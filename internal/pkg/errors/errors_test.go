package errors

import (
	"errors"
	"testing"
)

// TestCacheErrors 测试缓存相关错误
func TestCacheErrors(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{"ErrCacheNotFound", ErrCacheNotFound},
		{"ErrCacheExpired", ErrCacheExpired},
		{"ErrInvalidCacheKey", ErrInvalidCacheKey},
		{"ErrCacheServerDown", ErrCacheServerDown},
		{"ErrInvalidTTL", ErrInvalidTTL},
		{"ErrRedisNotInitialized", ErrRedisNotInitialized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.error == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.error.Error() == "" {
				t.Errorf("%s should have a non-empty error message", tt.name)
			}
		})
	}
}

// TestConfigErrors 测试配置相关错误
func TestConfigErrors(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{"ErrConfigNotFound", ErrConfigNotFound},
		{"ErrConfigInvalid", ErrConfigInvalid},
		{"ErrConfigNotInitialized", ErrConfigNotInitialized},
		{"ErrInvalidConfigPath", ErrInvalidConfigPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.error == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.error.Error() == "" {
				t.Errorf("%s should have a non-empty error message", tt.name)
			}
		})
	}
}

// TestDatabaseErrors 测试数据库相关错误
func TestDatabaseErrors(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{"ErrDatabaseNotInitialized", ErrDatabaseNotInitialized},
		{"ErrDatabaseConnectionFailed", ErrDatabaseConnectionFailed},
		{"ErrInvalidDSN", ErrInvalidDSN},
		{"ErrTransactionFailed", ErrTransactionFailed},
		{"ErrLockFailed", ErrLockFailed},
		{"ErrLockTimeout", ErrLockTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.error == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.error.Error() == "" {
				t.Errorf("%s should have a non-empty error message", tt.name)
			}
		})
	}
}

// TestValidationErrors 测试验证相关错误
func TestValidationErrors(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{"ErrValidationFailed", ErrValidationFailed},
		{"ErrInvalidInput", ErrInvalidInput},
		{"ErrMissingRequired", ErrMissingRequired},
		{"ErrInvalidFormat", ErrInvalidFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.error == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.error.Error() == "" {
				t.Errorf("%s should have a non-empty error message", tt.name)
			}
		})
	}
}

// TestBusinessLogicErrors 测试业务逻辑错误
func TestBusinessLogicErrors(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{"ErrResourceNotFound", ErrResourceNotFound},
		{"ErrResourceExists", ErrResourceExists},
		{"ErrPermissionDenied", ErrPermissionDenied},
		{"ErrOperationNotAllowed", ErrOperationNotAllowed},
		{"ErrQuotaExceeded", ErrQuotaExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.error == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.error.Error() == "" {
				t.Errorf("%s should have a non-empty error message", tt.name)
			}
		})
	}
}

// TestNetworkIOErrors 测试网络和I/O错误
func TestNetworkIOErrors(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{"ErrNetworkTimeout", ErrNetworkTimeout},
		{"ErrFileNotFound", ErrFileNotFound},
		{"ErrFileCorrupted", ErrFileCorrupted},
		{"ErrDiskFull", ErrDiskFull},
		{"ErrIOError", ErrIOError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.error == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.error.Error() == "" {
				t.Errorf("%s should have a non-empty error message", tt.name)
			}
		})
	}
}

// TestWrapError 测试错误包装函数
func TestWrapError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		message  string
		expected string
	}{
		{
			name:     "wrap nil error",
			err:      nil,
			message:  "test message",
			expected: "",
		},
		{
			name:     "wrap normal error",
			err:      errors.New("original error"),
			message:  "wrapped",
			expected: "wrapped: original error",
		},
		{
			name:     "wrap with empty message",
			err:      errors.New("original error"),
			message:  "",
			expected: ": original error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.err, tt.message)
			if tt.err == nil {
				if result != nil {
					t.Errorf("WrapError with nil error should return nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Error("WrapError with non-nil error should not return nil")
				} else if result.Error() != tt.expected {
					t.Errorf("WrapError() = %q, want %q", result.Error(), tt.expected)
				}
			}
		})
	}
}

