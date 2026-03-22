package service

import (
	"context"
	"fmt"

	v1 "github.com/luminance/backend/api/luminance/v1"
	"github.com/luminance/backend/internal/biz"
)

// UserService 用户服务实现
type UserService struct {
	v1.UnimplementedUserServiceServer
	uc *biz.UserUseCase
}

// NewUserService 创建用户服务
func NewUserService(uc *biz.UserUseCase) *UserService {
	return &UserService{uc: uc}
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	// 获取客户端 IP
	clientIP := getClientIP(ctx)

	result, err := s.uc.Register(ctx, req.Username, req.Password, clientIP)
	if err != nil {
		return nil, mapBizError(err)
	}

	return &v1.RegisterResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		Profile:      toUserProfile(result.User),
	}, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	loginReq := &biz.LoginRequest{
		AuthType:   req.AuthType,
		Username:   req.Username,
		Password:   req.Password,
		WechatCode: req.WechatCode,
		ClientIP:   getClientIP(ctx),
	}
	if loginReq.AuthType == "" {
		loginReq.AuthType = "password" // 默认密码认证
	}

	result, err := s.uc.Login(ctx, loginReq)
	if err != nil {
		return nil, mapBizError(err)
	}

	return &v1.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		Profile:      toUserProfile(result.User),
	}, nil
}

// Refresh 刷新 Token
func (s *UserService) Refresh(ctx context.Context, req *v1.RefreshRequest) (*v1.RefreshResponse, error) {
	accessToken, expiresIn, err := s.uc.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, mapBizError(err)
	}

	return &v1.RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	}, nil
}

// Logout 用户登出
func (s *UserService) Logout(ctx context.Context, req *v1.LogoutRequest) (*v1.LogoutResponse, error) {
	// 从 context 获取 userID 和 token
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	token, err := getTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.uc.Logout(ctx, token, userID); err != nil {
		return nil, mapBizError(err)
	}

	return &v1.LogoutResponse{Success: true}, nil
}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(ctx context.Context, req *v1.GetProfileRequest) (*v1.UserProfile, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.uc.GetProfile(ctx, userID)
	if err != nil {
		return nil, mapBizError(err)
	}

	return toUserProfile(user), nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx context.Context, req *v1.UpdateProfileRequest) (*v1.UserProfile, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 目前只支持更新头像
	if req.Avatar != nil && *req.Avatar != "" {
		if err := s.uc.UpdateAvatar(ctx, userID, *req.Avatar); err != nil {
			return nil, mapBizError(err)
		}
	}

	// 返回更新后的资料
	user, err := s.uc.GetProfile(ctx, userID)
	if err != nil {
		return nil, mapBizError(err)
	}

	return toUserProfile(user), nil
}

// toUserProfile 转换为 proto UserProfile
func toUserProfile(user *biz.User) *v1.UserProfile {
	profile := &v1.UserProfile{
		Id:              int64(user.ID),
		Username:        user.Username,
		IsMember:        user.IsMember,
		MemberLevel:     user.MemberLevel,
		FreeTrialCount:  user.FreeTrialCount,
		Avatar:          user.Avatar,
		CreatedAt:       user.CreatedAt.Unix(),
	}

	if user.MemberExpireAt != nil {
		profile.MemberExpireAt = user.MemberExpireAt.Unix()
	}

	return profile
}

// mapBizError 将 biz 层错误映射为 gRPC 错误
func mapBizError(err error) error {
	// 简化处理，实际应使用 status.Error(codes.Code, message)
	return err
}

// getClientIP 从 context 获取客户端 IP
func getClientIP(ctx context.Context) string {
	// 实际应从 gRPC metadata 或 HTTP header 获取
	return ""
}

// getUserIDFromContext 从 context 获取用户 ID
func getUserIDFromContext(ctx context.Context) (int32, error) {
	// 实际应从 JWT claims 中获取
	return 0, fmt.Errorf("unauthorized")
}

// getTokenFromContext 从 context 获取 token
func getTokenFromContext(ctx context.Context) (string, error) {
	// 实际应从 HTTP header 或 gRPC metadata 获取
	return "", fmt.Errorf("no token")
}
