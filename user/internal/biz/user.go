package biz

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strings"

	"math/big"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"user/internal/pkg/tracing"
)

var (
	// ErrInvalidCredentials 当提供的凭证无效时返回
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrEmailAlreadyExists 当邮箱已被注册时返回
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrInvalidVerificationCode 当验证码无效时返回
	ErrInvalidVerificationCode = errors.New("invalid verification code")

	// ErrVerificationCodeExpired 当验证码过期时返回
	ErrVerificationCodeExpired = errors.New("verification code expired")
)

// isUniqueConstraintError 判断错误是否为唯一约束错误（邮箱已存在）
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// GORM 的唯一约束错误信息通常包含 "UNIQUE constraint failed" 或 "Duplicate entry"
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate") ||
		strings.Contains(errStr, "unique") ||
		strings.Contains(errStr, "constraint failed")
}

// VerificationCode 验证码实体，用于存储和验证用户注册验证码
type VerificationCode struct {
	Email     string
	Code      string
	ExpiresAt time.Time
}

// User 用户基本信息表
type User struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Email        string    `gorm:"column:email;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	Nickname     string    `gorm:"column:nickname;not null;default:'新用户'" json:"nickname"`
	AvatarURL    string    `gorm:"column:avatar_url" json:"avatar_url,omitempty"`
	IsPremium    uint8     `gorm:"column:is_premium;not null;default:0" json:"is_premium"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

type UpdateUserRequest struct {
	// 允许用户更新昵称。使用指针 *string，可以接收 "" (零值) 或 nil (不更新)
	Nickname *string `json:"nickname"`

	//  *string 来表示：nil (不更新), 指向非空字符串的指针 (更新),
	AvatarURL *string `json:"avatar_url"`
}

// TableName 指定表名
func (User) TableName() string {
	return "user"
}

// UserRepository 用户数据访问接口
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, id int64, req *UpdateUserRequest) error
}

// CodeRepository 认证数据访问接口，定义了验证码相关的数据操作方法
type CodeRepository interface {
	// 验证码相关操作
	StoreVerificationCode(ctx context.Context, email, code string, expiresAt time.Time) error
	GetVerificationCode(ctx context.Context, email string) (*VerificationCode, error)
	DeleteVerificationCode(ctx context.Context, email string) error
	// 发送频率限制
	CheckAndSetSendRateLimit(ctx context.Context, email string, duration time.Duration) (bool, error)
}

// GreeterUsecase is a Greeter usecase.
type UserUsecase struct {
	userRepo UserRepository
	codeRepo CodeRepository
	authRepo AuthRepository
	log      *log.Helper
}

// NewUserUsecase new a User usecase.
func NewUserUsecase(userRepo UserRepository, codeRepo CodeRepository, authRepo AuthRepository, logger log.Logger) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
		codeRepo: codeRepo,
		authRepo: authRepo,
		log:      log.NewHelper(logger),
	}
}

// ErrTooManyRequests 发送请求过于频繁
var ErrTooManyRequests = errors.New("too many requests, please try again later")

// SendRegisterCode 发送注册验证码
func (uc *UserUsecase) SendRegisterCode(ctx context.Context, email string) error {
	ctx, span := tracing.StartSpan(ctx, "UserUsecase.SendRegisterCode")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation": "send_register_code",
		"email": email,
	})

	uc.log.WithContext(ctx).Infof("Sending registration code to email: %s", email)

	// 验证邮箱格式
	if email == "" {
		uc.log.WithContext(ctx).Warn("Empty email provided")
		return errors.New("email is required")
	}

	// 检查邮箱是否已注册
	_, err := uc.userRepo.GetByEmail(ctx, email)
	if err == nil {
		uc.log.WithContext(ctx).Infof("Email already registered: %s", email)
		return ErrEmailAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		uc.log.WithContext(ctx).Errorf("Database error when checking email: %s, error: %v", email, err)
		return err
	}

	// 检查发送频率限制（60秒内只能发送一次）
	// 这可以防止并发请求重复发送验证码
	ok, err := uc.codeRepo.CheckAndSetSendRateLimit(ctx, email, 60*time.Second)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to check rate limit for email: %s, error: %v", email, err)
		return err
	}
	if !ok {
		uc.log.WithContext(ctx).Warnf("Send verification code too frequently for email: %s", email)
		return ErrTooManyRequests
	}

	// 生成验证码
	code := generateVerificationCode()
	expiresAt := time.Now().Add(10 * time.Minute) // 10分钟过期

	// 存储验证码
	err = uc.codeRepo.StoreVerificationCode(ctx, email, code, expiresAt)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to store verification code for email: %s, error: %v", email, err)
		return err
	}

	// 发送邮件验证码
	err = uc.sendVerificationEmail(ctx, email, code)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to send verification email to: %s, error: %v", email, err)
		// 即使邮件发送失败，也不删除验证码，用户可能需要重新发送
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	uc.log.WithContext(ctx).Infof("Verification code sent successfully to: %s", email)
	return nil
}

