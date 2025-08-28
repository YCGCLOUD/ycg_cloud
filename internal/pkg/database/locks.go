package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"cloudpan/internal/pkg/config"
)

// validateTableName 验证表名安全性，防止SQL注入
func validateTableName(tableName string) error {
	// 检查表名是否为空
	if strings.TrimSpace(tableName) == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// 使用正则表达式验证表名格式（只允许字母、数字和下划线）
	tableNameRegex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !tableNameRegex.MatchString(tableName) {
		return fmt.Errorf("invalid table name format: %s", tableName)
	}

	// 检查表名长度
	if len(tableName) > 64 {
		return fmt.Errorf("table name too long: %s", tableName)
	}

	// 检查是否包含SQL关键字（简单黑名单）
	sqlKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
		"UNION", "WHERE", "ORDER", "GROUP", "HAVING", "EXEC", "EXECUTE",
	}
	upper := strings.ToUpper(tableName)
	for _, keyword := range sqlKeywords {
		if upper == keyword {
			return fmt.Errorf("table name cannot be SQL keyword: %s", tableName)
		}
	}

	return nil
}

// IsolationLevel 事务隔离级别枚举
type IsolationLevel int

const (
	ReadUncommitted IsolationLevel = iota
	ReadCommitted
	RepeatableRead
	Serializable
)

// String 返回隔离级别字符串
func (il IsolationLevel) String() string {
	switch il {
	case ReadUncommitted:
		return "READ UNCOMMITTED"
	case ReadCommitted:
		return "READ COMMITTED"
	case RepeatableRead:
		return "REPEATABLE READ"
	case Serializable:
		return "SERIALIZABLE"
	default:
		return "REPEATABLE READ"
	}
}

// LockType 锁类型枚举
type LockType int

const (
	SharedLock    LockType = iota // 共享锁
	ExclusiveLock                 // 排他锁
)

// String 返回锁类型字符串
func (lt LockType) String() string {
	switch lt {
	case SharedLock:
		return "LOCK IN SHARE MODE"
	case ExclusiveLock:
		return "FOR UPDATE"
	default:
		return "FOR UPDATE"
	}
}

// TransactionManager 事务管理器
type TransactionManager struct {
	db             *gorm.DB
	defaultLevel   IsolationLevel
	defaultTimeout time.Duration
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{
		db:             db,
		defaultLevel:   RepeatableRead,
		defaultTimeout: 30 * time.Second,
	}
}

// WithIsolationLevel 设置事务隔离级别
func (tm *TransactionManager) WithIsolationLevel(level IsolationLevel, fn func(tx *gorm.DB) error) error {
	return tm.db.Transaction(func(tx *gorm.DB) error {
		// 设置事务隔离级别
		if err := tx.Exec(fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", level.String())).Error; err != nil {
			return fmt.Errorf("failed to set isolation level: %w", err)
		}
		return fn(tx)
	}, &sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  false,
	})
}

// WithReadOnlyTransaction 只读事务
func (tm *TransactionManager) WithReadOnlyTransaction(fn func(tx *gorm.DB) error) error {
	return tm.db.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	}, &sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  true,
	})
}

// DatabaseLockManager 数据库锁管理器
type DatabaseLockManager struct {
	db *gorm.DB
}

// NewDatabaseLockManager 创建数据库锁管理器
func NewDatabaseLockManager(db *gorm.DB) *DatabaseLockManager {
	return &DatabaseLockManager{db: db}
}

// AcquirePessimisticLock 获取悲观锁
// 注意：tableName 必须是可信的表名，不能来自用户输入
func (dlm *DatabaseLockManager) AcquirePessimisticLock(ctx context.Context, tx *gorm.DB, tableName string, lockType LockType, where string, args ...interface{}) error {
	// 验证表名安全性（防止SQL注入）
	if err := validateTableName(tableName); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	// 使用GORM的查询构建器而不是直接字符串拼接
	var result int
	query := tx.WithContext(ctx).Table(tableName).Select("1").Where(where, args...)

	// 添加锁提示（使用tagged switch优化）
	switch lockType {
	case SharedLock:
		query = query.Set("gorm:query_option", "FOR SHARE")
	case ExclusiveLock:
		query = query.Set("gorm:query_option", "FOR UPDATE")
	}

	if err := query.Scan(&result).Error; err != nil {
		return fmt.Errorf("failed to acquire pessimistic lock: %w", err)
	}

	log.Printf("Acquired pessimistic lock on table %s with condition: %s", tableName, where)
	return nil
}

