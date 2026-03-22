package data

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Key 设计（前缀统一，便于扫描和清理）：
//
//	auth:fail:{username}       -> String(count), TTL 15m
//	auth:fail:ip:{ip}          -> String(count), TTL 15m
//	auth:register:ip:{ip}      -> String(count), TTL 15m
//	auth:refresh:{uid}:{jti}   -> String("1"),   TTL 30d
//	auth:blacklist:{jti}       -> String("1"),   TTL 2h（与 Access Token 一致）

const (
	keyLoginFailUser   = "auth:fail:%s"
	keyLoginFailIP     = "auth:fail:ip:%s"
	keyRegisterIP      = "auth:register:ip:%s"
	keyRefreshToken    = "auth:refresh:%s:%s"
	keyBlacklist       = "auth:blacklist:%s"
	loginLockThreshold = 5 // 连续失败 5 次即锁定
)

// RedisClient 封装 go-redis 客户端，提供认证相关的原子操作。
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient 创建并返回一个连接到 addr 的 RedisClient。
// addr 格式：host:port，例如 "127.0.0.1:6379"。
func NewRedisClient(addr string) *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

// Ping 检测 Redis 连通性，供健康检查调用。
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// ---------------------------------------------------------------------------
// 登录失败计数
// ---------------------------------------------------------------------------

// IncrLoginFail 将指定 key 的登录失败计数加 1；若 key 不存在则创建并设置 TTL 15 分钟。
// key 可以是用户名或 IP（调用方自行传入格式化后的 key）。
func (r *RedisClient) IncrLoginFail(ctx context.Context, key string) (int64, error) {
	pipe := r.client.TxPipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 15*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("redis IncrLoginFail: %w", err)
	}
	return incrCmd.Val(), nil
}

// IncrLoginFailByUsername 封装按用户名维度的失败计数。
func (r *RedisClient) IncrLoginFailByUsername(ctx context.Context, username string) (int64, error) {
	return r.IncrLoginFail(ctx, fmt.Sprintf(keyLoginFailUser, username))
}

// IncrLoginFailByIP 封装按 IP 维度的失败计数。
func (r *RedisClient) IncrLoginFailByIP(ctx context.Context, ip string) (int64, error) {
	return r.IncrLoginFail(ctx, fmt.Sprintf(keyLoginFailIP, ip))
}

// IsLoginLocked 判断指定 key 的失败计数是否已达锁定阈值（>= loginLockThreshold）。
func (r *RedisClient) IsLoginLocked(ctx context.Context, key string) (bool, error) {
	val, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("redis IsLoginLocked: %w", err)
	}
	return val >= loginLockThreshold, nil
}

// ClearLoginFail 清除指定 key 的登录失败记录（登录成功后调用）。
func (r *RedisClient) ClearLoginFail(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis ClearLoginFail: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Refresh Token 管理
// ---------------------------------------------------------------------------

// StoreRefreshToken 存储 Refresh Token 标识（uid + jti 组合为 key），value 固定为 "1"。
func (r *RedisClient) StoreRefreshToken(ctx context.Context, uid, jti string, ttl time.Duration) error {
	key := fmt.Sprintf(keyRefreshToken, uid, jti)
	if err := r.client.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("redis StoreRefreshToken: %w", err)
	}
	return nil
}

// CheckRefreshToken 验证 Refresh Token 是否存在且有效（未被撤销）。
func (r *RedisClient) CheckRefreshToken(ctx context.Context, uid, jti string) (bool, error) {
	key := fmt.Sprintf(keyRefreshToken, uid, jti)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("redis CheckRefreshToken: %w", err)
	}
	return val == "1", nil
}

// RevokeRefreshToken 撤销（删除）指定的 Refresh Token，用于登出或令牌轮换。
func (r *RedisClient) RevokeRefreshToken(ctx context.Context, uid, jti string) error {
	key := fmt.Sprintf(keyRefreshToken, uid, jti)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis RevokeRefreshToken: %w", err)
	}
	return nil
}

// RevokeAllRefreshTokens 撤销某个用户的所有 Refresh Token（强制下线）。
// 注意：SCAN 是 O(N) 操作，适合低频场景（如密码修改、安全事件）。
func (r *RedisClient) RevokeAllRefreshTokens(ctx context.Context, uid string) error {
	pattern := fmt.Sprintf(keyRefreshToken, uid, "*")
	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("redis RevokeAllRefreshTokens scan: %w", err)
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("redis RevokeAllRefreshTokens del: %w", err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Access Token 黑名单
// ---------------------------------------------------------------------------

// BlacklistAccessToken 将已注销的 Access Token 的 jti 加入黑名单，TTL 与 Token 剩余有效期一致。
func (r *RedisClient) BlacklistAccessToken(ctx context.Context, jti string, ttl time.Duration) error {
	key := fmt.Sprintf(keyBlacklist, jti)
	if err := r.client.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("redis BlacklistAccessToken: %w", err)
	}
	return nil
}

// IsBlacklisted 检查 Access Token 的 jti 是否在黑名单中。
func (r *RedisClient) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf(keyBlacklist, jti)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("redis IsBlacklisted: %w", err)
	}
	return val == "1", nil
}

// ---------------------------------------------------------------------------
// 注册频率限制
// ---------------------------------------------------------------------------

// IncrRegisterCount 将指定 IP 的注册次数加 1；首次创建时设置 TTL 15 分钟。
func (r *RedisClient) IncrRegisterCount(ctx context.Context, ip string) (int64, error) {
	key := fmt.Sprintf(keyRegisterIP, ip)
	pipe := r.client.TxPipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 15*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("redis IncrRegisterCount: %w", err)
	}
	return incrCmd.Val(), nil
}

// GetRegisterCount 获取指定 IP 在当前窗口期内的注册次数。
func (r *RedisClient) GetRegisterCount(ctx context.Context, ip string) (int64, error) {
	key := fmt.Sprintf(keyRegisterIP, ip)
	val, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("redis GetRegisterCount: %w", err)
	}
	return val, nil
}

// Close 关闭 Redis 连接。
func (r *RedisClient) Close() error {
	return r.client.Close()
}
