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

	"cloudpan/internal/api/routes"
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

	// 4. 设置路由
	r := routes.SetupRouter()

	// 5. 创建HTTP服务器
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port),
		Handler:        r,
		ReadTimeout:    config.AppConfig.Server.ReadTimeout,
		WriteTimeout:   config.AppConfig.Server.WriteTimeout,
		MaxHeaderBytes: config.AppConfig.Server.MaxHeaderBytes,
	}

	// 6. 启动服务器（在goroutine中）
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("HXLOS Cloud Storage started successfully on %s", srv.Addr)
	log.Printf("Environment: %s, Debug: %v", config.AppConfig.App.Env, config.AppConfig.App.Debug)

	// 7. 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 8. 优雅关闭服务器，等待现有连接完成
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// 9. 关闭数据库连接
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
