package logger

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

// TestLogConfig 测试日志配置结构
func TestLogConfig(t *testing.T) {
	config := LogConfig{
		Level:      "info",
		Format:     "json",
		Output:     "console",
		FilePath:   "/tmp/test.log",
		MaxSize:    100,
		MaxAge:     7,
		MaxBackups: 3,
		Compress:   true,
	}

	// 验证字段值
	if config.Level != "info" {
		t.Errorf("Expected Level to be 'info', got %s", config.Level)
	}
	if config.Format != "json" {
		t.Errorf("Expected Format to be 'json', got %s", config.Format)
	}
	if config.Output != "console" {
		t.Errorf("Expected Output to be 'console', got %s", config.Output)
	}
	if config.FilePath != "/tmp/test.log" {
		t.Errorf("Expected FilePath to be '/tmp/test.log', got %s", config.FilePath)
	}
	if config.MaxSize != 100 {
		t.Errorf("Expected MaxSize to be 100, got %d", config.MaxSize)
	}
	if config.MaxAge != 7 {
		t.Errorf("Expected MaxAge to be 7, got %d", config.MaxAge)
	}
	if config.MaxBackups != 3 {
		t.Errorf("Expected MaxBackups to be 3, got %d", config.MaxBackups)
	}
	if !config.Compress {
		t.Error("Expected Compress to be true")
	}
}

// TestGetLogLevel 测试日志级别转换
func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zapcore.Level
		hasError bool
	}{
		{"debug", zapcore.DebugLevel, false},
		{"info", zapcore.InfoLevel, false},
		{"warn", zapcore.WarnLevel, false},
		{"warning", zapcore.WarnLevel, false},
		{"error", zapcore.ErrorLevel, false},
		{"panic", zapcore.PanicLevel, false},
		{"fatal", zapcore.FatalLevel, false},
		{"DEBUG", zapcore.DebugLevel, false}, // 测试大小写不敏感
		{"Info", zapcore.InfoLevel, false},
		{"invalid", zapcore.InfoLevel, true}, // 无效级别应返回默认级别和错误
		{"", zapcore.InfoLevel, true},        // 空字符串应返回默认级别和错误
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := getLogLevel(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input %s, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", tt.input, err)
				}
			}

			if level != tt.expected {
				t.Errorf("Expected level %v for input %s, got %v", tt.expected, tt.input, level)
			}
		})
	}
}

// TestCustomTimeEncoder 测试自定义时间编码器
func TestCustomTimeEncoder(t *testing.T) {
	// 创建一个测试时间
	testTime := time.Date(2024, 1, 1, 12, 30, 45, 123456789, time.UTC)

	// 创建一个简单的缓冲编码器来测试
	var result string
	encoder := &testEncoder{result: &result}

	customTimeEncoder(testTime, encoder)

	expected := "2024-01-01 12:30:45.123"
	if result != expected {
		t.Errorf("Expected time format %s, got %s", expected, result)
	}
}

// testEncoder 用于测试的简单编码器
type testEncoder struct {
	result *string
}

func (e *testEncoder) AppendString(s string) {
	*e.result = s
}

func (e *testEncoder) AppendBool(bool)              {}
func (e *testEncoder) AppendByteString([]byte)      {}
func (e *testEncoder) AppendComplex128(complex128)  {}
func (e *testEncoder) AppendComplex64(complex64)    {}
func (e *testEncoder) AppendFloat64(float64)        {}
func (e *testEncoder) AppendFloat32(float32)        {}
func (e *testEncoder) AppendInt(int)                {}
func (e *testEncoder) AppendInt64(int64)            {}
func (e *testEncoder) AppendInt32(int32)            {}
func (e *testEncoder) AppendInt16(int16)            {}
func (e *testEncoder) AppendInt8(int8)              {}
func (e *testEncoder) AppendUint(uint)              {}
func (e *testEncoder) AppendUint64(uint64)          {}
func (e *testEncoder) AppendUint32(uint32)          {}
func (e *testEncoder) AppendUint16(uint16)          {}
func (e *testEncoder) AppendUint8(uint8)            {}
func (e *testEncoder) AppendUintptr(uintptr)        {}
func (e *testEncoder) AppendDuration(time.Duration) {}
func (e *testEncoder) AppendTime(time.Time)         {}

