package data

import (
	"context"
	"fmt"
	"time"
	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
)

// RedisClient 定义Redis客户端接口，方便测试
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
}

// codeRepository 验证码数据访问实现
type codeRepository struct {
	data   *Data
	logger *log.Helper
}

// NewCodeRepository 创建验证码数据访问实例
func NewCodeRepository(data *Data, logger log.Logger) biz.CodeRepository {
	return &codeRepository{
		data:   data,
		logger: log.NewHelper(logger),
	}
}

// StoreVerificationCode 存储验证码到Redis
func (r *codeRepository) StoreVerificationCode(ctx context.Context, email, code string, expiresAt time.Time) error {
	r.logger.Log(log.LevelInfo, "Storing verification code for email: ", email)

	// 构造Redis键
	key := fmt.Sprintf("verification_code:%s", email)

	// 设置过期时间
	expiration := time.Until(expiresAt)
	r.logger.Log(log.LevelWarn, "expiration", expiration)

	// 存储验证码到Redis
	err := r.data.RedisClient().Set(ctx, key, code, expiration).Err()
	if err != nil {
		r.logger.Log(log.LevelError, "Failed to store verification code for email: ", email, ", error: ", err)
		return err
	}

	r.logger.Log(log.LevelInfo, "Successfully stored verification code for email: ", email)
	return nil
}

// GetVerificationCode 从Redis获取验证码
func (r *codeRepository) GetVerificationCode(ctx context.Context, email string) (*biz.VerificationCode, error) {
	r.logger.Log(log.LevelInfo, "Getting verification code for email: ", email)

	// 构造Redis键
	key := fmt.Sprintf("verification_code:%s", email)

	// 从Redis获取验证码
	code, err := r.data.RedisClient().Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.logger.Log(log.LevelWarn, "Verification code not found or expired for email: ", email)
			return nil, fmt.Errorf("验证码不存在或已过期")
		}
		r.logger.Log(log.LevelError, "Failed to get verification code for email: ", email, ", error: ", err)
		return nil, err
	}

	// 获取TTL以计算过期时间
	ttl, err := r.data.RedisClient().TTL(ctx, key).Result()
	if err != nil {
		r.logger.Log(log.LevelError, "Failed to get TTL for verification code of email: ", email, ", error: ", err)
		return nil, err
	}

	verificationCode := &biz.VerificationCode{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(ttl),
	}

	r.logger.Log(log.LevelInfo, "Successfully retrieved verification code for email: ", email)
	return verificationCode, nil
}

// DeleteVerificationCode 从Redis删除验证码
func (r *codeRepository) DeleteVerificationCode(ctx context.Context, email string) error {
	r.logger.Log(log.LevelInfo, "Deleting verification code for email: ", email)

	// 构造Redis键
	key := fmt.Sprintf("verification_code:%s", email)

	// 从Redis删除验证码
	_, err := r.data.RedisClient().Del(ctx, key).Result()
	if err != nil {
		r.logger.Log(log.LevelError, "Failed to delete verification code for email: ", email, ", error: ", err)
		return err
	}

	r.logger.Log(log.LevelInfo, "Successfully deleted verification code for email: ", email)
	return nil
}
