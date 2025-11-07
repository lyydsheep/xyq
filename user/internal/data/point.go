package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"user/internal/biz"

	"gorm.io/gorm"
	"user/internal/pkg/tracing"
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
	ctx, span := tracing.StartSpan(ctx, "UserPointRepository.Create")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"user_id": userPoint.UserID,
	})

	r.logger.WithContext(ctx).Infof("Creating user point for user_id: %d", userPoint.UserID)
	err := r.db.WithContext(ctx).Create(userPoint).Error
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to create user point for user_id: %d, error_reason: %v", userPoint.UserID, err)
		return err
	}

	r.logger.WithContext(ctx).Infof("Successfully created user point with id: %d for user_id: %d", userPoint.ID, userPoint.UserID)
	return nil
}

func (r *userPointRepository) GetByUserID(ctx context.Context, userID int64) (*biz.UserPoint, error) {
	ctx, span := tracing.StartSpan(ctx, "UserPointRepository.GetByUserID")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"user_id": userID,
	})

	r.logger.WithContext(ctx).Infof("Getting user point for user_id: %d", userID)
	var p biz.UserPoint
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to get user point for user_id: %d, error_reason: %v", userID, err)
		return nil, err
	}

	r.logger.WithContext(ctx).Infof("Successfully retrieved user point with id: %d for user_id: %d", p.ID, userID)
	return &p, nil
}
