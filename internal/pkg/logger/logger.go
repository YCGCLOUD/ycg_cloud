package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局日志实例
var Logger *zap.Logger

// SugaredLogger 全局Sugar日志实例（支持格式化）
var SugaredLogger *zap.SugaredLogger

// LogConfig 日志配置结构
//
// LogConfig定义了日志系统的完整配置选项，支持灵活的日志级别、格式和输出配置：
//
// 级别控制：
// - debug: 最详细的日志，包含所有调试信息
// - info: 一般信息日志，默认级别
// - warn/warning: 警告级别，需要关注但不影响功能
// - error: 错误级别，表示发生了错误但程序可以继续
// - panic: 严重错误，会触发panic
// - fatal: 致命错误，会终止程序
//
// 格式支持：
// - json: 结构化JSON格式，适合日志收集和分析
// - console: 人类可读格式，适合开发调试
//
// 输出方式：
// - console: 仅输出到控制台
// - file: 仅输出到文件
// - both: 同时输出到控制台和文件
//
// 文件轮转：
// - MaxSize: 单个日志文件最大大小(MB)
// - MaxAge: 日志文件最大保留天数
// - MaxBackups: 最大备份文件数量
// - Compress: 是否压缩轮转的日志文件
type LogConfig struct {
	Level      string `yaml:"level" mapstructure:"level"`             // 日志级别：debug/info/warn/error/panic/fatal
	Format     string `yaml:"format" mapstructure:"format"`           // 日志格式：json/console
	Output     string `yaml:"output" mapstructure:"output"`           // 输出方式：file/console/both
	FilePath   string `yaml:"file_path" mapstructure:"file_path"`     // 日志文件路径
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size"`       // 最大文件大小(MB)
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age"`         // 最大保留天数
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"` // 最大备份文件数
	Compress   bool   `yaml:"compress" mapstructure:"compress"`       // 是否压缩历史文件
}

// RequestID 请求ID键
type RequestID string

const (
	// RequestIDKey 请求ID在上下文中的键
	RequestIDKey RequestID = "request_id"
	// UserIDKey 用户ID在上下文中的键
	UserIDKey RequestID = "user_id"
)

// InitLogger 初始化日志系统
//
// InitLogger根据提供的配置初始化整个日志系统，包括：
// 1. 创建必要的日志目录
// 2. 解析和验证日志级别
// 3. 创建适当的编码器（JSON或Console格式）
// 4. 配置输出目标（文件、控制台或两者）
// 5. 设置日志轮转策略
// 6. 创建全局Logger和SugaredLogger实例
//
// 初始化成功后，可以通过全局变量Logger、SugaredLogger或包级别的便捷函数使用日志系统。
//
// 参数:
//   - config: 日志配置结构，包含所有必要的配置选项
//
// 返回:
//   - error: 初始化错误，nil表示成功
//
// 使用示例:
//
//	config := LogConfig{
//	    Level: "info",
//	    Format: "json",
//	    Output: "both",
//	    FilePath: "logs/app.log",
//	    MaxSize: 100,
//	    MaxAge: 7,
//	    MaxBackups: 10,
//	    Compress: true,
//	}
//	err := InitLogger(config)
func InitLogger(config LogConfig) error {
	// 确保日志目录存在
	if err := createLogDirectoryIfNeeded(config); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 获取日志级别
	level, err := getLogLevel(config.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// 创建编码器
	encoder := createEncoder(config)

	// 创建输出Writer
	writeSyncer, err := createWriteSyncer(config)
	if err != nil {
		return err
	}

	// 创建和设置Logger
	return setupLogger(encoder, writeSyncer, level, config)
}

// createLogDirectoryIfNeeded 创建日志目录（如果需要）
func createLogDirectoryIfNeeded(config LogConfig) error {
	if config.Output == "file" || config.Output == "both" {
		return ensureLogDir(config.FilePath)
	}
	return nil
}

// createEncoder 创建日志编码器
func createEncoder(config LogConfig) zapcore.Encoder {
	encoderConfig := getEncoderConfig()
	if config.Format == "json" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// createWriteSyncer 创建输出Writer
func createWriteSyncer(config LogConfig) (zapcore.WriteSyncer, error) {
	var writers []zapcore.WriteSyncer

	// 添加控制台输出
	if config.Output == "console" || config.Output == "both" {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	// 添加文件输出
	if config.Output == "file" || config.Output == "both" {
		fileWriter := createFileWriter(config)
		writers = append(writers, zapcore.AddSync(fileWriter))
	}

	// 组合多个输出
	if len(writers) == 1 {
		return writers[0], nil
	}
	return zapcore.NewMultiWriteSyncer(writers...), nil
}

// createFileWriter 创建文件Writer
func createFileWriter(config LogConfig) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   config.FilePath,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		Compress:   config.Compress,
	}
}

// setupLogger 设置Logger
func setupLogger(encoder zapcore.Encoder, writeSyncer zapcore.WriteSyncer, level zapcore.Level, config LogConfig) error {
	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建Logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	SugaredLogger = Logger.Sugar()

	// 替换全局Logger
	zap.ReplaceGlobals(Logger)

	Logger.Info("Logger initialized successfully",
		zap.String("level", config.Level),
		zap.String("format", config.Format),
		zap.String("output", config.Output),
	)

	return nil
}

// getEncoderConfig 获取编码器配置
func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// customTimeEncoder 自定义时间编码器
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// getLogLevel 获取日志级别
func getLogLevel(level string) (zapcore.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown level: %s", level)
	}
}

