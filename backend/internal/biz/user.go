package biz

import (
	"context"
	"time"
)

// User 用户实体
type User struct {
	ID             int32
	Username       string
	PasswordHash   string
	Salt           string
	IsMember       bool
	MemberLevel    int32
	MemberExpireAt *time.Time
	FreeTrialCount int32
	Avatar         string
	CreatedAt      time.Time
}

// UserRepo 用户数据访问接口
type UserRepo interface {
	Create(ctx context.Context, user *User) error
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id int32) (*User, error)
	UpdateAvatar(ctx context.Context, id int32, avatar string) error
}
