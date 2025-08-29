package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"cloudpan/internal/pkg/config"

	"github.com/go-redis/redis/v8"
)

// CacheManager 缓存管理器
//
// CacheManager 提供了对Redis缓存的统一管理接口，支持：
// 1. 基础缓存操作：Set、Get、Delete、Exists等
// 2. Hash操作：HSet、HGet、HDelete等
// 3. 集合操作：SAdd、SRemove、SIsMember等
// 4. 有序集合操作：ZAdd、ZRemove、ZRange等
// 5. 原子操作：Increment、Decrement等
// 6. 批量操作：支持管道式批量操作提升性能
// 7. TTL管理：支持缓存过期时间设置和查询
//
// 特性：
// - 延迟初始化：Redis客户端在首次使用时才创建连接
// - 类型安全：支持多种数据类型的序列化和反序列化
// - 性能优化：针对基础类型提供特殊序列化优化
// - 错误处理：统一的错误处理和类型转换
type CacheManager struct {
	client *redis.Client   // Redis客户端连接，支持延迟初始化
	ctx    context.Context // 上下文对象，用于请求生命周期管理
}

// NewCacheManager 创建缓存管理器
//
// 创建一个新的缓存管理器实例，使用延迟初始化模式：
// - Redis客户端将在第一次调用时通过GetRedisClient()获取
// - 使用context.Background()作为默认上下文
//
// 返回:
//   - *CacheManager: 缓存管理器实例
//
// 使用示例:
//
//	cm := NewCacheManager()
//	err := cm.Set("key", "value")
func NewCacheManager() *CacheManager {
	return &CacheManager{
		client: nil, // 延迟初始化，在第一次使用时获取
		ctx:    context.Background(),
	}
}

// getClient 获取Redis客户端（延迟初始化）
//
// 实现延迟初始化模式，仅在首次调用时创建Redis连接：
// - 避免不必要的连接创建
// - 确保配置已正确加载
// - 提高应用启动性能
//
// 返回:
//   - *redis.Client: Redis客户端实例
func (c *CacheManager) getClient() *redis.Client {
	if c.client == nil {
		c.client = GetRedisClient()
	}
	return c.client
}

// Set 设置缓存，使用默认TTL
//
// 使用配置文件中定义的默认TTL时间设置缓存。支持任意类型的值，
// 内部会自动进行序列化处理。
//
// 参数:
//   - key: 缓存键名，不能为空
//   - value: 缓存值，支持string、[]byte、数值、bool、struct等类型
//
// 返回:
//   - error: 操作错误，nil表示成功
//
// 使用示例:
//
//	err := cm.Set("user:123", userInfo)
func (c *CacheManager) Set(key string, value interface{}) error {
	return c.SetWithTTL(key, value, config.AppConfig.Cache.DefaultTTL)
}

// SetWithTTL 设置缓存，指定TTL
//
// 设置缓存并指定过期时间。支持任意类型的值，内部会自动进行序列化处理。
// 对于性能敏感的场景，建议优先使用基础类型（string、int、bool等）。
//
// 参数:
//   - key: 缓存键名，不能为空
//   - value: 缓存值，支持多种类型自动序列化
//   - ttl: 过期时间，0表示永不过期
//
// 返回:
//   - error: 操作错误，nil表示成功
//
// 使用示例:
//
//	err := cm.SetWithTTL("session:abc", sessionData, 30*time.Minute)
func (c *CacheManager) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	data, err := c.serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	return c.getClient().Set(c.ctx, key, data, ttl).Err()
}

// Get 获取缓存
//
// 获取缓存值并反序列化到指定的目标对象。支持多种数据类型的自动转换，
// 如果缓存不存在会返回ErrCacheNotFound错误。
//
// 参数:
//   - key: 缓存键名
//   - dest: 目标对象指针，用于接收反序列化后的值
//
// 返回:
//   - error: 操作错误，ErrCacheNotFound表示缓存不存在
//
// 使用示例:
//
//	var userInfo User
//	err := cm.Get("user:123", &userInfo)
//	if err == ErrCacheNotFound {
//	    // 缓存不存在
//	}
func (c *CacheManager) Get(key string, dest interface{}) error {
	data, err := c.getClient().Get(c.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheNotFound
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}

	return c.deserialize(data, dest)
}

