package data

import (
	"context"
	"fmt"
	"testing"
	"time"
	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDataRepository_StoreVerificationCode 测试存储验证码功能
func TestDataRepository_StoreVerificationCode(t *testing.T) {
	// 使用过去的时间作为过期时间，表示永不过期
	pastTime := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name        string
		email       string
		code        string
		setupMock   func(redismock.ClientMock)
		wantErr     bool
		expectedErr string
	}{
		{
			name:  "成功存储验证码",
			email: "test@example.com",
			code:  "123456",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test@example.com"
				// 使用0持续时间表示无过期
				mock.ExpectSet(key, "123456", time.Duration(0)).SetVal("OK")
			},
			wantErr: false,
		},
		{
			name:  "存储验证码失败 - Redis错误",
			email: "test@example.com",
			code:  "123456",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test@example.com"
				mock.ExpectSet(key, "123456", time.Duration(0)).SetErr(fmt.Errorf("redis connection error"))
			},
			wantErr:     true,
			expectedErr: "redis connection error",
		},
		{
			name:  "存储验证码 - 空邮箱",
			email: "",
			code:  "123456",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:"
				mock.ExpectSet(key, "123456", time.Duration(0)).SetVal("OK")
			},
			wantErr: false,
		},
		{
			name:  "存储验证码 - 特殊字符邮箱",
			email: "test+tag@example-domain.co.uk",
			code:  "123456",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test+tag@example-domain.co.uk"
				mock.ExpectSet(key, "123456", time.Duration(0)).SetVal("OK")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建Redis mock
			client, mock := redismock.NewClientMock()

			// 设置mock期望
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			// 创建Data结构体
			data := &Data{
				rds: client,
			}

			// 创建repository
			repo := NewCodeRepository(data, log.DefaultLogger)

			// 执行测试
			err := repo.StoreVerificationCode(context.Background(), tt.email, tt.code, pastTime)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			// 验证所有期望都被满足
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestDataRepository_GetVerificationCode 测试获取验证码功能
func TestDataRepository_GetVerificationCode(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(10 * time.Minute)

	tests := []struct {
		name        string
		email       string
		setupMock   func(redismock.ClientMock)
		wantCode    *biz.VerificationCode
		wantErr     bool
		expectedErr string
	}{
		{
			name:  "成功获取有效验证码",
			email: "test@example.com",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test@example.com"
				mock.ExpectGet(key).SetVal("123456")
				mock.ExpectTTL(key).SetVal(10 * time.Minute)
			},
			wantCode: &biz.VerificationCode{
				Email: "test@example.com",
				Code:  "123456",
			},
			wantErr: false,
		},
		{
			name:  "验证码不存在",
			email: "nonexistent@example.com",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:nonexistent@example.com"
				mock.ExpectGet(key).SetErr(redis.Nil)
			},
			wantCode:    nil,
			wantErr:     true,
			expectedErr: "验证码不存在或已过期",
		},
		{
			name:  "Redis连接错误",
			email: "test@example.com",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test@example.com"
				mock.ExpectGet(key).SetErr(fmt.Errorf("connection error"))
			},
			wantCode:    nil,
			wantErr:     true,
			expectedErr: "connection error",
		},
		{
			name:  "TTL获取错误",
			email: "test@example.com",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test@example.com"
				mock.ExpectGet(key).SetVal("123456")
				mock.ExpectTTL(key).SetErr(fmt.Errorf("ttl error"))
			},
			wantCode:    nil,
			wantErr:     true,
			expectedErr: "ttl error",
		},
		{
			name:  "空邮箱验证码",
			email: "",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:"
				mock.ExpectGet(key).SetVal("123456")
				mock.ExpectTTL(key).SetVal(10 * time.Minute)
			},
			wantCode: &biz.VerificationCode{
				Email: "",
				Code:  "123456",
			},
			wantErr: false,
		},
		{
			name:  "特殊字符邮箱验证码",
			email: "test+tag@example-domain.co.uk",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test+tag@example-domain.co.uk"
				mock.ExpectGet(key).SetVal("123456")
				mock.ExpectTTL(key).SetVal(10 * time.Minute)
			},
			wantCode: &biz.VerificationCode{
				Email: "test+tag@example-domain.co.uk",
				Code:  "123456",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建Redis mock
			client, mock := redismock.NewClientMock()

			// 设置mock期望
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			// 创建Data结构体
			data := &Data{
				rds: client,
			}

			// 创建repository
			repo := NewCodeRepository(data, log.DefaultLogger)

			// 执行测试
			code, err := repo.GetVerificationCode(context.Background(), tt.email)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
				assert.Nil(t, code)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, code)
				assert.Equal(t, tt.wantCode.Email, code.Email)
				assert.Equal(t, tt.wantCode.Code, code.Code)
				// 验证过期时间在预期范围内
				assert.WithinDuration(t, futureTime, code.ExpiresAt, time.Second)
			}

			// 验证所有期望都被满足
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestDataRepository_DeleteVerificationCode 测试删除验证码功能
func TestDataRepository_DeleteVerificationCode(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		setupMock   func(redismock.ClientMock)
		wantErr     bool
		expectedErr string
	}{
		{
			name:  "成功删除验证码",
			email: "test@example.com",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test@example.com"
				mock.ExpectDel(key).SetVal(1)
			},
			wantErr: false,
		},
		{
			name:  "删除不存在的验证码",
			email: "nonexistent@example.com",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:nonexistent@example.com"
				mock.ExpectDel(key).SetVal(0)
			},
			wantErr: false,
		},
		{
			name:  "Redis连接错误",
			email: "test@example.com",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test@example.com"
				mock.ExpectDel(key).SetErr(fmt.Errorf("connection error"))
			},
			wantErr:     true,
			expectedErr: "connection error",
		},
		{
			name:  "删除空邮箱验证码",
			email: "",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:"
				mock.ExpectDel(key).SetVal(1)
			},
			wantErr: false,
		},
		{
			name:  "删除特殊字符邮箱验证码",
			email: "test+tag@example-domain.co.uk",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test+tag@example-domain.co.uk"
				mock.ExpectDel(key).SetVal(1)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建Redis mock
			client, mock := redismock.NewClientMock()

			// 设置mock期望
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			// 创建Data结构体
			data := &Data{
				rds: client,
			}

			// 创建repository
			repo := NewCodeRepository(data, log.DefaultLogger)

			// 执行测试
			err := repo.DeleteVerificationCode(context.Background(), tt.email)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			// 验证所有期望都被满足
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestDataRepository_Integration 测试完整的验证码生命周期
func TestDataRepository_Integration(t *testing.T) {
	client, mock := redismock.NewClientMock()

	// 创建Data结构体
	data := &Data{
		rds: client,
	}

	// 创建repository
	repo := NewCodeRepository(data, log.DefaultLogger)

	email := "integration@example.com"
	code := "654321"
	// 使用过去的时间表示永不过期
	pastTime := time.Now().Add(-1 * time.Hour)
	key := "verification_code:integration@example.com"

	// 1. 存储验证码
	mock.ExpectSet(key, code, time.Duration(0)).SetVal("OK")
	err := repo.StoreVerificationCode(context.Background(), email, code, pastTime)
	assert.NoError(t, err)

	// 2. 获取验证码
	mock.ExpectGet(key).SetVal(code)
	mock.ExpectTTL(key).SetVal(-1 * time.Second) // -1表示永不过期
	storedCode, err := repo.GetVerificationCode(context.Background(), email)
	assert.NoError(t, err)
	assert.Equal(t, email, storedCode.Email)
	assert.Equal(t, code, storedCode.Code)

	// 3. 删除验证码
	mock.ExpectDel(key).SetVal(1)
	err = repo.DeleteVerificationCode(context.Background(), email)
	assert.NoError(t, err)

	// 验证所有期望都被满足
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDataRepository_ConcurrentAccess 测试并发访问
func TestDataRepository_ConcurrentAccess(t *testing.T) {
	client, mock := redismock.NewClientMock()

	// 创建Data结构体
	data := &Data{
		rds: client,
	}

	// 创建repository
	repo := NewCodeRepository(data, log.DefaultLogger)

	email := "concurrent@example.com"
	code := "999999"
	// 使用过去的时间表示永不过期
	pastTime := time.Now().Add(-1 * time.Hour)
	key := "verification_code:concurrent@example.com"

	// 设置多个并发的存储操作期望
	for i := 0; i < 5; i++ {
		mock.ExpectSet(key, code, time.Duration(0)).SetVal("OK")
	}

	// 并发存储验证码
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			err := repo.StoreVerificationCode(context.Background(), email, code, pastTime)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 5; i++ {
		<-done
	}

	// 验证所有期望都被满足
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestDataRepository_EdgeCases 测试边界条件
func TestDataRepository_EdgeCases(t *testing.T) {
	// 使用过去的时间表示永不过期
	pastTime := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name        string
		email       string
		code        string
		setupMock   func(redismock.ClientMock)
		wantErr     bool
		expectedErr string
	}{
		{
			name:  "极长验证码",
			email: "test@example.com",
			code:  string(make([]byte, 100)), // 减少长度避免问题
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test@example.com"
				longCode := string(make([]byte, 100))
				mock.ExpectSet(key, longCode, time.Duration(0)).SetVal("OK")
			},
			wantErr: false,
		},
		{
			name:  "Unicode字符验证码",
			email: "test@example.com",
			code:  "验证码123",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test@example.com"
				mock.ExpectSet(key, "验证码123", time.Duration(0)).SetVal("OK")
			},
			wantErr: false,
		},
		{
			name:  "空验证码",
			email: "test@example.com",
			code:  "",
			setupMock: func(mock redismock.ClientMock) {
				key := "verification_code:test@example.com"
				mock.ExpectSet(key, "", time.Duration(0)).SetVal("OK")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建Redis mock
			client, mock := redismock.NewClientMock()

			// 设置mock期望
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			// 创建Data结构体
			data := &Data{
				rds: client,
			}

			// 创建repository
			repo := NewCodeRepository(data, log.DefaultLogger)

			// 执行测试
			err := repo.StoreVerificationCode(context.Background(), tt.email, tt.code, pastTime)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			// 验证所有期望都被满足
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
