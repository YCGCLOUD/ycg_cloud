package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pkgErrors "cloudpan/internal/pkg/errors"
	"cloudpan/internal/pkg/logger"
)

// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
	Code      int         `json:"code"`              // HTTP状态码
	Message   string      `json:"message"`           // 错误信息
	Error     string      `json:"error,omitempty"`   // 错误类型
	Details   interface{} `json:"details,omitempty"` // 详细信息
	RequestID string      `json:"request_id"`        // 请求ID
	Timestamp time.Time   `json:"timestamp"`         // 时间戳
}

// ErrorHandlerConfig 错误处理中间件配置
type ErrorHandlerConfig struct {
	// EnableStackTrace 是否在开发环境下返回堆栈信息
	EnableStackTrace bool
	// LogStackTrace 是否记录堆栈信息到日志
	LogStackTrace bool
	// ErrorCodeMapping 自定义错误码映射
	ErrorCodeMapping map[error]int
}

// DefaultErrorHandlerConfig 默认配置
func DefaultErrorHandlerConfig() ErrorHandlerConfig {
	return ErrorHandlerConfig{
		EnableStackTrace: false, // 生产环境应设为false
		LogStackTrace:    true,
		ErrorCodeMapping: map[error]int{
			pkgErrors.ErrResourceNotFound:         http.StatusNotFound,
			pkgErrors.ErrPermissionDenied:         http.StatusForbidden,
			pkgErrors.ErrValidationFailed:         http.StatusBadRequest,
			pkgErrors.ErrInvalidInput:             http.StatusBadRequest,
			pkgErrors.ErrMissingRequired:          http.StatusBadRequest,
			pkgErrors.ErrInvalidFormat:            http.StatusBadRequest,
			pkgErrors.ErrResourceExists:           http.StatusConflict,
			pkgErrors.ErrOperationNotAllowed:      http.StatusMethodNotAllowed,
			pkgErrors.ErrQuotaExceeded:            http.StatusForbidden,
			pkgErrors.ErrNetworkTimeout:           http.StatusRequestTimeout,
			pkgErrors.ErrDatabaseConnectionFailed: http.StatusInternalServerError,
			pkgErrors.ErrCacheServerDown:          http.StatusInternalServerError,
		},
	}
}

// ErrorHandler 全局错误处理中间件
func ErrorHandler(config ...ErrorHandlerConfig) gin.HandlerFunc {
	cfg := DefaultErrorHandlerConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 处理panic
				handlePanic(c, err, cfg)
			}
		}()

		// 处理请求
		c.Next()

		// 处理Gin错误
		if len(c.Errors) > 0 {
			handleGinErrors(c, cfg)
		}
	}
}

// handlePanic 处理panic错误
func handlePanic(c *gin.Context, err interface{}, cfg ErrorHandlerConfig) {
	requestID := getRequestID(c)

	// 记录panic日志
	stack := debug.Stack()
	logger.Logger.Error("Panic recovered",
		zap.String("request_id", requestID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Any("panic", err),
		zap.String("stack", string(stack)),
	)

	// 构建错误响应
	response := ErrorResponse{
		Code:      http.StatusInternalServerError,
		Message:   "Internal server error",
		Error:     "panic",
		RequestID: requestID,
		Timestamp: time.Now(),
	}

	// 在开发环境下可能包含更多信息
	if cfg.EnableStackTrace {
		response.Details = map[string]interface{}{
			"panic": fmt.Sprintf("%v", err),
			"stack": strings.Split(string(stack), "\n"),
		}
	}

	c.JSON(http.StatusInternalServerError, response)
	c.Abort()
}

// handleGinErrors 处理Gin错误
func handleGinErrors(c *gin.Context, cfg ErrorHandlerConfig) {
	requestID := getRequestID(c)
	lastError := c.Errors.Last()

	if lastError == nil {
		return
	}

	err := lastError.Err
	statusCode := getErrorStatusCode(err, cfg)

	// 记录错误日志
	logLevel := getLogLevel(statusCode)
	fields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Int("status_code", statusCode),
		zap.Error(err),
	}

	// 添加用户ID（如果存在）
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			fields = append(fields, zap.String("user_id", uid))
		}
	}

	switch logLevel {
	case "error":
		logger.Logger.Error("Request failed", fields...)
	case "warn":
		logger.Logger.Warn("Request warning", fields...)
	default:
		logger.Logger.Info("Request info", fields...)
	}

	// 构建错误响应
	response := ErrorResponse{
		Code:      statusCode,
		Message:   getErrorMessage(err),
		Error:     getErrorType(err),
		RequestID: requestID,
		Timestamp: time.Now(),
	}

	// 添加详细信息（如果需要）
	if details := getErrorDetails(err, cfg); details != nil {
		response.Details = details
	}

	// 如果已经写入了响应，则不再处理
	if c.Writer.Written() {
		return
	}

	c.JSON(statusCode, response)
	c.Abort()
}

// getRequestID 获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if rid, ok := requestID.(string); ok {
			return rid
		}
	}
	return "unknown"
}

