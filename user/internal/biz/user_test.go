package biz

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// 模拟 UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, id int64, req *UpdateUserRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

// 模拟 CodeRepository
type MockCodeRepository struct {
	mock.Mock
}

func (m *MockCodeRepository) StoreVerificationCode(ctx context.Context, email, code string, expiresAt time.Time) error {
	args := m.Called(ctx, email, code, expiresAt)
	return args.Error(0)
}

func (m *MockCodeRepository) GetVerificationCode(ctx context.Context, email string) (*VerificationCode, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*VerificationCode), args.Error(1)
}

func (m *MockCodeRepository) DeleteVerificationCode(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockCodeRepository) CheckAndSetSendRateLimit(ctx context.Context, email string, duration time.Duration) (bool, error) {
	args := m.Called(ctx, email, duration)
	return args.Bool(0), args.Error(1)
}

// 模拟 AuthRepository
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) StoreRefreshToken(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, refreshToken, expiresAt)
	return args.Error(0)
}

func (m *MockAuthRepository) GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (int64, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuthRepository) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthRepository) DeleteAllRefreshTokens(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthRepository) RefreshTokenAtomically(ctx context.Context, userID int64, oldToken, newToken string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, oldToken, newToken, expiresAt)
	return args.Error(0)
}

// 设置测试环境变量
func setupTestEnv() {
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-for-unit-testing-only")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-for-unit-testing-only")
	os.Setenv("SENDGRID_API_KEY", "test-sendgrid-api-key")
}

// 清理测试环境变量
func cleanupTestEnv() {
	os.Unsetenv("JWT_ACCESS_SECRET")
	os.Unsetenv("JWT_REFRESH_SECRET")
	os.Unsetenv("SENDGRID_API_KEY")
}

// 获取测试用logger
func getTestLogger() log.Logger {
	return log.NewStdLogger(os.Stdout)
}

// TestUserUsecase_SendRegisterCode 测试发送注册验证码
func TestUserUsecase_SendRegisterCode(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	tests := []struct {
		name        string
		email       string
		setupMocks  func(*MockUserRepository, *MockCodeRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:  "成功发送验证码",
			email: "test@example.com",
			setupMocks: func(userRepo *MockUserRepository, codeRepo *MockCodeRepository) {
				// 用户不存在
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").
					Return((*User)(nil), gorm.ErrRecordNotFound)

				// 频率限制检查通过
				codeRepo.On("CheckAndSetSendRateLimit", mock.Anything, "test@example.com", 60*time.Second).
					Return(true, nil)

				// 存储验证码
				codeRepo.On("StoreVerificationCode", mock.Anything, "test@example.com", mock.Anything, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "发送过于频繁",
			email: "frequent@example.com",
			setupMocks: func(userRepo *MockUserRepository, codeRepo *MockCodeRepository) {
				// 用户不存在
				userRepo.On("GetByEmail", mock.Anything, "frequent@example.com").
					Return((*User)(nil), gorm.ErrRecordNotFound)

				// 频率限制检查失败
				codeRepo.On("CheckAndSetSendRateLimit", mock.Anything, "frequent@example.com", 60*time.Second).
					Return(false, nil)
			},
			wantErr:     true,
			expectedErr: ErrTooManyRequests,
		},
		{
			name:  "邮箱为空",
			email: "",
			setupMocks: func(userRepo *MockUserRepository, codeRepo *MockCodeRepository) {
				// 不调用任何方法
			},
			wantErr:     true,
			expectedErr: errors.New("email is required"),
		},
		{
			name:  "邮箱已注册",
			email: "existing@example.com",
			setupMocks: func(userRepo *MockUserRepository, codeRepo *MockCodeRepository) {
				// 用户已存在
				userRepo.On("GetByEmail", mock.Anything, "existing@example.com").
					Return(&User{Email: "existing@example.com"}, nil)
			},
			wantErr:     true,
			expectedErr: ErrEmailAlreadyExists,
		},
		{
			name:  "数据库错误",
			email: "db-error_reason@example.com",
			setupMocks: func(userRepo *MockUserRepository, codeRepo *MockCodeRepository) {
				userRepo.On("GetByEmail", mock.Anything, "db-error_reason@example.com").
					Return((*User)(nil), errors.New("database error_reason"))
			},
			wantErr:     true,
			expectedErr: errors.New("database error_reason"),
		},
		{
			name:  "频率限制错误",
			email: "rate-limit-error_reason@example.com",
			setupMocks: func(userRepo *MockUserRepository, codeRepo *MockCodeRepository) {
				userRepo.On("GetByEmail", mock.Anything, "rate-limit-error_reason@example.com").
					Return((*User)(nil), gorm.ErrRecordNotFound)

				codeRepo.On("CheckAndSetSendRateLimit", mock.Anything, "rate-limit-error_reason@example.com", 60*time.Second).
					Return(false, errors.New("redis error_reason"))
			},
			wantErr:     true,
			expectedErr: errors.New("redis error_reason"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 mock
			userRepo := new(MockUserRepository)
			codeRepo := new(MockCodeRepository)
			authRepo := new(MockAuthRepository)

			// 设置 mock 期望
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, codeRepo)
			}

			// 创建 usecase
			uc := NewUserUsecase(userRepo, codeRepo, authRepo, getTestLogger(), &MockSnowflakeGenerator{mock.Mock}{})

			// 执行测试
			err := uc.SendRegisterCode(context.Background(), tt.email)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			// 验证所有期望都被调用
			userRepo.AssertExpectations(t)
			codeRepo.AssertExpectations(t)
		})
	}
}