// Register 用户注册
func (uc *UserUsecase) Register(ctx context.Context, email, password, code, nickname string) (*User, error) {
	ctx, span := tracing.StartSpan(ctx, "UserUsecase.Register")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation": "register",
		"email": email,
		"nickname": nickname,
	})

	uc.log.WithContext(ctx).Infof("Registering user with email: %s", email)

	// 参数验证
	if email == "" || password == "" || code == "" {
		uc.log.WithContext(ctx).Warn("Missing required fields for registration")
		return nil, errors.New("email, password and code are required")
	}

	// 验证验证码
	storedCode, err := uc.codeRepo.GetVerificationCode(ctx, email)
	if err != nil {
		uc.log.WithContext(ctx).Warnf("Failed to get verification code for email: %s, error: %v", email, err)
		return nil, ErrInvalidVerificationCode
	}

	if storedCode.Code != code {
		uc.log.WithContext(ctx).Warnf("Invalid verification code for email: %s", email)
		return nil, ErrInvalidVerificationCode
	}

	if time.Now().After(storedCode.ExpiresAt) {
		uc.log.WithContext(ctx).Warnf("Verification code expired for email: %s", email)
		return nil, ErrVerificationCodeExpired
	}

	// 密码强度验证
	if len(password) < 6 {
		uc.log.WithContext(ctx).Warnf("Password too short for email: %s", email)
		return nil, errors.New("password must be at least 6 characters long")
	}

	// 删除验证码
	err = uc.codeRepo.DeleteVerificationCode(ctx, email)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to delete verification code for email: %s, error: %v", email, err)
		// 不返回错误，因为用户已经通过验证
	}

	// 密码哈希
	hashedPassword, err := hashPassword(password)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to hash password for email: %s, error: %v", email, err)
		return nil, err
	}

	// 创建用户
	// 注意：这里不提前检查邮箱是否已存在，而是直接尝试创建
	// 如果邮箱已存在，数据库的唯一约束会阻止插入并返回错误
	// 这种方式避免了竞态条件问题
	user := &User{
		Email:        email,
		PasswordHash: hashedPassword,
		Nickname:     nickname,
		IsPremium:    0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		// 检查是否是唯一约束错误（邮箱已存在）
		if isUniqueConstraintError(err) {
			uc.log.WithContext(ctx).Infof("Email already registered during registration: %s", email)
			return nil, ErrEmailAlreadyExists
		}
		uc.log.WithContext(ctx).Errorf("Failed to create user with email: %s, error: %v", email, err)
		return nil, err
	}

	// 清空密码哈希，不返回给调用方
	user.PasswordHash = ""

	uc.log.WithContext(ctx).Infof("Successfully registered user with id: %d, email: %s", user.ID, email)
	return user, nil
}

