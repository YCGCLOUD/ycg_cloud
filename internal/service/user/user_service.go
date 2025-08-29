package user

import (
	"context"

	"cloudpan/internal/repository/models"
)

// UserService 用户服务接口
//
// 提供用户相关的业务逻辑操作，包括：
// 1. 用户CRUD操作
// 2. 用户验证和检查
// 3. 用户状态管理
// 4. 用户搜索和查询
//
// 使用示例：
//
//	service := NewUserService(userRepo, cacheManager)
//	user, err := service.CreateUser(ctx, userData)
//	exists, err := service.CheckUserExists(ctx, email, username)
type UserService interface {
	// 用户创建和管理
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uint) error

	// 用户验证和检查
	CheckUserExists(ctx context.Context, email, username string) (bool, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	ValidatePassword(ctx context.Context, userID uint, password string) (bool, error)

	// 用户状态管理
	ActivateUser(ctx context.Context, userID uint) error
	DeactivateUser(ctx context.Context, userID uint) error
	SuspendUser(ctx context.Context, userID uint, reason string) error
	VerifyEmail(ctx context.Context, userID uint) error
	VerifyPhone(ctx context.Context, userID uint) error

	// 用户查询
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, int64, error)
	SearchUsers(ctx context.Context, keyword string, limit, offset int) ([]*models.User, int64, error)
	GetActiveUsersCount(ctx context.Context) (int64, error)

	// 存储配额管理
	UpdateStorageUsed(ctx context.Context, userID uint, size int64) error
	CheckStorageQuota(ctx context.Context, userID uint, requiredSize int64) (bool, error)
	GetStorageStats(ctx context.Context, userID uint) (*UserStorageStats, error)

	// 用户偏好设置
	GetUserPreferences(ctx context.Context, userID uint, category string) (map[string]interface{}, error)
	SetUserPreference(ctx context.Context, userID uint, category, key, value string) error
	DeleteUserPreference(ctx context.Context, userID uint, category, key string) error
}

// UserStorageStats 用户存储统计信息
type UserStorageStats struct {
	UserID           uint    `json:"user_id"`
	StorageQuota     int64   `json:"storage_quota"`     // 存储配额
	StorageUsed      int64   `json:"storage_used"`      // 已使用存储
	StorageAvailable int64   `json:"storage_available"` // 可用存储
	UsagePercent     float64 `json:"usage_percent"`     // 使用百分比
	FileCount        int64   `json:"file_count"`        // 文件数量
}
