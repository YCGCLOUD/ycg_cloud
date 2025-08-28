package middleware

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"cloudpan/internal/pkg/utils"
)

// APIVersionConfig API版本配置
type APIVersionConfig struct {
	DefaultVersion    string            // 默认版本
	SupportedVersions []string          // 支持的版本列表
	VersionHeader     string            // 版本头名称
	VersionParam      string            // 版本参数名
	VersionPrefix     string            // URL版本前缀
	DeprecatedMap     map[string]string // 已弃用版本映射
}

// DefaultAPIVersionConfig 默认API版本配置
func DefaultAPIVersionConfig() *APIVersionConfig {
	return &APIVersionConfig{
		DefaultVersion:    "v1",
		SupportedVersions: []string{"v1", "v2"},
		VersionHeader:     "API-Version",
		VersionParam:      "version",
		VersionPrefix:     "/api/",
		DeprecatedMap: map[string]string{
			"v1": "v2", // v1已弃用，建议使用v2
		},
	}
}

// APIVersionMiddleware 创建API版本管理中间件
func APIVersionMiddleware(config ...*APIVersionConfig) gin.HandlerFunc {
	var cfg *APIVersionConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = DefaultAPIVersionConfig()
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		version := extractAPIVersion(c, cfg)

		// 验证版本是否支持
		if !isVersionSupported(version, cfg.SupportedVersions) {
			utils.ErrorWithMessage(c, utils.CodeValidationError,
				fmt.Sprintf("不支持的API版本: %s，支持的版本: %s",
					version, strings.Join(cfg.SupportedVersions, ", ")))
			return
		}

		// 设置版本信息到上下文
		c.Set("api_version", version)
		c.Set("api_version_config", cfg)

		// 检查版本弃用警告
		if recommendedVersion, deprecated := cfg.DeprecatedMap[version]; deprecated {
			c.Header("API-Deprecated", "true")
			c.Header("API-Deprecated-Version", version)
			c.Header("API-Recommended-Version", recommendedVersion)
			c.Header("Warning",
				fmt.Sprintf(`299 - "API version %s is deprecated. Please use %s"`,
					version, recommendedVersion))
		}

		// 添加版本信息到响应头
		c.Header("API-Version", version)
		c.Header("API-Supported-Versions", strings.Join(cfg.SupportedVersions, ", "))

		c.Next()
	})
}

// extractAPIVersion 提取API版本
func extractAPIVersion(c *gin.Context, cfg *APIVersionConfig) string {
	var version string

	// 1. 优先从URL路径提取版本
	if strings.HasPrefix(c.Request.URL.Path, cfg.VersionPrefix) {
		version = extractVersionFromPath(c.Request.URL.Path, cfg.VersionPrefix)
	}

	// 2. 从Header提取版本
	if version == "" {
		version = c.GetHeader(cfg.VersionHeader)
	}

	// 3. 从查询参数提取版本
	if version == "" {
		version = c.Query(cfg.VersionParam)
	}

	// 4. 使用默认版本
	if version == "" {
		version = cfg.DefaultVersion
	}

	return normalizeVersion(version)
}

// extractVersionFromPath 从路径提取版本
func extractVersionFromPath(path, prefix string) string {
	// 匹配 /api/v1/xxx 或 /api/v2/xxx 格式
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `(v\d+)`)
	matches := re.FindStringSubmatch(path)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// normalizeVersion 标准化版本格式
func normalizeVersion(version string) string {
	version = strings.ToLower(strings.TrimSpace(version))

	// 确保版本以v开头
	if version != "" && !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	return version
}

// isVersionSupported 检查版本是否支持
func isVersionSupported(version string, supportedVersions []string) bool {
	for _, supported := range supportedVersions {
		if version == supported {
			return true
		}
	}
	return false
}

// GetAPIVersion 获取当前请求的API版本
func GetAPIVersion(c *gin.Context) string {
	if version, exists := c.Get("api_version"); exists {
		if v, ok := version.(string); ok {
			return v
		}
	}
	return "v1" // 默认版本
}

// GetAPIVersionConfig 获取API版本配置
func GetAPIVersionConfig(c *gin.Context) *APIVersionConfig {
	if config, exists := c.Get("api_version_config"); exists {
		if cfg, ok := config.(*APIVersionConfig); ok {
			return cfg
		}
	}
	return DefaultAPIVersionConfig()
}

// IsVersionDeprecated 检查版本是否已弃用
func IsVersionDeprecated(c *gin.Context, version string) (bool, string) {
	cfg := GetAPIVersionConfig(c)
	if recommendedVersion, deprecated := cfg.DeprecatedMap[version]; deprecated {
		return true, recommendedVersion
	}
	return false, ""
}

// VersionedRoute 版本化路由结构
type VersionedRoute struct {
	Version string
	Method  string
	Path    string
	Handler gin.HandlerFunc
}

// RegisterVersionedRoutes 注册版本化路由
func RegisterVersionedRoutes(router *gin.Engine, routes []VersionedRoute) {
	for _, route := range routes {
		// 构建完整路径
		fullPath := fmt.Sprintf("/api/%s%s", route.Version, route.Path)

		// 注册路由
		switch strings.ToUpper(route.Method) {
		case "GET":
			router.GET(fullPath, route.Handler)
		case "POST":
			router.POST(fullPath, route.Handler)
		case "PUT":
			router.PUT(fullPath, route.Handler)
		case "PATCH":
			router.PATCH(fullPath, route.Handler)
		case "DELETE":
			router.DELETE(fullPath, route.Handler)
		case "OPTIONS":
			router.OPTIONS(fullPath, route.Handler)
		case "HEAD":
			router.HEAD(fullPath, route.Handler)
		default:
			panic(fmt.Sprintf("不支持的HTTP方法: %s", route.Method))
		}
	}
}

// APIVersionResponse API版本响应结构
type APIVersionResponse struct {
	CurrentVersion     string            `json:"current_version"`
	SupportedVersions  []string          `json:"supported_versions"`
	DeprecatedVersions map[string]string `json:"deprecated_versions,omitempty"`
	DefaultVersion     string            `json:"default_version"`
}

// VersionInfoHandler API版本信息处理器
func VersionInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := GetAPIVersionConfig(c)
		currentVersion := GetAPIVersion(c)

		response := APIVersionResponse{
			CurrentVersion:     currentVersion,
			SupportedVersions:  cfg.SupportedVersions,
			DeprecatedVersions: cfg.DeprecatedMap,
			DefaultVersion:     cfg.DefaultVersion,
		}

		utils.Success(c, response)
	}
}
