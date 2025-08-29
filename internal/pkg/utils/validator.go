package utils

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 验证器接口
type Validator interface {
	ValidateEmail(email string) error
	ValidateUsername(username string) error
	ValidateDisplayName(name string) error
	ValidateRequired(value, fieldName string) error
	ValidateLength(value string, min, max int, fieldName string) error
	ValidatePattern(value, pattern, fieldName string) error
}

// defaultValidator 默认验证器实现
type defaultValidator struct{}

// NewValidator 创建新的验证器
func NewValidator() Validator {
	return &defaultValidator{}
}

// validateEmailBasicFormat 验证邮箱基本格式
func validateEmailBasicFormat(email string) error {
	// 使用Go标准库验证邮箱格式
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("邮箱格式不正确")
	}

	// 额外的邮箱格式检查
	if len(email) > 254 {
		return fmt.Errorf("邮箱长度不能超过254个字符")
	}
	return nil
}

// validateEmailParts 验证邮箱各部分格式
func validateEmailParts(email string) error {
	// 检查本地部分和域名部分
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return fmt.Errorf("邮箱格式不正确")
	}

	localPart := parts[0]
	domainPart := parts[1]

	// 验证本地部分
	if len(localPart) == 0 || len(localPart) > 64 {
		return fmt.Errorf("邮箱用户名部分长度必须在1-64个字符之间")
	}

	// 验证域名部分
	if len(domainPart) == 0 || len(domainPart) > 253 {
		return fmt.Errorf("邮箱域名部分长度必须在1-253个字符之间")
	}
	return nil
}

// validateEmailSpecialChars 验证邮箱特殊字符
func validateEmailSpecialChars(email string) error {
	// 检查是否包含连续的点
	if strings.Contains(email, "..") {
		return fmt.Errorf("邮箱不能包含连续的点")
	}

	// 检查是否以点开头或结尾
	localPart := strings.Split(email, "@")[0]
	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return fmt.Errorf("邮箱用户名不能以点开头或结尾")
	}
	return nil
}

// validateEmailDomain 验证邮箱域名格式
func validateEmailDomain(email string) error {
	domainPart := strings.Split(email, "@")[1]
	// 检查域名格式
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !domainRegex.MatchString(domainPart) {
		return fmt.Errorf("邮箱域名格式不正确")
	}
	return nil
}

// ValidateEmail 验证邮箱格式
func (v *defaultValidator) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("邮箱不能为空")
	}

	email = strings.TrimSpace(email)

	// 验证基本格式
	if err := validateEmailBasicFormat(email); err != nil {
		return err
	}

	// 验证各部分
	if err := validateEmailParts(email); err != nil {
		return err
	}

	// 验证特殊字符
	if err := validateEmailSpecialChars(email); err != nil {
		return err
	}

	// 验证域名
	return validateEmailDomain(email)
}

// getReservedUsernames 获取保留用户名列表
func getReservedUsernames() []string {
	return []string{
		"admin", "root", "user", "test", "api", "www", "ftp", "mail",
		"support", "help", "info", "contact", "about", "login", "register",
		"password", "settings", "profile", "dashboard", "system", "service",
		"guest", "public", "private", "config", "null", "undefined",
		"administrator", "moderator", "operator", "bot", "robot",
	}
}

// validateUsernameFormat 验证用户名格式
func validateUsernameFormat(username string) error {
	// 用户名只能包含字母、数字、下划线、连字符
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validUsername.MatchString(username) {
		return fmt.Errorf("用户名只能包含字母、数字、下划线和连字符")
	}
	return nil
}

// validateUsernameStartEnd 验证用户名开头和结尾
func validateUsernameStartEnd(username string) error {
	// 不能以数字开头
	if regexp.MustCompile(`^[0-9]`).MatchString(username) {
		return fmt.Errorf("用户名不能以数字开头")
	}

	// 不能以连字符或下划线开头或结尾
	if strings.HasPrefix(username, "-") || strings.HasPrefix(username, "_") ||
		strings.HasSuffix(username, "-") || strings.HasSuffix(username, "_") {
		return fmt.Errorf("用户名不能以连字符或下划线开头或结尾")
	}
	return nil
}

