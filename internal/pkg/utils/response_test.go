package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestGin() (*gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	recorder := httptest.NewRecorder()
	return router, recorder
}

func TestResponseCode_GetMessage(t *testing.T) {
	tests := []struct {
		code     ResponseCode
		expected string
	}{
		{CodeSuccess, "操作成功"},
		{CodeBadRequest, "请求参数错误"},
		{CodeUnauthorized, "未认证"},
		{CodeValidationError, "数据验证失败"},
		{ResponseCode(9999), "未知错误"}, // unknown code
	}

	for _, tt := range tests {
		result := tt.code.GetMessage()
		assert.Equal(t, tt.expected, result)
	}
}

func TestResponseCode_GetHTTPStatus(t *testing.T) {
	tests := []struct {
		code     ResponseCode
		expected int
	}{
		{CodeSuccess, http.StatusOK},
		{CodeBadRequest, http.StatusBadRequest},
		{CodeUnauthorized, http.StatusUnauthorized},
		{CodeForbidden, http.StatusForbidden},
		{CodeNotFound, http.StatusNotFound},
		{CodeInternalError, http.StatusInternalServerError},
		{CodeValidationError, http.StatusBadRequest},
		{CodeDuplicateData, http.StatusConflict},
		{CodeDataNotFound, http.StatusNotFound},
		{CodeInvalidToken, http.StatusUnauthorized},
		{CodeTokenExpired, http.StatusUnauthorized},
		{CodePermissionDenied, http.StatusForbidden},
		{CodeQuotaExceeded, http.StatusForbidden},
		{ResponseCode(9999), http.StatusInternalServerError}, // unknown code
	}

	for _, tt := range tests {
		result := tt.code.GetHTTPStatus()
		assert.Equal(t, tt.expected, result)
	}
}

func TestSuccess(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		Success(c, map[string]string{"message": "hello"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "操作成功", response.Message)
	assert.Equal(t, "test-request-id", response.RequestID)
	assert.NotZero(t, response.Timestamp)
	assert.NotNil(t, response.Data)
}

func TestSuccessWithMessage(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		SuccessWithMessage(c, "自定义成功消息", "test data")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "自定义成功消息", response.Message)
	assert.Equal(t, "test data", response.Data)
}

func TestError(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		Error(c, CodeValidationError)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeValidationError, response.Code)
	assert.Equal(t, "数据验证失败", response.Message)
	assert.Equal(t, "test-request-id", response.RequestID)
}

func TestErrorWithMessage(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		ErrorWithMessage(c, CodeNotFound, "用户不存在")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeNotFound, response.Code)
	assert.Equal(t, "用户不存在", response.Message)
}

func TestErrorWithData(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		ErrorWithData(c, CodeValidationError, "验证失败", map[string]string{"field": "email"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeValidationError, response.Code)
	assert.Equal(t, "验证失败", response.Message)
	assert.NotNil(t, response.Data)
}

func TestValidationError(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		ValidationError(c, map[string]string{"email": "invalid format"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeValidationError, response.Code)
	assert.Equal(t, "数据验证失败", response.Message)
}

func TestUnauthorized(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		Unauthorized(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeUnauthorized, response.Code)
}

func TestForbidden(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		Forbidden(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeForbidden, response.Code)
}

func TestNotFound(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		NotFound(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeNotFound, response.Code)
}

func TestNotFoundWithMessage(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		NotFoundWithMessage(c, "文件不存在")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeNotFound, response.Code)
	assert.Equal(t, "文件不存在", response.Message)
}

func TestInternalError(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		InternalError(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeInternalError, response.Code)
}

func TestInternalErrorWithMessage(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		InternalErrorWithMessage(c, "数据库连接失败")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeInternalError, response.Code)
	assert.Equal(t, "数据库连接失败", response.Message)
}

func TestSuccessList(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		data := []string{"item1", "item2", "item3"}
		pagination := NewPagination(1, 10, 100)
		SuccessList(c, data, pagination)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response ListResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "操作成功", response.Message)
	assert.NotNil(t, response.Data)
	assert.NotNil(t, response.Pagination)
	assert.Equal(t, "test-request-id", response.RequestID)
}

func TestSuccessListWithMessage(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		data := []string{"item1", "item2"}
		pagination := NewPagination(1, 10, 2)
		SuccessListWithMessage(c, "获取列表成功", data, pagination)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response ListResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "获取列表成功", response.Message)
}

