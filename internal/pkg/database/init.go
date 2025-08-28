package database

import (
	"fmt"
	"log"
)

// Init 初始化所有数据库连接
func Init() error {
	log.Println("Initializing database connections...")

	// 初始化MySQL连接池
	if err := InitMySQL(); err != nil {
		return fmt.Errorf("failed to initialize MySQL connection pool: %w", err)
	}

	// 初始化并发控制机制（包括 Redis 分布式锁）
	if err := InitConcurrencyControl(); err != nil {
		return fmt.Errorf("failed to initialize concurrency control: %w", err)
	}

	log.Println("Database initialization completed successfully")
	return nil
}

// Shutdown 优雅关闭所有数据库连接
func Shutdown() error {
	log.Println("Shutting down database connections...")

	// 关闭MySQL连接
	if err := Close(); err != nil {
		return fmt.Errorf("failed to close MySQL connection: %w", err)
	}

	log.Println("Database shutdown completed")
	return nil
}

// Status 检查所有数据库连接状态
func Status() map[string]interface{} {
	status := make(map[string]interface{})

	// MySQL连接状态
	if err := HealthCheck(); err != nil {
		status["mysql"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		status["mysql"] = map[string]interface{}{
			"status": "healthy",
			"stats":  GetConnectionStats(),
		}
	}

	// 迁移状态
	status["migration"] = CheckMigrationStatus()

	// 注册的模型数量
	status["models"] = map[string]interface{}{
		"registered_count": len(GetRegisteredModels()),
		"models":           GetRegisteredModels(),
	}

	return status
}
