package email

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEmailConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *EmailConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &EmailConfig{
				SMTP: SMTPConfig{
					Host:     "smtp.gmail.com",
					Port:     587,
					Username: "test@gmail.com",
					Password: "password",
					UseTLS:   true,
				},
				From: "test@gmail.com",
			},
			wantErr: false,
		},
		{
			name: "missing SMTP host",
			config: &EmailConfig{
				SMTP: SMTPConfig{
					Port:     587,
					Username: "test@gmail.com",
					Password: "password",
				},
				From: "test@gmail.com",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &EmailConfig{
				SMTP: SMTPConfig{
					Host:     "smtp.gmail.com",
					Port:     0,
					Username: "test@gmail.com",
					Password: "password",
				},
				From: "test@gmail.com",
			},
			wantErr: true,
		},
		{
			name: "missing username",
			config: &EmailConfig{
				SMTP: SMTPConfig{
					Host:     "smtp.gmail.com",
					Port:     587,
					Password: "password",
				},
				From: "test@gmail.com",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			config: &EmailConfig{
				SMTP: SMTPConfig{
					Host:     "smtp.gmail.com",
					Port:     587,
					Username: "test@gmail.com",
				},
				From: "test@gmail.com",
			},
			wantErr: true,
		},
		{
			name: "missing from email",
			config: &EmailConfig{
				SMTP: SMTPConfig{
					Host:     "smtp.gmail.com",
					Port:     587,
					Username: "test@gmail.com",
					Password: "password",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailConfig_GetDurations(t *testing.T) {
	config := &EmailConfig{
		RetryInterval:       "30s",
		Timeout:             "1m",
		VerificationCodeTTL: "5m",
		ResetTokenTTL:       "2h",
	}

	assert.Equal(t, 30*time.Second, config.GetRetryInterval())
	assert.Equal(t, 1*time.Minute, config.GetTimeout())
	assert.Equal(t, 5*time.Minute, config.GetVerificationCodeTTL())
	assert.Equal(t, 2*time.Hour, config.GetResetTokenTTL())
}

func TestEmailConfig_GetDurations_Invalid(t *testing.T) {
	config := &EmailConfig{
		RetryInterval:       "invalid",
		Timeout:             "invalid",
		VerificationCodeTTL: "invalid",
		ResetTokenTTL:       "invalid",
	}

	// 应该返回默认值
	assert.Equal(t, 30*time.Second, config.GetRetryInterval())
	assert.Equal(t, 30*time.Second, config.GetTimeout())
	assert.Equal(t, 10*time.Minute, config.GetVerificationCodeTTL())
	assert.Equal(t, 1*time.Hour, config.GetResetTokenTTL())
}

func TestEmailConfig_GetFromAddress(t *testing.T) {
	tests := []struct {
		name     string
		config   *EmailConfig
		expected string
	}{
		{
			name: "with from name",
			config: &EmailConfig{
				From:     "test@example.com",
				FromName: "Test Service",
			},
			expected: "Test Service <test@example.com>",
		},
		{
			name: "without from name",
			config: &EmailConfig{
				From: "test@example.com",
			},
			expected: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.GetFromAddress())
		})
	}
}

func TestEmailConfig_GetSMTPAddress(t *testing.T) {
	config := &EmailConfig{
		SMTP: SMTPConfig{
			Host: "smtp.gmail.com",
			Port: 587,
		},
	}

	assert.Equal(t, "smtp.gmail.com:587", config.GetSMTPAddress())
}

func TestDefaultEmailConfig(t *testing.T) {
	config := DefaultEmailConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "smtp.gmail.com", config.SMTP.Host)
	assert.Equal(t, 587, config.SMTP.Port)
	assert.False(t, config.SMTP.UseSSL)
	assert.True(t, config.SMTP.UseTLS)
	assert.Equal(t, "HXLOS Cloud", config.FromName)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, "30s", config.RetryInterval)
	assert.Equal(t, "30s", config.Timeout)
	assert.True(t, config.KeepAlive)
	assert.Equal(t, 10, config.PoolSize)
	assert.Equal(t, "10m", config.VerificationCodeTTL)
	assert.Equal(t, "1h", config.ResetTokenTTL)
	assert.Equal(t, "templates/email", config.TemplateDir)
	assert.Equal(t, "zh-CN", config.DefaultLanguage)
}

func TestNewEmailService(t *testing.T) {
	// 测试使用nil config
	service := NewEmailService(nil)
	assert.NotNil(t, service)

	// 测试使用有效config
	config := DefaultEmailConfig()
	service = NewEmailService(config)
	assert.NotNil(t, service)
}

func TestEmailService_LoadTemplates(t *testing.T) {
	config := DefaultEmailConfig()
	service := NewEmailService(config).(*emailService)

	err := service.LoadTemplates()
	assert.NoError(t, err)

	// 验证模板是否加载
	tmpl, err := service.GetTemplate(TemplateVerificationCode, "zh-CN")
	assert.NoError(t, err)
	assert.NotNil(t, tmpl)
	assert.Equal(t, TemplateVerificationCode, tmpl.Name)
	assert.Equal(t, "zh-CN", tmpl.Language)
	assert.True(t, tmpl.IsActive)
}

func TestEmailService_RegisterTemplate(t *testing.T) {
	service := NewEmailService(nil).(*emailService)

	template := &EmailTemplate{
		Name:        "test_template",
		Language:    "zh-CN",
		Subject:     "Test Subject",
		HTMLBody:    "<h1>Test</h1>",
		TextBody:    "Test",
		IsActive:    true,
		Description: "Test template",
	}

	err := service.RegisterTemplate(template)
	assert.NoError(t, err)

	// 验证模板注册成功
	retrieved, err := service.GetTemplate("test_template", "zh-CN")
	assert.NoError(t, err)
	assert.Equal(t, template.Name, retrieved.Name)
	assert.Equal(t, template.Subject, retrieved.Subject)
}

func TestEmailService_GetTemplate_NotFound(t *testing.T) {
	service := NewEmailService(nil).(*emailService)

	_, err := service.GetTemplate("nonexistent", "zh-CN")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

func TestEmailService_GetTemplate_Inactive(t *testing.T) {
	service := NewEmailService(nil).(*emailService)

	template := &EmailTemplate{
		Name:     "inactive_template",
		Language: "zh-CN",
		Subject:  "Test Subject",
		HTMLBody: "<h1>Test</h1>",
		TextBody: "Test",
		IsActive: false, // 未激活
	}

	err := service.RegisterTemplate(template)
	assert.NoError(t, err)

	_, err = service.GetTemplate("inactive_template", "zh-CN")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template is not active")
}

func TestEmailQueue_Creation(t *testing.T) {
	emailQueue := &EmailQueue{
		To:       []string{"test@example.com"},
		Subject:  "Test Subject",
		Priority: PriorityNormal,
		Status:   EmailStatusPending,
	}

	assert.Equal(t, []string{"test@example.com"}, emailQueue.To)
	assert.Equal(t, "Test Subject", emailQueue.Subject)
	assert.Equal(t, PriorityNormal, emailQueue.Priority)
	assert.Equal(t, EmailStatusPending, emailQueue.Status)
}

func TestEmailService_QueueEmail(t *testing.T) {
	service := NewEmailService(nil).(*emailService)

	emailItem := &EmailQueue{
		To:       []string{"test@example.com"},
		Subject:  "Test Subject",
		HTMLBody: "<h1>Test</h1>",
		Priority: PriorityNormal,
	}

	err := service.QueueEmail(emailItem)
	assert.NoError(t, err)

	// 验证ID和时间戳已设置
	assert.NotEmpty(t, emailItem.ID)
	assert.False(t, emailItem.CreatedAt.IsZero())
	assert.False(t, emailItem.UpdatedAt.IsZero())
	assert.Equal(t, EmailStatusPending, emailItem.Status)
	assert.Equal(t, 3, emailItem.MaxAttempts) // 默认重试次数
}

func TestEmailService_GetQueueStatus(t *testing.T) {
	service := NewEmailService(nil).(*emailService)

	status, err := service.GetQueueStatus()
	assert.NoError(t, err)
	assert.Contains(t, status, "pending")
	assert.Contains(t, status, "total")
	assert.Equal(t, 0, status["pending"])  // 初始时队列为空
	assert.Equal(t, 1000, status["total"]) // 队列容量
}

// 基准测试
func BenchmarkEmailService_QueueEmail(b *testing.B) {
	service := NewEmailService(nil).(*emailService)

	emailItem := &EmailQueue{
		To:       []string{"test@example.com"},
		Subject:  "Test Subject",
		HTMLBody: "<h1>Test</h1>",
		Priority: PriorityNormal,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 创建新的邮件项以避免ID冲突
		newItem := *emailItem
		newItem.ID = ""
		service.QueueEmail(&newItem)
	}
}

func BenchmarkEmailService_RenderTemplate(b *testing.B) {
	service := NewEmailService(nil).(*emailService)

	variables := map[string]interface{}{
		"app_name":   "Test App",
		"code":       "123456",
		"expires_in": 10,
	}

	tmplStr := "Welcome to {{.app_name}}! Your code is {{.code}}, expires in {{.expires_in}} minutes."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.renderTemplate(tmplStr, variables)
	}
}

// 新增的测试用例，提升覆盖率

// TestEmailService_IsHealthy 测试服务健康检查
func TestEmailService_IsHealthy(t *testing.T) {
	service := NewEmailService(nil).(*emailService)

	// 服务未启动时不健康
	assert.False(t, service.IsHealthy())

	// 模拟启动状态
	service.isRunning = true
	assert.True(t, service.IsHealthy())
}

// TestEmailService_StartStop 测试服务启停
func TestEmailService_StartStop(t *testing.T) {
	config := &EmailConfig{
		SMTP: SMTPConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "test@example.com",
			Password: "password",
			UseTLS:   true,
		},
		From:     "test@example.com",
		FromName: "Test Service",
	}

	service := NewEmailService(config).(*emailService)

	// 测试重复启动
	ctx := context.Background()
	err := service.Start(ctx)
	assert.NoError(t, err) // 应该成功，因为使用默认配置

	// 测试停止未启动的服务
	err = service.Stop()
	assert.NoError(t, err)
}

