package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"cloudpan/internal/pkg/logger"
)

// RequestLogger HTTP请求日志中间件配置
type RequestLoggerConfig struct {
	// SkipPaths 跳过记录的路径列表
	SkipPaths []string
	// LogRequestBody 是否记录请求体
	LogRequestBody bool
	// LogResponseBody 是否记录响应体
	LogResponseBody bool
	// MaxBodySize 最大记录的请求/响应体大小
	MaxBodySize int64
	// SensitiveHeaders 敏感headers，记录时会脱敏
	SensitiveHeaders []string
}

// DefaultRequestLoggerConfig 默认配置
func DefaultRequestLoggerConfig() RequestLoggerConfig {
	return RequestLoggerConfig{
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
		LogRequestBody:  false, // 默认不记录请求体（可能包含敏感信息）
		LogResponseBody: false, // 默认不记录响应体（避免日志过大）
		MaxBodySize:     4096,  // 4KB
		SensitiveHeaders: []string{
			"Authorization",
			"Cookie",
			"Set-Cookie",
			"X-Auth-Token",
			"X-API-Key",
		},
	}
}

// RequestLogger 创建请求日志中间件
func RequestLogger(config ...RequestLoggerConfig) gin.HandlerFunc {
	cfg := DefaultRequestLoggerConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// 构建跳过路径的map，提高查找效率
	skipPathsMap := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPathsMap[path] = true
	}

	// 构建敏感header的map
	sensitiveHeadersMap := make(map[string]bool)
	for _, header := range cfg.SensitiveHeaders {
		sensitiveHeadersMap[strings.ToLower(header)] = true
	}

	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			// 跳过指定路径
			if skipPathsMap[param.Path] {
				return ""
			}

			// 生成请求ID
			requestID := generateRequestID()

			// 构建访问日志条目
			entry := logger.AccessLogEntry{
				Timestamp:    param.TimeStamp,
				RequestID:    requestID,
				Method:       param.Method,
				Path:         param.Path,
				Query:        param.Request.URL.RawQuery,
				StatusCode:   param.StatusCode,
				ResponseTime: param.Latency.Milliseconds(),
				IPAddress:    param.ClientIP,
				UserAgent:    param.Request.UserAgent(),
				RequestSize:  param.Request.ContentLength,
				Protocol:     param.Request.Proto,
			}

			// 从上下文中获取用户ID（如果存在）
			if userID := param.Keys["user_id"]; userID != nil {
				if uid, ok := userID.(string); ok {
					entry.UserID = uid
				}
			}

			// 记录访问日志
			logger.LogAccess(entry)

			return "" // 返回空字符串，避免Gin默认日志输出
		},
		Output: io.Discard, // 禁用Gin默认输出
	})
}

// DetailedRequestLogger 详细请求日志中间件
func DetailedRequestLogger(config ...RequestLoggerConfig) gin.HandlerFunc {
	cfg := DefaultRequestLoggerConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// 构建跳过路径的map
	skipPathsMap := buildSkipPathsMap(cfg.SkipPaths)

	return func(c *gin.Context) {
		// 跳过指定路径
		if skipPathsMap[c.Request.URL.Path] {
			c.Next()
			return
		}

		// 设置请求ID和开始时间
		requestID := setupRequestLogging(c)
		startTime := time.Now()

		// 处理请求体和响应体记录
		requestBody := readRequestBody(c, cfg)
		responseWriter := setupResponseCapture(c, cfg)

		// 记录请求开始日志
		logRequestStart(c, requestID)

		// 处理请求
		c.Next()

		// 记录请求完成日志
		logRequestCompletion(c, requestID, startTime, requestBody, responseWriter)
	}
}

// buildSkipPathsMap 构建跳过路径的map
func buildSkipPathsMap(skipPaths []string) map[string]bool {
	skipPathsMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipPathsMap[path] = true
	}
	return skipPathsMap
}

