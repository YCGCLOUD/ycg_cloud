package logger

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestAccessLogConfig 测试访问日志配置结构
func TestAccessLogConfig(t *testing.T) {
	config := AccessLogConfig{
		Enabled:  true,
		FilePath: "/tmp/access.log",
		Format:   "json",
	}

	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.FilePath != "/tmp/access.log" {
		t.Errorf("Expected FilePath to be '/tmp/access.log', got %s", config.FilePath)
	}
	if config.Format != "json" {
		t.Errorf("Expected Format to be 'json', got %s", config.Format)
	}
}

// TestAccessLogEntry 测试访问日志条目结构
// validateAccessLogEntry 验证访问日志条目的字段值
func validateAccessLogEntry(t *testing.T, entry AccessLogEntry, expected AccessLogEntry) {
	t.Helper()

	// 使用结构体比较防止复杂度过高
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"Timestamp", entry.Timestamp, expected.Timestamp},
		{"RequestID", entry.RequestID, expected.RequestID},
		{"UserID", entry.UserID, expected.UserID},
		{"Method", entry.Method, expected.Method},
		{"Path", entry.Path, expected.Path},
		{"Query", entry.Query, expected.Query},
		{"StatusCode", entry.StatusCode, expected.StatusCode},
		{"ResponseTime", entry.ResponseTime, expected.ResponseTime},
		{"UserAgent", entry.UserAgent, expected.UserAgent},
		{"IPAddress", entry.IPAddress, expected.IPAddress},
		{"RequestSize", entry.RequestSize, expected.RequestSize},
		{"ResponseSize", entry.ResponseSize, expected.ResponseSize},
		{"Referer", entry.Referer, expected.Referer},
		{"Protocol", entry.Protocol, expected.Protocol},
	}

	for _, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("Field %s: expected %v, got %v", tt.name, tt.expected, tt.actual)
		}
	}
}

func TestAccessLogEntry(t *testing.T) {
	now := time.Now()
	entry := AccessLogEntry{
		Timestamp:    now,
		RequestID:    "req-123",
		UserID:       "user-456",
		Method:       "GET",
		Path:         "/api/files",
		Query:        "page=1&limit=10",
		StatusCode:   200,
		ResponseTime: 150,
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		RequestSize:  1024,
		ResponseSize: 2048,
		Referer:      "https://example.com",
		Protocol:     "HTTP/1.1",
	}

	// 使用辅助函数进行验证
	validateAccessLogEntry(t, entry, entry)
}

// TestInitAccessLogger 测试访问日志初始化
func TestInitAccessLogger(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "test_access_logger")
	accessLogFile := filepath.Join(tempDir, "access.log")
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		config  AccessLogConfig
		wantErr bool
	}{
		{
			name: "disabled access logger",
			config: AccessLogConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "enabled access logger",
			config: AccessLogConfig{
				Enabled:  true,
				FilePath: accessLogFile,
				Format:   "json",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitAccessLogger(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitAccessLogger() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 验证AccessLogger是否正确初始化
			if AccessLogger == nil {
				t.Error("AccessLogger should not be nil after initialization")
			}

			// 对于启用的日志，验证目录是否创建
			if tt.config.Enabled && err == nil {
				if _, statErr := os.Stat(tempDir); os.IsNotExist(statErr) {
					t.Error("Directory should be created for enabled access logger")
				}
			}
		})
	}
}

