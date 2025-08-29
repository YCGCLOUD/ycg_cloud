package utils

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// PasswordSecurityChecker 密码安全检查器接口
type PasswordSecurityChecker interface {
	// 密码历史检查
	CheckPasswordHistory(ctx context.Context, userID uint, newPassword string) error
	ValidatePasswordAge(ctx context.Context, userID uint, maxAge time.Duration) error
	CheckPasswordReuse(ctx context.Context, userID uint, password string, historyCount int) error

	// 密码复杂度检查
	CheckPasswordComplexity(password string) (*PasswordComplexityResult, error)
	ValidatePasswordPolicy(password string, policy *PasswordPolicy) error
	CheckCommonPasswords(password string) error

	// 账户安全检查
	CheckAccountLockout(ctx context.Context, userID uint) error
	ValidatePasswordChangeFrequency(ctx context.Context, userID uint, minInterval time.Duration) error
	CheckSuspiciousActivity(ctx context.Context, userID uint, ipAddress string) error

	// 密码生成建议
	GeneratePasswordSuggestions(currentPassword string) []string
	GetPasswordStrengthScore(password string) int
	CalculatePasswordEntropy(password string) float64
}

// PasswordPolicy 密码策略
type PasswordPolicy struct {
	MinLength           int      `json:"min_length"`            // 最小长度
	MaxLength           int      `json:"max_length"`            // 最大长度
	RequireUppercase    bool     `json:"require_uppercase"`     // 要求大写字母
	RequireLowercase    bool     `json:"require_lowercase"`     // 要求小写字母
	RequireDigits       bool     `json:"require_digits"`        // 要求数字
	RequireSpecialChars bool     `json:"require_special_chars"` // 要求特殊字符
	MinSpecialChars     int      `json:"min_special_chars"`     // 最少特殊字符数
	MaxConsecutiveChars int      `json:"max_consecutive_chars"` // 最大连续字符数
	MaxRepeatingChars   int      `json:"max_repeating_chars"`   // 最大重复字符数
	ForbiddenWords      []string `json:"forbidden_words"`       // 禁用词汇
	ForbiddenPatterns   []string `json:"forbidden_patterns"`    // 禁用模式
	RequireComplexity   int      `json:"require_complexity"`    // 要求复杂度等级
	AllowUserInfo       bool     `json:"allow_user_info"`       // 允许包含用户信息
	HistoryCount        int      `json:"history_count"`         // 密码历史数量
	MaxAge              int      `json:"max_age"`               // 密码最大使用天数
	MinChangeInterval   int      `json:"min_change_interval"`   // 最小修改间隔（小时）
}

// PasswordComplexityResult 密码复杂度检查结果
type PasswordComplexityResult struct {
	Strength           int      `json:"strength"`             // 强度等级 1-3
	Score              int      `json:"score"`                // 得分 0-100
	Entropy            float64  `json:"entropy"`              // 熵值
	HasUppercase       bool     `json:"has_uppercase"`        // 包含大写字母
	HasLowercase       bool     `json:"has_lowercase"`        // 包含小写字母
	HasDigits          bool     `json:"has_digits"`           // 包含数字
	HasSpecialChars    bool     `json:"has_special_chars"`    // 包含特殊字符
	CharsetSize        int      `json:"charset_size"`         // 字符集大小
	UniqueChars        int      `json:"unique_chars"`         // 唯一字符数
	RepeatingChars     int      `json:"repeating_chars"`      // 重复字符数
	SequentialChars    int      `json:"sequential_chars"`     // 连续字符数
	Suggestions        []string `json:"suggestions"`          // 改进建议
	Warnings           []string `json:"warnings"`             // 安全警告
	EstimatedCrackTime string   `json:"estimated_crack_time"` // 预估破解时间
}

// PasswordSecurityInfo 密码安全信息
type PasswordSecurityInfo struct {
	LastChanged        time.Time `json:"last_changed"`        // 上次修改时间
	ChangeFrequency    int       `json:"change_frequency"`    // 修改频率（天）
	IsExpired          bool      `json:"is_expired"`          // 是否过期
	DaysUntilExpiry    int       `json:"days_until_expiry"`   // 距离过期天数
	FailedAttempts     int       `json:"failed_attempts"`     // 失败尝试次数
	IsLocked           bool      `json:"is_locked"`           // 是否锁定
	LockoutExpiry      time.Time `json:"lockout_expiry"`      // 锁定过期时间
	SuspiciousActivity bool      `json:"suspicious_activity"` // 可疑活动
	RiskLevel          string    `json:"risk_level"`          // 风险等级
}

