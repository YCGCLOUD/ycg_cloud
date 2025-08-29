package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"sync"
	"time"

	"github.com/jordan-wright/email"
)

// EmailService 邮件服务接口
//
// 提供完整的邮件发送和管理功能，包括：
// 1. 邮件发送：支持纯文本、HTML和模板邮件
// 2. 模板管理：动态加载和渲染邮件模板
// 3. 队列管理：异步邮件发送和重试机制
// 4. 连接池：SMTP连接池管理，提高性能
//
// 使用示例：
//
//	service := NewEmailService(config)
//	service.Start(ctx)
//	service.SendVerificationCode(ctx, "user@example.com", "123456")
//	service.Stop()
type EmailService interface {
	// 发送邮件
	SendEmail(ctx context.Context, to []string, subject, body string) error
	SendHTMLEmail(ctx context.Context, to []string, subject, htmlBody, textBody string) error
	SendTemplateEmail(ctx context.Context, templateName string, to []string, variables map[string]interface{}) error

	// 发送特定类型邮件
	SendVerificationCode(ctx context.Context, to string, code string) error
	SendPasswordReset(ctx context.Context, to string, resetURL string) error
	SendWelcomeEmail(ctx context.Context, to string, username string) error
	SendSecurityAlert(ctx context.Context, to string, alertType string, details map[string]interface{}) error

	// 队列管理
	QueueEmail(email *EmailQueue) error
	ProcessQueue(ctx context.Context) error
	GetQueueStatus() (map[string]int, error)

	// 模板管理
	LoadTemplates() error
	RegisterTemplate(template *EmailTemplate) error
	GetTemplate(name, language string) (*EmailTemplate, error)

	// 服务管理
	Start(ctx context.Context) error
	Stop() error
	IsHealthy() bool
}

// emailService 邮件服务实现
type emailService struct {
	config    *EmailConfig
	pool      *smtpPool
	templates map[string]*EmailTemplate
	queue     chan *EmailQueue
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	isRunning bool
}

// NewEmailService 创建邮件服务实例
func NewEmailService(config *EmailConfig) EmailService {
	if config == nil {
		config = DefaultEmailConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &emailService{
		config:    config,
		pool:      newSMTPPool(config),
		templates: make(map[string]*EmailTemplate),
		queue:     make(chan *EmailQueue, 1000), // 队列容量1000
		ctx:       ctx,
		cancel:    cancel,
	}

	return service
}

// Start 启动邮件服务
func (s *emailService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("email service is already running")
	}

	// 验证配置
	if err := s.config.Validate(); err != nil {
		return fmt.Errorf("invalid email config: %w", err)
	}

	// 加载模板
	if err := s.LoadTemplates(); err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// 启动队列处理器
	s.wg.Add(1)
	go s.queueProcessor()

	s.isRunning = true
	log.Println("Email service started successfully")
	return nil
}

// Stop 停止邮件服务
func (s *emailService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	s.cancel()
	close(s.queue)
	s.wg.Wait()

	s.pool.Close()
	s.isRunning = false
	log.Println("Email service stopped")
	return nil
}

// IsHealthy 检查服务健康状态
func (s *emailService) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning && s.pool.IsHealthy()
}

// SendEmail 发送纯文本邮件
func (s *emailService) SendEmail(ctx context.Context, to []string, subject, body string) error {
	return s.SendHTMLEmail(ctx, to, subject, "", body)
}

// SendHTMLEmail 发送HTML邮件
func (s *emailService) SendHTMLEmail(ctx context.Context, to []string, subject, htmlBody, textBody string) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	e := email.NewEmail()
	e.From = s.config.GetFromAddress()
	e.To = to
	e.Subject = subject

	if htmlBody != "" {
		e.HTML = []byte(htmlBody)
	}
	if textBody != "" {
		e.Text = []byte(textBody)
	}

	return s.sendEmail(ctx, e)
}

// SendTemplateEmail 发送模板邮件
func (s *emailService) SendTemplateEmail(ctx context.Context, templateName string, to []string, variables map[string]interface{}) error {
	tmpl, err := s.GetTemplate(templateName, s.config.DefaultLanguage)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// 渲染主题
	subject, err := s.renderTemplate(tmpl.Subject, variables)
	if err != nil {
		return fmt.Errorf("failed to render subject: %w", err)
	}

	// 渲染HTML内容
	htmlBody, err := s.renderTemplate(tmpl.HTMLBody, variables)
	if err != nil {
		return fmt.Errorf("failed to render HTML body: %w", err)
	}

	// 渲染文本内容
	textBody, err := s.renderTemplate(tmpl.TextBody, variables)
	if err != nil {
		return fmt.Errorf("failed to render text body: %w", err)
	}

	return s.SendHTMLEmail(ctx, to, subject, htmlBody, textBody)
}

// SendVerificationCode 发送验证码邮件
func (s *emailService) SendVerificationCode(ctx context.Context, to string, code string) error {
	variables := map[string]interface{}{
		"code":       code,
		"expires_in": s.config.GetVerificationCodeTTL().Minutes(),
		"app_name":   s.config.FromName,
	}

	return s.SendTemplateEmail(ctx, TemplateVerificationCode, []string{to}, variables)
}

// SendPasswordReset 发送密码重置邮件
func (s *emailService) SendPasswordReset(ctx context.Context, to string, resetURL string) error {
	variables := map[string]interface{}{
		"reset_url":  resetURL,
		"expires_in": s.config.GetResetTokenTTL().Hours(),
		"app_name":   s.config.FromName,
	}

	return s.SendTemplateEmail(ctx, TemplatePasswordReset, []string{to}, variables)
}

