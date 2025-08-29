package utils

import (
	"strings"
	"testing"

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
