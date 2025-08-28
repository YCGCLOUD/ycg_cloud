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
type LogConfig struct {
	Level      string `yaml:"level" mapstructure:"level"`             // 日志级别
	Format     string `yaml:"format" mapstructure:"format"`           // 日志格式 json/console
	Output     string `yaml:"output" mapstructure:"output"`           // 输出方式 file/console/both
	FilePath   string `yaml:"file_path" mapstructure:"file_path"`     // 日志文件路径
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size"`       // 最大文件大小(MB)
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age"`         // 最大保留天数
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"` // 最大备份文件数
	Compress   bool   `yaml:"compress" mapstructure:"compress"`       // 是否压缩
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
func WithRequestID(ctx context.Context, requestID string) *zap.Logger {
	return Logger.With(zap.String("request_id", requestID))
}

// WithUserID 为日志添加用户ID
func WithUserID(ctx context.Context, userID string) *zap.Logger {
	logger := Logger
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		logger = logger.With(zap.String("request_id", requestID.(string)))
	}
	return logger.With(zap.String("user_id", userID))
}

// WithContext 从上下文中提取信息并添加到日志
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
