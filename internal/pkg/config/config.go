package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var (
	// AppConfig 全局配置实例
	AppConfig *Config
)

// Load 加载配置文件
//
// 此函数负责从多个源加载和合并配置信息，包括：
// 1. 默认配置文件 (config.yaml)
// 2. 环境特定配置文件 (config.dev.yaml, config.prod.yaml 等)
// 3. 环境变量 (.env 文件和系统环境变量)
//
// 加载顺序和优先级：
// 1. 默认配置文件作为基础
// 2. 环境特定配置覆盖默认配置
// 3. 环境变量具有最高优先级，可以覆盖所有配置
//
// 支持的环境变量格式：
// - CLOUDPAN_DATABASE_MYSQL_HOST (数据库主机)
// - CLOUDPAN_REDIS_HOST (Redis主机)
// - CLOUDPAN_JWT_SECRET (JWT密钥)
// 等等...
//
// 返回值：
//   - error: 配置加载或验证失败时的错误信息
func Load() error {
	// 设置基础配置
	if err := setupViperConfig(); err != nil {
		return fmt.Errorf("failed to setup viper config: %w", err)
	}

	// 加载配置文件
	if err := loadConfigFiles(); err != nil {
		return fmt.Errorf("failed to load config files: %w", err)
	}

	// 处理环境变量
	if err := setupEnvironmentVars(); err != nil {
		return fmt.Errorf("failed to setup environment variables: %w", err)
	}

	// 解析和验证配置
	return parseAndValidateConfig()
}

// setupViperConfig 设置Viper基础配置
func setupViperConfig() error {
	// 设置配置文件搜索路径
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")
	viper.AddConfigPath("/etc/cloudpan")

	// 设置配置文件类型
	viper.SetConfigType("yaml")
	return nil
}

// loadConfigFiles 加载配置文件
func loadConfigFiles() error {
	// 首先加载默认配置
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read default config file: %w", err)
	}

	// 加载环境特定配置
	return loadEnvironmentConfig()
}

// loadEnvironmentConfig 加载环境特定配置
func loadEnvironmentConfig() error {
	env := getEnvironment()
	envConfigName := getEnvConfigName(env)

	viper.SetConfigName(envConfigName)

	// 尝试读取环境特定配置（不是必须的）
	if err := viper.MergeInConfig(); err != nil {
		// 如果是文件不存在错误，忽略它
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to merge %s config file: %w", envConfigName, err)
		}
	}

	return nil
}

// getEnvironment 获取当前环境
func getEnvironment() string {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}
	return env
}

// getEnvConfigName 获取环境配置文件名
func getEnvConfigName(env string) string {
	// 映射环境名称到配置文件名
	envFileMap := map[string]string{
		"development": "dev",
		"testing":     "test",
		"production":  "prod",
	}

	envConfigSuffix := envFileMap[env]
	if envConfigSuffix == "" {
		envConfigSuffix = env // 如果没有映射，直接使用原环境名
	}
	return fmt.Sprintf("config.%s", envConfigSuffix)
}

// setupEnvironmentVars 设置环境变量支持
func setupEnvironmentVars() error {
	// 加载.env文件（敏感信息）
	env := getEnvironment()
	if err := loadEnvFile(env); err != nil {
		// .env文件不是必须的，只记录警告
		fmt.Printf("Warning: failed to load .env file: %v\n", err)
	}

	// 支持环境变量覆盖配置（必须在加载.env文件后设置）
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("CLOUDPAN")

	// 手动绑定关键的环境变量到配置路径
	bindEnvVars()
	return nil
}

// parseAndValidateConfig 解析和验证配置
func parseAndValidateConfig() error {
	// 解析配置到结构体
	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证必要的配置项
	if err := validateConfig(AppConfig); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 创建必要的目录
	if err := createDirectories(AppConfig); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	return nil
}

// LoadFromFile 从指定文件加载配置
func LoadFromFile(configPath string) error {
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// 支持环境变量覆盖
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("CLOUDPAN")

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(AppConfig); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	if err := createDirectories(AppConfig); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	return nil
}

// validateConfig 验证配置的有效性
func validateConfig(cfg *Config) error {
	validators := []func(*Config) error{
		validateAppConfig,
		validateServerConfig,
		validateDatabaseConfig,
		validateRedisConfig,
		validateJWTConfig,
		validateStorageConfig,
		validateEmailConfig,
	}

	for _, validator := range validators {
		if err := validator(cfg); err != nil {
			return err
		}
	}

	return nil
}