// TestWrapErrorf 测试格式化错误包装函数
func TestWrapErrorf(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "wrap nil error",
			err:      nil,
			format:   "test %s",
			args:     []interface{}{"message"},
			expected: "",
		},
		{
			name:     "wrap with format",
			err:      errors.New("original error"),
			format:   "operation %s failed",
			args:     []interface{}{"create"},
			expected: "operation create failed: original error",
		},
		{
			name:     "wrap with multiple args",
			err:      errors.New("original error"),
			format:   "operation %s failed on %s",
			args:     []interface{}{"create", "user"},
			expected: "operation create failed on user: original error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapErrorf(tt.err, tt.format, tt.args...)
			if tt.err == nil {
				if result != nil {
					t.Errorf("WrapErrorf with nil error should return nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Error("WrapErrorf with non-nil error should not return nil")
				} else if result.Error() != tt.expected {
					t.Errorf("WrapErrorf() = %q, want %q", result.Error(), tt.expected)
				}
			}
		})
	}
}

// TestNewValidationError 测试验证错误创建函数
func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		message  string
		expected string
	}{
		{
			name:     "normal validation error",
			field:    "email",
			message:  "invalid format",
			expected: "validation failed for field 'email': invalid format",
		},
		{
			name:     "empty field",
			field:    "",
			message:  "required",
			expected: "validation failed for field '': required",
		},
		{
			name:     "empty message",
			field:    "username",
			message:  "",
			expected: "validation failed for field 'username': ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewValidationError(tt.field, tt.message)
			if result == nil {
				t.Error("NewValidationError should not return nil")
			} else if result.Error() != tt.expected {
				t.Errorf("NewValidationError() = %q, want %q", result.Error(), tt.expected)
			}
		})
	}
}

// TestNewResourceError 测试资源错误创建函数
func TestNewResourceError(t *testing.T) {
	tests := []struct {
		name      string
		resource  string
		operation string
		err       error
		expected  string
	}{
		{
			name:      "normal resource error",
			resource:  "user",
			operation: "create",
			err:       errors.New("database error"),
			expected:  "failed to create user: database error",
		},
		{
			name:      "empty resource",
			resource:  "",
			operation: "delete",
			err:       errors.New("not found"),
			expected:  "failed to delete : not found",
		},
		{
			name:      "empty operation",
			resource:  "file",
			operation: "",
			err:       errors.New("io error"),
			expected:  "failed to  file: io error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewResourceError(tt.resource, tt.operation, tt.err)
			if result == nil {
				t.Error("NewResourceError should not return nil")
			} else if result.Error() != tt.expected {
				t.Errorf("NewResourceError() = %q, want %q", result.Error(), tt.expected)
			}
		})
	}
}

