package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"cloudpan/internal/pkg/utils"
)

// 测试用的JWT密钥
const testJWTSecret = "test-jwt-secret-key-for-unit-testing-very-long-secret"

func setupTestAuthMiddleware() *AuthMiddleware {
	logger := zap.NewNop()
	middleware, _ := NewAuthMiddleware(testJWTSecret, logger)
	return middleware
}

func generateTestTokens() (string, string, error) {
	jwtManager, err := utils.NewDefaultJWTManager(testJWTSecret)
	if err != nil {
		return "", "", err
	}

	accessToken, err := jwtManager.GenerateAccessToken(1, "testuser", "test@example.com", "user")
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwtManager.GenerateRefreshToken(1, "testuser", "test@example.com", "user")
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func TestNewAuthMiddleware(t *testing.T) {
	logger := zap.NewNop()

	t.Run("成功创建认证中间件", func(t *testing.T) {
		middleware, err := NewAuthMiddleware(testJWTSecret, logger)
		assert.NoError(t, err)
		assert.NotNil(t, middleware)
	})

	t.Run("JWT密钥过短时失败", func(t *testing.T) {
		middleware, err := NewAuthMiddleware("short", logger)
		assert.Error(t, err)
		assert.Nil(t, middleware)
	})
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authMiddleware := setupTestAuthMiddleware()

	t.Run("有效的访问令牌", func(t *testing.T) {
		accessToken, _, err := generateTestTokens()
		assert.NoError(t, err)

		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.GET("/protected", func(c *gin.Context) {
			// 验证用户信息是否已存储到上下文
			userID, exists := c.Get("user_id")
			assert.True(t, exists)
			assert.Equal(t, uint64(1), userID)

			username, exists := c.Get("username")
			assert.True(t, exists)
			assert.Equal(t, "testuser", username)

			email, exists := c.Get("email")
			assert.True(t, exists)
			assert.Equal(t, "test@example.com", email)

			role, exists := c.Get("role")
			assert.True(t, exists)
			assert.Equal(t, "user", role)

			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("缺少Authorization头", func(t *testing.T) {
		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求（不设置Authorization头）
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("无效的Bearer格式", func(t *testing.T) {
		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求（无效的Bearer格式）
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("无效的令牌", func(t *testing.T) {
		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求（无效令牌）
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("使用刷新令牌而非访问令牌", func(t *testing.T) {
		_, refreshToken, err := generateTestTokens()
		assert.NoError(t, err)

		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求（使用刷新令牌）
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+refreshToken)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthMiddleware_OptionalAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authMiddleware := setupTestAuthMiddleware()

	t.Run("有效的访问令牌", func(t *testing.T) {
		accessToken, _, err := generateTestTokens()
		assert.NoError(t, err)

		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.OptionalAuth())
		router.GET("/optional", func(c *gin.Context) {
			// 验证用户信息是否已存储到上下文
			userID, exists := c.Get("user_id")
			assert.True(t, exists)
			assert.Equal(t, uint64(1), userID)

			c.JSON(http.StatusOK, gin.H{"message": "success", "authenticated": true})
		})

		// 创建请求
		req := httptest.NewRequest("GET", "/optional", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("没有令牌时继续处理", func(t *testing.T) {
		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.OptionalAuth())
		router.GET("/optional", func(c *gin.Context) {
			// 验证用户信息不存在
			_, exists := c.Get("user_id")
			assert.False(t, exists)

			c.JSON(http.StatusOK, gin.H{"message": "success", "authenticated": false})
		})

		// 创建请求（不设置Authorization头）
		req := httptest.NewRequest("GET", "/optional", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("无效令牌时继续处理", func(t *testing.T) {
		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.OptionalAuth())
		router.GET("/optional", func(c *gin.Context) {
			// 验证用户信息不存在
			_, exists := c.Get("user_id")
			assert.False(t, exists)

			c.JSON(http.StatusOK, gin.H{"message": "success", "authenticated": false})
		})

		// 创建请求（无效令牌）
		req := httptest.NewRequest("GET", "/optional", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authMiddleware := setupTestAuthMiddleware()

	t.Run("用户角色匹配", func(t *testing.T) {
		accessToken, _, err := generateTestTokens()
		assert.NoError(t, err)

		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.Use(authMiddleware.RequireRole("user"))
		router.GET("/admin", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求
		req := httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("用户角色不匹配", func(t *testing.T) {
		accessToken, _, err := generateTestTokens()
		assert.NoError(t, err)

		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.Use(authMiddleware.RequireRole("admin")) // 需要admin角色，但用户是user
		router.GET("/admin", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求
		req := httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("缺少认证信息", func(t *testing.T) {
		// 创建测试路由（跳过认证中间件）
		router := gin.New()
		router.Use(authMiddleware.RequireRole("user"))
		router.GET("/admin", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求
		req := httptest.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthMiddleware_RequireAnyRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authMiddleware := setupTestAuthMiddleware()

	t.Run("用户具有其中一个角色", func(t *testing.T) {
		accessToken, _, err := generateTestTokens()
		assert.NoError(t, err)

		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.Use(authMiddleware.RequireAnyRole("admin", "user", "moderator"))
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("用户不具有任何所需角色", func(t *testing.T) {
		accessToken, _, err := generateTestTokens()
		assert.NoError(t, err)

		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.Use(authMiddleware.RequireAnyRole("admin", "superuser")) // 用户是user，不匹配
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAuthMiddleware_HelperFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authMiddleware := setupTestAuthMiddleware()

	t.Run("GetCurrentUser辅助函数", func(t *testing.T) {
		accessToken, _, err := generateTestTokens()
		assert.NoError(t, err)

		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.GET("/user", func(c *gin.Context) {
			user := GetCurrentUser(c)
			assert.NotNil(t, user)
			assert.Equal(t, uint64(1), user.UserID)
			assert.Equal(t, "testuser", user.Username)
			assert.Equal(t, "test@example.com", user.Email)
			assert.Equal(t, "user", user.Role)

			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求
		req := httptest.NewRequest("GET", "/user", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetCurrentUserID辅助函数", func(t *testing.T) {
		accessToken, _, err := generateTestTokens()
		assert.NoError(t, err)

		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.GET("/user", func(c *gin.Context) {
			userID, exists := GetCurrentUserID(c)
			assert.True(t, exists)
			assert.Equal(t, uint64(1), userID)

			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求
		req := httptest.NewRequest("GET", "/user", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("IsAuthenticated辅助函数", func(t *testing.T) {
		accessToken, _, err := generateTestTokens()
		assert.NoError(t, err)

		// 创建测试路由
		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.GET("/user", func(c *gin.Context) {
			authenticated := IsAuthenticated(c)
			assert.True(t, authenticated)

			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 创建请求
		req := httptest.NewRequest("GET", "/user", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAuthMiddleware_ExtractToken(t *testing.T) {
	authMiddleware := setupTestAuthMiddleware()

	t.Run("正确提取Bearer令牌", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer test-token-123")

		token := authMiddleware.extractToken(c)
		assert.Equal(t, "test-token-123", token)
	})

	t.Run("缺少Authorization头", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/test", nil)

		token := authMiddleware.extractToken(c)
		assert.Empty(t, token)
	})

	t.Run("无效的Bearer格式", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Basic dGVzdDp0ZXN0")

		token := authMiddleware.extractToken(c)
		assert.Empty(t, token)
	})

	t.Run("Bearer后无令牌", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer ")

		token := authMiddleware.extractToken(c)
		assert.Empty(t, token)
	})
}

func TestAuthMiddleware_HasRole(t *testing.T) {
	authMiddleware := setupTestAuthMiddleware()

	t.Run("相同角色", func(t *testing.T) {
		assert.True(t, authMiddleware.hasRole("user", "user"))
		assert.True(t, authMiddleware.hasRole("admin", "admin"))
	})

	t.Run("角色层次结构", func(t *testing.T) {
		assert.True(t, authMiddleware.hasRole("admin", "user"))       // admin >= user
		assert.True(t, authMiddleware.hasRole("admin", "moderator"))  // admin >= moderator
		assert.False(t, authMiddleware.hasRole("user", "admin"))      // user < admin
		assert.False(t, authMiddleware.hasRole("moderator", "admin")) // moderator < admin
	})

	t.Run("未知角色精确匹配", func(t *testing.T) {
		assert.True(t, authMiddleware.hasRole("custom", "custom"))
		assert.False(t, authMiddleware.hasRole("custom", "user"))
		assert.False(t, authMiddleware.hasRole("user", "custom"))
	})
}