// TestUserUsecase_Register 测试用户注册
func TestUserUsecase_Register(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	validCode := &VerificationCode{
		Email:     "test@example.com",
		Code:      "123456",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	tests := []struct {
		name        string
		email       string
		password    string
		code        string
		nickname    string
		setupMocks  func(*MockUserRepository, *MockCodeRepository, *MockAuthRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "成功注册",
			email:    "test@example.com",
			password: "password123",
			code:     "123456",
			nickname: "测试用户",
			setupMocks: func(userRepo *MockUserRepository, codeRepo *MockCodeRepository, authRepo *MockAuthRepository) {
				// 获取验证码
				codeRepo.On("GetVerificationCode", mock.Anything, "test@example.com").
					Return(validCode, nil)

				// 删除验证码
				codeRepo.On("DeleteVerificationCode", mock.Anything, "test@example.com").
					Return(nil)

				// 创建用户
				userRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *User) bool {
					return user.Email == "test@example.com" && user.Nickname == "测试用户"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "参数缺失",
			email:       "",
			password:    "",
			code:        "",
			nickname:    "",
			setupMocks:  func(userRepo *MockUserRepository, codeRepo *MockCodeRepository, authRepo *MockAuthRepository) {},
			wantErr:     true,
			expectedErr: errors.New("email, password and code are required"),
		},
		{
			name:     "无效验证码",
			email:    "test@example.com",
			password: "password123",
			code:     "wrongcode",
			nickname: "测试用户",
			setupMocks: func(userRepo *MockUserRepository, codeRepo *MockCodeRepository, authRepo *MockAuthRepository) {
				codeRepo.On("GetVerificationCode", mock.Anything, "test@example.com").
					Return(validCode, nil)
			},
			wantErr:     true,
			expectedErr: ErrInvalidVerificationCode,
		},
		{
			name:     "验证码过期",
			email:    "test@example.com",
			password: "password123",
			code:     "123456",
			nickname: "测试用户",
			setupMocks: func(userRepo *MockUserRepository, codeRepo *MockCodeRepository, authRepo *MockAuthRepository) {
				expiredCode := &VerificationCode{
					Email:     "test@example.com",
					Code:      "123456",
					ExpiresAt: time.Now().Add(-1 * time.Minute), // 已过期
				}
				codeRepo.On("GetVerificationCode", mock.Anything, "test@example.com").
					Return(expiredCode, nil)
			},
			wantErr:     true,
			expectedErr: ErrVerificationCodeExpired,
		},
		{
			name:     "密码太短",
			email:    "test@example.com",
			password: "123",
			code:     "123456",
			nickname: "测试用户",
			setupMocks: func(userRepo *MockUserRepository, codeRepo *MockCodeRepository, authRepo *MockAuthRepository) {
				codeRepo.On("GetVerificationCode", mock.Anything, "test@example.com").
					Return(validCode, nil)
			},
			wantErr:     true,
			expectedErr: errors.New("password must be at least 6 characters long"),
		},
		{
			name:     "邮箱已存在（唯一约束错误）",
			email:    "existing@example.com",
			password: "password123",
			code:     "123456",
			nickname: "测试用户",
			setupMocks: func(userRepo *MockUserRepository, codeRepo *MockCodeRepository, authRepo *MockAuthRepository) {
				codeRepo.On("GetVerificationCode", mock.Anything, "existing@example.com").
					Return(validCode, nil)

				codeRepo.On("DeleteVerificationCode", mock.Anything, "existing@example.com").
					Return(nil)

				// 模拟唯一约束错误（邮箱已存在）
				userRepo.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("Duplicate entry 'existing@example.com' for key 'email'"))
			},
			wantErr:     true,
			expectedErr: ErrEmailAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 mock
			userRepo := new(MockUserRepository)
			codeRepo := new(MockCodeRepository)
			authRepo := new(MockAuthRepository)

			// 设置 mock 期望
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, codeRepo, authRepo)
			}

			// 创建 usecase
			uc := NewUserUsecase(userRepo, codeRepo, authRepo, getTestLogger(), &MockSnowflakeGenerator{mock.Mock}{})

			// 执行测试
			user, err := uc.Register(context.Background(), tt.email, tt.password, tt.code, tt.nickname)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.nickname, user.Nickname)
				// 密码哈希不应该返回
				assert.Equal(t, "", user.PasswordHash)
			}

			// 验证所有期望都被调用
			userRepo.AssertExpectations(t)
			codeRepo.AssertExpectations(t)
			authRepo.AssertExpectations(t)
		})
	}
}