// Login 用户登录
func (uc *UserUsecase) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	ctx, span := tracing.StartSpan(ctx, "UserUsecase.Login")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation": "login",
		"email": email,
	})

	uc.log.WithContext(ctx).Infof("User login attempt with email: %s", email)

	// 参数验证
	if email == "" || password == "" {
		uc.log.WithContext(ctx).Warn("Missing email or password for login")
		return nil, errors.New("email and password are required")
	}

	// 获取用户
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.WithContext(ctx).Warnf("User not found with email: %s", email)
			return nil, ErrInvalidCredentials // 为了安全，不暴露用户是否存在
		}
		uc.log.WithContext(ctx).Errorf("Database error when getting user with email: %s, error: %v", email, err)
		return nil, err
	}

	// 验证密码
	if !checkPasswordHash(password, user.PasswordHash) {
		uc.log.WithContext(ctx).Warnf("Invalid password for user with email: %s", email)
		return nil, ErrInvalidCredentials
	}

	// 生成令牌
	accessToken, accessExpiresIn, err := generateAccessToken(user.ID)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to generate access token for user id: %d, error: %v", user.ID, err)
		return nil, err
	}

	refreshToken, refreshExpiresIn, err := generateRefreshToken(user.ID)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to generate refresh token for user id: %d, error: %v", user.ID, err)
		return nil, err
	}

	// 存储刷新令牌
	refreshTokenExpiresAt := time.Now().Add(time.Duration(refreshExpiresIn) * time.Second)

	err = uc.authRepo.StoreRefreshToken(ctx, user.ID, refreshToken, refreshTokenExpiresAt)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to store refresh token for user id: %d, error: %v", user.ID, err)
		return nil, err
	}

	uc.log.WithContext(ctx).Infof("User login successful for user id: %d, email: %s", user.ID, email)
	return &TokenPair{
		AccessToken:      accessToken,
		AccessExpiresIn:  accessExpiresIn,
		RefreshToken:     refreshToken,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}

// generateVerificationCode 生成6位数字验证码
func generateVerificationCode() string {
	// 生成真正的数字验证码
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code[i] = byte(n.Int64()) + '0'
	}
	return string(code)
}

