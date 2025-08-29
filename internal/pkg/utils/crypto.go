package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// 密码强度等级常量
const (
	PasswordWeek   = 1 // 弱密码
	PasswordMedium = 2 // 中等强度密码
	PasswordStrong = 3 // 强密码
)

// BCrypt成本因子常量
const (
	MinCost     = 4  // 最小成本（测试用）
	DefaultCost = 12 // 默认成本（推荐）
	MaxCost     = 31 // 最大成本
)

// PasswordHasher 密码哈希器接口
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, plainPassword string) bool
	ValidatePasswordStrength(password string) (int, error)
	GenerateSecurePassword(length int) (string, error)
}

// bcryptHasher BCrypt密码哈希器实现
type bcryptHasher struct {
	cost int
}

// NewPasswordHasher 创建新的密码哈希器
func NewPasswordHasher(cost int) PasswordHasher {
	if cost < MinCost || cost > MaxCost {
		cost = DefaultCost
	}
	return &bcryptHasher{cost: cost}
}

// NewDefaultPasswordHasher 创建默认的密码哈希器
func NewDefaultPasswordHasher() PasswordHasher {
	return &bcryptHasher{cost: DefaultCost}
}

// HashPassword 使用BCrypt加密密码
func (h *bcryptHasher) HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", fmt.Errorf("密码不能为空")
	}

	// 使用BCrypt加密密码
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}

	return string(bytes), nil
}

// VerifyPassword 验证密码是否正确
func (h *bcryptHasher) VerifyPassword(hashedPassword, plainPassword string) bool {
	if len(hashedPassword) == 0 || len(plainPassword) == 0 {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// checkPasswordComplexity 检查密码复杂度
func checkPasswordComplexity(password string) int {
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	complexityCount := 0
	if hasUpper {
		complexityCount++
	}
	if hasLower {
		complexityCount++
	}
	if hasDigit {
		complexityCount++
	}
	if hasSpecial {
		complexityCount++
	}
	return complexityCount
}

// checkWeakPasswords 检查常见弱密码
func checkWeakPasswords(password string) error {
	weakPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123",
		"password123", "admin", "root", "user", "test", "guest",
		"111111", "000000", "654321", "123123", "987654321",
	}

	passwordLower := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if strings.Contains(passwordLower, weak) {
			return fmt.Errorf("密码过于简单，不能包含常见的弱密码模式")
		}
	}
	return nil
}

// checkPasswordPatterns 检查密码模式
func checkPasswordPatterns(password string) error {
	// 检查重复字符
	if hasRepeatedChars(password, 3) {
		return fmt.Errorf("密码不能包含3个或更多连续相同的字符")
	}

	// 检查顺序字符
	if hasSequentialChars(password, 4) {
		return fmt.Errorf("密码不能包含4个或更多连续的顺序字符")
	}
	return nil
}

// calculatePasswordStrength 计算密码强度
func calculatePasswordStrength(password string, complexityCount int) int {
	// 根据复杂度和长度判断密码强度
	if len(password) >= 12 && complexityCount >= 3 {
		return PasswordStrong
	}
	if len(password) >= 8 && complexityCount >= 2 {
		return PasswordMedium
	}
	return PasswordWeek
}

// ValidatePasswordStrength 验证密码强度
func (h *bcryptHasher) ValidatePasswordStrength(password string) (int, error) {
	if len(password) < 6 {
		return PasswordWeek, fmt.Errorf("密码长度至少6位")
	}

	if len(password) > 128 {
		return PasswordWeek, fmt.Errorf("密码长度不能超过128位")
	}

	// 检查密码复杂度
	complexityCount := checkPasswordComplexity(password)

	// 检查常见弱密码
	if err := checkWeakPasswords(password); err != nil {
		return PasswordWeek, err
	}

	// 检查密码模式
	if err := checkPasswordPatterns(password); err != nil {
		return PasswordWeek, err
	}

	// 计算密码强度
	strength := calculatePasswordStrength(password, complexityCount)
	if strength == PasswordWeek {
		return PasswordWeek, fmt.Errorf("密码强度不足，建议包含大小写字母、数字和特殊字符")
	}

	return strength, nil
}

// generatePasswordCharsets 获取密码生成字符集
func generatePasswordCharsets() (string, string, string, string) {
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits := "0123456789"
	special := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	return lowercase, uppercase, digits, special
}

// addRequiredChars 添加必需的字符类型
func addRequiredChars(password *strings.Builder, lowercase, uppercase, digits, special string) error {
	// 至少一个小写字母
	if char, err := randomChar(lowercase); err == nil {
		password.WriteByte(char)
	} else {
		return fmt.Errorf("生成小写字母失败: %w", err)
	}

	// 至少一个大写字母
	if char, err := randomChar(uppercase); err == nil {
		password.WriteByte(char)
	} else {
		return fmt.Errorf("生成大写字母失败: %w", err)
	}

	// 至少一个数字
	if char, err := randomChar(digits); err == nil {
		password.WriteByte(char)
	} else {
		return fmt.Errorf("生成数字失败: %w", err)
	}

	// 至少一个特殊字符
	if char, err := randomChar(special); err == nil {
		password.WriteByte(char)
	} else {
		return fmt.Errorf("生成特殊字符失败: %w", err)
	}

	return nil
}

