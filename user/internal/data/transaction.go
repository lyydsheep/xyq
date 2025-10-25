package data

import (
	"context"
	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// pointTransactionRepository 点数交易流水数据访问实现
type pointTransactionRepository struct {
	db     *gorm.DB
	logger *log.Helper
}

// NewPointTransactionRepository 创建点数交易流水数据访问实例
func NewPointTransactionRepository(db *gorm.DB, logger log.Logger) biz.PointTransactionRepository {
	return &pointTransactionRepository{db: db, logger: log.NewHelper(logger)}
}

func (r *pointTransactionRepository) Create(ctx context.Context, transaction *biz.PointTransaction) error {
	r.logger.Log(log.LevelInfo, "Creating point transaction for user_id: ", transaction.UserID, ", type: ", transaction.Type, ", amount: ", transaction.Amount)
	err := r.db.WithContext(ctx).Create(transaction).Error
	if err != nil {
		r.logger.Log(log.LevelError, "Failed to create point transaction for user_id: ", transaction.UserID, ", error: ", err)
		return err
	}
	r.logger.Log(log.LevelInfo, "Successfully created point transaction with id: ", transaction.ID, " for user_id: ", transaction.UserID)
	return nil
}

func (r *pointTransactionRepository) GetByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*biz.PointTransaction, int64, error) {
	r.logger.Log(log.LevelInfo, "Getting point transactions for user_id: ", userID, ", page: ", page, ", pageSize: ", pageSize)
	var transactions []*biz.PointTransaction
	var total int64

	offset := (page - 1) * pageSize

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&biz.PointTransaction{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		r.logger.Log(log.LevelError, "Failed to count point transactions for user_id: ", userID, ", error: ", err)
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&transactions).Error

	if err != nil {
		r.logger.Log(log.LevelError, "Failed to get point transactions for user_id: ", userID, ", error: ", err)
		return nil, 0, err
	}

	r.logger.Log(log.LevelInfo, "Successfully retrieved ", len(transactions), " point transactions for user_id: ", userID, ", total count: ", total)
	return transactions, total, nil
}

func (r *pointTransactionRepository) GetByID(ctx context.Context, id int64) (*biz.PointTransaction, error) {
	r.logger.Log(log.LevelInfo, "Getting point transaction with id: ", id)
	var t biz.PointTransaction
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error
	if err != nil {
		r.logger.Log(log.LevelError, "Failed to get point transaction with id: ", id, ", error: ", err)
		return nil, err
	}
	r.logger.Log(log.LevelInfo, "Successfully retrieved point transaction with id: ", id, " for user_id: ", t.UserID)
	return &t, nil
}
