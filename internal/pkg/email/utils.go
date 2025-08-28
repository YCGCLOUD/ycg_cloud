package email

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ValidateEmailAddress 验证邮箱地址格式
func ValidateEmailAddress(email string) error {
	if email == "" {
		return fmt.Errorf("email address is required")
	}

	// 正则表达式验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email address format")
	}

	return nil
}

// ValidateEmailList 批量验证邮箱地址
func ValidateEmailList(emails []string) error {
	if len(emails) == 0 {
		return fmt.Errorf("email list is empty")
	}

	for i, email := range emails {
		if err := ValidateEmailAddress(email); err != nil {
			return fmt.Errorf("invalid email at index %d: %w", i, err)
		}
	}

	return nil
}

// NormalizeEmailAddress 规范化邮箱地址
func NormalizeEmailAddress(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizeEmailList 批量规范化邮箱地址
func NormalizeEmailList(emails []string) []string {
	normalized := make([]string, len(emails))
	for i, email := range emails {
		normalized[i] = NormalizeEmailAddress(email)
	}
	return normalized
}

// IsTemporaryEmailProvider 检查是否为临时邮箱提供商
func IsTemporaryEmailProvider(email string) bool {
	// 常见临时邮箱域名列表
	tempDomains := []string{
		"10minutemail.com",
		"guerrillamail.com",
		"mailinator.com",
		"tempmail.org",
		"trash-mail.com",
		"yopmail.com",
		"getnada.com",
		"maildrop.cc",
		"emailondeck.com",
		"sharklasers.com",
	}

	emailLower := strings.ToLower(email)
	for _, domain := range tempDomains {
		if strings.HasSuffix(emailLower, "@"+domain) {
			return true
		}
	}

	return false
}

// GetEmailDomain 获取邮箱域名
func GetEmailDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return strings.ToLower(parts[1])
}

// IsBusinessEmail 检查是否为企业邮箱
func IsBusinessEmail(email string) bool {
	domain := GetEmailDomain(email)
	if domain == "" {
		return false
	}

	// 常见个人邮箱域名
	personalDomains := []string{
		"gmail.com",
		"yahoo.com",
		"hotmail.com",
		"outlook.com",
		"qq.com",
		"163.com",
		"126.com",
		"sina.com",
		"sohu.com",
		"139.com",
	}

	for _, personalDomain := range personalDomains {
		if domain == personalDomain {
			return false
		}
	}

	return true
}

// SanitizeEmailContent 清理邮件内容，防止注入攻击
func SanitizeEmailContent(content string) string {
	// 移除潜在的恶意脚本标签
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	content = scriptRegex.ReplaceAllString(content, "")

	// 移除潜在的恶意样式标签
	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	content = styleRegex.ReplaceAllString(content, "")

	// 移除javascript:协议链接
	jsRegex := regexp.MustCompile(`(?i)javascript:`)
	content = jsRegex.ReplaceAllString(content, "")

	return content
}

// GenerateUnsubscribeURL 生成退订链接
func GenerateUnsubscribeURL(baseURL, userID, token string) string {
	if baseURL == "" {
		baseURL = "https://example.com"
	}
	return fmt.Sprintf("%s/unsubscribe?user=%s&token=%s", baseURL, userID, token)
}

// FormatEmailAddress 格式化邮箱地址显示
func FormatEmailAddress(email, name string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}

// ParseEmailAddress 解析邮箱地址
func ParseEmailAddress(address string) (email, name string) {
	// 匹配 "Name <email@domain.com>" 格式
	nameEmailRegex := regexp.MustCompile(`^(.+?)\s*<(.+)>$`)
	matches := nameEmailRegex.FindStringSubmatch(strings.TrimSpace(address))

	if len(matches) == 3 {
		name = strings.Trim(strings.TrimSpace(matches[1]), `"'`)
		email = strings.TrimSpace(matches[2])
	} else {
		email = strings.TrimSpace(address)
	}

	return email, name
}

