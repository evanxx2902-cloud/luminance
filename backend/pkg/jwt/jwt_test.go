package jwt

import (
	"testing"
)

func TestGenerateAccessToken(t *testing.T) {
	token, jti, err := GenerateAccessToken(1, "testuser")
	if err != nil {
		t.Fatalf("GenerateAccessToken() 返回错误: %v", err)
	}
	if token == "" {
		t.Error("token 不应为空")
	}
	if jti == "" {
		t.Error("jti 不应为空")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token, jti, err := GenerateRefreshToken(1)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() 返回错误: %v", err)
	}
	if token == "" {
		t.Error("token 不应为空")
	}
	if jti == "" {
		t.Error("jti 不应为空")
	}
}

func TestParseToken(t *testing.T) {
	// 测试 access token 解析
	tokenStr, jti, err := GenerateAccessToken(42, "alice")
	if err != nil {
		t.Fatalf("生成 token 失败: %v", err)
	}

	claims, err := ParseToken(tokenStr)
	if err != nil {
		t.Fatalf("ParseToken() 返回错误: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("UserID = %d, 期望 42", claims.UserID)
	}
	if claims.Username != "alice" {
		t.Errorf("Username = %s, 期望 alice", claims.Username)
	}
	if claims.Type != "access" {
		t.Errorf("Type = %s, 期望 access", claims.Type)
	}
	if claims.ID != jti {
		t.Errorf("JTI = %s, 期望 %s", claims.ID, jti)
	}

	// 测试无效 token
	_, err = ParseToken("invalid.token.string")
	if err == nil {
		t.Error("无效 token 应返回错误")
	}
}

func TestRefreshTokenClaims(t *testing.T) {
	tokenStr, _, err := GenerateRefreshToken(99)
	if err != nil {
		t.Fatalf("生成 refresh token 失败: %v", err)
	}

	claims, err := ParseToken(tokenStr)
	if err != nil {
		t.Fatalf("解析 refresh token 失败: %v", err)
	}
	if claims.UserID != 99 {
		t.Errorf("UserID = %d, 期望 99", claims.UserID)
	}
	if claims.Type != "refresh" {
		t.Errorf("Type = %s, 期望 refresh", claims.Type)
	}
}

func TestUniqueJTI(t *testing.T) {
	_, jti1, _ := GenerateAccessToken(1, "user")
	_, jti2, _ := GenerateAccessToken(1, "user")
	if jti1 == jti2 {
		t.Error("每次生成的 JTI 应唯一")
	}
}
