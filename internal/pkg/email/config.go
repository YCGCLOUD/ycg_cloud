package email

import (
	"fmt"
	"time"
)

// SMTPConfig SMTP服务器配置
type SMTPConfig struct {
	Host     string `mapstructure:"host" json:"host"`         // SMTP服务器地址
	Port     int    `mapstructure:"port" json:"port"`         // SMTP端口
	Username string `mapstructure:"username" json:"username"` // 用户名
	Password string `mapstructure:"password" json:"password"` // 密码
	UseSSL   bool   `mapstructure:"use_ssl" json:"use_ssl"`   // 是否使用SSL
	UseTLS   bool   `mapstructure:"use_tls" json:"use_tls"`   // 是否使用TLS
}

// EmailConfig 邮件服务配置
type EmailConfig struct {
	SMTP                SMTPConfig `mapstructure:"smtp" json:"smtp"`
	From                string     `mapstructure:"from" json:"from"`                                   // 发件人邮箱
	FromName            string     `mapstructure:"from_name" json:"from_name"`                         // 发件人名称
	ReplyTo             string     `mapstructure:"reply_to" json:"reply_to"`                           // 回复邮箱
	MaxRetries          int        `mapstructure:"max_retries" json:"max_retries"`                     // 最大重试次数
	RetryInterval       string     `mapstructure:"retry_interval" json:"retry_interval"`               // 重试间隔
	Timeout             string     `mapstructure:"timeout" json:"timeout"`                             // 超时时间
	KeepAlive           bool       `mapstructure:"keep_alive" json:"keep_alive"`                       // 保持连接
	PoolSize            int        `mapstructure:"pool_size" json:"pool_size"`                         // 连接池大小
	VerificationCodeTTL string     `mapstructure:"verification_code_ttl" json:"verification_code_ttl"` // 验证码有效期
	ResetTokenTTL       string     `mapstructure:"reset_token_ttl" json:"reset_token_ttl"`             // 重置令牌有效期
	TemplateDir         string     `mapstructure:"template_dir" json:"template_dir"`                   // 模板目录
	DefaultLanguage     string     `mapstructure:"default_language" json:"default_language"`           // 默认语言
}

// GetRetryInterval 获取重试间隔时间
func (c *EmailConfig) GetRetryInterval() time.Duration {
	if c.RetryInterval == "" {
		return 30 * time.Second
	}
	duration, err := time.ParseDuration(c.RetryInterval)
	if err != nil {
		return 30 * time.Second
	}
	return duration
}

// GetTimeout 获取超时时间
func (c *EmailConfig) GetTimeout() time.Duration {
	if c.Timeout == "" {
		return 30 * time.Second
	}
	duration, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return 30 * time.Second
	}
	return duration
}

// GetVerificationCodeTTL 获取验证码有效期
func (c *EmailConfig) GetVerificationCodeTTL() time.Duration {
	if c.VerificationCodeTTL == "" {
		return 10 * time.Minute
	}
	duration, err := time.ParseDuration(c.VerificationCodeTTL)
	if err != nil {
		return 10 * time.Minute
	}
	return duration
}

// GetResetTokenTTL 获取重置令牌有效期
func (c *EmailConfig) GetResetTokenTTL() time.Duration {
	if c.ResetTokenTTL == "" {
		return 1 * time.Hour
	}
	duration, err := time.ParseDuration(c.ResetTokenTTL)
	if err != nil {
		return 1 * time.Hour
	}
	return duration
}

// Validate 验证配置
func (c *EmailConfig) Validate() error {
	if c.SMTP.Host == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if c.SMTP.Port <= 0 || c.SMTP.Port > 65535 {
		return fmt.Errorf("invalid SMTP port: %d", c.SMTP.Port)
	}
	if c.SMTP.Username == "" {
		return fmt.Errorf("SMTP username is required")
	}
	if c.SMTP.Password == "" {
		return fmt.Errorf("SMTP password is required")
	}
	if c.From == "" {
		return fmt.Errorf("from email is required")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}
	if c.PoolSize <= 0 {
		c.PoolSize = 10 // 默认连接池大小
	}
	return nil
}

