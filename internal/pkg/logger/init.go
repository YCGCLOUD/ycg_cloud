package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// InitConfig 日志初始化配置
type InitConfig struct {
	AppLog    LogConfig       `yaml:"app_log" mapstructure:"app_log"`
	AccessLog AccessLogConfig `yaml:"access_log" mapstructure:"access_log"`
}

// InitializeLoggerSystem 初始化整个日志系统
func InitializeLoggerSystem(config InitConfig) error {
	// 1. 创建日志目录
	if err := createLogDirectories(config); err != nil {
		return fmt.Errorf("failed to create log directories: %w", err)
	}

	// 2. 初始化应用日志
	if err := InitLogger(config.AppLog); err != nil {
		return fmt.Errorf("failed to initialize app logger: %w", err)
	}

	// 3. 初始化访问日志
	if err := InitAccessLogger(config.AccessLog); err != nil {
		return fmt.Errorf("failed to initialize access logger: %w", err)
	}

	// 4. 初始化结构化日志
	InitStructuredLogger()

	Logger.Info("Logger system initialized successfully",
		String("app_log_level", config.AppLog.Level),
		String("app_log_format", config.AppLog.Format),
		String("app_log_output", config.AppLog.Output),
		Bool("access_log_enabled", config.AccessLog.Enabled),
	)

	return nil
}

