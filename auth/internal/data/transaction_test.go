package data

import (
	"auth/internal/biz"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestPointTransactionRepository_Create 测试点数交易流水创建功能
func TestPointTransactionRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		transaction *biz.PointTransaction
		mockFn      func(sqlmock.Sqlmock)
		wantErr     bool
		expectErr   string
	}{
		{
			name: "成功创建消费交易流水",
			transaction: &biz.PointTransaction{
				UserID:      1,
				Type:        biz.PointTransactionConsume,
				Amount:      20,
				Description: stringPtr("购买书籍"),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `point_transaction`").
					WithArgs(
						1,         // user_id
						"CONSUME", // type
						20,        // amount
						nil,
						"购买书籍", // description
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "成功创建充值交易流水",
			transaction: &biz.PointTransaction{
				UserID:        1,
				Type:          biz.PointTransactionRecharge,
				Amount:        100,
				RelatedBookID: nil,
				Description:   stringPtr("充值"),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `point_transaction`").
					WithArgs(
						1,          // user_id
						"RECHARGE", // type
						100,        // amount
						nil,        // related_book_id
						"充值",       // description
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "创建交易流水失败 - 数据库错误",
			transaction: &biz.PointTransaction{
				UserID:      1,
				Type:        biz.PointTransactionConsume,
				Amount:      20,
				Description: stringPtr("购买书籍"),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `point_transaction`").
					WithArgs(
						1,         // user_id
						"CONSUME", // type
						20,        // amount
						nil,
						"购买书籍", // description
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
			repo := NewPointTransactionRepository(db, log.DefaultLogger)
			tt.mockFn(mock)

			err := repo.Create(context.Background(), tt.transaction)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectErr != "" {
					assert.Contains(t, err.Error(), tt.expectErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.transaction.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestPointTransactionRepository_GetByUserID 测试根据用户ID获取交易流水
func TestPointTransactionRepository_GetByUserID(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		page      int
		pageSize  int
		mockFn    func(sqlmock.Sqlmock)
		wantTxns  []*biz.PointTransaction
		wantTotal int64
		wantErr   bool
		expectErr string
	}{
		{
			name:     "成功获取用户交易流水",
			userID:   1,
			page:     1,
			pageSize: 10,
			mockFn: func(mock sqlmock.Sqlmock) {
				// 模拟总数查询
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `point_transaction` WHERE user_id = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

				// 模拟分页查询
				rows := sqlmock.NewRows([]string{"id", "user_id", "type", "amount", "related_book_id", "description", "created_at", "updated_at"}).
					AddRow(1, 1, "CONSUME", 20, int64Ptr(101), "购买书籍", time.Now(), time.Now()).
					AddRow(2, 1, "RECHARGE", 100, nil, "充值", time.Now(), time.Now()).
					AddRow(3, 1, "CONSUME", 30, int64Ptr(102), "购买书籍", time.Now(), time.Now())

				mock.ExpectQuery("SELECT \\* FROM `point_transaction` WHERE user_id = \\? ORDER BY created_at DESC LIMIT \\?").
					WithArgs(1, 10).
					WillReturnRows(rows)
			},
			wantTxns: []*biz.PointTransaction{
				{
					ID:            1,
					UserID:        1,
					Type:          biz.PointTransactionConsume,
					Amount:        20,
					RelatedBookID: int64Ptr(101),
					Description:   stringPtr("购买书籍"),
				},
				{
					ID:            2,
					UserID:        1,
					Type:          biz.PointTransactionRecharge,
					Amount:        100,
					RelatedBookID: nil,
					Description:   stringPtr("充值"),
				},
				{
					ID:            3,
					UserID:        1,
					Type:          biz.PointTransactionConsume,
					Amount:        30,
					RelatedBookID: int64Ptr(102),
					Description:   stringPtr("购买书籍"),
				},
			},
			wantTotal: 5,
			wantErr:   false,
		},
		{
			name:     "用户交易流水为空",
			userID:   999,
			page:     1,
			pageSize: 10,
			mockFn: func(mock sqlmock.Sqlmock) {
				// 模拟总数查询
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `point_transaction` WHERE user_id = \\?").
					WithArgs(999).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				// 模拟分页查询
				mock.ExpectQuery("SELECT \\* FROM `point_transaction` WHERE user_id = \\? ORDER BY created_at DESC LIMIT \\?").
					WithArgs(999, 10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "type", "amount", "related_book_id", "description", "created_at", "updated_at"}))
			},
			wantTxns:  []*biz.PointTransaction{},
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:     "查询总数失败",
			userID:   1,
			page:     1,
			pageSize: 10,
			mockFn: func(mock sqlmock.Sqlmock) {
				// 模拟总数查询失败
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `point_transaction` WHERE user_id = \\?").
					WithArgs(1).
					WillReturnError(fmt.Errorf("database connection error"))
			},
			wantTxns:  nil,
			wantTotal: 0,
			wantErr:   true,
			expectErr: "database connection error",
		},
		{
			name:     "查询分页数据失败",
			userID:   1,
			page:     1,
			pageSize: 10,
			mockFn: func(mock sqlmock.Sqlmock) {
				// 模拟总数查询成功
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `point_transaction` WHERE user_id = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

				// 模拟分页查询失败
				mock.ExpectQuery("SELECT \\* FROM `point_transaction` WHERE user_id = \\? ORDER BY created_at DESC LIMIT \\?").
					WithArgs(1, 10).
					WillReturnError(fmt.Errorf("database connection error"))
			},
			wantTxns:  nil,
			wantTotal: 0,
			wantErr:   true,
			expectErr: "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := NewPointTransactionRepository(db, log.DefaultLogger)
			tt.mockFn(mock)

			transactions, total, err := repo.GetByUserID(context.Background(), tt.userID, tt.page, tt.pageSize)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectErr != "" {
					assert.Contains(t, err.Error(), tt.expectErr)
				}
				assert.Nil(t, transactions)
				assert.Equal(t, int64(0), total)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total)
				assert.Equal(t, len(tt.wantTxns), len(transactions))

				for i, wantTxn := range tt.wantTxns {
					assert.Equal(t, wantTxn.ID, transactions[i].ID)
					assert.Equal(t, wantTxn.UserID, transactions[i].UserID)
					assert.Equal(t, wantTxn.Type, transactions[i].Type)
					assert.Equal(t, wantTxn.Amount, transactions[i].Amount)
					assert.Equal(t, wantTxn.RelatedBookID, transactions[i].RelatedBookID)
					assert.Equal(t, wantTxn.Description, transactions[i].Description)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestPointTransactionRepository_GetByID 测试根据ID获取交易流水
func TestPointTransactionRepository_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		transactionID int64
		mockFn        func(sqlmock.Sqlmock)
		wantTxn       *biz.PointTransaction
		wantErr       bool
		expectErr     string
	}{
		{
			name:          "成功获取交易流水",
			transactionID: 1,
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "type", "amount", "related_book_id", "description", "created_at", "updated_at"}).
					AddRow(1, 1, "CONSUME", 20, int64Ptr(101), "购买书籍", time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `point_transaction` WHERE id = \\? ORDER BY `point_transaction`.`id` LIMIT \\?").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantTxn: &biz.PointTransaction{
				ID:            1,
				UserID:        1,
				Type:          biz.PointTransactionConsume,
				Amount:        20,
				RelatedBookID: int64Ptr(101),
				Description:   stringPtr("购买书籍"),
			},
			wantErr: false,
		},
		{
			name:          "交易流水不存在",
			transactionID: 999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `point_transaction` WHERE id = \\? ORDER BY `point_transaction`.`id` LIMIT \\?").
					WithArgs(999, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantTxn:   nil,
			wantErr:   true,
			expectErr: "record not found",
		},
		{
			name:          "查询失败",
			transactionID: 2,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `point_transaction` WHERE id = \\? ORDER BY `point_transaction`.`id` LIMIT \\?").
					WithArgs(2, 1).
					WillReturnError(fmt.Errorf("database connection error"))
			},
			wantTxn:   nil,
			wantErr:   true,
			expectErr: "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := NewPointTransactionRepository(db, log.DefaultLogger)
			tt.mockFn(mock)

			transaction, err := repo.GetByID(context.Background(), tt.transactionID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectErr != "" {
					assert.Contains(t, err.Error(), tt.expectErr)
				}
				assert.Nil(t, transaction)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTxn.ID, transaction.ID)
				assert.Equal(t, tt.wantTxn.UserID, transaction.UserID)
				assert.Equal(t, tt.wantTxn.Type, transaction.Type)
				assert.Equal(t, tt.wantTxn.Amount, transaction.Amount)
				assert.Equal(t, tt.wantTxn.RelatedBookID, transaction.RelatedBookID)
				assert.Equal(t, tt.wantTxn.Description, transaction.Description)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestPointTransactionRepository_Integration 测试点数交易流水操作的集成场景
func TestPointTransactionRepository_Integration(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:   "创建并获取交易流水",
			userID: 1,
			mockFn: func(mock sqlmock.Sqlmock) {
				// 创建消费交易流水
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `point_transaction`").
					WithArgs(
						1,             // user_id
						"CONSUME",     // type
						20,            // amount
						int64Ptr(101), // related_book_id
						"购买书籍",        // description
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				// 根据ID获取交易流水
				rows := sqlmock.NewRows([]string{"id", "user_id", "type", "amount", "related_book_id", "description", "created_at", "updated_at"}).
					AddRow(1, 1, "CONSUME", 20, int64Ptr(101), "购买书籍", time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `point_transaction` WHERE id = \\? ORDER BY `point_transaction`.`id` LIMIT \\?").
					WithArgs(1, 1).
					WillReturnRows(rows)

				// 创建充值交易流水
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `point_transaction`").
					WithArgs(
						1,          // user_id
						"RECHARGE", // type
						100,        // amount
						nil,        // related_book_id
						"充值",       // description
					).
					WillReturnResult(sqlmock.NewResult(2, 1))
				mock.ExpectCommit()

				// 获取用户交易流水列表
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `point_transaction` WHERE user_id = \\?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

				listRows := sqlmock.NewRows([]string{"id", "user_id", "type", "amount", "related_book_id", "description", "created_at", "updated_at"}).
					AddRow(2, 1, "RECHARGE", 100, nil, "充值", time.Now(), time.Now()).
					AddRow(1, 1, "CONSUME", 20, int64Ptr(101), "购买书籍", time.Now(), time.Now())

				mock.ExpectQuery("SELECT \\* FROM `point_transaction` WHERE user_id = \\? ORDER BY created_at DESC LIMIT \\?").
					WithArgs(1, 10).
					WillReturnRows(listRows)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := NewPointTransactionRepository(db, log.DefaultLogger)
			tt.mockFn(mock)

			// 创建消费交易流水
			consumeTxn := &biz.PointTransaction{
				UserID:        tt.userID,
				Type:          biz.PointTransactionConsume,
				Amount:        20,
				RelatedBookID: int64Ptr(101),
				Description:   stringPtr("购买书籍"),
			}
			err := repo.Create(context.Background(), consumeTxn)
			assert.NoError(t, err)
			assert.NotZero(t, consumeTxn.ID)

			// 根据ID获取交易流水
			retrievedTxn, err := repo.GetByID(context.Background(), consumeTxn.ID)
			assert.NoError(t, err)
			assert.Equal(t, consumeTxn.ID, retrievedTxn.ID)
			assert.Equal(t, consumeTxn.UserID, retrievedTxn.UserID)
			assert.Equal(t, consumeTxn.Type, retrievedTxn.Type)
			assert.Equal(t, consumeTxn.Amount, retrievedTxn.Amount)
			assert.Equal(t, consumeTxn.RelatedBookID, retrievedTxn.RelatedBookID)
			assert.Equal(t, consumeTxn.Description, retrievedTxn.Description)

			// 创建充值交易流水
			rechargeTxn := &biz.PointTransaction{
				UserID:      tt.userID,
				Type:        biz.PointTransactionRecharge,
				Amount:      100,
				Description: stringPtr("充值"),
			}
			err = repo.Create(context.Background(), rechargeTxn)
			assert.NoError(t, err)
			assert.NotZero(t, rechargeTxn.ID)

			// 获取用户交易流水列表
			transactions, total, err := repo.GetByUserID(context.Background(), tt.userID, 1, 10)
			assert.NoError(t, err)
			assert.Equal(t, int64(2), total)
			assert.Equal(t, 2, len(transactions))
			assert.Equal(t, rechargeTxn.ID, transactions[0].ID) // 充值记录应该在前，因为是按创建时间倒序排列
			assert.Equal(t, consumeTxn.ID, transactions[1].ID)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// 辅助函数：字符串指针
func stringPtr(s string) *string {
	return &s
}

// 辅助函数：int64指针
func int64Ptr(i int64) *int64 {
	return &i
}
