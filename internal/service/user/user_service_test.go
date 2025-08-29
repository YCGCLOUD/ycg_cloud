package user

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"cloudpan/internal/repository/models"
	userrepo "cloudpan/internal/repository/user"
)

// MockUserRepository 简化的mock用户仓储
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUUID(ctx context.Context, uuid string) (*models.User, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ValidatePassword(ctx context.Context, hashedPassword, plainPassword string) bool {
	args := m.Called(ctx, hashedPassword, plainPassword)
	return args.Bool(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*models.User, int64, error) {
	args := m.Called(ctx, keyword, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) GetActiveUsersCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) UpdateStorageUsed(ctx context.Context, userID uint, size int64) error {
	args := m.Called(ctx, userID, size)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserFileCount(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetUserPreferences(ctx context.Context, userID uint, category string) ([]*models.UserPreference, error) {
	args := m.Called(ctx, userID, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserPreference), args.Error(1)
}

func (m *MockUserRepository) SetUserPreference(ctx context.Context, userID uint, category, key, value string) error {
	args := m.Called(ctx, userID, category, key, value)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUserPreference(ctx context.Context, userID uint, category, key string) error {
	args := m.Called(ctx, userID, category, key)
	return args.Error(0)
}

func (m *MockUserRepository) GetTotalUsersCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetUsersByStatus(ctx context.Context, status string, limit, offset int) ([]*models.User, int64, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

// 测试辅助函数
func createTestUser() *models.User {
	return &models.User{
		UUID:         "test-uuid-123",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Status:       "active",
	}
}

func createTestUserWithID(id uint) *models.User {
	user := createTestUser()
	user.ID = id
	return user
}

// 简化的业务逻辑服务，用于测试核心功能
type SimpleUserService struct {
	repo userrepo.UserRepository
}

func NewSimpleUserService(repo userrepo.UserRepository) *SimpleUserService {
	return &SimpleUserService{repo: repo}
}

func (s *SimpleUserService) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	return s.repo.ExistsByEmail(ctx, email)
}

func (s *SimpleUserService) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	return s.repo.ExistsByUsername(ctx, username)
}

func (s *SimpleUserService) ValidateUserData(ctx context.Context, email, username string) error {
	if email == "" {
		return errors.New("邮箱不能为空")
	}
	if username == "" {
		return errors.New("用户名不能为空")
	}

	exists, err := s.CheckEmailExists(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("邮箱已被注册")
	}

	exists, err = s.CheckUsernameExists(ctx, username)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("用户名已被注册")
	}

	return nil
}

func (s *SimpleUserService) CreateUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("用户数据不能为空")
	}

	if err := s.ValidateUserData(ctx, user.Email, user.Username); err != nil {
		return err
	}

	return s.repo.Create(ctx, user)
}

func (s *SimpleUserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	if id == 0 {
		return nil, errors.New("用户ID不能为空")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *SimpleUserService) GetUserCount(ctx context.Context) (int64, error) {
	return s.repo.GetActiveUsersCount(ctx)
}

// 测试用例
func TestCheckEmailExists(t *testing.T) {
	ctx := context.Background()

	t.Run("邮箱存在", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		mockRepo.On("ExistsByEmail", ctx, "test@example.com").Return(true, nil)

		exists, err := service.CheckEmailExists(ctx, "test@example.com")
		assert.NoError(t, err)
		assert.True(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("邮箱不存在", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		mockRepo.On("ExistsByEmail", ctx, "notfound@example.com").Return(false, nil)

		exists, err := service.CheckEmailExists(ctx, "notfound@example.com")
		assert.NoError(t, err)
		assert.False(t, exists)
		mockRepo.AssertExpectations(t)
	})
}

func TestCheckUsernameExists(t *testing.T) {
	ctx := context.Background()

	t.Run("用户名存在", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		mockRepo.On("ExistsByUsername", ctx, "testuser").Return(true, nil)

		exists, err := service.CheckUsernameExists(ctx, "testuser")
		assert.NoError(t, err)
		assert.True(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("用户名不存在", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		mockRepo.On("ExistsByUsername", ctx, "notfound").Return(false, nil)

		exists, err := service.CheckUsernameExists(ctx, "notfound")
		assert.NoError(t, err)
		assert.False(t, exists)
		mockRepo.AssertExpectations(t)
	})
}

func TestValidateUserData(t *testing.T) {
	ctx := context.Background()

	t.Run("数据验证成功", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		mockRepo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)
		mockRepo.On("ExistsByUsername", ctx, "testuser").Return(false, nil)

		err := service.ValidateUserData(ctx, "test@example.com", "testuser")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("邮箱为空", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		err := service.ValidateUserData(ctx, "", "testuser")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "邮箱不能为空")
	})

	t.Run("用户名为空", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		err := service.ValidateUserData(ctx, "test@example.com", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "用户名不能为空")
	})

	t.Run("邮箱已存在", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		mockRepo.On("ExistsByEmail", ctx, "test@example.com").Return(true, nil)

		err := service.ValidateUserData(ctx, "test@example.com", "testuser")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "邮箱已被注册")
		mockRepo.AssertExpectations(t)
	})

	t.Run("用户名已存在", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		mockRepo.On("ExistsByEmail", ctx, "test@example.com").Return(false, nil)
		mockRepo.On("ExistsByUsername", ctx, "testuser").Return(true, nil)

		err := service.ValidateUserData(ctx, "test@example.com", "testuser")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "用户名已被注册")
		mockRepo.AssertExpectations(t)
	})
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("成功创建用户", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)
		user := createTestUser()

		mockRepo.On("ExistsByEmail", ctx, user.Email).Return(false, nil)
		mockRepo.On("ExistsByUsername", ctx, user.Username).Return(false, nil)
		mockRepo.On("Create", ctx, user).Return(nil)

		err := service.CreateUser(ctx, user)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("用户数据为空", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		err := service.CreateUser(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "用户数据不能为空")
	})

	t.Run("创建失败", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)
		user := createTestUser()

		mockRepo.On("ExistsByEmail", ctx, user.Email).Return(false, nil)
		mockRepo.On("ExistsByUsername", ctx, user.Username).Return(false, nil)
		mockRepo.On("Create", ctx, user).Return(errors.New("数据库错误"))

		err := service.CreateUser(ctx, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "数据库错误")
		mockRepo.AssertExpectations(t)
	})
}

func TestGetUserByID(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取用户", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)
		user := createTestUserWithID(1)

		mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)

		result, err := service.GetUserByID(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, user, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("用户ID为空", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		result, err := service.GetUserByID(ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "用户ID不能为空")
	})

	t.Run("获取用户失败", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		mockRepo.On("GetByID", ctx, uint(1)).Return(nil, errors.New("用户不存在"))

		result, err := service.GetUserByID(ctx, 1)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetUserCount(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取用户数量", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		mockRepo.On("GetActiveUsersCount", ctx).Return(int64(100), nil)

		count, err := service.GetUserCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(100), count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("获取用户数量失败", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewSimpleUserService(mockRepo)

		mockRepo.On("GetActiveUsersCount", ctx).Return(int64(0), errors.New("数据库错误"))

		count, err := service.GetUserCount(ctx)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
		mockRepo.AssertExpectations(t)
	})
}

func TestNewSimpleUserService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewSimpleUserService(mockRepo)
	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
}
