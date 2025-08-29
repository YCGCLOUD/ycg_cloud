package errors

import (
	"errors"
	"fmt"
)

// ValidationError 验证错误结构体
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error 实现error接口
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// InternalError 内部错误结构体
type InternalError struct {
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

// Error 实现error接口
func (e *InternalError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap 实现错误包装
func (e *InternalError) Unwrap() error {
	return e.Cause
}

// 缓存相关错误
var (
	// ErrCacheNotFound 缓存未找到
	ErrCacheNotFound = errors.New("cache not found")
	// ErrCacheExpired 缓存已过期
	ErrCacheExpired = errors.New("cache expired")
	// ErrInvalidCacheKey 无效的缓存键
	ErrInvalidCacheKey = errors.New("invalid cache key")
	// ErrCacheServerDown 缓存服务器故障
	ErrCacheServerDown = errors.New("cache server down")
	// ErrInvalidTTL 无效的TTL值
	ErrInvalidTTL = errors.New("invalid TTL value")
	// ErrRedisNotInitialized Redis未初始化
	ErrRedisNotInitialized = errors.New("redis not initialized")
)

// 配置相关错误
var (
	// ErrConfigNotFound 配置文件未找到
	ErrConfigNotFound = errors.New("config file not found")
	// ErrConfigInvalid 配置文件无效
	ErrConfigInvalid = errors.New("config file invalid")
	// ErrConfigNotInitialized 配置未初始化
	ErrConfigNotInitialized = errors.New("config not initialized")
	// ErrInvalidConfigPath 无效的配置路径
	ErrInvalidConfigPath = errors.New("invalid config path")
)

// 数据库相关错误
var (
	// ErrDatabaseNotInitialized 数据库未初始化
	ErrDatabaseNotInitialized = errors.New("database not initialized")
	// ErrDatabaseConnectionFailed 数据库连接失败
	ErrDatabaseConnectionFailed = errors.New("database connection failed")
	// ErrInvalidDSN 无效的数据库连接字符串
	ErrInvalidDSN = errors.New("invalid database DSN")
	// ErrTransactionFailed 事务执行失败
	ErrTransactionFailed = errors.New("transaction failed")
	// ErrLockFailed 锁获取失败
	ErrLockFailed = errors.New("lock acquisition failed")
	// ErrLockTimeout 锁超时
	ErrLockTimeout = errors.New("lock timeout")
)

// 验证相关错误
var (
	// ErrValidationFailed 验证失败
	ErrValidationFailed = errors.New("validation failed")
	// ErrInvalidInput 无效输入
	ErrInvalidInput = errors.New("invalid input")
	// ErrMissingRequired 缺少必需参数
	ErrMissingRequired = errors.New("missing required parameter")
	// ErrInvalidFormat 格式无效
	ErrInvalidFormat = errors.New("invalid format")
)

// 业务逻辑错误
var (
	// ErrResourceNotFound 资源未找到
	ErrResourceNotFound = errors.New("resource not found")
	// ErrResourceExists 资源已存在
	ErrResourceExists = errors.New("resource already exists")
	// ErrPermissionDenied 权限被拒绝
	ErrPermissionDenied = errors.New("permission denied")
	// ErrOperationNotAllowed 操作不被允许
	ErrOperationNotAllowed = errors.New("operation not allowed")
	// ErrQuotaExceeded 配额超出
	ErrQuotaExceeded = errors.New("quota exceeded")
)

// 网络和I/O错误
var (
	// ErrNetworkTimeout 网络超时
	ErrNetworkTimeout = errors.New("network timeout")
	// ErrFileNotFound 文件未找到
	ErrFileNotFound = errors.New("file not found")
	// ErrFileCorrupted 文件损坏
	ErrFileCorrupted = errors.New("file corrupted")
	// ErrDiskFull 磁盘空间不足
	ErrDiskFull = errors.New("disk full")
	// ErrIOError I/O错误
	ErrIOError = errors.New("I/O error")
)

// 错误包装函数
// WrapError 包装错误，添加上下文信息
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapErrorf 包装错误，添加格式化的上下文信息
func WrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", message, err)
}

// NewValidationError 创建验证错误
func NewValidationError(field string, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewInternalError 创建内部错误
func NewInternalError(message string) *InternalError {
	return &InternalError{
		Message: message,
	}
}

// NewInternalErrorWithCause 创建带原因的内部错误
func NewInternalErrorWithCause(message string, cause error) *InternalError {
	return &InternalError{
		Message: message,
		Cause:   cause,
	}
}

// NewResourceError 创建资源相关错误
func NewResourceError(resource string, operation string, err error) error {
	return fmt.Errorf("failed to %s %s: %w", operation, resource, err)
}

// IsNotFoundError 检查是否为资源未找到错误
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrResourceNotFound) ||
		errors.Is(err, ErrFileNotFound) ||
		errors.Is(err, ErrCacheNotFound) ||
		errors.Is(err, ErrConfigNotFound)
}

// IsPermissionError 检查是否为权限错误
func IsPermissionError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrPermissionDenied) ||
		errors.Is(err, ErrOperationNotAllowed)
}

// IsValidationError 检查是否为验证错误
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrValidationFailed) ||
		errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrMissingRequired) ||
		errors.Is(err, ErrInvalidFormat)
}

// IsRetryableError 检查错误是否可重试
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNetworkTimeout) ||
		errors.Is(err, ErrCacheServerDown) ||
		errors.Is(err, ErrDatabaseConnectionFailed) ||
		errors.Is(err, ErrLockTimeout)
}
