package logger

import (
	"io"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// AccessLogger 访问日志实例
var AccessLogger *zap.Logger

// AccessLogConfig 访问日志配置
type AccessLogConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`     // 是否启用
	FilePath string `yaml:"file_path" mapstructure:"file_path"` // 日志文件路径
	Format   string `yaml:"format" mapstructure:"format"`       // 日志格式
}

// AccessLogEntry 访问日志条目结构
type AccessLogEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	RequestID    string    `json:"request_id"`
	UserID       string    `json:"user_id,omitempty"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Query        string    `json:"query,omitempty"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int64     `json:"response_time"` // 毫秒
	UserAgent    string    `json:"user_agent,omitempty"`
	IPAddress    string    `json:"ip_address"`
	RequestSize  int64     `json:"request_size,omitempty"`
	ResponseSize int64     `json:"response_size,omitempty"`
	Referer      string    `json:"referer,omitempty"`
	Protocol     string    `json:"protocol,omitempty"`
}

// InitAccessLogger 初始化访问日志系统
func InitAccessLogger(config AccessLogConfig) error {
	if !config.Enabled {
		AccessLogger = zap.NewNop() // 禁用时使用空Logger
		return nil
	}

	// 确保日志目录存在
	if err := ensureLogDir(config.FilePath); err != nil {
		return err
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "timestamp",
		LevelKey:      zapcore.OmitKey, // 访问日志不需要level
		NameKey:       zapcore.OmitKey,
		CallerKey:     zapcore.OmitKey,
		MessageKey:    zapcore.OmitKey, // 使用结构化字段，不需要message
		StacktraceKey: zapcore.OmitKey,
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeTime:    zapcore.ISO8601TimeEncoder,
	}

	// 创建JSON编码器
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 创建文件Writer
	fileWriter := &lumberjack.Logger{
		Filename:   config.FilePath,
		MaxSize:    100, // 100MB
		MaxAge:     30,  // 30天
		MaxBackups: 10,  // 10个备份文件
		Compress:   true,
	}

	// 创建核心
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(fileWriter),
		zapcore.InfoLevel,
	)

	// 创建Logger
	AccessLogger = zap.New(core)

	return nil
}

// LogAccess 记录访问日志
func LogAccess(entry AccessLogEntry) {
	if AccessLogger == nil {
		return
	}

	fields := []zap.Field{
		zap.Time("timestamp", entry.Timestamp),
		zap.String("request_id", entry.RequestID),
		zap.String("method", entry.Method),
		zap.String("path", entry.Path),
		zap.Int("status_code", entry.StatusCode),
		zap.Int64("response_time", entry.ResponseTime),
		zap.String("ip_address", entry.IPAddress),
	}

	// 条件性添加字段
	if entry.UserID != "" {
		fields = append(fields, zap.String("user_id", entry.UserID))
	}
	if entry.Query != "" {
		fields = append(fields, zap.String("query", entry.Query))
	}
	if entry.UserAgent != "" {
		fields = append(fields, zap.String("user_agent", entry.UserAgent))
	}
	if entry.RequestSize > 0 {
		fields = append(fields, zap.Int64("request_size", entry.RequestSize))
	}
	if entry.ResponseSize > 0 {
		fields = append(fields, zap.Int64("response_size", entry.ResponseSize))
	}
	if entry.Referer != "" {
		fields = append(fields, zap.String("referer", entry.Referer))
	}
	if entry.Protocol != "" {
		fields = append(fields, zap.String("protocol", entry.Protocol))
	}

	AccessLogger.Info("", fields...)
}

// GetAccessLogWriter 获取访问日志Writer（用于Gin中间件）
func GetAccessLogWriter(config AccessLogConfig) (io.Writer, error) {
	if !config.Enabled {
		return io.Discard, nil
	}

	// 确保日志目录存在
	if err := ensureLogDir(config.FilePath); err != nil {
		return nil, err
	}

	// 返回文件Writer
	return &lumberjack.Logger{
		Filename:   config.FilePath,
		MaxSize:    100, // 100MB
		MaxAge:     30,  // 30天
		MaxBackups: 10,  // 10个备份文件
		Compress:   true,
	}, nil
}

// SyncAccessLogger 同步访问日志缓冲区
func SyncAccessLogger() error {
	if AccessLogger != nil {
		return AccessLogger.Sync()
	}
	return nil
}
