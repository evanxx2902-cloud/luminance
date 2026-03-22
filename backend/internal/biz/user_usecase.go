package biz

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/luminance/backend/pkg/jwt"
)

// 常量定义
const (
	MaxAvatarSize    = 1 << 20 // 1MB
	MaxLoginFailures = 5
	LockoutDuration  = 15 * time.Minute
	MaxRegisterPerIP = 3
	RegisterWindow   = 15 * time.Minute
)

// 错误定义
var (
	ErrInvalidUsername  = errors.New("invalid username format")
	ErrInvalidPassword  = errors.New("invalid password format")
	ErrUserExists       = errors.New("user already exists")
	ErrUserNotFound     = errors.New("user not found")
	ErrAccountLocked    = errors.New("account locked due to too many failed attempts")
	ErrIPBlocked        = errors.New("IP blocked due to too many requests")
	ErrTooManyRegisters = errors.New("too many registration attempts from this IP")
	ErrInvalidAvatar    = errors.New("invalid avatar data")
	ErrAvatarTooLarge   = errors.New("avatar size exceeds 1MB")
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenBlacklisted = errors.New("token has been revoked")
)

// RedisClient 接口（用于解耦）
type RedisClient interface {
	IncrLoginFail(ctx context.Context, key string) (int64, error)
	IsLoginLocked(ctx context.Context, key string) (bool, error)
	ClearLoginFail(ctx context.Context, key string) error
	StoreRefreshToken(ctx context.Context, uid, jti string, ttl time.Duration) error
	CheckRefreshToken(ctx context.Context, uid, jti string) (bool, error)
	RevokeRefreshToken(ctx context.Context, uid, jti string) error
	RevokeAllRefreshTokens(ctx context.Context, uid string) error
	BlacklistAccessToken(ctx context.Context, jti string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
	IncrRegisterCount(ctx context.Context, ip string) (int64, error)
	GetRegisterCount(ctx context.Context, ip string) (int64, error)
}

// UserUseCase 用户用例
type UserUseCase struct {
	repo    UserRepo
	redis   RedisClient
	authReg *AuthRegistry
}

// NewUserUseCase 创建用户用例
func NewUserUseCase(repo UserRepo, redis RedisClient, authReg *AuthRegistry) *UserUseCase {
	return &UserUseCase{
		repo:    repo,
		redis:   redis,
		authReg: authReg,
	}
}

// RegisterResult 注册结果
type RegisterResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         *User
}

// LoginResult 登录结果
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         *User
}

// 用户名格式：3-32位字母数字下划线
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

// Register 用户注册
func (uc *UserUseCase) Register(ctx context.Context, username, password, clientIP string) (*RegisterResult, error) {
	// 校验用户名
	if !usernameRegex.MatchString(username) {
		return nil, ErrInvalidUsername
	}

	// 校验密码长度
	if len(password) < 8 || len(password) > 64 {
		return nil, ErrInvalidPassword
	}

	// 检查 IP 注册频率
	count, err := uc.redis.GetRegisterCount(ctx, clientIP)
	if err != nil {
		return nil, fmt.Errorf("failed to check register count: %w", err)
	}
	if count >= MaxRegisterPerIP {
		return nil, ErrTooManyRegisters
	}

	// 创建用户（密码哈希在 repo 层处理）
	user := &User{
		Username:       username,
		PasswordHash:   password, // 临时存储，repo 会处理哈希
		IsMember:       false,
		MemberLevel:    0,
		FreeTrialCount: 1,
		CreatedAt:      time.Now(),
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 增加注册计数
	if _, err := uc.redis.IncrRegisterCount(ctx, clientIP); err != nil {
		fmt.Printf("failed to incr register count: %v\n", err)
	}

	// 生成 JWT
	return uc.generateTokens(ctx, user)
}

// Login 用户登录
func (uc *UserUseCase) Login(ctx context.Context, req *LoginRequest) (*LoginResult, error) {
	// 检查 IP 是否被封禁
	ipKey := fmt.Sprintf("auth:fail:ip:%s", req.ClientIP)
	locked, err := uc.redis.IsLoginLocked(ctx, ipKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check ip lock: %w", err)
	}
	if locked {
		return nil, ErrIPBlocked
	}

	// 检查账号是否被锁定（在认证前检查）
	if req.Username != "" {
		userKey := fmt.Sprintf("auth:fail:%s", req.Username)
		locked, err := uc.redis.IsLoginLocked(ctx, userKey)
		if err != nil {
			return nil, fmt.Errorf("failed to check user lock: %w", err)
		}
		if locked {
			return nil, ErrAccountLocked
		}
	}

	// 执行认证
	user, err := uc.authReg.Authenticate(ctx, req)
	if err != nil {
		// 认证失败，记录失败次数
		if req.Username != "" {
			userKey := fmt.Sprintf("auth:fail:%s", req.Username)
			if _, err := uc.redis.IncrLoginFail(ctx, userKey); err != nil {
				fmt.Printf("failed to incr login fail: %v\n", err)
			}
		}
		if _, err := uc.redis.IncrLoginFail(ctx, ipKey); err != nil {
			fmt.Printf("failed to incr ip fail: %v\n", err)
		}
		return nil, err
	}

	// 清除失败计数
	userKey := fmt.Sprintf("auth:fail:%s", user.Username)
	if err := uc.redis.ClearLoginFail(ctx, userKey); err != nil {
		fmt.Printf("failed to clear login fail: %v\n", err)
	}
	if err := uc.redis.ClearLoginFail(ctx, ipKey); err != nil {
		fmt.Printf("failed to clear ip fail: %v\n", err)
	}

	// 生成 JWT
	result, err := uc.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		User:         result.User,
	}, nil
}