// fillRemainingChars 填充剩余字符
func fillRemainingChars(password *strings.Builder, length int, allChars string) error {
	for password.Len() < length {
		if char, err := randomChar(allChars); err == nil {
			password.WriteByte(char)
		} else {
			return fmt.Errorf("生成密码失败: %w", err)
		}
	}
	return nil
}

// shufflePassword 打乱密码字符顺序
func shufflePassword(password string) (string, error) {
	result := []byte(password)
	for i := len(result) - 1; i > 0; i-- {
		if j, err := randomInt(i + 1); err == nil {
			result[i], result[j] = result[j], result[i]
		} else {
			return "", fmt.Errorf("打乱密码顺序失败: %w", err)
		}
	}
	return string(result), nil
}

// GenerateSecurePassword 生成安全的密码
func (h *bcryptHasher) GenerateSecurePassword(length int) (string, error) {
	if length < 8 {
		length = 12 // 默认12位
	}
	if length > 128 {
		length = 128 // 最大128位
	}

	// 获取字符集
	lowercase, uppercase, digits, special := generatePasswordCharsets()

	// 确保至少包含每种字符类型
	var password strings.Builder
	if err := addRequiredChars(&password, lowercase, uppercase, digits, special); err != nil {
		return "", err
	}

	// 填充剩余长度
	allChars := lowercase + uppercase + digits + special
	if err := fillRemainingChars(&password, length, allChars); err != nil {
		return "", err
	}

	// 打乱字符顺序
	return shufflePassword(password.String())
}

// 全局便利函数

// HashPassword 加密密码（使用默认哈希器）
func HashPassword(password string) (string, error) {
	hasher := NewDefaultPasswordHasher()
	return hasher.HashPassword(password)
}

// VerifyPassword 验证密码（使用默认哈希器）
func VerifyPassword(hashedPassword, plainPassword string) bool {
	hasher := NewDefaultPasswordHasher()
	return hasher.VerifyPassword(hashedPassword, plainPassword)
}

// ValidatePasswordStrength 验证密码强度（使用默认哈希器）
func ValidatePasswordStrength(password string) (int, error) {
	hasher := NewDefaultPasswordHasher()
	return hasher.ValidatePasswordStrength(password)
}

// GenerateSecurePassword 生成安全密码（使用默认哈希器）
func GenerateSecurePassword(length int) (string, error) {
	hasher := NewDefaultPasswordHasher()
	return hasher.GenerateSecurePassword(length)
}

// ComparePasswords 安全地比较两个密码是否相同（防止时序攻击）
func ComparePasswords(password1, password2 string) bool {
	return subtle.ConstantTimeCompare([]byte(password1), []byte(password2)) == 1
}

// GenerateSalt 生成随机盐值
func GenerateSalt(length int) (string, error) {
	if length <= 0 {
		length = 32
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成盐值失败: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// 辅助函数

// randomChar 生成指定字符集中的随机字符
func randomChar(charset string) (byte, error) {
	if len(charset) == 0 {
		return 0, fmt.Errorf("字符集不能为空")
	}

	bytes := make([]byte, 1)
	if _, err := rand.Read(bytes); err != nil {
		return 0, err
	}

	return charset[int(bytes[0])%len(charset)], nil
}

// randomInt 生成0到max-1之间的随机整数
func randomInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("max必须大于0")
	}

	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return 0, err
	}

	// 将字节转换为无符号整数
	num := uint32(bytes[0])<<24 | uint32(bytes[1])<<16 | uint32(bytes[2])<<8 | uint32(bytes[3])
	return int(num) % max, nil
}

// hasRepeatedChars 检查是否有重复字符
func hasRepeatedChars(password string, maxRepeat int) bool {
	if maxRepeat <= 0 {
		return false
	}

	count := 1
	for i := 1; i < len(password); i++ {
		if password[i] == password[i-1] {
			count++
			if count >= maxRepeat {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

// hasSequentialChars 检查是否有顺序字符
func hasSequentialChars(password string, maxSequential int) bool {
	if maxSequential <= 0 || len(password) < maxSequential {
		return false
	}

	for i := 0; i <= len(password)-maxSequential; i++ {
		// 检查递增序列
		isAscending := true
		for j := 1; j < maxSequential; j++ {
			if password[i+j] != password[i+j-1]+1 {
				isAscending = false
				break
			}
		}
		if isAscending {
			return true
		}

		// 检查递减序列
		isDescending := true
		for j := 1; j < maxSequential; j++ {
			if password[i+j] != password[i+j-1]-1 {
				isDescending = false
				break
			}
		}
		if isDescending {
			return true
		}
	}
	return false
}