// createLogDirectories 创建日志目录
func createLogDirectories(config InitConfig) error {
	directories := make(map[string]bool)

	// 应用日志目录
	if config.AppLog.Output == "file" || config.AppLog.Output == "both" {
		if config.AppLog.FilePath != "" {
			dir := filepath.Dir(config.AppLog.FilePath)
			directories[dir] = true
		}
	}

	// 访问日志目录
	if config.AccessLog.Enabled && config.AccessLog.FilePath != "" {
		dir := filepath.Dir(config.AccessLog.FilePath)
		directories[dir] = true
	}

	// 创建目录
	for dir := range directories {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// DefaultLogConfig 获取默认日志配置
func DefaultLogConfig() InitConfig {
	return InitConfig{
		AppLog: LogConfig{
			Level:      "info",
			Format:     "json",
			Output:     "both",
			FilePath:   "logs/app.log",
			MaxSize:    100, // 100MB
			MaxAge:     30,  // 30天
			MaxBackups: 5,   // 5个备份文件
			Compress:   true,
		},
		AccessLog: AccessLogConfig{
			Enabled:  true,
			FilePath: "logs/access.log",
			Format:   "json",
		},
	}
}

// ProductionLogConfig 生产环境日志配置
func ProductionLogConfig() InitConfig {
	return InitConfig{
		AppLog: LogConfig{
			Level:      "warn",
			Format:     "json",
			Output:     "file",
			FilePath:   "logs/app.log",
			MaxSize:    200, // 200MB
			MaxAge:     60,  // 60天
			MaxBackups: 10,  // 10个备份文件
			Compress:   true,
		},
		AccessLog: AccessLogConfig{
			Enabled:  true,
			FilePath: "logs/access.log",
			Format:   "json",
		},
	}
}

// DevelopmentLogConfig 开发环境日志配置
func DevelopmentLogConfig() InitConfig {
	return InitConfig{
		AppLog: LogConfig{
			Level:      "debug",
			Format:     "console",
			Output:     "both",
			FilePath:   "logs/app.log",
			MaxSize:    50, // 50MB
			MaxAge:     7,  // 7天
			MaxBackups: 3,  // 3个备份文件
			Compress:   false,
		},
		AccessLog: AccessLogConfig{
			Enabled:  true,
			FilePath: "logs/access.log",
			Format:   "json",
		},
	}
}

// TestingLogConfig 测试环境日志配置
func TestingLogConfig() InitConfig {
	return InitConfig{
		AppLog: LogConfig{
			Level:      "debug",
			Format:     "console",
			Output:     "console",
			FilePath:   "",
			MaxSize:    10,
			MaxAge:     1,
			MaxBackups: 1,
			Compress:   false,
		},
		AccessLog: AccessLogConfig{
			Enabled:  false,
			FilePath: "",
			Format:   "json",
		},
	}
}

// GetLogConfigByEnv 根据环境获取日志配置
func GetLogConfigByEnv(env string) InitConfig {
	switch env {
	case "production", "prod":
		return ProductionLogConfig()
	case "development", "dev":
		return DevelopmentLogConfig()
	case "testing", "test":
		return TestingLogConfig()
	default:
		return DefaultLogConfig()
	}
}

// LogRotationInfo 日志轮转信息
type LogRotationInfo struct {
	CurrentSize int64  `json:"current_size"`
	MaxSize     int64  `json:"max_size"`
	FileCount   int    `json:"file_count"`
	MaxBackups  int    `json:"max_backups"`
	OldestFile  string `json:"oldest_file,omitempty"`
	NewestFile  string `json:"newest_file,omitempty"`
}

// GetLogRotationInfo 获取日志轮转信息
func GetLogRotationInfo(logFilePath string) (*LogRotationInfo, error) {
	if logFilePath == "" {
		return nil, fmt.Errorf("log file path is empty")
	}

	// 获取当前日志文件信息
	info, err := getCurrentLogFileInfo(logFilePath)
	if err != nil {
		return nil, err
	}

	// 获取轮转文件信息
	if err := addRotationFileInfo(info, logFilePath); err != nil {
		// 即使获取轮转文件信息失败，也返回基础信息
		return info, nil
	}

	return info, nil
}

// getCurrentLogFileInfo 获取当前日志文件信息
func getCurrentLogFileInfo(logFilePath string) (*LogRotationInfo, error) {
	fileInfo, err := os.Stat(logFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &LogRotationInfo{
				CurrentSize: 0,
				FileCount:   0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get log file info: %w", err)
	}

	return &LogRotationInfo{
		CurrentSize: fileInfo.Size(),
	}, nil
}

// addRotationFileInfo 添加轮转文件信息
func addRotationFileInfo(info *LogRotationInfo, logFilePath string) error {
	dir := filepath.Dir(logFilePath)
	baseName := filepath.Base(logFilePath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	logFiles := findRotationLogFiles(entries, baseName)
	info.FileCount = len(logFiles)

	if len(logFiles) > 0 {
		info.OldestFile = logFiles[0]
		info.NewestFile = logFiles[len(logFiles)-1]
	}

	return nil
}

// findRotationLogFiles 查找轮转日志文件
func findRotationLogFiles(entries []os.DirEntry, baseName string) []string {
	var logFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if isRotationLogFile(name, baseName) {
			logFiles = append(logFiles, name)
		}
	}
	return logFiles
}

// isRotationLogFile 检查是否是轮转日志文件
func isRotationLogFile(name, baseName string) bool {
	if name == baseName {
		return true
	}

	return len(name) > len(baseName) &&
		name[:len(baseName)] == baseName &&
		(name[len(baseName)] == '.' || name[len(baseName)] == '-')
}

// CleanupOldLogs 清理旧日志文件
func CleanupOldLogs(logFilePath string, maxAge int) error {
	if logFilePath == "" || maxAge <= 0 {
		return nil
	}

	dir := filepath.Dir(logFilePath)
	baseName := filepath.Base(logFilePath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -maxAge)
	cleanedCount := cleanupExpiredLogFiles(entries, dir, baseName, cutoff)

	logCleanupResult(logFilePath, cleanedCount, maxAge)
	return nil
}

// cleanupExpiredLogFiles 清理过期的日志文件
func cleanupExpiredLogFiles(entries []os.DirEntry, dir, baseName string, cutoff time.Time) int {
	var cleanedCount int

	for _, entry := range entries {
		if shouldSkipEntry(entry, baseName) {
			continue
		}

		if shouldCleanupFile(entry, dir, cutoff) {
			cleanedCount++
		}
	}

	return cleanedCount
}

// shouldSkipEntry 检查是否应该跳过该条目
func shouldSkipEntry(entry os.DirEntry, baseName string) bool {
	if entry.IsDir() {
		return true
	}

	name := entry.Name()
	if name == baseName {
		return true // 不删除当前日志文件
	}

	// 检查是否是轮转的日志文件
	return !(len(name) > len(baseName) && name[:len(baseName)] == baseName)
}

// shouldCleanupFile 检查是否应该清理文件
func shouldCleanupFile(entry os.DirEntry, dir string, cutoff time.Time) bool {
	name := entry.Name()
	fullPath := filepath.Join(dir, name)

	fileInfo, err := entry.Info()
	if err != nil {
		return false
	}

	if fileInfo.ModTime().Before(cutoff) {
		return os.Remove(fullPath) == nil
	}

	return false
}

// logCleanupResult 记录清理结果
func logCleanupResult(logFilePath string, cleanedCount, maxAge int) {
	if cleanedCount > 0 {
		Logger.Info("Cleaned up old log files",
			String("log_file", logFilePath),
			Int("cleaned_count", cleanedCount),
			Int("max_age_days", maxAge),
		)
	}
}

// ForceRotateLog 强制轮转日志文件
func ForceRotateLog() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

// 便捷的字段构造函数
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}