func TestNewPagination(t *testing.T) {
	tests := []struct {
		currentPage int
		pageSize    int
		totalCount  int64
		expected    *Pagination
	}{
		{
			currentPage: 1,
			pageSize:    10,
			totalCount:  100,
			expected: &Pagination{
				CurrentPage: 1,
				PageSize:    10,
				TotalCount:  100,
				TotalPages:  10,
				HasPrevious: false,
				HasNext:     true,
				NextPage:    2,
			},
		},
		{
			currentPage: 5,
			pageSize:    10,
			totalCount:  100,
			expected: &Pagination{
				CurrentPage:  5,
				PageSize:     10,
				TotalCount:   100,
				TotalPages:   10,
				HasPrevious:  true,
				HasNext:      true,
				PreviousPage: 4,
				NextPage:     6,
			},
		},
		{
			currentPage: 10,
			pageSize:    10,
			totalCount:  100,
			expected: &Pagination{
				CurrentPage:  10,
				PageSize:     10,
				TotalCount:   100,
				TotalPages:   10,
				HasPrevious:  true,
				HasNext:      false,
				PreviousPage: 9,
			},
		},
		{
			currentPage: 0, // invalid, should be corrected to 1
			pageSize:    0, // invalid, should be corrected to 20
			totalCount:  0,
			expected: &Pagination{
				CurrentPage: 1,
				PageSize:    20,
				TotalCount:  0,
				TotalPages:  1,
				HasPrevious: false,
				HasNext:     false,
			},
		},
	}

	for _, tt := range tests {
		result := NewPagination(tt.currentPage, tt.pageSize, tt.totalCount)
		assert.Equal(t, tt.expected.CurrentPage, result.CurrentPage)
		assert.Equal(t, tt.expected.PageSize, result.PageSize)
		assert.Equal(t, tt.expected.TotalCount, result.TotalCount)
		assert.Equal(t, tt.expected.TotalPages, result.TotalPages)
		assert.Equal(t, tt.expected.HasPrevious, result.HasPrevious)
		assert.Equal(t, tt.expected.HasNext, result.HasNext)
		assert.Equal(t, tt.expected.PreviousPage, result.PreviousPage)
		assert.Equal(t, tt.expected.NextPage, result.NextPage)
	}
}

func TestDefaultPageRequest(t *testing.T) {
	req := DefaultPageRequest()
	assert.Equal(t, 1, req.Page)
	assert.Equal(t, 20, req.PageSize)
	assert.Equal(t, "id", req.SortBy)
	assert.Equal(t, "desc", req.SortDir)
}

func TestParsePageRequest(t *testing.T) {
	// Test with valid query parameters
	req := httptest.NewRequest("GET", "/test?page=2&page_size=50&sort_by=name&sort_dir=asc", nil)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	pageReq := ParsePageRequest(c)
	assert.Equal(t, 2, pageReq.Page)
	assert.Equal(t, 50, pageReq.PageSize)
	assert.Equal(t, "name", pageReq.SortBy)
	assert.Equal(t, "asc", pageReq.SortDir)

	// Test with invalid/missing parameters (should use defaults)
	req2 := httptest.NewRequest("GET", "/test?page=0&page_size=200&sort_dir=invalid", nil)
	c.Request = req2

	pageReq2 := ParsePageRequest(c)
	assert.Equal(t, 1, pageReq2.Page)         // corrected from 0
	assert.Equal(t, 20, pageReq2.PageSize)    // corrected to default 20
	assert.Equal(t, "id", pageReq2.SortBy)    // default
	assert.Equal(t, "desc", pageReq2.SortDir) // corrected from invalid
}

func TestPageRequest_Methods(t *testing.T) {
	req := PageRequest{
		Page:     3,
		PageSize: 25,
		SortBy:   "name",
		SortDir:  "asc",
	}

	assert.Equal(t, 50, req.GetOffset()) // (3-1) * 25
	assert.Equal(t, 25, req.GetLimit())
	assert.Equal(t, "name asc", req.GetOrderBy())

	// Test ValidateSortField
	allowedFields := []string{"id", "name", "email"}
	assert.True(t, req.ValidateSortField(allowedFields))

	req.SortBy = "invalid_field"
	assert.False(t, req.ValidateSortField(allowedFields))
}

func TestCreated(t *testing.T) {
	router, recorder := setupTestGin()

	router.POST("/test", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		Created(c, map[string]string{"id": "123"})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "创建成功", response.Message)
}

func TestUpdated(t *testing.T) {
	router, recorder := setupTestGin()

	router.PUT("/test", func(c *gin.Context) {
		Updated(c, map[string]string{"id": "123"})
	})

	req := httptest.NewRequest("PUT", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "更新成功", response.Message)
}

