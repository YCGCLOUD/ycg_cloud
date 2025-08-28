package logger

import (
	"time"

	"go.uber.org/zap"
)

// StructuredLog 结构化日志帮助器
type StructuredLog struct {
	logger *zap.Logger
}

// NewStructuredLog 创建结构化日志实例
func NewStructuredLog(logger *zap.Logger) *StructuredLog {
	return &StructuredLog{logger: logger}
}

// UserAction 用户操作日志
type UserAction struct {
	UserID     string                 `json:"user_id"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource,omitempty"`
	ResourceID string                 `json:"resource_id,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// DatabaseOperation 数据库操作日志
type DatabaseOperation struct {
	Operation    string        `json:"operation"`
	Table        string        `json:"table"`
	Duration     time.Duration `json:"duration"`
	RowsAffected int64         `json:"rows_affected,omitempty"`
	Error        string        `json:"error,omitempty"`
	SQL          string        `json:"sql,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

// FileOperation 文件操作日志
type FileOperation struct {
	UserID      string        `json:"user_id"`
	Operation   string        `json:"operation"` // upload, download, delete, etc.
	FileName    string        `json:"file_name"`
	FileSize    int64         `json:"file_size,omitempty"`
	FileType    string        `json:"file_type,omitempty"`
	StoragePath string        `json:"storage_path,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
	Error       string        `json:"error,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// SecurityEvent 安全事件日志
type SecurityEvent struct {
	EventType   string                 `json:"event_type"` // login_failed, permission_denied, etc.
	UserID      string                 `json:"user_id,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Resource    string                 `json:"resource,omitempty"`
	Severity    string                 `json:"severity"` // low, medium, high, critical
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// SystemEvent 系统事件日志
type SystemEvent struct {
	Component string                 `json:"component"`
	Event     string                 `json:"event"`
	Level     string                 `json:"level"` // info, warn, error
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// LogUserAction 记录用户操作日志
func (sl *StructuredLog) LogUserAction(action UserAction) {
	action.Timestamp = time.Now()

	fields := []zap.Field{
		zap.String("log_type", "user_action"),
		zap.String("user_id", action.UserID),
		zap.String("action", action.Action),
		zap.Time("timestamp", action.Timestamp),
	}

	if action.Resource != "" {
		fields = append(fields, zap.String("resource", action.Resource))
	}
	if action.ResourceID != "" {
		fields = append(fields, zap.String("resource_id", action.ResourceID))
	}
	if action.IPAddress != "" {
		fields = append(fields, zap.String("ip_address", action.IPAddress))
	}
	if action.UserAgent != "" {
		fields = append(fields, zap.String("user_agent", action.UserAgent))
	}
	if action.Details != nil {
		fields = append(fields, zap.Any("details", action.Details))
	}

	sl.logger.Info("User action recorded", fields...)
}

// LogDatabaseOperation 记录数据库操作日志
func (sl *StructuredLog) LogDatabaseOperation(op DatabaseOperation) {
	op.Timestamp = time.Now()

	fields := []zap.Field{
		zap.String("log_type", "database_operation"),
		zap.String("operation", op.Operation),
		zap.String("table", op.Table),
		zap.Duration("duration", op.Duration),
		zap.Time("timestamp", op.Timestamp),
	}

	if op.RowsAffected > 0 {
		fields = append(fields, zap.Int64("rows_affected", op.RowsAffected))
	}
	if op.SQL != "" {
		fields = append(fields, zap.String("sql", op.SQL))
	}

	if op.Error != "" {
		fields = append(fields, zap.String("error", op.Error))
		sl.logger.Error("Database operation failed", fields...)
	} else {
		sl.logger.Info("Database operation completed", fields...)
	}
}

// LogFileOperation 记录文件操作日志
func (sl *StructuredLog) LogFileOperation(op FileOperation) {
	op.Timestamp = time.Now()

	fields := []zap.Field{
		zap.String("log_type", "file_operation"),
		zap.String("user_id", op.UserID),
		zap.String("operation", op.Operation),
		zap.String("file_name", op.FileName),
		zap.Time("timestamp", op.Timestamp),
	}

	if op.FileSize > 0 {
		fields = append(fields, zap.Int64("file_size", op.FileSize))
	}
	if op.FileType != "" {
		fields = append(fields, zap.String("file_type", op.FileType))
	}
	if op.StoragePath != "" {
		fields = append(fields, zap.String("storage_path", op.StoragePath))
	}
	if op.Duration > 0 {
		fields = append(fields, zap.Duration("duration", op.Duration))
	}

	if op.Error != "" {
		fields = append(fields, zap.String("error", op.Error))
		sl.logger.Error("File operation failed", fields...)
	} else {
		sl.logger.Info("File operation completed", fields...)
	}
}

// LogSecurityEvent 记录安全事件日志
func (sl *StructuredLog) LogSecurityEvent(event SecurityEvent) {
	event.Timestamp = time.Now()

	fields := []zap.Field{
		zap.String("log_type", "security_event"),
		zap.String("event_type", event.EventType),
		zap.String("ip_address", event.IPAddress),
		zap.String("severity", event.Severity),
		zap.String("description", event.Description),
		zap.Time("timestamp", event.Timestamp),
	}

	if event.UserID != "" {
		fields = append(fields, zap.String("user_id", event.UserID))
	}
	if event.UserAgent != "" {
		fields = append(fields, zap.String("user_agent", event.UserAgent))
	}
	if event.Resource != "" {
		fields = append(fields, zap.String("resource", event.Resource))
	}
	if event.Details != nil {
		fields = append(fields, zap.Any("details", event.Details))
	}

	// 根据严重程度选择日志级别
	switch event.Severity {
	case "critical":
		sl.logger.Error("Critical security event", fields...)
	case "high":
		sl.logger.Error("High severity security event", fields...)
	case "medium":
		sl.logger.Warn("Medium severity security event", fields...)
	default:
		sl.logger.Info("Security event", fields...)
	}
}

// LogSystemEvent 记录系统事件日志
func (sl *StructuredLog) LogSystemEvent(event SystemEvent) {
	event.Timestamp = time.Now()

	fields := []zap.Field{
		zap.String("log_type", "system_event"),
		zap.String("component", event.Component),
		zap.String("event", event.Event),
		zap.String("level", event.Level),
		zap.String("message", event.Message),
		zap.Time("timestamp", event.Timestamp),
	}

	if event.Details != nil {
		fields = append(fields, zap.Any("details", event.Details))
	}

	// 根据级别选择日志级别
	switch event.Level {
	case "error":
		sl.logger.Error("System event", fields...)
	case "warn":
		sl.logger.Warn("System event", fields...)
	default:
		sl.logger.Info("System event", fields...)
	}
}

// 全局结构化日志实例
var StructuredLogger *StructuredLog

// InitStructuredLogger 初始化结构化日志
func InitStructuredLogger() {
	if Logger != nil {
		StructuredLogger = NewStructuredLog(Logger)
	}
}

// 便捷方法
func LogUserAction(action UserAction) {
	if StructuredLogger != nil {
		StructuredLogger.LogUserAction(action)
	}
}

func LogDatabaseOperation(op DatabaseOperation) {
	if StructuredLogger != nil {
		StructuredLogger.LogDatabaseOperation(op)
	}
}

func LogFileOperation(op FileOperation) {
	if StructuredLogger != nil {
		StructuredLogger.LogFileOperation(op)
	}
}

func LogSecurityEvent(event SecurityEvent) {
	if StructuredLogger != nil {
		StructuredLogger.LogSecurityEvent(event)
	}
}

func LogSystemEvent(event SystemEvent) {
	if StructuredLogger != nil {
		StructuredLogger.LogSystemEvent(event)
	}
}
