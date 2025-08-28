package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"sync"
	"time"
)

// SMTPConnection SMTP连接包装
type SMTPConnection struct {
	client    *smtp.Client
	createdAt time.Time
	lastUsed  time.Time
	inUse     bool
}

// smtpPool SMTP连接池
type smtpPool struct {
	config      *EmailConfig
	connections chan *SMTPConnection
	mu          sync.RWMutex
	closed      bool
	maxIdle     time.Duration
	maxLifetime time.Duration
}

// newSMTPPool 创建SMTP连接池
func newSMTPPool(config *EmailConfig) *smtpPool {
	pool := &smtpPool{
		config:      config,
		connections: make(chan *SMTPConnection, config.PoolSize),
		maxIdle:     5 * time.Minute,  // 最大空闲时间
		maxLifetime: 30 * time.Minute, // 最大生存时间
	}

	// 启动清理协程
	go pool.cleaner()

	return pool
}

// Get 获取SMTP连接
func (p *smtpPool) Get() (*SMTPConnection, error) {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return nil, fmt.Errorf("connection pool is closed")
	}
	p.mu.RUnlock()

	// 尝试从池中获取连接
	select {
	case conn := <-p.connections:
		if p.isValidConnection(conn) {
			conn.inUse = true
			conn.lastUsed = time.Now()
			return conn, nil
		}
		// 连接无效，关闭并创建新连接
		p.closeConnection(conn)
	default:
		// 池中没有可用连接
	}

	// 创建新连接
	return p.createConnection()
}

// Put 归还SMTP连接
func (p *smtpPool) Put(conn *SMTPConnection) {
	if conn == nil {
		return
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		p.closeConnection(conn)
		return
	}

	conn.inUse = false
	conn.lastUsed = time.Now()

	// 检查连接是否有效
	if !p.isValidConnection(conn) {
		p.closeConnection(conn)
		return
	}

	// 尝试归还到池中
	select {
	case p.connections <- conn:
		// 成功归还
	default:
		// 池已满，关闭连接
		p.closeConnection(conn)
	}
}

// Close 关闭连接池
func (p *smtpPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true
	close(p.connections)

	// 关闭所有连接
	for conn := range p.connections {
		p.closeConnection(conn)
	}
}

// IsHealthy 检查连接池健康状态
func (p *smtpPool) IsHealthy() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return !p.closed
}

// createConnection 创建新的SMTP连接
func (p *smtpPool) createConnection() (*SMTPConnection, error) {
	// 建立TCP连接
	client, err := smtp.Dial(p.config.GetSMTPAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to dial SMTP server: %w", err)
	}

	// 发送EHLO命令
	if err := client.Hello("localhost"); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			// 关闭连接时出错，但不影响主要错误的返回
		}
		return nil, fmt.Errorf("failed to send EHLO: %w", err)
	}

	// 检查是否支持STARTTLS
	if ok, _ := client.Extension("STARTTLS"); ok && p.config.IsTLSEnabled() {
		tlsConfig := &tls.Config{
			ServerName:         p.config.SMTP.Host,
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12, // 强制使用TLS 1.2或更高版本
		}

		if err := client.StartTLS(tlsConfig); err != nil {
			if closeErr := client.Close(); closeErr != nil {
				// 关闭连接时出错，但不影响主要错误的返回
			}
			return nil, fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// 进行身份验证
	auth := smtp.PlainAuth("", p.config.SMTP.Username, p.config.SMTP.Password, p.config.SMTP.Host)
	if err := client.Auth(auth); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			// 关闭连接时出错，但不影响主要错误的返回
		}
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	conn := &SMTPConnection{
		client:    client,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		inUse:     true,
	}

	return conn, nil
}

// isValidConnection 检查连接是否有效
func (p *smtpPool) isValidConnection(conn *SMTPConnection) bool {
	if conn == nil || conn.client == nil {
		return false
	}

	now := time.Now()

	// 检查连接是否超过最大生存时间
	if now.Sub(conn.createdAt) > p.maxLifetime {
		return false
	}

	// 检查连接是否超过最大空闲时间
	if !conn.inUse && now.Sub(conn.lastUsed) > p.maxIdle {
		return false
	}

	// 尝试发送NOOP命令检查连接
	if err := conn.client.Noop(); err != nil {
		return false
	}

	return true
}

// closeConnection 关闭SMTP连接
func (p *smtpPool) closeConnection(conn *SMTPConnection) {
	if conn != nil && conn.client != nil {
		if err := conn.client.Quit(); err != nil {
			// 记录Quit命令错误，但不影响后续关闭操作
			// 这里可以添加日志记录，但为了避免循环依赖，暂时忽略
		}
		if err := conn.client.Close(); err != nil {
			// 记录Close命令错误，但不影响后续操作
			// 这里可以添加日志记录，但为了避免循环依赖，暂时忽略
		}
	}
}

// cleaner 清理过期连接
func (p *smtpPool) cleaner() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		// 等待清理时间到达
		<-ticker.C
		p.cleanupExpiredConnections()

		// 检查连接池是否已关闭
		p.mu.RLock()
		if p.closed {
			p.mu.RUnlock()
			return
		}
		p.mu.RUnlock()
	}
}

// cleanupExpiredConnections 清理过期连接
func (p *smtpPool) cleanupExpiredConnections() {
	var validConnections []*SMTPConnection

	// 获取所有连接
	for {
		select {
		case conn := <-p.connections:
			if p.isValidConnection(conn) {
				validConnections = append(validConnections, conn)
			} else {
				p.closeConnection(conn)
			}
		default:
			goto done
		}
	}

done:
	// 将有效连接放回池中
	for _, conn := range validConnections {
		// 非阻塞式归还连接到池中
		select {
		case p.connections <- conn:
			// 成功归还到池中
		default:
			// 池已满，关闭连接以防止资源泄露
			p.closeConnection(conn)
		}
	}
}

// GetStats 获取连接池统计信息
func (p *smtpPool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"available_connections": len(p.connections),
		"max_connections":       cap(p.connections),
		"is_closed":             p.closed,
		"max_idle_time":         p.maxIdle.String(),
		"max_lifetime":          p.maxLifetime.String(),
	}
}
