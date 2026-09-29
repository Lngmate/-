package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 加密用户密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 校验用户密码（兼容 bcrypt 加密密码与既有测试账号的纯文本明文密码）
func CheckPassword(plainPassword, hashedPassword string) bool {
	// 1. 尝试使用 bcrypt 解密对比
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err == nil {
		return true
	}

	// 2. 兼容初始种子数据的明文密码（如 "123"）
	return plainPassword == hashedPassword
}