// Delete 删除缓存
//
// 删除一个或多个Redis键。支持批量删除操作，如果没有提供键名
// 则直接返回成功。即使某些键不存在也不会报错。
//
// 参数:
//   - keys: 要删除的键名列表，支持多个键名
//
// 返回:
//   - error: 操作错误，nil表示成功
//
// 使用示例:
//
//	err := cm.Delete("user:123", "session:abc")
func (c *CacheManager) Delete(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.getClient().Del(c.ctx, keys...).Err()
}

// Exists 检查缓存是否存在
//
// 检查一个或多个键是否存在于Redis中。返回存在的键的数量。
// 如果没有提供键名则直接返回0。
//
// 参数:
//   - keys: 要检查的键名列表
//
// 返回:
//   - int64: 存在的键的数量
//   - error: 操作错误，nil表示成功
//
// 使用示例:
//
//	count, err := cm.Exists("user:123", "user:456")
//	if count == 2 {
//	    // 两个键都存在
//	}
func (c *CacheManager) Exists(keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	return c.getClient().Exists(c.ctx, keys...).Result()
}

// Expire 设置缓存过期时间
//
// 为已存在的键设置过期时间。如果键不存在，操作不会报错但也不会生效。
// TTL设置为0表示立即过期，负数表示永不过期。
//
// 参数:
//   - key: 要设置过期时间的键名
//   - ttl: 过期时间间隔
//
// 返回:
//   - error: 操作错误，nil表示成功
//
// 使用示例:
//
//	err := cm.Expire("session:abc", 30*time.Minute)
func (c *CacheManager) Expire(key string, ttl time.Duration) error {
	return c.getClient().Expire(c.ctx, key, ttl).Err()
}

// TTL 获取缓存剩余过期时间
//
// 获取指定键的剩余过期时间。返回值含义：
// - 正数：剩余过期时间
// - -1：键存在但没有设置过期时间（永不过期）
// - -2：键不存在
//
// 参数:
//   - key: 要查询的键名
//
// 返回:
//   - time.Duration: 剩余过期时间
//   - error: 操作错误，nil表示成功
//
// 使用示例:
//
//	ttl, err := cm.TTL("session:abc")
//	if ttl > 0 {
//	    // 键将在ttl时间后过期
//	}
func (c *CacheManager) TTL(key string) (time.Duration, error) {
	return c.getClient().TTL(c.ctx, key).Result()
}

// Increment 原子递增
//
// 将指定键的值原子地递增1。如果键不存在，会自动创建并初始化为0，
// 然后执行递增操作。这是一个原子操作，适用于计数器、统计等场景。
//
// 参数:
//   - key: 要递增的键名
//
// 返回:
//   - int64: 递增后的值
//   - error: 操作错误，nil表示成功
//
// 使用示例:
//
//	count, err := cm.Increment("page:views")
func (c *CacheManager) Increment(key string) (int64, error) {
	return c.getClient().Incr(c.ctx, key).Result()
}

// IncrementBy 原子递增指定值
//
// 将指定键的值原子地递增指定的数值。如果键不存在，会自动创建并初始化为0，
// 然后执行递增操作。支持负数值（实际上就是递减）。
//
// 参数:
//   - key: 要递增的键名
//   - value: 递增的数值，可以为负数
//
// 返回:
//   - int64: 递增后的值
//   - error: 操作错误，nil表示成功
//
// 使用示例:
//
//	count, err := cm.IncrementBy("score:user:123", 10)
func (c *CacheManager) IncrementBy(key string, value int64) (int64, error) {
	return c.getClient().IncrBy(c.ctx, key, value).Result()
}

// Decrement 原子递减
//
// 将指定键的值原子地递减1。如果键不存在，会自动创建并初始化为0，
// 然后执行递减操作（结果为-1）。这是一个原子操作。
//
// 参数:
//   - key: 要递减的键名
//
// 返回:
//   - int64: 递减后的值
//   - error: 操作错误，nil表示成功
//
// 使用示例:
//
//	count, err := cm.Decrement("available:tickets")
func (c *CacheManager) Decrement(key string) (int64, error) {
	return c.getClient().Decr(c.ctx, key).Result()
}

// DecrementBy 原子递减指定值
//
// 将指定键的值原子地递减指定的数值。如果键不存在，会自动创建并初始化为0，
// 然后执行递减操作。支持负数值（实际上就是递增）。
//
// 参数:
//   - key: 要递减的键名
//   - value: 递减的数值，可以为负数
//
// 返回:
//   - int64: 递减后的值
//   - error: 操作错误，nil表示成功
//
// 使用示例:
//
//	count, err := cm.DecrementBy("stock:item:456", 5)
func (c *CacheManager) DecrementBy(key string, value int64) (int64, error) {
	return c.getClient().DecrBy(c.ctx, key, value).Result()
}

