package test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"cloudpan/internal/api/middleware"
	"cloudpan/internal/pkg/logger"
	"cloudpan/internal/pkg/utils"
)

// TestIntegration 集成测试
func TestIntegration(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 初始化日志系统
	err := setupTestLogger()
	assert.NoError(t, err)

	// 创建测试路由
	router := setupTestRouter()

	// 测试成功响应
	t.Run("TestSuccessResponse", func(t *testing.T) {
		testSuccessResponse(t, router)
	})

	// 测试错误响应
	t.Run("TestErrorResponse", func(t *testing.T) {
		testErrorResponse(t, router)
	})

	// 测试中间件集成
	t.Run("TestMiddlewareIntegration", func(t *testing.T) {
		testMiddlewareIntegration(t, router)
	})

	// 测试panic恢复
	t.Run("TestPanicRecovery", func(t *testing.T) {
		testPanicRecovery(t, router)
	})

	// 清理测试日志文件
	cleanupTestFiles()
}

// setupTestLogger 设置测试日志系统
func setupTestLogger() error {
	// 使用测试环境日志配置
	logConfig := logger.LogConfig{
		Level:      "debug",
		Format:     "json",
		Output:     "file",
		FilePath:   "test_logs/app.log",
		MaxSize:    10,
		MaxAge:     1,
		MaxBackups: 1,
		Compress:   false,
	}

	accessConfig := logger.AccessLogConfig{
		Enabled:  true,
		FilePath: "test_logs/access.log",
		Format:   "json",
	}

	// 确保日志目录存在
	if err := os.MkdirAll("test_logs", 0750); err != nil {
		return err
	}

	// 初始化日志系统
	if err := logger.InitLogger(logConfig); err != nil {
		return err
	}

	if err := logger.InitAccessLogger(accessConfig); err != nil {
		return err
	}

	logger.InitStructuredLogger()
	return nil
}

// setupTestRouter 设置测试路由
func setupTestRouter() *gin.Engine {
	router := gin.New()

	// 添加中间件
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())

	// 测试路由
	router.GET("/test/success", func(c *gin.Context) {
		c.Set("user_id", "test-user-123")

		// 记录结构化日志
		logger.LogUserAction(logger.UserAction{
			UserID:    "test-user-123",
			Action:    "view_page",
			Resource:  "test_page",
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})

		utils.Success(c, map[string]interface{}{
			"message": "测试成功",
			"data":    []string{"item1", "item2", "item3"},
		})
	})

	router.GET("/test/error", func(c *gin.Context) {
		c.Set("user_id", "test-user-456")

		// 记录安全事件
		logger.LogSecurityEvent(logger.SecurityEvent{
			EventType:   "access_denied",
			UserID:      "test-user-456",
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			Resource:    "restricted_resource",
			Severity:    "medium",
			Description: "用户尝试访问受限资源",
		})

		utils.ErrorWithMessage(c, utils.CodePermissionDenied, "您没有权限访问此资源")
	})

	router.GET("/test/validation", func(c *gin.Context) {
		// 验证错误示例
		errors := map[string]string{
			"email": "邮箱格式不正确",
			"phone": "手机号码格式不正确",
		}
		utils.ValidationError(c, errors)
	})

	router.GET("/test/panic", func(c *gin.Context) {
		// 触发panic测试恢复机制
		panic("这是一个测试panic")
	})

	router.GET("/test/list", func(c *gin.Context) {
		c.Set("user_id", "test-user-789")

		// 解析分页参数
		pageReq := utils.ParsePageRequest(c)

		// 模拟数据
		data := []map[string]interface{}{
			{"id": 1, "name": "文件1.txt", "size": 1024},
			{"id": 2, "name": "文件2.pdf", "size": 2048},
			{"id": 3, "name": "文件3.jpg", "size": 3072},
		}

		// 创建分页信息
		pagination := utils.NewPagination(pageReq.Page, pageReq.PageSize, 100)

		utils.SuccessList(c, data, pagination)
	})

	return router
}

// testSuccessResponse 测试成功响应
func testSuccessResponse(t *testing.T, router *gin.Engine) {
	req := httptest.NewRequest("GET", "/test/success", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response utils.Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, utils.CodeSuccess, response.Code)
	assert.Equal(t, "操作成功", response.Message)
	assert.NotEmpty(t, response.RequestID)
	assert.NotZero(t, response.Timestamp)
	assert.NotNil(t, response.Data)

	// 验证响应头包含请求ID
	assert.NotEmpty(t, recorder.Header().Get("X-Request-ID"))
}

// testErrorResponse 测试错误响应
func testErrorResponse(t *testing.T, router *gin.Engine) {
	req := httptest.NewRequest("GET", "/test/error", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code)

	var response utils.Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, utils.CodePermissionDenied, response.Code)
	assert.Equal(t, "您没有权限访问此资源", response.Message)
	assert.NotEmpty(t, response.RequestID)
}

// testMiddlewareIntegration 测试中间件集成
func testMiddlewareIntegration(t *testing.T, router *gin.Engine) {
	// 测试带分页参数的列表请求
	req := httptest.NewRequest("GET", "/test/list?page=2&page_size=10&sort_by=name&sort_dir=asc", nil)
	req.Header.Set("User-Agent", "TestClient/1.0")
	req.Header.Set("X-Forwarded-For", "192.168.1.100")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response utils.ListResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, utils.CodeSuccess, response.Code)
	assert.NotNil(t, response.Data)
	assert.NotNil(t, response.Pagination)
	assert.Equal(t, 2, response.Pagination.CurrentPage)
	assert.Equal(t, 10, response.Pagination.PageSize)
	assert.Equal(t, int64(100), response.Pagination.TotalCount)

	// 验证请求ID存在
	assert.NotEmpty(t, response.RequestID)
	assert.NotEmpty(t, recorder.Header().Get("X-Request-ID"))
}

