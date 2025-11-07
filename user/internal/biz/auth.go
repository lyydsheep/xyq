package biz

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"strconv"
	"time"
	error_reason "user/api/error_reason"
	"user/internal/pkg/tracing"
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

	// 从环境变量获取JWT访问令牌密钥
	secret := os.Getenv("JWT_ACCESS_SECRET")
	if secret == "" {
		return "", 0, error_reason.ErrorAuthDatabaseError("JWT访问令牌密钥未配置")
	}

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
	tokenString, err := token.SignedString([]byte(secret))
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

	// 从环境变量获取JWT刷新令牌密钥
	secret := os.Getenv("JWT_REFRESH_SECRET")
	if secret == "" {
		return "", 0, error_reason.ErrorAuthDatabaseError("JWT刷新令牌密钥未配置")
	}

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
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresIn, nil
}

// RefreshToken 刷新访问令牌
func (uc *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	ctx, span := tracing.StartSpan(ctx, "AuthUsecase.RefreshToken")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation":    "refresh_token",
		"token_length": len(refreshToken),
	})

	uc.log.WithContext(ctx).Info("Refreshing token")

	// 参数验证
	if refreshToken == "" {
		uc.log.WithContext(ctx).Warn("Empty refresh token provided")
		return nil, error_reason.ErrorUserRefreshTokenInvalid("刷新令牌不能为空")
	}

	// 验证刷新令牌
	userID, err := uc.authRepo.GetUserIDByRefreshToken(ctx, refreshToken)
	if err != nil {
		uc.log.WithContext(ctx).Warn("Invalid refresh token provided")
		return nil, error_reason.ErrorUserRefreshTokenInvalid("刷新令牌无效")
	}

	// 使用事务确保令牌刷新的原子性
	return uc.refreshTokenInTransaction(ctx, userID, refreshToken)
}

// refreshTokenInTransaction 在事务中刷新令牌
func (uc *AuthUsecase) refreshTokenInTransaction(ctx context.Context, userID int64, oldRefreshToken string) (*TokenPair, error) {
	// 生成新的令牌对
	accessToken, accessExpiresIn, err := generateAccessToken(userID)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to generate access token during refresh for user id: %d, error_reason: %v", userID, err)
		return nil, error_reason.ErrorUserInternalError("访问令牌生成失败")
	}

	newRefreshToken, refreshExpiresIn, err := generateRefreshToken(userID)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to generate refresh token during refresh for user id: %d, error_reason: %v", userID, err)
		return nil, error_reason.ErrorUserInternalError("刷新令牌生成失败")
	}

	// 使用原子操作刷新令牌
	refreshTokenExpiresAt := time.Now().Add(time.Duration(refreshExpiresIn) * time.Second)
	err = uc.authRepo.RefreshTokenAtomically(ctx, userID, oldRefreshToken, newRefreshToken, refreshTokenExpiresAt)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to refresh token atomically for user id: %d, error_reason: %v", userID, err)
		return nil, error_reason.ErrorUserDatabaseError("令牌刷新失败")
	}

	uc.log.WithContext(ctx).Infof("Token refresh successful for user id: %d", userID)
	tracing.AddSpanEvent(ctx, "token_refresh_success", map[string]interface{}{
		"user_id":            userID,
		"access_expires_in":  accessExpiresIn,
		"refresh_expires_in": refreshExpiresIn,
	})

	return &TokenPair{
		AccessToken:      accessToken,
		AccessExpiresIn:  accessExpiresIn,
		RefreshToken:     newRefreshToken,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}

// Logout 用户登出
func (uc *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	ctx, span := tracing.StartSpan(ctx, "AuthUsecase.Logout")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation":    "logout",
		"token_length": len(refreshToken),
	})

	uc.log.WithContext(ctx).Info("User logout")

	// 参数验证
	if refreshToken == "" {
		uc.log.WithContext(ctx).Warn("Empty refresh token provided for logout")
		return error_reason.ErrorUserRefreshTokenInvalid("刷新令牌不能为空")
	}

	// 删除刷新令牌
	err := uc.authRepo.DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to delete refresh token during logout, error_reason: %v", err)
		return error_reason.ErrorUserDatabaseError("令牌删除失败")
	}

	uc.log.WithContext(ctx).Info("User logout successful")
	return nil
}

// ValidateToken 验证访问令牌（JWT版本）
func (uc *AuthUsecase) ValidateToken(ctx context.Context, accessToken string) (int64, error) {
	ctx, span := tracing.StartSpan(ctx, "AuthUsecase.ValidateToken")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation":    "validate_token",
		"token_length": len(accessToken),
	})

	// 参数验证
	if accessToken == "" {
		uc.log.WithContext(ctx).Warn("Empty access token provided for validation")
		return 0, error_reason.ErrorUserInvalidToken("访问令牌不能为空")
	}

	// 从环境变量获取JWT访问令牌密钥
	secret := os.Getenv("JWT_ACCESS_SECRET")
	if secret == "" {
		uc.log.WithContext(ctx).Error("JWT_ACCESS_SECRET environment variable is required")
		return 0, error_reason.ErrorAuthDatabaseError("JWT访问令牌密钥未配置")
	}

	// 解析和验证JWT令牌
	token, err := jwt.ParseWithClaims(accessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		uc.log.WithContext(ctx).Warnf("Failed to parse access token, error_reason: %v", err)
		return 0, error_reason.ErrorUserInvalidToken("访问令牌格式无效")
	}

	// 验证令牌是否有效
	if !token.Valid {
		uc.log.WithContext(ctx).Warn("Invalid access token provided")
		return 0, error_reason.ErrorUserInvalidToken("访问令牌无效")
	}

	// 获取声明
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		// 检查是否过期
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			uc.log.WithContext(ctx).Warn("Access token has expired")
			return 0, error_reason.ErrorUserTokenExpired("访问令牌已过期")
		}

		// 解析用户ID
		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			uc.log.WithContext(ctx).Warn("Failed to parse user id from access token")
			return 0, error_reason.ErrorUserInvalidToken("访问令牌用户信息无效")
		}
		uc.log.WithContext(ctx).Infof("Token validation successful for user id: %d", userID)
		return userID, nil
	} else {
		uc.log.WithContext(ctx).Warn("Failed to get claims from access token")
		return 0, error_reason.ErrorUserInvalidToken("访问令牌格式无效")
	}
}