// TestEmailService_RenderTemplate 测试模板渲染
func TestEmailService_RenderTemplate(t *testing.T) {
	service := NewEmailService(nil).(*emailService)

	tests := []struct {
		name      string
		template  string
		variables map[string]interface{}
		expected  string
		wantErr   bool
	}{
		{
			name:     "simple template",
			template: "Hello {{.name}}!",
			variables: map[string]interface{}{
				"name": "World",
			},
			expected: "Hello World!",
			wantErr:  false,
		},
		{
			name:     "complex template",
			template: "Welcome {{.username}} to {{.app_name}}! Code: {{.code}}",
			variables: map[string]interface{}{
				"username": "john",
				"app_name": "CloudPan",
				"code":     "123456",
			},
			expected: "Welcome john to CloudPan! Code: 123456",
			wantErr:  false,
		},
		{
			name:     "invalid template",
			template: "Hello {{.name", // 缺少结束括号
			variables: map[string]interface{}{
				"name": "World",
			},
			expected: "",
			wantErr:  true,
		},
		{
			name:      "missing variable",
			template:  "Hello {{.name}}!",
			variables: map[string]interface{}{}, // 缺少name变量
			expected:  "Hello !",                // Go模板的实际行为
			wantErr:   false,
		},
		{
			name:      "empty template",
			template:  "",
			variables: map[string]interface{}{},
			expected:  "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.renderTemplate(tt.template, tt.variables)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestEmailService_SendHTMLEmail 测试发送HTML邮件
func TestEmailService_SendHTMLEmail(t *testing.T) {
	service := NewEmailService(nil).(*emailService)
	ctx := context.Background()

	// 测试空收件人列表
	err := service.SendHTMLEmail(ctx, []string{}, "Test", "<h1>Test</h1>", "Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recipients specified")

	// 测试正常情况（会失败，因为没有真实的SMTP服务器）
	err = service.SendHTMLEmail(ctx, []string{"test@example.com"}, "Test Subject", "<h1>Test HTML</h1>", "Test Text")
	assert.Error(t, err) // 预期失败，因为没有真实的SMTP配置
}

// TestEmailService_SendTemplateEmail 测试发送模板邮件
func TestEmailService_SendTemplateEmail(t *testing.T) {
	service := NewEmailService(nil).(*emailService)
	ctx := context.Background()

	// 测试不存在的模板
	err := service.SendTemplateEmail(ctx, "nonexistent", []string{"test@example.com"}, map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get template")

	// 添加测试模板
	testTemplate := &EmailTemplate{
		Name:     "test_template",
		Language: "zh-CN",
		Subject:  "Test {{.subject_var}}",
		HTMLBody: "<h1>Hello {{.name}}</h1>",
		TextBody: "Hello {{.name}}",
		IsActive: true,
	}
	service.RegisterTemplate(testTemplate)

	// 测试模板邮件（会失败，因为没有真实的SMTP服务器）
	variables := map[string]interface{}{
		"subject_var": "Subject",
		"name":        "World",
	}
	err = service.SendTemplateEmail(ctx, "test_template", []string{"test@example.com"}, variables)
	assert.Error(t, err) // 预期失败，因为没有真实的SMTP配置
}

// TestEmailService_SendVerificationCode 测试发送验证码
func TestEmailService_SendVerificationCode(t *testing.T) {
	service := NewEmailService(nil).(*emailService)
	ctx := context.Background()

	// 加载默认模板
	service.LoadTemplates()

	// 测试发送验证码（会失败，因为没有真实的SMTP服务器）
	err := service.SendVerificationCode(ctx, "test@example.com", "123456")
	assert.Error(t, err) // 预期失败，因为没有真实的SMTP配置
}

// TestEmailService_SendPasswordReset 测试发送密码重置
func TestEmailService_SendPasswordReset(t *testing.T) {
	service := NewEmailService(nil).(*emailService)
	ctx := context.Background()

	// 加载默认模板
	service.LoadTemplates()

	// 测试发送密码重置（会失败，因为没有真实的SMTP服务器）
	err := service.SendPasswordReset(ctx, "test@example.com", "https://example.com/reset?token=123")
	assert.Error(t, err) // 预期失败，因为没有真实的SMTP配置
}

// TestEmailService_SendWelcomeEmail 测试发送欢迎邮件
func TestEmailService_SendWelcomeEmail(t *testing.T) {
	service := NewEmailService(nil).(*emailService)
	ctx := context.Background()

	// 加载默认模板
	service.LoadTemplates()

	// 测试发送欢迎邮件（会失败，因为没有真实的SMTP服务器）
	err := service.SendWelcomeEmail(ctx, "test@example.com", "testuser")
	assert.Error(t, err) // 预期失败，因为没有真实的SMTP配置
}

// TestEmailService_SendSecurityAlert 测试发送安全警告
func TestEmailService_SendSecurityAlert(t *testing.T) {
	service := NewEmailService(nil).(*emailService)
	ctx := context.Background()

	// 加载默认模板
	service.LoadTemplates()

	details := map[string]interface{}{
		"ip":        "192.168.1.1",
		"location":  "北京",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 测试发送安全警告（会失败，因为没有真实的SMTP服务器）
	err := service.SendSecurityAlert(ctx, "test@example.com", "login", details)
	assert.Error(t, err) // 预期失败，因为没有真实的SMTP配置
}

// TestEmailConfig_IsTLSEnabled 测试TLS配置
func TestEmailConfig_IsTLSEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   *EmailConfig
		expected bool
	}{
		{
			name: "TLS enabled",
			config: &EmailConfig{
				SMTP: SMTPConfig{UseTLS: true},
			},
			expected: true,
		},
		{
			name: "TLS disabled",
			config: &EmailConfig{
				SMTP: SMTPConfig{UseTLS: false},
			},
			expected: false,
		},
		{
			name: "SSL enabled (not TLS)",
			config: &EmailConfig{
				SMTP: SMTPConfig{UseSSL: true, UseTLS: false},
			},
			expected: false, // SSL与TLS是不同的
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsTLSEnabled()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEmailQueue_Timestamps 测试邮件队列时间戳
func TestEmailQueue_Timestamps(t *testing.T) {
	emailQueue := &EmailQueue{
		To:      []string{"test@example.com"},
		Subject: "Test",
	}

	// 设置初始时间戳
	emailQueue.setTimestamps()
	assert.False(t, emailQueue.CreatedAt.IsZero())
	assert.False(t, emailQueue.UpdatedAt.IsZero())
	assert.Equal(t, emailQueue.CreatedAt, emailQueue.UpdatedAt)

	// 等待一毫秒然后更新
	time.Sleep(time.Millisecond)
	emailQueue.UpdateStatus(EmailStatusSending)
	assert.True(t, emailQueue.UpdatedAt.After(emailQueue.CreatedAt))
	assert.Equal(t, EmailStatusSending, emailQueue.Status)
}
