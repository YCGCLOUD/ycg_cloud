package config

import "time"

// Config 应用配置结构体
type Config struct {
	App        App              `yaml:"app" mapstructure:"app"`
	Server     ServerConfig     `yaml:"server" mapstructure:"server"`
	Database   DatabaseConfig   `yaml:"database" mapstructure:"database"`
	Redis      RedisConfig      `yaml:"redis" mapstructure:"redis"`
	JWT        JWTConfig        `yaml:"jwt" mapstructure:"jwt"`
	Storage    StorageConfig    `yaml:"storage" mapstructure:"storage"`
	User       UserConfig       `yaml:"user" mapstructure:"user"`
	Email      EmailConfig      `yaml:"email" mapstructure:"email"`
	Security   SecurityConfig   `yaml:"security" mapstructure:"security"`
	Log        LogConfig        `yaml:"log" mapstructure:"log"`
	Cache      CacheConfig      `yaml:"cache" mapstructure:"cache"`
	Queue      QueueConfig      `yaml:"queue" mapstructure:"queue"`
	WebSocket  WebSocketConfig  `yaml:"websocket" mapstructure:"websocket"`
	Monitoring MonitoringConfig `yaml:"monitoring" mapstructure:"monitoring"`
	I18n       I18nConfig       `yaml:"i18n" mapstructure:"i18n"`
	ThirdParty ThirdPartyConfig `yaml:"third_party" mapstructure:"third_party"`
}

// App 应用配置
type App struct {
	Name    string `yaml:"name" mapstructure:"name"`
	Version string `yaml:"version" mapstructure:"version"`
	Env     string `yaml:"env" mapstructure:"env"`
	Debug   bool   `yaml:"debug" mapstructure:"debug"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host           string        `yaml:"host" mapstructure:"host"`
	Port           int           `yaml:"port" mapstructure:"port"`
	Mode           string        `yaml:"mode" mapstructure:"mode"`
	ReadTimeout    time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes" mapstructure:"max_header_bytes"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL MySQLConfig `yaml:"mysql" mapstructure:"mysql"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host            string        `yaml:"host" mapstructure:"host"`
	Port            int           `yaml:"port" mapstructure:"port"`
	Username        string        `yaml:"username" mapstructure:"username"`
	Password        string        `yaml:"password" mapstructure:"password"`
	DBName          string        `yaml:"dbname" mapstructure:"dbname"`
	Charset         string        `yaml:"charset" mapstructure:"charset"`
	ParseTime       bool          `yaml:"parse_time" mapstructure:"parse_time"`
	Loc             string        `yaml:"loc" mapstructure:"loc"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns" mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" mapstructure:"conn_max_idle_time"`
	Timezone        string        `yaml:"timezone" mapstructure:"timezone"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string        `yaml:"host" mapstructure:"host"`
	Port         int           `yaml:"port" mapstructure:"port"`
	Password     string        `yaml:"password" mapstructure:"password"`
	DB           int           `yaml:"db" mapstructure:"db"`
	Protocol     int           `yaml:"protocol" mapstructure:"protocol"`
	PoolSize     int           `yaml:"pool_size" mapstructure:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns" mapstructure:"min_idle_conns"`
	MaxRetries   int           `yaml:"max_retries" mapstructure:"max_retries"`
	DialTimeout  time.Duration `yaml:"dial_timeout" mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	PoolTimeout  time.Duration `yaml:"pool_timeout" mapstructure:"pool_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" mapstructure:"idle_timeout"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret             string `yaml:"secret" mapstructure:"secret"`
	ExpireHours        int    `yaml:"expire_hours" mapstructure:"expire_hours"`
	RefreshExpireHours int    `yaml:"refresh_expire_hours" mapstructure:"refresh_expire_hours"`
	Issuer             string `yaml:"issuer" mapstructure:"issuer"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Local LocalStorageConfig `yaml:"local" mapstructure:"local"`
	OSS   OSSStorageConfig   `yaml:"oss" mapstructure:"oss"`
}

// LocalStorageConfig 本地存储配置
type LocalStorageConfig struct {
	Enabled      bool     `yaml:"enabled" mapstructure:"enabled"`
	RootPath     string   `yaml:"root_path" mapstructure:"root_path"`
	TempPath     string   `yaml:"temp_path" mapstructure:"temp_path"`
	MaxSize      int64    `yaml:"max_size" mapstructure:"max_size"`
	AllowedTypes []string `yaml:"allowed_types" mapstructure:"allowed_types"`
}

// OSSStorageConfig OSS存储配置
type OSSStorageConfig struct {
	Enabled         bool   `yaml:"enabled" mapstructure:"enabled"`
	Provider        string `yaml:"provider" mapstructure:"provider"`
	Endpoint        string `yaml:"endpoint" mapstructure:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id" mapstructure:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret" mapstructure:"access_key_secret"`
	BucketName      string `yaml:"bucket_name" mapstructure:"bucket_name"`
	Region          string `yaml:"region" mapstructure:"region"`
	Domain          string `yaml:"domain" mapstructure:"domain"`
	Secure          bool   `yaml:"secure" mapstructure:"secure"`
	AutoSwitchSize  int64  `yaml:"auto_switch_size" mapstructure:"auto_switch_size"`
}

// UserConfig 用户配置
type UserConfig struct {
	DefaultQuota int64          `yaml:"default_quota" mapstructure:"default_quota"`
	MaxQuota     int64          `yaml:"max_quota" mapstructure:"max_quota"`
	Avatar       AvatarConfig   `yaml:"avatar" mapstructure:"avatar"`
	Password     PasswordConfig `yaml:"password" mapstructure:"password"`
}

// AvatarConfig 头像配置
type AvatarConfig struct {
	MaxSize          int64    `yaml:"max_size" mapstructure:"max_size"`
	AllowedTypes     []string `yaml:"allowed_types" mapstructure:"allowed_types"`
	PathTemplate     string   `yaml:"path_template" mapstructure:"path_template"`
	FilenameTemplate string   `yaml:"filename_template" mapstructure:"filename_template"`
}

// PasswordConfig 密码配置
type PasswordConfig struct {
	MinLength      int  `yaml:"min_length" mapstructure:"min_length"`
	MaxLength      int  `yaml:"max_length" mapstructure:"max_length"`
	RequireNumber  bool `yaml:"require_number" mapstructure:"require_number"`
	RequireLetter  bool `yaml:"require_letter" mapstructure:"require_letter"`
	RequireSpecial bool `yaml:"require_special" mapstructure:"require_special"`
	BcryptCost     int  `yaml:"bcrypt_cost" mapstructure:"bcrypt_cost"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTP       SMTPConfig       `yaml:"smtp" mapstructure:"smtp"`
	Templates  TemplatesConfig  `yaml:"templates" mapstructure:"templates"`
	VerifyCode VerifyCodeConfig `yaml:"verify_code" mapstructure:"verify_code"`
}