// setupRequestLogging 设置请求日志
func setupRequestLogging(c *gin.Context) string {
	// 生成请求ID并设置到上下文
	requestID := generateRequestID()
	c.Set("request_id", requestID)
	c.Header("X-Request-ID", requestID)
	return requestID
}

// readRequestBody 读取请求体
func readRequestBody(c *gin.Context, cfg RequestLoggerConfig) string {
	if !cfg.LogRequestBody || c.Request.Body == nil {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, cfg.MaxBodySize))
	if err != nil {
		return ""
	}

	// 重置请求体供后续处理使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	return string(body)
}

// setupResponseCapture 设置响应捕获
func setupResponseCapture(c *gin.Context, cfg RequestLoggerConfig) *bodyLogWriter {
	if !cfg.LogResponseBody {
		return nil
	}

	blw := &bodyLogWriter{
		body:           &bytes.Buffer{},
		ResponseWriter: c.Writer,
		maxSize:        cfg.MaxBodySize,
	}
	c.Writer = blw
	return blw
}

// logRequestStart 记录请求开始日志
func logRequestStart(c *gin.Context, requestID string) {
	logger.WithRequestID(c.Request.Context(), requestID).Info("HTTP request started",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("query", c.Request.URL.RawQuery),
		zap.String("ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.Int64("content_length", c.Request.ContentLength),
	)
}

// logRequestCompletion 记录请求完成日志
func logRequestCompletion(c *gin.Context, requestID string, startTime time.Time, requestBody string, responseWriter *bodyLogWriter) {
	duration := time.Since(startTime)

	// 构建基础日志字段
	fields := buildLogFields(c, requestID, duration)

	// 添加请求体和响应体
	addOptionalFields(&fields, c, requestBody, responseWriter)

	// 记录日志
	if c.Writer.Status() >= 400 {
		logger.Logger.Error("HTTP request completed with error", fields...)
	} else {
		logger.Logger.Info("HTTP request completed", fields...)
	}
}

// buildLogFields 构建日志字段
func buildLogFields(c *gin.Context, requestID string, duration time.Duration) []zap.Field {
	return []zap.Field{
		zap.String("request_id", requestID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("query", c.Request.URL.RawQuery),
		zap.Int("status_code", c.Writer.Status()),
		zap.Duration("duration", duration),
		zap.String("ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.Int64("content_length", c.Request.ContentLength),
		zap.Int("response_size", c.Writer.Size()),
	}
}

// addOptionalFields 添加可选字段
func addOptionalFields(fields *[]zap.Field, c *gin.Context, requestBody string, responseWriter *bodyLogWriter) {
	// 添加用户ID
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			*fields = append(*fields, zap.String("user_id", uid))
		}
	}

	// 添加请求体
	if requestBody != "" {
		*fields = append(*fields, zap.String("request_body", requestBody))
	}

	// 添加响应体
	if responseWriter != nil {
		*fields = append(*fields, zap.String("response_body", responseWriter.body.String()))
	}

	// 添加错误信息
	if len(c.Errors) > 0 {
		*fields = append(*fields, zap.String("errors", c.Errors.String()))
	}
}

// bodyLogWriter 响应体记录器
type bodyLogWriter struct {
	gin.ResponseWriter
	body    *bytes.Buffer
	maxSize int64
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	// 记录响应体（限制大小）
	if w.body.Len() < int(w.maxSize) {
		remaining := int(w.maxSize) - w.body.Len()
		if len(b) > remaining {
			w.body.Write(b[:remaining])
		} else {
			w.body.Write(b)
		}
	}
	return w.ResponseWriter.Write(b)
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// RequestIDMiddleware 请求ID中间件（轻量级版本）
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// UserIDMiddleware 用户ID中间件（需要在认证中间件之后使用）
func UserIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里假设认证中间件已经设置了用户ID
		// 实际实现时需要根据具体的认证机制来获取用户ID
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(string); ok && uid != "" {
				// 用户ID已存在，直接继续
				c.Next()
				return
			}
		}

		// 如果没有用户ID，可能是匿名访问
		c.Next()
	}
}