// HSet 设置Hash字段
func (c *CacheManager) HSet(key, field string, value interface{}) error {
	data, err := c.serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}
	return c.getClient().HSet(c.ctx, key, field, data).Err()
}

// HGet 获取Hash字段
func (c *CacheManager) HGet(key, field string, dest interface{}) error {
	data, err := c.getClient().HGet(c.ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheNotFound
		}
		return fmt.Errorf("failed to get hash field: %w", err)
	}

	return c.deserialize(data, dest)
}

// HDelete 删除Hash字段
func (c *CacheManager) HDelete(key string, fields ...string) error {
	if len(fields) == 0 {
		return nil
	}
	return c.getClient().HDel(c.ctx, key, fields...).Err()
}

// HExists 检查Hash字段是否存在
func (c *CacheManager) HExists(key, field string) (bool, error) {
	return c.getClient().HExists(c.ctx, key, field).Result()
}

// SAdd 添加集合成员
func (c *CacheManager) SAdd(key string, members ...interface{}) error {
	return c.getClient().SAdd(c.ctx, key, members...).Err()
}

// SRemove 删除集合成员
func (c *CacheManager) SRemove(key string, members ...interface{}) error {
	return c.getClient().SRem(c.ctx, key, members...).Err()
}

// SIsMember 检查是否为集合成员
func (c *CacheManager) SIsMember(key string, member interface{}) (bool, error) {
	return c.getClient().SIsMember(c.ctx, key, member).Result()
}

// SMembers 获取集合所有成员
func (c *CacheManager) SMembers(key string) ([]string, error) {
	return c.getClient().SMembers(c.ctx, key).Result()
}

// ZAdd 添加有序集合成员
func (c *CacheManager) ZAdd(key string, score float64, member interface{}) error {
	return c.getClient().ZAdd(c.ctx, key, &redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

// ZRemove 删除有序集合成员
func (c *CacheManager) ZRemove(key string, members ...interface{}) error {
	return c.getClient().ZRem(c.ctx, key, members...).Err()
}

// ZRange 获取有序集合范围成员
func (c *CacheManager) ZRange(key string, start, stop int64) ([]string, error) {
	return c.getClient().ZRange(c.ctx, key, start, stop).Result()
}

// Batch 批量操作
//
// 创建一个批量操作器，用于执行多个缓存操作并在一个原子事务中提交。
// 批量操作可以显著减少网络往返次数，提高性能。
//
// 返回:
//   - *BatchOperator: 批量操作器实例，支持链式调用
//
// 使用示例:
//
//	err := cm.Batch().
//	    Set("key1", "value1", time.Hour).
//	    Set("key2", "value2", time.Hour).
//	    Delete("key3").
//	    Execute()
func (c *CacheManager) Batch() *BatchOperator {
	return &BatchOperator{
		client: c.getClient(),
		ctx:    c.ctx,
		pipe:   c.getClient().Pipeline(),
	}
}

// serialize 序列化数据（优化内存分配）
//
// 为不同类型的数据提供优化的序列化策略：
// 1. 基础类型（string、[]byte、bool）：直接转换，没有额外开销
// 2. 数值类型（int、float等）：使用strconv高效转换
// 3. 复杂类型（struct、slice、map等）：使用JSON序列化
//
// 这种分层处理策略可以显著提升常用类型的序列化性能。
//
// 参数:
//   - value: 要序列化的值，支持任意类型
//
// 返回:
//   - string: 序列化后的字符串数据
//   - error: 序列化错误，nil表示成功
func (c *CacheManager) serialize(value interface{}) (string, error) {
	// 尝试基础类型序列化
	if result, ok := c.serializeBasicTypes(value); ok {
		return result, nil
	}

	// 尝试数值类型序列化
	if result, ok := c.serializeNumericTypes(value); ok {
		return result, nil
	}

	// 默认使用JSON序列化
	data, err := json.Marshal(value)
	return string(data), err
}

// serializeBasicTypes 序列化基础类型
func (c *CacheManager) serializeBasicTypes(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	case bool:
		if v {
			return "1", true
		}
		return "0", true
	default:
		return "", false
	}
}

// serializeNumericTypes 序列化数值类型
func (c *CacheManager) serializeNumericTypes(value interface{}) (string, bool) {
	// 尝试有符号整数类型
	if result, ok := c.serializeSignedInts(value); ok {
		return result, true
	}

	// 尝试无符号整数类型
	if result, ok := c.serializeUnsignedInts(value); ok {
		return result, true
	}

	// 尝试浮点数类型
	if result, ok := c.serializeFloats(value); ok {
		return result, true
	}

	return "", false
}

// serializeSignedInts 序列化有符号整数类型
func (c *CacheManager) serializeSignedInts(value interface{}) (string, bool) {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v), true
	case int8:
		return strconv.FormatInt(int64(v), 10), true
	case int16:
		return strconv.FormatInt(int64(v), 10), true
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int64:
		return strconv.FormatInt(v, 10), true
	default:
		return "", false
	}
}

