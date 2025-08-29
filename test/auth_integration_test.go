package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"cloudpan/internal/api/middleware"
	"cloudpan/internal/pkg/utils"
)

// AuthIntegrationTest 认证集成测试套件
type AuthIntegrationTest struct {
	router     *gin.Engine
	authMW     *middleware.AuthMiddleware
	testSecret string
}

// setupAuthIntegrationTest 设置认证集成测试环境
func setupAuthIntegrationTest() *AuthIntegrationTest {
	gin.SetMode(gin.TestMode)

	testSecret := "integration-test-jwt-secret-key-very-long-for-security"
	logger := zap.NewNop()

	// 创建认证中间件
	authMW, _ := middleware.NewAuthMiddleware(testSecret, logger)

	// 创建路由
	router := gin.New()
	router.Use(gin.Recovery())

	return &AuthIntegrationTest{
		router:     router,
		authMW:     authMW,
		testSecret: testSecret,
	}
}

// TestAuthenticationFlow 测试完整的认证流程
func TestAuthenticationFlow(t *testing.T) {
	test := setupAuthIntegrationTest()

	// 模拟用户数据
	testUser := map[string]interface{}{
		"user_id":  1,
		"username": "testuser",
		"email":    "test@example.com",
		"role":     "user",
	}

	// 1. 生成JWT令牌
	jwtManager, err := utils.NewDefaultJWTManager(test.testSecret)
	assert.NoError(t, err)

	accessToken, err := jwtManager.GenerateAccessToken(
		uint64(testUser["user_id"].(int)),
		testUser["username"].(string),
		testUser["email"].(string),
		testUser["role"].(string),
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	refreshToken, err := jwtManager.GenerateRefreshToken(
		uint64(testUser["user_id"].(int)),
		testUser["username"].(string),
		testUser["email"].(string),
		testUser["role"].(string),
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	// 2. 测试无需认证的端点
	test.router.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "public endpoint"})
	})

	req := httptest.NewRequest("GET", "/public", nil)
	w := httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. 测试需要认证的端点
	test.router.GET("/protected", test.authMW.RequireAuth(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{
			"message":  "protected endpoint",
			"user_id":  userID,
			"username": username,
		})
	})

	// 3.1 无令牌访问受保护端点
	req = httptest.NewRequest("GET", "/protected", nil)
	w = httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 3.2 使用有效令牌访问受保护端点
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "protected endpoint", response["message"])
	assert.Equal(t, float64(1), response["user_id"]) // JSON数字解析为float64
	assert.Equal(t, "testuser", response["username"])

	// 4. 测试角色验证
	test.router.GET("/admin", test.authMW.RequireAuth(), test.authMW.RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin endpoint"})
	})

	// 4.1 用户角色访问管理员端点（应该失败）
	req = httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// 4.2 生成管理员令牌并测试
	adminToken, err := jwtManager.GenerateAccessToken(2, "admin", "admin@example.com", "admin")
	assert.NoError(t, err)

	req = httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. 测试令牌刷新
	newAccessToken, newRefreshToken, err := jwtManager.RefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, accessToken, newAccessToken)
	assert.NotEqual(t, refreshToken, newRefreshToken)

	// 验证新令牌可以正常使用
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+newAccessToken)
	w = httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestTokenSecurity 测试令牌安全性
func TestTokenSecurity(t *testing.T) {
	test := setupAuthIntegrationTest()
	jwtManager, _ := utils.NewDefaultJWTManager(test.testSecret)

	// 设置受保护端点
	test.router.GET("/secure", test.authMW.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "secure"})
	})

	// 1. 测试伪造的令牌
	fakeToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	req := httptest.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+fakeToken)
	w := httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 2. 测试过期令牌
	// 生成一个已过期的令牌（过期时间为1毫秒）
	expiredManager, _ := utils.NewJWTManager(test.testSecret, time.Millisecond, time.Hour)
	expiredToken, _ := expiredManager.GenerateAccessToken(1, "test", "test@example.com", "user")

	// 等待令牌过期
	time.Sleep(100 * time.Millisecond)

	req = httptest.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w = httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 3. 测试错误的令牌类型（使用刷新令牌作为访问令牌）
	refreshToken, _ := jwtManager.GenerateRefreshToken(1, "test", "test@example.com", "user")

	req = httptest.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	w = httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 4. 测试恶意修改的令牌
	validToken, _ := jwtManager.GenerateAccessToken(1, "test", "test@example.com", "user")

	// 尝试修改令牌的最后几个字符
	maliciousToken := validToken[:len(validToken)-5] + "XXXXX"

	req = httptest.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+maliciousToken)
	w = httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestConcurrentAuthentication 测试并发认证安全性
