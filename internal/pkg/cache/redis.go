package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloudpan/internal/pkg/config"

	"github.com/go-redis/redis/v8"
)

// Redis连接管理器
var (
	RedisClient *redis.Client
)

// InitRedis 初始化Redis连接
func InitRedis() error {
	if config.AppConfig == nil {
		return fmt.Errorf("config not initialized")
	}

	cfg := config.AppConfig.Redis

	// 创建Redis客户端
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolTimeout:  cfg.PoolTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Redis connected successfully: %s:%d", cfg.Host, cfg.Port)
	return nil
}

// GetRedisClient 获取Redis客户端
func GetRedisClient() *redis.Client {
	if RedisClient == nil {
		log.Fatal("Redis not initialized. Call InitRedis() first")
	}
	return RedisClient
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if RedisClient == nil {
		return nil
	}
	return RedisClient.Close()
}

// HealthCheck 检查Redis健康状态
func HealthCheck() error {
	if RedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return RedisClient.Ping(ctx).Err()
}

// GetConnectionStats 获取连接统计信息
func GetConnectionStats() map[string]interface{} {
	if RedisClient == nil {
		return map[string]interface{}{
			"status": "not_initialized",
		}
	}

	stats := RedisClient.PoolStats()
	return map[string]interface{}{
		"status":      "connected",
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"timeouts":    stats.Timeouts,
		"total_conns": stats.TotalConns,
		"idle_conns":  stats.IdleConns,
		"stale_conns": stats.StaleConns,
	}
}