// validateRequired 通用的必填字段验证函数
func validateRequired(fieldName, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// validateRange 验证数值范围
func validateRange(fieldName string, value, min, max int) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d", fieldName, min, max)
	}
	return nil
}

// validateMinLength 验证最小长度
func validateMinLength(fieldName, value string, minLength int) error {
	if len(value) < minLength {
		return fmt.Errorf("%s must be at least %d characters long", fieldName, minLength)
	}
	return nil
}

// validateAppConfig 验证应用配置
func validateAppConfig(cfg *Config) error {
	return validateRequired("app.name", cfg.App.Name)
}

// validateServerConfig 验证服务器配置
func validateServerConfig(cfg *Config) error {
	return validateRange("server.port", cfg.Server.Port, 1, 65535)
}

// validateDatabaseConfig 验证数据库配置
func validateDatabaseConfig(cfg *Config) error {
	if err := validateRequired("database.mysql.host", cfg.Database.MySQL.Host); err != nil {
		return err
	}
	if err := validateRequired("database.mysql.username", cfg.Database.MySQL.Username); err != nil {
		return err
	}
	return validateRequired("database.mysql.dbname", cfg.Database.MySQL.DBName)
}

// validateRedisConfig 验证Redis配置
func validateRedisConfig(cfg *Config) error {
	return validateRequired("redis.host", cfg.Redis.Host)
}

// validateJWTConfig 验证JWT配置
func validateJWTConfig(cfg *Config) error {
	if err := validateRequired("jwt.secret", cfg.JWT.Secret); err != nil {
		return err
	}
	return validateMinLength("jwt.secret", cfg.JWT.Secret, 32)
}

// validateStorageConfig 验证存储配置
func validateStorageConfig(cfg *Config) error {
	if cfg.Storage.Local.Enabled && cfg.Storage.Local.RootPath == "" {
		return fmt.Errorf("storage.local.root_path is required when local storage is enabled")
	}

	if cfg.Storage.OSS.Enabled {
		return validateOSSConfig(cfg)
	}
	return nil
}

// validateOSSConfig 验证OSS配置
func validateOSSConfig(cfg *Config) error {
	if cfg.Storage.OSS.AccessKeyID == "" {
		return fmt.Errorf("storage.oss.access_key_id is required when OSS is enabled")
	}
	if cfg.Storage.OSS.AccessKeySecret == "" {
		return fmt.Errorf("storage.oss.access_key_secret is required when OSS is enabled")
	}
	if cfg.Storage.OSS.BucketName == "" {
		return fmt.Errorf("storage.oss.bucket_name is required when OSS is enabled")
	}
	return nil
}

// validateEmailConfig 验证邮件配置
func validateEmailConfig(cfg *Config) error {
	if cfg.Email.SMTP.Host == "" {
		return fmt.Errorf("email.smtp.host is required")
	}
	if cfg.Email.SMTP.FromEmail == "" {
		return fmt.Errorf("email.smtp.from_email is required")
	}
	return nil
}

// createDirectories 创建必要的目录
func createDirectories(cfg *Config) error {
	directories := collectDirectoriesToCreate(cfg)
	return createDirectoriesFromList(directories)
}

// collectDirectoriesToCreate 收集需要创建的目录
func collectDirectoriesToCreate(cfg *Config) []string {
	var directories []string

	// 收集存储目录
	directories = append(directories, collectStorageDirectories(cfg)...)

	// 收集日志目录
	directories = append(directories, collectLogDirectories(cfg)...)

	// 收集国际化目录
	directories = append(directories, collectI18nDirectories(cfg)...)

	return directories
}

// collectStorageDirectories 收集存储目录
func collectStorageDirectories(cfg *Config) []string {
	var directories []string

	if cfg.Storage.Local.Enabled {
		directories = append(directories, cfg.Storage.Local.RootPath)
		if cfg.Storage.Local.TempPath != "" {
			directories = append(directories, cfg.Storage.Local.TempPath)
		}
	}

	return directories
}

// collectLogDirectories 收集日志目录
func collectLogDirectories(cfg *Config) []string {
	var directories []string

	// 主日志目录
	if cfg.Log.Output == "file" && cfg.Log.FilePath != "" {
		logDir := filepath.Dir(cfg.Log.FilePath)
		directories = append(directories, logDir)
	}

	// 访问日志目录
	if cfg.Log.AccessLog.Enabled && cfg.Log.AccessLog.FilePath != "" {
		accessLogDir := filepath.Dir(cfg.Log.AccessLog.FilePath)
		directories = append(directories, accessLogDir)
	}

	return directories
}

