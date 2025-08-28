package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"cloudpan/internal/pkg/config"
	"cloudpan/internal/pkg/database"
)

func TestMain(m *testing.M) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 设置配置文件路径并加载测试配置
	os.Setenv("APP_ENV", "test")
	os.Chdir("../../../") // 切换到项目根目录
	if err := config.Load(); err != nil {
		// 配置加载失败不影响单元测试，使用默认配置
		config.AppConfig = &config.Config{
			App: config.App{
				Name:    "cloudpan",
				Version: "1.0.0",
				Env:     "test",
				Debug:   true,
			},
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8080,
			},
		}
	}

	// 初始化数据库（如果需要）
	if err := database.Init(); err != nil {
		// 数据库初始化失败不影响单元测试
	}

	code := m.Run()

	// 清理
	database.Shutdown()
	os.Exit(code)
}

func TestSetupRouter(t *testing.T) {
	router := SetupRouter()
	assert.NotNil(t, router)
}

func TestHealthCheckHandler(t *testing.T) {
	router := SetupRouter()

	t.Run("TestHealthCheck", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
		assert.NotEmpty(t, response["message"])
		assert.NotEmpty(t, response["version"])
		assert.NotZero(t, response["timestamp"])
	})
}

func TestDatabaseHealthHandler(t *testing.T) {
	router := SetupRouter()

	t.Run("TestDatabaseHealth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health/database", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		// 数据库可能未连接，但应该返回响应
		assert.True(t, recorder.Code == http.StatusOK || recorder.Code == http.StatusServiceUnavailable)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
		assert.NotNil(t, response["databases"])
	})
}

func TestSystemStatsHandler(t *testing.T) {
	router := SetupRouter()

	t.Run("TestSystemStats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/system/stats", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
		assert.NotEmpty(t, response["message"])
		assert.NotNil(t, response["data"])

		// 检查数据结构
		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["application"])
		assert.NotNil(t, data["server"])
		assert.NotNil(t, data["database"])
		assert.NotZero(t, data["timestamp"])
	})
}

func TestAPIVersionRoutes(t *testing.T) {
	router := SetupRouter()

	t.Run("TestVersionInfo", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/system/version", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response["data"])

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["current_version"])
		assert.NotNil(t, data["supported_versions"])
		assert.NotEmpty(t, data["default_version"])
	})

	t.Run("TestV2Route", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/system/version", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestLanguageRoutes(t *testing.T) {
	router := SetupRouter()

	t.Run("TestLanguageInfo", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/system/language", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response["data"])

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["current_language"])
		assert.NotNil(t, data["supported_languages"])
		assert.NotEmpty(t, data["default_language"])
	})

	t.Run("TestLanguageWithParam", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/system/language?lang=en-US", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "en-US", recorder.Header().Get("Content-Language"))
	})
}

func TestCORSMiddleware(t *testing.T) {
	router := SetupRouter()

	t.Run("TestPreflightRequest", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/v1/system/stats", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
		assert.NotEmpty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(t, recorder.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("TestCORSHeaders", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/system/stats", nil)
		req.Header.Set("Origin", "https://example.com")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		// 在debug模式下，应该允许所有源
		corsHeader := recorder.Header().Get("Access-Control-Allow-Origin")
		assert.True(t, corsHeader == "*" || corsHeader == "https://example.com")
	})
}

func TestBusinessRoutes(t *testing.T) {
	router := SetupRouter()

	t.Run("TestUserRoutes", func(t *testing.T) {
		// 测试用户列表
		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// 测试创建用户
		req = httptest.NewRequest("POST", "/api/v1/users", nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// 测试获取用户详情
		req = httptest.NewRequest("GET", "/api/v1/users/123", nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("TestFileRoutes", func(t *testing.T) {
		// 测试文件列表
		req := httptest.NewRequest("GET", "/api/v1/files", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// 测试文件上传
		req = httptest.NewRequest("POST", "/api/v1/files/upload", nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// 测试文件下载
		req = httptest.NewRequest("GET", "/api/v1/files/123/download", nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("TestTeamRoutes", func(t *testing.T) {
		// 测试团队列表
		req := httptest.NewRequest("GET", "/api/v1/teams", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// 测试创建团队
		req = httptest.NewRequest("POST", "/api/v1/teams", nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("TestMessageRoutes", func(t *testing.T) {
		// 测试消息列表
		req := httptest.NewRequest("GET", "/api/v1/messages", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// 测试发送消息
		req = httptest.NewRequest("POST", "/api/v1/messages", nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestMiddlewareIntegration(t *testing.T) {
	router := SetupRouter()

	t.Run("TestRequestIDMiddleware", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.NotEmpty(t, recorder.Header().Get("X-Request-ID"))
	})

	t.Run("TestAPIVersionMiddleware", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/system/stats", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "v1", recorder.Header().Get("API-Version"))
		assert.NotEmpty(t, recorder.Header().Get("API-Supported-Versions"))
	})

	t.Run("TestI18nMiddleware", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		req.Header.Set("Accept-Language", "en-US")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "en-US", recorder.Header().Get("Content-Language"))
	})
}