// Refresh 刷新 Access Token
func (uc *UserUseCase) Refresh(ctx context.Context, refreshToken string) (string, int64, error) {
	// 解析 Refresh Token
	claims, err := jwt.ParseToken(refreshToken)
	if err != nil {
		return "", 0, ErrInvalidToken
	}

	// 验证类型
	if claims.Type != "refresh" {
		return "", 0, ErrInvalidToken
	}

	// 检查是否在 Redis 中
	exists, err := uc.redis.CheckRefreshToken(ctx, fmt.Sprintf("%d", claims.UserID), claims.ID)
	if err != nil {
		return "", 0, fmt.Errorf("failed to check refresh token: %w", err)
	}
	if !exists {
		return "", 0, ErrInvalidToken
	}

	// 生成新的 Access Token
	accessToken, _, err := jwt.GenerateAccessToken(claims.UserID, claims.Username)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, int64(jwt.AccessTokenExpiry.Seconds()), nil
}

// Logout 用户登出
func (uc *UserUseCase) Logout(ctx context.Context, accessToken string, userID int32) error {
	// 解析 Access Token 获取 JTI
	claims, err := jwt.ParseToken(accessToken)
	if err != nil {
		return ErrInvalidToken
	}

	// 计算剩余有效期
	remainingTTL := time.Until(claims.ExpiresAt.Time)
	if remainingTTL > 0 {
		// 加入黑名单
		if err := uc.redis.BlacklistAccessToken(ctx, claims.ID, remainingTTL); err != nil {
			return fmt.Errorf("failed to blacklist token: %w", err)
		}
	}

	// 撤销该用户的所有 Refresh Token
	if err := uc.redis.RevokeAllRefreshTokens(ctx, fmt.Sprintf("%d", userID)); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	return nil
}

// GetProfile 获取用户资料
func (uc *UserUseCase) GetProfile(ctx context.Context, userID int32) (*User, error) {
	return uc.repo.FindByID(ctx, userID)
}

// UpdateAvatar 更新头像
func (uc *UserUseCase) UpdateAvatar(ctx context.Context, userID int32, avatarBase64 string) error {
	// 解码 base64 检查大小
	data, err := base64.StdEncoding.DecodeString(avatarBase64)
	if err != nil {
		return ErrInvalidAvatar
	}

	if len(data) > MaxAvatarSize {
		return ErrAvatarTooLarge
	}

	return uc.repo.UpdateAvatar(ctx, userID, avatarBase64)
}

// CheckTokenBlacklist 检查 Token 是否在黑名单中
func (uc *UserUseCase) CheckTokenBlacklist(ctx context.Context, jti string) (bool, error) {
	return uc.redis.IsBlacklisted(ctx, jti)
}

// generateTokens 生成双 Token
func (uc *UserUseCase) generateTokens(ctx context.Context, user *User) (*RegisterResult, error) {
	// 生成 Access Token
	accessToken, _, err := jwt.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 生成 Refresh Token
	refreshToken, refreshJTI, err := jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 存储 Refresh Token
	if err := uc.redis.StoreRefreshToken(ctx, fmt.Sprintf("%d", user.ID), refreshJTI, jwt.RefreshTokenExpiry); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &RegisterResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(jwt.AccessTokenExpiry.Seconds()),
		User:         user,
	}, nil
}