// TestCreateEncoder 测试编码器创建
func TestCreateEncoder(t *testing.T) {
	tests := []struct {
		name   string
		config LogConfig
	}{
		{
			name: "json encoder",
			config: LogConfig{
				Format: "json",
			},
		},
		{
			name: "console encoder",
			config: LogConfig{
				Format: "console",
			},
		},
		{
			name: "default to console",
			config: LogConfig{
				Format: "unknown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := createEncoder(tt.config)
			if encoder == nil {
				t.Error("createEncoder should not return nil")
			}
		})
	}
}

// TestCreateFileWriter 测试文件Writer创建
func TestCreateFileWriter(t *testing.T) {
	config := LogConfig{
		FilePath:   "/tmp/test.log",
		MaxSize:    100,
		MaxAge:     7,
		MaxBackups: 3,
		Compress:   true,
	}

	writer := createFileWriter(config)
	if writer == nil {
		t.Error("createFileWriter should not return nil")
	}

	if writer.Filename != config.FilePath {
		t.Errorf("Expected filename %s, got %s", config.FilePath, writer.Filename)
	}
	if writer.MaxSize != config.MaxSize {
		t.Errorf("Expected max size %d, got %d", config.MaxSize, writer.MaxSize)
	}
	if writer.MaxAge != config.MaxAge {
		t.Errorf("Expected max age %d, got %d", config.MaxAge, writer.MaxAge)
	}
	if writer.MaxBackups != config.MaxBackups {
		t.Errorf("Expected max backups %d, got %d", config.MaxBackups, writer.MaxBackups)
	}
	if writer.Compress != config.Compress {
		t.Errorf("Expected compress %v, got %v", config.Compress, writer.Compress)
	}
}

// TestEnsureLogDir 测试日志目录创建
func TestEnsureLogDir(t *testing.T) {
	// 创建临时目录进行测试
	tempDir := filepath.Join(os.TempDir(), "test_logs", "sub", "dir")
	testFile := filepath.Join(tempDir, "test.log")

	// 清理：确保目录不存在
	os.RemoveAll(filepath.Join(os.TempDir(), "test_logs"))

	// 测试目录创建
	err := ensureLogDir(testFile)
	if err != nil {
		t.Fatalf("ensureLogDir failed: %v", err)
	}

	// 验证目录是否创建
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Directory should be created")
	}

	// 清理
	os.RemoveAll(filepath.Join(os.TempDir(), "test_logs"))
}

// TestWithRequestID 测试请求ID日志
func TestWithRequestID(t *testing.T) {
	// 初始化一个测试logger
	setupTestLogger(t)

	ctx := context.Background()
	requestID := "test-request-123"

	logger := WithRequestID(ctx, requestID)
	if logger == nil {
		t.Error("WithRequestID should not return nil")
	}

	// 这里我们无法直接验证requestID是否被添加到日志中，
	// 但我们可以验证返回的logger不是nil且不等于原始logger
	if logger == Logger {
		t.Error("WithRequestID should return a new logger instance")
	}
}

// TestWithUserID 测试用户ID日志
func TestWithUserID(t *testing.T) {
	// 初始化一个测试logger
	setupTestLogger(t)

	// 测试不带请求ID的情况
	ctx := context.Background()
	userID := "user-123"

	logger := WithUserID(ctx, userID)
	if logger == nil {
		t.Error("WithUserID should not return nil")
	}

	// 测试带请求ID的情况
	ctxWithRequestID := context.WithValue(ctx, RequestIDKey, "request-123")
	loggerWithRequestID := WithUserID(ctxWithRequestID, userID)
	if loggerWithRequestID == nil {
		t.Error("WithUserID should not return nil when context has request ID")
	}
}

// TestRequestIDKey 测试请求ID键常量
func TestRequestIDKey(t *testing.T) {
	if RequestIDKey != "request_id" {
		t.Errorf("Expected RequestIDKey to be 'request_id', got %s", RequestIDKey)
	}
}