// GetSMTPAddress 获取SMTP服务器地址
func (c *EmailConfig) GetSMTPAddress() string {
	return fmt.Sprintf("%s:%d", c.SMTP.Host, c.SMTP.Port)
}

// IsSSLEnabled 检查是否启用SSL
func (c *EmailConfig) IsSSLEnabled() bool {
	return c.SMTP.UseSSL
}

// IsTLSEnabled 检查是否启用TLS
func (c *EmailConfig) IsTLSEnabled() bool {
	return c.SMTP.UseTLS
}

// GetFromAddress 获取完整的发件人地址
func (c *EmailConfig) GetFromAddress() string {
	if c.FromName != "" {
		return fmt.Sprintf("%s <%s>", c.FromName, c.From)
	}
	return c.From
}

// DefaultEmailConfig 默认邮件配置
func DefaultEmailConfig() *EmailConfig {
	return &EmailConfig{
		SMTP: SMTPConfig{
			Host:   "smtp.gmail.com",
			Port:   587,
			UseSSL: false,
			UseTLS: true,
		},
		FromName:            "HXLOS Cloud",
		MaxRetries:          3,
		RetryInterval:       "30s",
		Timeout:             "30s",
		KeepAlive:           true,
		PoolSize:            10,
		VerificationCodeTTL: "10m",
		ResetTokenTTL:       "1h",
		TemplateDir:         "templates/email",
		DefaultLanguage:     "zh-CN",
	}
}

// EmailTemplate 邮件模板配置
type EmailTemplate struct {
	Name        string            `json:"name"`        // 模板名称
	Subject     string            `json:"subject"`     // 邮件主题
	HTMLBody    string            `json:"html_body"`   // HTML内容
	TextBody    string            `json:"text_body"`   // 纯文本内容
	Variables   map[string]string `json:"variables"`   // 模板变量
	Language    string            `json:"language"`    // 语言
	IsActive    bool              `json:"is_active"`   // 是否激活
	Description string            `json:"description"` // 模板描述
}

// TemplateType 邮件模板类型常量
const (
	TemplateVerificationCode = "verification_code" // 验证码模板
	TemplatePasswordReset    = "password_reset"    // 密码重置模板
	TemplateWelcome          = "welcome"           // 欢迎邮件模板
	TemplateAccountLocked    = "account_locked"    // 账户锁定模板
	TemplateSecurityAlert    = "security_alert"    // 安全警告模板
	TemplateTeamInvitation   = "team_invitation"   // 团队邀请模板
	TemplateFileShared       = "file_shared"       // 文件分享模板
)

// EmailQueue 邮件队列项
type EmailQueue struct {
	ID          string                 `json:"id"`
	To          []string               `json:"to"`
	CC          []string               `json:"cc"`
	BCC         []string               `json:"bcc"`
	Subject     string                 `json:"subject"`
	HTMLBody    string                 `json:"html_body"`
	TextBody    string                 `json:"text_body"`
	Template    string                 `json:"template"`
	Variables   map[string]interface{} `json:"variables"`
	Priority    int                    `json:"priority"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	ScheduledAt time.Time              `json:"scheduled_at"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Status      string                 `json:"status"`
	ErrorMsg    string                 `json:"error_msg"`
}

// 邮件队列状态常量
const (
	EmailStatusPending   = "pending"   // 待发送
	EmailStatusSending   = "sending"   // 发送中
	EmailStatusSent      = "sent"      // 已发送
	EmailStatusFailed    = "failed"    // 发送失败
	EmailStatusRetrying  = "retrying"  // 重试中
	EmailStatusCancelled = "cancelled" // 已取消
)

// 邮件优先级常量
const (
	PriorityLow    = 1  // 低优先级
	PriorityNormal = 5  // 普通优先级
	PriorityHigh   = 8  // 高优先级
	PriorityUrgent = 10 // 紧急优先级
)
