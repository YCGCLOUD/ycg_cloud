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
	// 尝试加载配置，如果失败则使用测试配置
	if err := config.Load(); err != nil {
		// 配置加载失败，使用测试环境的Redis配置
		config.AppConfig = &config.Config{
			Redis: config.RedisConfig{
				Host:         "dbconn.sealosbja.site",
				Port:         38650,
				Password:     "dttd62tf",
				DB:           1,
				PoolSize:     5,
				MinIdleConns: 2,
				MaxRetries:   3,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolTimeout:  4 * time.Second,
				IdleTimeout:  300 * time.Second,
			},
			Cache: config.CacheConfig{
				DefaultTTL:          time.Hour,
				UserInfoTTL:         30 * time.Minute,
				FileInfoTTL:         10 * time.Minute,
				VerificationCodeTTL: 5 * time.Minute,
			},
		}
	}

	// 验证Redis配置是否存在
	if config.AppConfig.Redis.Host == "" {
		s.T().Skip("Redis配置为空，跳过缓存测试")
		return
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

// TestSortedSetOperations 测试有序集合操作
func (s *CacheTestSuite) TestSortedSetOperations() {
	key := "test:zset"
	member1 := "member1"
	member2 := "member2"
	score1 := 10.5
	score2 := 20.5

	// 添加有序集合成员
	err := s.manager.ZAdd(key, score1, member1)
	assert.NoError(s.T(), err)

	err = s.manager.ZAdd(key, score2, member2)
	assert.NoError(s.T(), err)

	// 获取范围成员
	members, err := s.manager.ZRange(key, 0, -1)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), members, 2)
	assert.Equal(s.T(), member1, members[0]) // 分数较低的在前
	assert.Equal(s.T(), member2, members[1])

	// 删除成员
	err = s.manager.ZRemove(key, member1)
	assert.NoError(s.T(), err)

	// 确认删除成功
	members, err = s.manager.ZRange(key, 0, -1)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), members, 1)
	assert.Equal(s.T(), member2, members[0])
}

// TestSerializationTypes 测试不同类型的序列化
func (s *CacheTestSuite) TestSerializationTypes() {
	// 测试字符串类型
	key1 := "test:string"
	value1 := "test_string"
	err := s.manager.Set(key1, value1)
	assert.NoError(s.T(), err)
	var result1 string
	err = s.manager.Get(key1, &result1)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value1, result1)

	// 测试整数类型
	key2 := "test:int"
	value2 := 12345
	err = s.manager.Set(key2, value2)
	assert.NoError(s.T(), err)
	var result2 int
	err = s.manager.Get(key2, &result2)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value2, result2)

	// 测试浮点数类型
	key3 := "test:float"
	value3 := 123.45
	err = s.manager.Set(key3, value3)
	assert.NoError(s.T(), err)
	var result3 float64
	err = s.manager.Get(key3, &result3)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value3, result3)

	// 测试布尔类型
	key4 := "test:bool"
	value4 := true
	err = s.manager.Set(key4, value4)
	assert.NoError(s.T(), err)
	var result4 bool
	err = s.manager.Get(key4, &result4)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value4, result4)

	// 测试字节数组
	key5 := "test:bytes"
	value5 := []byte("test_bytes")
	err = s.manager.Set(key5, value5)
	assert.NoError(s.T(), err)
	var result5 []byte
	err = s.manager.Get(key5, &result5)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value5, result5)

	// 测试复杂数值类型
	key6 := "test:int64"
	value6 := int64(9223372036854775807)
	err = s.manager.Set(key6, value6)
	assert.NoError(s.T(), err)
	var result6 int64
	err = s.manager.Get(key6, &result6)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value6, result6)

	// 测试uint类型
	key7 := "test:uint"
	value7 := uint(4294967295)
	err = s.manager.Set(key7, value7)
	assert.NoError(s.T(), err)
	var result7 uint
	err = s.manager.Get(key7, &result7)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value7, result7)
}