// PessimisticLockQuery 带悲观锁的查询
func (dlm *DatabaseLockManager) PessimisticLockQuery(tx *gorm.DB, model interface{}, lockType LockType, where ...interface{}) error {
	query := tx.Set("gorm:query_option", lockType.String())

	if len(where) > 0 {
		query = query.Where(where[0], where[1:]...)
	}

	return query.First(model).Error
}

// OptimisticLockUpdate 乐观锁更新
func (dlm *DatabaseLockManager) OptimisticLockUpdate(tx *gorm.DB, model interface{}, version int64, updates map[string]interface{}) error {
	// 添加版本号递增
	updates["version"] = version + 1

	result := tx.Model(model).Where("version = ?", version).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("optimistic lock update failed: %w", result.Error)
	}

	// 检查是否有记录被更新
	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock conflict: record has been modified by another process")
	}

	return nil
}

// OptimisticLockDelete 乐观锁删除
func (dlm *DatabaseLockManager) OptimisticLockDelete(tx *gorm.DB, model interface{}, version int64) error {
	result := tx.Where("version = ?", version).Delete(model)
	if result.Error != nil {
		return fmt.Errorf("optimistic lock delete failed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock conflict: record has been modified or deleted by another process")
	}

	return nil
}

// RedisDistributedLock Redis分布式锁
type RedisDistributedLock struct {
	client *redis.Client
	key    string
	value  string
	ttl    time.Duration
}

// RedisLockManager Redis分布式锁管理器
type RedisLockManager struct {
	client *redis.Client
}

// NewRedisLockManager 创建Redis分布式锁管理器
func NewRedisLockManager(client *redis.Client) *RedisLockManager {
	return &RedisLockManager{
		client: client,
	}
}

// NewLock 创建分布式锁
func (rlm *RedisLockManager) NewLock(key string, ttl time.Duration) (*RedisDistributedLock, error) {
	// 生成随机值作为锁的标识
	value, err := generateRandomValue(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate lock value: %w", err)
	}

	return &RedisDistributedLock{
		client: rlm.client,
		key:    fmt.Sprintf("lock:%s", key),
		value:  value,
		ttl:    ttl,
	}, nil
}

// TryLock 尝试获取锁
func (rdl *RedisDistributedLock) TryLock(ctx context.Context) (bool, error) {
	result, err := rdl.client.SetNX(ctx, rdl.key, rdl.value, rdl.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if result {
		log.Printf("Acquired distributed lock: %s", rdl.key)
	}

	return result, nil
}

// Lock 阻塞获取锁
func (rdl *RedisDistributedLock) Lock(ctx context.Context, retryInterval time.Duration) error {
	for {
		acquired, err := rdl.TryLock(ctx)
		if err != nil {
			return err
		}

		if acquired {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
			// 继续重试
		}
	}
}

// Unlock 释放锁
func (rdl *RedisDistributedLock) Unlock(ctx context.Context) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := rdl.client.Eval(ctx, script, []string{rdl.key}, rdl.value).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result.(int64) == 1 {
		log.Printf("Released distributed lock: %s", rdl.key)
	} else {
		log.Printf("Lock %s was not owned by this instance", rdl.key)
	}

	return nil
}

