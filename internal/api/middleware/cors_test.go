package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TestDefaultCORS", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		// 测试预检请求
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
		assert.Equal(t, "https://example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, recorder.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, recorder.Header().Get("Access-Control-Allow-Headers"), "Authorization")
		assert.Equal(t, "true", recorder.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("TestCORSWithOrigin", func(t *testing.T) {
		router := gin.New()
		opts := &CORSOptions{
			AllowedOrigins:   []string{"https://example.com", "https://test.com"},
			AllowedMethods:   []string{"GET", "POST"},
			AllowCredentials: true,
		}
		router.Use(CORS(opts))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		// 测试允许的源
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "https://example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", recorder.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("TestCORSWithDisallowedOrigin", func(t *testing.T) {
		router := gin.New()
		opts := &CORSOptions{
			AllowedOrigins: []string{"https://example.com"},
			AllowedMethods: []string{"GET", "POST"},
		}
		router.Use(CORS(opts))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		// 测试不允许的源
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://malicious.com")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "", recorder.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("TestCORSWildcardSubdomain", func(t *testing.T) {
		router := gin.New()
		opts := &CORSOptions{
			AllowedOrigins: []string{"*.example.com"},
			AllowedMethods: []string{"GET"},
		}
		router.Use(CORS(opts))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		// 测试子域名匹配
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://api.example.com")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "https://api.example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("TestProductionCORS", func(t *testing.T) {
		router := gin.New()
		router.Use(ProductionCORS([]string{"https://cloudpan.hxlos.com"}))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://cloudpan.hxlos.com")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "https://cloudpan.hxlos.com", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", recorder.Header().Get("Access-Control-Allow-Credentials"))
	})
}

func TestIsOriginAllowed(t *testing.T) {
	t.Run("TestExactMatch", func(t *testing.T) {
		allowed := []string{"https://example.com", "https://test.com"}
		assert.True(t, isOriginAllowed("https://example.com", allowed))
		assert.False(t, isOriginAllowed("https://other.com", allowed))
	})

	t.Run("TestWildcardMatch", func(t *testing.T) {
		allowed := []string{"*"}
		assert.True(t, isOriginAllowed("https://any.com", allowed))
		assert.True(t, isOriginAllowed("http://localhost:3000", allowed))
	})

	t.Run("TestSubdomainMatch", func(t *testing.T) {
		allowed := []string{"*.example.com"}
		assert.True(t, isOriginAllowed("https://api.example.com", allowed))
		assert.True(t, isOriginAllowed("https://www.example.com", allowed))
		assert.True(t, isOriginAllowed("https://example.com", allowed))
		assert.False(t, isOriginAllowed("https://example.org", allowed))
		assert.False(t, isOriginAllowed("https://malicious-example.com", allowed))
	})

	t.Run("TestEmptyOrigin", func(t *testing.T) {
		allowed := []string{"https://example.com"}
		assert.False(t, isOriginAllowed("", allowed))
	})
}

func TestDefaultCORSOptions(t *testing.T) {
	opts := DefaultCORSOptions()

	assert.Equal(t, []string{"*"}, opts.AllowedOrigins)
	assert.Contains(t, opts.AllowedMethods, "GET")
	assert.Contains(t, opts.AllowedMethods, "POST")
	assert.Contains(t, opts.AllowedMethods, "PUT")
	assert.Contains(t, opts.AllowedMethods, "DELETE")
	assert.Contains(t, opts.AllowedHeaders, "Authorization")
	assert.Contains(t, opts.AllowedHeaders, "Content-Type")
	assert.True(t, opts.AllowCredentials)
	assert.Equal(t, 86400, opts.MaxAge)
}
