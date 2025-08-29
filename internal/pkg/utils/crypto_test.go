package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPasswordHasher(t *testing.T) {
	t.Run("正常创建密码哈希器", func(t *testing.T) {
		hasher := NewPasswordHasher(12)
		assert.NotNil(t, hasher)
	})

	t.Run("成本过低时使用默认值", func(t *testing.T) {
		hasher := NewPasswordHasher(3)
		assert.NotNil(t, hasher)
	})

	t.Run("成本过高时使用默认值", func(t *testing.T) {
		hasher := NewPasswordHasher(32)
		assert.NotNil(t, hasher)
	})
}

func TestNewDefaultPasswordHasher(t *testing.T) {
	hasher := NewDefaultPasswordHasher()
	assert.NotNil(t, hasher)
}

func TestHashPassword(t *testing.T) {
	hasher := NewDefaultPasswordHasher()

	t.Run("正常密码加密", func(t *testing.T) {
		password := "testPassword123!"
		hash, err := hasher.HashPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)
	})

	t.Run("空密码应该返回错误", func(t *testing.T) {
		hash, err := hasher.HashPassword("")
		assert.Error(t, err)
		assert.Empty(t, hash)
		assert.Contains(t, err.Error(), "密码不能为空")
	})
}

func TestVerifyPassword(t *testing.T) {
	hasher := NewDefaultPasswordHasher()
	password := "testPassword123!"
	hash, _ := hasher.HashPassword(password)

	t.Run("正确密码验证", func(t *testing.T) {
		result := hasher.VerifyPassword(hash, password)
		assert.True(t, result)
	})

	t.Run("错误密码验证", func(t *testing.T) {
		result := hasher.VerifyPassword(hash, "wrongPassword")
		assert.False(t, result)
	})

	t.Run("空哈希值验证", func(t *testing.T) {
		result := hasher.VerifyPassword("", password)
		assert.False(t, result)
	})

	t.Run("空密码验证", func(t *testing.T) {
		result := hasher.VerifyPassword(hash, "")
		assert.False(t, result)
	})
}

func TestValidatePasswordStrength(t *testing.T) {
	hasher := NewDefaultPasswordHasher()

	t.Run("强密码测试", func(t *testing.T) {
		password := "MySecure#Pass789!"
		strength, err := hasher.ValidatePasswordStrength(password)
		assert.NoError(t, err)
		assert.Equal(t, PasswordStrong, strength)
	})

	t.Run("中等强度密码测试", func(t *testing.T) {
		password := "MySecure9"
		strength, err := hasher.ValidatePasswordStrength(password)
		assert.NoError(t, err)
		assert.Equal(t, PasswordMedium, strength)
	})

	t.Run("弱密码长度不足", func(t *testing.T) {
		password := "123"
		strength, err := hasher.ValidatePasswordStrength(password)
		assert.Error(t, err)
		assert.Equal(t, PasswordWeek, strength)
		assert.Contains(t, err.Error(), "密码长度至少6位")
	})

	t.Run("密码过长", func(t *testing.T) {
		password := strings.Repeat("a", 129)
		strength, err := hasher.ValidatePasswordStrength(password)
		assert.Error(t, err)
		assert.Equal(t, PasswordWeek, strength)
		assert.Contains(t, err.Error(), "密码长度不能超过128位")
	})

	t.Run("包含弱密码模式", func(t *testing.T) {
		password := "password123"
		strength, err := hasher.ValidatePasswordStrength(password)
		assert.Error(t, err)
		assert.Equal(t, PasswordWeek, strength)
		assert.Contains(t, err.Error(), "密码过于简单")
	})

	t.Run("包含重复字符", func(t *testing.T) {
		password := "MySecuresss123!"
		strength, err := hasher.ValidatePasswordStrength(password)
		assert.Error(t, err)
		assert.Equal(t, PasswordWeek, strength)
		assert.Contains(t, err.Error(), "连续相同的字符")
	})

	t.Run("包含顺序字符", func(t *testing.T) {
		password := "MySecure1234!"
		strength, err := hasher.ValidatePasswordStrength(password)
		assert.Error(t, err)
		assert.Equal(t, PasswordWeek, strength)
		assert.Contains(t, err.Error(), "连续的顺序字符")
	})

	t.Run("强度不足", func(t *testing.T) {
		password := "simpleA"
		strength, err := hasher.ValidatePasswordStrength(password)
		assert.Error(t, err)
		assert.Equal(t, PasswordWeek, strength)
		assert.Contains(t, err.Error(), "密码强度不足")
	})
}