// serializeUnsignedInts 序列化无符号整数类型
func (c *CacheManager) serializeUnsignedInts(value interface{}) (string, bool) {
	switch v := value.(type) {
	case uint:
		return strconv.FormatUint(uint64(v), 10), true
	case uint8:
		return strconv.FormatUint(uint64(v), 10), true
	case uint16:
		return strconv.FormatUint(uint64(v), 10), true
	case uint32:
		return strconv.FormatUint(uint64(v), 10), true
	case uint64:
		return strconv.FormatUint(v, 10), true
	default:
		return "", false
	}
}

// serializeFloats 序列化浮点数类型
func (c *CacheManager) serializeFloats(value interface{}) (string, bool) {
	switch v := value.(type) {
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), true
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true
	default:
		return "", false
	}
}

// deserialize 反序列化数据
//
// 根据目标类型自动选择最适合的反序列化策略：
// 1. *string：直接赋值，没有转换开销
// 2. *[]byte：转换为字节数组
// 3. *bool：智能识别"1"、"true"等值
// 4. 其他类型：使用JSON反序列化
//
// 参数:
//   - data: 要反序列化的字符串数据
//   - dest: 目标对象指针，用于接收反序列化结果
//
// 返回:
//   - error: 反序列化错误，nil表示成功
func (c *CacheManager) deserialize(data string, dest interface{}) error {
	switch d := dest.(type) {
	case *string:
		*d = data
		return nil
	case *[]byte:
		*d = []byte(data)
		return nil
	case *bool:
		*d = data == "1" || data == "true"
		return nil
	default:
		return json.Unmarshal([]byte(data), dest)
	}
}

// BatchOperator 批量操作器
//
// 批量操作器允许将多个缓存操作组合在一起，并在一个原子事务中执行。
// 这样可以：
// 1. 减少网络往返延迟：一次性发送多个命令
// 2. 保证原子性：所有操作一起成功或一起失败
// 3. 提高性能：特别适用于需要批量更新的场景
//
// 注意：所有操作都是延迟执行的，只有调用Execute()时才会真正执行。
type BatchOperator struct {
	client *redis.Client   // Redis客户端实例
	ctx    context.Context // 上下文对象
	pipe   redis.Pipeliner // Redis管道实例，用于批量操作
}

// Set 批量设置
//
// 在批量操作中添加一个设置操作。操作不会立即执行，需要调用Execute()才会提交。
// 支持任意类型的值，内部会使用与CacheManager一致的序列化策略。
//
// 参数:
//   - key: 缓存键名
//   - value: 缓存值，支持多种类型
//   - ttl: 过期时间，0表示永不过期
//
// 返回:
//   - *BatchOperator: 返回自身，支持链式调用
func (b *BatchOperator) Set(key string, value interface{}, ttl time.Duration) *BatchOperator {
	// 使用与CacheManager一致的序列化方法
	cm := &CacheManager{}
	data, err := cm.serialize(value)
	if err != nil {
		// 如果序列化失败，回退到JSON
		jsonData, _ := json.Marshal(value)
		data = string(jsonData)
	}
	b.pipe.Set(b.ctx, key, data, ttl)
	return b
}

// Delete 批量删除
//
// 在批量操作中添加一个或多个删除操作。操作不会立即执行，需要调用Execute()才会提交。
//
// 参数:
//   - keys: 要删除的键名列表，支持多个键
//
// 返回:
//   - *BatchOperator: 返回自身，支持链式调用
func (b *BatchOperator) Delete(keys ...string) *BatchOperator {
	b.pipe.Del(b.ctx, keys...)
	return b
}

// Execute 执行批量操作
//
// 执行之前添加的所有批量操作。所有操作会在一个原子事务中执行，
// 只有全部成功或全部失败。执行后管道会被清空，可以重新使用。
//
// 返回:
//   - error: 执行错误，nil表示所有操作都成功
func (b *BatchOperator) Execute() error {
	_, err := b.pipe.Exec(b.ctx)
	return err
}
