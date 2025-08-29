package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 常用字符集定义
const (
	// LettersLowercase 小写字母
	LettersLowercase = "abcdefghijklmnopqrstuvwxyz"
	// LettersUppercase 大写字母
	LettersUppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Letters 所有字母
	Letters = LettersLowercase + LettersUppercase
	// Digits 数字
	Digits = "0123456789"
	// Alphanumeric 字母数字
	Alphanumeric = Letters + Digits
	// SpecialChars 特殊字符
	SpecialChars = "!@#$%^&*()-_=+[]{}|;:,.<>?"
	// HexChars 十六进制字符
	HexChars = "0123456789abcdef"
)

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int, charset string) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}
	if charset == "" {
		charset = Alphanumeric
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range result {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		result[i] = charset[randomIndex.Int64()]
	}

	return string(result), nil
}

// GenerateAlphanumeric 生成字母数字随机字符串
func GenerateAlphanumeric(length int) (string, error) {
	return GenerateRandomString(length, Alphanumeric)
}

// GenerateNumeric 生成数字随机字符串
func GenerateNumeric(length int) (string, error) {
	return GenerateRandomString(length, Digits)
}

// GenerateRandomCode 生成随机验证码（纯数字）
func GenerateRandomCode(length int) string {
	code, err := GenerateNumeric(length)
	if err != nil {
		// 如果生成失败，使用备用方案
		return "123456"[:length]
	}
	return code
}

// GenerateHex 生成十六进制随机字符串
func GenerateHex(length int) (string, error) {
	return GenerateRandomString(length, HexChars)
}

// GenerateSecureToken 生成安全令牌
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateUUID 生成类似UUID的字符串（不是标准UUID）
func GenerateUUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}

	// 设置版本和变体位
	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant is 10

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16]), nil
}

// IsEmpty 检查字符串是否为空（包括只有空白字符）
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty 检查字符串是否非空
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// TrimAndLower 修剪空白并转小写
func TrimAndLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// TrimAndUpper 修剪空白并转大写
func TrimAndUpper(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

// Truncate 截断字符串到指定长度
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}

// TruncateWithEllipsis 截断字符串并添加省略号
func TruncateWithEllipsis(s string, maxLen int) string {
	if maxLen <= 3 {
		return Truncate(s, maxLen)
	}
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	return Truncate(s, maxLen-3) + "..."
}

// PadLeft 左填充字符串
func PadLeft(s string, length int, padChar rune) string {
	if utf8.RuneCountInString(s) >= length {
		return s
	}
	padding := strings.Repeat(string(padChar), length-utf8.RuneCountInString(s))
	return padding + s
}

// PadRight 右填充字符串
func PadRight(s string, length int, padChar rune) string {
	if utf8.RuneCountInString(s) >= length {
		return s
	}
	padding := strings.Repeat(string(padChar), length-utf8.RuneCountInString(s))
	return s + padding
}

// Reverse 反转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ToSnakeCase 转换为蛇形命名
func ToSnakeCase(s string) string {
	// 在大写字母前插入下划线
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	s = re.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(s)
}

// ToCamelCase 转换为驼峰命名
func ToCamelCase(s string) string {
	words := strings.FieldsFunc(s, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})

	if len(words) == 0 {
		return ""
	}

	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			result += strings.ToUpper(string(words[i][0])) + strings.ToLower(words[i][1:])
		}
	}
	return result
}

// ToPascalCase 转换为帕斯卡命名
func ToPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})

	var result strings.Builder
	for _, word := range words {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(string(word[0])))
			result.WriteString(strings.ToLower(word[1:]))
		}
	}
	return result.String()
}

// ContainsIgnoreCase 忽略大小写检查包含
func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// StartsWithIgnoreCase 忽略大小写检查前缀
func StartsWithIgnoreCase(s, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
}

// EndsWithIgnoreCase 忽略大小写检查后缀
func EndsWithIgnoreCase(s, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(s), strings.ToLower(suffix))
}

