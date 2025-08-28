package cache

import (
	"fmt"
	"testing"
	"time"

	"cloudpan/internal/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// CacheTestSuite Redis缓存测试套件
type CacheTestSuite struct {
	suite.Suite
	manager    *CacheManager
	ttlManager *TTLManager
	wrapper    *CacheWrapper
}

// SetupSuite 测试套件初始化
func (s *CacheTestSuite) SetupSuite() {
	// 初始化测试配置
	config.AppConfig = &config.Config{
		Redis: config.RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Password:     "",
			DB:           1, // 使用测试数据库
			PoolSize:     5,
			MinIdleConns: 2,
		},
		Cache: config.CacheConfig{
			DefaultTTL:          time.Hour,
			UserInfoTTL:         30 * time.Minute,
			FileInfoTTL:         10 * time.Minute,
			VerificationCodeTTL: 5 * time.Minute,
		},
	}
}

// SetupTest 每个测试前的准备
func (s *CacheTestSuite) SetupTest() {
	// 检查Redis是否可用
	if err := InitRedis(); err != nil {
		s.T().Skip("Redis不可用，跳过测试")
	}

	s.manager = NewCacheManager()
	s.ttlManager = NewTTLManager()
	s.wrapper = NewCacheWrapper()
}

// TearDownTest 每个测试后的清理
func (s *CacheTestSuite) TearDownTest() {
	if RedisClient != nil {
		// 清理测试数据
		RedisClient.FlushDB(RedisClient.Context())
	}
}

// TearDownSuite 测试套件清理
func (s *CacheTestSuite) TearDownSuite() {
	if RedisClient != nil {
		CloseRedis()
	}
}

// TestRedisConnection 测试Redis连接
func (s *CacheTestSuite) TestRedisConnection() {
	err := HealthCheck()
	assert.NoError(s.T(), err)

	stats := GetConnectionStats()
	assert.Equal(s.T(), "connected", stats["status"])
}

// TestCacheBasicOperations 测试基础缓存操作
func (s *CacheTestSuite) TestCacheBasicOperations() {
	key := "test:basic"
	value := "test_value"

	// 测试设置缓存
	err := s.manager.Set(key, value)
	assert.NoError(s.T(), err)

	// 测试获取缓存
	var result string
	err = s.manager.Get(key, &result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value, result)

	// 测试检查存在性
	exists, err := s.manager.Exists(key)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), exists)

	// 测试删除缓存
	err = s.manager.Delete(key)
	assert.NoError(s.T(), err)

	// 确认删除成功
	err = s.manager.Get(key, &result)
	assert.Equal(s.T(), ErrCacheNotFound, err)
}

// TestCacheWithTTL 测试带TTL的缓存操作
func (s *CacheTestSuite) TestCacheWithTTL() {
	key := "test:ttl"
	value := "test_value"
	ttl := 2 * time.Second

	// 设置带TTL的缓存
	err := s.manager.SetWithTTL(key, value, ttl)
	assert.NoError(s.T(), err)

	// 立即获取应该成功
	var result string
	err = s.manager.Get(key, &result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value, result)

	// 检查TTL
	remainingTTL, err := s.manager.TTL(key)
	assert.NoError(s.T(), err)
	assert.True(s.T(), remainingTTL > 0 && remainingTTL <= ttl)

	// 等待过期
	time.Sleep(ttl + 100*time.Millisecond)

	// 过期后获取应该失败
	err = s.manager.Get(key, &result)
	assert.Equal(s.T(), ErrCacheNotFound, err)
}

// TestCacheJsonSerialization 测试JSON序列化
func (s *CacheTestSuite) TestCacheJsonSerialization() {
	type TestStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	key := "test:json"
	value := TestStruct{ID: 123, Name: "test"}

	// 设置复杂对象
	err := s.manager.Set(key, value)
	assert.NoError(s.T(), err)

	// 获取复杂对象
	var result TestStruct
	err = s.manager.Get(key, &result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value, result)
}

