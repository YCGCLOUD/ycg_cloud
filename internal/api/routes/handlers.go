package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"cloudpan/internal/api/middleware"
	"cloudpan/internal/pkg/config"
	"cloudpan/internal/pkg/database"
)

// HealthCheckHandler 基础健康检查处理器
func HealthCheckHandler(c *gin.Context) {
	response := gin.H{
		"status":      "ok",
		"message":     middleware.T(c, "common.success"),
		"module":      "cloudpan",
		"version":     config.AppConfig.App.Version,
		"timestamp":   time.Now().Unix(),
		"language":    middleware.GetLanguage(c),
		"api_version": middleware.GetAPIVersion(c),
	}

	c.JSON(http.StatusOK, response)
}

// DatabaseHealthHandler 数据库健康检查处理器
func DatabaseHealthHandler(c *gin.Context) {
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

	response := gin.H{
		"status":      "ok",
		"message":     middleware.T(c, "common.success"),
		"databases":   status,
		"timestamp":   time.Now().Unix(),
		"language":    middleware.GetLanguage(c),
		"api_version": middleware.GetAPIVersion(c),
	}

	c.JSON(statusCode, response)
}

// SystemStatsHandler 系统统计信息处理器
func SystemStatsHandler(c *gin.Context) {
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
		"database":    database.Status(),
		"timestamp":   time.Now().Unix(),
		"language":    middleware.GetLanguage(c),
		"api_version": middleware.GetAPIVersion(c),
	}

	response := gin.H{
		"code":      200,
		"message":   middleware.T(c, "common.success"),
		"data":      stats,
		"timestamp": time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}
