package main

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
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

	// 引用确保依赖被保留
	_ = sql.Drivers   // 引用sql标准库
	_ = mysql.Open    // 引用mysql驱动确保依赖被保留
	_ = gorm.Config{} // 引用gorm确保依赖被保留

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
