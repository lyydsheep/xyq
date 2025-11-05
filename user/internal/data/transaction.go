package data

import (
	"context"
	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"user/internal/pkg/tracing"
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
	ctx, span := tracing.StartSpan(ctx, "PointTransactionRepository.Create")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"user_id": transaction.UserID,
		"transaction_type": transaction.Type,
		"amount": transaction.Amount,
	})

	r.logger.WithContext(ctx).Infof("Creating point transaction for user_id: %d, type: %s, amount: %d", transaction.UserID, transaction.Type, transaction.Amount)
	err := r.db.WithContext(ctx).Create(transaction).Error
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to create point transaction for user_id: %d, error: %v", transaction.UserID, err)
		return err
	}

	r.logger.WithContext(ctx).Infof("Successfully created point transaction with id: %d for user_id: %d", transaction.ID, transaction.UserID)
	return nil
}

func (r *pointTransactionRepository) GetByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*biz.PointTransaction, int64, error) {
	ctx, span := tracing.StartSpan(ctx, "PointTransactionRepository.GetByUserID")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"user_id": userID,
		"page": page,
		"page_size": pageSize,
	})

	r.logger.WithContext(ctx).Infof("Getting point transactions for user_id: %d, page: %d, pageSize: %d", userID, page, pageSize)
	var transactions []*biz.PointTransaction
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&biz.PointTransaction{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to count point transactions for user_id: %d, error: %v", userID, err)
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&transactions).Error

	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to get point transactions for user_id: %d, error: %v", userID, err)
		return nil, 0, err
	}

	r.logger.WithContext(ctx).Infof("Successfully retrieved %d point transactions for user_id: %d, total count: %d", len(transactions), userID, total)
	return transactions, total, nil
}

func (r *pointTransactionRepository) GetByID(ctx context.Context, id int64) (*biz.PointTransaction, error) {
	ctx, span := tracing.StartSpan(ctx, "PointTransactionRepository.GetByID")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"transaction_id": id,
	})

	r.logger.WithContext(ctx).Infof("Getting point transaction with id: %d", id)
	var t biz.PointTransaction
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to get point transaction with id: %d, error: %v", id, err)
		return nil, err
	}

	r.logger.WithContext(ctx).Infof("Successfully retrieved point transaction with id: %d for user_id: %d", id, t.UserID)
	return &t, nil
}