// TestUserIDKey 测试用户ID键常量
func TestUserIDKey(t *testing.T) {
	if UserIDKey != "user_id" {
		t.Errorf("Expected UserIDKey to be 'user_id', got %s", UserIDKey)
	}
}

// TestInitLogger 测试日志系统初始化
func TestInitLogger(t *testing.T) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "test_init_logger")
	logFile := filepath.Join(tempDir, "test.log")
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		config  LogConfig
		wantErr bool
	}{
		{
			name: "valid console config",
			config: LogConfig{
				Level:  "info",
				Format: "json",
				Output: "console",
			},
			wantErr: false,
		},
		{
			name: "valid file config",
			config: LogConfig{
				Level:      "debug",
				Format:     "console",
				Output:     "file",
				FilePath:   logFile,
				MaxSize:    10,
				MaxAge:     7,
				MaxBackups: 3,
				Compress:   false,
			},
			wantErr: false,
		},
		{
			name: "invalid level",
			config: LogConfig{
				Level:  "invalid",
				Format: "json",
				Output: "console",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitLogger(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitLogger() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 如果初始化成功，验证全局logger是否设置
			if err == nil {
				if Logger == nil {
					t.Error("Global Logger should be set after successful initialization")
				}
				if SugaredLogger == nil {
					t.Error("Global SugaredLogger should be set after successful initialization")
				}
			}
		})
	}
}

// setupTestLogger 设置测试用的logger
func setupTestLogger(t *testing.T) {
	config := LogConfig{
		Level:  "info",
		Format: "json",
		Output: "console",
	}

	err := InitLogger(config)
	if err != nil {
		t.Fatalf("Failed to setup test logger: %v", err)
	}
}

// TestGetEncoderConfig 测试编码器配置
func TestGetEncoderConfig(t *testing.T) {
	config := getEncoderConfig()

	// 验证关键字段
	if config.TimeKey != "timestamp" {
		t.Errorf("Expected TimeKey to be 'timestamp', got %s", config.TimeKey)
	}
	if config.LevelKey != "level" {
		t.Errorf("Expected LevelKey to be 'level', got %s", config.LevelKey)
	}
	if config.MessageKey != "message" {
		t.Errorf("Expected MessageKey to be 'message', got %s", config.MessageKey)
	}
	if config.CallerKey != "caller" {
		t.Errorf("Expected CallerKey to be 'caller', got %s", config.CallerKey)
	}
}

// TestCreateLogDirectoryIfNeeded 测试条件性日志目录创建
func TestCreateLogDirectoryIfNeeded(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "test_conditional_dir")
	logFile := filepath.Join(tempDir, "test.log")
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		config  LogConfig
		wantErr bool
	}{
		{
			name: "console output - no directory needed",
			config: LogConfig{
				Output: "console",
			},
			wantErr: false,
		},
		{
			name: "file output - directory needed",
			config: LogConfig{
				Output:   "file",
				FilePath: logFile,
			},
			wantErr: false,
		},
		{
			name: "both output - directory needed",
			config: LogConfig{
				Output:   "both",
				FilePath: logFile,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createLogDirectoryIfNeeded(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("createLogDirectoryIfNeeded() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 对于需要文件输出的情况，验证目录是否创建
			if (tt.config.Output == "file" || tt.config.Output == "both") && err == nil {
				if _, statErr := os.Stat(tempDir); os.IsNotExist(statErr) {
					t.Error("Directory should be created for file output")
				}
			}
		})
	}
}

// BenchmarkGetLogLevel 基准测试日志级别获取
func BenchmarkGetLogLevel(b *testing.B) {
	levels := []string{"debug", "info", "warn", "error", "panic", "fatal"}

	for i := 0; i < b.N; i++ {
		level := levels[i%len(levels)]
		_, _ = getLogLevel(level)
	}
}

// BenchmarkCreateEncoder 基准测试编码器创建
func BenchmarkCreateEncoder(b *testing.B) {
	config := LogConfig{Format: "json"}

	for i := 0; i < b.N; i++ {
		_ = createEncoder(config)
	}
}
