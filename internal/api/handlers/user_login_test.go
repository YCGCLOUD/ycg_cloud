package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"cloudpan/internal/pkg/utils"
	"cloudpan/internal/repository/models"
	"cloudpan/internal/service/user"
)

// 由于user_register_test.go中已经定义了MockUserService，
// 这里我们直接使用简化的mock，避免重复定义

// MockLoginUserService 简化的用户服务Mock（仅用于登录测试）
type MockLoginUserService struct {
	mock.Mock
}

func (m *MockLoginUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockLoginUserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockLoginUserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// 为了满足接口要求，需要实现其他方法（简化实现）
func (m *MockLoginUserService) CreateUser(ctx context.Context, user *models.User) error {
	return nil
}
func (m *MockLoginUserService) GetUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	return nil, nil
}
func (m *MockLoginUserService) UpdateUser(ctx context.Context, user *models.User) error {
	return nil
}
func (m *MockLoginUserService) DeleteUser(ctx context.Context, id uint) error {
	return nil
}
func (m *MockLoginUserService) CheckUserExists(ctx context.Context, email, username string) (bool, error) {
	return false, nil
}
func (m *MockLoginUserService) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	return false, nil
}
func (m *MockLoginUserService) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	return false, nil
}
func (m *MockLoginUserService) ValidatePassword(ctx context.Context, userID uint, password string) (bool, error) {
	return false, nil
}
func (m *MockLoginUserService) UpdatePassword(ctx context.Context, userID uint, hashedPassword string) error {
	return nil
}
func (m *MockLoginUserService) ActivateUser(ctx context.Context, userID uint) error {
	return nil
}
func (m *MockLoginUserService) DeactivateUser(ctx context.Context, userID uint) error {
	return nil
}
func (m *MockLoginUserService) SuspendUser(ctx context.Context, userID uint, reason string) error {
	return nil
}
func (m *MockLoginUserService) VerifyEmail(ctx context.Context, userID uint) error {
	return nil
}
func (m *MockLoginUserService) VerifyPhone(ctx context.Context, userID uint) error {
	return nil
}
func (m *MockLoginUserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	return nil, 0, nil
}
func (m *MockLoginUserService) SearchUsers(ctx context.Context, keyword string, limit, offset int) ([]*models.User, int64, error) {
	return nil, 0, nil
}
func (m *MockLoginUserService) GetActiveUsersCount(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockLoginUserService) UpdateStorageUsed(ctx context.Context, userID uint, size int64) error {
	return nil
}
func (m *MockLoginUserService) CheckStorageQuota(ctx context.Context, userID uint, requiredSize int64) (bool, error) {
	return false, nil
}
func (m *MockLoginUserService) GetStorageStats(ctx context.Context, userID uint) (*user.UserStorageStats, error) {
	return nil, nil
}
func (m *MockLoginUserService) GetUserPreferences(ctx context.Context, userID uint, category string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockLoginUserService) SetUserPreference(ctx context.Context, userID uint, category, key, value string) error {
	return nil
}
func (m *MockLoginUserService) DeleteUserPreference(ctx context.Context, userID uint, category, key string) error {
	return nil
}

// 测试用的JWT密钥
const testJWTSecret = "test-jwt-secret-key-for-unit-testing-very-long-secret"

func setupTestLoginHandler(userService *MockLoginUserService) *UserLoginHandler {
	logger := zap.NewNop()
	handler, _ := NewUserLoginHandler(userService, logger, testJWTSecret)
	return handler
}

func setupTestUser() *models.User {
	passwordHash, _ := utils.HashPassword("testPassword123!")
	displayName := "Test User"
	avatarURL := "https://example.com/avatar.jpg"
	now := time.Now()

	user := &models.User{
		UUID:         "test-uuid-1234",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: passwordHash,
		DisplayName:  &displayName,
		AvatarURL:    &avatarURL,
		Status:       "active",
	}
	// 设置BaseModel字段
	user.ID = 1
	user.CreatedAt = now
	user.UpdatedAt = now
	return user
}

func TestNewUserLoginHandler(t *testing.T) {
	logger := zap.NewNop()
	mockUserService := &MockLoginUserService{}

	t.Run("成功创建登录处理器", func(t *testing.T) {
		handler, err := NewUserLoginHandler(mockUserService, logger, testJWTSecret)
		assert.NoError(t, err)
		assert.NotNil(t, handler)
	})

	t.Run("JWT密钥为空时失败", func(t *testing.T) {
		handler, err := NewUserLoginHandler(mockUserService, logger, "")
		assert.Error(t, err)
		assert.Nil(t, handler)
	})

	t.Run("JWT密钥过短时失败", func(t *testing.T) {
		handler, err := NewUserLoginHandler(mockUserService, logger, "short")
		assert.Error(t, err)
		assert.Nil(t, handler)
	})
}

func TestUserLoginHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("邮箱登录成功", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)
		testUser := setupTestUser()

		// 设置mock期望
		mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(testUser, nil)

		// 创建测试请求
		loginReq := LoginRequest{
			Identifier: "test@example.com",
			Password:   "testPassword123!",
			LoginType:  "email",
		}
		reqBody, _ := json.Marshal(loginReq)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.Login(c)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)

		var response utils.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, utils.CodeSuccess, response.Code)

		// 验证返回的登录响应
		responseData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, responseData["access_token"])
		assert.NotEmpty(t, responseData["refresh_token"])
		assert.Equal(t, "Bearer", responseData["token_type"])

		mockUserService.AssertExpectations(t)
	})

	t.Run("用户名登录成功", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)
		testUser := setupTestUser()

		// 设置mock期望
		mockUserService.On("GetUserByUsername", mock.Anything, "testuser").Return(testUser, nil)

		// 创建测试请求
		loginReq := LoginRequest{
			Identifier: "testuser",
			Password:   "testPassword123!",
			LoginType:  "username",
		}
		reqBody, _ := json.Marshal(loginReq)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.Login(c)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)
		mockUserService.AssertExpectations(t)
	})

	t.Run("自动检测登录类型", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)
		testUser := setupTestUser()

		// 设置mock期望（自动检测为邮箱）
		mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(testUser, nil)

		// 创建测试请求（不指定登录类型）
		loginReq := LoginRequest{
			Identifier: "test@example.com",
			Password:   "testPassword123!",
		}
		reqBody, _ := json.Marshal(loginReq)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.Login(c)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)
		mockUserService.AssertExpectations(t)
	})

	t.Run("无效的请求参数", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer([]byte("invalid json")))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.Login(c)

		// 验证结果
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("缺少登录标识符", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)

		loginReq := LoginRequest{
			Password: "testPassword123!",
		}
		reqBody, _ := json.Marshal(loginReq)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.Login(c)

		// 验证结果
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("用户不存在", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)

		// 设置mock期望（用户不存在）
		mockUserService.On("GetUserByEmail", mock.Anything, "nonexistent@example.com").Return(nil, fmt.Errorf("user not found"))

		loginReq := LoginRequest{
			Identifier: "nonexistent@example.com",
			Password:   "testPassword123!",
		}
		reqBody, _ := json.Marshal(loginReq)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.Login(c)

		// 验证结果
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserService.AssertExpectations(t)
	})

	t.Run("密码错误", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)
		testUser := setupTestUser()

		// 设置mock期望
		mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(testUser, nil)

		loginReq := LoginRequest{
			Identifier: "test@example.com",
			Password:   "wrongPassword",
		}
		reqBody, _ := json.Marshal(loginReq)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.Login(c)

		// 验证结果
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserService.AssertExpectations(t)
	})

	t.Run("用户账户已禁用", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)
		testUser := setupTestUser()
		testUser.Status = "inactive" // 设置为禁用状态

		// 设置mock期望
		mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(testUser, nil)

		loginReq := LoginRequest{
			Identifier: "test@example.com",
			Password:   "testPassword123!",
		}
		reqBody, _ := json.Marshal(loginReq)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.Login(c)

		// 验证结果
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserService.AssertExpectations(t)
	})
}

func TestUserLoginHandler_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("成功刷新令牌", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)
		testUser := setupTestUser()

		// 生成一个有效的刷新令牌
		refreshToken, err := handler.jwtManager.GenerateRefreshToken(
			uint64(testUser.ID), testUser.Username, testUser.Email, "user")
		assert.NoError(t, err)

		// 设置mock期望
		mockUserService.On("GetUserByID", mock.Anything, uint(testUser.ID)).Return(testUser, nil)

		// 创建测试请求
		refreshReq := RefreshTokenRequest{
			RefreshToken: refreshToken,
		}
		reqBody, _ := json.Marshal(refreshReq)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/refresh", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.RefreshToken(c)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code)

		var response utils.Response
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, utils.CodeSuccess, response.Code)

		// 验证返回的刷新响应
		responseData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, responseData["access_token"])
		assert.NotEmpty(t, responseData["refresh_token"])
		assert.NotEqual(t, refreshToken, responseData["refresh_token"]) // 应该是新的刷新令牌

		mockUserService.AssertExpectations(t)
	})

	t.Run("无效的请求参数", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/refresh", bytes.NewBuffer([]byte("invalid json")))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.RefreshToken(c)

		// 验证结果
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("缺少刷新令牌", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)

		refreshReq := RefreshTokenRequest{
			RefreshToken: "",
		}
		reqBody, _ := json.Marshal(refreshReq)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/refresh", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.RefreshToken(c)

		// 验证结果
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("无效的刷新令牌", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)

		refreshReq := RefreshTokenRequest{
			RefreshToken: "invalid-refresh-token",
		}
		reqBody, _ := json.Marshal(refreshReq)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/refresh", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.RefreshToken(c)

		// 验证结果
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("使用访问令牌而非刷新令牌", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)
		testUser := setupTestUser()

		// 生成一个访问令牌（错误类型）
		accessToken, err := handler.jwtManager.GenerateAccessToken(
			uint64(testUser.ID), testUser.Username, testUser.Email, "user")
		assert.NoError(t, err)

		refreshReq := RefreshTokenRequest{
			RefreshToken: accessToken, // 使用访问令牌作为刷新令牌
		}
		reqBody, _ := json.Marshal(refreshReq)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/refresh", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.RefreshToken(c)

		// 验证结果
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("用户不存在", func(t *testing.T) {
		mockUserService := &MockLoginUserService{}
		handler := setupTestLoginHandler(mockUserService)
		testUser := setupTestUser()

		// 生成一个有效的刷新令牌
		refreshToken, err := handler.jwtManager.GenerateRefreshToken(
			uint64(testUser.ID), testUser.Username, testUser.Email, "user")
		assert.NoError(t, err)

		// 设置mock期望（用户不存在）
		mockUserService.On("GetUserByID", mock.Anything, uint(testUser.ID)).Return(nil, fmt.Errorf("user not found"))

		refreshReq := RefreshTokenRequest{
			RefreshToken: refreshToken,
		}
		reqBody, _ := json.Marshal(refreshReq)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/refresh", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 执行测试
		handler.RefreshToken(c)

		// 验证结果
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserService.AssertExpectations(t)
	})
}
