package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("HXLOS Cloud Storage - 项目初始化成功!")
	fmt.Println("Go模块: cloudpan")
	fmt.Println("Go版本: 1.23.12")
	fmt.Println("Gin框架: 已添加")
	fmt.Println("Gorm ORM: 已添加")

	// 数据库连接示例（注释掉实际连接，仅作为依赖引用）
	// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	// if err != nil {
	//     fmt.Println("数据库连接失败:", err)
	// } else {
	//     fmt.Println("数据库连接成功")
	// }
	_ = mysql.Open // 引用mysql驱动确保依赖被保留
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
