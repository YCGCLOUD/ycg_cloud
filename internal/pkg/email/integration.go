package email

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// EmailManager 邮件管理器，负责管理邮件服务的生命周期
type EmailManager struct {
	service EmailService
	config  *EmailConfig
	mu      sync.RWMutex
	started bool
}

// NewEmailManager 创建邮件管理器
func NewEmailManager(config *EmailConfig) *EmailManager {
	return &EmailManager{
		config: config,
	}
}

// Initialize 初始化邮件服务
func (m *EmailManager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.service != nil {
		return fmt.Errorf("email service already initialized")
	}

	// 创建邮件服务
	m.service = NewEmailService(m.config)

	log.Println("Email service initialized successfully")
	return nil
}

// Start 启动邮件服务
func (m *EmailManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("email service already started")
	}

	if m.service == nil {
		return fmt.Errorf("email service not initialized")
	}

	if err := m.service.Start(ctx); err != nil {
		return fmt.Errorf("failed to start email service: %w", err)
	}

	m.started = true
	log.Println("Email service started successfully")
	return nil
}

// Stop 停止邮件服务
func (m *EmailManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	if m.service == nil {
		return fmt.Errorf("email service not initialized")
	}

	if err := m.service.Stop(); err != nil {
		return fmt.Errorf("failed to stop email service: %w", err)
	}

	m.started = false
	log.Println("Email service stopped successfully")
	return nil
}

// GetService 获取邮件服务实例
func (m *EmailManager) GetService() EmailService {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.service
}

// IsStarted 检查服务是否已启动
func (m *EmailManager) IsStarted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.started
}

// IsHealthy 检查服务健康状态
func (m *EmailManager) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.started || m.service == nil {
		return false
	}

	return m.service.IsHealthy()
}

// UpdateConfig 更新配置（需要重启服务生效）
func (m *EmailManager) UpdateConfig(config *EmailConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("cannot update config while service is running")
	}

	m.config = config
	if m.service != nil {
		// 重新创建服务实例
		m.service = NewEmailService(config)
	}

	return nil
}

// GetStats 获取邮件服务统计信息
func (m *EmailManager) GetStats() (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.service == nil {
		return nil, fmt.Errorf("email service not initialized")
	}

	queueStatus, err := m.service.GetQueueStatus()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"initialized": m.service != nil,
		"started":     m.started,
		"healthy":     m.IsHealthy(),
		"queue":       queueStatus,
	}

	return stats, nil
}

// 全局邮件管理器实例
var (
	globalEmailManager *EmailManager
	globalEmailOnce    sync.Once
)

// GetGlobalEmailManager 获取全局邮件管理器实例
func GetGlobalEmailManager() *EmailManager {
	globalEmailOnce.Do(func() {
		globalEmailManager = NewEmailManager(DefaultEmailConfig())
	})
	return globalEmailManager
}

// InitializeGlobalEmailService 初始化全局邮件服务
func InitializeGlobalEmailService(config *EmailConfig) error {
	manager := GetGlobalEmailManager()

	if config != nil {
		if err := manager.UpdateConfig(config); err != nil {
			return err
		}
	}

	return manager.Initialize()
}

// StartGlobalEmailService 启动全局邮件服务
func StartGlobalEmailService(ctx context.Context) error {
	manager := GetGlobalEmailManager()
	return manager.Start(ctx)
}

// StopGlobalEmailService 停止全局邮件服务
func StopGlobalEmailService() error {
	manager := GetGlobalEmailManager()
	return manager.Stop()
}

// GetGlobalEmailService 获取全局邮件服务实例
func GetGlobalEmailService() EmailService {
	manager := GetGlobalEmailManager()
	return manager.GetService()
}

// IsGlobalEmailServiceHealthy 检查全局邮件服务健康状态
func IsGlobalEmailServiceHealthy() bool {
	manager := GetGlobalEmailManager()
	return manager.IsHealthy()
}

// 便捷函数，直接使用全局邮件服务

// SendVerificationCodeGlobal 发送验证码（使用全局服务）
func SendVerificationCodeGlobal(ctx context.Context, to string, code string) error {
	service := GetGlobalEmailService()
	if service == nil {
		return fmt.Errorf("email service not available")
	}
	return service.SendVerificationCode(ctx, to, code)
}

// SendPasswordResetGlobal 发送密码重置邮件（使用全局服务）
func SendPasswordResetGlobal(ctx context.Context, to string, resetURL string) error {
	service := GetGlobalEmailService()
	if service == nil {
		return fmt.Errorf("email service not available")
	}
	return service.SendPasswordReset(ctx, to, resetURL)
}

// SendWelcomeEmailGlobal 发送欢迎邮件（使用全局服务）
func SendWelcomeEmailGlobal(ctx context.Context, to string, username string) error {
	service := GetGlobalEmailService()
	if service == nil {
		return fmt.Errorf("email service not available")
	}
	return service.SendWelcomeEmail(ctx, to, username)
}

// SendSecurityAlertGlobal 发送安全警告邮件（使用全局服务）
func SendSecurityAlertGlobal(ctx context.Context, to string, alertType string, details map[string]interface{}) error {
	service := GetGlobalEmailService()
	if service == nil {
		return fmt.Errorf("email service not available")
	}
	return service.SendSecurityAlert(ctx, to, alertType, details)
}

// QueueEmailGlobal 将邮件加入全局队列
func QueueEmailGlobal(email *EmailQueue) error {
	service := GetGlobalEmailService()
	if service == nil {
		return fmt.Errorf("email service not available")
	}
	return service.QueueEmail(email)
}

// GetGlobalEmailStats 获取全局邮件服务统计信息
func GetGlobalEmailStats() (map[string]interface{}, error) {
	manager := GetGlobalEmailManager()
	return manager.GetStats()
}