// GetEmailProvider 获取邮箱服务提供商
func GetEmailProvider(email string) string {
	domain := GetEmailDomain(email)
	if domain == "" {
		return "unknown"
	}

	providers := map[string]string{
		"gmail.com":   "Google",
		"yahoo.com":   "Yahoo",
		"hotmail.com": "Microsoft",
		"outlook.com": "Microsoft",
		"live.com":    "Microsoft",
		"qq.com":      "Tencent",
		"163.com":     "NetEase",
		"126.com":     "NetEase",
		"sina.com":    "Sina",
		"sohu.com":    "Sohu",
		"139.com":     "China Mobile",
	}

	if provider, exists := providers[domain]; exists {
		return provider
	}

	return "Other"
}

// EstimateDeliveryTime 估算邮件投递时间
func EstimateDeliveryTime(emailCount int, provider string) time.Duration {
	baseTime := 5 * time.Second

	// 根据邮件数量调整
	if emailCount > 100 {
		baseTime += time.Duration(emailCount/100) * time.Second
	}

	// 根据服务提供商调整
	switch provider {
	case "Google", "Microsoft":
		// 大厂商通常更快
		baseTime = baseTime * 80 / 100
	case "Other":
		// 其他提供商可能较慢
		baseTime = baseTime * 120 / 100
	}

	return baseTime
}

// GetOptimalSendTime 获取最佳发送时间
func GetOptimalSendTime(timezone string) time.Time {
	// 通常工作日上午9-11点是最佳邮件发送时间
	now := time.Now()

	// 如果是周末，推迟到下周一
	if now.Weekday() == time.Saturday {
		now = now.AddDate(0, 0, 2)
	} else if now.Weekday() == time.Sunday {
		now = now.AddDate(0, 0, 1)
	}

	// 设置为上午10点
	optimal := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())

	// 如果已经过了今天的最佳时间，则推迟到明天
	if optimal.Before(now) {
		optimal = optimal.AddDate(0, 0, 1)
		// 再次检查是否是周末
		if optimal.Weekday() == time.Saturday {
			optimal = optimal.AddDate(0, 0, 2)
		} else if optimal.Weekday() == time.Sunday {
			optimal = optimal.AddDate(0, 0, 1)
		}
	}

	return optimal
}

// CalculateEmailPriority 计算邮件优先级
func CalculateEmailPriority(templateType string, urgent bool) int {
	basePriority := PriorityNormal

	switch templateType {
	case TemplateVerificationCode:
		basePriority = PriorityHigh
	case TemplatePasswordReset:
		basePriority = PriorityHigh
	case TemplateSecurityAlert:
		basePriority = PriorityUrgent
	case TemplateWelcome:
		basePriority = PriorityNormal
	case TemplateAccountLocked:
		basePriority = PriorityHigh
	case TemplateTeamInvitation:
		basePriority = PriorityNormal
	case TemplateFileShared:
		basePriority = PriorityLow
	}

	if urgent {
		basePriority = PriorityUrgent
	}

	return basePriority
}

// CreateEmailQueue 创建邮件队列项
func CreateEmailQueue(templateName string, to []string, variables map[string]interface{}, priority int) *EmailQueue {
	return &EmailQueue{
		ID:          fmt.Sprintf("email_%d_%d", time.Now().Unix(), time.Now().Nanosecond()),
		To:          to,
		Template:    templateName,
		Variables:   variables,
		Priority:    priority,
		Status:      EmailStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		MaxAttempts: 3,
	}
}

// CreateDirectEmailQueue 创建直接邮件队列项
func CreateDirectEmailQueue(to []string, subject, htmlBody, textBody string, priority int) *EmailQueue {
	return &EmailQueue{
		ID:          fmt.Sprintf("email_%d_%d", time.Now().Unix(), time.Now().Nanosecond()),
		To:          to,
		Subject:     subject,
		HTMLBody:    htmlBody,
		TextBody:    textBody,
		Priority:    priority,
		Status:      EmailStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		MaxAttempts: 3,
	}
}

// GetEmailStatistics 获取邮件统计信息
type EmailStatistics struct {
	TotalSent    int            `json:"total_sent"`
	TotalFailed  int            `json:"total_failed"`
	TotalPending int            `json:"total_pending"`
	SuccessRate  float64        `json:"success_rate"`
	ByTemplate   map[string]int `json:"by_template"`
	ByProvider   map[string]int `json:"by_provider"`
	ByHour       map[int]int    `json:"by_hour"`
	LastUpdated  time.Time      `json:"last_updated"`
}
