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
	ctx, span := tracing.StartSpan(ctx, "CodeRepository.StoreVerificationCode")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"email":       email,
		"code_length": len(code),
	})

	r.logger.WithContext(ctx).Infof("Storing verification code for email: %s", email)

	key := fmt.Sprintf("verification_code:%s", email)
	expiration := time.Until(expiresAt)

	err := r.data.RedisClient().Set(ctx, key, code, expiration).Err()
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to store verification code for email: %s, error_reason: %v", email, err)
		return err
	}

	r.logger.WithContext(ctx).Infof("Successfully stored verification code for email: %s", email)
	return nil
}

// GetVerificationCode 从Redis获取验证码
func (r *codeRepository) GetVerificationCode(ctx context.Context, email string) (*biz.VerificationCode, error) {
	ctx, span := tracing.StartSpan(ctx, "CodeRepository.GetVerificationCode")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"email": email,
	})

	r.logger.WithContext(ctx).Infof("Getting verification code for email: %s", email)

	key := fmt.Sprintf("verification_code:%s", email)
	code, err := r.data.RedisClient().Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.logger.WithContext(ctx).Warnf("Verification code not found or expired for email: %s", email)
			return nil, fmt.Errorf("验证码不存在或已过期")
		}
		r.logger.WithContext(ctx).Errorf("Failed to get verification code for email: %s, error_reason: %v", email, err)
		return nil, err
	}

	ttl, err := r.data.RedisClient().TTL(ctx, key).Result()
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to get TTL for verification code of email: %s, error_reason: %v", email, err)
		return nil, err
	}

	verificationCode := &biz.VerificationCode{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(ttl),
	}

	r.logger.WithContext(ctx).Infof("Successfully retrieved verification code for email: %s", email)
	return verificationCode, nil
}

// DeleteVerificationCode 从Redis删除验证码
func (r *codeRepository) DeleteVerificationCode(ctx context.Context, email string) error {
	ctx, span := tracing.StartSpan(ctx, "CodeRepository.DeleteVerificationCode")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"email": email,
	})

	r.logger.WithContext(ctx).Infof("Deleting verification code for email: %s", email)

	key := fmt.Sprintf("verification_code:%s", email)
	_, err := r.data.RedisClient().Del(ctx, key).Result()
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to delete verification code for email: %s, error_reason: %v", email, err)
		return err
	}

	r.logger.WithContext(ctx).Infof("Successfully deleted verification code for email: %s", email)
	return nil
}

// CheckAndSetSendRateLimit 检查并设置发送频率限制
// 如果在指定时间内已经发送过验证码，返回 false；否则设置限制并返回 true
func (r *codeRepository) CheckAndSetSendRateLimit(ctx context.Context, email string, duration time.Duration) (bool, error) {
	ctx, span := tracing.StartSpan(ctx, "CodeRepository.CheckAndSetSendRateLimit")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"email":            email,
		"duration_seconds": duration.Seconds(),
	})

	r.logger.WithContext(ctx).Infof("Checking send rate limit for email: %s", email)

	key := fmt.Sprintf("rate_limit:send_code:%s", email)
	// SetNX 返回一个 bool 值表示是否成功设置，我们需要检查这个值
	success, err := r.data.RedisClient().SetNX(ctx, key, time.Now().Unix(), duration).Result()
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to set rate limit for email: %s, error_reason: %v", email, err)
		return false, err
	}

	if !success {
		r.logger.WithContext(ctx).Warnf("Rate limit exceeded for email: %s", email)
		return false, nil
	}

	r.logger.WithContext(ctx).Infof("Rate limit set successfully for email: %s", email)
	return true, nil
}
