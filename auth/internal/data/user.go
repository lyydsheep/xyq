package data

import (
	"auth/internal/biz"
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// userRepository 用户数据访问实现
type userRepository struct {
	db     *gorm.DB
	logger *log.Helper
}

func (r *userRepository) Update(ctx context.Context, id int64, req *biz.UpdateUserRequest) error {
	r.logger.Log(log.LevelInfo, "Updating user with id: ", id)

	// 1. 构建一个只包含需要更新字段的 Map
	updates := make(map[string]interface{})

	// 检查 Nickname 是否为 nil，如果不为 nil，则表示需要更新（可能是新值或空字符串 ""）
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname // 解引用指针，获取 string 值
	}

	// 检查 AvatarURL 是否为 nil，如果不为 nil，则表示需要更新（可能是新 URL 或空字符串 ""）
	if req.AvatarURL != nil {
		updates["avatar_url"] = *req.AvatarURL // 解引用指针，获取 string 值
	}

	// 2. 检查是否有字段需要更新
	if len(updates) == 0 {
		r.logger.Log(log.LevelInfo, "No fields to update for user id: ", id)
		return nil
	}

	err := r.db.WithContext(ctx).Model(&biz.User{}).Where("id = ?", id).Updates(updates).Error

	if err != nil {
		r.logger.Log(log.LevelError, "Failed to update user with id: ", id, ", error: ", err)
		return err
	}
	r.logger.Log(log.LevelInfo, "Successfully updated user with id: ", id)
	return nil
}

// NewUserRepository 创建用户数据访问实例
func NewUserRepository(db *gorm.DB, logger log.Logger) biz.UserRepository {
	return &userRepository{db: db, logger: log.NewHelper(logger)}
}

func (r *userRepository) Create(ctx context.Context, user *biz.User) error {
	r.logger.Log(log.LevelInfo, "Creating user with email: ", user.Email)
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		r.logger.Log(log.LevelError, "Failed to create user with email: ", user.Email, ", error: ", err)
		return err
	}
	r.logger.Log(log.LevelInfo, "Successfully created user with id: ", user.ID, ", email: ", user.Email)
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*biz.User, error) {
	r.logger.Log(log.LevelInfo, "Getting user with id: ", id)
	var u biz.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		r.logger.Log(log.LevelError, "Failed to get user with id: ", id, ", error: ", err)
		return nil, err
	}
	r.logger.Log(log.LevelInfo, "Successfully retrieved user with id: ", id, ", email: ", u.Email)
	return &u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*biz.User, error) {
	r.logger.Log(log.LevelInfo, "Getting user with email: ", email)
	var u biz.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		r.logger.Log(log.LevelError, "Failed to get user with email: ", email, ", error: ", err)
		return nil, err
	}
	r.logger.Log(log.LevelInfo, "Successfully retrieved user with id: ", u.ID, ", email: ", email)
	return &u, nil
}
