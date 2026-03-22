package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/luminance/backend/internal/biz"
	"github.com/luminance/backend/pkg/crypto"
)

// userRepo 用户仓库实现
type userRepo struct {
	db *sql.DB
}

// NewUserRepo 创建用户仓库
func NewUserRepo(db *sql.DB) biz.UserRepo {
	return &userRepo{db: db}
}

// Create 创建用户
func (r *userRepo) Create(ctx context.Context, user *biz.User) error {
	// 生成盐值并哈希密码
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return err
	}
	hash := crypto.HashPassword(user.PasswordHash, salt)

	query := `
		INSERT INTO users (username, password_hash, salt, is_member, member_level,
						   free_trial_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	now := time.Now()
	return r.db.QueryRowContext(ctx, query,
		user.Username,
		hash,
		salt,
		user.IsMember,
		user.MemberLevel,
		user.FreeTrialCount,
		now,
		now,
	).Scan(&user.ID)
}

// FindByUsername 根据用户名查找用户
func (r *userRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	query := `
		SELECT id, username, password_hash, salt, is_member, member_level,
			   member_expire_at, free_trial_count, avatar, created_at
		FROM users
		WHERE username = $1
	`
	user := &biz.User{}
	var memberExpireAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Salt,
		&user.IsMember,
		&user.MemberLevel,
		&memberExpireAt,
		&user.FreeTrialCount,
		&user.Avatar,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	if memberExpireAt.Valid {
		user.MemberExpireAt = &memberExpireAt.Time
	}
	return user, nil
}

// FindByID 根据 ID 查找用户
func (r *userRepo) FindByID(ctx context.Context, id int32) (*biz.User, error) {
	query := `
		SELECT id, username, password_hash, salt, is_member, member_level,
			   member_expire_at, free_trial_count, avatar, created_at
		FROM users
		WHERE id = $1
	`
	user := &biz.User{}
	var memberExpireAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Salt,
		&user.IsMember,
		&user.MemberLevel,
		&memberExpireAt,
		&user.FreeTrialCount,
		&user.Avatar,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	if memberExpireAt.Valid {
		user.MemberExpireAt = &memberExpireAt.Time
	}
	return user, nil
}

// UpdateAvatar 更新头像
func (r *userRepo) UpdateAvatar(ctx context.Context, id int32, avatar string) error {
	query := `
		UPDATE users
		SET avatar = $1, updated_at = $2
		WHERE id = $3
	`
	result, err := r.db.ExecContext(ctx, query, avatar, time.Now(), id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return biz.ErrUserNotFound
	}
	return nil
}
