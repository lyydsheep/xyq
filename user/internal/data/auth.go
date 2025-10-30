package data

import (
	"context"
	"fmt"
	"time"

	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
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
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	// 将用户ID作为值存储在Redis中
	err := r.data.RedisClient().Set(ctx, key, userID, time.Until(expiresAt)).Err()
	if err != nil {
		r.logger.Errorf("Failed to store refresh token: %v", err)
		return err
	}
	return nil
}

// GetUserIDByRefreshToken 根据刷新令牌获取用户ID
func (r *authRepository) GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (int64, error) {
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	val, err := r.data.RedisClient().Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("refresh token not found")
		}
		r.logger.Errorf("Failed to get refresh token: %v", err)
		return 0, err
	}
	return val, nil
}

// DeleteRefreshToken 删除刷新令牌
func (r *authRepository) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	err := r.data.RedisClient().Del(ctx, key).Err()
	if err != nil {
		r.logger.Errorf("Failed to delete refresh token: %v", err)
		return err
	}
	return nil
}

// DeleteAllRefreshTokens 删除用户的所有刷新令牌
func (r *authRepository) DeleteAllRefreshTokens(ctx context.Context, userID int64) error {
	// 查找所有以 "refresh_token:" 开头且用户ID匹配的键
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
		r.logger.Errorf("Failed to scan refresh tokens: %v", err)
		return err
	}

	if len(keys) > 0 {
		err := r.data.RedisClient().Del(ctx, keys...).Err()
		if err != nil {
			r.logger.Errorf("Failed to delete refresh tokens: %v", err)
			return err
		}
	}
	return nil
}

// RefreshTokenAtomically 原子性地刷新令牌
func (r *authRepository) RefreshTokenAtomically(ctx context.Context, userID int64, oldToken, newToken string, expiresAt time.Time) error {
	// 使用 Redis 事务确保原子性
	pipe := r.data.RedisClient().Pipeline()

	// 删除旧令牌
	oldKey := fmt.Sprintf("refresh_token:%s", oldToken)
	pipe.Del(ctx, oldKey)

	// 存储新令牌
	newKey := fmt.Sprintf("refresh_token:%s", newToken)
	pipe.Set(ctx, newKey, userID, time.Until(expiresAt))

	// 执行事务
	_, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.Errorf("Failed to refresh token atomically: %v", err)
		return err
	}
	return nil
}