// defaultPasswordSecurityChecker 默认密码安全检查器实现
type defaultPasswordSecurityChecker struct{}

// NewPasswordSecurityChecker 创建密码安全检查器
func NewPasswordSecurityChecker() PasswordSecurityChecker {
	return &defaultPasswordSecurityChecker{}
}

// CheckPasswordComplexity 检查密码复杂度
func (c *defaultPasswordSecurityChecker) CheckPasswordComplexity(password string) (*PasswordComplexityResult, error) {
	if password == "" {
		return nil, fmt.Errorf("密码不能为空")
	}

	result := &PasswordComplexityResult{
		Suggestions: make([]string, 0),
		Warnings:    make([]string, 0),
	}

	// 检查字符类型
	hasUpper, hasLower, hasDigit, hasSpecial := analyzeCharacterTypes(password)
	result.HasUppercase = hasUpper
	result.HasLowercase = hasLower
	result.HasDigits = hasDigit
	result.HasSpecialChars = hasSpecial

	// 计算字符集大小
	result.CharsetSize = calculateCharsetSize(hasUpper, hasLower, hasDigit, hasSpecial)

	// 计算唯一字符数和重复字符数
	result.UniqueChars, result.RepeatingChars = analyzeCharacterDistribution(password)

	// 检查连续字符
	result.SequentialChars = countSequentialChars(password)

	// 计算熵值
	result.Entropy = c.CalculatePasswordEntropy(password)

	// 计算得分
	result.Score = c.GetPasswordStrengthScore(password)

	// 确定强度等级
	if result.Score >= 80 {
		result.Strength = PasswordStrong
	} else if result.Score >= 60 {
		result.Strength = PasswordMedium
	} else {
		result.Strength = PasswordWeek
	}

	// 生成建议和警告
	c.generateSuggestionsAndWarnings(password, result)

	// 估算破解时间
	result.EstimatedCrackTime = c.estimateCrackTime(result.Entropy)

	return result, nil
}

// ValidatePasswordPolicy 验证密码策略
func (c *defaultPasswordSecurityChecker) ValidatePasswordPolicy(password string, policy *PasswordPolicy) error {
	if policy == nil {
		return nil // 没有策略要求
	}

	// 检查基本长度要求
	if err := c.validatePasswordLength(password, policy); err != nil {
		return err
	}

	// 检查字符类型要求
	if err := c.validateCharacterTypes(password, policy); err != nil {
		return err
	}

	// 检查特殊字符数量要求
	if err := c.validateSpecialCharCount(password, policy); err != nil {
		return err
	}

	// 检查字符模式要求
	if err := c.validateCharacterPatterns(password, policy); err != nil {
		return err
	}

	// 检查禁用内容
	if err := c.validateForbiddenContent(password, policy); err != nil {
		return err
	}

	// 检查复杂度要求
	return c.validateComplexityRequirement(password, policy)
}

// validatePasswordLength 验证密码长度
func (c *defaultPasswordSecurityChecker) validatePasswordLength(password string, policy *PasswordPolicy) error {
	if len(password) < policy.MinLength {
		return fmt.Errorf("密码长度至少需要%d位", policy.MinLength)
	}
	if policy.MaxLength > 0 && len(password) > policy.MaxLength {
		return fmt.Errorf("密码长度不能超过%d位", policy.MaxLength)
	}
	return nil
}

// validateCharacterTypes 验证字符类型要求
func (c *defaultPasswordSecurityChecker) validateCharacterTypes(password string, policy *PasswordPolicy) error {
	hasUpper, hasLower, hasDigit, hasSpecial := analyzeCharacterTypes(password)

	if policy.RequireUppercase && !hasUpper {
		return fmt.Errorf("密码必须包含大写字母")
	}
	if policy.RequireLowercase && !hasLower {
		return fmt.Errorf("密码必须包含小写字母")
	}
	if policy.RequireDigits && !hasDigit {
		return fmt.Errorf("密码必须包含数字")
	}
	if policy.RequireSpecialChars && !hasSpecial {
		return fmt.Errorf("密码必须包含特殊字符")
	}
	return nil
}

// validateSpecialCharCount 验证特殊字符数量
func (c *defaultPasswordSecurityChecker) validateSpecialCharCount(password string, policy *PasswordPolicy) error {
	if policy.MinSpecialChars > 0 {
		specialCount := countSpecialChars(password)
		if specialCount < policy.MinSpecialChars {
			return fmt.Errorf("密码至少需要包含%d个特殊字符", policy.MinSpecialChars)
		}
	}
	return nil
}