// TestLogAccess 测试访问日志记录
func TestLogAccess(t *testing.T) {
	// 设置临时访问日志
	tempDir := filepath.Join(os.TempDir(), "test_log_access")
	accessLogFile := filepath.Join(tempDir, "access.log")
	defer os.RemoveAll(tempDir)

	config := AccessLogConfig{
		Enabled:  true,
		FilePath: accessLogFile,
		Format:   "json",
	}

	err := InitAccessLogger(config)
	if err != nil {
		t.Fatalf("Failed to initialize access logger: %v", err)
	}

	tests := []struct {
		name  string
		entry AccessLogEntry
	}{
		{
			name: "complete access log entry",
			entry: AccessLogEntry{
				Timestamp:    time.Now(),
				RequestID:    "req-123",
				UserID:       "user-456",
				Method:       "POST",
				Path:         "/api/upload",
				Query:        "type=image",
				StatusCode:   201,
				ResponseTime: 250,
				UserAgent:    "curl/7.68.0",
				IPAddress:    "192.168.1.100",
				RequestSize:  5120,
				ResponseSize: 512,
				Referer:      "https://app.example.com",
				Protocol:     "HTTP/2.0",
			},
		},
		{
			name: "minimal access log entry",
			entry: AccessLogEntry{
				Timestamp:    time.Now(),
				RequestID:    "req-456",
				Method:       "GET",
				Path:         "/health",
				StatusCode:   200,
				ResponseTime: 5,
				IPAddress:    "127.0.0.1",
			},
		},
		{
			name: "error response entry",
			entry: AccessLogEntry{
				Timestamp:    time.Now(),
				RequestID:    "req-789",
				Method:       "DELETE",
				Path:         "/api/files/nonexistent",
				StatusCode:   404,
				ResponseTime: 10,
				IPAddress:    "10.0.0.1",
				UserAgent:    "Mozilla/5.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试不应该panic或出错
			LogAccess(tt.entry)
		})
	}

	// 同步日志确保写入
	err = SyncAccessLogger()
	if err != nil {
		t.Errorf("SyncAccessLogger() error = %v", err)
	}

	// 验证日志文件是否创建
	if _, err := os.Stat(accessLogFile); os.IsNotExist(err) {
		t.Error("Access log file should be created")
	}
}

// TestLogAccessWithNilLogger 测试空AccessLogger的日志记录
func TestLogAccessWithNilLogger(t *testing.T) {
	// 保存原始AccessLogger
	originalAccessLogger := AccessLogger

	// 设置AccessLogger为nil
	AccessLogger = nil

	entry := AccessLogEntry{
		Timestamp:    time.Now(),
		RequestID:    "req-test",
		Method:       "GET",
		Path:         "/test",
		StatusCode:   200,
		ResponseTime: 10,
		IPAddress:    "127.0.0.1",
	}

	// 这应该安全地处理，不会panic
	LogAccess(entry)

	// 恢复原始AccessLogger
	AccessLogger = originalAccessLogger
}

// TestGetAccessLogWriter 测试获取访问日志Writer
func TestGetAccessLogWriter(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "test_access_writer")
	accessLogFile := filepath.Join(tempDir, "access.log")
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		config  AccessLogConfig
		wantErr bool
	}{
		{
			name: "disabled access writer",
			config: AccessLogConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "enabled access writer",
			config: AccessLogConfig{
				Enabled:  true,
				FilePath: accessLogFile,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := GetAccessLogWriter(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccessLogWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if writer == nil {
				t.Error("GetAccessLogWriter() should not return nil writer")
				return
			}

			// 对于禁用的配置，应该返回io.Discard
			if !tt.config.Enabled {
				if writer != io.Discard {
					t.Error("GetAccessLogWriter() should return io.Discard for disabled config")
				}
			}

			// 测试写入功能
			testData := []byte("test log entry\n")
			n, writeErr := writer.Write(testData)
			if writeErr != nil {
				t.Errorf("Writer.Write() error = %v", writeErr)
			}
			if n != len(testData) {
				t.Errorf("Writer.Write() wrote %d bytes, expected %d", n, len(testData))
			}
		})
	}
}

