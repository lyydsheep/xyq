package biz

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestAuthUsecase_RefreshToken 测试令牌刷新
func TestAuthUsecase_RefreshToken(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	tests := []struct {
		name         string
		refreshToken string
		setupMocks   func(*MockAuthRepository)
		wantErr      bool
		expectedErr  error
	}{
		{
			name:         "成功刷新令牌",
			refreshToken: "valid-refresh-token",
			setupMocks: func(authRepo *MockAuthRepository) {
				// 模拟成功获取用户ID
				authRepo.On("GetUserIDByRefreshToken", mock.Anything, "valid-refresh-token").
					Return(int64(123), nil)

				// 模拟原子刷新成功
				authRepo.On("RefreshTokenAtomically", mock.Anything, int64(123), "valid-refresh-token", mock.Anything, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "刷新令牌为空",
			refreshToken: "",
			setupMocks: func(authRepo *MockAuthRepository) {
				// 不调用任何方法
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name:         "无效的刷新令牌",
			refreshToken: "invalid-token",
			setupMocks: func(authRepo *MockAuthRepository) {
				// 模拟令牌不存在
				authRepo.On("GetUserIDByRefreshToken", mock.Anything, "invalid-token").
					Return(int64(0), errors.New("token not found"))
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name:         "用户ID获取失败",
			refreshToken: "error_reason-token",
			setupMocks: func(authRepo *MockAuthRepository) {
				// 模拟数据库错误
				authRepo.On("GetUserIDByRefreshToken", mock.Anything, "error_reason-token").
					Return(int64(0), errors.New("database error_reason"))
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name:         "正常刷新流程",
			refreshToken: "normal-refresh-token",
			setupMocks: func(authRepo *MockAuthRepository) {
				authRepo.On("GetUserIDByRefreshToken", mock.Anything, "normal-refresh-token").
					Return(int64(456), nil)

				authRepo.On("RefreshTokenAtomically", mock.Anything, int64(456), "normal-refresh-token", mock.Anything, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "原子刷新失败",
			refreshToken: "atomic-fail-token",
			setupMocks: func(authRepo *MockAuthRepository) {
				authRepo.On("GetUserIDByRefreshToken", mock.Anything, "atomic-fail-token").
					Return(int64(123), nil)

				// 模拟原子刷新失败
				authRepo.On("RefreshTokenAtomically", mock.Anything, int64(123), "atomic-fail-token", mock.Anything, mock.Anything).
					Return(errors.New("redis error_reason"))
			},
			wantErr:     true,
			expectedErr: errors.New("redis error_reason"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 mock
			authRepo := new(MockAuthRepository)

			// 设置 mock 期望
			if tt.setupMocks != nil {
				tt.setupMocks(authRepo)
			}

			// 创建 usecase
			uc := NewAuthUsecase(authRepo, getTestLogger())

			// 执行测试
			tokenPair, err := uc.RefreshToken(context.Background(), tt.refreshToken)

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

				// 验证访问令牌JWT格式
				accessToken, err := jwt.ParseWithClaims(tokenPair.AccessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte("test-access-secret-key-for-unit-testing-only"), nil
				})
				assert.NoError(t, err)
				assert.True(t, accessToken.Valid)

				// 验证刷新令牌JWT格式
				refreshToken, err := jwt.ParseWithClaims(tokenPair.RefreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte("test-refresh-secret-key-for-unit-testing-only"), nil
				})
				assert.NoError(t, err)
				assert.True(t, refreshToken.Valid)
			}

			// 验证所有期望都被调用
			authRepo.AssertExpectations(t)
		})
	}
}

// TestAuthUsecase_Logout 测试用户登出
func TestAuthUsecase_Logout(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	tests := []struct {
		name         string
		refreshToken string
		setupMocks   func(*MockAuthRepository)
		wantErr      bool
		expectedErr  error
	}{
		{
			name:         "成功登出",
			refreshToken: "valid-refresh-token",
			setupMocks: func(authRepo *MockAuthRepository) {
				// 模拟成功删除刷新令牌
				authRepo.On("DeleteRefreshToken", mock.Anything, "valid-refresh-token").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "刷新令牌为空",
			refreshToken: "",
			setupMocks: func(authRepo *MockAuthRepository) {
				// 不调用任何方法
			},
			wantErr:     true,
			expectedErr: errors.New("refresh token is required"),
		},
		{
			name:         "删除刷新令牌失败",
			refreshToken: "delete-fail-token",
			setupMocks: func(authRepo *MockAuthRepository) {
				// 模拟删除失败
				authRepo.On("DeleteRefreshToken", mock.Anything, "delete-fail-token").
					Return(errors.New("redis error_reason"))
			},
			wantErr:     true,
			expectedErr: errors.New("redis error_reason"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 mock
			authRepo := new(MockAuthRepository)

			// 设置 mock 期望
			if tt.setupMocks != nil {
				tt.setupMocks(authRepo)
			}

			// 创建 usecase
			uc := NewAuthUsecase(authRepo, getTestLogger())

			// 执行测试
			err := uc.Logout(context.Background(), tt.refreshToken)

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
			authRepo.AssertExpectations(t)
		})
	}
}

// TestAuthUsecase_ValidateToken 测试令牌验证
func TestAuthUsecase_ValidateToken(t *testing.T) {
	setupTestEnv()
	defer cleanupTestEnv()

	// 生成一个有效的访问令牌用于测试
	validAccessToken, _, err := generateAccessToken(123)
	require.NoError(t, err)

	// 生成一个过期的访问令牌
	expiredClaims := &jwt.RegisteredClaims{
		Subject:   "123",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 已过期
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredAccessToken, err := expiredToken.SignedString([]byte("test-access-secret-key-for-unit-testing-only"))
	require.NoError(t, err)

	tests := []struct {
		name           string
		accessToken    string
		setupMocks     func(*MockAuthRepository)
		wantErr        bool
		expectedErr    error
		expectedUserID int64
	}{
		{
			name:        "成功验证有效令牌",
			accessToken: validAccessToken,
			setupMocks: func(authRepo *MockAuthRepository) {
				// 不需要调用mock方法
			},
			wantErr:        false,
			expectedUserID: 123,
		},
		{
			name:        "访问令牌为空",
			accessToken: "",
			setupMocks: func(authRepo *MockAuthRepository) {
				// 不调用任何方法
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name:        "无效的令牌格式",
			accessToken: "invalid-token-format",
			setupMocks: func(authRepo *MockAuthRepository) {
				// 不调用任何方法
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name:        "令牌已过期",
			accessToken: expiredAccessToken,
			setupMocks: func(authRepo *MockAuthRepository) {
				// 不调用任何方法
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken, // ParseWithClaims在解析过期token时会失败
		},
		{
			name:        "错误的签名",
			accessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMiLCJleHAiOjk5OTk5OTk5OTksImlhdCI6MTUxNjIzOTAyMiwibmJmIjoxNTE2MjM5MDIyLCJpZCI6InJlZnJlc2hfMTIzIn0.wrong-signature",
			setupMocks: func(authRepo *MockAuthRepository) {
				// 不调用任何方法
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name: "无效的用户ID",
			accessToken: func() string {
				// 创建一个包含无效用户ID的令牌
				claims := &jwt.RegisteredClaims{
					Subject:   "invalid-user-id",
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					NotBefore: jwt.NewNumericDate(time.Now()),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenStr, _ := token.SignedString([]byte("test-access-secret-key-for-unit-testing-only"))
				return tokenStr
			}(),
			setupMocks: func(authRepo *MockAuthRepository) {
				// 不调用任何方法
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name:        "缺少环境变量",
			accessToken: validAccessToken,
			setupMocks: func(authRepo *MockAuthRepository) {
				// 在执行测试前删除环境变量
				os.Unsetenv("JWT_ACCESS_SECRET")
			},
			wantErr:     true,
			expectedErr: errors.New("JWT_ACCESS_SECRET environment variable is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 记录原始环境变量值
			originalSecret := os.Getenv("JWT_ACCESS_SECRET")

			// 创建 mock
			authRepo := new(MockAuthRepository)

			// 设置 mock 期望
			if tt.setupMocks != nil {
				tt.setupMocks(authRepo)
			}

			// 创建 usecase
			uc := NewAuthUsecase(authRepo, getTestLogger())

			// 执行测试
			userID, err := uc.ValidateToken(context.Background(), tt.accessToken)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
				assert.Equal(t, int64(0), userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUserID, userID)
			}

			// 恢复环境变量
			if originalSecret != "" {
				os.Setenv("JWT_ACCESS_SECRET", originalSecret)
			}

			// 验证所有期望都被调用
			authRepo.AssertExpectations(t)
		})
	}
}