// TestHashOperations 测试Hash操作
func (s *CacheTestSuite) TestHashOperations() {
	key := "test:hash"
	field1 := "field1"
	value1 := "value1"
	field2 := "field2"
	value2 := 123

	// 设置Hash字段
	err := s.manager.HSet(key, field1, value1)
	assert.NoError(s.T(), err)

	err = s.manager.HSet(key, field2, value2)
	assert.NoError(s.T(), err)

	// 获取Hash字段
	var result1 string
	err = s.manager.HGet(key, field1, &result1)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value1, result1)

	var result2 int
	err = s.manager.HGet(key, field2, &result2)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value2, result2)

	// 检查字段存在性
	exists, err := s.manager.HExists(key, field1)
	assert.NoError(s.T(), err)
	assert.True(s.T(), exists)

	// 删除字段
	err = s.manager.HDelete(key, field1)
	assert.NoError(s.T(), err)

	// 确认删除成功
	exists, err = s.manager.HExists(key, field1)
	assert.NoError(s.T(), err)
	assert.False(s.T(), exists)
}

// TestSetOperations 测试集合操作
func (s *CacheTestSuite) TestSetOperations() {
	key := "test:set"
	member1 := "member1"
	member2 := "member2"

	// 添加集合成员
	err := s.manager.SAdd(key, member1, member2)
	assert.NoError(s.T(), err)

	// 检查成员是否存在
	isMember, err := s.manager.SIsMember(key, member1)
	assert.NoError(s.T(), err)
	assert.True(s.T(), isMember)

	// 获取所有成员
	members, err := s.manager.SMembers(key)
	assert.NoError(s.T(), err)
	assert.Contains(s.T(), members, member1)
	assert.Contains(s.T(), members, member2)

	// 删除成员
	err = s.manager.SRemove(key, member1)
	assert.NoError(s.T(), err)

	// 确认删除成功
	isMember, err = s.manager.SIsMember(key, member1)
	assert.NoError(s.T(), err)
	assert.False(s.T(), isMember)
}

// TestIncrementOperations 测试原子递增操作
func (s *CacheTestSuite) TestIncrementOperations() {
	key := "test:counter"

	// 递增（从0开始）
	count, err := s.manager.Increment(key)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), count)

	// 按值递增
	count, err = s.manager.IncrementBy(key, 5)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(6), count)

	// 递减
	count, err = s.manager.Decrement(key)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(5), count)

	// 按值递减
	count, err = s.manager.DecrementBy(key, 3)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(2), count)
}

// TestTTLManager TTL管理器测试
func (s *CacheTestSuite) TestTTLManager() {
	// 测试不同类型的TTL
	userSessionTTL := s.ttlManager.GetTTL("user_session")
	assert.Equal(s.T(), 2*time.Hour, userSessionTTL)

	fileInfoTTL := s.ttlManager.GetTTL("file_info")
	assert.Equal(s.T(), config.AppConfig.Cache.FileInfoTTL, fileInfoTTL)

	defaultTTL := s.ttlManager.GetTTL("unknown_type")
	assert.Equal(s.T(), config.AppConfig.Cache.DefaultTTL, defaultTTL)

	// 测试TTL验证
	err := s.ttlManager.ValidateTTL(time.Hour)
	assert.NoError(s.T(), err)

	err = s.ttlManager.ValidateTTL(-time.Hour)
	assert.Equal(s.T(), ErrInvalidTTL, err)

	err = s.ttlManager.ValidateTTL(8 * 24 * time.Hour)
	assert.Equal(s.T(), ErrInvalidTTL, err)
}

