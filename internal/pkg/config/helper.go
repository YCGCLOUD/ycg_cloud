package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Validator 配置验证接口
type Validator interface {
	Validate() error
}

// ConfigHelper 配置帮助类
type ConfigHelper struct {
	config *Config
}

// NewConfigHelper 创建配置帮助类实例
func NewConfigHelper(cfg *Config) *ConfigHelper {
	return &ConfigHelper{config: cfg}
}

// GetStoragePath 获取用户存储路径
func (h *ConfigHelper) GetStoragePath(userID int64) string {
	if h.config.Storage.Local.RootPath == "" {
		return ""
	}

	return fmt.Sprintf("%s/user-%d", h.config.Storage.Local.RootPath, userID)
}

// GetAvatarPath 获取用户头像存储路径
func (h *ConfigHelper) GetAvatarPath(userID int64) string {
	template := h.config.User.Avatar.PathTemplate
	if template == "" {
		template = "/storage/user-{user_id}/avatars/"
	}

	path := strings.ReplaceAll(template, "{user_id}", strconv.FormatInt(userID, 10))

	// 确保路径以存储根目录开始
	if h.config.Storage.Local.RootPath != "" {
		if !strings.HasPrefix(path, h.config.Storage.Local.RootPath) {
			path = h.config.Storage.Local.RootPath + path
		}
	}

	return path
}

// GetAvatarFilename 获取头像文件名
func (h *ConfigHelper) GetAvatarFilename(ext string) string {
	template := h.config.User.Avatar.FilenameTemplate
	if template == "" {
		template = "avatar_{timestamp}.{ext}"
	}

	filename := strings.ReplaceAll(template, "{timestamp}", strconv.FormatInt(time.Now().Unix(), 10))
	filename = strings.ReplaceAll(filename, "{ext}", ext)

	return filename
}

// IsAllowedFileType 检查文件类型是否允许
func (h *ConfigHelper) IsAllowedFileType(mimeType string) bool {
	allowedTypes := h.config.Storage.Local.AllowedTypes
	if len(allowedTypes) == 0 {
		return true // 如果没有限制，则允许所有类型
	}

	for _, allowedType := range allowedTypes {
		if allowedType == mimeType {
			return true
		}
	}

	return false
}

// IsAllowedAvatarType 检查头像文件类型是否允许
func (h *ConfigHelper) IsAllowedAvatarType(mimeType string) bool {
	allowedTypes := h.config.User.Avatar.AllowedTypes
	if len(allowedTypes) == 0 {
		return false // 头像必须有类型限制
	}

	for _, allowedType := range allowedTypes {
		if allowedType == mimeType {
			return true
		}
	}

	return false
}

// ShouldUseOSS 判断是否应该使用OSS存储
func (h *ConfigHelper) ShouldUseOSS(fileSize int64) bool {
	if !h.config.Storage.OSS.Enabled {
		return false
	}

	return fileSize >= h.config.Storage.OSS.AutoSwitchSize
}

// GetCacheTTL 获取缓存TTL
func (h *ConfigHelper) GetCacheTTL(cacheType string) time.Duration {
	switch cacheType {
	case "user_info":
		return h.config.Cache.UserInfoTTL
	case "file_info":
		return h.config.Cache.FileInfoTTL
	case "verification_code":
		return h.config.Cache.VerificationCodeTTL
	default:
		return h.config.Cache.DefaultTTL
	}
}

// GetJWTExpiration 获取JWT过期时间
func (h *ConfigHelper) GetJWTExpiration() time.Duration {
	return time.Duration(h.config.JWT.ExpireHours) * time.Hour
}

// GetJWTRefreshExpiration 获取JWT刷新过期时间
func (h *ConfigHelper) GetJWTRefreshExpiration() time.Duration {
	return time.Duration(h.config.JWT.RefreshExpireHours) * time.Hour
}

// IsPasswordValid 验证密码强度
func (h *ConfigHelper) IsPasswordValid(password string) error {
	cfg := h.config.User.Password

	validators := []func(string, *PasswordConfig) error{
		validatePasswordLength,
		validatePasswordNumber,
		validatePasswordLetter,
		validatePasswordSpecial,
	}

	for _, validator := range validators {
		if err := validator(password, &cfg); err != nil {
			return err
		}
	}

	return nil
}

// validatePasswordLength 验证密码长度
func validatePasswordLength(password string, cfg *PasswordConfig) error {
	if len(password) < cfg.MinLength {
		return fmt.Errorf("password must be at least %d characters long", cfg.MinLength)
	}
	if len(password) > cfg.MaxLength {
		return fmt.Errorf("password must be no more than %d characters long", cfg.MaxLength)
	}
	return nil
}

// validatePasswordNumber 验证密码数字要求
func validatePasswordNumber(password string, cfg *PasswordConfig) error {
	if !cfg.RequireNumber {
		return nil
	}

	for _, char := range password {
		if char >= '0' && char <= '9' {
			return nil
		}
	}
	return fmt.Errorf("password must contain at least one number")
}