// SMTPConfig SMTP配置
type SMTPConfig struct {
	Host      string `yaml:"host" mapstructure:"host"`
	Port      int    `yaml:"port" mapstructure:"port"`
	Username  string `yaml:"username" mapstructure:"username"`
	Password  string `yaml:"password" mapstructure:"password"`
	FromName  string `yaml:"from_name" mapstructure:"from_name"`
	FromEmail string `yaml:"from_email" mapstructure:"from_email"`
}

// TemplatesConfig 邮件模板配置
type TemplatesConfig struct {
	VerifyCode    string `yaml:"verify_code" mapstructure:"verify_code"`
	PasswordReset string `yaml:"password_reset" mapstructure:"password_reset"`
	Welcome       string `yaml:"welcome" mapstructure:"welcome"`
}

// VerifyCodeConfig 验证码配置
type VerifyCodeConfig struct {
	Length        int `yaml:"length" mapstructure:"length"`
	ExpireMinutes int `yaml:"expire_minutes" mapstructure:"expire_minutes"`
	MaxAttempts   int `yaml:"max_attempts" mapstructure:"max_attempts"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	CORS      CORSConfig      `yaml:"cors" mapstructure:"cors"`
	RateLimit RateLimitConfig `yaml:"rate_limit" mapstructure:"rate_limit"`
	Antivirus AntivirusConfig `yaml:"antivirus" mapstructure:"antivirus"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins" mapstructure:"allow_origins"`
	AllowMethods     []string `yaml:"allow_methods" mapstructure:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers" mapstructure:"allow_headers"`
	ExposeHeaders    []string `yaml:"expose_headers" mapstructure:"expose_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" mapstructure:"allow_credentials"`
	MaxAge           int      `yaml:"max_age" mapstructure:"max_age"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled" mapstructure:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute" mapstructure:"requests_per_minute"`
	Burst             int  `yaml:"burst" mapstructure:"burst"`
}

