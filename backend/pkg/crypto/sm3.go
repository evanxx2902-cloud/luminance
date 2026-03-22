package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"

	"github.com/tjfoc/gmsm/sm3"
)

// GenerateSalt 生成 32 字节随机盐值，返回 hex 编码字符串
func GenerateSalt() (string, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

// HashPassword 使用 HMAC-SM3 计算密码哈希
func HashPassword(password, salt string) string {
	h := hmac.New(sm3.New, []byte(salt))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyPassword 验证密码
func VerifyPassword(password, salt, hash string) bool {
	return HashPassword(password, salt) == hash
}
