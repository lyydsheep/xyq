package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

// TestAuthRepository_StoreRefreshToken 测试存储刷新令牌
func TestAuthRepository_StoreRefreshToken(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		token     string
		expiresAt time.Time
		mockFn    func(mock redismock.ClientMock)
		wantErr   bool
	}{
		{
			name:      "成功存储刷新令牌",
			userID:    1,
			token:     "refresh_token_123456",
			expiresAt: time.Now().Add(7 * 24 * time.Hour),
			mockFn: func(mock redismock.ClientMock) {
				key := fmt.Sprintf("refresh_token:%s", "refresh_token_123456")
				mock.ExpectSet(key, int64(1), time.Until(time.Now().Add(7*24*time.Hour))).SetVal("OK")
			},
			wantErr: false,
		},
		{
			name:      "存储刷新令牌失败",
			userID:    2,
			token:     "invalid_token",
			expiresAt: time.Now().Add(24 * time.Hour),
			mockFn: func(mock redismock.ClientMock) {
				key := fmt.Sprintf("refresh_token:%s", "invalid_token")
				mock.ExpectSet(key, int64(2), time.Until(time.Now().Add(24*time.Hour))).RedisNil()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 Redis mock
			rds, mock := redismock.NewClientMock()

			// 创建 Data 结构体
			data := &Data{
				rds: rds,
				db:  nil, // 测试不需要数据库连接
			}

			// 创建 auth repository
			repo := NewAuthRepository(data, log.DefaultLogger)

			// 设置 mock 期望
			tt.mockFn(mock)

			err := repo.StoreRefreshToken(context.Background(), tt.userID, tt.token, tt.expiresAt)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 验证所有期望都被调用
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestAuthRepository_GetUserIDByRefreshToken 测试根据刷新令牌获取用户ID
func TestAuthRepository_GetUserIDByRefreshToken(t *testing.T) {
	tests := []struct {
		name         string
		token        string
		mockFn       func(mock redismock.ClientMock)
		expectedID   int64
		wantErr      bool
		expectErrMsg string
	}{
		{
			name:  "成功获取用户ID",
			token: "valid_token",
			mockFn: func(mock redismock.ClientMock) {
				key := fmt.Sprintf("refresh_token:%s", "valid_token")
				mock.ExpectGet(key).SetVal("123")
			},
			expectedID: 123,
			wantErr:    false,
		},
		{
			name:  "刷新令牌不存在",
			token: "nonexistent_token",
			mockFn: func(mock redismock.ClientMock) {
				key := fmt.Sprintf("refresh_token:%s", "nonexistent_token")
				mock.ExpectGet(key).RedisNil()
			},
			expectedID:   0,
			wantErr:      true,
			expectErrMsg: "refresh token not found",
		},
		{
			name:  "Redis返回错误",
			token: "error_token",
			mockFn: func(mock redismock.ClientMock) {
				key := fmt.Sprintf("refresh_token:%s", "error_token")
				mock.ExpectGet(key).SetErr(assert.AnError)
			},
			expectedID: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 Redis mock
			rds, mock := redismock.NewClientMock()

			// 创建 Data 结构体
			data := &Data{
				rds: rds,
				db:  nil,
			}

			// 创建 auth repository
			repo := NewAuthRepository(data, log.DefaultLogger)

			// 设置 mock 期望
			tt.mockFn(mock)

			userID, err := repo.GetUserIDByRefreshToken(context.Background(), tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectErrMsg)
				}
				assert.Equal(t, int64(0), userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}

			// 验证所有期望都被调用
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestAuthRepository_DeleteRefreshToken 测试删除刷新令牌
func TestAuthRepository_DeleteRefreshToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		mockFn  func(mock redismock.ClientMock)
		wantErr bool
	}{
		{
			name:  "成功删除刷新令牌",
			token: "valid_token",
			mockFn: func(mock redismock.ClientMock) {
				key := fmt.Sprintf("refresh_token:%s", "valid_token")
				mock.ExpectDel(key).SetVal(1)
			},
			wantErr: false,
		},
		{
			name:  "令牌不存在（删除返回0）",
			token: "nonexistent_token",
			mockFn: func(mock redismock.ClientMock) {
				key := fmt.Sprintf("refresh_token:%s", "nonexistent_token")
				mock.ExpectDel(key).SetVal(0)
			},
			wantErr: false, // 删除不存在的令牌不算是错误
		},
		{
			name:  "删除令牌失败",
			token: "error_token",
			mockFn: func(mock redismock.ClientMock) {
				key := fmt.Sprintf("refresh_token:%s", "error_token")
				mock.ExpectDel(key).RedisNil()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 Redis mock
			rds, mock := redismock.NewClientMock()

			// 创建 Data 结构体
			data := &Data{
				rds: rds,
				db:  nil,
			}

			// 创建 auth repository
			repo := NewAuthRepository(data, log.DefaultLogger)

			// 设置 mock 期望
			tt.mockFn(mock)

			err := repo.DeleteRefreshToken(context.Background(), tt.token)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 验证所有期望都被调用
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestAuthRepository_DeleteAllRefreshTokens 测试删除用户的所有刷新令牌
func TestAuthRepository_DeleteAllRefreshTokens(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		mockFn  func(mock redismock.ClientMock)
		wantErr bool
	}{
		{
			name:   "成功删除用户的所有刷新令牌",
			userID: 123,
			mockFn: func(mock redismock.ClientMock) {
				// 模拟 SCAN 操作找到匹配的键
				keys := []string{"refresh_token:token1", "refresh_token:token2", "refresh_token:token3"}
				mock.ExpectScan(0, "refresh_token:*", -1).SetVal(keys, 0)

				// 模拟 GET 操作获取每个键对应的用户ID
				mock.ExpectGet("refresh_token:token1").SetVal("123")
				mock.ExpectGet("refresh_token:token2").SetVal("123")
				mock.ExpectGet("refresh_token:token3").SetVal("123")

				// 模拟 DEL 操作删除所有匹配的键
				mock.ExpectDel("refresh_token:token1", "refresh_token:token2", "refresh_token:token3").SetVal(3)
			},
			wantErr: false,
		},
		{
			name:   "用户没有刷新令牌",
			userID: 999,
			mockFn: func(mock redismock.ClientMock) {
				// 模拟 SCAN 操作返回空结果
				keys := []string{}
				mock.ExpectScan(0, "refresh_token:*", -1).SetVal(keys, 0)
				// 没有 GET 和 DEL 操作
			},
			wantErr: false,
		},
		{
			name:   "SCAN操作出错",
			userID: 456,
			mockFn: func(mock redismock.ClientMock) {
				// 模拟 SCAN 操作出错
				mock.ExpectScan(0, "refresh_token:*", -1).SetErr(assert.AnError)
			},
			wantErr: true,
		},
		{
			name:   "GET操作出错但继续处理",
			userID: 789,
			mockFn: func(mock redismock.ClientMock) {
				// 模拟 SCAN 操作返回一些键
				keys := []string{"refresh_token:token1"}
				mock.ExpectScan(0, "refresh_token:*", -1).SetVal(keys, 0)

				// 模拟 GET 操作出错（但实际实现中会忽略错误继续处理）
				mock.ExpectGet("refresh_token:token1").SetErr(assert.AnError)
			},
			wantErr: false, // 实际实现中GET错误不会导致整体失败
		},
		{
			name:   "DEL操作出错",
			userID: 111,
			mockFn: func(mock redismock.ClientMock) {
				// 模拟 SCAN 操作找到匹配的键
				keys := []string{"refresh_token:token1"}
				mock.ExpectScan(0, "refresh_token:*", -1).SetVal(keys, 0)

				// 模拟 GET 操作返回正确的用户ID
				mock.ExpectGet("refresh_token:token1").SetVal("111")

				// 模拟 DEL 操作出错
				mock.ExpectDel("refresh_token:token1").RedisNil()
			},
			wantErr: true,
		},
		{
			name:   "过滤不匹配的用户ID",
			userID: 222,
			mockFn: func(mock redismock.ClientMock) {
				// 模拟 SCAN 操作找到一些键
				keys := []string{"refresh_token:token1", "refresh_token:token2"}
				mock.ExpectScan(0, "refresh_token:*", -1).SetVal(keys, 0)

				// 模拟 GET 操作返回不匹配的用户ID
				mock.ExpectGet("refresh_token:token1").SetVal("333") // 不匹配
				mock.ExpectGet("refresh_token:token2").SetVal("222") // 匹配

				// 模拟 DEL 操作只删除匹配的键
				mock.ExpectDel("refresh_token:token2").SetVal(1)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 Redis mock
			rds, mock := redismock.NewClientMock()

			// 创建 Data 结构体
			data := &Data{
				rds: rds,
				db:  nil,
			}

			// 创建 auth repository
			repo := NewAuthRepository(data, log.DefaultLogger)

			// 设置 mock 期望
			tt.mockFn(mock)

			err := repo.DeleteAllRefreshTokens(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 验证所有期望都被调用
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestAuthRepository_RefreshTokenAtomically 测试原子性地刷新令牌
func TestAuthRepository_RefreshTokenAtomically(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		oldToken  string
		newToken  string
		expiresAt time.Time
		mockFn    func(mock redismock.ClientMock)
		wantErr   bool
	}{
		{
			name:      "成功原子性刷新令牌",
			userID:    123,
			oldToken:  "old_token",
			newToken:  "new_token",
			expiresAt: time.Now().Add(7 * 24 * time.Hour),
			mockFn: func(mock redismock.ClientMock) {
				// 模拟 DEL 操作删除旧令牌
				oldKey := fmt.Sprintf("refresh_token:%s", "old_token")
				mock.ExpectDel(oldKey).SetVal(1)

				// 模拟 SET 操作存储新令牌
				newKey := fmt.Sprintf("refresh_token:%s", "new_token")
				mock.ExpectSet(newKey, int64(123), time.Until(time.Now().Add(7*24*time.Hour))).SetVal("OK")
			},
			wantErr: false,
		},
		{
			name:      "删除旧令牌失败",
			userID:    456,
			oldToken:  "old_token_error",
			newToken:  "new_token",
			expiresAt: time.Now().Add(24 * time.Hour),
			mockFn: func(mock redismock.ClientMock) {
				// 模拟 DEL 操作失败
				oldKey := fmt.Sprintf("refresh_token:%s", "old_token_error")
				mock.ExpectDel(oldKey).RedisNil()

				// 不应该有 SET 操作
			},
			wantErr: true,
		},
		{
			name:      "存储新令牌失败",
			userID:    789,
			oldToken:  "old_token",
			newToken:  "new_token_error",
			expiresAt: time.Now().Add(12 * time.Hour),
			mockFn: func(mock redismock.ClientMock) {
				// 模拟 DEL 操作成功删除旧令牌
				oldKey := fmt.Sprintf("refresh_token:%s", "old_token")
				mock.ExpectDel(oldKey).SetVal(1)

				// 模拟 SET 操作失败
				newKey := fmt.Sprintf("refresh_token:%s", "new_token_error")
				mock.ExpectSet(newKey, int64(789), time.Until(time.Now().Add(12*time.Hour))).RedisNil()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 Redis mock
			rds, mock := redismock.NewClientMock()

			// 创建 Data 结构体
			data := &Data{
				rds: rds,
				db:  nil,
			}

			// 创建 auth repository
			repo := NewAuthRepository(data, log.DefaultLogger)

			// 设置 mock 期望
			tt.mockFn(mock)

			err := repo.RefreshTokenAtomically(context.Background(), tt.userID, tt.oldToken, tt.newToken, tt.expiresAt)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 验证所有期望都被调用
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestAuthRepository_NewAuthRepository 测试构造函数
func TestAuthRepository_NewAuthRepository(t *testing.T) {
	// 创建测试用的 Data 结构体
	data := &Data{
		rds: nil,
		db:  nil,
	}

	// 创建 auth repository
	repo := NewAuthRepository(data, log.DefaultLogger)

	// 验证返回的是正确的类型
	assert.NotNil(t, repo)

	// 验证实现了 biz.AuthRepository 接口
	var authRepo biz.AuthRepository = repo
	assert.NotNil(t, authRepo)

	// 验证内部字段已正确设置
	authRepoImpl := repo.(*authRepository)
	assert.Equal(t, data, authRepoImpl.data)
	assert.NotNil(t, authRepoImpl.logger)
}
