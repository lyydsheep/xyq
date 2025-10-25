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
	"gorm.io/gorm"
)

// TestUserPointRepository_Create 测试用户点数创建功能
func TestUserPointRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		point     *biz.UserPoint
		mockFn    func(sqlmock.Sqlmock)
		wantErr   bool
		expectErr string
	}{
		{
			name: "成功创建用户点数",
			point: &biz.UserPoint{
				UserID:        1,
				CurrentPoints: 100,
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `user_point`").
					WithArgs(
						1,   // user_id
						100, // current_points
						sqlmock.AnyArg(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "创建用户点数失败 - 数据库错误",
			point: &biz.UserPoint{
				UserID:        1,
				CurrentPoints: 100,
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `user_point`").
					WithArgs(
						1,   // user_id
						100, // current_points
						sqlmock.AnyArg(),
					).
					WillReturnError(fmt.Errorf("database connection error"))
				mock.ExpectRollback()
			},
			wantErr:   true,
			expectErr: "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := NewUserPointRepository(db, log.DefaultLogger)
			tt.mockFn(mock)

			err := repo.Create(context.Background(), tt.point)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectErr != "" {
					assert.Contains(t, err.Error(), tt.expectErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.point.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestUserPointRepository_GetByUserID 测试根据用户ID获取点数
func TestUserPointRepository_GetByUserID(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		mockFn    func(sqlmock.Sqlmock)
		wantPoint *biz.UserPoint
		wantErr   bool
		expectErr string
	}{
		{
			name:   "成功获取用户点数",
			userID: 1,
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "current_points", "created_at", "updated_at"}).
					AddRow(1, 1, 100, time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `user_point` WHERE user_id = \\? ORDER BY `user_point`.`id` LIMIT \\?").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantPoint: &biz.UserPoint{
				ID:            1,
				UserID:        1,
				CurrentPoints: 100,
			},
			wantErr: false,
		},
		{
			name:   "用户点数不存在",
			userID: 999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `user_point` WHERE user_id = \\? ORDER BY `user_point`.`id` LIMIT \\?").
					WithArgs(999, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantPoint: nil,
			wantErr:   true,
			expectErr: "record not found",
		},
		{
			name:   "查询失败",
			userID: 2,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `user_point` WHERE user_id = \\? ORDER BY `user_point`.`id` LIMIT \\?").
					WithArgs(2, 1).
					WillReturnError(fmt.Errorf("database connection error"))
			},
			wantPoint: nil,
			wantErr:   true,
			expectErr: "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := NewUserPointRepository(db, log.DefaultLogger)
			tt.mockFn(mock)

			point, err := repo.GetByUserID(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectErr != "" {
					assert.Contains(t, err.Error(), tt.expectErr)
				}
				assert.Nil(t, point)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPoint.ID, point.ID)
				assert.Equal(t, tt.wantPoint.UserID, point.UserID)
				assert.Equal(t, tt.wantPoint.CurrentPoints, point.CurrentPoints)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