// TestErrorHandling 测试错误处理
func (s *CacheTestSuite) TestErrorHandling() {
	// 测试获取不存在的键
	var result string
	err := s.manager.Get("nonexistent:key", &result)
	assert.Equal(s.T(), ErrCacheNotFound, err)

	// 测试空键列表的删除
	err = s.manager.Delete()
	assert.NoError(s.T(), err)

	// 测试空键列表的存在性检查
	exists, err := s.manager.Exists()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), exists)

	// 测试Hash操作中获取不存在的字段
	err = s.manager.HGet("nonexistent:hash", "field", &result)
	assert.Equal(s.T(), ErrCacheNotFound, err)

	// 测试空字段列表的Hash删除
	err = s.manager.HDelete("test:hash")
	assert.NoError(s.T(), err)

	// 测试设置过期时间
	key := "test:expire"
	err = s.manager.Set(key, "value")
	assert.NoError(s.T(), err)

	err = s.manager.Expire(key, time.Second)
	assert.NoError(s.T(), err)

	// 验证TTL设置成功
	ttl, err := s.manager.TTL(key)
	assert.NoError(s.T(), err)
	assert.True(s.T(), ttl > 0)
}

// TestAdvancedBatchOperations 测试高级批量操作
func (s *CacheTestSuite) TestAdvancedBatchOperations() {
	batch := s.manager.Batch()

	// 批量设置多种类型的数据
	batch.Set("batch:string", "value1", time.Hour)
	batch.Set("batch:int", 123, time.Hour)
	batch.Set("batch:float", 123.45, time.Hour)
	batch.Set("batch:bool", true, time.Hour)

	// 执行批量操作
	err := batch.Execute()
	assert.NoError(s.T(), err)

	// 验证批量操作结果
	var stringResult string
	err = s.manager.Get("batch:string", &stringResult)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "value1", stringResult)

	var intResult int
	err = s.manager.Get("batch:int", &intResult)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 123, intResult)

	var floatResult float64
	err = s.manager.Get("batch:float", &floatResult)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 123.45, floatResult)

	var boolResult bool
	err = s.manager.Get("batch:bool", &boolResult)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), true, boolResult)

	// 测试批量删除
	batch2 := s.manager.Batch()
	batch2.Delete("batch:string", "batch:int")
	err = batch2.Execute()
	assert.NoError(s.T(), err)

	// 验证删除成功
	var result string
	err = s.manager.Get("batch:string", &result)
	assert.Equal(s.T(), ErrCacheNotFound, err)
}

// TestExtendedKeyBuilder 测试扩展键构建器功能
func (s *CacheTestSuite) TestExtendedKeyBuilder() {
	kb := NewKeyBuilder()

	// 测试团队相关键
	teamID := "team123"
	userID := "user456"
	assert.Equal(s.T(), "team:team123", kb.TeamInfo(teamID))
	assert.Equal(s.T(), "team:members:team123", kb.TeamMembers(teamID))
	assert.Equal(s.T(), "team:files:team123", kb.TeamFiles(teamID))
	assert.Equal(s.T(), "team:perms:team123:user456", kb.TeamPermissions(teamID, userID))

	// 测试验证码相关键
	assert.Equal(s.T(), "code:sms:13800138000", kb.VerifyCode("sms", "13800138000"))
	assert.Equal(s.T(), "attempt:email:test@example.com", kb.VerifyAttempt("email", "test@example.com"))
	assert.Equal(s.T(), "block:login:user123", kb.VerifyBlock("login", "user123"))

	// 测试限流相关键
	assert.Equal(s.T(), "rate:192.168.1.1:/api/upload", kb.RateLimit("192.168.1.1", "/api/upload"))
	assert.Equal(s.T(), "user_rate:user123:download", kb.UserRateLimit("user123", "download"))
	assert.Equal(s.T(), "api_rate:apikey123:/api/search", kb.APIRateLimit("apikey123", "/api/search"))

	// 测试锁相关键
	assert.Equal(s.T(), "lock:file:file123", kb.FileLock("file123"))
	assert.Equal(s.T(), "lock:user:user123", kb.UserLock("user123"))
	assert.Equal(s.T(), "lock:team:team123", kb.TeamLock("team123"))
	assert.Equal(s.T(), "lock:upload:upload123", kb.UploadLock("upload123"))
}

// TestManagerInitialization 测试管理器初始化
func (s *CacheTestSuite) TestManagerInitialization() {
	// 测试新建缓存管理器
	manager := NewCacheManager()
	assert.NotNil(s.T(), manager)

	// 测试延迟初始化的客户端
	assert.NotNil(s.T(), manager.getClient())

	// 测试TTL管理器
	ttlManager := NewTTLManager()
	assert.NotNil(s.T(), ttlManager)

	// 测试缓存包装器
	wrapper := NewCacheWrapper()
	assert.NotNil(s.T(), wrapper)
}

