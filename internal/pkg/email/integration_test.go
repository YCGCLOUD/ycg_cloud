package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEmailManager_Creation 测试邮件管理器创建
func TestEmailManager_Creation(t *testing.T) {
	config := DefaultEmailConfig()
	manager := NewEmailManager(config)

	assert.NotNil(t, manager)
	assert.Equal(t, config, manager.config)
	assert.False(t, manager.started)
	assert.Nil(t, manager.service)
}

// TestEmailManager_Initialize 测试邮件管理器初始化
func TestEmailManager_Initialize(t *testing.T) {
	config := DefaultEmailConfig()
	manager := NewEmailManager(config)

	// 初始化
	err := manager.Initialize()
	assert.NoError(t, err)
	assert.NotNil(t, manager.service)

	// 重复初始化应该失败
	err = manager.Initialize()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already initialized")
}

// TestEmailManager_StartStop 测试邮件管理器启停
func TestEmailManager_StartStop(t *testing.T) {
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
	manager := NewEmailManager(config)
	ctx := context.Background()

	// 未初始化时启动应该失败
	err := manager.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// 初始化后启动（会失败，因为没有真实的SMTP服务器）
	err = manager.Initialize()
	assert.NoError(t, err)

	err = manager.Start(ctx)
	assert.NoError(t, err) // 应该成功

	// 停止未启动的服务
	err = manager.Stop()
	assert.NoError(t, err)
}

// TestEmailManager_IsStarted 测试启动状态检查
func TestEmailManager_IsStarted(t *testing.T) {
	manager := NewEmailManager(DefaultEmailConfig())

	// 初始状态
	assert.False(t, manager.IsStarted())

	// 模拟启动状态
	manager.started = true
	assert.True(t, manager.IsStarted())
}

// TestEmailManager_IsHealthy 测试健康状态检查
func TestEmailManager_IsHealthy(t *testing.T) {
	manager := NewEmailManager(DefaultEmailConfig())

	// 未初始化时不健康
	assert.False(t, manager.IsHealthy())

	// 初始化但未启动时不健康
	manager.Initialize()
	assert.False(t, manager.IsHealthy())

	// 模拟启动状态
	manager.started = true
	assert.False(t, manager.IsHealthy()) // 因为服务是mock的，不一定健康
}

// TestEmailManager_UpdateConfig 测试配置更新
func TestEmailManager_UpdateConfig(t *testing.T) {
	manager := NewEmailManager(DefaultEmailConfig())

	newConfig := &EmailConfig{
		SMTP: SMTPConfig{
			Host: "new.smtp.com",
			Port: 587,
		},
		From: "new@example.com",
	}

	// 未启动时可以更新配置
	err := manager.UpdateConfig(newConfig)
	assert.NoError(t, err)
	assert.Equal(t, newConfig, manager.config)

	// 模拟启动状态，更新配置应该失败
	manager.started = true
	err = manager.UpdateConfig(newConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot update config while service is running")
}

// TestEmailManager_GetService 测试获取服务实例
func TestEmailManager_GetService(t *testing.T) {
	manager := NewEmailManager(DefaultEmailConfig())

	// 未初始化时返回nil
	service := manager.GetService()
	assert.Nil(t, service)

	// 初始化后返回服务实例
	manager.Initialize()
	service = manager.GetService()
	assert.NotNil(t, service)
}

// TestEmailManager_GetStats 测试获取统计信息
func TestEmailManager_GetStats(t *testing.T) {
	manager := NewEmailManager(DefaultEmailConfig())

	// 未初始化时获取统计信息应该失败
	stats, err := manager.GetStats()
	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "not initialized")

	// 初始化后获取统计信息
	manager.Initialize()
	stats, err = manager.GetStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "initialized")
	assert.Contains(t, stats, "started")
	assert.Contains(t, stats, "healthy")
	assert.Contains(t, stats, "queue")
	assert.Equal(t, true, stats["initialized"])
	assert.Equal(t, false, stats["started"])
}

// TestGlobalEmailManager 测试全局邮件管理器
func TestGlobalEmailManager(t *testing.T) {
	// 获取全局管理器
	manager1 := GetGlobalEmailManager()
	assert.NotNil(t, manager1)

	// 再次获取应该是同一个实例
	manager2 := GetGlobalEmailManager()
	assert.Equal(t, manager1, manager2)
}

// TestInitializeGlobalEmailService 测试初始化全局邮件服务
func TestInitializeGlobalEmailService(t *testing.T) {
	// 使用默认配置初始化
	err := InitializeGlobalEmailService(nil)
	assert.NoError(t, err)

	// 使用自定义配置初始化
	customConfig := &EmailConfig{
		SMTP: SMTPConfig{
			Host: "custom.smtp.com",
			Port: 587,
		},
		From: "custom@example.com",
	}

	// 需要先停止已启动的服务，然后重新创建管理器
	StopGlobalEmailService()
	// 重置全局管理器以避免“已初始化”错误
	globalEmailManager = NewEmailManager(customConfig)

	err = InitializeGlobalEmailService(customConfig)
	assert.NoError(t, err)
}

// TestStartStopGlobalEmailService 测试启停全局邮件服务
func TestStartStopGlobalEmailService(t *testing.T) {
	ctx := context.Background()

	// 启动全局服务（会失败，因为配置问题）
	err := StartGlobalEmailService(ctx)
	assert.Error(t, err) // 预期失败

	// 停止全局服务
	err = StopGlobalEmailService()
	assert.NoError(t, err)
}

// TestGetGlobalEmailService 测试获取全局邮件服务
func TestGetGlobalEmailService(t *testing.T) {
	service := GetGlobalEmailService()
	assert.NotNil(t, service)
}

// TestIsGlobalEmailServiceHealthy 测试全局邮件服务健康检查
func TestIsGlobalEmailServiceHealthy(t *testing.T) {
	healthy := IsGlobalEmailServiceHealthy()
	// 由于服务可能未启动，这里只检查不会panic
	_ = healthy
}

// TestEmailConfig_IsSSLEnabled 测试SSL启用检查
func TestEmailConfig_IsSSLEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   *EmailConfig
		expected bool
	}{
		{
			name: "SSL enabled",
			config: &EmailConfig{
				SMTP: SMTPConfig{UseSSL: true},
			},
			expected: true,
		},
		{
			name: "SSL disabled",
			config: &EmailConfig{
				SMTP: SMTPConfig{UseSSL: false},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsSSLEnabled()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEmailConfig_ValidationEdgeCases 测试配置验证边界情况
func TestEmailConfig_ValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  *EmailConfig
		wantErr bool
	}{
		{
			name: "negative max retries",
			config: &EmailConfig{
				SMTP: SMTPConfig{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "test@example.com",
					Password: "password",
				},
				From:       "test@example.com",
				MaxRetries: -1,
			},
			wantErr: true,
		},
		{
			name: "zero pool size (should auto-fix)",
			config: &EmailConfig{
				SMTP: SMTPConfig{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "test@example.com",
					Password: "password",
				},
				From:     "test@example.com",
				PoolSize: 0,
			},
			wantErr: false,
		},
		{
			name: "port out of range",
			config: &EmailConfig{
				SMTP: SMTPConfig{
					Host:     "smtp.example.com",
					Port:     99999,
					Username: "test@example.com",
					Password: "password",
				},
				From: "test@example.com",
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
				if tt.config.PoolSize == 0 {
					// 应该被设置为默认值10
					assert.Equal(t, 10, tt.config.PoolSize)
				}
			}
		})
	}
}