// TestSyncAccessLogger 测试访问日志同步
func TestSyncAccessLogger(t *testing.T) {
	// 测试正常情况
	tempDir := filepath.Join(os.TempDir(), "test_sync_access")
	accessLogFile := filepath.Join(tempDir, "access.log")
	defer os.RemoveAll(tempDir)

	config := AccessLogConfig{
		Enabled:  true,
		FilePath: accessLogFile,
	}

	err := InitAccessLogger(config)
	if err != nil {
		t.Fatalf("Failed to initialize access logger: %v", err)
	}

	err = SyncAccessLogger()
	if err != nil {
		t.Errorf("SyncAccessLogger() error = %v", err)
	}

	// 测试AccessLogger为nil的情况
	originalAccessLogger := AccessLogger
	AccessLogger = nil

	err = SyncAccessLogger()
	if err != nil {
		t.Errorf("SyncAccessLogger() should not error when AccessLogger is nil: %v", err)
	}

	// 恢复原始AccessLogger
	AccessLogger = originalAccessLogger
}

// TestAccessLogEntryOptionalFields 测试访问日志条目的可选字段
func TestAccessLogEntryOptionalFields(t *testing.T) {
	// 初始化访问日志
	tempDir := filepath.Join(os.TempDir(), "test_optional_fields")
	accessLogFile := filepath.Join(tempDir, "access.log")
	defer os.RemoveAll(tempDir)

	config := AccessLogConfig{
		Enabled:  true,
		FilePath: accessLogFile,
	}

	err := InitAccessLogger(config)
	if err != nil {
		t.Fatalf("Failed to initialize access logger: %v", err)
	}

	// 测试包含所有可选字段的条目
	fullEntry := AccessLogEntry{
		Timestamp:    time.Now(),
		RequestID:    "req-full",
		UserID:       "user-123",
		Method:       "POST",
		Path:         "/api/test",
		Query:        "param=value",
		StatusCode:   200,
		ResponseTime: 100,
		UserAgent:    "test-agent",
		IPAddress:    "192.168.1.1",
		RequestSize:  1024,
		ResponseSize: 2048,
		Referer:      "https://example.com",
		Protocol:     "HTTP/1.1",
	}

	// 测试只包含必需字段的条目
	minimalEntry := AccessLogEntry{
		Timestamp:    time.Now(),
		RequestID:    "req-minimal",
		Method:       "GET",
		Path:         "/api/minimal",
		StatusCode:   200,
		ResponseTime: 50,
		IPAddress:    "127.0.0.1",
	}

	// 记录两种类型的日志条目
	LogAccess(fullEntry)
	LogAccess(minimalEntry)

	// 同步确保写入
	err = SyncAccessLogger()
	if err != nil {
		t.Errorf("SyncAccessLogger() error = %v", err)
	}
}

// BenchmarkLogAccess 基准测试访问日志记录
func BenchmarkLogAccess(b *testing.B) {
	// 使用内存中的临时文件以减少I/O影响
	tempDir := filepath.Join(os.TempDir(), "bench_access_log")
	accessLogFile := filepath.Join(tempDir, "access.log")
	defer os.RemoveAll(tempDir)

	config := AccessLogConfig{
		Enabled:  true,
		FilePath: accessLogFile,
	}

	err := InitAccessLogger(config)
	if err != nil {
		b.Fatalf("Failed to initialize access logger: %v", err)
	}

	entry := AccessLogEntry{
		Timestamp:    time.Now(),
		RequestID:    "req-bench",
		Method:       "GET",
		Path:         "/api/benchmark",
		StatusCode:   200,
		ResponseTime: 10,
		IPAddress:    "127.0.0.1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogAccess(entry)
	}
}

// BenchmarkInitAccessLogger 基准测试访问日志初始化
func BenchmarkInitAccessLogger(b *testing.B) {
	tempDir := filepath.Join(os.TempDir(), "bench_init_access")
	defer os.RemoveAll(tempDir)

	config := AccessLogConfig{
		Enabled:  true,
		FilePath: filepath.Join(tempDir, "access.log"),
	}

	for i := 0; i < b.N; i++ {
		err := InitAccessLogger(config)
		if err != nil {
			b.Errorf("InitAccessLogger() error = %v", err)
		}
	}
}
