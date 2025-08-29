package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"cloudpan/internal/pkg/email"
	"cloudpan/internal/repository/models"
	"cloudpan/internal/service/user"
)

// Mock对象

// MockUserService 用户服务Mock
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) CheckUserExists(ctx context.Context, email, username string) (bool, error) {
	args := m.Called(ctx, email, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// 实现其他必需的接口方法（简化实现）
func (m *MockUserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) ValidatePassword(ctx context.Context, userID uint, password string) (bool, error) {
	args := m.Called(ctx, userID, password)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) ActivateUser(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserService) DeactivateUser(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserService) SuspendUser(ctx context.Context, userID uint, reason string) error {
	args := m.Called(ctx, userID, reason)
	return args.Error(0)
}

func (m *MockUserService) VerifyEmail(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserService) VerifyPhone(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) SearchUsers(ctx context.Context, keyword string, limit, offset int) ([]*models.User, int64, error) {
	args := m.Called(ctx, keyword, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) GetActiveUsersCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserService) UpdateStorageUsed(ctx context.Context, userID uint, size int64) error {
	args := m.Called(ctx, userID, size)
	return args.Error(0)
}

func (m *MockUserService) CheckStorageQuota(ctx context.Context, userID uint, requiredSize int64) (bool, error) {
	args := m.Called(ctx, userID, requiredSize)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) GetStorageStats(ctx context.Context, userID uint) (*user.UserStorageStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.UserStorageStats), args.Error(1)
}

func (m *MockUserService) GetUserPreferences(ctx context.Context, userID uint, category string) (map[string]interface{}, error) {
	args := m.Called(ctx, userID, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUserService) SetUserPreference(ctx context.Context, userID uint, category, key, value string) error {
	args := m.Called(ctx, userID, category, key, value)
	return args.Error(0)
}

func (m *MockUserService) DeleteUserPreference(ctx context.Context, userID uint, category, key string) error {
	args := m.Called(ctx, userID, category, key)
	return args.Error(0)
}

// MockEmailService 邮件服务Mock
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendVerificationCode(ctx context.Context, to string, code string) error {
	args := m.Called(ctx, to, code)
	return args.Error(0)
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, to string, username string) error {
	args := m.Called(ctx, to, username)
	return args.Error(0)
}

// 实现其他必需的接口方法（简化实现）
func (m *MockEmailService) SendEmail(ctx context.Context, to []string, subject, body string) error {
	args := m.Called(ctx, to, subject, body)
	return args.Error(0)
}

func (m *MockEmailService) SendHTMLEmail(ctx context.Context, to []string, subject, htmlBody, textBody string) error {
	args := m.Called(ctx, to, subject, htmlBody, textBody)
	return args.Error(0)
}

func (m *MockEmailService) SendTemplateEmail(ctx context.Context, templateName string, to []string, variables map[string]interface{}) error {
	args := m.Called(ctx, templateName, to, variables)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordReset(ctx context.Context, to string, resetURL string) error {
	args := m.Called(ctx, to, resetURL)
	return args.Error(0)
}

func (m *MockEmailService) SendSecurityAlert(ctx context.Context, to string, alertType string, details map[string]interface{}) error {
	args := m.Called(ctx, to, alertType, details)
	return args.Error(0)
}

func (m *MockEmailService) QueueEmail(email *email.EmailQueue) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockEmailService) ProcessQueue(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEmailService) GetQueueStatus() (map[string]int, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MockEmailService) LoadTemplates() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEmailService) RegisterTemplate(template *email.EmailTemplate) error {
	args := m.Called(template)
	return args.Error(0)
}

func (m *MockEmailService) GetTemplate(name, language string) (*email.EmailTemplate, error) {
	args := m.Called(name, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*email.EmailTemplate), args.Error(1)
}

func (m *MockEmailService) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEmailService) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEmailService) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}

// MockCacheManager 缓存管理器Mock
type MockCacheManager struct {
	mock.Mock
	data map[string]string
}

func NewMockCacheManager() *MockCacheManager {
	return &MockCacheManager{
		data: make(map[string]string),
	}
}

// SetWithTTL Mock实现，匹配实际接口签名
func (m *MockCacheManager) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	if args.Error(0) == nil {
		// 将value转换为字符串存储
		if str, ok := value.(string); ok {
			m.data[key] = str
		} else {
			m.data[key] = "mock_value"
		}
	}
	return args.Error(0)
}