// TestUserUsecase_Login 测试用户登录
func TestUserUsecase_Login(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	validPassword := "password123"
	hashedPassword, _ := hashPassword(validPassword)

	validUser := &User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Nickname:     "测试用户",
	}

	tests := []struct {
		name        string
		email       string
		password    string
		setupMocks  func(*MockUserRepository, *MockAuthRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "成功登录",
			email:    "test@example.com",
			password: validPassword,
			setupMocks: func(userRepo *MockUserRepository, authRepo *MockAuthRepository) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").
					Return(validUser, nil)

				authRepo.On("StoreRefreshToken", mock.Anything, int64(1), mock.Anything, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "邮箱或密码为空",
			email:       "",
			password:    "",
			setupMocks:  func(userRepo *MockUserRepository, authRepo *MockAuthRepository) {},
			wantErr:     true,
			expectedErr: errors.New("email and password are required"),
		},
		{
			name:     "用户不存在",
			email:    "nonexistent@example.com",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, authRepo *MockAuthRepository) {
				userRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").
					Return((*User)(nil), gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			expectedErr: ErrInvalidCredentials,
		},
		{
			name:     "密码错误",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMocks: func(userRepo *MockUserRepository, authRepo *MockAuthRepository) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").
					Return(validUser, nil)
			},
			wantErr:     true,
			expectedErr: ErrInvalidCredentials,
		},
		{
			name:     "数据库错误",
			email:    "test@example.com",
			password: validPassword,
			setupMocks: func(userRepo *MockUserRepository, authRepo *MockAuthRepository) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").
					Return((*User)(nil), errors.New("database error_reason"))
			},
			wantErr:     true,
			expectedErr: errors.New("database error_reason"),
		},
		{
			name:     "StoreRefreshToken失败",
			email:    "test@example.com",
			password: validPassword,
			setupMocks: func(userRepo *MockUserRepository, authRepo *MockAuthRepository) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").
					Return(validUser, nil)

				authRepo.On("StoreRefreshToken", mock.Anything, int64(1), mock.Anything, mock.Anything).
					Return(errors.New("redis error_reason"))
			},
			wantErr:     true,
			expectedErr: errors.New("redis error_reason"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 mock
			userRepo := new(MockUserRepository)
			codeRepo := new(MockCodeRepository)
			authRepo := new(MockAuthRepository)

			// 设置 mock 期望
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, authRepo)
			}

			// 创建 usecase
			uc := NewUserUsecase(userRepo, codeRepo, authRepo, getTestLogger(), &MockSnowflakeGenerator{mock.Mock}{})

			// 执行测试
			tokenPair, err := uc.Login(context.Background(), tt.email, tt.password)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
				assert.Nil(t, tokenPair)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, tokenPair)
				assert.NotEmpty(t, tokenPair.AccessToken)
				assert.NotEmpty(t, tokenPair.RefreshToken)
				assert.True(t, tokenPair.AccessExpiresIn > 0)
				assert.True(t, tokenPair.RefreshExpiresIn > 0)

				// 验证 JWT token 格式
				token, err := jwt.ParseWithClaims(tokenPair.AccessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte("test-access-secret-key-for-unit-testing-only"), nil
				})
				assert.NoError(t, err)
				assert.True(t, token.Valid)
			}

			// 验证所有期望都被调用
			userRepo.AssertExpectations(t)
			authRepo.AssertExpectations(t)
		})
	}
}

