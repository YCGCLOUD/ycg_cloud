package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandlerBasic(t *testing.T) {
	// 设置gin为测试模式
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupHandler   func(c *gin.Context)
		expectedStatus int
	}{
		{
			name: "No errors - success response",
			setupHandler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Error with abort",
			setupHandler: func(c *gin.Context) {
				c.AbortWithStatus(http.StatusBadRequest)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建路由器
			r := gin.New()
			r.Use(ErrorHandler())

			// 添加测试路由
			r.GET("/test", tt.setupHandler)

			// 创建测试请求
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// 执行请求
			r.ServeHTTP(w, req)

			// 验证状态码
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