// TestTTLManagerExtended 测试TTL管理器扩展功能
func (s *CacheTestSuite) TestTTLManagerExtended() {
	ttlManager := NewTTLManager()

	// 测试各种预定义TTL类型
	assert.Equal(s.T(), 2*time.Hour, ttlManager.GetTTL("user_session"))
	assert.Equal(s.T(), 1*time.Hour, ttlManager.GetTTL("user_permissions"))
	assert.Equal(s.T(), 30*time.Minute, ttlManager.GetTTL("file_preview"))
	assert.Equal(s.T(), 1*time.Hour, ttlManager.GetTTL("file_share"))
	assert.Equal(s.T(), 24*time.Hour, ttlManager.GetTTL("file_upload"))
	assert.Equal(s.T(), 30*time.Minute, ttlManager.GetTTL("team_info"))
	assert.Equal(s.T(), 15*time.Minute, ttlManager.GetTTL("team_members"))
	assert.Equal(s.T(), 15*time.Minute, ttlManager.GetTTL("verify_attempt"))
	assert.Equal(s.T(), 1*time.Hour, ttlManager.GetTTL("verify_block"))
	assert.Equal(s.T(), 1*time.Minute, ttlManager.GetTTL("rate_limit"))
	assert.Equal(s.T(), 5*time.Minute, ttlManager.GetTTL("user_rate_limit"))
	assert.Equal(s.T(), 1*time.Minute, ttlManager.GetTTL("api_rate_limit"))
	assert.Equal(s.T(), 10*time.Minute, ttlManager.GetTTL("lock"))
	assert.Equal(s.T(), 15*time.Minute, ttlManager.GetTTL("search_result"))
	assert.Equal(s.T(), 24*time.Hour, ttlManager.GetTTL("search_history"))
	assert.Equal(s.T(), 10*time.Minute, ttlManager.GetTTL("stats_user"))
	assert.Equal(s.T(), 5*time.Minute, ttlManager.GetTTL("stats_file"))
	assert.Equal(s.T(), 1*time.Minute, ttlManager.GetTTL("stats_system"))
	assert.Equal(s.T(), 1*time.Hour, ttlManager.GetTTL("message"))
	assert.Equal(s.T(), 30*time.Minute, ttlManager.GetTTL("conversation"))
	assert.Equal(s.T(), 5*time.Minute, ttlManager.GetTTL("online_users"))

	// 测试基于配置的TTL
	assert.Equal(s.T(), config.AppConfig.Cache.UserInfoTTL, ttlManager.GetTTL("user_profile"))
	assert.Equal(s.T(), config.AppConfig.Cache.UserInfoTTL, ttlManager.GetTTL("user_info"))
	assert.Equal(s.T(), config.AppConfig.Cache.FileInfoTTL, ttlManager.GetTTL("file_info"))
	assert.Equal(s.T(), config.AppConfig.Cache.VerificationCodeTTL, ttlManager.GetTTL("verify_code"))
	assert.Equal(s.T(), config.AppConfig.Cache.DefaultTTL, ttlManager.GetTTL("unknown_type"))

	// 测试快捷TTL方法
	assert.Equal(s.T(), 5*time.Minute, ttlManager.GetShortTTL())
	assert.Equal(s.T(), 30*time.Minute, ttlManager.GetMediumTTL())
	assert.Equal(s.T(), 2*time.Hour, ttlManager.GetLongTTL())
	assert.Equal(s.T(), 24*time.Hour, ttlManager.GetPersistentTTL())

	// 测试TTL验证
	assert.NoError(s.T(), ttlManager.ValidateTTL(time.Hour))
	assert.NoError(s.T(), ttlManager.ValidateTTL(0))
	assert.Equal(s.T(), ErrInvalidTTL, ttlManager.ValidateTTL(-time.Hour))
	assert.Equal(s.T(), ErrInvalidTTL, ttlManager.ValidateTTL(8*24*time.Hour))
}

// TestCacheWrapperExtended 测试缓存包装器扩展功能
func (s *CacheTestSuite) TestCacheWrapperExtended() {
	wrapper := NewCacheWrapper()
	fileID := "test_file_789"

	// 测试SetByType方法
	err := wrapper.SetByType("test:by:type", "test_value", "user_session")
	assert.NoError(s.T(), err)

	// 验证TTL是否正确设置
	ttl, err := s.manager.TTL("test:by:type")
	assert.NoError(s.T(), err)
	assert.True(s.T(), ttl > 0)

	// 测试文件信息缓存
	fileInfo := map[string]interface{}{
		"id":   fileID,
		"name": "test.txt",
		"size": 1024.0, // 使用float64避免JSON反序列化类型问题
	}
	err = wrapper.SetFileInfo(fileID, fileInfo)
	assert.NoError(s.T(), err)

	var retrievedFileInfo map[string]interface{}
	err = wrapper.GetFileInfo(fileID, &retrievedFileInfo)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), fileInfo, retrievedFileInfo)
}

