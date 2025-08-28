package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestLoggerBasic(t *testing.T) {
	// 设置gin为测试模式
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		setupHandler   func(c *gin.Context)
		expectedStatus int
	}{
		{
			name:   "GET request success",
			method: "GET",
			path:   "/api/test",
			setupHandler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST request",
			method: "POST",
			path:   "/api/create",
			setupHandler: func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"created": true})
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建路由器
			r := gin.New()
			r.Use(RequestLogger())

			// 添加测试路由
			r.Handle(tt.method, tt.path, tt.setupHandler)

			// 创建测试请求
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// 执行请求
			r.ServeHTTP(w, req)

			// 验证状态码
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
