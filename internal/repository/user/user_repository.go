package user

import (
	"context"

	"cloudpan/internal/repository/models"
)

// UserRepository 用户数据仓库接口
//
// 提供用户相关的数据访问操作，包括：
// 1. 用户CRUD操作
// 2. 用户查询和检索
// 3. 用户验证和校验
// 4. 用户偏好设置管理
//
// 使用示例：
//
//	repo := NewUserRepository(db)
//	user, err := repo.GetByEmail(ctx, email)
//	exists, err := repo.ExistsByEmail(ctx, email)
type UserRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint) (*models.User, error)
	GetByUUID(ctx context.Context, uuid string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error

	// 存在性检查
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)

	// 密码验证
	ValidatePassword(ctx context.Context, hashedPassword, plainPassword string) bool

	// 用户查询
	List(ctx context.Context, limit, offset int) ([]*models.User, int64, error)
	Search(ctx context.Context, keyword string, limit, offset int) ([]*models.User, int64, error)
	GetActiveUsersCount(ctx context.Context) (int64, error)

	// 存储管理
	UpdateStorageUsed(ctx context.Context, userID uint, size int64) error
	GetUserFileCount(ctx context.Context, userID uint) (int64, error)

	// 用户偏好设置
	GetUserPreferences(ctx context.Context, userID uint, category string) ([]*models.UserPreference, error)
	SetUserPreference(ctx context.Context, userID uint, category, key, value string) error
	DeleteUserPreference(ctx context.Context, userID uint, category, key string) error

	// 统计信息
	GetTotalUsersCount(ctx context.Context) (int64, error)
	GetUsersByStatus(ctx context.Context, status string, limit, offset int) ([]*models.User, int64, error)
}