// RemoveNonAlphanumeric 移除非字母数字字符
func RemoveNonAlphanumeric(s string) string {
	return regexp.MustCompile("[^a-zA-Z0-9]+").ReplaceAllString(s, "")
}

// RemoveNonPrintable 移除不可打印字符
func RemoveNonPrintable(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, s)
}

// SanitizeFilename 清理文件名，移除不安全字符
func SanitizeFilename(filename string) string {
	// 移除路径分隔符和其他危险字符
	unsafe := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	filename = unsafe.ReplaceAllString(filename, "_")

	// 移除首尾的点和空格
	filename = strings.Trim(filename, ". ")

	// 如果文件名为空或只包含点，使用默认名称
	if filename == "" || strings.Trim(filename, ".") == "" {
		filename = "unnamed"
	}

	return filename
}

// IsValidEmail 简单的邮箱格式验证
func IsValidEmail(email string) bool {
	// 简单的邮箱正则表达式
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// IsValidUsername 验证用户名格式
func IsValidUsername(username string) bool {
	// 用户名只能包含字母、数字、下划线和连字符，长度3-32
	if len(username) < 3 || len(username) > 32 {
		return false
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return re.MatchString(username)
}

// MaskEmail 脱敏邮箱地址
func MaskEmail(email string) string {
	if !IsValidEmail(email) {
		return email
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		return strings.Repeat("*", len(local)) + "@" + domain
	}

	return string(local[0]) + strings.Repeat("*", len(local)-2) + string(local[len(local)-1]) + "@" + domain
}

// MaskPhone 脱敏电话号码
func MaskPhone(phone string) string {
	// 移除非数字字符
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	if len(digits) < 7 {
		return phone
	}

	if len(digits) == 11 { // 中国手机号
		return digits[:3] + "****" + digits[7:]
	}

	// 其他格式
	mid := len(digits) / 2
	start := max(1, mid-2)
	end := min(len(digits)-1, mid+2)

	result := digits[:start] + strings.Repeat("*", end-start) + digits[end:]
	return result
}

// StringToInt 字符串转整数，带默认值
func StringToInt(s string, defaultValue int) int {
	if val, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return val
	}
	return defaultValue
}

// StringToInt64 字符串转64位整数，带默认值
func StringToInt64(s string, defaultValue int64) int64 {
	if val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
		return val
	}
	return defaultValue
}

// StringToFloat64 字符串转浮点数，带默认值
func StringToFloat64(s string, defaultValue float64) float64 {
	if val, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
		return val
	}
	return defaultValue
}

// StringToBool 字符串转布尔值，带默认值
func StringToBool(s string, defaultValue bool) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "true", "1", "yes", "on", "enabled":
		return true
	case "false", "0", "no", "off", "disabled":
		return false
	default:
		return defaultValue
	}
}

// JoinNonEmpty 连接非空字符串
func JoinNonEmpty(sep string, strs ...string) string {
	var nonEmpty []string
	for _, s := range strs {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}
	return strings.Join(nonEmpty, sep)
}

// SplitAndTrim 分割字符串并去除空白
func SplitAndTrim(s, sep string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, sep)
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// EscapeHTML 转义HTML特殊字符
func EscapeHTML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(s)
}

// UnescapeHTML 反转义HTML特殊字符
func UnescapeHTML(s string) string {
	replacer := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&#39;", "'",
	)
	return replacer.Replace(s)
}

// ToHex 字符串转十六进制
func ToHex(s string) string {
	return hex.EncodeToString([]byte(s))
}

// FromHex 十六进制转字符串
func FromHex(hexStr string) (string, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", fmt.Errorf("invalid hex string: %w", err)
	}
	return string(bytes), nil
}

// ToBase64 字符串转Base64
func ToBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// FromBase64 Base64转字符串
func FromBase64(b64Str string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		return "", fmt.Errorf("invalid base64 string: %w", err)
	}
	return string(bytes), nil
}

// 辅助函数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
