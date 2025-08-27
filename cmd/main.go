package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("HXLOS Cloud Storage - 项目初始化成功!")
	fmt.Println("Go模块: cloudpan")
	fmt.Println("Go版本: 1.23.12")
	fmt.Println("Gin框架: 已添加")
	fmt.Println("Gorm ORM: 已添加")
	fmt.Println("MySQL驱动: 已添加 (支持MySQL 8.0.31)")
	fmt.Println("Redis客户端: 已添加 (支持Redis 7.0.6)")
	fmt.Println("JWT认证: 已添加 (支持golang-jwt/jwt/v5)")

	// MySQL 8.0.31 直接连接示例（注释掉实际连接）
	// MySQL 8.0.31 连接字符串示例：
	// dsn := "username:password@tcp(localhost:3306)/database_name?charset=utf8mb4&parseTime=True&loc=Local&allowNativePasswords=true"
	// db, err := sql.Open("mysql", dsn)
	// if err != nil {
	//     fmt.Println("MySQL 8.0.31 直接连接失败:", err)
	// } else {
	//     defer db.Close()
	//     fmt.Println("MySQL 8.0.31 直接连接成功")
	// }

	// Gorm + MySQL 8.0.31 连接示例（注释掉实际连接）
	// gormDsn := "username:password@tcp(localhost:3306)/database_name?charset=utf8mb4&parseTime=True&loc=Local"
	// gormDB, err := gorm.Open(mysql.Open(gormDsn), &gorm.Config{})
	// if err != nil {
	//     fmt.Println("Gorm MySQL 8.0.31 连接失败:", err)
	// } else {
	//     fmt.Println("Gorm MySQL 8.0.31 连接成功")
	// }

	// Redis 7.0.6 客户端连接示例（注释掉实际连接）
	// Redis 7.0.6 连接配置：
	// rdb := redis.NewClient(&redis.Options{
	//     Addr:     "localhost:6379",
	//     Password: "", // 无密码
	//     DB:       0,  // 默认数据库
	//     Protocol: 3,  // Redis 7.0.6 支持RESP3协议
	// })
	// ctx := context.Background()
	// pong, err := rdb.Ping(ctx).Result()
	// if err != nil {
	//     fmt.Println("Redis 7.0.6 连接失败:", err)
	// } else {
	//     fmt.Println("Redis 7.0.6 连接成功:", pong)
	// }

	// Redis Stream 消息队列示例（支持Redis 7.0.6新特性）
	// streamName := "cloudpan:events"
	// 添加消息到Stream
	// xaddResult := rdb.XAdd(ctx, &redis.XAddArgs{
	//     Stream: streamName,
	//     Values: map[string]interface{}{
	//         "event": "file_upload",
	//         "user_id": "123",
	//         "timestamp": time.Now().Unix(),
	//     },
	// })
	// fmt.Println("Redis Stream 消息添加:", xaddResult.Val())

	// JWT认证示例（golang-jwt/jwt/v5）
	// JWT秘钥和配置：
	// jwtSecret := []byte("cloudpan-jwt-secret-key")
	// 生成JWT Token
	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	//     "user_id": "123",
	//     "username": "admin",
	//     "exp": time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
	//     "iat": time.Now().Unix(),
	// })
	// tokenString, err := token.SignedString(jwtSecret)
	// if err != nil {
	//     fmt.Println("JWT Token生成失败:", err)
	// } else {
	//     fmt.Println("JWT Token生成成功:", tokenString[:50]+"...")
	// }
	// 解析JWT Token
	// claims := jwt.MapClaims{}
	// parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
	//     return jwtSecret, nil
	// })
	// if err != nil {
	//     fmt.Println("JWT Token解析失败:", err)
	// } else if parsedToken.Valid {
	//     fmt.Println("JWT Token解析成功:", claims["username"])
	// }

	// 引用确保依赖被保留
	_ = sql.Drivers            // 引用sql标准库
	_ = mysql.Open             // 引用mysql驱动确保依赖被保留
	_ = gorm.Config{}          // 引用gorm确保依赖被保留
	_ = &redis.Options{}       // 引用redis客户端确保依赖被保留
	_ = jwt.SigningMethodHS256 // 引用JWT确保依赖被保留
	_ = context.TODO           // 引用context包

	// 创建Gin引擎
	r := gin.Default()

	// 添加健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "HXLOS Cloud Storage Service is running",
			"module":  "cloudpan",
		})
	})

	fmt.Println("服务启动在 :8080 端口")
	r.Run(":8080")
}
