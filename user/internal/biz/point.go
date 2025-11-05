package biz

import (
	"context"
	"time"
)

// UserPoint 用户点数表
type UserPoint struct {
	ID            int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID        int64     `gorm:"column:user_id;uniqueIndex;not null" json:"user_id"`
	CurrentPoints uint32    `gorm:"column:current_points;not null;default:0" json:"current_points"`
	TotalConsumed uint32    `gorm:"column:total_consumed;not null;default:0" json:"total_consumed"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定表名
func (UserPoint) TableName() string {
	return "user_point"
}

// UserPointRepository 用户点数数据访问接口
type UserPointRepository interface {
	Create(ctx context.Context, userPoint *UserPoint) error
	GetByUserID(ctx context.Context, userID int64) (*UserPoint, error)
}