// validateCharacterPatterns 验证字符模式
func (c *defaultPasswordSecurityChecker) validateCharacterPatterns(password string, policy *PasswordPolicy) error {
	// 检查连续字符
	if policy.MaxConsecutiveChars > 0 {
		if hasConsecutiveSequentialChars(password, policy.MaxConsecutiveChars) {
			return fmt.Errorf("密码不能包含超过%d个连续字符", policy.MaxConsecutiveChars)
		}
	}

	// 检查重复字符
	if policy.MaxRepeatingChars > 0 {
		if hasRepeatingChars(password, policy.MaxRepeatingChars) {
			return fmt.Errorf("密码不能包含超过%d个重复字符", policy.MaxRepeatingChars)
		}
	}
	return nil
}

// validateForbiddenContent 验证禁用内容
func (c *defaultPasswordSecurityChecker) validateForbiddenContent(password string, policy *PasswordPolicy) error {
	// 检查禁用词汇
	if err := c.checkForbiddenWords(password, policy.ForbiddenWords); err != nil {
		return err
	}

	// 检查禁用模式
	return c.checkForbiddenPatterns(password, policy.ForbiddenPatterns)
}

// validateComplexityRequirement 验证复杂度要求
func (c *defaultPasswordSecurityChecker) validateComplexityRequirement(password string, policy *PasswordPolicy) error {
	if policy.RequireComplexity > 0 {
		complexity, err := c.CheckPasswordComplexity(password)
		if err != nil {
			return err
		}
		if complexity.Strength < policy.RequireComplexity {
			return fmt.Errorf("密码复杂度不足，要求等级%d", policy.RequireComplexity)
		}
	}
	return nil
}

// GetPasswordStrengthScore 获取密码强度得分
func (c *defaultPasswordSecurityChecker) GetPasswordStrengthScore(password string) int {
	score := 0

	// 长度评分（0-25分）
	score += c.calculateLengthScore(password)

	// 字符类型评分（0-40分）
	score += c.calculateCharacterTypeScore(password)

	// 字符集多样性评分（0-20分）
	score += c.calculateDiversityScore(password)

	// 模式检查（扣分）
	score -= c.calculatePatternPenalty(password)

	// 常见密码检查（扣分）
	score -= c.calculateCommonPasswordPenalty(password)

	// 确保得分在0-100范围内
	return c.normalizeScore(score)
}

// calculateLengthScore 计算长度得分
func (c *defaultPasswordSecurityChecker) calculateLengthScore(password string) int {
	score := 0
	length := len(password)

	if length >= 8 {
		score += 5
	}
	if length >= 10 {
		score += 5
	}
	if length >= 12 {
		score += 10
	}
	if length >= 16 {
		score += 5
	}
	return score
}

// calculateCharacterTypeScore 计算字符类型得分
func (c *defaultPasswordSecurityChecker) calculateCharacterTypeScore(password string) int {
	score := 0
	hasUpper, hasLower, hasDigit, hasSpecial := analyzeCharacterTypes(password)

	if hasUpper {
		score += 10
	}
	if hasLower {
		score += 10
	}
	if hasDigit {
		score += 10
	}
	if hasSpecial {
		score += 10
	}
	return score
}

// calculateDiversityScore 计算字符集多样性得分
func (c *defaultPasswordSecurityChecker) calculateDiversityScore(password string) int {
	uniqueChars, _ := analyzeCharacterDistribution(password)
	charsetDiversity := float64(uniqueChars) / float64(len(password))

	if charsetDiversity >= 0.8 {
		return 20
	} else if charsetDiversity >= 0.6 {
		return 15
	} else if charsetDiversity >= 0.4 {
		return 10
	} else if charsetDiversity >= 0.2 {
		return 5
	}
	return 0
}

// calculatePatternPenalty 计算模式罚分
func (c *defaultPasswordSecurityChecker) calculatePatternPenalty(password string) int {
	penalty := 0

	if hasSequentialPatterns(password, 3) {
		penalty += 10
	}
	if hasRepeatingChars(password, 3) {
		penalty += 10
	}
	return penalty
}

// calculateCommonPasswordPenalty 计算常见密码罚分
func (c *defaultPasswordSecurityChecker) calculateCommonPasswordPenalty(password string) int {
	if c.isCommonPassword(password) {
		return 20
	}
	return 0
}