// validateUsernameConsecutiveChars 验证用户名连续字符
func validateUsernameConsecutiveChars(username string) error {
	// 不能包含连续的连字符或下划线
	if strings.Contains(username, "--") || strings.Contains(username, "__") ||
		strings.Contains(username, "_-") || strings.Contains(username, "-_") {
		return fmt.Errorf("用户名不能包含连续的特殊字符")
	}
	return nil
}

// validateUsernameReserved 验证用户名是否为保留名称
func validateUsernameReserved(username string) error {
	reservedNames := getReservedUsernames()
	for _, reserved := range reservedNames {
		if strings.EqualFold(username, reserved) {
			return fmt.Errorf("该用户名为系统保留，不可使用")
		}
	}
	return nil
}

// ValidateUsername 验证用户名格式
func (v *defaultValidator) ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("用户名不能为空")
	}

	username = strings.TrimSpace(username)

	// 检查长度
	if err := v.ValidateLength(username, 3, 50, "用户名"); err != nil {
		return err
	}

	// 验证格式
	if err := validateUsernameFormat(username); err != nil {
		return err
	}

	// 验证开头和结尾
	if err := validateUsernameStartEnd(username); err != nil {
		return err
	}

	// 验证连续字符
	if err := validateUsernameConsecutiveChars(username); err != nil {
		return err
	}

	// 验证保留名称
	return validateUsernameReserved(username)
}

// ValidateDisplayName 验证显示名称
func (v *defaultValidator) ValidateDisplayName(name string) error {
	if name == "" {
		return nil // 显示名称可以为空
	}

	name = strings.TrimSpace(name)

	// 检查长度
	if err := v.ValidateLength(name, 1, 100, "显示名称"); err != nil {
		return err
	}

	// 检查是否包含不允许的字符
	for _, r := range name {
		if r < 32 || r == 127 { // 控制字符
			return fmt.Errorf("显示名称不能包含控制字符")
		}
	}

	// 不能全是空白字符
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("显示名称不能全是空白字符")
	}

	return nil
}

// ValidateRequired 验证必填字段
func (v *defaultValidator) ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s不能为空", fieldName)
	}
	return nil
}

// ValidateLength 验证字符串长度
func (v *defaultValidator) ValidateLength(value string, min, max int, fieldName string) error {
	length := utf8.RuneCountInString(value)

	if min > 0 && length < min {
		return fmt.Errorf("%s长度不能少于%d个字符", fieldName, min)
	}

	if max > 0 && length > max {
		return fmt.Errorf("%s长度不能超过%d个字符", fieldName, max)
	}

	return nil
}

// ValidatePattern 验证正则表达式模式
func (v *defaultValidator) ValidatePattern(value, pattern, fieldName string) error {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return fmt.Errorf("验证%s时发生错误: %w", fieldName, err)
	}

	if !matched {
		return fmt.Errorf("%s格式不正确", fieldName)
	}

	return nil
}

// 全局便利函数

var defaultValidatorInstance = NewValidator()

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) error {
	return defaultValidatorInstance.ValidateEmail(email)
}

// ValidateUsername 验证用户名格式
func ValidateUsername(username string) error {
	return defaultValidatorInstance.ValidateUsername(username)
}

// ValidateDisplayName 验证显示名称
func ValidateDisplayName(name string) error {
	return defaultValidatorInstance.ValidateDisplayName(name)
}

// ValidateRequired 验证必填字段
func ValidateRequired(value, fieldName string) error {
	return defaultValidatorInstance.ValidateRequired(value, fieldName)
}

// ValidateLength 验证字符串长度
func ValidateLength(value string, min, max int, fieldName string) error {
	return defaultValidatorInstance.ValidateLength(value, min, max, fieldName)
}

// ValidatePattern 验证正则表达式模式
func ValidatePattern(value, pattern, fieldName string) error {
	return defaultValidatorInstance.ValidatePattern(value, pattern, fieldName)
}