// SendWelcomeEmail 发送欢迎邮件
func (s *emailService) SendWelcomeEmail(ctx context.Context, to string, username string) error {
	variables := map[string]interface{}{
		"username": username,
		"app_name": s.config.FromName,
	}

	return s.SendTemplateEmail(ctx, TemplateWelcome, []string{to}, variables)
}

// SendSecurityAlert 发送安全警告邮件
func (s *emailService) SendSecurityAlert(ctx context.Context, to string, alertType string, details map[string]interface{}) error {
	variables := map[string]interface{}{
		"alert_type": alertType,
		"details":    details,
		"app_name":   s.config.FromName,
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
	}

	return s.SendTemplateEmail(ctx, TemplateSecurityAlert, []string{to}, variables)
}

// QueueEmail 将邮件加入队列
func (s *emailService) QueueEmail(emailItem *EmailQueue) error {
	if emailItem.ID == "" {
		emailItem.ID = generateEmailID()
	}
	if emailItem.CreatedAt.IsZero() {
		emailItem.CreatedAt = time.Now()
	}
	if emailItem.UpdatedAt.IsZero() {
		emailItem.UpdatedAt = time.Now()
	}
	if emailItem.Status == "" {
		emailItem.Status = EmailStatusPending
	}
	if emailItem.MaxAttempts == 0 {
		emailItem.MaxAttempts = s.config.MaxRetries
	}

	select {
	case s.queue <- emailItem:
		return nil
	default:
		return fmt.Errorf("email queue is full")
	}
}

// ProcessQueue 处理邮件队列
func (s *emailService) ProcessQueue(ctx context.Context) error {
	// 这个方法主要用于手动触发队列处理
	// 实际的队列处理在queueProcessor中进行
	return nil
}

// GetQueueStatus 获取队列状态
func (s *emailService) GetQueueStatus() (map[string]int, error) {
	status := map[string]int{
		"pending": len(s.queue),
		"total":   cap(s.queue),
	}
	return status, nil
}

// LoadTemplates 加载邮件模板
func (s *emailService) LoadTemplates() error {
	// 注册默认模板
	defaultTemplates := s.getDefaultTemplates()
	for _, tmpl := range defaultTemplates {
		s.templates[tmpl.Name+"_"+tmpl.Language] = tmpl
	}

	// 如果配置了模板目录，从文件系统加载模板
	if s.config.TemplateDir != "" {
		return s.loadTemplatesFromDir(s.config.TemplateDir)
	}

	return nil
}

// RegisterTemplate 注册模板
func (s *emailService) RegisterTemplate(template *EmailTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if template.Language == "" {
		template.Language = s.config.DefaultLanguage
	}

	key := template.Name + "_" + template.Language
	s.templates[key] = template
	return nil
}

// GetTemplate 获取模板
func (s *emailService) GetTemplate(name, language string) (*EmailTemplate, error) {
	if language == "" {
		language = s.config.DefaultLanguage
	}

	key := name + "_" + language
	template, exists := s.templates[key]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", key)
	}

	if !template.IsActive {
		return nil, fmt.Errorf("template is not active: %s", key)
	}

	return template, nil
}

// sendEmail 发送邮件的内部方法
func (s *emailService) sendEmail(ctx context.Context, e *email.Email) error {
	conn, err := s.pool.Get()
	if err != nil {
		return fmt.Errorf("failed to get SMTP connection: %w", err)
	}
	defer s.pool.Put(conn)

	// 设置超时
	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.GetTimeout())
	defer cancel()

	// 检查上下文是否已取消
	select {
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	default:
	}

	// 发送邮件
	return e.Send(s.config.GetSMTPAddress(), s.getSMTPAuth())
}

// getSMTPAuth 获取SMTP认证
func (s *emailService) getSMTPAuth() smtp.Auth {
	return smtp.PlainAuth("", s.config.SMTP.Username, s.config.SMTP.Password, s.config.SMTP.Host)
}

// renderTemplate 渲染模板
func (s *emailService) renderTemplate(tmplStr string, variables map[string]interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// queueProcessor 队列处理器
func (s *emailService) queueProcessor() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case emailItem, ok := <-s.queue:
			if !ok {
				return
			}
			s.processEmailItem(emailItem)
		}
	}
}

// processEmailItem 处理队列中的邮件项
func (s *emailService) processEmailItem(emailItem *EmailQueue) {
	emailItem.Status = EmailStatusSending
	emailItem.UpdatedAt = time.Now()

	var err error
	if emailItem.Template != "" {
		// 使用模板发送
		err = s.SendTemplateEmail(s.ctx, emailItem.Template, emailItem.To, emailItem.Variables)
	} else {
		// 直接发送
		err = s.SendHTMLEmail(s.ctx, emailItem.To, emailItem.Subject, emailItem.HTMLBody, emailItem.TextBody)
	}

	if err != nil {
		emailItem.Attempts++
		emailItem.ErrorMsg = err.Error()
		emailItem.UpdatedAt = time.Now()

		if emailItem.Attempts < emailItem.MaxAttempts {
			// 重试
			emailItem.Status = EmailStatusRetrying
			time.AfterFunc(s.config.GetRetryInterval(), func() {
				s.queue <- emailItem
			})
		} else {
			// 达到最大重试次数，标记为失败
			emailItem.Status = EmailStatusFailed
		}
	} else {
		emailItem.Status = EmailStatusSent
		emailItem.UpdatedAt = time.Now()
	}
}

// loadTemplatesFromDir 从目录加载模板
func (s *emailService) loadTemplatesFromDir(_ string) error {
	// 这里可以实现从文件系统加载模板的逻辑
	// 暂时使用默认模板
	return nil
}

// generateEmailID 生成邮件ID
func generateEmailID() string {
	return fmt.Sprintf("email_%d", time.Now().UnixNano())
}