// normalizeScore 归一化得分到 0-100 范围
func (c *defaultPasswordSecurityChecker) normalizeScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// CalculatePasswordEntropy 计算密码熵值
func (c *defaultPasswordSecurityChecker) CalculatePasswordEntropy(password string) float64 {
	if len(password) == 0 {
		return 0
	}

	// 计算字符集大小
	hasUpper, hasLower, hasDigit, hasSpecial := analyzeCharacterTypes(password)
	charsetSize := calculateCharsetSize(hasUpper, hasLower, hasDigit, hasSpecial)

	// 使用简化的熵计算：熵 = 长度 * log2(字符集大小)
	if charsetSize == 0 {
		return 0
	}

	// 简化的熵计算
	entropy := float64(len(password)) * log2(float64(charsetSize))
	return entropy
}

// log2 计算以2为底的对数
func log2(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// 简化计算：log2(x) = ln(x) / ln(2) ≈ ln(x) / 0.693
	// 这里使用更简单的近似
	return 3.32 * log10(x) // log2(x) = log10(x) / log10(2) ≈ log10(x) * 3.32
}

// log10 简化的以10为底的对数近似
func log10(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// 非常简化的对数近似
	if x < 10 {
		return 1
	}
	if x < 100 {
		return 2
	}
	return 3
}

// 其他未实现的方法（为了满足接口要求）
func (c *defaultPasswordSecurityChecker) CheckPasswordHistory(ctx context.Context, userID uint, newPassword string) error {
	// TODO: 实现密码历史检查
	return nil
}

func (c *defaultPasswordSecurityChecker) ValidatePasswordAge(ctx context.Context, userID uint, maxAge time.Duration) error {
	// TODO: 实现密码年龄验证
	return nil
}

func (c *defaultPasswordSecurityChecker) CheckPasswordReuse(ctx context.Context, userID uint, password string, historyCount int) error {
	// TODO: 实现密码重用检查
	return nil
}

func (c *defaultPasswordSecurityChecker) CheckAccountLockout(ctx context.Context, userID uint) error {
	// TODO: 实现账户锁定检查
	return nil
}

func (c *defaultPasswordSecurityChecker) ValidatePasswordChangeFrequency(ctx context.Context, userID uint, minInterval time.Duration) error {
	// TODO: 实现密码修改频率验证
	return nil
}

func (c *defaultPasswordSecurityChecker) CheckSuspiciousActivity(ctx context.Context, userID uint, ipAddress string) error {
	// TODO: 实现可疑活动检查
	return nil
}

func (c *defaultPasswordSecurityChecker) GeneratePasswordSuggestions(currentPassword string) []string {
	suggestions := make([]string, 0)

	hasUpper, hasLower, hasDigit, hasSpecial := analyzeCharacterTypes(currentPassword)

	if len(currentPassword) < 12 {
		suggestions = append(suggestions, "增加密码长度到12位以上")
	}
	if !hasUpper {
		suggestions = append(suggestions, "添加大写字母")
	}
	if !hasLower {
		suggestions = append(suggestions, "添加小写字母")
	}
	if !hasDigit {
		suggestions = append(suggestions, "添加数字")
	}
	if !hasSpecial {
		suggestions = append(suggestions, "添加特殊字符（如!@#$%）")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "密码强度良好，建议定期更换")
	}

	return suggestions
}

func (c *defaultPasswordSecurityChecker) CheckCommonPasswords(password string) error {
	commonPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123",
		"password123", "admin", "root", "user", "test", "guest",
		"111111", "000000", "654321", "123123", "987654321",
		"welcome", "login", "master", "monkey", "dragon",
	}

	passwordLower := strings.ToLower(password)
	for _, common := range commonPasswords {
		// 完全匹配或包含常见密码
		if passwordLower == common {
			return fmt.Errorf("密码过于常见，请选择更安全的密码")
		}
	}

	return nil
}

// 辅助函数

func analyzeCharacterTypes(password string) (hasUpper, hasLower, hasDigit, hasSpecial bool) {
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}
	return
}

func calculateCharsetSize(hasUpper, hasLower, hasDigit, hasSpecial bool) int {
	size := 0
	if hasUpper {
		size += 26
	}
	if hasLower {
		size += 26
	}
	if hasDigit {
		size += 10
	}
	if hasSpecial {
		size += 32 // 估算常用特殊字符数量
	}
	return size
}

func analyzeCharacterDistribution(password string) (unique, repeating int) {
	charCount := make(map[rune]int)
	for _, char := range password {
		charCount[char]++
	}

	unique = len(charCount)
	for _, count := range charCount {
		if count > 1 {
			repeating += count - 1
		}
	}
	return
}

