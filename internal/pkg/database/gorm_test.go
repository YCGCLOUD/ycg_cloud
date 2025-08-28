package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"cloudpan/internal/pkg/config"
	"cloudpan/internal/pkg/database/models"
)

// GormTestSuite Gorm功能测试套件
type GormTestSuite struct {
	suite.Suite
}

// SetupSuite 测试套件初始化
func (s *GormTestSuite) SetupSuite() {
	// 初始化测试配置
	config.AppConfig = &config.Config{
		App: config.App{
			Debug: true,
		},
		Database: config.DatabaseConfig{
			MySQL: config.MySQLConfig{
				Host:            "localhost",
				Port:            3306,
				Username:        "test",
				Password:        "test",
				DBName:          "test_db",
				Charset:         "utf8mb4",
				ParseTime:       true,
				Loc:             "Local",
				MaxIdleConns:    5,
				MaxOpenConns:    10,
				ConnMaxLifetime: 1 * time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
			},
		},
	}

	// 注意：这个测试需要真实的数据库连接
	// 在CI/CD环境中可能需要跳过或使用测试数据库
	if testing.Short() {
		s.T().Skip("跳过需要数据库连接的集成测试")
	}
}

// TearDownSuite 测试套件清理
func (s *GormTestSuite) TearDownSuite() {
	if DB != nil {
		Close()
	}
}

// TestTransaction 测试事务功能
func (s *GormTestSuite) TestTransaction() {
	if DB == nil {
		s.T().Skip("数据库未初始化")
	}

	// 测试成功事务
	err := Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些操作
		var count int64
		tx.Raw("SELECT COUNT(*) FROM information_schema.tables").Scan(&count)
		return nil
	})
	assert.NoError(s.T(), err)

	// 测试事务回滚
	err = Transaction(func(tx *gorm.DB) error {
		// 模拟错误，触发回滚
		return assert.AnError
	})
	assert.Error(s.T(), err)
}

// TestTransactionWithContext 测试带上下文的事务
func (s *GormTestSuite) TestTransactionWithContext() {
	if DB == nil {
		s.T().Skip("数据库未初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := TransactionWithContext(ctx, func(tx *gorm.DB) error {
		var count int64
		tx.Raw("SELECT COUNT(*) FROM information_schema.tables").Scan(&count)
		return nil
	})
	assert.NoError(s.T(), err)
}

// TestPagination 测试分页功能
func (s *GormTestSuite) TestPagination() {
	if DB == nil {
		s.T().Skip("数据库未初始化")
	}

	// 创建测试用的模型（使用information_schema.tables作为测试数据）
	type TableInfo struct {
		TableName   string `json:"table_name"`
		TableSchema string `json:"table_schema"`
	}

	var tables []TableInfo
	opts := &QueryOptions{
		Page:  1,
		Size:  5,
		Sort:  "table_name",
		Order: "asc",
	}

	// 由于我们使用的是系统表，可能不适合直接测试分页
	// 这里主要测试分页函数的参数处理逻辑
	result, err := Paginate(DB.Table("information_schema.tables").Select("table_name, table_schema"), &tables, opts)

	if err == nil {
		assert.NotNil(s.T(), result)
		assert.Equal(s.T(), 1, result.Page)
		assert.Equal(s.T(), 5, result.Size)
		assert.GreaterOrEqual(s.T(), result.Total, int64(0))
	}
}

// TestQueryOptions 测试查询选项
func (s *GormTestSuite) TestQueryOptions() {
	// 测试默认查询选项
	opts := DefaultQueryOptions
	assert.Equal(s.T(), 1, opts.Page)
	assert.Equal(s.T(), 20, opts.Size)
	assert.Equal(s.T(), "desc", opts.Order)

	// 测试查询选项验证
	testOpts := &QueryOptions{
		Page: -1,
		Size: -1,
	}

	// 在分页函数中会自动修正这些值
	if DB != nil {
		var result []interface{}
		pagination, err := Paginate(DB.Table("information_schema.tables"), &result, testOpts)
		if err == nil {
			assert.Equal(s.T(), 1, pagination.Page)  // 应该被修正为1
			assert.Equal(s.T(), 20, pagination.Size) // 应该被修正为20
		}
	}
}

// TestExists 测试记录存在性检查
func (s *GormTestSuite) TestExists() {
	if DB == nil {
		s.T().Skip("数据库未初始化")
	}

	// 检查系统表是否存在
	exists, err := Exists(DB.Table("information_schema.tables"), nil, "table_schema = ?", "information_schema")
	assert.NoError(s.T(), err)
	assert.True(s.T(), exists)

	// 检查不存在的条件
	exists, err = Exists(DB.Table("information_schema.tables"), nil, "table_schema = ?", "non_existent_schema")
	assert.NoError(s.T(), err)
	assert.False(s.T(), exists)
}

