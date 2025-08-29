package email

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSMTPPool_Creation(t *testing.T) {
	config := &EmailConfig{
		SMTP: SMTPConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "test@example.com",
			Password: "password",
			UseTLS:   true,
		},
		PoolSize: 5,
	}

	pool := newSMTPPool(config)
	assert.NotNil(t, pool)
	assert.Equal(t, config, pool.config)
	assert.Equal(t, 5*time.Minute, pool.maxIdle)
	assert.Equal(t, 30*time.Minute, pool.maxLifetime)
	assert.False(t, pool.closed)

	// 清理
	pool.Close()
}

func TestSMTPPool_IsHealthy(t *testing.T) {
	config := DefaultEmailConfig()
	pool := newSMTPPool(config)

	// 新创建的连接池应该是健康的
	assert.True(t, pool.IsHealthy())

	// 关闭后应该不健康
	pool.Close()
	assert.False(t, pool.IsHealthy())
}

func TestSMTPPool_Close(t *testing.T) {
	config := DefaultEmailConfig()
	pool := newSMTPPool(config)

	// 第一次关闭
	pool.Close()
	assert.True(t, pool.closed)

	// 重复关闭应该不会出错
	pool.Close()
	assert.True(t, pool.closed)
}

func TestSMTPPool_GetConnection_PoolClosed(t *testing.T) {
	config := DefaultEmailConfig()
	pool := newSMTPPool(config)

	// 关闭连接池
	pool.Close()

	// 尝试获取连接应该失败
	conn, err := pool.Get()
	assert.Error(t, err)
	assert.Nil(t, conn)
	assert.Contains(t, err.Error(), "connection pool is closed")
}

func TestSMTPPool_PutConnection_PoolClosed(t *testing.T) {
	config := DefaultEmailConfig()
	pool := newSMTPPool(config)

	// 创建一个模拟连接
	mockConn := &SMTPConnection{
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		inUse:     false,
	}

	// 关闭连接池
	pool.Close()

	// 尝试归还连接应该直接关闭连接
	pool.Put(mockConn)
	// 如果能执行到这里说明没有panic，测试通过
}

func TestSMTPPool_PutNilConnection(t *testing.T) {
	config := DefaultEmailConfig()
	pool := newSMTPPool(config)
	defer pool.Close()

	// 归还nil连接应该安全处理
	pool.Put(nil)
	// 如果能执行到这里说明没有panic，测试通过
}

