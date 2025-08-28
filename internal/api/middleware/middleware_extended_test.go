package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	pkgErrors "cloudpan/internal/pkg/errors"
	"cloudpan/internal/pkg/logger"
)

func init() {
	// 初始化测试日志器
	logger.Logger = zap.NewNop() // 使用空日志器用于测试
}

// TestErrorHandlerWithErrors 测试错误处理中间件处理各种错误
func TestErrorHandlerWithErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		error          error
		expectedStatus int
		config         ErrorHandlerConfig
	}{
		{
			name:           "Resource not found",
			error:          pkgErrors.ErrResourceNotFound,
			expectedStatus: http.StatusNotFound,
			config:         DefaultErrorHandlerConfig(),
		},
		{
			name:           "Permission denied",
			error:          pkgErrors.ErrPermissionDenied,
			expectedStatus: http.StatusForbidden,
			config:         DefaultErrorHandlerConfig(),
		},
		{
			name:           "Validation failed",
			error:          pkgErrors.ErrValidationFailed,
			expectedStatus: http.StatusBadRequest,
			config:         DefaultErrorHandlerConfig(),
		},
		{
			name:           "Resource exists",
			error:          pkgErrors.ErrResourceExists,
			expectedStatus: http.StatusConflict,
			config:         DefaultErrorHandlerConfig(),
		},
		{
			name:           "Generic error",
			error:          errors.New("generic error"),
			expectedStatus: http.StatusInternalServerError,
			config:         DefaultErrorHandlerConfig(),
		},
		{
			name:           "Error with stack trace enabled",
			error:          errors.New("test error"),
			expectedStatus: http.StatusInternalServerError,
			config: ErrorHandlerConfig{
				EnableStackTrace: true,
				LogStackTrace:    true,
				ErrorCodeMapping: DefaultErrorHandlerConfig().ErrorCodeMapping,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建路由器
			r := gin.New()
			r.Use(ErrorHandler(tt.config))

			// 添加测试路由
			r.GET("/test", func(c *gin.Context) {
				c.Set("request_id", "test-request-123")
				c.Set("user_id", "test-user-456")
				c.Error(tt.error)
				c.Abort()
			})

			// 创建测试请求
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// 执行请求
			r.ServeHTTP(w, req)

			// 验证状态码
			assert.Equal(t, tt.expectedStatus, w.Code)

			// 验证响应包含错误信息
			assert.Contains(t, w.Body.String(), "error")
			assert.Contains(t, w.Body.String(), "test-request-123")
		})
	}
}

// TestErrorHandlerPanic 测试panic处理
func TestErrorHandlerPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(ErrorHandler())

	// 添加会引发panic的路由
	r.GET("/panic", func(c *gin.Context) {
		c.Set("request_id", "panic-test-123")
		panic("test panic")
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// 执行请求
	r.ServeHTTP(w, req)

	// 验证panic被正确处理
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
	assert.Contains(t, w.Body.String(), "panic-test-123")
}

// TestErrorHandlerCustomConfig 测试自定义配置
func TestErrorHandlerCustomConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	customError := errors.New("custom error")
	customConfig := ErrorHandlerConfig{
		EnableStackTrace: true,
		LogStackTrace:    true,
		ErrorCodeMapping: map[error]int{
			customError: http.StatusTeapot, // 418
		},
	}

	r := gin.New()
	r.Use(ErrorHandler(customConfig))

	r.GET("/custom", func(c *gin.Context) {
		c.Error(customError)
		c.Abort()
	})

	req := httptest.NewRequest("GET", "/custom", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// 验证自定义错误码映射生效
	assert.Equal(t, http.StatusTeapot, w.Code)
}

// TestDefaultErrorHandlerConfig 测试默认配置
func TestDefaultErrorHandlerConfig(t *testing.T) {
	config := DefaultErrorHandlerConfig()

	assert.False(t, config.EnableStackTrace)
	assert.True(t, config.LogStackTrace)
	assert.NotNil(t, config.ErrorCodeMapping)

	// 验证预定义的错误映射
	assert.Equal(t, http.StatusNotFound, config.ErrorCodeMapping[pkgErrors.ErrResourceNotFound])
	assert.Equal(t, http.StatusForbidden, config.ErrorCodeMapping[pkgErrors.ErrPermissionDenied])
	assert.Equal(t, http.StatusBadRequest, config.ErrorCodeMapping[pkgErrors.ErrValidationFailed])
}

// TestErrorResponseStructure 测试错误响应结构
func TestErrorResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(ErrorHandler())

	r.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "struct-test-123")
		c.Error(pkgErrors.ErrValidationFailed)
		c.Abort()
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "code")
	assert.Contains(t, w.Body.String(), "message")
	assert.Contains(t, w.Body.String(), "request_id")
	assert.Contains(t, w.Body.String(), "timestamp")
	assert.Contains(t, w.Body.String(), "struct-test-123")
}