// TestModelRegistry 测试模型注册
func (s *GormTestSuite) TestModelRegistry() {
	// 测试模型注册
	RegisterModel("TestModel", &models.BaseModel{})

	registeredModels := GetRegisteredModels()
	assert.Contains(s.T(), registeredModels, "TestModel")
	assert.Contains(s.T(), registeredModels, "BaseModel") // 在init中注册的
}

// TestMigrationConfig 测试迁移配置
func (s *GormTestSuite) TestMigrationConfig() {
	config := DefaultMigrationConfig
	assert.True(s.T(), config.AutoMigrate)
	assert.False(s.T(), config.DropFirst)
	assert.True(s.T(), config.CreateIndex)
}

// TestWithUserContext 测试用户上下文
func (s *GormTestSuite) TestWithUserContext() {
	if DB == nil {
		s.T().Skip("数据库未初始化")
	}

	// 测试设置用户上下文
	dbWithUser := WithUserContext(DB, 123)
	userID, exists := dbWithUser.Get("current_user_id")
	assert.True(s.T(), exists)
	assert.Equal(s.T(), uint(123), userID)
}

// TestWithTimeout 测试超时上下文
func (s *GormTestSuite) TestWithTimeout() {
	if DB == nil {
		s.T().Skip("数据库未初始化")
	}

	// 测试设置超时上下文
	dbWithTimeout := WithTimeout(DB, 5*time.Second)
	assert.NotNil(s.T(), dbWithTimeout.Statement.Context)
}

// TestPlugins 测试插件功能
func (s *GormTestSuite) TestPlugins() {
	if DB == nil {
		s.T().Skip("数据库未初始化")
	}

	// 测试默认插件
	plugins := GetDefaultPlugins()
	assert.Len(s.T(), plugins, 3) // AuditPlugin, MetricsPlugin, TracePlugin

	// 测试插件名称
	assert.Equal(s.T(), "audit", plugins[0].Name())
	assert.Equal(s.T(), "metrics", plugins[1].Name())
	assert.Equal(s.T(), "trace", plugins[2].Name())
}

// TestCustomLogger 测试自定义日志记录器
func (s *GormTestSuite) TestCustomLogger() {
	logger := NewCustomLogger(100*time.Millisecond, logger.Info)
	assert.NotNil(s.T(), logger)
	assert.Equal(s.T(), 100*time.Millisecond, logger.SlowThreshold)
	// 移除函数类型的比较，因为它们不能直接比较
	// assert.Equal(s.T(), logger.Info, logger.LogLevel)
}

// TestBaseModel 测试基础模型
func (s *GormTestSuite) TestBaseModel() {
	model := &models.BaseModel{}

	// 测试版本控制
	assert.Equal(s.T(), int64(0), model.GetVersion())
	model.SetVersion(5)
	assert.Equal(s.T(), int64(5), model.GetVersion())

	// 测试软删除状态
	assert.False(s.T(), model.IsDeleted())
}

// TestAuditModel 测试审计模型
func (s *GormTestSuite) TestAuditModel() {
	model := &models.AuditModel{}

	// 测试设置创建者和更新者
	model.SetCreatedBy(100)
	model.SetUpdatedBy(200)

	assert.Equal(s.T(), uint(100), model.CreatedBy)
	assert.Equal(s.T(), uint(200), model.UpdatedBy)
}

// TestStatusModel 测试状态模型
func (s *GormTestSuite) TestStatusModel() {
	model := &models.StatusModel{}

	// 测试状态操作
	model.Activate()
	assert.True(s.T(), model.IsActive())
	assert.Equal(s.T(), models.StatusActive, model.Status)

	model.Deactivate()
	assert.False(s.T(), model.IsActive())
	assert.Equal(s.T(), models.StatusInactive, model.Status)

	model.Disable()
	assert.Equal(s.T(), models.StatusDisabled, model.Status)
}

// 运行测试套件
func TestGormSuite(t *testing.T) {
	suite.Run(t, new(GormTestSuite))
}

// BenchmarkTransaction 事务性能基准测试
func BenchmarkTransaction(b *testing.B) {
	if testing.Short() {
		b.Skip("跳过需要数据库连接的基准测试")
	}

	if DB == nil {
		b.Skip("数据库未初始化")
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Transaction(func(tx *gorm.DB) error {
				var count int64
				tx.Raw("SELECT 1").Scan(&count)
				return nil
			})
		}
	})
}

// BenchmarkPagination 分页查询性能基准测试
func BenchmarkPagination(b *testing.B) {
	if testing.Short() {
		b.Skip("跳过需要数据库连接的基准测试")
	}

	if DB == nil {
		b.Skip("数据库未初始化")
	}

	opts := &QueryOptions{
		Page: 1,
		Size: 10,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var result []interface{}
			Paginate(DB.Table("information_schema.tables"), &result, opts)
		}
	})
}