// collectI18nDirectories 收集国际化目录
func collectI18nDirectories(cfg *Config) []string {
	var directories []string

	if cfg.I18n.Path != "" {
		directories = append(directories, cfg.I18n.Path)
	}

	return directories
}

// createDirectoriesFromList 从目录列表创建目录
func createDirectoriesFromList(directories []string) error {
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return AppConfig
}

// IsProduction 判断是否为生产环境
func IsProduction() bool {
	return AppConfig != nil && AppConfig.App.Env == "production"
}

// IsDevelopment 判断是否为开发环境
func IsDevelopment() bool {
	return AppConfig != nil && AppConfig.App.Env == "development"
}

// IsTesting 判断是否为测试环境
func IsTesting() bool {
	return AppConfig != nil && AppConfig.App.Env == "testing"
}

// GetDSN 获取MySQL数据库连接字符串（兼容MySQL 8.0.31）
func GetDSN() string {
	if AppConfig == nil {
		return ""
	}

	mysql := AppConfig.Database.MySQL
	// 解决URL编码问题，对loc参数进行特殊处理
	loc := mysql.Loc
	if loc == "Local" {
		loc = "Local" // 保持不变，MySQL驱动会自动处理
	}

	// 根据记忆中的MySQL 8.0.31要求，添加allowNativePasswords=true参数
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s&allowNativePasswords=true",
		mysql.Username,
		mysql.Password,
		mysql.Host,
		mysql.Port,
		mysql.DBName,
		mysql.Charset,
		mysql.ParseTime,
		loc,
	)
}

// GetRedisAddr 获取Redis连接地址
func GetRedisAddr() string {
	if AppConfig == nil {
		return ""
	}

	return fmt.Sprintf("%s:%d", AppConfig.Redis.Host, AppConfig.Redis.Port)
}

// GetServerAddr 获取服务器监听地址
func GetServerAddr() string {
	if AppConfig == nil {
		return ":8080"
	}

	return fmt.Sprintf("%s:%d", AppConfig.Server.Host, AppConfig.Server.Port)
}

// loadEnvFile 加载环境特定的.env文件并设置到环境变量
func loadEnvFile(env string) error {
	// 映射环境名称到.env文件后缀
	envFileMap := map[string]string{
		"development": "dev",
		"testing":     "test",
		"production":  "prod",
	}

	// 获取映射后的文件后缀
	envFileSuffix := envFileMap[env]
	if envFileSuffix == "" {
		envFileSuffix = env // 如果没有映射，直接使用原环境名
	}

	// 尝试加载环境特定的.env文件
	envFiles := []string{
		fmt.Sprintf(".env.%s", envFileSuffix), // .env.dev
		".env",                                // 通用.env文件
	}

	for _, envFile := range envFiles {
		if _, err := os.Stat(envFile); err == nil {
			// 读取.env文件内容并设置为环境变量
			if err := loadEnvFileToEnvironment(envFile); err != nil {
				return fmt.Errorf("failed to load %s: %w", envFile, err)
			}
			fmt.Printf("Loaded environment file: %s\n", envFile)
			return nil
		}
	}

	return fmt.Errorf("no .env file found")
}

// loadEnvFileToEnvironment 读取.env文件并设置为环境变量
func loadEnvFileToEnvironment(envFile string) error {
	// 验证文件路径安全性
	if !isValidEnvFilePath(envFile) {
		return fmt.Errorf("invalid env file path: %s", envFile)
	}

	file, err := os.Open(envFile) // #nosec G304 - Path is validated by isValidEnvFilePath
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := &envFileScanner{}
	lines, err := scanner.readLines(file)
	if err != nil {
		return err
	}

	return processEnvLines(lines)
}

// isValidEnvFilePath 验证环境文件路径是否安全
func isValidEnvFilePath(path string) bool {
	// 防止路径遍历攻击
	if strings.Contains(path, "..") || strings.Contains(path, "/") || strings.Contains(path, "\\") {
		return false
	}

	// 只允许.env文件
	return strings.HasSuffix(path, ".env") || strings.Contains(path, ".env.")
}