func TestGenerateSecurePassword(t *testing.T) {
	hasher := NewDefaultPasswordHasher()

	t.Run("生成默认长度密码", func(t *testing.T) {
		password, err := hasher.GenerateSecurePassword(12)
		assert.NoError(t, err)
		assert.Len(t, password, 12)
	})

	t.Run("生成最小长度密码", func(t *testing.T) {
		password, err := hasher.GenerateSecurePassword(6)
		assert.NoError(t, err)
		assert.Len(t, password, 12) // 应该使用默认长度12
	})

	t.Run("生成超长密码", func(t *testing.T) {
		password, err := hasher.GenerateSecurePassword(200)
		assert.NoError(t, err)
		assert.Len(t, password, 128) // 应该限制为最大长度128
	})

	t.Run("验证生成密码的复杂度", func(t *testing.T) {
		password, err := hasher.GenerateSecurePassword(16)
		assert.NoError(t, err)

		// 检查是否包含各种字符类型
		hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
		hasDigit := strings.ContainsAny(password, "0123456789")
		hasSpecial := strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?")

		assert.True(t, hasUpper, "密码应该包含大写字母")
		assert.True(t, hasLower, "密码应该包含小写字母")
		assert.True(t, hasDigit, "密码应该包含数字")
		assert.True(t, hasSpecial, "密码应该包含特殊字符")
	})
}

