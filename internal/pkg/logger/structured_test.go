package logger

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TestNewStructuredLog 测试结构化日志创建
func TestNewStructuredLog(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sl := NewStructuredLog(logger)

	// NewStructuredLog 总是返回有效的指针，无需nil检查
	if sl.logger != logger {
		t.Error("StructuredLog should contain the provided logger")
	}
}

// TestLogUserAction 测试用户操作日志记录
func TestLogUserAction(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sl := NewStructuredLog(logger)

	tests := []struct {
		name   string
		action UserAction
	}{
		{
			name: "basic user action",
			action: UserAction{
				UserID: "user123",
				Action: "login",
			},
		},
		{
			name: "complete user action",
			action: UserAction{
				UserID:     "user123",
				Action:     "upload_file",
				Resource:   "file",
				ResourceID: "file456",
				IPAddress:  "192.168.1.1",
				UserAgent:  "Mozilla/5.0",
				Details: map[string]interface{}{
					"file_size": 1024,
					"file_type": "image/png",
				},
			},
		},
		{
			name: "minimal user action",
			action: UserAction{
				UserID: "user456",
				Action: "logout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试不应该panic或出错
			sl.LogUserAction(tt.action)
		})
	}
}

// TestLogDatabaseOperation 测试数据库操作日志记录
func TestLogDatabaseOperation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sl := NewStructuredLog(logger)

	tests := []struct {
		name string
		op   DatabaseOperation
	}{
		{
			name: "successful database operation",
			op: DatabaseOperation{
				Operation:    "SELECT",
				Table:        "users",
				Duration:     time.Millisecond * 50,
				RowsAffected: 10,
				SQL:          "SELECT * FROM users WHERE active = 1",
			},
		},
		{
			name: "failed database operation",
			op: DatabaseOperation{
				Operation: "INSERT",
				Table:     "users",
				Duration:  time.Millisecond * 100,
				Error:     "duplicate key error",
				SQL:       "INSERT INTO users (email) VALUES ('test@example.com')",
			},
		},
		{
			name: "minimal database operation",
			op: DatabaseOperation{
				Operation: "DELETE",
				Table:     "sessions",
				Duration:  time.Millisecond * 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试不应该panic或出错
			sl.LogDatabaseOperation(tt.op)
		})
	}
}

// TestLogFileOperation 测试文件操作日志记录
func TestLogFileOperation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sl := NewStructuredLog(logger)

	tests := []struct {
		name string
		op   FileOperation
	}{
		{
			name: "successful file upload",
			op: FileOperation{
				UserID:      "user123",
				Operation:   "upload",
				FileName:    "document.pdf",
				FileSize:    2048,
				FileType:    "application/pdf",
				StoragePath: "/storage/user123/document.pdf",
				Duration:    time.Second * 2,
			},
		},
		{
			name: "failed file download",
			op: FileOperation{
				UserID:    "user456",
				Operation: "download",
				FileName:  "missing.jpg",
				Error:     "file not found",
			},
		},
		{
			name: "file deletion",
			op: FileOperation{
				UserID:    "user789",
				Operation: "delete",
				FileName:  "old_file.txt",
				Duration:  time.Millisecond * 50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试不应该panic或出错
			sl.LogFileOperation(tt.op)
		})
	}
}