// TestCacheWrapper 测试缓存包装器
func (s *CacheTestSuite) TestCacheWrapper() {
	userID := "test_user_123"
	token := "test_token_456"
	fileID := "test_file_789"

	// 测试用户会话
	sessionData := map[string]interface{}{
		"user_id": userID,
		"role":    "user",
	}
	err := s.wrapper.SetUserSession(token, sessionData)
	assert.NoError(s.T(), err)

	var retrievedSession map[string]interface{}
	err = s.wrapper.GetUserSession(token, &retrievedSession)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), sessionData, retrievedSession)

	// 测试用户权限
	permissions := []string{"read", "write"}
	err = s.wrapper.SetUserPermissions(userID, permissions)
	assert.NoError(s.T(), err)

	retrievedPerms, err := s.wrapper.GetUserPermissions(userID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), permissions, retrievedPerms)

	// 测试在线用户
	err = s.wrapper.SetOnlineUser(userID)
	assert.NoError(s.T(), err)

	isOnline := s.wrapper.IsUserOnline(userID)
	assert.True(s.T(), isOnline)

	// 测试验证码
	code := "123456"
	err = s.wrapper.SetVerificationCode("email", "test@example.com", code)
	assert.NoError(s.T(), err)

	retrievedCode, err := s.wrapper.GetVerificationCode("email", "test@example.com")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), code, retrievedCode)

	// 测试限流
	count, err := s.wrapper.IncrementRateLimit("127.0.0.1", "/api/test")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), count)

	// 测试清理缓存
	err = s.wrapper.ClearUserCache(userID)
	assert.NoError(s.T(), err)

	err = s.wrapper.ClearFileCache(fileID)
	assert.NoError(s.T(), err)
}

// TestBatchOperations 测试批量操作
func (s *CacheTestSuite) TestBatchOperations() {
	batch := s.manager.Batch()

	// 批量设置多个键值对
	batch.Set("batch:key1", "value1", time.Hour)
	batch.Set("batch:key2", "value2", time.Hour)

	// 执行批量操作
	err := batch.Execute()
	assert.NoError(s.T(), err)

	// 验证设置成功
	var result1, result2 string
	err = s.manager.Get("batch:key1", &result1)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "value1", result1)

	err = s.manager.Get("batch:key2", &result2)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "value2", result2)
}

// TestKeyBuilder 测试键构建器
func (s *CacheTestSuite) TestKeyBuilder() {
	kb := NewKeyBuilder()

	// 测试用户相关键
	userID := "user123"
	assert.Equal(s.T(), "session:token123", kb.UserSession("token123"))
	assert.Equal(s.T(), "permissions:user123", kb.UserPermissions(userID))
	assert.Equal(s.T(), "profile:user123", kb.UserProfile(userID))

	// 测试文件相关键
	fileID := "file456"
	assert.Equal(s.T(), "file:file456", kb.FileInfo(fileID))
	assert.Equal(s.T(), "share:token789", kb.FileShare("token789"))
	assert.Equal(s.T(), "chunk:upload123:1", kb.FileChunk("upload123", 1))

	// 测试验证码相关键
	assert.Equal(s.T(), "code:email:test@example.com", kb.VerifyCode("email", "test@example.com"))
	assert.Equal(s.T(), "rate:127.0.0.1:/api/test", kb.RateLimit("127.0.0.1", "/api/test"))
}

// 运行测试套件
func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}

// 基准测试
func BenchmarkCacheSet(b *testing.B) {
	if testing.Short() {
		b.Skip("跳过需要Redis连接的基准测试")
	}

	// 初始化
	config.AppConfig = &config.Config{
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		Cache: config.CacheConfig{
			DefaultTTL: time.Hour,
		},
	}

	if err := InitRedis(); err != nil {
		b.Skip("Redis不可用")
	}
	defer CloseRedis()

	manager := NewCacheManager()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench:key:%d", i)
			manager.Set(key, "test_value")
			i++
		}
	})
}

func BenchmarkCacheGet(b *testing.B) {
	if testing.Short() {
		b.Skip("跳过需要Redis连接的基准测试")
	}

	// 初始化
	config.AppConfig = &config.Config{
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		Cache: config.CacheConfig{
			DefaultTTL: time.Hour,
		},
	}

	if err := InitRedis(); err != nil {
		b.Skip("Redis不可用")
	}
	defer CloseRedis()

	manager := NewCacheManager()

	// 预设一些数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench:key:%d", i)
		manager.Set(key, "test_value")
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench:key:%d", i%1000)
			var result string
			manager.Get(key, &result)
			i++
		}
	})
}
