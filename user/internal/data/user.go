package data

import (
	"context"
	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"user/internal/pkg/tracing"
)

// userRepository 用户数据访问实现
type userRepository struct {
	db     *gorm.DB
	logger *log.Helper
}

func (r *userRepository) Update(ctx context.Context, id int64, req *biz.UpdateUserRequest) error {
	ctx, span := tracing.StartSpan(ctx, "UserRepository.Update")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"user_id": id,
		"has_nickname": req.Nickname != nil,
		"has_avatar_url": req.AvatarURL != nil,
	})

	r.logger.WithContext(ctx).Infof("Updating user with id: %d", id)

	updates := make(map[string]interface{})

	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}

	if req.AvatarURL != nil {
		updates["avatar_url"] = *req.AvatarURL
	}

	if len(updates) == 0 {
		r.logger.WithContext(ctx).Infof("No fields to update for user id: %d", id)
		return nil
	}

	err := r.db.WithContext(ctx).Model(&biz.User{}).Where("id = ?", id).Updates(updates).Error

	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to update user with id: %d, error: %v", id, err)
		return err
	}

	r.logger.WithContext(ctx).Infof("Successfully updated user with id: %d", id)
	return nil
}

// NewUserRepository 创建用户数据访问实例
func NewUserRepository(db *gorm.DB, logger log.Logger) biz.UserRepository {
	return &userRepository{db: db, logger: log.NewHelper(logger)}
}

func (r *userRepository) Create(ctx context.Context, user *biz.User) error {
	ctx, span := tracing.StartSpan(ctx, "UserRepository.Create")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"email": user.Email,
	})

	r.logger.WithContext(ctx).Infof("Creating user with email: %s", user.Email)
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to create user with email: %s, error: %v", user.Email, err)
		return err
	}

	r.logger.WithContext(ctx).Infof("Successfully created user with id: %d, email: %s", user.ID, user.Email)
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*biz.User, error) {
	ctx, span := tracing.StartSpan(ctx, "UserRepository.GetByID")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"user_id": id,
	})

	r.logger.WithContext(ctx).Infof("Getting user with id: %d", id)
	var u biz.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to get user with id: %d, error: %v", id, err)
		return nil, err
	}

	r.logger.WithContext(ctx).Infof("Successfully retrieved user with id: %d, email: %s", id, u.Email)
	return &u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*biz.User, error) {
	ctx, span := tracing.StartSpan(ctx, "UserRepository.GetByEmail")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"email": email,
	})

	r.logger.WithContext(ctx).Infof("Getting user with email: %s", email)
	var u biz.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		r.logger.WithContext(ctx).Errorf("Failed to get user with email: %s, error: %v", email, err)
		return nil, err
	}

	r.logger.WithContext(ctx).Infof("Successfully retrieved user with id: %d, email: %s", u.ID, email)
	return &u, nil
}