// TestIsNotFoundError 测试未找到错误检查函数
func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "resource not found",
			err:      ErrResourceNotFound,
			expected: true,
		},
		{
			name:     "file not found",
			err:      ErrFileNotFound,
			expected: true,
		},
		{
			name:     "cache not found",
			err:      ErrCacheNotFound,
			expected: true,
		},
		{
			name:     "config not found",
			err:      ErrConfigNotFound,
			expected: true,
		},
		{
			name:     "wrapped resource not found",
			err:      WrapError(ErrResourceNotFound, "wrapped"),
			expected: true,
		},
		{
			name:     "other error",
			err:      ErrPermissionDenied,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFoundError(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFoundError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsPermissionError 测试权限错误检查函数
func TestIsPermissionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "permission denied",
			err:      ErrPermissionDenied,
			expected: true,
		},
		{
			name:     "operation not allowed",
			err:      ErrOperationNotAllowed,
			expected: true,
		},
		{
			name:     "wrapped permission denied",
			err:      WrapError(ErrPermissionDenied, "wrapped"),
			expected: true,
		},
		{
			name:     "other error",
			err:      ErrResourceNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPermissionError(tt.err)
			if result != tt.expected {
				t.Errorf("IsPermissionError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsValidationError 测试验证错误检查函数
func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "validation failed",
			err:      ErrValidationFailed,
			expected: true,
		},
		{
			name:     "invalid input",
			err:      ErrInvalidInput,
			expected: true,
		},
		{
			name:     "missing required",
			err:      ErrMissingRequired,
			expected: true,
		},
		{
			name:     "invalid format",
			err:      ErrInvalidFormat,
			expected: true,
		},
		{
			name:     "wrapped validation error",
			err:      WrapError(ErrValidationFailed, "wrapped"),
			expected: true,
		},
		{
			name:     "other error",
			err:      ErrResourceNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidationError(tt.err)
			if result != tt.expected {
				t.Errorf("IsValidationError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsRetryableError 测试可重试错误检查函数
func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "network timeout",
			err:      ErrNetworkTimeout,
			expected: true,
		},
		{
			name:     "cache server down",
			err:      ErrCacheServerDown,
			expected: true,
		},
		{
			name:     "database connection failed",
			err:      ErrDatabaseConnectionFailed,
			expected: true,
		},
		{
			name:     "lock timeout",
			err:      ErrLockTimeout,
			expected: true,
		},
		{
			name:     "wrapped retryable error",
			err:      WrapError(ErrNetworkTimeout, "wrapped"),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      ErrPermissionDenied,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestErrorConstistency 测试错误常量的一致性
func TestErrorConstistency(t *testing.T) {
	// 确保所有预定义错误都不为空且有有意义的消息
	errorMap := map[string]error{
		"ErrCacheNotFound":            ErrCacheNotFound,
		"ErrCacheExpired":             ErrCacheExpired,
		"ErrInvalidCacheKey":          ErrInvalidCacheKey,
		"ErrCacheServerDown":          ErrCacheServerDown,
		"ErrInvalidTTL":               ErrInvalidTTL,
		"ErrRedisNotInitialized":      ErrRedisNotInitialized,
		"ErrConfigNotFound":           ErrConfigNotFound,
		"ErrConfigInvalid":            ErrConfigInvalid,
		"ErrConfigNotInitialized":     ErrConfigNotInitialized,
		"ErrInvalidConfigPath":        ErrInvalidConfigPath,
		"ErrDatabaseNotInitialized":   ErrDatabaseNotInitialized,
		"ErrDatabaseConnectionFailed": ErrDatabaseConnectionFailed,
		"ErrInvalidDSN":               ErrInvalidDSN,
		"ErrTransactionFailed":        ErrTransactionFailed,
		"ErrLockFailed":               ErrLockFailed,
		"ErrLockTimeout":              ErrLockTimeout,
		"ErrValidationFailed":         ErrValidationFailed,
		"ErrInvalidInput":             ErrInvalidInput,
		"ErrMissingRequired":          ErrMissingRequired,
		"ErrInvalidFormat":            ErrInvalidFormat,
		"ErrResourceNotFound":         ErrResourceNotFound,
		"ErrResourceExists":           ErrResourceExists,
		"ErrPermissionDenied":         ErrPermissionDenied,
		"ErrOperationNotAllowed":      ErrOperationNotAllowed,
		"ErrQuotaExceeded":            ErrQuotaExceeded,
		"ErrNetworkTimeout":           ErrNetworkTimeout,
		"ErrFileNotFound":             ErrFileNotFound,
		"ErrFileCorrupted":            ErrFileCorrupted,
		"ErrDiskFull":                 ErrDiskFull,
		"ErrIOError":                  ErrIOError,
	}

	for name, err := range errorMap {
		t.Run(name, func(t *testing.T) {
			if err == nil {
				t.Errorf("%s should not be nil", name)
			}
			if err.Error() == "" {
				t.Errorf("%s should have a non-empty error message", name)
			}
			if len(err.Error()) < 5 {
				t.Errorf("%s error message is too short: %q", name, err.Error())
			}
		})
	}
}
