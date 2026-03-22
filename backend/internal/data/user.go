package data

import (
	"context"
	"time"

	"github.com/luminance/backend/ent"
	"github.com/luminance/backend/ent/user"
	"github.com/luminance/backend/internal/biz"
	"github.com/luminance/backend/pkg/crypto"
)

// userRepo 用户仓库实现
type userRepo struct {
	client *ent.Client
}

// NewUserRepo 创建用户仓库
func NewUserRepo(client *ent.Client) biz.UserRepo {
	return &userRepo{client: client}
}

// toBizUser 将 ent.User 转换为 biz.User
func toBizUser(u *ent.User) *biz.User {
	return &biz.User{
		ID:             u.ID,
		Username:       u.Username,
		PasswordHash:   u.PasswordHash,
		Salt:           u.Salt,
		IsMember:       u.IsMember,
		MemberLevel:    int32(u.MemberLevel),
		MemberExpireAt: u.MemberExpireAt,
		FreeTrialCount: u.FreeTrialCount,
		Avatar:         u.Avatar,
		CreatedAt:      u.CreatedAt,
	}
}

// Create 创建用户
func (r *userRepo) Create(ctx context.Context, u *biz.User) error {
	// 生成盐值并哈希密码
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return err
	}
	hash := crypto.HashPassword(u.PasswordHash, salt)

	created, err := r.client.User.Create().
		SetUsername(u.Username).
		SetPasswordHash(hash).
		SetSalt(salt).
		SetIsMember(u.IsMember).
		SetMemberLevel(int16(u.MemberLevel)).
		SetFreeTrialCount(u.FreeTrialCount).
		SetAvatar(u.Avatar).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return err
	}

	u.ID = created.ID
	return nil
}

// FindByUsername 根据用户名查找用户
func (r *userRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	u, err := r.client.User.Query().
		Where(user.Username(username)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return toBizUser(u), nil
}

// FindByID 根据 ID 查找用户
func (r *userRepo) FindByID(ctx context.Context, id int32) (*biz.User, error) {
	u, err := r.client.User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return toBizUser(u), nil
}

// UpdateAvatar 更新头像
func (r *userRepo) UpdateAvatar(ctx context.Context, id int32, avatar string) error {
	n, err := r.client.User.UpdateOneID(id).
		SetAvatar(avatar).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return biz.ErrUserNotFound
		}
		return err
	}
	_ = n
	return nil
}