// 特殊验证函数

// ValidatePhoneNumber 验证手机号码格式
func ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return nil // 手机号可以为空
	}

	phone = strings.TrimSpace(phone)

	// 移除常见的分隔符
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	// 检查是否只包含数字和可选的+号
	if !regexp.MustCompile(`^\+?[0-9]+$`).MatchString(phone) {
		return fmt.Errorf("手机号码格式不正确")
	}

	// 检查长度（不包括国际区号前缀）
	digits := strings.TrimPrefix(phone, "+")
	if len(digits) < 7 || len(digits) > 15 {
		return fmt.Errorf("手机号码长度必须在7-15位之间")
	}

	return nil
}

// ValidateURL 验证URL格式
func ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL不能为空")
	}

	// 简单的URL格式验证
	urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(:[0-9]+)?(/.*)?$`)
	if !urlRegex.MatchString(url) {
		return fmt.Errorf("URL格式不正确")
	}

	return nil
}

// ValidateAge 验证年龄
func ValidateAge(age int) error {
	if age < 0 {
		return fmt.Errorf("年龄不能为负数")
	}

	if age > 150 {
		return fmt.Errorf("年龄不能超过150岁")
	}

	return nil
}

// ValidateConfirmPassword 验证确认密码
func ValidateConfirmPassword(password, confirmPassword string) error {
	if password != confirmPassword {
		return fmt.Errorf("密码和确认密码不一致")
	}
	return nil
}

// ValidateAcceptTerms 验证用户是否接受条款
func ValidateAcceptTerms(accept bool) error {
	if !accept {
		return fmt.Errorf("必须接受服务条款")
	}
	return nil
}

// 验证码相关验证

// ValidateVerificationCode 验证验证码格式
func ValidateVerificationCode(code string) error {
	if code == "" {
		return fmt.Errorf("验证码不能为空")
	}

	code = strings.TrimSpace(code)

	// 验证码应该是6位数字
	if !regexp.MustCompile(`^[0-9]{6}$`).MatchString(code) {
		return fmt.Errorf("验证码必须是6位数字")
	}

	return nil
}

// ValidateCodeType 验证验证码类型
func ValidateCodeType(codeType string) error {
	validTypes := []string{"register", "password_reset", "login", "change_email"}

	for _, validType := range validTypes {
		if codeType == validType {
			return nil
		}
	}

	return fmt.Errorf("验证码类型不正确")
}

// 批量验证函数

// ValidateUserRegistration 验证用户注册数据
func ValidateUserRegistration(email, username, password, confirmPassword, displayName string, acceptTerms bool) error {
	// 验证邮箱
	if err := ValidateEmail(email); err != nil {
		return fmt.Errorf("邮箱验证失败: %w", err)
	}

	// 验证用户名
	if err := ValidateUsername(username); err != nil {
		return fmt.Errorf("用户名验证失败: %w", err)
	}

	// 验证密码强度
	if _, err := ValidatePasswordStrength(password); err != nil {
		return fmt.Errorf("密码验证失败: %w", err)
	}

	// 验证确认密码
	if err := ValidateConfirmPassword(password, confirmPassword); err != nil {
		return fmt.Errorf("确认密码验证失败: %w", err)
	}

	// 验证显示名称
	if err := ValidateDisplayName(displayName); err != nil {
		return fmt.Errorf("显示名称验证失败: %w", err)
	}

	// 验证接受条款
	if err := ValidateAcceptTerms(acceptTerms); err != nil {
		return fmt.Errorf("服务条款验证失败: %w", err)
	}

	return nil
}

// 辅助函数

// IsAlpha 检查字符串是否只包含字母
func IsAlpha(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return len(s) > 0
}

// IsAlphanumeric 检查字符串是否只包含字母和数字
func IsAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

// IsNumeric 检查字符串是否只包含数字
func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

// ContainsWhitespace 检查字符串是否包含空白字符
func ContainsWhitespace(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) {
			return true
		}
	}
	return false
}