// ensureLogDir 确保日志目录存在
func ensureLogDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0750)
}

// WithRequestID 为日志添加请求ID
//
// 为日志记录添加请求ID字段，用于追踪同一请求的所有日志。这对于微服务架构
// 和分布式系统中的请求追踪非常有用。返回的Logger会自动在所有日志中包含请求ID。
//
// 参数:
//   - ctx: 上下文对象（当前未使用，为了API一致性保留）
//   - requestID: 请求的唯一标识符
//
// 返回:
//   - *zap.Logger: 包含请求ID的Logger实例
//
// 使用示例:
//
//	logger := WithRequestID(ctx, "req-123-456")
//	logger.Info("Processing request")
func WithRequestID(ctx context.Context, requestID string) *zap.Logger {
	return Logger.With(zap.String("request_id", requestID))
}

// WithUserID 为日志添加用户ID
//
// 为日志记录添加用户ID字段，同时保留上下文中的请求ID（如果存在）。
// 这对于用户行为追踪和安全审计非常有用。
//
// 参数:
//   - ctx: 上下文对象，用于提取请求ID
//   - userID: 用户的唯一标识符
//
// 返回:
//   - *zap.Logger: 包含用户ID和可能的请求ID的Logger实例
//
// 使用示例:
//
//	logger := WithUserID(ctx, "user-789")
//	logger.Info("User action performed")
func WithUserID(ctx context.Context, userID string) *zap.Logger {
	logger := Logger
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		logger = logger.With(zap.String("request_id", requestID.(string)))
	}
	return logger.With(zap.String("user_id", userID))
}

// WithContext 从上下文中提取信息并添加到日志
//
// 从上下文中自动提取请求ID和用户ID等信息，并添加到日志记录中。
// 这是一个便捷函数，可以一次性提取上下文中的所有相关信息。
//
// 参数:
//   - ctx: 包含请求ID和用户ID等信息的上下文对象
//
// 返回:
//   - *zap.Logger: 包含上下文信息的Logger实例
//
// 使用示例:
//
//	logger := WithContext(ctx)
//	logger.Info("Operation completed")  // 自动包含请求ID和用户ID
func WithContext(ctx context.Context) *zap.Logger {
	logger := Logger

	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		logger = logger.With(zap.String("request_id", requestID.(string)))
	}

	if userID := ctx.Value(UserIDKey); userID != nil {
		logger = logger.With(zap.String("user_id", userID.(string)))
	}

	return logger
}

// Sync 同步日志缓冲区
func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

// Close 关闭日志系统
func Close() error {
	return Sync()
}

// Info 简化的Info日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Debug 简化的Debug日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Warn 简化的Warn日志
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error 简化的Error日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal 简化的Fatal日志
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// Panic 简化的Panic日志
func Panic(msg string, fields ...zap.Field) {
	Logger.Panic(msg, fields...)
}
