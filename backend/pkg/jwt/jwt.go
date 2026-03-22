package jwt

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenExpiry  = 2 * time.Hour           // 2 小时
	RefreshTokenExpiry = 30 * 24 * time.Hour     // 30 天
)

type Claims struct {
	UserID   int32  `json:"sub"`
	Username string `json:"username"`
	Type     string `json:"type"` // "access" | "refresh"
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成 Access Token，返回 token 字符串和 jti
func GenerateAccessToken(userID int32, username string) (string, string, error) {
	jti := uuid.New().String()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   string(rune(userID)),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(getSecret()))
	return tokenString, jti, err
}

// GenerateRefreshToken 生成 Refresh Token
func GenerateRefreshToken(userID int32) (string, string, error) {
	jti := uuid.New().String()
	claims := Claims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   string(rune(userID)),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(getSecret()))
	return tokenString, jti, err
}

// ParseToken 解析并验证 Token
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(getSecret()), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

// getSecret 从环境变量获取 JWT 密钥
func getSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "default-secret-change-in-production" // 仅开发用
	}
	return secret
}
