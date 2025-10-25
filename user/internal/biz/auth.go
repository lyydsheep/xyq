package biz

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
)

// Auth Errors
var (
	// ErrInvalidToken 当令牌无效时返回
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired 当令牌过期时返回
	ErrTokenExpired = errors.New("token expired")
)

// TokenPair 令牌对，包含访问令牌和刷新令牌
type TokenPair struct {
	AccessToken      string
	AccessExpiresIn  int32
	RefreshToken     string
	RefreshExpiresIn int32
}

// AuthRepository 认证数据访问接口，定义了令牌相关的数据操作方法
type AuthRepository interface {
	// Token相关操作
	StoreRefreshToken(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error
	GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (int64, error)
	DeleteRefreshToken(ctx context.Context, refreshToken string) error
	DeleteAllRefreshTokens(ctx context.Context, userID int64) error
	// 事务方法
	RefreshTokenAtomically(ctx context.Context, userID int64, oldToken, newToken string, expiresAt time.Time) error
}

// AuthUsecase 认证业务逻辑，处理用户注册、登录、令牌刷新等认证相关操作
type AuthUsecase struct {
	authRepo AuthRepository // 认证数据访问接口
	log      *log.Helper    // 日志助手
}

// NewAuthUsecase 创建认证业务逻辑实例
//
// 参数:
//   - userRepo: 用户数据访问接口
//   - authRepo: 认证数据访问接口
//   - logger: 日志记录器
//
// 返回值:
//   - *AuthUsecase: 认证业务逻辑实例
func NewAuthUsecase(authRepo AuthRepository, logger log.Logger) *AuthUsecase {
	return &AuthUsecase{
		authRepo: authRepo,
		log:      log.NewHelper(logger),
	}
}

// generateAccessToken 生成访问令牌（JWT）
func generateAccessToken(userID int64) (string, int32, error) {
	// 设置过期时间为1小时
	expiresIn := int32(3600)
	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second)

	// 创建声明
	claims := &jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并获得完整的编码后的字符串token
	// 注意：在生产环境中，应该使用环境变量或配置文件来存储密钥
	// TODO 补充密钥
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresIn, nil
}

// generateRefreshToken 生成刷新令牌（JWT）
func generateRefreshToken(userID int64) (string, int32, error) {
	// 设置过期时间为7天
	expiresIn := int32(7 * 24 * 3600)
	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second)

	// 创建声明
	claims := &jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		ID:        "refresh_" + string(rune(userID)),
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并获得完整的编码后的字符串token
	// 注意：在生产环境中，应该使用环境变量或配置文件来存储密钥
	tokenString, err := token.SignedString([]byte("your-refresh-secret-key"))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresIn, nil
}

// RefreshToken 刷新访问令牌
func (uc *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	uc.log.Log(log.LevelInfo, "Refreshing token")

	// 参数验证
	if refreshToken == "" {
		uc.log.Log(log.LevelWarn, "Empty refresh token provided")
		return nil, ErrInvalidToken
	}

	// 验证刷新令牌
	userID, err := uc.authRepo.GetUserIDByRefreshToken(ctx, refreshToken)
	if err != nil {
		uc.log.Log(log.LevelWarn, "Invalid refresh token provided")
		return nil, ErrInvalidToken
	}

	// 使用事务确保令牌刷新的原子性
	return uc.refreshTokenInTransaction(ctx, userID, refreshToken)
}

// refreshTokenInTransaction 在事务中刷新令牌
func (uc *AuthUsecase) refreshTokenInTransaction(ctx context.Context, userID int64, oldRefreshToken string) (*TokenPair, error) {
	// 生成新的令牌对
	accessToken, accessExpiresIn, err := generateAccessToken(userID)
	if err != nil {
		uc.log.Log(log.LevelError, "Failed to generate access token during refresh for user id: ", userID, ", error: ", err)
		return nil, err
	}

	newRefreshToken, refreshExpiresIn, err := generateRefreshToken(userID)
	if err != nil {
		uc.log.Log(log.LevelError, "Failed to generate refresh token during refresh for user id: ", userID, ", error: ", err)
		return nil, err
	}

	// 使用原子操作刷新令牌
	refreshTokenExpiresAt := time.Now().Add(time.Duration(refreshExpiresIn) * time.Second)
	err = uc.authRepo.RefreshTokenAtomically(ctx, userID, oldRefreshToken, newRefreshToken, refreshTokenExpiresAt)
	if err != nil {
		uc.log.Log(log.LevelError, "Failed to refresh token atomically for user id: ", userID, ", error: ", err)
		return nil, err
	}

	uc.log.Log(log.LevelInfo, "Token refresh successful for user id: ", userID)
	return &TokenPair{
		AccessToken:      accessToken,
		AccessExpiresIn:  accessExpiresIn,
		RefreshToken:     newRefreshToken,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}

// Logout 用户登出
func (uc *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	uc.log.Log(log.LevelInfo, "User logout")

	// 参数验证
	if refreshToken == "" {
		uc.log.Log(log.LevelWarn, "Empty refresh token provided for logout")
		return errors.New("refresh token is required")
	}

	// 删除刷新令牌
	err := uc.authRepo.DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		uc.log.Log(log.LevelError, "Failed to delete refresh token during logout, error: ", err)
		return err
	}

	uc.log.Log(log.LevelInfo, "User logout successful")
	return nil
}

// ValidateToken 验证访问令牌（JWT版本）
func (uc *AuthUsecase) ValidateToken(ctx context.Context, accessToken string) (int64, error) {
	// 参数验证
	if accessToken == "" {
		uc.log.Log(log.LevelWarn, "Empty access token provided for validation")
		return 0, ErrInvalidToken
	}

	// 解析和验证JWT令牌
	token, err := jwt.ParseWithClaims(accessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil
	})

	if err != nil {
		uc.log.Log(log.LevelWarn, "Failed to parse access token, error: ", err)
		return 0, ErrInvalidToken
	}

	// 验证令牌是否有效
	if !token.Valid {
		uc.log.Log(log.LevelWarn, "Invalid access token provided")
		return 0, ErrInvalidToken
	}

	// 获取声明
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		// 检查是否过期
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			uc.log.Log(log.LevelWarn, "Access token has expired")
			return 0, ErrTokenExpired
		}

		// 解析用户ID
		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			uc.log.Log(log.LevelWarn, "Failed to parse user id from access token")
			return 0, ErrInvalidToken
		}
		uc.log.Log(log.LevelInfo, "Token validation successful for user id: ", userID)
		return userID, nil
	} else {
		uc.log.Log(log.LevelWarn, "Failed to get claims from access token")
		return 0, ErrInvalidToken
	}
}