// AntivirusConfig 病毒扫描配置
type AntivirusConfig struct {
	Enabled      bool          `yaml:"enabled" mapstructure:"enabled"`
	ClamAVSocket string        `yaml:"clamav_socket" mapstructure:"clamav_socket"`
	ScanTimeout  time.Duration `yaml:"scan_timeout" mapstructure:"scan_timeout"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string          `yaml:"level" mapstructure:"level"`
	Format     string          `yaml:"format" mapstructure:"format"`
	Output     string          `yaml:"output" mapstructure:"output"`
	FilePath   string          `yaml:"file_path" mapstructure:"file_path"`
	MaxSize    int             `yaml:"max_size" mapstructure:"max_size"`
	MaxAge     int             `yaml:"max_age" mapstructure:"max_age"`
	MaxBackups int             `yaml:"max_backups" mapstructure:"max_backups"`
	Compress   bool            `yaml:"compress" mapstructure:"compress"`
	AccessLog  AccessLogConfig `yaml:"access_log" mapstructure:"access_log"`
}

// AccessLogConfig 访问日志配置
type AccessLogConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	FilePath string `yaml:"file_path" mapstructure:"file_path"`
	Format   string `yaml:"format" mapstructure:"format"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	DefaultTTL          time.Duration `yaml:"default_ttl" mapstructure:"default_ttl"`
	UserInfoTTL         time.Duration `yaml:"user_info_ttl" mapstructure:"user_info_ttl"`
	FileInfoTTL         time.Duration `yaml:"file_info_ttl" mapstructure:"file_info_ttl"`
	VerificationCodeTTL time.Duration `yaml:"verification_code_ttl" mapstructure:"verification_code_ttl"`
}

// QueueConfig 消息队列配置
type QueueConfig struct {
	RedisStream RedisStreamConfig `yaml:"redis_stream" mapstructure:"redis_stream"`
}

// RedisStreamConfig Redis Stream配置
type RedisStreamConfig struct {
	Enabled       bool              `yaml:"enabled" mapstructure:"enabled"`
	Streams       map[string]string `yaml:"streams" mapstructure:"streams"`
	ConsumerGroup string            `yaml:"consumer_group" mapstructure:"consumer_group"`
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	Enabled          bool          `yaml:"enabled" mapstructure:"enabled"`
	Path             string        `yaml:"path" mapstructure:"path"`
	CheckOrigin      bool          `yaml:"check_origin" mapstructure:"check_origin"`
	ReadBufferSize   int           `yaml:"read_buffer_size" mapstructure:"read_buffer_size"`
	WriteBufferSize  int           `yaml:"write_buffer_size" mapstructure:"write_buffer_size"`
	HandshakeTimeout time.Duration `yaml:"handshake_timeout" mapstructure:"handshake_timeout"`
	ReadDeadline     time.Duration `yaml:"read_deadline" mapstructure:"read_deadline"`
	WriteDeadline    time.Duration `yaml:"write_deadline" mapstructure:"write_deadline"`
	PingPeriod       time.Duration `yaml:"ping_period" mapstructure:"ping_period"`
	PongWait         time.Duration `yaml:"pong_wait" mapstructure:"pong_wait"`
	MaxMessageSize   int64         `yaml:"max_message_size" mapstructure:"max_message_size"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Metrics MetricsConfig `yaml:"metrics" mapstructure:"metrics"`
	Health  HealthConfig  `yaml:"health" mapstructure:"health"`
	PProf   PProfConfig   `yaml:"pprof" mapstructure:"pprof"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Path    string `yaml:"path" mapstructure:"path"`
}

// HealthConfig 健康检查配置
type HealthConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Path    string `yaml:"path" mapstructure:"path"`
}

// PProfConfig 性能分析配置
type PProfConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Path    string `yaml:"path" mapstructure:"path"`
}

// I18nConfig 国际化配置
type I18nConfig struct {
	DefaultLanguage string   `yaml:"default_language" mapstructure:"default_language"`
	Languages       []string `yaml:"languages" mapstructure:"languages"`
	Path            string   `yaml:"path" mapstructure:"path"`
}

// ThirdPartyConfig 第三方服务配置
type ThirdPartyConfig struct {
	SMS SMSConfig `yaml:"sms" mapstructure:"sms"`
	Geo GeoConfig `yaml:"geo" mapstructure:"geo"`
}

// SMSConfig 短信服务配置
type SMSConfig struct {
	Enabled   bool   `yaml:"enabled" mapstructure:"enabled"`
	Provider  string `yaml:"provider" mapstructure:"provider"`
	AppID     string `yaml:"app_id" mapstructure:"app_id"`
	AppSecret string `yaml:"app_secret" mapstructure:"app_secret"`
}

// GeoConfig 地理位置服务配置
type GeoConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	Provider string `yaml:"provider" mapstructure:"provider"`
	APIKey   string `yaml:"api_key" mapstructure:"api_key"`
}