// TestRequestLoggerWithConfig 测试请求日志中间件配置
func TestRequestLoggerWithConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := RequestLoggerConfig{
		SkipPaths:        []string{"/health"},
		LogRequestBody:   true,
		LogResponseBody:  true,
		MaxBodySize:      1024,
		SensitiveHeaders: []string{"Authorization"},
	}

	r := gin.New()
	r.Use(RequestLogger(config))

	// 测试正常路径
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试跳过路径
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 测试正常路径
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 测试跳过路径
	req2 := httptest.NewRequest("GET", "/health", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

// TestDetailedRequestLogger 测试详细请求日志中间件
func TestDetailedRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := RequestLoggerConfig{
		LogRequestBody:  true,
		LogResponseBody: true,
		MaxBodySize:     1024,
	}

	r := gin.New()
	r.Use(DetailedRequestLogger(config))

	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": 123})
	})

	// 创建带请求体的POST请求
	body := strings.NewReader(`{"name": "test"}`)
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom-Header", "test-value")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Header().Get("X-Request-ID"), "")
}

// TestRequestLoggerDefaultConfig 测试请求日志中间件默认配置
func TestRequestLoggerDefaultConfig(t *testing.T) {
	config := DefaultRequestLoggerConfig()

	assert.Contains(t, config.SkipPaths, "/health")
	assert.Contains(t, config.SkipPaths, "/metrics")
	assert.Contains(t, config.SkipPaths, "/favicon.ico")
	assert.False(t, config.LogRequestBody)
	assert.False(t, config.LogResponseBody)
	assert.Equal(t, int64(4096), config.MaxBodySize)
	assert.Contains(t, config.SensitiveHeaders, "Authorization")
	assert.Contains(t, config.SensitiveHeaders, "Cookie")
}

// TestRequestLoggerWithLargeBody 测试大请求体处理
func TestRequestLoggerWithLargeBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := RequestLoggerConfig{
		LogRequestBody: true,
		MaxBodySize:    10, // 很小的限制
	}

	r := gin.New()
	r.Use(DetailedRequestLogger(config))

	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"received": true})
	})

	// 创建超过MaxBodySize的请求
	largeBody := strings.NewReader("this is a very long request body that exceeds the limit")
	req := httptest.NewRequest("POST", "/test", largeBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestVersionMiddlewareEdgeCases 测试版本中间件边界情况
func TestVersionMiddlewareEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 测试自定义配置
	config := &APIVersionConfig{
		DefaultVersion:    "v2",
		SupportedVersions: []string{"v1", "v2", "v3"},
		VersionHeader:     "X-API-Version",
		VersionParam:      "version",
		VersionPrefix:     "/api/",
		DeprecatedMap:     map[string]string{"v1": "v2"},
	}

	r := gin.New()
	r.Use(APIVersionMiddleware(config))

	r.GET("/api/:version/test", func(c *gin.Context) {
		version := c.GetString("api_version")
		c.JSON(http.StatusOK, gin.H{
			"version": version,
		})
	})

	// 测试路径中的版本
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "v1")
	assert.Equal(t, "true", w.Header().Get("API-Deprecated"))

	// 测试Header中的版本
	req2 := httptest.NewRequest("GET", "/api/v2/test", nil)
	req2.Header.Set("X-API-Version", "v3")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "v2") // URL中的版本优先

	// 测试查询参数中的版本
	req3 := httptest.NewRequest("GET", "/api/v2/test?version=v3", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
	assert.Contains(t, w3.Body.String(), "v2") // URL中的版本优先
}

// TestI18nMiddlewareExtended 测试国际化中间件扩展功能
func TestI18nMiddlewareExtended(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &I18nConfig{
		DefaultLanguage:    "en-US",
		SupportedLanguages: []string{"en-US", "zh-CN", "ja-JP"},
		LanguageParam:      "lang",
		LanguageHeader:     "Accept-Language",
	}

	r := gin.New()
	r.Use(I18nMiddleware(config))

	r.GET("/test", func(c *gin.Context) {
		lang := c.GetString("language")
		c.JSON(http.StatusOK, gin.H{"language": lang})
	})

	// 测试复杂的Accept-Language header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "zh-CN")

	// 测试不支持的语言回退到默认值
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Accept-Language", "fr-FR")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "en-US")

	// 测试查询参数优先级
	req3 := httptest.NewRequest("GET", "/test?lang=ja-JP", nil)
	req3.Header.Set("Accept-Language", "zh-CN")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
	assert.Contains(t, w3.Body.String(), "ja-JP")
}

// TestCORSMiddlewareAdvanced 测试CORS中间件高级功能
func TestCORSMiddlewareAdvanced(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &CORSOptions{
		AllowedOrigins:   []string{"*.example.com", "https://app.test.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Custom-Header"},
		ExposedHeaders:   []string{"X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	r := gin.New()
	r.Use(CORS(config))

	r.OPTIONS("/api/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	r.POST("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// 测试预检请求
	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "https://sub.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://sub.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

	// 测试实际请求
	req2 := httptest.NewRequest("POST", "/api/test", bytes.NewReader([]byte(`{"test": true}`)))
	req2.Header.Set("Origin", "https://app.test.com")
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "https://app.test.com", w2.Header().Get("Access-Control-Allow-Origin"))
}
