package data

import (
	"context"
	"fmt"
	"testing"
	"time"
	"user/internal/biz"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestUserRepository_Create 测试用户创建功能
func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		user    *biz.User
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "成功创建用户",
			user: &biz.User{
				Email:        "test@example.com",
				PasswordHash: "hashed_password",
				Nickname:     "测试用户",
				IsPremium:    0,
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `user`").
					WithArgs(
						"test@example.com",
						"hashed_password",
						"测试用户",
						"", // avatar_url
						0,  // is_premium
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "创建用户失败 - 邮箱已存在",
			user: &biz.User{
				Email:        "existing@example.com",
				PasswordHash: "hashed_password",
				Nickname:     "测试用户",
				IsPremium:    0,
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `user`").
					WithArgs(
						"existing@example.com",
						"hashed_password",
						"测试用户",
						"", // avatar_url
						0,  // is_premium
					).
					WillReturnError(fmt.Errorf("duplicate entry"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := NewUserRepository(db, log.DefaultLogger)
			tt.mockFn(mock)

			err := repo.Create(context.Background(), tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.user.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestUserRepository_GetByID 测试根据ID获取用户
func TestUserRepository_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		mockFn    func(sqlmock.Sqlmock)
		wantUser  *biz.User
		wantErr   bool
		expectErr string
	}{
		{
			name:   "成功获取用户",
			userID: 1,
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "nickname", "avatar_url", "is_premium", "created_at", "updated_at"}).
					AddRow(1, "test@example.com", "hashed_password", "测试用户", "", 0, time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `user` WHERE id = \\? ORDER BY `user`.`id` LIMIT \\?").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantUser: &biz.User{
				ID:           1,
				Email:        "test@example.com",
				PasswordHash: "hashed_password",
				Nickname:     "测试用户",
				IsPremium:    0,
			},
			wantErr: false,
		},
		{
			name:   "用户不存在",
			userID: 999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `user` WHERE id = \\? ORDER BY `user`.`id` LIMIT \\?").
					WithArgs(999, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantUser:  nil,
			wantErr:   true,
			expectErr: "record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := NewUserRepository(db, log.DefaultLogger)
			tt.mockFn(mock)

			user, err := repo.GetByID(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectErr != "" {
					assert.Contains(t, err.Error(), tt.expectErr)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser.ID, user.ID)
				assert.Equal(t, tt.wantUser.Email, user.Email)
				assert.Equal(t, tt.wantUser.Nickname, user.Nickname)
				assert.Equal(t, tt.wantUser.IsPremium, user.IsPremium)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestUserRepository_GetByEmail 测试根据邮箱获取用户
func TestUserRepository_GetByEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		mockFn    func(sqlmock.Sqlmock)
		wantUser  *biz.User
		wantErr   bool
		expectErr string
	}{
		{
			name:  "成功获取用户",
			email: "test@example.com",
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "nickname", "avatar_url", "is_premium", "created_at", "updated_at"}).
					AddRow(1, "test@example.com", "hashed_password", "测试用户", "", 0, time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `user` WHERE email = \\? ORDER BY `user`.`id` LIMIT \\?").
					WithArgs("test@example.com", 1).
					WillReturnRows(rows)
			},
			wantUser: &biz.User{
				ID:           1,
				Email:        "test@example.com",
				PasswordHash: "hashed_password",
				Nickname:     "测试用户",
				IsPremium:    0,
			},
			wantErr: false,
		},
		{
			name:  "用户不存在",
			email: "nonexistent@example.com",
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `user` WHERE email = \\? ORDER BY `user`.`id` LIMIT \\?").
					WithArgs("nonexistent@example.com", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantUser:  nil,
			wantErr:   true,
			expectErr: "record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := NewUserRepository(db, log.DefaultLogger)
			tt.mockFn(mock)

			user, err := repo.GetByEmail(context.Background(), tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectErr != "" {
					assert.Contains(t, err.Error(), tt.expectErr)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser.ID, user.ID)
				assert.Equal(t, tt.wantUser.Email, user.Email)
				assert.Equal(t, tt.wantUser.Nickname, user.Nickname)
				assert.Equal(t, tt.wantUser.IsPremium, user.IsPremium)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库失败: %v", err)
	}

	// 配置GORM
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 禁用GORM日志，避免测试输出混乱
	})
	if err != nil {
		t.Fatalf("连接GORM数据库失败: %v", err)
	}

	return gormDB, mock
}

// TestUserRepository_Update 测试用户更新功能
func TestUserRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		req     *biz.UpdateUserRequest
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:   "成功更新昵称",
			userID: 1,
			req: &biz.UpdateUserRequest{
				Nickname: stringPtr("新昵称"),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `user` SET `nickname`=\\?,`updated_at`=\\? WHERE id = \\?").
					WithArgs("新昵称", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:   "成功更新头像URL",
			userID: 1,
			req: &biz.UpdateUserRequest{
				AvatarURL: stringPtr("https://example.com/avatar.jpg"),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `user` SET `avatar_url`=\\?,`updated_at`=\\? WHERE id = \\?").
					WithArgs("https://example.com/avatar.jpg", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:   "成功更新昵称和头像URL",
			userID: 1,
			req: &biz.UpdateUserRequest{
				Nickname:  stringPtr("新昵称"),
				AvatarURL: stringPtr("https://example.com/avatar.jpg"),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `user` SET `avatar_url`=\\?,`nickname`=\\?,`updated_at`=\\? WHERE id = \\?").
					WithArgs("https://example.com/avatar.jpg", "新昵称", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:   "成功将头像URL设置为NULL",
			userID: 1,
			req: &biz.UpdateUserRequest{
				// 使用空字符串指针来表示将数据库字段设置为NULL或空值
				AvatarURL: stringPtr(""),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `user` SET `avatar_url`=\\?,`updated_at`=\\? WHERE id = \\?").
					WithArgs("", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:   "没有字段需要更新",
			userID: 1,
			req:    &biz.UpdateUserRequest{
				// 所有字段都为nil，表示不更新任何字段
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				// 不应该有任何数据库操作
			},
			wantErr: false,
		},
		{
			name:   "更新用户失败 - 用户不存在",
			userID: 999,
			req: &biz.UpdateUserRequest{
				Nickname: stringPtr("不存在的用户"),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `user` SET `nickname`=\\?,`updated_at`=\\? WHERE id = \\?").
					WithArgs("不存在的用户", sqlmock.AnyArg(), 999).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name:   "更新用户失败 - 数据库错误",
			userID: 1,
			req: &biz.UpdateUserRequest{
				Nickname: stringPtr("测试昵称"),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `user` SET `nickname`=\\?,`updated_at`=\\? WHERE id = \\?").
					WithArgs("测试昵称", sqlmock.AnyArg(), 1).
					WillReturnError(fmt.Errorf("database connection error_reason"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := NewUserRepository(db, log.DefaultLogger)
			tt.mockFn(mock)

			err := repo.Update(context.Background(), tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
