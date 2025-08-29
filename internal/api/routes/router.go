package routes

import (
	"github.com/gin-gonic/gin"

	"cloudpan/internal/api/middleware"
	"cloudpan/internal/pkg/config"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 创建Gin引擎
	r := gin.New()

	// 添加基础中间件
	setupMiddleware(r)

	// 添加健康检查路由
	setupHealthRoutes(r)

	// 添加API路由
	setupAPIRoutes(r)

	return r
}

// setupMiddleware 设置中间件
func setupMiddleware(r *gin.Engine) {
	// 基础中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 请求ID中间件
	r.Use(middleware.RequestIDMiddleware())

	// 请求日志中间件
	r.Use(middleware.RequestLogger())

	// 错误处理中间件
	r.Use(middleware.ErrorHandler())

	// CORS中间件
	if config.AppConfig.App.Debug {
		// 开发环境允许所有源
		r.Use(middleware.CORSMiddleware())
	} else {
		// 生产环境限制源
		allowedOrigins := []string{
			"https://cloudpan.hxlos.com",
			"https://www.hxlos.com",
		}
		r.Use(middleware.ProductionCORS(allowedOrigins))
	}

	// API版本管理中间件
	r.Use(middleware.APIVersionMiddleware())

	// 国际化中间件
	i18nConfig := middleware.DefaultI18nConfig()
	i18nConfig.TranslationPath = "locales"
	r.Use(middleware.I18nMiddleware(i18nConfig))
}

// setupHealthRoutes 设置健康检查路由
func setupHealthRoutes(r *gin.Engine) {
	r.GET("/health", HealthCheckHandler)
	r.GET("/health/database", DatabaseHealthHandler)
}

// setupAPIRoutes 设置API路由
func setupAPIRoutes(r *gin.Engine) {
	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 系统信息
		v1.GET("/system/stats", SystemStatsHandler)
		v1.GET("/system/version", middleware.VersionInfoHandler())
		v1.GET("/system/language", middleware.LanguageInfoHandler())

		// 预留其他业务路由
		setupUserRoutes(v1)
		setupFileRoutes(v1)
		setupTeamRoutes(v1)
		setupMessageRoutes(v1)
	}

	// API v2 路由组（预留）
	v2 := r.Group("/api/v2")
	{
		v2.GET("/system/stats", SystemStatsHandler)
		v2.GET("/system/version", middleware.VersionInfoHandler())
		v2.GET("/system/language", middleware.LanguageInfoHandler())
	}
}

// setupUserRoutes 设置用户相关路由
func setupUserRoutes(rg *gin.RouterGroup) {
	// 认证相关路由（不需要认证）
	auth := rg.Group("/auth")
	{
		auth.POST("/register", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "用户注册接口 - 待实现"})
		})
		auth.POST("/send-code", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "发送验证码接口 - 待实现"})
		})
		auth.POST("/login", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "用户登录接口 - 待实现"})
		})
		auth.POST("/refresh", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Token刷新接口 - 待实现"})
		})
		auth.POST("/forgot-password", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "忘记密码接口 - 待实现"})
		})
		auth.POST("/reset-password", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "重置密码接口 - 待实现"})
		})
	}

	// 用户管理路由（需要认证）
	users := rg.Group("/users")
	{
		// 预留用户路由
		users.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "用户列表接口 - 待实现"})
		})
		users.GET("/profile", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "获取用户信息接口 - 待实现"})
		})
		users.PUT("/profile", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "更新用户信息接口 - 待实现"})
		})
		users.POST("/change-password", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "修改密码接口 - 待实现"})
		})
		users.GET("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "获取用户详情接口 - 待实现"})
		})
		users.PUT("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "更新用户接口 - 待实现"})
		})
		users.DELETE("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "删除用户接口 - 待实现"})
		})
	}
}

// setupFileRoutes 设置文件相关路由
func setupFileRoutes(rg *gin.RouterGroup) {
	files := rg.Group("/files")
	{
		// 预留文件路由
		files.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "文件列表接口 - 待实现"})
		})
		files.POST("/upload", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "文件上传接口 - 待实现"})
		})
		files.GET("/:id/download", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "文件下载接口 - 待实现"})
		})
		files.DELETE("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "删除文件接口 - 待实现"})
		})
	}
}

// setupTeamRoutes 设置团队相关路由
func setupTeamRoutes(rg *gin.RouterGroup) {
	teams := rg.Group("/teams")
	{
		// 预留团队路由
		teams.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "团队列表接口 - 待实现"})
		})
		teams.POST("", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "创建团队接口 - 待实现"})
		})
		teams.GET("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "获取团队详情接口 - 待实现"})
		})
	}
}

// setupMessageRoutes 设置消息相关路由
func setupMessageRoutes(rg *gin.RouterGroup) {
	messages := rg.Group("/messages")
	{
		// 预留消息路由
		messages.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "消息列表接口 - 待实现"})
		})
		messages.POST("", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "发送消息接口 - 待实现"})
		})
		messages.PUT("/:id/read", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "标记消息已读接口 - 待实现"})
		})
	}
}