// Extend 延长锁的过期时间
func (rdl *RedisDistributedLock) Extend(ctx context.Context, newTTL time.Duration) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := rdl.client.Eval(ctx, script, []string{rdl.key}, rdl.value, int64(newTTL.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	if result.(int64) == 1 {
		rdl.ttl = newTTL
		log.Printf("Extended distributed lock: %s, new TTL: %v", rdl.key, newTTL)
	}

	return nil
}

// IsLocked 检查锁是否仍然有效
func (rdl *RedisDistributedLock) IsLocked(ctx context.Context) (bool, error) {
	value, err := rdl.client.Get(ctx, rdl.key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return value == rdl.value, nil
}

// LockWithAutoRenewal 带自动续期的锁
func (rdl *RedisDistributedLock) LockWithAutoRenewal(ctx context.Context, renewalInterval time.Duration) error {
	// 首先获取锁
	if err := rdl.Lock(ctx, 100*time.Millisecond); err != nil {
		return err
	}

	// 启动自动续期协程
	go func() {
		ticker := time.NewTicker(renewalInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := rdl.Extend(ctx, rdl.ttl); err != nil {
					log.Printf("Failed to renew lock %s: %v", rdl.key, err)
					return
				}
			}
		}
	}()

	return nil
}

// ConcurrencyControlManager 并发控制管理器
type ConcurrencyControlManager struct {
	txManager    *TransactionManager
	dbLockMgr    *DatabaseLockManager
	redisLockMgr *RedisLockManager
}

// NewConcurrencyControlManager 创建并发控制管理器
func NewConcurrencyControlManager(db *gorm.DB, redisClient *redis.Client) *ConcurrencyControlManager {
	return &ConcurrencyControlManager{
		txManager:    NewTransactionManager(db),
		dbLockMgr:    NewDatabaseLockManager(db),
		redisLockMgr: NewRedisLockManager(redisClient),
	}
}

// WithDistributedLock 使用分布式锁执行操作
func (ccm *ConcurrencyControlManager) WithDistributedLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	lock, err := ccm.redisLockMgr.NewLock(key, ttl)
	if err != nil {
		return err
	}

	if err := lock.Lock(ctx, 100*time.Millisecond); err != nil {
		return err
	}
	defer func() {
		if unlockErr := lock.Unlock(ctx); unlockErr != nil {
			log.Printf("Failed to unlock distributed lock %s: %v", key, unlockErr)
		}
	}()

	return fn()
}

// WithPessimisticLock 使用悲观锁执行事务
func (ccm *ConcurrencyControlManager) WithPessimisticLock(ctx context.Context, tableName string, lockType LockType, where string, fn func(tx *gorm.DB) error, args ...interface{}) error {
	return ccm.txManager.db.Transaction(func(tx *gorm.DB) error {
		// 获取悲观锁
		if err := ccm.dbLockMgr.AcquirePessimisticLock(ctx, tx, tableName, lockType, where, args...); err != nil {
			return err
		}

		// 执行业务逻辑
		return fn(tx)
	})
}

// WithOptimisticLock 使用乐观锁执行更新
func (ccm *ConcurrencyControlManager) WithOptimisticLock(model interface{}, version int64, updates map[string]interface{}) error {
	return ccm.txManager.db.Transaction(func(tx *gorm.DB) error {
		return ccm.dbLockMgr.OptimisticLockUpdate(tx, model, version, updates)
	})
}

// generateRandomValue 生成随机值
func generateRandomValue(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

// 全局锁管理器实例
var (
	GlobalConcurrencyManager *ConcurrencyControlManager
)

// InitConcurrencyControl 初始化并发控制
func InitConcurrencyControl() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.AppConfig.Redis.Host, config.AppConfig.Redis.Port),
		Password:     config.AppConfig.Redis.Password,
		DB:           config.AppConfig.Redis.DB,
		PoolSize:     config.AppConfig.Redis.PoolSize,
		MinIdleConns: config.AppConfig.Redis.MinIdleConns,
		MaxRetries:   config.AppConfig.Redis.MaxRetries,
		DialTimeout:  config.AppConfig.Redis.DialTimeout,
		ReadTimeout:  config.AppConfig.Redis.ReadTimeout,
		WriteTimeout: config.AppConfig.Redis.WriteTimeout,
		PoolTimeout:  config.AppConfig.Redis.PoolTimeout,
	})

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// 初始化全局并发控制管理器
	GlobalConcurrencyManager = NewConcurrencyControlManager(DB, redisClient)

	log.Println("Concurrency control initialized successfully")
	return nil
}

// GetConcurrencyManager 获取并发控制管理器
func GetConcurrencyManager() *ConcurrencyControlManager {
	if GlobalConcurrencyManager == nil {
		log.Fatal("Concurrency control not initialized. Call InitConcurrencyControl() first")
	}
	return GlobalConcurrencyManager
}