// TestLogSecurityEvent 测试安全事件日志记录
func TestLogSecurityEvent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sl := NewStructuredLog(logger)

	tests := []struct {
		name  string
		event SecurityEvent
	}{
		{
			name: "critical security event",
			event: SecurityEvent{
				EventType:   "login_failed",
				UserID:      "user123",
				IPAddress:   "192.168.1.100",
				UserAgent:   "curl/7.68.0",
				Resource:    "/api/login",
				Severity:    "critical",
				Description: "Multiple failed login attempts detected",
				Details: map[string]interface{}{
					"attempts": 5,
					"window":   "5 minutes",
				},
			},
		},
		{
			name: "high severity event",
			event: SecurityEvent{
				EventType:   "permission_denied",
				UserID:      "user456",
				IPAddress:   "10.0.0.50",
				Severity:    "high",
				Description: "Access denied to admin resource",
			},
		},
		{
			name: "medium severity event",
			event: SecurityEvent{
				EventType:   "unusual_activity",
				IPAddress:   "203.0.113.1",
				Severity:    "medium",
				Description: "User accessing from unusual location",
			},
		},
		{
			name: "low severity event",
			event: SecurityEvent{
				EventType:   "password_change",
				UserID:      "user789",
				IPAddress:   "192.168.1.10",
				Severity:    "low",
				Description: "User changed password",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试不应该panic或出错
			sl.LogSecurityEvent(tt.event)
		})
	}
}

// TestLogSystemEvent 测试系统事件日志记录
func TestLogSystemEvent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sl := NewStructuredLog(logger)

	tests := []struct {
		name  string
		event SystemEvent
	}{
		{
			name: "error level system event",
			event: SystemEvent{
				Component: "database",
				Event:     "connection_failed",
				Level:     "error",
				Message:   "Failed to connect to database",
				Details: map[string]interface{}{
					"host":     "db.example.com",
					"port":     5432,
					"database": "cloudpan",
				},
			},
		},
		{
			name: "warn level system event",
			event: SystemEvent{
				Component: "storage",
				Event:     "disk_space_low",
				Level:     "warn",
				Message:   "Disk space is running low",
				Details: map[string]interface{}{
					"available": "500MB",
					"threshold": "1GB",
				},
			},
		},
		{
			name: "info level system event",
			event: SystemEvent{
				Component: "server",
				Event:     "startup",
				Level:     "info",
				Message:   "Server started successfully",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试不应该panic或出错
			sl.LogSystemEvent(tt.event)
		})
	}
}

// TestInitStructuredLogger 测试结构化日志初始化
func TestInitStructuredLogger(t *testing.T) {
	// 保存原始状态
	originalLogger := Logger
	originalStructuredLogger := StructuredLogger

	// 测试Logger为nil时的初始化
	Logger = nil
	InitStructuredLogger()
	if StructuredLogger != nil {
		t.Error("StructuredLogger should be nil when Logger is nil")
	}

	// 测试Logger存在时的初始化
	Logger = zaptest.NewLogger(t)
	InitStructuredLogger()
	if StructuredLogger == nil {
		t.Error("StructuredLogger should not be nil when Logger is set")
	}
	if StructuredLogger.logger != Logger {
		t.Error("StructuredLogger should use the global Logger")
	}

	// 恢复原始状态
	Logger = originalLogger
	StructuredLogger = originalStructuredLogger
}

// TestGlobalConvenienceFunctions 测试全局便捷函数
func TestGlobalConvenienceFunctions(t *testing.T) {
	// 保存原始状态
	originalStructuredLogger := StructuredLogger

	// 测试StructuredLogger为nil时的函数调用
	StructuredLogger = nil

	// 这些调用应该安全，不会panic
	LogUserAction(UserAction{UserID: "test", Action: "test"})
	LogDatabaseOperation(DatabaseOperation{Operation: "test", Table: "test"})
	LogFileOperation(FileOperation{UserID: "test", Operation: "test", FileName: "test"})
	LogSecurityEvent(SecurityEvent{EventType: "test", IPAddress: "test", Severity: "low", Description: "test"})
	LogSystemEvent(SystemEvent{Component: "test", Event: "test", Level: "info", Message: "test"})

	// 设置StructuredLogger并测试
	logger := zaptest.NewLogger(t)
	StructuredLogger = NewStructuredLog(logger)

	LogUserAction(UserAction{UserID: "test", Action: "test"})
	LogDatabaseOperation(DatabaseOperation{Operation: "test", Table: "test"})
	LogFileOperation(FileOperation{UserID: "test", Operation: "test", FileName: "test"})
	LogSecurityEvent(SecurityEvent{EventType: "test", IPAddress: "test", Severity: "low", Description: "test"})
	LogSystemEvent(SystemEvent{Component: "test", Event: "test", Level: "info", Message: "test"})

	// 恢复原始状态
	StructuredLogger = originalStructuredLogger
}

