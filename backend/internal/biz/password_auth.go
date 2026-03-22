package biz

import (
	"context"
	"errors"

	"github.com/luminance/backend/pkg/crypto"
)

// PasswordAuthenticator 用户名密码认证器
type PasswordAuthenticator struct {
	userRepo UserRepo
}

// NewPasswordAuthenticator 创建密码认证器
func NewPasswordAuthenticator(userRepo UserRepo) *PasswordAuthenticator {
	return &PasswordAuthenticator{userRepo: userRepo}
}

// Type 返回认证类型
func (p *PasswordAuthenticator) Type() string {
	return "password"
}

// Authenticate 执行密码认证
func (p *PasswordAuthenticator) Authenticate(ctx context.Context, req *LoginRequest) (*User, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("username and password are required")
	}

	// 查找用户
	user, err := p.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 验证密码
	if !crypto.VerifyPassword(req.Password, user.Salt, user.PasswordHash) {
		return nil, errors.New("invalid password")
	}

	return user, nil
}