// TestAdvancedSerialization 测试高级序列化功能
func (s *CacheTestSuite) TestAdvancedSerialization() {
	manager := s.manager

	// 测试空接口序列化
	key1 := "test:nil"
	var nilValue interface{} = nil
	err := manager.Set(key1, nilValue)
	assert.NoError(s.T(), err)

	var result1 interface{}
	err = manager.Get(key1, &result1)
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), result1)

	// 测试空数组序列化
	key2 := "test:empty:array"
	emptyArray := []string{}
	err = manager.Set(key2, emptyArray)
	assert.NoError(s.T(), err)

	var result2 []string
	err = manager.Get(key2, &result2)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), emptyArray, result2)

	// 测试空对象序列化
	key3 := "test:empty:map"
	emptyMap := map[string]interface{}{}
	err = manager.Set(key3, emptyMap)
	assert.NoError(s.T(), err)

	var result3 map[string]interface{}
	err = manager.Get(key3, &result3)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), emptyMap, result3)

	// 测试复杂嵌套结构
	key4 := "test:nested"
	nestedData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":     123,
			"name":   "test",
			"active": true,
			"tags":   []string{"admin", "user"},
			"profile": map[string]interface{}{
				"email": "test@example.com",
				"age":   25,
			},
		},
	}
	err = manager.Set(key4, nestedData)
	assert.NoError(s.T(), err)

	var result4 map[string]interface{}
	err = manager.Get(key4, &result4)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result4["user"])
}

// TestDeserializationEdgeCases 测试反序列化边界情况
func (s *CacheTestSuite) TestDeserializationEdgeCases() {
	manager := s.manager

	// 测试反序列化到不同类型的指针
	key := "test:deserialize"
	originalValue := "test_string_value"
	err := manager.Set(key, originalValue)
	assert.NoError(s.T(), err)

	// 反序列化到string指针
	var stringResult string
	err = manager.Get(key, &stringResult)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), originalValue, stringResult)

	// 反序列化到[]byte指针
	var bytesResult []byte
	err = manager.Get(key, &bytesResult)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), []byte(originalValue), bytesResult)

	// 测试bool反序列化
	boolKey := "test:bool:deserialize"
	err = manager.Set(boolKey, "1")
	assert.NoError(s.T(), err)

	var boolResult bool
	err = manager.Get(boolKey, &boolResult)
	assert.NoError(s.T(), err)
	assert.True(s.T(), boolResult)

	// 测试bool反序列化 - "true"
	boolKey2 := "test:bool:true"
	err = manager.Set(boolKey2, "true")
	assert.NoError(s.T(), err)

	var boolResult2 bool
	err = manager.Get(boolKey2, &boolResult2)
	assert.NoError(s.T(), err)
	assert.True(s.T(), boolResult2)

	// 测试bool反序列化 - "0"
	boolKey3 := "test:bool:false"
	err = manager.Set(boolKey3, "0")
	assert.NoError(s.T(), err)

	var boolResult3 bool
	err = manager.Get(boolKey3, &boolResult3)
	assert.NoError(s.T(), err)
	assert.False(s.T(), boolResult3)
}

// TestRedisConnectionManagement 测试Redis连接管理
func (s *CacheTestSuite) TestRedisConnectionManagement() {
	// 测试获取Redis客户端
	client := GetRedisClient()
	assert.NotNil(s.T(), client)

	// 测试健康检查
	err := HealthCheck()
	assert.NoError(s.T(), err)

	// 测试连接统计信息
	stats := GetConnectionStats()
	assert.NotNil(s.T(), stats)
	assert.Equal(s.T(), "connected", stats["status"])
	assert.Contains(s.T(), stats, "hits")
	assert.Contains(s.T(), stats, "misses")
	assert.Contains(s.T(), stats, "timeouts")
	assert.Contains(s.T(), stats, "total_conns")
	assert.Contains(s.T(), stats, "idle_conns")
	assert.Contains(s.T(), stats, "stale_conns")
}

