package biz

import (
	"context"
	"time"
)

// User 用户基本信息表
type User struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Email        string    `gorm:"column:email;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	Nickname     string    `gorm:"column:nickname;not null;default:'新用户'" json:"nickname"`
	AvatarURL    string    `gorm:"column:avatar_url" json:"avatar_url,omitempty"`
	IsPremium    uint8     `gorm:"column:is_premium;not null;default:0" json:"is_premium"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

type UpdateUserRequest struct {
	// 允许用户更新昵称。使用指针 *string，可以接收 "" (零值) 或 nil (不更新)
	Nickname *string `json:"nickname"`

	//  *string 来表示：nil (不更新), 指向非空字符串的指针 (更新),
	AvatarURL *string `json:"avatar_url"`
}

// TableName 指定表名
func (User) TableName() string {
	return "user"
}

// UserRepository 用户数据访问接口
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, id int64, req *UpdateUserRequest) error
}