// TestUserActionStruct 测试UserAction结构体
func TestUserActionStruct(t *testing.T) {
	action := UserAction{
		UserID:     "user123",
		Action:     "login",
		Resource:   "auth",
		ResourceID: "session456",
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		Details: map[string]interface{}{
			"method": "oauth",
		},
		Timestamp: time.Now(),
	}

	// 验证字段值
	if action.UserID != "user123" {
		t.Errorf("Expected UserID to be 'user123', got %s", action.UserID)
	}
	if action.Action != "login" {
		t.Errorf("Expected Action to be 'login', got %s", action.Action)
	}
	if action.Resource != "auth" {
		t.Errorf("Expected Resource to be 'auth', got %s", action.Resource)
	}
	if action.ResourceID != "session456" {
		t.Errorf("Expected ResourceID to be 'session456', got %s", action.ResourceID)
	}
	if action.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IPAddress to be '192.168.1.1', got %s", action.IPAddress)
	}
	if action.UserAgent != "Mozilla/5.0" {
		t.Errorf("Expected UserAgent to be 'Mozilla/5.0', got %s", action.UserAgent)
	}
	if action.Details["method"] != "oauth" {
		t.Errorf("Expected Details['method'] to be 'oauth', got %v", action.Details["method"])
	}
	if action.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}

// TestDatabaseOperationStruct 测试DatabaseOperation结构体
func TestDatabaseOperationStruct(t *testing.T) {
	op := DatabaseOperation{
		Operation:    "SELECT",
		Table:        "users",
		Duration:     time.Millisecond * 100,
		RowsAffected: 10,
		Error:        "",
		SQL:          "SELECT * FROM users",
		Timestamp:    time.Now(),
	}

	if op.Operation != "SELECT" {
		t.Errorf("Expected Operation to be 'SELECT', got %s", op.Operation)
	}
	if op.Table != "users" {
		t.Errorf("Expected Table to be 'users', got %s", op.Table)
	}
	if op.Duration != time.Millisecond*100 {
		t.Errorf("Expected Duration to be 100ms, got %v", op.Duration)
	}
	if op.RowsAffected != 10 {
		t.Errorf("Expected RowsAffected to be 10, got %d", op.RowsAffected)
	}
	if op.Error != "" {
		t.Errorf("Expected Error to be empty, got %s", op.Error)
	}
	if op.SQL != "SELECT * FROM users" {
		t.Errorf("Expected SQL to be 'SELECT * FROM users', got %s", op.SQL)
	}
	if op.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}

// TestFileOperationStruct 测试FileOperation结构体
func TestFileOperationStruct(t *testing.T) {
	op := FileOperation{
		UserID:      "user123",
		Operation:   "upload",
		FileName:    "test.jpg",
		FileSize:    1024,
		FileType:    "image/jpeg",
		StoragePath: "/storage/test.jpg",
		Duration:    time.Second,
		Error:       "",
		Timestamp:   time.Now(),
	}

	if op.UserID != "user123" {
		t.Errorf("Expected UserID to be 'user123', got %s", op.UserID)
	}
	if op.Operation != "upload" {
		t.Errorf("Expected Operation to be 'upload', got %s", op.Operation)
	}
	if op.FileName != "test.jpg" {
		t.Errorf("Expected FileName to be 'test.jpg', got %s", op.FileName)
	}
	if op.FileSize != 1024 {
		t.Errorf("Expected FileSize to be 1024, got %d", op.FileSize)
	}
	if op.FileType != "image/jpeg" {
		t.Errorf("Expected FileType to be 'image/jpeg', got %s", op.FileType)
	}
	if op.StoragePath != "/storage/test.jpg" {
		t.Errorf("Expected StoragePath to be '/storage/test.jpg', got %s", op.StoragePath)
	}
	if op.Duration != time.Second {
		t.Errorf("Expected Duration to be 1s, got %v", op.Duration)
	}
	if op.Error != "" {
		t.Errorf("Expected Error to be empty, got %s", op.Error)
	}
	if op.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}