// hashPassword 使用bcrypt对密码进行哈希处理
//
// 参数:
//   - password: 明文密码
//
// 返回值:
//   - string: 哈希后的密码
//   - error: 错误信息
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPasswordHash 验证密码是否与哈希值匹配
//
// 参数:
//   - password: 明文密码
//   - hash: 哈希后的密码
//
// 返回值:
//   - bool: 密码是否匹配
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// sendVerificationEmail 发送验证码邮件
func (uc *UserUsecase) sendVerificationEmail(ctx context.Context, email, code string) error {
	ctx, span := tracing.StartSpan(ctx, "UserUsecase.sendVerificationEmail")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation": "send_verification_email",
		"email": email,
		"code_length": len(code),
	})

	// 1. 从环境变量获取 API Key
	apiKey := os.Getenv("SENDGRID_API_KEY")
	if apiKey == "" {
		uc.log.WithContext(ctx).Error("SENDGRID_API_KEY environment variable is not set")
		return errors.New("SENDGRID_API_KEY environment variable is required")
	}

	// 检查是否为测试环境（API key以"test-"开头）
	isTestMode := strings.HasPrefix(apiKey, "test-")
	if isTestMode {
		uc.log.WithContext(ctx).Infof("Test mode: skipping actual email send, email: %s, code: %s", email, code)
		return nil
	}

	// 2. 定义发件人邮箱（需要是在 SendGrid 中验证过的域名下的邮箱）
	fromEmail := mail.NewEmail("用户系统", "noreply@lyydsheep.xyz")

	// 3. 提取邮箱的用户名部分作为收件人称呼
	emailPrefix := strings.Split(email, "@")[0]
	if len(emailPrefix) > 3 {
		// 只显示邮箱前缀的前3个字符和后缀（例如：use***@example.com）
		emailPrefix = emailPrefix[:3] + strings.Repeat("*", len(emailPrefix)-3)
	}

	// 4. 定义收件人
	toEmail := mail.NewEmail(emailPrefix, email)

	// 5. 定义邮件主题
	subject := "您的验证码 - 请在10分钟内使用"

	// 6. 构建纯文本内容
	plainTextContent := fmt.Sprintf(`您好！

您的注册验证码是：%s

此验证码将在10分钟后失效。为了保障您的账户安全，请勿将验证码告知他人。

如果您没有进行注册操作，请忽略此邮件。

感谢您的使用！
`, code)

	// 7. 构建HTML内容
	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>邮箱验证</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 40px auto; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 40px 30px; text-align: center; color: white; }
        .header h1 { font-size: 28px; margin-bottom: 10px; font-weight: 600; }
        .header p { font-size: 16px; opacity: 0.9; }
        .content { padding: 40px 30px; }
        .greeting { font-size: 16px; color: #333; margin-bottom: 25px; line-height: 1.6; }
        .code-box { background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%); border-radius: 12px; padding: 30px; text-align: center; margin: 30px 0; box-shadow: 0 4px 15px rgba(245, 87, 108, 0.3); }
        .code-label { font-size: 14px; color: white; margin-bottom: 10px; opacity: 0.9; }
        .code { font-size: 36px; font-weight: bold; color: white; letter-spacing: 8px; font-family: 'Courier New', monospace; }
        .warning { background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 25px 0; border-radius: 4px; }
        .warning-title { color: #856404; font-weight: 600; margin-bottom: 8px; font-size: 14px; }
        .warning-text { color: #856404; font-size: 13px; line-height: 1.6; }
        .footer { background-color: #f8f9fa; padding: 25px 30px; text-align: center; color: #666; font-size: 13px; line-height: 1.6; }
        .footer a { color: #667eea; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 邮箱验证码</h1>
            <p>您的安全验证信息</p>
        </div>

        <div class="content">
            <div class="greeting">
                您好！<br>
                感谢您注册我们的服务。请使用下面的验证码完成注册：
            </div>

            <div class="code-box">
                <div class="code-label">您的验证码</div>
                <div class="code">%s</div>
            </div>

            <div class="warning">
                <div class="warning-title">⏰ 重要提醒</div>
                <div class="warning-text">
                    • 验证码将在 <strong>10 分钟</strong> 后失效<br>
                    • 请勿将验证码告知他人<br>
                    • 如果您没有进行注册操作，请忽略此邮件
                </div>
            </div>
        </div>

        <div class="footer">
            <p>此邮件由系统自动发送，请勿直接回复。</p>
            <p>如有问题请联系 <a href="mailto:support@lyydsheep.xyz">support@lyydsheep.xyz</a></p>
            <p style="margin-top: 15px; color: #999;">© 2025 您的应用名称. 保留所有权利。</p>
        </div>
    </div>
</body>
</html>
`, code)

	// 8. 构造完整的邮件对象
	message := mail.NewSingleEmail(
		fromEmail,
		subject,
		toEmail,
		plainTextContent,
		htmlContent,
	)

	// 9. 创建 SendGrid 客户端
	client := sendgrid.NewSendClient(apiKey)

	// 10. 发送邮件
	uc.log.WithContext(ctx).Infof("Sending verification email to: %s", email)
	response, err := client.Send(message)

	// 11. 处理响应和错误
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to send email: %v", err)
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		uc.log.WithContext(ctx).Infof("Verification email sent successfully to: %s, status: %d", email, response.StatusCode)
		return nil
	} else {
		uc.log.WithContext(ctx).Errorf("Failed to send email, status: %d, body: %s", response.StatusCode, response.Body)
		return fmt.Errorf("failed to send verification email: status %d", response.StatusCode)
	}
}

// UpdateUser 更新用户信息
func (uc *UserUsecase) UpdateUser(ctx context.Context, id int64, req *UpdateUserRequest) error {
	uc.log.WithContext(ctx).Infof("Updating user with id: %d", id)

	// 参数验证
	if req == nil {
		uc.log.WithContext(ctx).Warn("UpdateUser request is nil")
		return errors.New("update request is required")
	}

	// 更新用户信息
	err := uc.userRepo.Update(ctx, id, req)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to update user with id: %d, error: %v", id, err)
		return err
	}

	uc.log.WithContext(ctx).Infof("Successfully updated user with id: %d", id)
	return nil
}

// GetUserByID 根据ID获取用户信息
func (uc *UserUsecase) GetUserByID(ctx context.Context, id int64) (*User, error) {
	uc.log.WithContext(ctx).Infof("Getting user with id: %d", id)

	// 参数验证
	if id <= 0 {
		uc.log.WithContext(ctx).Warnf("Invalid user id: %d", id)
		return nil, errors.New("invalid user id")
	}

	// 获取用户信息
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get user with id: %d, error: %v", id, err)
		return nil, err
	}

	uc.log.WithContext(ctx).Infof("Successfully got user with id: %d", id)
	return user, nil
}
