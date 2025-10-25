package biz

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"time"
)

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
}

// GreeterUsecase is a Greeter usecase.
type UserUsecase struct {
	userRepo UserRepository
	codeRepo CodeRepository
	log      *log.Helper
}

// NewGreeterUsecase new a Greeter usecase.
func NewUserUsecase(userRepo UserRepository, codeRepo CodeRepository, logger log.Logger) *UserUsecase {
	return &UserUsecase{userRepo: userRepo, codeRepo: codeRepo, log: log.NewHelper(logger)}
}

// SendRegisterCode 发送注册验证码
func (uc *UserUsecase) SendRegisterCode(ctx context.Context, email string) error {
	uc.log.Log(log.LevelInfo, "Sending registration code to email: ", email)

	// 验证邮箱格式
	if email == "" {
		uc.log.Log(log.LevelWarn, "Empty email provided")
		return errors.New("email is required")
	}

	// 检查邮箱是否已注册
	_, err := uc.userRepo.GetByEmail(ctx, email)
	if err == nil {
		uc.log.Log(log.LevelInfo, "Email already registered: ", email)
		return ErrEmailAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		uc.log.Log(log.LevelError, "Database error when checking email: ", email, ", error: ", err)
		return err
	}

	// 生成验证码
	code := generateVerificationCode()
	expiresAt := time.Now().Add(10 * time.Minute) // 10分钟过期

	// 存储验证码
	err = uc.codeRepo.StoreVerificationCode(ctx, email, code, expiresAt)
	if err != nil {
		uc.log.Log(log.LevelError, "Failed to store verification code for email: ", email, ", error: ", err)
		return err
	}

	// 这里应该发送邮件，暂时只记录日志
	uc.log.Log(log.LevelInfo, "Verification code for email ", email, ": ", code)

	return nil
}

// Register 用户注册
func (uc *UserUsecase) Register(ctx context.Context, email, password, code, nickname string) (*User, error) {
	uc.log.Log(log.LevelInfo, "Registering user with email: ", email)

	// 参数验证
	if email == "" || password == "" || code == "" {
		uc.log.Log(log.LevelWarn, "Missing required fields for registration")
		return nil, errors.New("email, password and code are required")
	}

	// 验证验证码
	storedCode, err := uc.codeRepo.GetVerificationCode(ctx, email)
	if err != nil {
		uc.log.Log(log.LevelWarn, "Failed to get verification code for email: ", email, ", error: ", err)
		return nil, ErrInvalidVerificationCode
	}

	if storedCode.Code != code {
		uc.log.Log(log.LevelWarn, "Invalid verification code for email: ", email)
		return nil, ErrInvalidVerificationCode
	}

	if time.Now().After(storedCode.ExpiresAt) {
		uc.log.Log(log.LevelWarn, "Verification code expired for email: ", email)
		return nil, ErrVerificationCodeExpired
	}

	// 删除验证码
	err = uc.codeRepo.DeleteVerificationCode(ctx, email)
	if err != nil {
		uc.log.Log(log.LevelError, "Failed to delete verification code for email: ", email, ", error: ", err)
		// 不返回错误，因为用户已经通过验证
	}

	// 检查邮箱是否已注册
	_, err = uc.userRepo.GetByEmail(ctx, email)
	if err == nil {
		uc.log.Log(log.LevelInfo, "Email already registered during registration: ", email)
		return nil, ErrEmailAlreadyExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		uc.log.Log(log.LevelError, "Database error when checking email during registration: ", email, ", error: ", err)
		return nil, err
	}

	// 密码强度验证
	if len(password) < 6 {
		uc.log.Log(log.LevelWarn, "Password too short for email: ", email)
		return nil, errors.New("password must be at least 6 characters long")
	}

	// 密码哈希
	hashedPassword, err := hashPassword(password)
	if err != nil {
		uc.log.Log(log.LevelError, "Failed to hash password for email: ", email, ", error: ", err)
		return nil, err
	}

	// 创建用户
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
		uc.log.Log(log.LevelError, "Failed to create user with email: ", email, ", error: ", err)
		return nil, err
	}

	// 清空密码哈希，不返回给调用方
	user.PasswordHash = ""

	uc.log.Log(log.LevelInfo, "Successfully registered user with id: ", user.ID, ", email: ", email)
	return user, nil
}