// getErrorStatusCode 获取错误对应的HTTP状态码
func getErrorStatusCode(err error, cfg ErrorHandlerConfig) int {
	// 检查自定义映射
	for mappedErr, code := range cfg.ErrorCodeMapping {
		if errors.Is(err, mappedErr) {
			return code
		}
	}

	// 根据错误类型判断
	if pkgErrors.IsNotFoundError(err) {
		return http.StatusNotFound
	}
	if pkgErrors.IsPermissionError(err) {
		return http.StatusForbidden
	}
	if pkgErrors.IsValidationError(err) {
		return http.StatusBadRequest
	}

	// 默认返回500
	return http.StatusInternalServerError
}

// getErrorMessage 获取错误信息
func getErrorMessage(err error) string {
	if err == nil {
		return "Unknown error"
	}

	// 使用错误映射表
	errorMessages := getErrorMessageMap()

	for errType, message := range errorMessages {
		if errors.Is(err, errType) {
			return message
		}
	}

	// 默认返回错误消息
	return err.Error()
}

// getErrorMessageMap 获取错误消息映射表
func getErrorMessageMap() map[error]string {
	return map[error]string{
		pkgErrors.ErrResourceNotFound:         "资源未找到",
		pkgErrors.ErrPermissionDenied:         "权限不足",
		pkgErrors.ErrValidationFailed:         "输入参数验证失败",
		pkgErrors.ErrInvalidInput:             "输入参数无效",
		pkgErrors.ErrMissingRequired:          "缺少必需参数",
		pkgErrors.ErrInvalidFormat:            "参数格式无效",
		pkgErrors.ErrResourceExists:           "资源已存在",
		pkgErrors.ErrOperationNotAllowed:      "操作不被允许",
		pkgErrors.ErrQuotaExceeded:            "配额超出限制",
		pkgErrors.ErrNetworkTimeout:           "网络超时",
		pkgErrors.ErrDatabaseConnectionFailed: "数据库连接失败",
		pkgErrors.ErrCacheServerDown:          "缓存服务异常",
	}
}

// getErrorType 获取错误类型
func getErrorType(err error) string {
	if err == nil {
		return "unknown"
	}

	// 使用错误类型映射表
	errorTypes := getErrorTypeMap()

	for errType, typeName := range errorTypes {
		if errors.Is(err, errType) {
			return typeName
		}
	}

	return "internal_error"
}

// getErrorTypeMap 获取错误类型映射表
func getErrorTypeMap() map[error]string {
	return map[error]string{
		pkgErrors.ErrResourceNotFound:         "resource_not_found",
		pkgErrors.ErrPermissionDenied:         "permission_denied",
		pkgErrors.ErrValidationFailed:         "validation_failed",
		pkgErrors.ErrInvalidInput:             "invalid_input",
		pkgErrors.ErrMissingRequired:          "missing_required",
		pkgErrors.ErrInvalidFormat:            "invalid_format",
		pkgErrors.ErrResourceExists:           "resource_exists",
		pkgErrors.ErrOperationNotAllowed:      "operation_not_allowed",
		pkgErrors.ErrQuotaExceeded:            "quota_exceeded",
		pkgErrors.ErrNetworkTimeout:           "network_timeout",
		pkgErrors.ErrDatabaseConnectionFailed: "database_error",
		pkgErrors.ErrCacheServerDown:          "cache_error",
	}
}

// getErrorDetails 获取错误详细信息
func getErrorDetails(err error, cfg ErrorHandlerConfig) interface{} {
	// 在开发环境下可能返回更多调试信息
	if !cfg.EnableStackTrace {
		return nil
	}

	// 可以根据需要添加更多详细信息
	return map[string]interface{}{
		"raw_error": err.Error(),
	}
}

// getLogLevel 根据状态码确定日志级别
func getLogLevel(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "error"
	case statusCode >= 400:
		return "warn"
	default:
		return "info"
	}
}

// CustomError 自定义错误辅助函数
func CustomError(c *gin.Context, statusCode int, errorType, message string, details interface{}) {
	requestID := getRequestID(c)

	response := ErrorResponse{
		Code:      statusCode,
		Message:   message,
		Error:     errorType,
		Details:   details,
		RequestID: requestID,
		Timestamp: time.Now(),
	}

	c.JSON(statusCode, response)
	c.Abort()
}

// ValidationError 验证错误辅助函数
func ValidationError(c *gin.Context, message string, details interface{}) {
	CustomError(c, http.StatusBadRequest, "validation_error", message, details)
}

// NotFoundError 未找到错误辅助函数
func NotFoundError(c *gin.Context, resource string) {
	message := fmt.Sprintf("%s not found", resource)
	CustomError(c, http.StatusNotFound, "not_found", message, nil)
}

// ForbiddenError 权限错误辅助函数
func ForbiddenError(c *gin.Context, message string) {
	if message == "" {
		message = "Permission denied"
	}
	CustomError(c, http.StatusForbidden, "permission_denied", message, nil)
}

// InternalError 内部错误辅助函数
func InternalError(c *gin.Context, err error) {
	requestID := getRequestID(c)

	// 记录错误日志
	logger.Logger.Error("Internal server error",
		zap.String("request_id", requestID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Error(err),
	)

	CustomError(c, http.StatusInternalServerError, "internal_error", "Internal server error", nil)
}
