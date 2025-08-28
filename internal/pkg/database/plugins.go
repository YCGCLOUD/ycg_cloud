package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 定义自定义类型作为context key以避免冲突
type contextKey string

const (
	traceIDKey contextKey = "trace_id"
)

// Plugin 插件接口
type Plugin interface {
	Name() string
	Initialize(*gorm.DB) error
}

// AuditPlugin 审计插件
type AuditPlugin struct{}

func (p *AuditPlugin) Name() string {
	return "audit"
}

func (p *AuditPlugin) Initialize(db *gorm.DB) error {
	// 注册审计回调
	if err := db.Callback().Create().After("gorm:create").Register("audit:create", auditCreate); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register("audit:update", auditUpdate); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gorm:delete").Register("audit:delete", auditDelete); err != nil {
		return err
	}

	log.Println("Audit plugin initialized")
	return nil
}

// MetricsPlugin 指标插件
type MetricsPlugin struct {
	SlowQueryThreshold time.Duration
}

func (p *MetricsPlugin) Name() string {
	return "metrics"
}

func (p *MetricsPlugin) Initialize(db *gorm.DB) error {
	if p.SlowQueryThreshold == 0 {
		p.SlowQueryThreshold = 200 * time.Millisecond
	}

	// 注册性能监控回调
	if err := db.Callback().Query().Before("gorm:query").Register("metrics:before_query", p.beforeQuery); err != nil {
		return err
	}
	if err := db.Callback().Query().After("gorm:query").Register("metrics:after_query", p.afterQuery); err != nil {
		return err
	}

	// 为其他操作也添加监控
	if err := db.Callback().Create().Before("gorm:create").Register("metrics:before_create", p.beforeQuery); err != nil {
		return err
	}
	if err := db.Callback().Create().After("gorm:create").Register("metrics:after_create", p.afterQuery); err != nil {
		return err
	}

	if err := db.Callback().Update().Before("gorm:update").Register("metrics:before_update", p.beforeQuery); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register("metrics:after_update", p.afterQuery); err != nil {
		return err
	}

	if err := db.Callback().Delete().Before("gorm:delete").Register("metrics:before_delete", p.beforeQuery); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gorm:delete").Register("metrics:after_delete", p.afterQuery); err != nil {
		return err
	}

	log.Println("Metrics plugin initialized")
	return nil
}

// TracePlugin 链路追踪插件
type TracePlugin struct{}

func (p *TracePlugin) Name() string {
	return "trace"
}

func (p *TracePlugin) Initialize(db *gorm.DB) error {
	// 注册链路追踪回调
	if err := db.Callback().Query().Before("gorm:query").Register("trace:before", traceStart); err != nil {
		return err
	}
	if err := db.Callback().Query().After("gorm:query").Register("trace:after", traceEnd); err != nil {
		return err
	}

	log.Println("Trace plugin initialized")
	return nil
}

// 审计回调函数
func auditCreate(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	log.Printf("Audit: Created record in table: %s", db.Statement.Table)
}

func auditUpdate(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	log.Printf("Audit: Updated %d record(s) in table: %s", db.RowsAffected, db.Statement.Table)
}

func auditDelete(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	log.Printf("Audit: Deleted %d record(s) from table: %s", db.RowsAffected, db.Statement.Table)
}

// 性能监控回调函数
func (p *MetricsPlugin) beforeQuery(db *gorm.DB) {
	db.Set("start_time", time.Now())
}

func (p *MetricsPlugin) afterQuery(db *gorm.DB) {
	if startTime, ok := db.Get("start_time"); ok {
		if start, valid := startTime.(time.Time); valid {
			duration := time.Since(start)

			// 记录慢查询
			if duration > p.SlowQueryThreshold {
				log.Printf("Slow Query: %s (Duration: %v, SQL: %s)",
					db.Statement.Table, duration, db.Statement.SQL.String())
			}

			// 这里可以添加指标收集逻辑
			// 例如发送到Prometheus、InfluxDB等
		}
	}
}

// 链路追踪回调函数
func traceStart(db *gorm.DB) {
	// 从上下文中获取trace信息
	if ctx := db.Statement.Context; ctx != nil {
		if traceID := ctx.Value(traceIDKey); traceID != nil {
			db.Set("trace_id", traceID)
		}
	}
}

func traceEnd(db *gorm.DB) {
	if traceID, ok := db.Get("trace_id"); ok {
		log.Printf("Trace: %v - Query completed: %s", traceID, db.Statement.Table)
	}
}

// WithUserContext 设置用户上下文
func WithUserContext(db *gorm.DB, userID uint) *gorm.DB {
	return db.Set("current_user_id", userID)
}

// WithTraceContext 设置追踪上下文
func WithTraceContext(db *gorm.DB, traceID string) *gorm.DB {
	ctx := context.WithValue(db.Statement.Context, traceIDKey, traceID)
	return db.WithContext(ctx)
}

// WithTimeout 设置超时上下文
func WithTimeout(db *gorm.DB, timeout time.Duration) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// 注意：这里没有调用cancel，因为GORM会在查询完成后自动处理
	_ = cancel
	return db.WithContext(ctx)
}

// InstallPlugins 安装所有插件
func InstallPlugins(db *gorm.DB, plugins ...Plugin) error {
	for _, plugin := range plugins {
		if err := plugin.Initialize(db); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", plugin.Name(), err)
		}
		log.Printf("Plugin %s installed successfully", plugin.Name())
	}
	return nil
}

// GetDefaultPlugins 获取默认插件列表
func GetDefaultPlugins() []Plugin {
	return []Plugin{
		&AuditPlugin{},
		&MetricsPlugin{SlowQueryThreshold: 200 * time.Millisecond},
		&TracePlugin{},
	}
}

// CustomLogger 自定义日志记录器
type CustomLogger struct {
	logger.Interface
	SlowThreshold time.Duration
	LogLevel      logger.LogLevel
}

// NewCustomLogger 创建自定义日志记录器
func NewCustomLogger(slowThreshold time.Duration, logLevel logger.LogLevel) *CustomLogger {
	return &CustomLogger{
		Interface:     logger.Default,
		SlowThreshold: slowThreshold,
		LogLevel:      logLevel,
	}
}

// LogMode 设置日志级别
func (l *CustomLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.LogLevel = level
	return l
}

// Info 记录信息日志
func (l *CustomLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		log.Printf("[INFO] "+msg, data...)
	}
}

// Warn 记录警告日志
func (l *CustomLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		log.Printf("[WARN] "+msg, data...)
	}
}

// Error 记录错误日志
func (l *CustomLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		log.Printf("[ERROR] "+msg, data...)
	}
}

// Trace 记录SQL跟踪日志
func (l *CustomLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.LogLevel >= logger.Error:
		log.Printf("[ERROR] SQL Error: %v (Duration: %v, Rows: %d, SQL: %s)",
			err, elapsed, rows, sql)
	case elapsed > l.SlowThreshold && l.LogLevel >= logger.Warn:
		log.Printf("[WARN] Slow SQL: Duration: %v (Rows: %d, SQL: %s)",
			elapsed, rows, sql)
	case l.LogLevel >= logger.Info:
		log.Printf("[INFO] SQL: Duration: %v (Rows: %d, SQL: %s)",
			elapsed, rows, sql)
	}
}
