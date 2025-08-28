package database

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"cloudpan/internal/pkg/config"
)

// LockTestSuite 锁机制测试套件
type LockTestSuite struct {
	suite.Suite
	redisClient *redis.Client
	lockManager *RedisLockManager
}

// SetupSuite 测试套件初始化
func (s *LockTestSuite) SetupSuite() {
	// 初始化测试配置
	config.AppConfig = &config.Config{
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       1, // 使用测试专用DB
		},
	}

	// 在CI/CD环境中可能需要跳过Redis测试
	if testing.Short() {
		s.T().Skip("跳过需要Redis连接的集成测试")
	}

	// 创建Redis客户端
	s.redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.redisClient.Ping(ctx).Err(); err != nil {
		s.T().Skip("Redis服务不可用，跳过测试")
		return
	}

	s.lockManager = NewRedisLockManager(s.redisClient)
}

// TearDownSuite 测试套件清理
func (s *LockTestSuite) TearDownSuite() {
	if s.redisClient != nil {
		s.redisClient.Close()
	}
}

// SetupTest 每个测试前的清理
func (s *LockTestSuite) SetupTest() {
	if s.redisClient != nil {
		// 清理测试数据
		ctx := context.Background()
		s.redisClient.FlushDB(ctx)
	}
}

// TestRedisDistributedLock 测试Redis分布式锁基本功能
func (s *LockTestSuite) TestRedisDistributedLock() {
	if s.lockManager == nil {
		s.T().Skip("Redis不可用")
	}

	ctx := context.Background()

	// 创建锁
	lock, err := s.lockManager.NewLock("test-resource", 10*time.Second)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), lock)

	// 测试获取锁
	acquired, err := lock.TryLock(ctx)
	assert.NoError(s.T(), err)
	assert.True(s.T(), acquired)

	// 测试锁状态
	isLocked, err := lock.IsLocked(ctx)
	assert.NoError(s.T(), err)
	assert.True(s.T(), isLocked)

	// 测试释放锁
	err = lock.Unlock(ctx)
	assert.NoError(s.T(), err)

	// 验证锁已释放
	isLocked, err = lock.IsLocked(ctx)
	assert.NoError(s.T(), err)
	assert.False(s.T(), isLocked)
}

// TestRedisDistributedLockConflict 测试分布式锁冲突
func (s *LockTestSuite) TestRedisDistributedLockConflict() {
	if s.lockManager == nil {
		s.T().Skip("Redis不可用")
	}

	ctx := context.Background()

	// 创建两个相同资源的锁
	lock1, err := s.lockManager.NewLock("test-resource", 10*time.Second)
	assert.NoError(s.T(), err)

	lock2, err := s.lockManager.NewLock("test-resource", 10*time.Second)
	assert.NoError(s.T(), err)

	// 第一个锁应该成功获取
	acquired1, err := lock1.TryLock(ctx)
	assert.NoError(s.T(), err)
	assert.True(s.T(), acquired1)

	// 第二个锁应该获取失败
	acquired2, err := lock2.TryLock(ctx)
	assert.NoError(s.T(), err)
	assert.False(s.T(), acquired2)

	// 释放第一个锁
	err = lock1.Unlock(ctx)
	assert.NoError(s.T(), err)

	// 现在第二个锁应该能获取成功
	acquired2, err = lock2.TryLock(ctx)
	assert.NoError(s.T(), err)
	assert.True(s.T(), acquired2)

	// 清理
	lock2.Unlock(ctx)
}

// TestRedisDistributedLockExtend 测试锁延期
func (s *LockTestSuite) TestRedisDistributedLockExtend() {
	if s.lockManager == nil {
		s.T().Skip("Redis不可用")
	}

	ctx := context.Background()

	// 创建短TTL的锁
	lock, err := s.lockManager.NewLock("test-resource", 2*time.Second)
	assert.NoError(s.T(), err)

	// 获取锁
	acquired, err := lock.TryLock(ctx)
	assert.NoError(s.T(), err)
	assert.True(s.T(), acquired)

	// 等待1秒
	time.Sleep(1 * time.Second)

	// 延长锁时间
	err = lock.Extend(ctx, 5*time.Second)
	assert.NoError(s.T(), err)

	// 再等待2秒（原来的锁应该已过期，但延期后应该仍有效）
	time.Sleep(2 * time.Second)

	// 验证锁仍然有效
	isLocked, err := lock.IsLocked(ctx)
	assert.NoError(s.T(), err)
	assert.True(s.T(), isLocked)

	// 清理
	lock.Unlock(ctx)
}

// TestRedisDistributedLockAutoExpire 测试锁自动过期
func (s *LockTestSuite) TestRedisDistributedLockAutoExpire() {
	if s.lockManager == nil {
		s.T().Skip("Redis不可用")
	}

	ctx := context.Background()

	// 创建短TTL的锁
	lock, err := s.lockManager.NewLock("test-resource", 1*time.Second)
	assert.NoError(s.T(), err)

	// 获取锁
	acquired, err := lock.TryLock(ctx)
	assert.NoError(s.T(), err)
	assert.True(s.T(), acquired)

	// 等待锁过期
	time.Sleep(2 * time.Second)

	// 验证锁已过期
	isLocked, err := lock.IsLocked(ctx)
	assert.NoError(s.T(), err)
	assert.False(s.T(), isLocked)

	// 其他锁现在应该能获取成功
	lock2, err := s.lockManager.NewLock("test-resource", 10*time.Second)
	assert.NoError(s.T(), err)

	acquired2, err := lock2.TryLock(ctx)
	assert.NoError(s.T(), err)
	assert.True(s.T(), acquired2)

	// 清理
	lock2.Unlock(ctx)
}

