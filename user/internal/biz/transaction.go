package biz

import (
	"context"
	"time"
)

// PointTransactionType 点数交易类型
type PointTransactionType string

const (
	PointTransactionConsume  PointTransactionType = "CONSUME"
	PointTransactionRecharge PointTransactionType = "RECHARGE"
)

// PointTransaction 点数交易流水表
type PointTransaction struct {
	ID            int64                `gorm:"column:id;primaryKey" json:"id"`
	UserID        int64                `gorm:"column:user_id;not null;index" json:"user_id"`
	Type          PointTransactionType `gorm:"column:type;not null" json:"type"`
	Amount        uint32               `gorm:"column:amount;not null" json:"amount"`
	RelatedBookID *int64               `gorm:"column:related_book_id" json:"related_book_id,omitempty"`
	Description   *string              `gorm:"column:description" json:"description,omitempty"`
	CreatedAt     time.Time            `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time            `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定表名
func (PointTransaction) TableName() string {
	return "point_transaction"
}

// PointTransactionRepository 点数交易流水数据访问接口
type PointTransactionRepository interface {
	Create(ctx context.Context, transaction *PointTransaction) error
	GetByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*PointTransaction, int64, error)
	GetByID(ctx context.Context, id int64) (*PointTransaction, error)
}
