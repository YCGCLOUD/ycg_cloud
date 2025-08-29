package handlers

import (
	"context"

	"github.com/stretchr/testify/mock"

	"cloudpan/internal/repository/models"
	"cloudpan/internal/service/user"
)

// MockUserService 模拟用户服务 - 共享测试mock
type MockUserService struct {
	mock.Mock
}

// 实现 user.UserService 接口的所有方法

// 用户创建和管理
func (m *MockUserService) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

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

// 用户验证和检查
func (m *MockUserService) CheckUserExists(ctx context.Context, email, username string) (bool, error) {
	args := m.Called(ctx, email, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) ValidatePassword(ctx context.Context, userID uint, password string) (bool, error) {
	args := m.Called(ctx, userID, password)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) UpdatePassword(ctx context.Context, userID uint, hashedPassword string) error {
	args := m.Called(ctx, userID, hashedPassword)
	return args.Error(0)
}

// 用户状态管理
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

// 用户查询
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

// 存储配额管理
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

// 用户偏好设置
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