// processEnvLines 处理环境变量行
func processEnvLines(lines []string) error {
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if shouldSkipLine(line) {
			continue
		}

		if err := parseAndSetEnvVar(line); err != nil {
			return err
		}
	}
	return nil
}

// shouldSkipLine 判断是否应该跳过该行
func shouldSkipLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "#")
}

// parseAndSetEnvVar 解析并设置环境变量
func parseAndSetEnvVar(line string) error {
	// 解析键值对
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return nil // 跳过无效行
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// 移除引号
	value = removeQuotes(value)

	// 设置环境变量
	if err := os.Setenv(key, value); err != nil {
		return fmt.Errorf("failed to set environment variable %s: %w", key, err)
	}
	return nil
}

// removeQuotes 移除值两边的引号
func removeQuotes(value string) string {
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return value[1 : len(value)-1]
	}
	return value
}

// envFileScanner 简单的.env文件扫描器
type envFileScanner struct{}

func (s *envFileScanner) readLines(file *os.File) ([]string, error) {
	var lines []string
	buf := make([]byte, 4096)
	current := ""

	for {
		n, err := file.Read(buf)
		if n > 0 {
			current += string(buf[:n])
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
	}

	// 分割行
	lines = strings.Split(current, "\n")
	// 处理Windows的\r\n
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, "\r")
	}

	return lines, nil
}

// bindEnvVars 手动绑定环境变量到配置路径
func bindEnvVars() {
	// 数据库相关环境变量绑定
	// 注意: viper.BindEnv 不返回错误，因为它只是设置内部映射
	viper.BindEnv("database.mysql.host", "CLOUDPAN_DATABASE_MYSQL_HOST")         // #nosec G104
	viper.BindEnv("database.mysql.port", "CLOUDPAN_DATABASE_MYSQL_PORT")         // #nosec G104
	viper.BindEnv("database.mysql.username", "CLOUDPAN_DATABASE_MYSQL_USERNAME") // #nosec G104
	viper.BindEnv("database.mysql.password", "CLOUDPAN_DATABASE_MYSQL_PASSWORD") // #nosec G104
	viper.BindEnv("database.mysql.dbname", "CLOUDPAN_DATABASE_MYSQL_DBNAME")     // #nosec G104

	// Redis相关环境变量绑定
	viper.BindEnv("redis.host", "CLOUDPAN_REDIS_HOST")         // #nosec G104
	viper.BindEnv("redis.port", "CLOUDPAN_REDIS_PORT")         // #nosec G104
	viper.BindEnv("redis.password", "CLOUDPAN_REDIS_PASSWORD") // #nosec G104
	viper.BindEnv("redis.db", "CLOUDPAN_REDIS_DB")             // #nosec G104

	// JWT相关环境变量绑定
	viper.BindEnv("jwt.secret", "CLOUDPAN_JWT_SECRET") // #nosec G104

	// 邮件相关环境变量绑定
	viper.BindEnv("email.smtp.username", "CLOUDPAN_EMAIL_SMTP_USERNAME")     // #nosec G104
	viper.BindEnv("email.smtp.password", "CLOUDPAN_EMAIL_SMTP_PASSWORD")     // #nosec G104
	viper.BindEnv("email.smtp.from_email", "CLOUDPAN_EMAIL_SMTP_FROM_EMAIL") // #nosec G104

	// OSS相关环境变量绑定
	viper.BindEnv("storage.oss.access_key_id", "CLOUDPAN_STORAGE_OSS_ACCESS_KEY_ID")         // #nosec G104
	viper.BindEnv("storage.oss.access_key_secret", "CLOUDPAN_STORAGE_OSS_ACCESS_KEY_SECRET") // #nosec G104
	viper.BindEnv("storage.oss.bucket_name", "CLOUDPAN_STORAGE_OSS_BUCKET_NAME")             // #nosec G104
	viper.BindEnv("storage.oss.endpoint", "CLOUDPAN_STORAGE_OSS_ENDPOINT")                   // #nosec G104
	viper.BindEnv("storage.oss.region", "CLOUDPAN_STORAGE_OSS_REGION")                       // #nosec G104

	// 服务器相关环境变量绑定
	viper.BindEnv("server.host", "CLOUDPAN_SERVER_HOST") // #nosec G104
	viper.BindEnv("server.port", "CLOUDPAN_SERVER_PORT") // #nosec G104

	// 日志相关环境变量绑定
	viper.BindEnv("log.level", "CLOUDPAN_LOG_LEVEL") // #nosec G104
}
