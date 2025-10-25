package data

import (
	"auth/internal/biz"
	"context"
	"github.com/go-kratos/kratos/v2/log"

	"gorm.io/gorm"
)

// userPointRepository 用户点数数据访问实现
type userPointRepository struct {
	db     *gorm.DB
	logger *log.Helper
}

// NewUserPointRepository 创建用户点数数据访问实例
func NewUserPointRepository(db *gorm.DB, logger log.Logger) biz.UserPointRepository {
	return &userPointRepository{db: db, logger: log.NewHelper(logger)}
}

func (r *userPointRepository) Create(ctx context.Context, userPoint *biz.UserPoint) error {
	r.logger.Log(log.LevelInfo, "Creating user point for user_id: ", userPoint.UserID)
	err := r.db.WithContext(ctx).Create(userPoint).Error
	if err != nil {
		r.logger.Log(log.LevelError, "Failed to create user point for user_id: ", userPoint.UserID, ", error: ", err)
		return err
	}
	r.logger.Log(log.LevelInfo, "Successfully created user point with id: ", userPoint.ID, " for user_id: ", userPoint.UserID)
	return nil
}

func (r *userPointRepository) GetByUserID(ctx context.Context, userID int64) (*biz.UserPoint, error) {
	r.logger.Log(log.LevelInfo, "Getting user point for user_id: ", userID)
	var p biz.UserPoint
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error
	if err != nil {
		r.logger.Log(log.LevelError, "Failed to get user point for user_id: ", userID, ", error: ", err)
		return nil, err
	}
	r.logger.Log(log.LevelInfo, "Successfully retrieved user point with id: ", p.ID, " for user_id: ", userID)
	return &p, nil
}