func TestConcurrentAuthentication(t *testing.T) {
	test := setupAuthIntegrationTest()
	jwtManager, _ := utils.NewDefaultJWTManager(test.testSecret)

	// 设置受保护端点
	test.router.GET("/concurrent", test.authMW.RequireAuth(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	// 生成多个用户的令牌
	tokens := make([]string, 10)
	for i := 0; i < 10; i++ {
		token, err := jwtManager.GenerateAccessToken(
			uint64(i+1),
			fmt.Sprintf("user%d", i+1),
			fmt.Sprintf("user%d@example.com", i+1),
			"user",
		)
		assert.NoError(t, err)
		tokens[i] = token
	}

	// 并发测试令牌验证
	results := make([]int, 10)
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			defer func() { done <- true }()

			req := httptest.NewRequest("GET", "/concurrent", nil)
			req.Header.Set("Authorization", "Bearer "+tokens[index])
			w := httptest.NewRecorder()
			test.router.ServeHTTP(w, req)

			results[index] = w.Code
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证所有请求都成功
	for i, code := range results {
		assert.Equal(t, http.StatusOK, code, "User %d token validation failed", i+1)
	}
}

// TestRoleHierarchy 测试角色层次结构
func TestRoleHierarchy(t *testing.T) {
	test := setupAuthIntegrationTest()
	jwtManager, _ := utils.NewDefaultJWTManager(test.testSecret)

	// 设置不同权限级别的端点
	test.router.GET("/user-endpoint", test.authMW.RequireAuth(), test.authMW.RequireRole("user"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "user endpoint"})
	})

	test.router.GET("/moderator-endpoint", test.authMW.RequireAuth(), test.authMW.RequireRole("moderator"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "moderator endpoint"})
	})

	test.router.GET("/admin-endpoint", test.authMW.RequireAuth(), test.authMW.RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin endpoint"})
	})

	// 生成不同角色的令牌
	userToken, _ := jwtManager.GenerateAccessToken(1, "user", "user@example.com", "user")
	moderatorToken, _ := jwtManager.GenerateAccessToken(2, "moderator", "moderator@example.com", "moderator")
	adminToken, _ := jwtManager.GenerateAccessToken(3, "admin", "admin@example.com", "admin")

	// 测试用户权限
	testCases := []struct {
		token    string
		endpoint string
		expected int
		desc     string
	}{
		{userToken, "/user-endpoint", http.StatusOK, "用户访问用户端点"},
		{userToken, "/moderator-endpoint", http.StatusForbidden, "用户访问版主端点"},
		{userToken, "/admin-endpoint", http.StatusForbidden, "用户访问管理员端点"},

		{moderatorToken, "/user-endpoint", http.StatusOK, "版主访问用户端点"},
		{moderatorToken, "/moderator-endpoint", http.StatusOK, "版主访问版主端点"},
		{moderatorToken, "/admin-endpoint", http.StatusForbidden, "版主访问管理员端点"},

		{adminToken, "/user-endpoint", http.StatusOK, "管理员访问用户端点"},
		{adminToken, "/moderator-endpoint", http.StatusOK, "管理员访问版主端点"},
		{adminToken, "/admin-endpoint", http.StatusOK, "管理员访问管理员端点"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			w := httptest.NewRecorder()
			test.router.ServeHTTP(w, req)
			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

// TestOptionalAuthentication 测试可选认证
func TestOptionalAuthentication(t *testing.T) {
	test := setupAuthIntegrationTest()
	jwtManager, _ := utils.NewDefaultJWTManager(test.testSecret)

	// 设置可选认证端点
	test.router.GET("/optional", test.authMW.OptionalAuth(), func(c *gin.Context) {
		userID, exists := middleware.GetCurrentUserID(c)
		if exists {
			c.JSON(http.StatusOK, gin.H{"authenticated": true, "user_id": userID})
		} else {
			c.JSON(http.StatusOK, gin.H{"authenticated": false})
		}
	})

	// 1. 无令牌访问
	req := httptest.NewRequest("GET", "/optional", nil)
	w := httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response["authenticated"].(bool))

	// 2. 有效令牌访问
	token, _ := jwtManager.GenerateAccessToken(1, "test", "test@example.com", "user")
	req = httptest.NewRequest("GET", "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response["authenticated"].(bool))
	assert.Equal(t, float64(1), response["user_id"])

	// 3. 无效令牌访问（应该继续处理）
	req = httptest.NewRequest("GET", "/optional", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	test.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response["authenticated"].(bool))
}