// validatePasswordLetter 验证密码字母要求
func validatePasswordLetter(password string, cfg *PasswordConfig) error {
	if !cfg.RequireLetter {
		return nil
	}

	for _, char := range password {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			return nil
		}
	}
	return fmt.Errorf("password must contain at least one letter")
}

// validatePasswordSpecial 验证密码特殊字符要求
func validatePasswordSpecial(password string, cfg *PasswordConfig) error {
	if !cfg.RequireSpecial {
		return nil
	}

	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, char := range password {
		for _, special := range specialChars {
			if char == special {
				return nil
			}
		}
	}
	return fmt.Errorf("password must contain at least one special character")
}

// GetEmailTemplateFilename 获取邮件模板文件名
func (h *ConfigHelper) GetEmailTemplateFilename(templateType string) string {
	switch templateType {
	case "verify_code":
		return h.config.Email.Templates.VerifyCode
	case "password_reset":
		return h.config.Email.Templates.PasswordReset
	case "welcome":
		return h.config.Email.Templates.Welcome
	default:
		return ""
	}
}

// GetQueueStreamName 获取队列流名称
func (h *ConfigHelper) GetQueueStreamName(streamType string) string {
	if !h.config.Queue.RedisStream.Enabled {
		return ""
	}

	streamName, exists := h.config.Queue.RedisStream.Streams[streamType]
	if !exists {
		return ""
	}

	return streamName
}

// EnvironmentVariables 环境变量映射
var EnvironmentVariables = map[string]string{
	// 数据库相关
	"database.mysql.host":     "DB_HOST",
	"database.mysql.port":     "DB_PORT",
	"database.mysql.username": "DB_USER",
	"database.mysql.password": "DB_PASSWORD",
	"database.mysql.dbname":   "DB_NAME",

	// Redis相关
	"redis.host":     "REDIS_HOST",
	"redis.port":     "REDIS_PORT",
	"redis.password": "REDIS_PASSWORD",
	"redis.db":       "REDIS_DB",

	// JWT相关
	"jwt.secret": "JWT_SECRET",

	// OSS相关
	"storage.oss.provider":          "OSS_PROVIDER",
	"storage.oss.endpoint":          "OSS_ENDPOINT",
	"storage.oss.access_key_id":     "OSS_ACCESS_KEY_ID",
	"storage.oss.access_key_secret": "OSS_ACCESS_KEY_SECRET",
	"storage.oss.bucket_name":       "OSS_BUCKET",
	"storage.oss.region":            "OSS_REGION",
	"storage.oss.domain":            "OSS_DOMAIN",

	// 邮件相关
	"email.smtp.host":       "SMTP_HOST",
	"email.smtp.port":       "SMTP_PORT",
	"email.smtp.username":   "SMTP_USER",
	"email.smtp.password":   "SMTP_PASSWORD",
	"email.smtp.from_email": "SMTP_FROM_EMAIL",

	// 服务器相关
	"server.host": "SERVER_HOST",
	"server.port": "SERVER_PORT",
	"server.mode": "GIN_MODE",

	// 应用相关
	"app.env":   "GO_ENV",
	"app.debug": "DEBUG",
}

// PrintEnvHelp 打印环境变量帮助信息
func PrintEnvHelp() {
	fmt.Println("Environment Variables:")
	fmt.Println("======================")

	for configKey, envVar := range EnvironmentVariables {
		value := os.Getenv(envVar)
		if value != "" {
			fmt.Printf("%-30s = %s (from %s)\n", configKey, value, envVar)
		} else {
			fmt.Printf("%-30s = <not set> (env: %s)\n", configKey, envVar)
		}
	}

	fmt.Println("\nPrefix: CLOUDPAN_")
	fmt.Println("You can also use CLOUDPAN_ prefix with any config key.")
	fmt.Println("For example: CLOUDPAN_DATABASE_MYSQL_HOST")
}

// GetEnvExample 获取环境变量示例
func GetEnvExample() string {
	return `# HXLOS Cloud Storage Environment Variables Example
# Copy this file to .env and modify the values

# Application
GO_ENV=production
DEBUG=false

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=cloudpan
DB_PASSWORD=your_db_password
DB_NAME=cloudpan

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# JWT
JWT_SECRET=your_super_secret_jwt_key_at_least_32_characters_long

# Email SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_email_password
SMTP_FROM_EMAIL=your_email@gmail.com

# OSS (Optional)
OSS_PROVIDER=aliyun
OSS_ENDPOINT=your_oss_endpoint
OSS_ACCESS_KEY_ID=your_access_key_id
OSS_ACCESS_KEY_SECRET=your_access_key_secret
OSS_BUCKET=your_bucket_name
OSS_REGION=your_region
OSS_DOMAIN=your_domain

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
GIN_MODE=release
`
}
