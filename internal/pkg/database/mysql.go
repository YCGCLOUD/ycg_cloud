package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"cloudpan/internal/pkg/config"
)

var (
	// DB 全局数据库实例
	DB *gorm.DB
)

// InitMySQL 初始化MySQL连接池
func InitMySQL() error {
	cfg := config.AppConfig.Database.MySQL

	// 构建DSN
	dsn := buildDSN(cfg)

	// 配置GORM日志
	var gormLogger logger.Interface
	if config.AppConfig.App.Debug {
		gormLogger = NewCustomLogger(200*time.Millisecond, logger.Info)
	} else {
		gormLogger = NewCustomLogger(500*time.Millisecond, logger.Warn)
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束创建，手动管理
		SkipDefaultTransaction:                   true, // 提高性能：跳过默认事务
		PrepareStmt:                              true, // 缓存预编译语句，提高性能
	})

	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// 获取底层sql.DB实例
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 配置连接池参数
	if err := configureConnectionPool(sqlDB, cfg); err != nil {
		return fmt.Errorf("failed to configure connection pool: %w", err)
	}

	// 测试连接
	if err := testConnection(sqlDB); err != nil {
		return fmt.Errorf("database connection test failed: %w", err)
	}

	// 设置全局DB实例
	DB = db

	// 安装默认插件
	if err := InstallPlugins(db, GetDefaultPlugins()...); err != nil {
		log.Printf("Warning: failed to install some plugins: %v", err)
	}

	// 执行自动迁移（如果配置开启）
	if config.AppConfig.App.Debug {
		if err := AutoMigrate(); err != nil {
			log.Printf("Warning: auto migration failed: %v", err)
		}
	}

	log.Printf("MySQL connected successfully: %s:%d/%s", cfg.Host, cfg.Port, cfg.DBName)
	log.Printf("Connection pool configured - MaxOpen: %d, MaxIdle: %d, MaxLifetime: %v",
		cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime)

	return nil
}

// buildDSN 构建MySQL连接字符串
func buildDSN(cfg config.MySQLConfig) string {
	// MySQL 8.0.31兼容配置
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s&allowNativePasswords=true",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.Charset,
		cfg.ParseTime,
		cfg.Loc,
	)

	// 添加时区配置（如果设置了）
	if cfg.Timezone != "" {
		dsn += "&time_zone=" + cfg.Timezone
	}

	return dsn
}

// configureConnectionPool 配置连接池参数
func configureConnectionPool(sqlDB sqlDB, cfg config.MySQLConfig) error {
	// 设置最大打开连接数
	// 默认值：100，防止连接数过多导致数据库压力
	maxOpenConns := cfg.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 100
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)

	// 设置最大空闲连接数
	// 默认值：10，保持足够的空闲连接以提高响应速度
	maxIdleConns := cfg.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 10
	}
	// 确保空闲连接数不超过最大连接数
	if maxIdleConns > maxOpenConns {
		maxIdleConns = maxOpenConns
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)

	// 设置连接最大生存时间
	// 默认值：1小时，防止长时间连接被MySQL服务器关闭
	connMaxLifetime := cfg.ConnMaxLifetime
	if connMaxLifetime <= 0 {
		connMaxLifetime = time.Hour
	}
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// 设置连接最大空闲时间
	// 默认值：10分钟，及时释放不活跃的连接
	connMaxIdleTime := cfg.ConnMaxIdleTime
	if connMaxIdleTime <= 0 {
		connMaxIdleTime = 10 * time.Minute
	}
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	return nil
}

// testConnection 测试数据库连接
func testConnection(sqlDB sqlDB) error {
	// 设置连接测试超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 执行ping测试
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database failed: %w", err)
	}

	return nil
}

// GetDB 获取数据库连接实例
func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("Database not initialized. Call InitMySQL() first")
	}
	return DB
}

// HealthCheck 健康检查
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetConnectionStats 获取连接池统计信息
func GetConnectionStats() map[string]interface{} {
	if DB == nil {
		return map[string]interface{}{
			"error": "database not initialized",
		}
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("failed to get underlying sql.DB: %v", err),
		}
	}

	stats := sqlDB.Stats()

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Println("Database connection closed")
	return nil
}

// sqlDB 接口定义，便于测试
type sqlDB interface {
	SetMaxOpenConns(n int)
	SetMaxIdleConns(n int)
	SetConnMaxLifetime(d time.Duration)
	SetConnMaxIdleTime(d time.Duration)
	PingContext(ctx context.Context) error
	Stats() sql.DBStats
	Close() error
}