// testPanicRecovery 测试panic恢复
func testPanicRecovery(t *testing.T, router *gin.Engine) {
	req := httptest.NewRequest("GET", "/test/panic", nil)
	recorder := httptest.NewRecorder()

	// 应该不会因为panic而崩溃
	assert.NotPanics(t, func() {
		router.ServeHTTP(recorder, req)
	})

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	// panic的响应可能不是标准的utils.Response格式
	// 只检查状态码和是否有响应内容
	assert.NotEmpty(t, recorder.Body.String())
}

// TestLoggerIntegration 测试日志系统集成
func TestLoggerIntegration(t *testing.T) {
	// 初始化测试日志
	err := setupTestLogger()
	assert.NoError(t, err)

	// 测试各种日志功能
	t.Run("TestBasicLogging", func(t *testing.T) {
		logger.Info("测试信息日志", zap.String("test", "value"))
		logger.Warn("测试警告日志", zap.String("warning", "test"))
		logger.Error("测试错误日志", zap.Error(fmt.Errorf("test error")))
	})

	t.Run("TestStructuredLogging", func(t *testing.T) {
		// 测试用户操作日志
		logger.LogUserAction(logger.UserAction{
			UserID:    "test-user",
			Action:    "login",
			IPAddress: "127.0.0.1",
			UserAgent: "Test Browser",
		})

		// 测试文件操作日志
		logger.LogFileOperation(logger.FileOperation{
			UserID:    "test-user",
			Operation: "upload",
			FileName:  "test.txt",
			FileSize:  1024,
			FileType:  "text/plain",
		})

		// 测试安全事件日志
		logger.LogSecurityEvent(logger.SecurityEvent{
			EventType:   "login_failed",
			IPAddress:   "127.0.0.1",
			Severity:    "high",
			Description: "多次登录失败",
		})

		// 测试系统事件日志
		logger.LogSystemEvent(logger.SystemEvent{
			Component: "database",
			Event:     "connection_established",
			Level:     "info",
			Message:   "数据库连接成功",
		})
	})

	// 同步日志缓冲区
	err = logger.Sync()
	assert.NoError(t, err)

	// 清理测试文件
	defer cleanupTestFiles()
}

// TestUtilsIntegration 测试工具函数集成
func TestUtilsIntegration(t *testing.T) {
	t.Run("TestStringUtils", func(t *testing.T) {
		// 测试随机字符串生成
		token, err := utils.GenerateSecureToken(32)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// 测试邮箱验证和脱敏
		email := "test@example.com"
		assert.True(t, utils.IsValidEmail(email))
		masked := utils.MaskEmail(email)
		assert.NotEqual(t, email, masked)
		assert.Contains(t, masked, "@example.com")

		// 测试字符串转换
		camelCase := utils.ToCamelCase("hello_world_test")
		assert.Equal(t, "helloWorldTest", camelCase)

		snakeCase := utils.ToSnakeCase("HelloWorldTest")
		assert.Equal(t, "hello_world_test", snakeCase)
	})

	t.Run("TestTimeUtils", func(t *testing.T) {
		// 测试时间格式化
		now := utils.FormatNow(utils.DateTimeLayout)
		assert.NotEmpty(t, now)

		// 测试时间解析
		parsed, err := utils.TryParseTime("2024-01-01 15:04:05")
		assert.NoError(t, err)
		assert.False(t, parsed.IsZero())

		// 测试人性化时间显示
		timeAgo := utils.TimeAgo(parsed)
		assert.NotEmpty(t, timeAgo)
	})

	t.Run("TestResponseUtils", func(t *testing.T) {
		// 测试分页创建
		pagination := utils.NewPagination(2, 20, 100)
		assert.Equal(t, 2, pagination.CurrentPage)
		assert.Equal(t, 20, pagination.PageSize)
		assert.Equal(t, int64(100), pagination.TotalCount)
		assert.Equal(t, 5, pagination.TotalPages)
		assert.True(t, pagination.HasPrevious)
		assert.True(t, pagination.HasNext)
		assert.Equal(t, 1, pagination.PreviousPage)
		assert.Equal(t, 3, pagination.NextPage)

		// 测试错误码
		assert.Equal(t, "操作成功", utils.CodeSuccess.GetMessage())
		assert.Equal(t, http.StatusOK, utils.CodeSuccess.GetHTTPStatus())
		assert.Equal(t, "数据验证失败", utils.CodeValidationError.GetMessage())
		assert.Equal(t, http.StatusBadRequest, utils.CodeValidationError.GetHTTPStatus())
	})
}

// cleanupTestFiles 清理测试文件
func cleanupTestFiles() {
	if err := os.RemoveAll("test_logs"); err != nil {
		log.Printf("Failed to cleanup test files: %v", err)
	}
}

// BenchmarkIntegration 性能基准测试
func BenchmarkIntegration(b *testing.B) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	setupTestLogger()
	router := setupTestRouter()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test/success", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}

	cleanupTestFiles()
}