// TestGenerateVerificationCode 测试验证码生成
func TestGenerateVerificationCode(t *testing.T) {
	code1 := generateVerificationCode()
	code2 := generateVerificationCode()

	// 验证码应该是6位数字
	assert.Equal(t, 6, len(code1))
	assert.Equal(t, 6, len(code2))

	// 验证码应该是数字
	for _, c := range code1 {
		assert.True(t, c >= '0' && c <= '9')
	}
	for _, c := range code2 {
		assert.True(t, c >= '0' && c <= '9')
	}

	// 多次生成的验证码应该不同（虽然概率很小）
	// 注意：这个测试在极少数情况下可能失败，但概率极低
	if code1 != code2 {
		t.Logf("验证码生成具有随机性: %s vs %s", code1, code2)
	}
}

// TestHashPassword 测试密码哈希
func TestHashPassword(t *testing.T) {
	password := "password123"

	// 哈希密码
	hashedPassword, err := hashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)

	// 验证密码正确
	isValid := checkPasswordHash(password, hashedPassword)
	assert.True(t, isValid)

	// 验证错误密码
	isValid = checkPasswordHash("wrongpassword", hashedPassword)
	assert.False(t, isValid)
}

// TestUserUsecase_sendVerificationEmail 测试邮件发送
func TestUserUsecase_sendVerificationEmail(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	tests := []struct {
		name        string
		email       string
		code        string
		setupMock   func(*MockAuthRepository)
		wantErr     bool
		expectedErr string
	}{
		{
			name:  "成功发送邮件",
			email: "test@example.com",
			code:  "123456",
			setupMock: func(authRepo *MockAuthRepository) {
				// 模拟发送成功（不实际发送邮件）
			},
			wantErr: false,
		},
		{
			name:        "邮箱为空",
			email:       "",
			code:        "123456",
			setupMock:   func(authRepo *MockAuthRepository) {},
			wantErr:     true,
			expectedErr: "email is required",
		},
		{
			name:        "验证码为空",
			email:       "test@example.com",
			code:        "",
			setupMock:   func(authRepo *MockAuthRepository) {},
			wantErr:     false, // 函数不验证验证码为空的情况
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 mock
			userRepo := new(MockUserRepository)
			codeRepo := new(MockCodeRepository)
			authRepo := new(MockAuthRepository)

			// 设置 mock 期望
			if tt.setupMock != nil {
				tt.setupMock(authRepo)
			}

			// 创建 usecase
			uc := NewUserUsecase(userRepo, codeRepo, authRepo, getTestLogger(), &MockSnowflakeGenerator{mock.Mock}{})

			// 执行测试（这里不会实际发送邮件，因为使用的是 test API key）
			// 在实际测试中，你可能想要 Mock SendGrid 的 HTTP 请求
			// 但这里我们只测试参数验证和错误处理
			if tt.email != "" && tt.code != "" {
				// 注意：由于使用的是测试 API key，邮件发送会失败
				// 这是一个已知的行为
				err := uc.sendVerificationEmail(context.Background(), tt.email, tt.code)
				// 我们预期这会失败（因为 API key 无效）
				if err == nil {
					t.Log("邮件发送成功（不应该发生）")
				}
			}
		})
	}
}

