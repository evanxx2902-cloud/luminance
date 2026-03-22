package crypto

import (
	"testing"
)

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() 返回错误: %v", err)
	}
	if len(salt1) != 64 { // 32 字节 hex 编码 = 64 字符
		t.Errorf("GenerateSalt() 长度 = %d, 期望 64", len(salt1))
	}

	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() 返回错误: %v", err)
	}
	if salt1 == salt2 {
		t.Error("两次 GenerateSalt() 返回相同的盐值，随机性有问题")
	}
}

func TestHashPassword(t *testing.T) {
	salt := "testsalt"
	hash1 := HashPassword("mypassword", salt)
	hash2 := HashPassword("mypassword", salt)
	if hash1 != hash2 {
		t.Error("相同密码和盐值应产生相同哈希")
	}

	hash3 := HashPassword("otherpassword", salt)
	if hash1 == hash3 {
		t.Error("不同密码应产生不同哈希")
	}

	hash4 := HashPassword("mypassword", "othersalt")
	if hash1 == hash4 {
		t.Error("不同盐值应产生不同哈希")
	}

	if len(hash1) == 0 {
		t.Error("哈希不应为空")
	}
}

func TestVerifyPassword(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() 错误: %v", err)
	}

	password := "testpassword123"
	hash := HashPassword(password, salt)

	if !VerifyPassword(password, salt, hash) {
		t.Error("VerifyPassword() 正确密码应返回 true")
	}

	if VerifyPassword("wrongpassword", salt, hash) {
		t.Error("VerifyPassword() 错误密码应返回 false")
	}

	if VerifyPassword(password, "wrongsalt", hash) {
		t.Error("VerifyPassword() 错误盐值应返回 false")
	}
}