func TestGlobalPasswordFunctions(t *testing.T) {
	t.Run("全局HashPassword函数", func(t *testing.T) {
		password := "testPassword123!"
		hash, err := HashPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("全局VerifyPassword函数", func(t *testing.T) {
		password := "testPassword123!"
		hash, _ := HashPassword(password)
		result := VerifyPassword(hash, password)
		assert.True(t, result)
	})

	t.Run("全局ValidatePasswordStrength函数", func(t *testing.T) {
		password := "MySecure#Pass789!"
		strength, err := ValidatePasswordStrength(password)
		assert.NoError(t, err)
		assert.Equal(t, PasswordStrong, strength)
	})

	t.Run("全局GenerateSecurePassword函数", func(t *testing.T) {
		password, err := GenerateSecurePassword(12)
		assert.NoError(t, err)
		assert.Len(t, password, 12)
	})
}

func TestComparePasswords(t *testing.T) {
	t.Run("相同密码比较", func(t *testing.T) {
		password := "testPassword"
		result := ComparePasswords(password, password)
		assert.True(t, result)
	})

	t.Run("不同密码比较", func(t *testing.T) {
		password1 := "testPassword1"
		password2 := "testPassword2"
		result := ComparePasswords(password1, password2)
		assert.False(t, result)
	})
}

func TestGenerateSalt(t *testing.T) {
	t.Run("生成默认长度盐值", func(t *testing.T) {
		salt, err := GenerateSalt(32)
		assert.NoError(t, err)
		assert.NotEmpty(t, salt)
	})

	t.Run("生成零长度盐值使用默认长度", func(t *testing.T) {
		salt, err := GenerateSalt(0)
		assert.NoError(t, err)
		assert.NotEmpty(t, salt)
	})

	t.Run("生成负长度盐值使用默认长度", func(t *testing.T) {
		salt, err := GenerateSalt(-1)
		assert.NoError(t, err)
		assert.NotEmpty(t, salt)
	})
}

func TestPasswordComplexityHelpers(t *testing.T) {
	t.Run("检查密码复杂度", func(t *testing.T) {
		complexity := checkPasswordComplexity("Password123!")
		assert.Equal(t, 4, complexity)
	})

	t.Run("检查只有小写字母的复杂度", func(t *testing.T) {
		complexity := checkPasswordComplexity("password")
		assert.Equal(t, 1, complexity)
	})

	t.Run("检查弱密码", func(t *testing.T) {
		err := checkWeakPasswords("password123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "密码过于简单")
	})

	t.Run("检查非弱密码", func(t *testing.T) {
		err := checkWeakPasswords("MyStrongSecurePass")
		assert.NoError(t, err)
	})

	t.Run("检查密码模式", func(t *testing.T) {
		err := checkPasswordPatterns("MyPassword!")
		assert.NoError(t, err)
	})

	t.Run("计算密码强度", func(t *testing.T) {
		strength := calculatePasswordStrength("Password123!", 4)
		assert.Equal(t, PasswordStrong, strength)
	})
}

func TestPasswordCharsets(t *testing.T) {
	t.Run("获取密码字符集", func(t *testing.T) {
		lowercase, uppercase, digits, special := generatePasswordCharsets()
		assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", lowercase)
		assert.Equal(t, "ABCDEFGHIJKLMNOPQRSTUVWXYZ", uppercase)
		assert.Equal(t, "0123456789", digits)
		assert.Equal(t, "!@#$%^&*()_+-=[]{}|;:,.<>?", special)
	})
}

func TestPasswordConstants(t *testing.T) {
	t.Run("密码强度常量", func(t *testing.T) {
		assert.Equal(t, 1, PasswordWeek)
		assert.Equal(t, 2, PasswordMedium)
		assert.Equal(t, 3, PasswordStrong)
	})

	t.Run("BCrypt成本常量", func(t *testing.T) {
		assert.Equal(t, 4, MinCost)
		assert.Equal(t, 12, DefaultCost)
		assert.Equal(t, 31, MaxCost)
	})
}

// ==== JWT测试 ====

func TestNewJWTManager(t *testing.T) {
	t.Run("正常创建JWT管理器", func(t *testing.T) {
		secretKey := "this-is-a-very-long-secret-key-for-testing-jwt-manager"
		manager, err := NewJWTManager(secretKey, time.Hour, 24*time.Hour)
		assert.NoError(t, err)
		assert.NotNil(t, manager)
	})

	t.Run("密钥过短的情况", func(t *testing.T) {
		secretKey := "short"
		manager, err := NewJWTManager(secretKey, time.Hour, 24*time.Hour)
		assert.Error(t, err)
		assert.Nil(t, manager)
		assert.Contains(t, err.Error(), "密钥长度不能小于")
	})

	t.Run("使用默认过期时间", func(t *testing.T) {
		secretKey := "this-is-a-very-long-secret-key-for-testing-jwt-manager"
		manager, err := NewJWTManager(secretKey, 0, 0)
		assert.NoError(t, err)
		assert.NotNil(t, manager)
	})
}

func TestNewDefaultJWTManager(t *testing.T) {
	secretKey := "this-is-a-very-long-secret-key-for-testing-jwt-manager"
	manager, err := NewDefaultJWTManager(secretKey)
	assert.NoError(t, err)
	assert.NotNil(t, manager)
}

func TestJWTTokenGeneration(t *testing.T) {
	secretKey := "this-is-a-very-long-secret-key-for-testing-jwt-manager"
	manager, _ := NewDefaultJWTManager(secretKey)

	t.Run("生成访问令牌", func(t *testing.T) {
		token, err := manager.GenerateAccessToken(12345, "testuser", "test@example.com", "user")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Contains(t, token, ".")
	})

	t.Run("生成刷新令牌", func(t *testing.T) {
		token, err := manager.GenerateRefreshToken(12345, "testuser", "test@example.com", "user")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Contains(t, token, ".")
	})
}

func TestJWTTokenValidation(t *testing.T) {
	secretKey := "this-is-a-very-long-secret-key-for-testing-jwt-manager"
	manager, _ := NewDefaultJWTManager(secretKey)

	t.Run("验证有效的访问令牌", func(t *testing.T) {
		token, _ := manager.GenerateAccessToken(12345, "testuser", "test@example.com", "user")
		claims, err := manager.ValidateToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, uint64(12345), claims.UserID)
		assert.Equal(t, "testuser", claims.Username)
		assert.Equal(t, "test@example.com", claims.Email)
		assert.Equal(t, "user", claims.Role)
		assert.Equal(t, "access", claims.TokenType)
	})

	t.Run("验证无效令牌", func(t *testing.T) {
		claims, err := manager.ValidateToken("invalid-token")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "令牌解析失败")
	})

	t.Run("验证空令牌", func(t *testing.T) {
		claims, err := manager.ValidateToken("")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestJWTTokenRefresh(t *testing.T) {
	secretKey := "this-is-a-very-long-secret-key-for-testing-jwt-manager"
	manager, _ := NewDefaultJWTManager(secretKey)

	t.Run("使用有效刷新令牌刷新", func(t *testing.T) {
		refreshToken, _ := manager.GenerateRefreshToken(12345, "testuser", "test@example.com", "user")
		newAccessToken, newRefreshToken, err := manager.RefreshToken(refreshToken)
		assert.NoError(t, err)
		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)
		assert.NotEqual(t, refreshToken, newRefreshToken)
	})

	t.Run("使用访问令牌刷新应该失败", func(t *testing.T) {
		accessToken, _ := manager.GenerateAccessToken(12345, "testuser", "test@example.com", "user")
		newAccessToken, newRefreshToken, err := manager.RefreshToken(accessToken)
		assert.Error(t, err)
		assert.Empty(t, newAccessToken)
		assert.Empty(t, newRefreshToken)
		assert.Contains(t, err.Error(), "令牌类型错误")
	})

	t.Run("使用无效令牌刷新", func(t *testing.T) {
		newAccessToken, newRefreshToken, err := manager.RefreshToken("invalid-token")
		assert.Error(t, err)
		assert.Empty(t, newAccessToken)
		assert.Empty(t, newRefreshToken)
		assert.Contains(t, err.Error(), "刷新令牌无效")
	})
}

// ==== 随机字符串生成测试 ====

func TestGenerateVerificationCode(t *testing.T) {
	t.Run("生成默认长度验证码", func(t *testing.T) {
		code, err := GenerateVerificationCode(6)
		assert.NoError(t, err)
		assert.Len(t, code, 6)
		// 验证只包含数字
		for _, char := range code {
			assert.True(t, char >= '0' && char <= '9', "验证码应该只包含数字")
		}
	})

	t.Run("生成零长度使用默认长度", func(t *testing.T) {
		code, err := GenerateVerificationCode(0)
		assert.NoError(t, err)
		assert.Len(t, code, 6)
	})

	t.Run("超过最大长度限制", func(t *testing.T) {
		code, err := GenerateVerificationCode(15)
		assert.NoError(t, err)
		assert.Len(t, code, 10) // 应该限制为最大10位
	})
}

func TestGenerateRandomToken(t *testing.T) {
	t.Run("生成随机令牌", func(t *testing.T) {
		token, err := GenerateRandomToken(32)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("使用默认字节长度", func(t *testing.T) {
		token, err := GenerateRandomToken(0)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

// ==== AES加密测试 ====

func TestNewAESCrypto(t *testing.T) {
	crypto := NewAESCrypto()
	assert.NotNil(t, crypto)
}

func TestAESKeyGeneration(t *testing.T) {
	crypto := NewAESCrypto()

	t.Run("生成AES密钥", func(t *testing.T) {
		key, err := crypto.GenerateKey()
		assert.NoError(t, err)
		assert.NotEmpty(t, key)
	})

	t.Run("全局函数生成AES密钥", func(t *testing.T) {
		key, err := GenerateAESKey()
		assert.NoError(t, err)
		assert.NotEmpty(t, key)
	})
}

func TestAESEncryptDecrypt(t *testing.T) {
	crypto := NewAESCrypto()
	key, _ := crypto.GenerateKey()
	plaintext := "This is a test message for AES encryption"

	t.Run("AES加密解密", func(t *testing.T) {
		ciphertext, err := crypto.Encrypt(plaintext, key)
		assert.NoError(t, err)
		assert.NotEmpty(t, ciphertext)
		assert.NotEqual(t, plaintext, ciphertext)

		decrypted, err := crypto.Decrypt(ciphertext, key)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("全局AES加密解密函数", func(t *testing.T) {
		ciphertext, err := EncryptAES(plaintext, key)
		assert.NoError(t, err)
		assert.NotEmpty(t, ciphertext)

		decrypted, err := DecryptAES(ciphertext, key)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("使用错误密钥解密", func(t *testing.T) {
		wrongKey, _ := crypto.GenerateKey()
		ciphertext, _ := crypto.Encrypt(plaintext, key)
		decrypted, err := crypto.Decrypt(ciphertext, wrongKey)
		assert.Error(t, err)
		assert.Empty(t, decrypted)
	})

	t.Run("无效的密钥格式", func(t *testing.T) {
		ciphertext, err := crypto.Encrypt(plaintext, "invalid-key")
		assert.Error(t, err)
		assert.Empty(t, ciphertext)
	})

	t.Run("无效的密文格式", func(t *testing.T) {
		decrypted, err := crypto.Decrypt("invalid-ciphertext", key)
		assert.Error(t, err)
		assert.Empty(t, decrypted)
	})
}

// ==== 哈希算法测试 ====

func TestMD5Hash(t *testing.T) {
	t.Run("MD5哈希计算", func(t *testing.T) {
		data := "test data"
		hash := MD5Hash(data)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 32) // MD5哈希长度为32个字符
	})

	t.Run("空字符串MD5哈希", func(t *testing.T) {
		hash := MD5Hash("")
		assert.NotEmpty(t, hash)
		assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", hash)
	})
}

func TestSHA256Hash(t *testing.T) {
	t.Run("SHA256哈希计算", func(t *testing.T) {
		data := "test data"
		hash := SHA256Hash(data)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64) // SHA256哈希长度为64个字符
	})

	t.Run("空字符串SHA256哈希", func(t *testing.T) {
		hash := SHA256Hash("")
		assert.NotEmpty(t, hash)
		assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", hash)
	})
}

func TestSHA256HashWithSalt(t *testing.T) {
	t.Run("带盐的SHA256哈希", func(t *testing.T) {
		data := "test data"
		salt := "salt123"
		hash := SHA256HashWithSalt(data, salt)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64)

		// 验证与不同盐值的哈希不同
		differentSaltHash := SHA256HashWithSalt(data, "different-salt")
		assert.NotEqual(t, hash, differentSaltHash)
	})
}
