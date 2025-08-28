package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// 在测试环境中跳过需要配置文件的测试
	t.Skip("跳过需要配置文件的测试")

	// 设置测试环境
	os.Setenv("GO_ENV", "testing")
	defer os.Unsetenv("GO_ENV")

	// 测试加载配置
	err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置是否正确加载
	if AppConfig == nil {
		t.Fatal("AppConfig is nil after loading")
	}

	// 验证基本配置
	if AppConfig.App.Name == "" {
		t.Error("App name should not be empty")
	}

	if AppConfig.Server.Port <= 0 {
		t.Error("Server port should be positive")
	}
}

func TestValidateConfig(t *testing.T) {
	// 创建一个有效的配置
	cfg := &Config{
		App: App{
			Name: "test-app",
		},
		Server: ServerConfig{
			Port: 8080,
		},
		Database: DatabaseConfig{
			MySQL: MySQLConfig{
				Host:     "localhost",
				Username: "test",
				DBName:   "test_db",
			},
		},
		Redis: RedisConfig{
			Host: "localhost",
		},
		JWT: JWTConfig{
			Secret: "this_is_a_very_long_secret_key_for_testing_purposes",
		},
		Storage: StorageConfig{
			Local: LocalStorageConfig{
				Enabled:  true,
				RootPath: "/tmp/test-storage",
			},
		},
		Email: EmailConfig{
			SMTP: SMTPConfig{
				Host:      "smtp.test.com",
				FromEmail: "test@test.com",
			},
		},
	}

	err := validateConfig(cfg)
	if err != nil {
		t.Errorf("Valid config should not return error: %v", err)
	}

	// 测试无效配置
	invalidCfg := &Config{}
	err = validateConfig(invalidCfg)
	if err == nil {
		t.Error("Invalid config should return error")
	}
}

func TestConfigHelper(t *testing.T) {
	cfg := createTestConfig()
	helper := NewConfigHelper(cfg)

	// 运行各个子测试
	testStoragePath(t, helper)
	testAvatarType(t, helper)
	testOSSUsage(t, helper)
	testPasswordValidation(t, helper)
	testCacheTTL(t, helper)
	testJWTExpiration(t, helper)
}

// createTestConfig 创建测试配置
func createTestConfig() *Config {
	return &Config{
		Storage: StorageConfig{
			Local: LocalStorageConfig{
				RootPath: "/test/storage",
			},
			OSS: OSSStorageConfig{
				Enabled:        true,
				AutoSwitchSize: 1024 * 1024, // 1MB
			},
		},
		User: UserConfig{
			Avatar: AvatarConfig{
				PathTemplate:     "/storage/user-{user_id}/avatars/",
				FilenameTemplate: "avatar_{timestamp}.{ext}",
				AllowedTypes:     []string{"image/jpeg", "image/png"},
			},
			Password: PasswordConfig{
				MinLength:     8,
				MaxLength:     32,
				RequireNumber: true,
				RequireLetter: true,
			},
		},
		Cache: CacheConfig{
			DefaultTTL:  time.Hour,
			UserInfoTTL: 30 * time.Minute,
		},
		JWT: JWTConfig{
			ExpireHours: 24,
		},
	}
}

// testStoragePath 测试存储路径生成
func testStoragePath(t *testing.T, helper *ConfigHelper) {
	storagePath := helper.GetStoragePath(123)
	expected := "/test/storage/user-123"
	if storagePath != expected {
		t.Errorf("Expected storage path %s, got %s", expected, storagePath)
	}
}

// testAvatarType 测试头像类型检查
func testAvatarType(t *testing.T, helper *ConfigHelper) {
	if !helper.IsAllowedAvatarType("image/jpeg") {
		t.Error("image/jpeg should be allowed for avatar")
	}

	if helper.IsAllowedAvatarType("text/plain") {
		t.Error("text/plain should not be allowed for avatar")
	}
}

