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
type CacheManager struct {
	client *redis.Client
	ctx    context.Context
}

// NewCacheManager 创建缓存管理器
func NewCacheManager() *CacheManager {
	return &CacheManager{
		client: nil, // 延迟初始化，在第一次使用时获取
		ctx:    context.Background(),
	}
}

// getClient 获取Redis客户端（延迟初始化）
func (c *CacheManager) getClient() *redis.Client {
	if c.client == nil {
		c.client = GetRedisClient()
	}
	return c.client
}

// Set 设置缓存，使用默认TTL
func (c *CacheManager) Set(key string, value interface{}) error {
	return c.SetWithTTL(key, value, config.AppConfig.Cache.DefaultTTL)
}

// SetWithTTL 设置缓存，指定TTL
func (c *CacheManager) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	data, err := c.serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	return c.getClient().Set(c.ctx, key, data, ttl).Err()
}

// Get 获取缓存
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
func (c *CacheManager) Delete(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.getClient().Del(c.ctx, keys...).Err()
}

// Exists 检查缓存是否存在
func (c *CacheManager) Exists(keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	return c.getClient().Exists(c.ctx, keys...).Result()
}

// Expire 设置缓存过期时间
func (c *CacheManager) Expire(key string, ttl time.Duration) error {
	return c.getClient().Expire(c.ctx, key, ttl).Err()
}

// TTL 获取缓存剩余过期时间
func (c *CacheManager) TTL(key string) (time.Duration, error) {
	return c.getClient().TTL(c.ctx, key).Result()
}

// Increment 原子递增
func (c *CacheManager) Increment(key string) (int64, error) {
	return c.getClient().Incr(c.ctx, key).Result()
}

// IncrementBy 原子递增指定值
func (c *CacheManager) IncrementBy(key string, value int64) (int64, error) {
	return c.getClient().IncrBy(c.ctx, key, value).Result()
}

// Decrement 原子递减
func (c *CacheManager) Decrement(key string) (int64, error) {
	return c.getClient().Decr(c.ctx, key).Result()
}

// DecrementBy 原子递减指定值
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
func (c *CacheManager) Batch() *BatchOperator {
	return &BatchOperator{
		client: c.getClient(),
		ctx:    c.ctx,
		pipe:   c.getClient().Pipeline(),
	}
}

// serialize 序列化数据（优化内存分配）
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
type BatchOperator struct {
	client *redis.Client
	ctx    context.Context
	pipe   redis.Pipeliner
}

// Set 批量设置
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
func (b *BatchOperator) Delete(keys ...string) *BatchOperator {
	b.pipe.Del(b.ctx, keys...)
	return b
}

// Execute 执行批量操作
func (b *BatchOperator) Execute() error {
	_, err := b.pipe.Exec(b.ctx)
	return err
}