func TestSMTPConnection_IsValid(t *testing.T) {
	config := DefaultEmailConfig()
	pool := newSMTPPool(config)
	defer pool.Close()

	now := time.Now()

	tests := []struct {
		name     string
		conn     *SMTPConnection
		expected bool
		maxIdle  time.Duration
		maxLife  time.Duration
	}{
		{
			name:     "nil connection",
			conn:     nil,
			expected: false,
		},
		{
			name: "connection with nil client",
			conn: &SMTPConnection{
				client:    nil,
				createdAt: now,
				lastUsed:  now,
			},
			expected: false,
		},
		{
			name: "expired connection (lifetime)",
			conn: &SMTPConnection{
				createdAt: now.Add(-2 * time.Hour), // 超过最大生存时间
				lastUsed:  now,
			},
			expected: false,
			maxLife:  time.Hour,
		},
		{
			name: "idle too long",
			conn: &SMTPConnection{
				createdAt: now,
				lastUsed:  now.Add(-20 * time.Minute), // 超过最大空闲时间
				inUse:     false,
			},
			expected: false,
			maxIdle:  10 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.maxIdle > 0 {
				pool.maxIdle = tt.maxIdle
			}
			if tt.maxLife > 0 {
				pool.maxLifetime = tt.maxLife
			}
			result := pool.isValidConnection(tt.conn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSMTPConnection_UpdateUsage(t *testing.T) {
	now := time.Now()
	conn := &SMTPConnection{
		createdAt: now,
		lastUsed:  now,
		inUse:     false,
	}

	// 验证初始状态
	assert.Equal(t, now, conn.createdAt)
	assert.Equal(t, now, conn.lastUsed)
	assert.False(t, conn.inUse)

	// 标记为使用中
	conn.inUse = true
	conn.lastUsed = time.Now()

	assert.True(t, conn.inUse)
	assert.True(t, conn.lastUsed.After(now) || conn.lastUsed.Equal(now))
}

// TestEmailTemplate_Validation 测试邮件模板验证
func TestEmailTemplate_Validation(t *testing.T) {
	tests := []struct {
		name     string
		template *EmailTemplate
		wantErr  bool
	}{
		{
			name: "valid template",
			template: &EmailTemplate{
				Name:        "test",
				Language:    "zh-CN",
				Subject:     "Test Subject",
				HTMLBody:    "<h1>Test</h1>",
				TextBody:    "Test",
				IsActive:    true,
				Description: "Test template",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			template: &EmailTemplate{
				Language: "zh-CN",
				Subject:  "Test Subject",
				HTMLBody: "<h1>Test</h1>",
			},
			wantErr: true,
		},
		{
			name: "missing language",
			template: &EmailTemplate{
				Name:     "test",
				Subject:  "Test Subject",
				HTMLBody: "<h1>Test</h1>",
			},
			wantErr: true,
		},
		{
			name: "missing subject",
			template: &EmailTemplate{
				Name:     "test",
				Language: "zh-CN",
				HTMLBody: "<h1>Test</h1>",
			},
			wantErr: true,
		},
		{
			name: "missing both bodies",
			template: &EmailTemplate{
				Name:     "test",
				Language: "zh-CN",
				Subject:  "Test Subject",
			},
			wantErr: true,
		},
		{
			name: "only HTML body",
			template: &EmailTemplate{
				Name:     "test",
				Language: "zh-CN",
				Subject:  "Test Subject",
				HTMLBody: "<h1>Test</h1>",
			},
			wantErr: false,
		},
		{
			name: "only text body",
			template: &EmailTemplate{
				Name:     "test",
				Language: "zh-CN",
				Subject:  "Test Subject",
				TextBody: "Test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEmailQueue_Methods 测试邮件队列方法
func TestEmailQueue_Methods(t *testing.T) {
	emailQueue := &EmailQueue{
		To:      []string{"test@example.com"},
		Subject: "Test Subject",
		Status:  EmailStatusPending,
	}

	// 测试设置时间戳
	emailQueue.setTimestamps()
	assert.False(t, emailQueue.CreatedAt.IsZero())
	assert.False(t, emailQueue.UpdatedAt.IsZero())

	// 测试更新状态
	oldUpdateTime := emailQueue.UpdatedAt
	time.Sleep(time.Millisecond) // 确保时间不同
	emailQueue.UpdateStatus(EmailStatusSending)
	assert.Equal(t, EmailStatusSending, emailQueue.Status)
	assert.True(t, emailQueue.UpdatedAt.After(oldUpdateTime))

	// 测试设置错误
	emailQueue.SetError("test error")
	assert.Equal(t, EmailStatusFailed, emailQueue.Status)
	assert.Equal(t, "test error", emailQueue.ErrorMsg)

	// 测试重置重试
	emailQueue.ResetRetry()
	assert.Equal(t, 0, emailQueue.Attempts)
	assert.Empty(t, emailQueue.ErrorMsg)
	assert.Equal(t, EmailStatusPending, emailQueue.Status)
}

// TestPriorityConstants 测试优先级常量
func TestPriorityConstants(t *testing.T) {
	assert.Equal(t, 1, PriorityHigh)
	assert.Equal(t, 2, PriorityNormal)
	assert.Equal(t, 3, PriorityLow)
}

// TestTemplateConstants 测试模板常量
func TestTemplateConstants(t *testing.T) {
	expectedTemplates := []string{
		TemplateVerificationCode,
		TemplatePasswordReset,
		TemplateWelcome,
		TemplateAccountLocked,
		TemplateSecurityAlert,
		TemplateTeamInvitation,
		TemplateFileShared,
	}

	for _, template := range expectedTemplates {
		assert.NotEmpty(t, template, "Template constant should not be empty")
	}
}

// TestEmailStatusConstants 测试邮件状态常量
func TestEmailStatusConstants(t *testing.T) {
	expectedStatuses := []string{
		EmailStatusPending,
		EmailStatusSending,
		EmailStatusSent,
		EmailStatusFailed,
		EmailStatusCancelled,
	}

	for _, status := range expectedStatuses {
		assert.NotEmpty(t, status, "Email status constant should not be empty")
	}
}