func TestDeleted(t *testing.T) {
	router, recorder := setupTestGin()

	router.DELETE("/test", func(c *gin.Context) {
		Deleted(c)
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "删除成功", response.Message)
}

func TestNoContent(t *testing.T) {
	router, recorder := setupTestGin()

	router.DELETE("/test", func(c *gin.Context) {
		NoContent(c)
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Empty(t, recorder.Body.String())
}

func TestTooManyRequests(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		TooManyRequests(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeTooManyRequests, response.Code)
}

func TestServiceUnavailable(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/test", func(c *gin.Context) {
		ServiceUnavailable(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeServiceUnavailable, response.Code)
}

func TestFileResponses(t *testing.T) {
	router, recorder := setupTestGin()

	router.GET("/file", func(c *gin.Context) {
		fileResp := &FileResponse{
			FileName:    "test.txt",
			FileSize:    1024,
			ContentType: "text/plain",
			DownloadURL: "/download/test.txt",
		}
		SuccessFile(c, fileResp)
	})

	req := httptest.NewRequest("GET", "/file", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.NotNil(t, response.Data)
}

func TestUploadResponse(t *testing.T) {
	router, recorder := setupTestGin()

	router.POST("/upload", func(c *gin.Context) {
		uploadResp := &UploadResponse{
			FileID:      "file123",
			FileName:    "test.txt",
			FileSize:    1024,
			ContentType: "text/plain",
			URL:         "/files/test.txt",
		}
		SuccessUpload(c, uploadResp)
	})

	req := httptest.NewRequest("POST", "/upload", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "文件上传成功", response.Message)
}

func TestAuthResponse(t *testing.T) {
	router, recorder := setupTestGin()

	router.POST("/login", func(c *gin.Context) {
		authResp := &AuthResponse{
			AccessToken:  "access_token_123",
			RefreshToken: "refresh_token_123",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			User:         map[string]string{"id": "123", "name": "test"},
		}
		SuccessAuth(c, authResp)
	})

	req := httptest.NewRequest("POST", "/login", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response Response
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "登录成功", response.Message)
}

func TestConvertToMap(t *testing.T) {
	// Test Response.ConvertToMap
	response := Response{
		Code:      CodeSuccess,
		Message:   "test message",
		Data:      "test data",
		RequestID: "test-id",
		Timestamp: 1234567890,
	}

	m := response.ConvertToMap()
	assert.Equal(t, CodeSuccess, m["code"])
	assert.Equal(t, "test message", m["message"])
	assert.Equal(t, "test data", m["data"])
	assert.Equal(t, "test-id", m["request_id"])
	assert.Equal(t, int64(1234567890), m["timestamp"])

	// Test ListResponse.ConvertToMap
	pagination := NewPagination(1, 10, 100)
	listResponse := ListResponse{
		Code:       CodeSuccess,
		Message:    "test message",
		Data:       []string{"item1", "item2"},
		Pagination: pagination,
		RequestID:  "test-id",
		Timestamp:  1234567890,
	}

	m2 := listResponse.ConvertToMap()
	assert.Equal(t, CodeSuccess, m2["code"])
	assert.Equal(t, "test message", m2["message"])
	assert.NotNil(t, m2["data"])
	assert.NotNil(t, m2["pagination"])
	assert.Equal(t, "test-id", m2["request_id"])
	assert.Equal(t, int64(1234567890), m2["timestamp"])
}

func TestGetRequestID(t *testing.T) {
	router := gin.New()

	// Test with request ID set
	router.GET("/with-id", func(c *gin.Context) {
		c.Set("request_id", "test-123")
		requestID := getRequestID(c)
		assert.Equal(t, "test-123", requestID)
		c.String(200, "ok")
	})

	// Test without request ID
	router.GET("/without-id", func(c *gin.Context) {
		requestID := getRequestID(c)
		assert.Equal(t, "unknown", requestID)
		c.String(200, "ok")
	})

	// Test with ID requests
	req1 := httptest.NewRequest("GET", "/with-id", nil)
	recorder1 := httptest.NewRecorder()
	router.ServeHTTP(recorder1, req1)

	req2 := httptest.NewRequest("GET", "/without-id", nil)
	recorder2 := httptest.NewRecorder()
	router.ServeHTTP(recorder2, req2)

	_ = recorder1 // 使用变量避免警告
	_ = recorder2 // 使用变量避免警告
}

// Benchmark tests
func BenchmarkSuccess(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		Success(c, map[string]string{"test": "data"})
	})

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}

func BenchmarkNewPagination(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewPagination(5, 20, 1000)
	}
}
