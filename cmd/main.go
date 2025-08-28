package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"cloudpan/internal/pkg/config"
	"cloudpan/internal/pkg/database"
)

func main() {
	fmt.Println("HXLOS Cloud Storage - 启动中...")

	// 1. 加载配置文件
	log.Println("Loading configuration...")
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println("Configuration loaded successfully")

	// 2. 初始化数据库连接池
	log.Println("Initializing database connections...")
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database connections initialized successfully")

	// 3. 设置Gin模式
	if !config.AppConfig.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 4. 创建Gin引擎
	r := gin.Default()

	// 5. 添加健康检查接口
	r.GET("/health", healthCheckHandler)
	r.GET("/health/database", databaseHealthHandler)
	r.GET("/api/v1/system/stats", systemStatsHandler)

	// 6. 创建HTTP服务器
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port),
		Handler:        r,
		ReadTimeout:    config.AppConfig.Server.ReadTimeout,
		WriteTimeout:   config.AppConfig.Server.WriteTimeout,
		MaxHeaderBytes: config.AppConfig.Server.MaxHeaderBytes,
	}

	// 7. 启动服务器（在goroutine中）
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("HXLOS Cloud Storage started successfully on %s", srv.Addr)
	log.Printf("Environment: %s, Debug: %v", config.AppConfig.App.Env, config.AppConfig.App.Debug)

	// 8. 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 9. 优雅关闭服务器，等待现有连接完成
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// 10. 关闭数据库连接
	if err := database.Shutdown(); err != nil {
		log.Printf("Failed to shutdown database: %v", err)
	}

	log.Println("Server exited")

	// 确保依赖被保留（防止go mod tidy移除）
	_ = sql.Drivers
	_ = mysql.Open
	_ = gorm.Config{}
	_ = &redis.Options{}
	_ = jwt.SigningMethodHS256
	_ = context.TODO
}

// healthCheckHandler 基础健康检查处理器
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "HXLOS Cloud Storage Service is running",
		"module":    "cloudpan",
		"version":   config.AppConfig.App.Version,
		"timestamp": time.Now().Unix(),
	})
}

// databaseHealthHandler 数据库健康检查处理器
func databaseHealthHandler(c *gin.Context) {
	status := database.Status()
	statusCode := http.StatusOK

	// 检查是否有不健康的数据库连接
	for _, dbStatus := range status {
		if dbInfo, ok := dbStatus.(map[string]interface{}); ok {
			if dbInfo["status"] != "healthy" {
				statusCode = http.StatusServiceUnavailable
				break
			}
		}
	}

	c.JSON(statusCode, gin.H{
		"status":    "ok",
		"databases": status,
		"timestamp": time.Now().Unix(),
	})
}

// systemStatsHandler 系统统计信息处理器
func systemStatsHandler(c *gin.Context) {
	stats := gin.H{
		"application": gin.H{
			"name":    config.AppConfig.App.Name,
			"version": config.AppConfig.App.Version,
			"env":     config.AppConfig.App.Env,
			"debug":   config.AppConfig.App.Debug,
		},
		"server": gin.H{
			"host":             config.AppConfig.Server.Host,
			"port":             config.AppConfig.Server.Port,
			"read_timeout":     config.AppConfig.Server.ReadTimeout.String(),
			"write_timeout":    config.AppConfig.Server.WriteTimeout.String(),
			"max_header_bytes": config.AppConfig.Server.MaxHeaderBytes,
		},
		"database":  database.Status(),
		"timestamp": time.Now().Unix(),
	}

	c.JSON(http.StatusOK, stats)
}