// testOSSUsage 测试OSS使用判断
func testOSSUsage(t *testing.T, helper *ConfigHelper) {
	if !helper.ShouldUseOSS(2 * 1024 * 1024) { // 2MB
		t.Error("Should use OSS for files larger than 1MB")
	}

	if helper.ShouldUseOSS(512 * 1024) { // 512KB
		t.Error("Should not use OSS for files smaller than 1MB")
	}
}

// testPasswordValidation 测试密码验证
func testPasswordValidation(t *testing.T, helper *ConfigHelper) {
	if err := helper.IsPasswordValid("test"); err == nil {
		t.Error("Short password should be invalid")
	}

	if err := helper.IsPasswordValid("testpassword"); err == nil {
		t.Error("Password without number should be invalid")
	}

	if err := helper.IsPasswordValid("testpassword123"); err != nil {
		t.Errorf("Valid password should be accepted: %v", err)
	}
}

// testCacheTTL 测试缓存TTL
func testCacheTTL(t *testing.T, helper *ConfigHelper) {
	ttl := helper.GetCacheTTL("user_info")
	if ttl != 30*time.Minute {
		t.Errorf("Expected user_info TTL to be 30 minutes, got %v", ttl)
	}

	defaultTTL := helper.GetCacheTTL("unknown")
	if defaultTTL != time.Hour {
		t.Errorf("Expected default TTL to be 1 hour, got %v", defaultTTL)
	}
}

// testJWTExpiration 测试JWT过期时间
func testJWTExpiration(t *testing.T, helper *ConfigHelper) {
	jwtExp := helper.GetJWTExpiration()
	if jwtExp != 24*time.Hour {
		t.Errorf("Expected JWT expiration to be 24 hours, got %v", jwtExp)
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// 测试环境变量映射
	if len(EnvironmentVariables) == 0 {
		t.Error("Environment variables map should not be empty")
	}

	// 检查一些重要的映射
	expectedMappings := map[string]string{
		"database.mysql.host": "DB_HOST",
		"jwt.secret":          "JWT_SECRET",
		"redis.host":          "REDIS_HOST",
	}

	for configKey, expectedEnv := range expectedMappings {
		if actualEnv, exists := EnvironmentVariables[configKey]; !exists || actualEnv != expectedEnv {
			t.Errorf("Expected %s to map to %s, got %s", configKey, expectedEnv, actualEnv)
		}
	}
}

func TestGetDSN(t *testing.T) {
	// 设置测试配置
	AppConfig = &Config{
		Database: DatabaseConfig{
			MySQL: MySQLConfig{
				Username:  "testuser",
				Password:  "testpass",
				Host:      "localhost",
				Port:      3306,
				DBName:    "testdb",
				Charset:   "utf8mb4",
				ParseTime: true,
				Loc:       "Local",
			},
		},
	}

	dsn := GetDSN()
	expected := "testuser:testpass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=true&loc=Local&allowNativePasswords=true"
	if dsn != expected {
		t.Errorf("Expected DSN %s, got %s", expected, dsn)
	}
}

func TestGetRedisAddr(t *testing.T) {
	AppConfig = &Config{
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	addr := GetRedisAddr()
	expected := "localhost:6379"
	if addr != expected {
		t.Errorf("Expected Redis address %s, got %s", expected, addr)
	}
}

func TestGetServerAddr(t *testing.T) {
	AppConfig = &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
	}

	addr := GetServerAddr()
	expected := "0.0.0.0:8080"
	if addr != expected {
		t.Errorf("Expected server address %s, got %s", expected, addr)
	}
}

func TestEnvironmentCheck(t *testing.T) {
	// 测试环境检查函数
	AppConfig = &Config{
		App: App{
			Env: "development",
		},
	}

	if !IsDevelopment() {
		t.Error("Should be development environment")
	}

	if IsProduction() {
		t.Error("Should not be production environment")
	}

	if IsTesting() {
		t.Error("Should not be testing environment")
	}

	// 切换到生产环境
	AppConfig.App.Env = "production"

	if IsProduction() == false {
		t.Error("Should be production environment")
	}

	if IsDevelopment() {
		t.Error("Should not be development environment")
	}
}