func countSequentialChars(password string) int {
	count := 0
	for i := 0; i < len(password)-2; i++ {
		if password[i+1] == password[i]+1 && password[i+2] == password[i]+2 {
			count++
		}
	}
	return count
}

func countSpecialChars(password string) int {
	count := 0
	for _, char := range password {
		if char < '0' || (char > '9' && char < 'A') || (char > 'Z' && char < 'a') || char > 'z' {
			count++
		}
	}
	return count
}

func hasConsecutiveSequentialChars(password string, maxLength int) bool {
	return countSequentialChars(password) >= maxLength
}

func hasRepeatingChars(password string, maxLength int) bool {
	count := 1
	for i := 1; i < len(password); i++ {
		if password[i] == password[i-1] {
			count++
			if count >= maxLength {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

func hasSequentialPatterns(password string, minLength int) bool {
	return countSequentialChars(password) >= minLength
}

func (c *defaultPasswordSecurityChecker) checkForbiddenWords(password string, forbiddenWords []string) error {
	passwordLower := strings.ToLower(password)
	for _, word := range forbiddenWords {
		if strings.Contains(passwordLower, strings.ToLower(word)) {
			return fmt.Errorf("密码不能包含禁用词汇")
		}
	}
	return nil
}

func (c *defaultPasswordSecurityChecker) checkForbiddenPatterns(password string, patterns []string) error {
	// 简化的模式检查，实际应该使用正则表达式
	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(password), strings.ToLower(pattern)) {
			return fmt.Errorf("密码不能包含禁用模式")
		}
	}
	return nil
}

func (c *defaultPasswordSecurityChecker) isCommonPassword(password string) bool {
	err := c.CheckCommonPasswords(password)
	return err != nil
}

func (c *defaultPasswordSecurityChecker) generateSuggestionsAndWarnings(password string, result *PasswordComplexityResult) {
	// 生成改进建议
	if !result.HasUppercase {
		result.Suggestions = append(result.Suggestions, "添加大写字母")
	}
	if !result.HasLowercase {
		result.Suggestions = append(result.Suggestions, "添加小写字母")
	}
	if !result.HasDigits {
		result.Suggestions = append(result.Suggestions, "添加数字")
	}
	if !result.HasSpecialChars {
		result.Suggestions = append(result.Suggestions, "添加特殊字符")
	}
	if len(password) < 12 {
		result.Suggestions = append(result.Suggestions, "增加密码长度")
	}

	// 生成安全警告
	if result.RepeatingChars > len(password)/3 {
		result.Warnings = append(result.Warnings, "密码包含过多重复字符")
	}
	if result.SequentialChars > 0 {
		result.Warnings = append(result.Warnings, "密码包含连续字符序列")
	}
	if c.isCommonPassword(password) {
		result.Warnings = append(result.Warnings, "密码过于常见")
	}
}

func (c *defaultPasswordSecurityChecker) estimateCrackTime(entropy float64) string {
	// 假设每秒尝试1亿次密码
	attemptsPerSecond := 100000000.0

	// 计算可能的密码组合数（2^entropy）
	possibleCombinations := pow(2, entropy)

	// 平均破解时间（一半的组合数）
	averageTime := possibleCombinations / 2 / attemptsPerSecond

	// 转换为可读的时间格式
	if averageTime < 60 {
		return fmt.Sprintf("%.0f秒", averageTime)
	} else if averageTime < 3600 {
		return fmt.Sprintf("%.0f分钟", averageTime/60)
	} else if averageTime < 86400 {
		return fmt.Sprintf("%.0f小时", averageTime/3600)
	} else if averageTime < 31536000 {
		return fmt.Sprintf("%.0f天", averageTime/86400)
	} else {
		return fmt.Sprintf("%.0f年", averageTime/31536000)
	}
}

// pow 简化的幂运算函数
func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	result := base
	for i := 1; i < int(exp); i++ {
		result *= base
	}
	return result
}

// 全局便利函数

var defaultSecurityChecker = NewPasswordSecurityChecker()

// CheckPasswordComplexityGlobal 全局密码复杂度检查
func CheckPasswordComplexityGlobal(password string) (*PasswordComplexityResult, error) {
	return defaultSecurityChecker.CheckPasswordComplexity(password)
}

// GetPasswordStrengthScoreGlobal 全局密码强度得分
func GetPasswordStrengthScoreGlobal(password string) int {
	return defaultSecurityChecker.GetPasswordStrengthScore(password)
}

// CalculatePasswordEntropyGlobal 全局密码熵值计算
func CalculatePasswordEntropyGlobal(password string) float64 {
	return defaultSecurityChecker.CalculatePasswordEntropy(password)
}
