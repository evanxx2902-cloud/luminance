package biz

import (
	"context"
	"fmt"
)

// LoginRequest 登录请求
type LoginRequest struct {
	AuthType   string
	Username   string
	Password   string
	WechatCode string
	ClientIP   string
}

// Authenticator 认证接口（Strategy 模式）
type Authenticator interface {
	Type() string
	Authenticate(ctx context.Context, req *LoginRequest) (*User, error)
}

// AuthRegistry 认证器注册表
type AuthRegistry struct {
	authenticators map[string]Authenticator
}

// NewAuthRegistry 创建认证器注册表
func NewAuthRegistry() *AuthRegistry {
	return &AuthRegistry{
		authenticators: make(map[string]Authenticator),
	}
}

// Register 注册认证器
func (r *AuthRegistry) Register(a Authenticator) {
	r.authenticators[a.Type()] = a
}

// Authenticate 执行认证
func (r *AuthRegistry) Authenticate(ctx context.Context, req *LoginRequest) (*User, error) {
	auth, ok := r.authenticators[req.AuthType]
	if !ok {
		return nil, fmt.Errorf("unsupported auth type: %s", req.AuthType)
	}
	return auth.Authenticate(ctx, req)
}