// TestSecurityEventStruct 测试SecurityEvent结构体
func TestSecurityEventStruct(t *testing.T) {
	event := SecurityEvent{
		EventType:   "login_failed",
		UserID:      "user123",
		IPAddress:   "192.168.1.1",
		UserAgent:   "curl/7.68.0",
		Resource:    "/api/login",
		Severity:    "high",
		Description: "Failed login attempt",
		Details: map[string]interface{}{
			"attempts": 3,
		},
		Timestamp: time.Now(),
	}

	if event.EventType != "login_failed" {
		t.Errorf("Expected EventType to be 'login_failed', got %s", event.EventType)
	}
	if event.UserID != "user123" {
		t.Errorf("Expected UserID to be 'user123', got %s", event.UserID)
	}
	if event.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IPAddress to be '192.168.1.1', got %s", event.IPAddress)
	}
	if event.UserAgent != "curl/7.68.0" {
		t.Errorf("Expected UserAgent to be 'curl/7.68.0', got %s", event.UserAgent)
	}
	if event.Resource != "/api/login" {
		t.Errorf("Expected Resource to be '/api/login', got %s", event.Resource)
	}
	if event.Severity != "high" {
		t.Errorf("Expected Severity to be 'high', got %s", event.Severity)
	}
	if event.Description != "Failed login attempt" {
		t.Errorf("Expected Description to be 'Failed login attempt', got %s", event.Description)
	}
	if event.Details["attempts"] != 3 {
		t.Errorf("Expected Details['attempts'] to be 3, got %v", event.Details["attempts"])
	}
	if event.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}

// TestSystemEventStruct 测试SystemEvent结构体
func TestSystemEventStruct(t *testing.T) {
	event := SystemEvent{
		Component: "database",
		Event:     "connection_lost",
		Level:     "error",
		Message:   "Database connection lost",
		Details: map[string]interface{}{
			"host": "db.example.com",
		},
		Timestamp: time.Now(),
	}

	if event.Component != "database" {
		t.Errorf("Expected Component to be 'database', got %s", event.Component)
	}
	if event.Event != "connection_lost" {
		t.Errorf("Expected Event to be 'connection_lost', got %s", event.Event)
	}
	if event.Level != "error" {
		t.Errorf("Expected Level to be 'error', got %s", event.Level)
	}
	if event.Message != "Database connection lost" {
		t.Errorf("Expected Message to be 'Database connection lost', got %s", event.Message)
	}
	if event.Details["host"] != "db.example.com" {
		t.Errorf("Expected Details['host'] to be 'db.example.com', got %v", event.Details["host"])
	}
	if event.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}

// BenchmarkLogUserAction 基准测试用户操作日志
func BenchmarkLogUserAction(b *testing.B) {
	logger := zap.NewNop() // 使用no-op logger提高性能
	sl := NewStructuredLog(logger)

	action := UserAction{
		UserID:    "user123",
		Action:    "test_action",
		IPAddress: "192.168.1.1",
	}

	for i := 0; i < b.N; i++ {
		sl.LogUserAction(action)
	}
}

// BenchmarkLogDatabaseOperation 基准测试数据库操作日志
func BenchmarkLogDatabaseOperation(b *testing.B) {
	logger := zap.NewNop()
	sl := NewStructuredLog(logger)

	op := DatabaseOperation{
		Operation: "SELECT",
		Table:     "users",
		Duration:  time.Millisecond * 50,
	}

	for i := 0; i < b.N; i++ {
		sl.LogDatabaseOperation(op)
	}
}
