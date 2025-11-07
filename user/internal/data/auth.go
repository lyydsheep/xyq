package data

import (
	"context"
	"fmt"
	"time"

	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"user/internal/pkg/tracing"
)

// authRepository 实现 biz.AuthRepository 接口
type authRepository struct {
	data   *Data
	logger *log.Helper
}

// NewAuthRepository 创建 AuthRepository 实例
func NewAuthRepository(data *Data, logger log.Logger) biz.AuthRepository {
	return &authRepository{
		data:   data,
		logger: log.NewHelper(logger),
	}
}

// StoreRefreshToken 存储刷新令牌
func (r *authRepository) StoreRefreshToken(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error {
	ctx, span := tracing.StartSpan(ctx, "AuthRepository.StoreRefreshToken")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"user_id":      userID,
		"token_length": len(refreshToken),
	})

	r.logger.WithContext(ctx).Infof("Storing refresh token for user_id: %d", userID)

	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	err := r.data.RedisClient().Set(ctx, key, userID, time.Until(expiresAt)).Err()
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to store refresh token for user_id: %d, error_reason: %v", userID, err)
		return err
	}

	r.logger.WithContext(ctx).Infof("Successfully stored refresh token for user_id: %d", userID)
	return nil
}

// GetUserIDByRefreshToken 根据刷新令牌获取用户ID
func (r *authRepository) GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (int64, error) {
	ctx, span := tracing.StartSpan(ctx, "AuthRepository.GetUserIDByRefreshToken")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"token_length": len(refreshToken),
	})

	r.logger.WithContext(ctx).Info("Getting user ID by refresh token")

	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	val, err := r.data.RedisClient().Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			r.logger.WithContext(ctx).Warn("Refresh token not found")
			return 0, fmt.Errorf("refresh token not found")
		}
		r.logger.WithContext(ctx).Errorf("Failed to get refresh token, error_reason: %v", err)
		return 0, err
	}

	r.logger.WithContext(ctx).Infof("Successfully retrieved user ID: %d by refresh token", val)
	return val, nil
}

// DeleteRefreshToken 删除刷新令牌
func (r *authRepository) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	ctx, span := tracing.StartSpan(ctx, "AuthRepository.DeleteRefreshToken")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"token_length": len(refreshToken),
	})

	r.logger.WithContext(ctx).Info("Deleting refresh token")

	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	err := r.data.RedisClient().Del(ctx, key).Err()
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to delete refresh token, error_reason: %v", err)
		return err
	}

	r.logger.WithContext(ctx).Info("Successfully deleted refresh token")
	return nil
}

// DeleteAllRefreshTokens 删除用户的所有刷新令牌
func (r *authRepository) DeleteAllRefreshTokens(ctx context.Context, userID int64) error {
	ctx, span := tracing.StartSpan(ctx, "AuthRepository.DeleteAllRefreshTokens")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"user_id": userID,
	})

	r.logger.WithContext(ctx).Infof("Deleting all refresh tokens for user_id: %d", userID)

	pattern := "refresh_token:*"
	iter := r.data.RedisClient().Scan(ctx, 0, pattern, -1).Iterator()
	var keys []string
	for iter.Next(ctx) {
		key := iter.Val()
		val, err := r.data.RedisClient().Get(ctx, key).Int64()
		if err == nil && val == userID {
			keys = append(keys, key)
		}
	}
	if err := iter.Err(); err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to scan refresh tokens for user_id: %d, error_reason: %v", userID, err)
		return err
	}

	if len(keys) > 0 {
		err := r.data.RedisClient().Del(ctx, keys...).Err()
		if err != nil {
			r.logger.WithContext(ctx).Errorf("Failed to delete refresh tokens for user_id: %d, error_reason: %v", userID, err)
			return err
		}
		r.logger.WithContext(ctx).Infof("Successfully deleted %d refresh tokens for user_id: %d", len(keys), userID)
	} else {
		r.logger.WithContext(ctx).Infof("No refresh tokens found to delete for user_id: %d", userID)
	}

	return nil
}

// RefreshTokenAtomically 原子性地刷新令牌
func (r *authRepository) RefreshTokenAtomically(ctx context.Context, userID int64, oldToken, newToken string, expiresAt time.Time) error {
	ctx, span := tracing.StartSpan(ctx, "AuthRepository.RefreshTokenAtomically")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"user_id":          userID,
		"old_token_length": len(oldToken),
		"new_token_length": len(newToken),
	})

	r.logger.WithContext(ctx).Infof("Atomically refreshing token for user_id: %d", userID)

	pipe := r.data.RedisClient().Pipeline()

	oldKey := fmt.Sprintf("refresh_token:%s", oldToken)
	pipe.Del(ctx, oldKey)

	newKey := fmt.Sprintf("refresh_token:%s", newToken)
	pipe.Set(ctx, newKey, userID, time.Until(expiresAt))

	_, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to refresh token atomically for user_id: %d, error_reason: %v", userID, err)
		return err
	}

	r.logger.WithContext(ctx).Infof("Successfully refreshed token atomically for user_id: %d", userID)
	tracing.AddSpanEvent(ctx, "token_atomic_refresh_success", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}