// Get Mock实现，匹配实际接口签名
func (m *MockCacheManager) Get(key string, dest interface{}) error {
	args := m.Called(key, dest)
	if args.Error(0) == nil {
		// 从内存中获取并设置到dest
		if value, exists := m.data[key]; exists {
			if strPtr, ok := dest.(*string); ok {
				*strPtr = value
			}
		}
	}
	return args.Error(0)
}

// Delete Mock实现，匹配实际接口签名
func (m *MockCacheManager) Delete(keys ...string) error {
	args := m.Called(keys)
	if args.Error(0) == nil {
		for _, key := range keys {
			delete(m.data, key)
		}
	}
	return args.Error(0)
}

func (m *MockCacheManager) Exists(keys ...string) (int64, error) {
	args := m.Called(keys)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheManager) Expire(key string, ttl time.Duration) error {
	args := m.Called(key, ttl)
	return args.Error(0)
}

func (m *MockCacheManager) TTL(key string) (time.Duration, error) {
	args := m.Called(key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockCacheManager) Increment(key string) (int64, error) {
	args := m.Called(key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheManager) Set(key string, value interface{}) error {
	return m.SetWithTTL(key, value, 0)
}

// 测试辅助函数

func setupTestHandler() (*UserRegisterHandler, *MockUserService, *MockEmailService, *MockCacheManager) {
	userService := &MockUserService{}
	emailService := &MockEmailService{}
	cacheManager := NewMockCacheManager()

	handler := NewUserRegisterHandler(userService, emailService, cacheManager)

	return handler, userService, emailService, cacheManager
}

func createTestRequest(method, url string, body interface{}) (*http.Request, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// 测试用例

// TestRegisterHandler_Register 测试用户注册接口
func TestRegisterHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常注册流程", func(t *testing.T) {
		handler, userService, emailService, cacheManager := setupTestHandler()

		// 设置Mock期望
		userService.On("CheckUserExists", mock.Anything, "test@example.com", "testuser").Return(false, nil)
		userService.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
		// 为异步发送欢迎邮件设置Mock期望
		emailService.On("SendWelcomeEmail", mock.Anything, "test@example.com", "testuser").Return(nil)

		// 预设缓存中的验证码
		cacheManager.data["email_code:register:test@example.com"] = "123456"
		cacheManager.On("Get", "email_code:register:test@example.com", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
			if strPtr, ok := args[1].(*string); ok {
				*strPtr = "123456"
			}
		})
		cacheManager.On("Delete", []string{"email_code:register:test@example.com"}).Return(nil)

		// 创建请求
		reqBody := RegisterRequest{
			Email:            "test@example.com",
			Username:         "testuser",
			Password:         "Str0ng@Passw0rd123!",
			ConfirmPassword:  "Str0ng@Passw0rd123!",
			VerificationCode: "123456",
			DisplayName:      "Test User",
			AcceptTerms:      true,
		}

		req, err := createTestRequest("POST", "/register", reqBody)
		assert.NoError(t, err)

		// 创建响应记录器
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// 执行处理器
		handler.Register(c)

		// 验证响应
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "创建成功", response["message"])
		assert.NotNil(t, response["data"])

		// 验证Mock调用
		userService.AssertExpectations(t)
		// 等待一段时间让异步操作完成
		time.Sleep(100 * time.Millisecond)
		emailService.AssertExpectations(t)
	})

	t.Run("密码和确认密码不一致", func(t *testing.T) {
		handler, _, _, _ := setupTestHandler()

		reqBody := RegisterRequest{
			Email:            "test@example.com",
			Username:         "testuser",
			Password:         "Str0ng@Passw0rd123!",
			ConfirmPassword:  "DifferentPassword",
			VerificationCode: "123456",
			AcceptTerms:      true,
		}

		req, err := createTestRequest("POST", "/register", reqBody)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "参数验证失败: 确认密码验证失败: 密码和确认密码不一致", response["message"])
	})

	t.Run("验证码错误", func(t *testing.T) {
		handler, _, _, cacheManager := setupTestHandler()

		// 设置验证码不匹配
		cacheManager.data["email_code:register:test@example.com"] = "654321"
		cacheManager.On("Get", "email_code:register:test@example.com", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
			if strPtr, ok := args[1].(*string); ok {
				*strPtr = "654321"
			}
		})

		reqBody := RegisterRequest{
			Email:            "test@example.com",
			Username:         "testuser",
			Password:         "Str0ng@Passw0rd123!",
			ConfirmPassword:  "Str0ng@Passw0rd123!",
			VerificationCode: "123456", // 错误的验证码
			AcceptTerms:      true,
		}

		req, err := createTestRequest("POST", "/register", reqBody)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Register(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "邮箱验证码错误或已过期: 验证码不正确", response["message"])
	})

	t.Run("用户已存在", func(t *testing.T) {
		handler, userService, _, cacheManager := setupTestHandler()

		// 设置用户已存在
		userService.On("CheckUserExists", mock.Anything, "existing@example.com", "existinguser").Return(true, nil)

		// 预设验证码
		cacheManager.data["email_code:register:existing@example.com"] = "123456"
		cacheManager.On("Get", "email_code:register:existing@example.com", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
			if strPtr, ok := args[1].(*string); ok {
				*strPtr = "123456"
			}
		})

		reqBody := RegisterRequest{
			Email:            "existing@example.com",
			Username:         "existinguser",
			Password:         "Str0ng@Passw0rd123!",
			ConfirmPassword:  "Str0ng@Passw0rd123!",
			VerificationCode: "123456",
			AcceptTerms:      true,
		}

		req, err := createTestRequest("POST", "/register", reqBody)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Register(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "用户已存在: 邮箱或用户名已被注册", response["message"])

		userService.AssertExpectations(t)
	})

	t.Run("无效的邮箱格式", func(t *testing.T) {
		handler, _, _, _ := setupTestHandler()

		reqBody := RegisterRequest{
			Email:            "invalid-email",
			Username:         "testuser",
			Password:         "Test123@456",
			ConfirmPassword:  "Test123@456",
			VerificationCode: "123456",
			AcceptTerms:      true,
		}

		req, err := createTestRequest("POST", "/register", reqBody)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("弱密码", func(t *testing.T) {
		handler, _, _, _ := setupTestHandler()

		reqBody := RegisterRequest{
			Email:            "test@example.com",
			Username:         "testuser",
			Password:         "password123", // 弱密码
			ConfirmPassword:  "password123",
			VerificationCode: "123456",
			AcceptTerms:      true,
		}

		req, err := createTestRequest("POST", "/register", reqBody)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("未接受服务条款", func(t *testing.T) {
		handler, _, _, _ := setupTestHandler()

		reqBody := RegisterRequest{
			Email:            "test@example.com",
			Username:         "testuser",
			Password:         "Str0ng@Passw0rd123!",
			ConfirmPassword:  "Str0ng@Passw0rd123!",
			VerificationCode: "123456",
			AcceptTerms:      false, // 未接受条款
		}

		req, err := createTestRequest("POST", "/register", reqBody)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestRegisterHandler_SendVerificationCode 测试发送验证码接口
func TestRegisterHandler_SendVerificationCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常发送验证码", func(t *testing.T) {
		handler, userService, emailService, cacheManager := setupTestHandler()

		// 设置Mock期望
		userService.On("CheckEmailExists", mock.Anything, "test@example.com").Return(false, nil)
		emailService.On("SendVerificationCode", mock.Anything, "test@example.com", mock.AnythingOfType("string")).Return(nil)
		cacheManager.On("Get", "email_send_limit:register:test@example.com", mock.AnythingOfType("*string")).Return(assert.AnError)
		cacheManager.On("SetWithTTL", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

		reqBody := SendVerificationCodeRequest{
			Email: "test@example.com",
			Type:  "register",
		}

		req, err := createTestRequest("POST", "/send-code", reqBody)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.SendVerificationCode(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "验证码发送成功", response["message"])

		userService.AssertExpectations(t)
		emailService.AssertExpectations(t)
	})

	t.Run("无效的邮箱格式", func(t *testing.T) {
		handler, _, _, _ := setupTestHandler()

		reqBody := SendVerificationCodeRequest{
			Email: "invalid-email",
			Type:  "register",
		}

		req, err := createTestRequest("POST", "/send-code", reqBody)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.SendVerificationCode(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("邮箱已被注册", func(t *testing.T) {
		handler, userService, _, cacheManager := setupTestHandler()

		// 设置邮箱已存在
		userService.On("CheckEmailExists", mock.Anything, "existing@example.com").Return(true, nil)
		cacheManager.On("Get", "email_send_limit:register:existing@example.com", mock.AnythingOfType("*string")).Return(assert.AnError)

		reqBody := SendVerificationCodeRequest{
			Email: "existing@example.com",
			Type:  "register",
		}

		req, err := createTestRequest("POST", "/send-code", reqBody)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.SendVerificationCode(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		userService.AssertExpectations(t)
	})

	t.Run("发送频率限制", func(t *testing.T) {
		handler, _, _, cacheManager := setupTestHandler()

		// 设置频率限制
		cacheManager.On("Get", "email_send_limit:register:test@example.com", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
			if strPtr, ok := args[1].(*string); ok {
				*strPtr = "1234567890"
			}
		})

		reqBody := SendVerificationCodeRequest{
			Email: "test@example.com",
			Type:  "register",
		}

		req, err := createTestRequest("POST", "/send-code", reqBody)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.SendVerificationCode(c)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("无效的验证码类型", func(t *testing.T) {
		handler, emailService, _, cacheManager := setupTestHandler()

		// 添加Mock设置用于checkCodeSendLimit方法
		cacheManager.On("Get", "email_send_limit:invalid_type:test@example.com", mock.AnythingOfType("*string")).Return(assert.AnError)

		// 添加Mock设置用于SendVerificationCode方法
		cacheManager.On("SetWithTTL", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

		// 添加Mock设置用于emailService.SendVerificationCode方法
		emailService.On("SendVerificationCode", mock.Anything, "test@example.com", mock.AnythingOfType("string")).Return(nil)

		reqBody := SendVerificationCodeRequest{
			Email: "test@example.com",
			Type:  "invalid_type",
		}

		req, err := createTestRequest("POST", "/send-code", reqBody)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.SendVerificationCode(c)

		// 应该返回400错误，因为Gin的binding验证会失败
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestRegisterHandler_ValidationFunctions 测试验证函数
func TestRegisterHandler_ValidationFunctions(t *testing.T) {
	handler := &UserRegisterHandler{}

	t.Run("测试isValidEmail", func(t *testing.T) {
		testCases := []struct {
			email string
			valid bool
		}{
			{"test@example.com", true},
			{"user.name@domain.co.uk", true},
			{"invalid-email", false},
			{"@domain.com", false},
			{"user@", false},
			{"", false},
		}

		for _, tc := range testCases {
			result := handler.isValidEmail(tc.email)
			assert.Equal(t, tc.valid, result, "Email: %s", tc.email)
		}
	})

	t.Run("测试validatePasswordStrength", func(t *testing.T) {
		testCases := []struct {
			password string
			valid    bool
		}{
			{"Str0ng@Passw0rd123!", true}, // 强密码
			{"Aa1@", false},               // 太短
			{"password", false},           // 弱密码
			{"PASSWORD123", false},        // 缺少特殊字符
			{"Str0ng@Passw0rd123!", true}, // 中等强度
			{"", false},                   // 空密码
		}

		for _, tc := range testCases {
			err := handler.validatePasswordStrength(tc.password)
			if tc.valid {
				assert.NoError(t, err, "Password: %s", tc.password)
			} else {
				assert.Error(t, err, "Password: %s", tc.password)
			}
		}
	})
}

// TestRegisterHandler_EdgeCases 测试边界情况
func TestRegisterHandler_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("空请求体", func(t *testing.T) {
		handler, _, _, _ := setupTestHandler()

		req, err := http.NewRequest("POST", "/register", nil)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("非法JSON", func(t *testing.T) {
		handler, _, _, _ := setupTestHandler()

		req, err := http.NewRequest("POST", "/register", bytes.NewBufferString("invalid json"))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("最大长度的字段", func(t *testing.T) {
		handler, userService, emailService, cacheManager := setupTestHandler()

		// 预设验证码
		cacheManager.data["email_code:register:very.long.email.address.that.is.still.valid@example.com"] = "123456"
		cacheManager.On("Get", "email_code:register:very.long.email.address.that.is.still.valid@example.com", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
			if strPtr, ok := args[1].(*string); ok {
				*strPtr = "123456"
			}
		})
		cacheManager.On("Delete", []string{"email_code:register:very.long.email.address.that.is.still.valid@example.com"}).Return(nil)

		userService.On("CheckUserExists", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(false, nil)
		userService.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
		emailService.On("SendWelcomeEmail", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

		// 最大长度的用户名（50位）
		longUsername := strings.Repeat("a", 47) + "123" // 50位
		// 最大长度的显示名（100位）
		longDisplayName := strings.Repeat("测试", 50) // 100个中文字符

		reqBody := RegisterRequest{
			Email:            "very.long.email.address.that.is.still.valid@example.com",
			Username:         longUsername,
			Password:         "Str0ng@Passw0rd123!",
			ConfirmPassword:  "Str0ng@Passw0rd123!",
			VerificationCode: "123456",
			DisplayName:      longDisplayName,
			AcceptTerms:      true,
		}

		req, err := createTestRequest("POST", "/register", reqBody)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Register(c)

		// 应该成功
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}