// TestComplexKeyBuilder 测试复杂键构建器场景
func (s *CacheTestSuite) TestComplexKeyBuilder() {
	kb := NewKeyBuilder()

	// 测试消息相关键
	conversationID := "conv123"
	messageID := "msg456"
	userID := "user789"
	assert.Equal(s.T(), "msg:conv:conv123", kb.Conversation(conversationID))
	assert.Equal(s.T(), "msg:msg456", kb.Message(messageID))
	assert.Equal(s.T(), "msg:read:conv123:user789", kb.MessageRead(conversationID, userID))
	assert.Equal(s.T(), "msg:user:user789", kb.UserMessages(userID))

	// 测试统计相关键
	assert.Equal(s.T(), "stats:user:user789", kb.UserStats(userID))
	fileID := "file123"
	assert.Equal(s.T(), "stats:file:file123", kb.FileStats(fileID))
	teamID := "team123"
	assert.Equal(s.T(), "stats:team:team123", kb.TeamStats(teamID))
	assert.Equal(s.T(), "stats:system", kb.SystemStats())

	// 测试搜索相关键
	indexType := "file"
	queryHash := "hash123"
	assert.Equal(s.T(), "search:index:file", kb.SearchIndex(indexType))
	assert.Equal(s.T(), "search:result:hash123", kb.SearchResult(queryHash))
	assert.Equal(s.T(), "search:history:user789", kb.SearchHistory(userID))

	// 测试更多文件相关键
	uploadID := "upload123"
	chunkNum := 5
	token := "token456"
	assert.Equal(s.T(), "file:file123", kb.FileInfo(fileID))
	assert.Equal(s.T(), "share:token456", kb.FileShare(token))
	assert.Equal(s.T(), "upload:upload123", kb.FileUpload(uploadID))
	assert.Equal(s.T(), "chunk:upload123:5", kb.FileChunk(uploadID, chunkNum))
	assert.Equal(s.T(), "preview:file123", kb.FilePreview(fileID))
	assert.Equal(s.T(), "download:file123", kb.FileDownload(fileID))

	// 测试更多用户相关键
	assert.Equal(s.T(), "profile:user789", kb.UserProfile(userID))
	assert.Equal(s.T(), "online:user789", kb.UserOnline(userID))
	assert.Equal(s.T(), "quota:user789", kb.UserQuota(userID))
}

// TestGlobalKeysInstance 测试全局Keys实例
func (s *CacheTestSuite) TestGlobalKeysInstance() {
	// 测试全局Keys实例
	assert.NotNil(s.T(), Keys)

	// 测试通过全局实例生成键
	userID := "global_user_123"
	fileID := "global_file_456"

	assert.Equal(s.T(), "session:token123", Keys.UserSession("token123"))
	assert.Equal(s.T(), "permissions:global_user_123", Keys.UserPermissions(userID))
	assert.Equal(s.T(), "file:global_file_456", Keys.FileInfo(fileID))
	assert.Equal(s.T(), "team:team123", Keys.TeamInfo("team123"))
	assert.Equal(s.T(), "code:email:test@example.com", Keys.VerifyCode("email", "test@example.com"))
	assert.Equal(s.T(), "rate:192.168.1.1:/api/test", Keys.RateLimit("192.168.1.1", "/api/test"))
	assert.Equal(s.T(), "lock:file:global_file_456", Keys.FileLock(fileID))
	assert.Equal(s.T(), "stats:system", Keys.SystemStats())
}

// TestCacheExpiration 测试缓存过期功能
func (s *CacheTestSuite) TestCacheExpiration() {
	key := "test:expiration"
	value := "test_value"

	// 设置缓存
	err := s.manager.Set(key, value)
	assert.NoError(s.T(), err)

	// 设置过期时间为1秒
	err = s.manager.Expire(key, time.Second)
	assert.NoError(s.T(), err)

	// 立即检查TTL
	ttl, err := s.manager.TTL(key)
	assert.NoError(s.T(), err)
	assert.True(s.T(), ttl > 0 && ttl <= time.Second)

	// 立即获取应该成功
	var result string
	err = s.manager.Get(key, &result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value, result)

	// 等待过期
	time.Sleep(1100 * time.Millisecond)

	// 过期后获取应该失败
	err = s.manager.Get(key, &result)
	assert.Equal(s.T(), ErrCacheNotFound, err)

	// 检查TTL应该返回-2（键不存在）
	ttl, err = s.manager.TTL(key)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), -2*time.Nanosecond, ttl)
}