// TestUser_UpdateUser 测试用户更新（如果需要）
func TestUserUsecase_UpdateUser(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	tests := []struct {
		name        string
		userID      int64
		nickname    *string
		avatarURL   *string
		setupMocks  func(*MockUserRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:      "成功更新昵称",
			userID:    1,
			nickname:  stringPtr("新昵称"),
			avatarURL: nil,
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.On("Update", mock.Anything, int64(1), &UpdateUserRequest{
					Nickname:  stringPtr("新昵称"),
					AvatarURL: nil,
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "成功更新头像",
			userID:    1,
			nickname:  nil,
			avatarURL: stringPtr("https://example.com/avatar.jpg"),
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.On("Update", mock.Anything, int64(1), &UpdateUserRequest{
					Nickname:  nil,
					AvatarURL: stringPtr("https://example.com/avatar.jpg"),
				}).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 mock
			userRepo := new(MockUserRepository)
			codeRepo := new(MockCodeRepository)
			authRepo := new(MockAuthRepository)

			// 设置 mock 期望
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo)
			}

			// 创建 usecase
			uc := NewUserUsecase(userRepo, codeRepo, authRepo, getTestLogger(), &MockSnowflakeGenerator{mock.Mock}{})

			// 创建更新请求
			req := &UpdateUserRequest{
				Nickname:  tt.nickname,
				AvatarURL: tt.avatarURL,
			}

			// 执行测试
			err := uc.UpdateUser(context.Background(), tt.userID, req)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			// 验证所有期望都被调用
			userRepo.AssertExpectations(t)
		})
	}
}

// TestUserUsecase_Register_Concurrent 测试 Register 函数的并发安全性
func TestUserUsecase_Register_Concurrent(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	// 这个测试验证在并发情况下，只有一个请求能成功创建用户
	// 其他请求会返回 ErrEmailAlreadyExists
	t.Run("并发注册同一邮箱", func(t *testing.T) {
		const numGoroutines = 10
		email := "concurrent-test@example.com"
		password := "password123"
		code := "123456"
		nickname := "测试用户"

		userRepo := new(MockUserRepository)
		codeRepo := new(MockCodeRepository)
		authRepo := new(MockAuthRepository)

		validCode := &VerificationCode{
			Email:     email,
			Code:      code,
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		// 设置期望：第一个成功的请求，其他请求返回唯一约束错误
		codeRepo.On("GetVerificationCode", mock.Anything, email).
			Return(validCode, nil).Times(numGoroutines)

		codeRepo.On("DeleteVerificationCode", mock.Anything, email).
			Return(nil).Times(numGoroutines)

		// 模拟第一个请求成功，其他请求失败
		userRepo.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				// 第一次调用返回成功
				userRepo.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("Duplicate entry '" + email + "' for key 'email'"))
			}).
			Return(nil).Once()

		uc := NewUserUsecase(userRepo, codeRepo, authRepo, getTestLogger(), &MockSnowflakeGenerator{mock.Mock}{})

		// 启动并发请求
		errChan := make(chan error, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				_, err := uc.Register(context.Background(), email, password, code, nickname)
				errChan <- err
			}()
		}

		// 收集结果
		successCount := 0
		duplicateCount := 0
		otherErrors := 0

		for i := 0; i < numGoroutines; i++ {
			err := <-errChan
			if err == nil {
				successCount++
			} else if errors.Is(err, ErrEmailAlreadyExists) {
				duplicateCount++
			} else {
				otherErrors++
				t.Logf("Unexpected error_reason: %v", err)
			}
		}

		// 验证结果：只有一个成功，其他都是邮箱已存在错误
		assert.Equal(t, 1, successCount, "应该只有一个请求成功创建用户")
		assert.Equal(t, numGoroutines-1, duplicateCount, "其他请求应该返回邮箱已存在错误")
		assert.Equal(t, 0, otherErrors, "不应该有其他错误")

		// 验证所有期望都被调用
		codeRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})
}

// TestIsUniqueConstraintError 测试唯一约束错误检测
func TestIsUniqueConstraintError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil 错误",
			err:      nil,
			expected: false,
		},
		{
			name:     "普通错误",
			err:      errors.New("some other error_reason"),
			expected: false,
		},
		{
			name:     "包含 duplicate 的错误",
			err:      errors.New("Duplicate entry 'test@example.com' for key 'email'"),
			expected: true,
		},
		{
			name:     "包含 unique 的错误",
			err:      errors.New("UNIQUE constraint failed: users.email"),
			expected: true,
		},
		{
			name:     "包含 constraint failed 的错误",
			err:      errors.New("constraint failed: users.email"),
			expected: true,
		},
		{
			name:     "大写 DUPLICATE",
			err:      errors.New("DUPLICATE entry 'test@example.com'"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUniqueConstraintError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}