// TestTransactionManager 测试事务管理器
func (s *LockTestSuite) TestTransactionManager() {
	if DB == nil {
		s.T().Skip("数据库未初始化")
	}

	tm := NewTransactionManager(DB)
	assert.NotNil(s.T(), tm)

	// 测试默认事务
	err := tm.db.Transaction(func(tx *gorm.DB) error {
		// 简单的事务操作
		var count int64
		tx.Raw("SELECT COUNT(*) FROM information_schema.tables").Scan(&count)
		return nil
	})
	assert.NoError(s.T(), err)
}

// TestIsolationLevel 测试隔离级别设置
func (s *LockTestSuite) TestIsolationLevel() {
	if DB == nil {
		s.T().Skip("数据库未初始化")
	}

	tm := NewTransactionManager(DB)

	// 测试设置不同隔离级别
	testLevels := []IsolationLevel{
		ReadCommitted,
		RepeatableRead,
		Serializable,
	}

	for _, level := range testLevels {
		err := tm.WithIsolationLevel(level, func(tx *gorm.DB) error {
			// 验证能正常执行查询
			var count int64
			tx.Raw("SELECT COUNT(*) FROM information_schema.tables").Scan(&count)
			return nil
		})
		assert.NoError(s.T(), err, "Failed at isolation level: %s", level.String())
	}
}

// TestOptimisticLocking 测试乐观锁
func (s *LockTestSuite) TestOptimisticLocking() {
	if DB == nil {
		s.T().Skip("数据库未初始化")
	}

	dlm := NewDatabaseLockManager(DB)

	// 创建一个模拟的模型用于测试
	type TestModel struct {
		ID      uint `gorm:"primarykey"`
		Name    string
		Version int64 `gorm:"default:1"`
	}

	// 模拟乐观锁更新
	updates := map[string]interface{}{
		"name": "updated_name",
	}

	// 测试版本冲突检测逻辑
	err := dlm.OptimisticLockUpdate(DB, &TestModel{}, 1, updates)
	// 由于没有实际的记录，应该返回乐观锁冲突错误
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "optimistic lock conflict")
}

// TestLockTypeString 测试锁类型字符串转换
func (s *LockTestSuite) TestLockTypeString() {
	assert.Equal(s.T(), "LOCK IN SHARE MODE", SharedLock.String())
	assert.Equal(s.T(), "FOR UPDATE", ExclusiveLock.String())
}

// TestIsolationLevelString 测试隔离级别字符串转换
func (s *LockTestSuite) TestIsolationLevelString() {
	assert.Equal(s.T(), "READ UNCOMMITTED", ReadUncommitted.String())
	assert.Equal(s.T(), "READ COMMITTED", ReadCommitted.String())
	assert.Equal(s.T(), "REPEATABLE READ", RepeatableRead.String())
	assert.Equal(s.T(), "SERIALIZABLE", Serializable.String())
}

// TestGenerateRandomValue 测试随机值生成
func (s *LockTestSuite) TestGenerateRandomValue() {
	value1, err := generateRandomValue(32)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), value1, 32)

	value2, err := generateRandomValue(32)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), value2, 32)

	// 两次生成的值应该不同
	assert.NotEqual(s.T(), value1, value2)
}

// TestConcurrencyControlManager 测试并发控制管理器
func (s *LockTestSuite) TestConcurrencyControlManager() {
	if s.redisClient == nil || DB == nil {
		s.T().Skip("Redis或数据库不可用")
	}

	ccm := NewConcurrencyControlManager(DB, s.redisClient)
	assert.NotNil(s.T(), ccm)
	assert.NotNil(s.T(), ccm.txManager)
	assert.NotNil(s.T(), ccm.dbLockMgr)
	assert.NotNil(s.T(), ccm.redisLockMgr)
}

// TestDistributedLockWithFunction 测试使用分布式锁执行函数
func (s *LockTestSuite) TestDistributedLockWithFunction() {
	if s.redisClient == nil {
		s.T().Skip("Redis不可用")
	}

	ccm := NewConcurrencyControlManager(DB, s.redisClient)
	ctx := context.Background()

	executed := false
	err := ccm.WithDistributedLock(ctx, "test-function", 5*time.Second, func() error {
		executed = true
		return nil
	})

	assert.NoError(s.T(), err)
	assert.True(s.T(), executed)
}

// 运行测试套件
func TestLockSuite(t *testing.T) {
	suite.Run(t, new(LockTestSuite))
}

// BenchmarkDistributedLock 分布式锁性能基准测试
func BenchmarkDistributedLock(b *testing.B) {
	if testing.Short() {
		b.Skip("跳过需要Redis的基准测试")
	}

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Skip("Redis不可用")
	}

	lockManager := NewRedisLockManager(client)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lock, err := lockManager.NewLock("benchmark-test", time.Second)
			if err != nil {
				b.Error(err)
				continue
			}

			acquired, err := lock.TryLock(ctx)
			if err != nil {
				b.Error(err)
				continue
			}

			if acquired {
				lock.Unlock(ctx)
			}
		}
	})
}
