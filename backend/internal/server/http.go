package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"

	"github.com/luminance/backend/internal/service"
	"github.com/luminance/backend/pkg/jwt"
)

// contextKey 用于存储在 context 中的键
type contextKey string

const (
	userIDKey contextKey = "user_id"
	tokenKey  contextKey = "token"
	claimsKey contextKey = "claims"
)

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(userService *service.UserService) *kratoshttp.Server {
	srv := kratoshttp.NewServer(
		kratoshttp.Address(":8000"),
		kratoshttp.Middleware(
			JWTAuthMiddleware(),
		),
	)

	// TODO: 注册 HTTP 路由
	// 注意：Kratos 的 HTTP 路由注册需要通过其 Router 接口
	// 这里暂时留空，待 main.go 中通过 srv.Route() 注册

	return srv
}

// JWTAuthMiddleware JWT 认证中间件
func JWTAuthMiddleware() middleware.Middleware {
	// 白名单路径
	whitelist := map[string]bool{
		"/api/v1/auth/register": true,
		"/api/v1/auth/login":    true,
		"/api/v1/auth/refresh":  true,
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 获取 HTTP 请求信息
			if tr, ok := http.RequestFromServerContext(ctx); ok {
				path := tr.URL.Path

				// 白名单跳过认证
				if whitelist[path] {
					return handler(ctx, req)
				}

				// 提取 token
				authHeader := tr.Header.Get("Authorization")
				if authHeader == "" {
					return nil, fmt.Errorf("missing authorization header")
				}

				// 解析 Bearer token
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
					return nil, fmt.Errorf("invalid authorization format")
				}
				tokenStr := parts[1]

				// 验证 token
				claims, err := jwt.ParseToken(tokenStr)
				if err != nil {
					return nil, fmt.Errorf("invalid token: %w", err)
				}

				// 将 claims 存入 context
				ctx = context.WithValue(ctx, userIDKey, claims.UserID)
				ctx = context.WithValue(ctx, tokenKey, tokenStr)
				ctx = context.WithValue(ctx, claimsKey, claims)
			}

			return handler(ctx, req)
		}
	}
}

// GetUserIDFromContext 从 context 获取用户 ID
func GetUserIDFromContext(ctx context.Context) (int32, bool) {
	userID, ok := ctx.Value(userIDKey).(int32)
	return userID, ok
}

// GetTokenFromContext 从 context 获取 token
func GetTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenKey).(string)
	return token, ok
}

// GetClaimsFromContext 从 context 获取 claims
func GetClaimsFromContext(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*jwt.Claims)
	return claims, ok
}
