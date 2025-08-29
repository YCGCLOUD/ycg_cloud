package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"cloudpan/internal/pkg/config"
)

// MockSQLDB 模拟sql.DB接口
type MockSQLDB struct {
	mock.Mock
}

func (m *MockSQLDB) SetMaxOpenConns(n int) {
	m.Called(n)
}

func (m *MockSQLDB) SetMaxIdleConns(n int) {
	m.Called(n)
}

func (m *MockSQLDB) SetConnMaxLifetime(d time.Duration) {
	m.Called(d)
}

func (m *MockSQLDB) SetConnMaxIdleTime(d time.Duration) {
	m.Called(d)
}

func (m *MockSQLDB) PingContext(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSQLDB) Stats() sql.DBStats {
	args := m.Called()
	return args.Get(0).(sql.DBStats)
}

func (m *MockSQLDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   config.MySQLConfig
		expected string
	}{
		{
			name: "基础配置",
			config: config.MySQLConfig{
				Host:      "localhost",
				Port:      3306,
				Username:  "root",
				Password:  "password",
				DBName:    "test",
				Charset:   "utf8mb4",
				ParseTime: true,
				Loc:       "Local",
			},
			expected: "root:password@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=true&loc=Local&allowNativePasswords=true",
		},
		{
			name: "包含时区配置",
			config: config.MySQLConfig{
				Host:      "127.0.0.1",
				Port:      3306,
				Username:  "user",
				Password:  "pass",
				DBName:    "mydb",
				Charset:   "utf8mb4",
				ParseTime: true,
				Loc:       "UTC",
				Timezone:  "+08:00",
			},
			expected: "user:pass@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=true&loc=UTC&allowNativePasswords=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDSN(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigureConnectionPool(t *testing.T) {
	tests := []struct {
		name   string
		config config.MySQLConfig
	}{
		{
			name: "默认配置",
			config: config.MySQLConfig{
				MaxIdleConns:    0, // 测试默认值
				MaxOpenConns:    0, // 测试默认值
				ConnMaxLifetime: 0, // 测试默认值
				ConnMaxIdleTime: 0, // 测试默认值
			},
		},
		{
			name: "自定义配置",
			config: config.MySQLConfig{
				MaxIdleConns:    20,
				MaxOpenConns:    200,
				ConnMaxLifetime: 2 * time.Hour,
				ConnMaxIdleTime: 30 * time.Minute,
			},
		},
		{
			name: "空闲连接数超过最大连接数",
			config: config.MySQLConfig{
				MaxIdleConns: 150,
				MaxOpenConns: 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockSQLDB)

			// 设置期望的调用
			expectedMaxOpen := tt.config.MaxOpenConns
			if expectedMaxOpen <= 0 {
				expectedMaxOpen = 100
			}
			mockDB.On("SetMaxOpenConns", expectedMaxOpen).Once()

			expectedMaxIdle := tt.config.MaxIdleConns
			if expectedMaxIdle <= 0 {
				expectedMaxIdle = 10
			}
			// 确保空闲连接数不超过最大连接数
			if expectedMaxIdle > expectedMaxOpen {
				expectedMaxIdle = expectedMaxOpen
			}
			mockDB.On("SetMaxIdleConns", expectedMaxIdle).Once()

			expectedLifetime := tt.config.ConnMaxLifetime
			if expectedLifetime <= 0 {
				expectedLifetime = time.Hour
			}
			mockDB.On("SetConnMaxLifetime", expectedLifetime).Once()

			expectedIdleTime := tt.config.ConnMaxIdleTime
			if expectedIdleTime <= 0 {
				expectedIdleTime = 30 * time.Minute // 修改为新的默认值
			}
			mockDB.On("SetConnMaxIdleTime", expectedIdleTime).Once()

			// 执行测试
			err := configureConnectionPool(mockDB, tt.config)

			// 验证结果
			assert.NoError(t, err)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestTestConnection(t *testing.T) {
	tests := []struct {
		name        string
		pingError   error
		expectError bool
	}{
		{
			name:        "连接成功",
			pingError:   nil,
			expectError: false,
		},
		{
			name:        "连接失败",
			pingError:   assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockSQLDB)
			mockDB.On("PingContext", mock.AnythingOfType("*context.timerCtx")).Return(tt.pingError)

			err := testConnection(mockDB)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetConnectionStats(t *testing.T) {
	// 保存原始DB实例
	originalDB := DB
	defer func() { DB = originalDB }()

	t.Run("数据库未初始化", func(t *testing.T) {
		DB = nil
		stats := GetConnectionStats()
		assert.Contains(t, stats, "error")
		assert.Equal(t, "database not initialized", stats["error"])
	})
}

func TestHealthCheck(t *testing.T) {
	// 保存原始DB实例
	originalDB := DB
	defer func() { DB = originalDB }()

	t.Run("数据库未初始化", func(t *testing.T) {
		DB = nil
		err := HealthCheck()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database not initialized")
	})
}

func TestClose(t *testing.T) {
	// 保存原始DB实例
	originalDB := DB
	defer func() { DB = originalDB }()

	t.Run("数据库未初始化", func(t *testing.T) {
		DB = nil
		err := Close()
		assert.NoError(t, err)
	})
}

// BenchmarkConnectionPool 连接池性能基准测试
func BenchmarkConnectionPool(b *testing.B) {
	// 这个基准测试需要真实的数据库连接
	// 在CI/CD环境中可能需要跳过
	if testing.Short() {
		b.Skip("跳过需要数据库连接的基准测试")
	}

	// 测试并发获取连接的性能
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if DB != nil {
				// 模拟简单的数据库操作
				var result int
				DB.Raw("SELECT 1").Scan(&result)
			}
		}
	})
}

// ExampleInitMySQL 演示如何初始化MySQL连接池
func ExampleInitMySQL() {
	// 首先需要加载配置
	// config.Load()

	// 初始化MySQL连接池
	// if err := InitMySQL(); err != nil {
	//     log.Fatal("Failed to initialize MySQL:", err)
	// }

	// 获取数据库连接
	// db := GetDB()

	// 执行数据库操作
	// var count int64
	// db.Raw("SELECT COUNT(*) FROM information_schema.tables").Scan(&count)

	// fmt.Printf("Tables count: %d\n", count)
}

// ExampleGetConnectionStats 演示如何获取连接池统计信息
func ExampleGetConnectionStats() {
	// 获取连接池统计信息
	// stats := GetConnectionStats()
	// fmt.Printf("Connection pool stats: %+v\n", stats)

	// 输出示例:
	// Connection pool stats: map[idle:5 in_use:2 max_open_connections:100 open_connections:7 ...]
}

// TestBuildDSNWithSpecialCharacters 测试包含特殊字符的DSN构建
func TestBuildDSNWithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		config   config.MySQLConfig
		expected string
	}{
		{
			name: "密码包含特殊字符",
			config: config.MySQLConfig{
				Host:      "localhost",
				Port:      3306,
				Username:  "root",
				Password:  "p@ssw0rd!@#$%",
				DBName:    "test",
				Charset:   "utf8mb4",
				ParseTime: true,
				Loc:       "Asia/Shanghai",
			},
			expected: "root:p%40ssw0rd%21%40%23%24%25@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai&allowNativePasswords=true",
		},
		{
			name: "空密码",
			config: config.MySQLConfig{
				Host:      "localhost",
				Port:      3306,
				Username:  "root",
				Password:  "",
				DBName:    "test",
				Charset:   "utf8mb4",
				ParseTime: false,
				Loc:       "Local",
			},
			expected: "root:@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=false&loc=Local&allowNativePasswords=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDSN(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCreateGormLogger 测试GORM日志器创建
func TestCreateGormLogger(t *testing.T) {
	// 保存原始配置
	originalConfig := config.AppConfig
	defer func() { config.AppConfig = originalConfig }()

	// 设置模拟配置
	config.AppConfig = &config.Config{
		App: config.App{
			Debug: false,
		},
	}

	// 测试非调试模式
	logger1 := createGormLogger()
	assert.NotNil(t, logger1)

	// 设置调试模式
	config.AppConfig.App.Debug = true
	logger2 := createGormLogger()
	assert.NotNil(t, logger2)
}

// TestConfigureConnectionPoolEdgeCases 测试连接池配置边界情况
func TestConfigureConnectionPoolEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		config config.MySQLConfig
	}{
		{
			name: "极大连接数",
			config: config.MySQLConfig{
				MaxOpenConns: 2000, // 超过安全限制
				MaxIdleConns: 50,
			},
		},
		{
			name: "极长生存时间",
			config: config.MySQLConfig{
				MaxOpenConns:    100,
				MaxIdleConns:    10,
				ConnMaxLifetime: 10 * time.Hour, // 超过安全限制
			},
		},
		{
			name: "空闲时间超过生存时间",
			config: config.MySQLConfig{
				MaxOpenConns:    100,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 2 * time.Hour, // 超过生存时间
			},
		},
		{
			name: "负值配置",
			config: config.MySQLConfig{
				MaxOpenConns:    -1,
				MaxIdleConns:    -1,
				ConnMaxLifetime: -time.Hour,
				ConnMaxIdleTime: -time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockSQLDB)

			// 设置期望的调用（会被安全处理）
			mockDB.On("SetMaxOpenConns", mock.AnythingOfType("int")).Once()
			mockDB.On("SetMaxIdleConns", mock.AnythingOfType("int")).Once()
			mockDB.On("SetConnMaxLifetime", mock.AnythingOfType("time.Duration")).Once()
			mockDB.On("SetConnMaxIdleTime", mock.AnythingOfType("time.Duration")).Once()

			err := configureConnectionPool(mockDB, tt.config)
			assert.NoError(t, err)
			mockDB.AssertExpectations(t)
		})
	}
}

// TestGetDB 测试获取数据库实例
func TestGetDB(t *testing.T) {
	// 保存原始DB实例
	originalDB := DB
	defer func() { DB = originalDB }()

	t.Run("数据库未初始化", func(t *testing.T) {
		DB = nil
		db := GetDB()
		assert.Nil(t, db)
	})

	t.Run("数据库已初始化", func(t *testing.T) {
		// 这里使用模拟的DB实例
		// 在实际测试中，这需要真实的GORM实例
		// DB = &gorm.DB{} // 无法直接创建，需要通过gorm.Open
		DB = originalDB // 使用原始实例进行测试
		if DB != nil {
			db := GetDB()
			assert.NotNil(t, db)
			assert.Equal(t, DB, db)
		}
	})
}

// TestMockSQLDBInterface 测试MockSQLDB接口实现
func TestMockSQLDBInterface(t *testing.T) {
	mockDB := new(MockSQLDB)

	// 测试所有接口方法
	mockDB.On("SetMaxOpenConns", 100).Once()
	mockDB.On("SetMaxIdleConns", 10).Once()
	mockDB.On("SetConnMaxLifetime", time.Hour).Once()
	mockDB.On("SetConnMaxIdleTime", 30*time.Minute).Once()
	mockDB.On("PingContext", mock.Anything).Return(nil).Once()
	mockDB.On("Stats").Return(sql.DBStats{}).Once()
	mockDB.On("Close").Return(nil).Once()

	// 执行所有方法
	mockDB.SetMaxOpenConns(100)
	mockDB.SetMaxIdleConns(10)
	mockDB.SetConnMaxLifetime(time.Hour)
	mockDB.SetConnMaxIdleTime(30 * time.Minute)
	err := mockDB.PingContext(context.Background())
	assert.NoError(t, err)

	stats := mockDB.Stats()
	assert.NotNil(t, stats)

	err = mockDB.Close()
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}

// TestDSNComponents 测试DSN组件
func TestDSNComponents(t *testing.T) {
	config := config.MySQLConfig{
		Host:      "db.example.com",
		Port:      3307,
		Username:  "testuser",
		Password:  "testpass",
		DBName:    "testdb",
		Charset:   "utf8",
		ParseTime: true,
		Loc:       "UTC",
	}

	dsn := buildDSN(config)

	// 验证DSN包含所有必要组件
	assert.Contains(t, dsn, "testuser:testpass")
	assert.Contains(t, dsn, "@tcp(db.example.com:3307)")
	assert.Contains(t, dsn, "/testdb")
	assert.Contains(t, dsn, "charset=utf8")
	assert.Contains(t, dsn, "parseTime=true")
	assert.Contains(t, dsn, "loc=UTC")
	assert.Contains(t, dsn, "allowNativePasswords=true")
}

// TestConnectionPoolDefaults 测试连接池默认值
func TestConnectionPoolDefaults(t *testing.T) {
	mockDB := new(MockSQLDB)
	config := config.MySQLConfig{} // 所有值都是零值

	// 验证默认值的设置
	mockDB.On("SetMaxOpenConns", 100).Once()               // 默认最大连接数
	mockDB.On("SetMaxIdleConns", 10).Once()                // 默认最大空闲连接数
	mockDB.On("SetConnMaxLifetime", time.Hour).Once()      // 默认连接生存时间
	mockDB.On("SetConnMaxIdleTime", 30*time.Minute).Once() // 默认空闲时间

	err := configureConnectionPool(mockDB, config)
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}
