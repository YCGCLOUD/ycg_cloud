package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPIVersionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TestDefaultVersion", func(t *testing.T) {
		router := gin.New()
		router.Use(APIVersionMiddleware())
		router.GET("/test", func(c *gin.Context) {
			version := GetAPIVersion(c)
			c.JSON(200, gin.H{"version": version})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "v1", recorder.Header().Get("API-Version"))

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "v1", response["version"])
	})

	t.Run("TestVersionFromPath", func(t *testing.T) {
		router := gin.New()
		router.Use(APIVersionMiddleware())
		router.GET("/api/v2/test", func(c *gin.Context) {
			version := GetAPIVersion(c)
			c.JSON(200, gin.H{"version": version})
		})

		req := httptest.NewRequest("GET", "/api/v2/test", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "v2", recorder.Header().Get("API-Version"))

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "v2", response["version"])
	})

	t.Run("TestVersionFromHeader", func(t *testing.T) {
		router := gin.New()
		router.Use(APIVersionMiddleware())
		router.GET("/test", func(c *gin.Context) {
			version := GetAPIVersion(c)
			c.JSON(200, gin.H{"version": version})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("API-Version", "v2")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "v2", recorder.Header().Get("API-Version"))
	})

	t.Run("TestVersionFromQuery", func(t *testing.T) {
		router := gin.New()
		router.Use(APIVersionMiddleware())
		router.GET("/test", func(c *gin.Context) {
			version := GetAPIVersion(c)
			c.JSON(200, gin.H{"version": version})
		})

		req := httptest.NewRequest("GET", "/test?version=v2", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "v2", recorder.Header().Get("API-Version"))
	})

	t.Run("TestUnsupportedVersion", func(t *testing.T) {
		router := gin.New()
		router.Use(APIVersionMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/test?version=v99", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("TestDeprecatedVersion", func(t *testing.T) {
		config := DefaultAPIVersionConfig()
		router := gin.New()
		router.Use(APIVersionMiddleware(config))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/test?version=v1", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "true", recorder.Header().Get("API-Deprecated"))
		assert.Equal(t, "v1", recorder.Header().Get("API-Deprecated-Version"))
		assert.Equal(t, "v2", recorder.Header().Get("API-Recommended-Version"))
		assert.Contains(t, recorder.Header().Get("Warning"), "deprecated")
	})

	t.Run("TestCustomConfig", func(t *testing.T) {
		config := &APIVersionConfig{
			DefaultVersion:    "v2",
			SupportedVersions: []string{"v2", "v3"},
			VersionHeader:     "X-API-Version",
			VersionParam:      "api_version",
		}

		router := gin.New()
		router.Use(APIVersionMiddleware(config))
		router.GET("/test", func(c *gin.Context) {
			version := GetAPIVersion(c)
			c.JSON(200, gin.H{"version": version})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "v2", recorder.Header().Get("API-Version"))
	})
}

func TestExtractVersionFromPath(t *testing.T) {
	t.Run("TestValidPath", func(t *testing.T) {
		version := extractVersionFromPath("/api/v1/users", "/api/")
		assert.Equal(t, "v1", version)

		version = extractVersionFromPath("/api/v2/files", "/api/")
		assert.Equal(t, "v2", version)
	})

	t.Run("TestInvalidPath", func(t *testing.T) {
		version := extractVersionFromPath("/users", "/api/")
		assert.Equal(t, "", version)

		version = extractVersionFromPath("/api/users", "/api/")
		assert.Equal(t, "", version)
	})
}

func TestNormalizeVersion(t *testing.T) {
	t.Run("TestNormalizeVersion", func(t *testing.T) {
		assert.Equal(t, "v1", normalizeVersion("1"))
		assert.Equal(t, "v1", normalizeVersion("v1"))
		assert.Equal(t, "v2", normalizeVersion("V2"))
		assert.Equal(t, "v1", normalizeVersion(" v1 "))
		assert.Equal(t, "", normalizeVersion(""))
	})
}

func TestIsVersionSupported(t *testing.T) {
	supportedVersions := []string{"v1", "v2"}

	assert.True(t, isVersionSupported("v1", supportedVersions))
	assert.True(t, isVersionSupported("v2", supportedVersions))
	assert.False(t, isVersionSupported("v3", supportedVersions))
	assert.False(t, isVersionSupported("", supportedVersions))
}

func TestVersionInfoHandler(t *testing.T) {
	router := gin.New()
	router.Use(APIVersionMiddleware())
	router.GET("/version", VersionInfoHandler())

	req := httptest.NewRequest("GET", "/version", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "v1", data["current_version"])
	assert.Equal(t, "v1", data["default_version"])
	assert.Contains(t, data["supported_versions"], "v1")
	assert.Contains(t, data["supported_versions"], "v2")
}

func TestIsVersionDeprecated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	config := DefaultAPIVersionConfig()
	c.Set("api_version_config", config)

	deprecated, recommended := IsVersionDeprecated(c, "v1")
	assert.True(t, deprecated)
	assert.Equal(t, "v2", recommended)

	deprecated, _ = IsVersionDeprecated(c, "v2")
	assert.False(t, deprecated)
}
