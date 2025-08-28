package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSOptions CORS配置选项
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSOptions 默认CORS配置
func DefaultCORSOptions() *CORSOptions {
	return &CORSOptions{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"X-Requested-With",
			"Accept",
			"Cache-Control",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Headers",
			"Cache-Control",
			"Content-Language",
			"Content-Type",
			"Expires",
			"Last-Modified",
			"Pragma",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24小时
	}
}

// CORS 创建CORS中间件
func CORS(options ...*CORSOptions) gin.HandlerFunc {
	var opts *CORSOptions
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	} else {
		opts = DefaultCORSOptions()
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 设置CORS头部
		setCORSHeaders(c, origin, opts)

		// 处理预检请求
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

// setCORSHeaders 设置CORS头部
func setCORSHeaders(c *gin.Context, origin string, opts *CORSOptions) {
	// 设置允许的源
	setAllowOriginHeader(c, origin, opts.AllowedOrigins)

	// 设置允许的方法
	if len(opts.AllowedMethods) > 0 {
		c.Header("Access-Control-Allow-Methods", strings.Join(opts.AllowedMethods, ", "))
	}

	// 设置允许的头部
	if len(opts.AllowedHeaders) > 0 {
		c.Header("Access-Control-Allow-Headers", strings.Join(opts.AllowedHeaders, ", "))
	}

	// 设置暴露的头部
	if len(opts.ExposedHeaders) > 0 {
		c.Header("Access-Control-Expose-Headers", strings.Join(opts.ExposedHeaders, ", "))
	}

	// 设置是否允许凭证
	if opts.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	// 设置预检请求缓存时间
	if opts.MaxAge > 0 {
		c.Header("Access-Control-Max-Age", string(rune(opts.MaxAge)))
	}
}

// setAllowOriginHeader 设置允许的源头部
func setAllowOriginHeader(c *gin.Context, origin string, allowedOrigins []string) {
	if isOriginAllowed(origin, allowedOrigins) {
		c.Header("Access-Control-Allow-Origin", origin)
	} else if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		c.Header("Access-Control-Allow-Origin", "*")
	}
}

// isOriginAllowed 检查源是否被允许
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
		// 支持子域名匹配，例如 *.example.com
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:]
			// 从origin中提取域名部分（去掉协议）
			originWithoutProtocol := origin
			if strings.Contains(origin, "://") {
				parts := strings.SplitN(origin, "://", 2)
				if len(parts) == 2 {
					originWithoutProtocol = parts[1]
				}
			}
			// 检查是否为子域名（以 .domain 结尾）
			if strings.HasSuffix(originWithoutProtocol, "."+domain) {
				return true
			}
			// 检查是否为精确的域名匹配
			if originWithoutProtocol == domain {
				return true
			}
		}
	}

	return false
}

// CORSMiddleware 便捷的CORS中间件函数
func CORSMiddleware() gin.HandlerFunc {
	return CORS()
}

// ProductionCORS 生产环境CORS配置
func ProductionCORS(allowedOrigins []string) gin.HandlerFunc {
	opts := &CORSOptions{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Accept",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           3600, // 1小时
	}

	return CORS(opts)
}
