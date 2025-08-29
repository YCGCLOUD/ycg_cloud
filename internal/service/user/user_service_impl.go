package user

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"cloudpan/internal/pkg/cache"
	"cloudpan/internal/repository/models"
	userrepo "cloudpan/internal/repository/user"
)

// userService 用户服务实现
type userService struct {
	userRepo     userrepo.UserRepository
	cacheManager *cache.CacheManager
	db           *gorm.DB
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo userrepo.UserRepository, cacheManager *cache.CacheManager, db *gorm.DB) UserService {
	return &userService{
		userRepo:     userRepo,
		cacheManager: cacheManager,
		db:           db,
	}
}

// CreateUser 创建用户
func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return fmt.Errorf("用户数据不能为空")
	}

	// 检查邮箱是否已存在
	exists, err := s.CheckEmailExists(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("检查邮箱存在性失败: %w", err)
	}
	if exists {
		return fmt.Errorf("邮箱已被注册")
	}

	// 检查用户名是否已存在
	exists, err = s.CheckUsernameExists(ctx, user.Username)
	if err != nil {
		return fmt.Errorf("检查用户名存在性失败: %w", err)
	}
	if exists {
		return fmt.Errorf("用户名已被注册")
	}

	// 保存用户到数据库
	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	// 清除相关缓存
	s.clearUserCache(ctx, user.Email, user.Username, user.UUID)

	return nil
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	if id == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("user:id:%d", id)
	if cachedUser, err := s.getUserFromCache(ctx, cacheKey); err == nil && cachedUser != nil {
		return cachedUser, nil
	}

	// 从数据库获取
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	// 存储到缓存
	s.setUserCache(ctx, cacheKey, user, 10*time.Minute)

	return user, nil
}

// GetUserByUUID 根据UUID获取用户
func (s *userService) GetUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	if uuid == "" {
		return nil, fmt.Errorf("用户UUID不能为空")
	}

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("user:uuid:%s", uuid)
	if cachedUser, err := s.getUserFromCache(ctx, cacheKey); err == nil && cachedUser != nil {
		return cachedUser, nil
	}

	// 从数据库获取
	user, err := s.userRepo.GetByUUID(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	// 存储到缓存
	s.setUserCache(ctx, cacheKey, user, 10*time.Minute)

	return user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("user:email:%s", email)
	if cachedUser, err := s.getUserFromCache(ctx, cacheKey); err == nil && cachedUser != nil {
		return cachedUser, nil
	}

	// 从数据库获取
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	// 存储到缓存
	s.setUserCache(ctx, cacheKey, user, 10*time.Minute)

	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("user:username:%s", username)
	if cachedUser, err := s.getUserFromCache(ctx, cacheKey); err == nil && cachedUser != nil {
		return cachedUser, nil
	}

	// 从数据库获取
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	// 存储到缓存
	s.setUserCache(ctx, cacheKey, user, 10*time.Minute)

	return user, nil
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(ctx context.Context, user *models.User) error {
	if user == nil || user.ID == 0 {
		return fmt.Errorf("用户数据不能为空")
	}

	// 更新数据库
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	// 清除相关缓存
	s.clearUserCache(ctx, user.Email, user.Username, user.UUID)
	if err := s.cacheManager.Delete(fmt.Sprintf("user:id:%d", user.ID)); err != nil {
		// 缓存删除失败，记录错误但不影响主流程
		_ = err // 明确忽略错误
	}

	return nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	if id == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	// 获取用户信息用于清除缓存
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 软删除用户
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	// 清除相关缓存
	s.clearUserCache(ctx, user.Email, user.Username, user.UUID)
	if err := s.cacheManager.Delete(fmt.Sprintf("user:id:%d", id)); err != nil {
		// 缓存删除失败，记录错误但不影响主流程
		_ = err // 明确忽略错误
	}

	return nil
}

// CheckUserExists 检查用户是否存在（邮箱或用户名）
func (s *userService) CheckUserExists(ctx context.Context, email, username string) (bool, error) {
	if email == "" && username == "" {
		return false, fmt.Errorf("邮箱和用户名不能同时为空")
	}

	// 检查邮箱
	if email != "" {
		exists, err := s.CheckEmailExists(ctx, email)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}

	// 检查用户名
	if username != "" {
		exists, err := s.CheckUsernameExists(ctx, username)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}

	return false, nil
}

// CheckEmailExists 检查邮箱是否存在
func (s *userService) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, fmt.Errorf("邮箱不能为空")
	}

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("user_exists:email:%s", email)
	var cached string
	if err := s.cacheManager.Get(cacheKey, &cached); err == nil {
		return cached == "true", nil
	}

	// 从数据库检查
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return false, fmt.Errorf("检查邮箱存在性失败: %w", err)
	}

	// 缓存结果
	existsStr := "false"
	if exists {
		existsStr = "true"
	}
	if err := s.cacheManager.SetWithTTL(cacheKey, existsStr, 5*time.Minute); err != nil {
		// 缓存设置失败，记录错误但不影响主流程
		_ = err // 明确忽略错误
	}

	return exists, nil
}

// CheckUsernameExists 检查用户名是否存在
func (s *userService) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, fmt.Errorf("用户名不能为空")
	}

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("user_exists:username:%s", username)
	var cached string
	if err := s.cacheManager.Get(cacheKey, &cached); err == nil {
		return cached == "true", nil
	}

	// 从数据库检查
	exists, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return false, fmt.Errorf("检查用户名存在性失败: %w", err)
	}

	// 缓存结果
	existsStr := "false"
	if exists {
		existsStr = "true"
	}
	if err := s.cacheManager.SetWithTTL(cacheKey, existsStr, 5*time.Minute); err != nil {
		// 缓存设置失败，记录错误但不影响主流程
		_ = err // 明确忽略错误
	}

	return exists, nil
}

// ValidatePassword 验证用户密码
func (s *userService) ValidatePassword(ctx context.Context, userID uint, password string) (bool, error) {
	if userID == 0 || password == "" {
		return false, fmt.Errorf("用户ID和密码不能为空")
	}

	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("获取用户失败: %w", err)
	}

	return s.userRepo.ValidatePassword(ctx, user.PasswordHash, password), nil
}

// UpdatePassword 更新用户密码
func (s *userService) UpdatePassword(ctx context.Context, userID uint, hashedPassword string) error {
	if userID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}
	if hashedPassword == "" {
		return fmt.Errorf("密码哈希值不能为空")
	}

	// 直接更新数据库中的密码字段
	result := s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password_hash", hashedPassword)
	if result.Error != nil {
		return fmt.Errorf("更新密码失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}

	// 清除用户相关缓存
	user, err := s.GetUserByID(ctx, userID)
	if err == nil {
		s.clearUserCache(ctx, user.Email, user.Username, user.UUID)
		if err := s.cacheManager.Delete(fmt.Sprintf("user:id:%d", userID)); err != nil {
			// 缓存删除失败，记录错误但不影响主流程
			_ = err // 明确忽略错误
		}
	}

	return nil
}

// ActivateUser 激活用户
func (s *userService) ActivateUser(ctx context.Context, userID uint) error {
	return s.updateUserStatus(ctx, userID, "active")
}

// DeactivateUser 停用用户
func (s *userService) DeactivateUser(ctx context.Context, userID uint) error {
	return s.updateUserStatus(ctx, userID, "inactive")
}

// SuspendUser 暂停用户
func (s *userService) SuspendUser(ctx context.Context, userID uint, reason string) error {
	return s.updateUserStatus(ctx, userID, "suspended")
}

// VerifyEmail 验证用户邮箱
func (s *userService) VerifyEmail(ctx context.Context, userID uint) error {
	if userID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}

	now := time.Now()
	user.EmailVerified = true
	user.EmailVerifiedAt = &now

	return s.UpdateUser(ctx, user)
}

// VerifyPhone 验证用户手机
func (s *userService) VerifyPhone(ctx context.Context, userID uint) error {
	if userID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}

	now := time.Now()
	user.PhoneVerified = true
	user.PhoneVerifiedAt = &now

	return s.UpdateUser(ctx, user)
}

// ListUsers 获取用户列表
func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	users, total, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户列表失败: %w", err)
	}

	return users, total, nil
}

// SearchUsers 搜索用户
func (s *userService) SearchUsers(ctx context.Context, keyword string, limit, offset int) ([]*models.User, int64, error) {
	if keyword == "" {
		return s.ListUsers(ctx, limit, offset)
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	users, total, err := s.userRepo.Search(ctx, keyword, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索用户失败: %w", err)
	}

	return users, total, nil
}

// GetActiveUsersCount 获取活跃用户数量
func (s *userService) GetActiveUsersCount(ctx context.Context) (int64, error) {
	// 尝试从缓存获取
	cacheKey := "stats:active_users_count"
	var cached string
	if err := s.cacheManager.Get(cacheKey, &cached); err == nil {
		return parseIntFromString(cached), nil
	}

	// 从数据库获取
	count, err := s.userRepo.GetActiveUsersCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取活跃用户数量失败: %w", err)
	}

	// 缓存结果
	if err := s.cacheManager.SetWithTTL(cacheKey, fmt.Sprintf("%d", count), 1*time.Hour); err != nil {
		// 缓存设置失败，记录错误但不影响主流程
		_ = err // 明确忽略错误
	}

	return count, nil
}

// UpdateStorageUsed 更新用户存储使用量
func (s *userService) UpdateStorageUsed(ctx context.Context, userID uint, size int64) error {
	if userID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	if err := s.userRepo.UpdateStorageUsed(ctx, userID, size); err != nil {
		return fmt.Errorf("更新存储使用量失败: %w", err)
	}

	// 清除相关缓存
	if err := s.cacheManager.Delete(fmt.Sprintf("user:id:%d", userID)); err != nil {
		// 缓存删除失败，记录错误但不影响主流程
		_ = err // 明确忽略错误
	}
	if err := s.cacheManager.Delete(fmt.Sprintf("storage_stats:%d", userID)); err != nil {
		// 缓存删除失败，记录错误但不影响主流程
		_ = err // 明确忽略错误
	}

	return nil
}

// CheckStorageQuota 检查用户存储配额
func (s *userService) CheckStorageQuota(ctx context.Context, userID uint, requiredSize int64) (bool, error) {
	if userID == 0 {
		return false, fmt.Errorf("用户ID不能为空")
	}

	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("获取用户失败: %w", err)
	}

	return user.HasStorageSpace(requiredSize), nil
}

// GetStorageStats 获取用户存储统计
func (s *userService) GetStorageStats(ctx context.Context, userID uint) (*UserStorageStats, error) {
	if userID == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("storage_stats:%d", userID)
	if cached, err := s.getStorageStatsFromCache(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// 获取用户信息
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	// 获取文件数量
	fileCount, err := s.userRepo.GetUserFileCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取文件数量失败: %w", err)
	}

	stats := &UserStorageStats{
		UserID:           user.ID,
		StorageQuota:     user.StorageQuota,
		StorageUsed:      user.StorageUsed,
		StorageAvailable: user.StorageQuota - user.StorageUsed,
		UsagePercent:     user.GetStorageUsagePercent(),
		FileCount:        fileCount,
	}

	// 缓存结果
	s.setStorageStatsCache(ctx, cacheKey, stats, 5*time.Minute)

	return stats, nil
}

// GetUserPreferences 获取用户偏好设置
func (s *userService) GetUserPreferences(ctx context.Context, userID uint, category string) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	preferences, err := s.userRepo.GetUserPreferences(ctx, userID, category)
	if err != nil {
		return nil, fmt.Errorf("获取用户偏好设置失败: %w", err)
	}

	result := make(map[string]interface{})
	for _, pref := range preferences {
		switch pref.ValueType {
		case "boolean":
			result[pref.Key] = pref.GetBoolValue()
		case "json":
			result[pref.Key] = pref.GetStringValue() // JSON 作为字符串返回，前端解析
		default:
			result[pref.Key] = pref.GetStringValue()
		}
	}

	return result, nil
}

// SetUserPreference 设置用户偏好
func (s *userService) SetUserPreference(ctx context.Context, userID uint, category, key, value string) error {
	if userID == 0 || category == "" || key == "" {
		return fmt.Errorf("用户ID、分类和键不能为空")
	}

	return s.userRepo.SetUserPreference(ctx, userID, category, key, value)
}

// DeleteUserPreference 删除用户偏好
func (s *userService) DeleteUserPreference(ctx context.Context, userID uint, category, key string) error {
	if userID == 0 || category == "" || key == "" {
		return fmt.Errorf("用户ID、分类和键不能为空")
	}

	return s.userRepo.DeleteUserPreference(ctx, userID, category, key)
}

// 辅助方法

// updateUserStatus 更新用户状态
func (s *userService) updateUserStatus(ctx context.Context, userID uint, status string) error {
	if userID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}

	user.Status = status
	return s.UpdateUser(ctx, user)
}

// clearUserCache 清除用户相关缓存
func (s *userService) clearUserCache(_ context.Context, email, username, uuid string) {
	if email != "" {
		if err := s.cacheManager.Delete(fmt.Sprintf("user:email:%s", email)); err != nil {
			_ = err // 明确忽略错误
		}
		if err := s.cacheManager.Delete(fmt.Sprintf("user_exists:email:%s", email)); err != nil {
			_ = err // 明确忽略错误
		}
	}
	if username != "" {
		if err := s.cacheManager.Delete(fmt.Sprintf("user:username:%s", username)); err != nil {
			_ = err // 明确忽略错误
		}
		if err := s.cacheManager.Delete(fmt.Sprintf("user_exists:username:%s", username)); err != nil {
			_ = err // 明确忽略错误
		}
	}
	if uuid != "" {
		if err := s.cacheManager.Delete(fmt.Sprintf("user:uuid:%s", uuid)); err != nil {
			_ = err // 明确忽略错误
		}
	}
}

// getUserFromCache 从缓存获取用户
func (s *userService) getUserFromCache(_ context.Context, _ string) (*models.User, error) {
	// 简化实现，实际项目中可以使用 JSON 序列化
	// 这里返回错误表示缓存未命中
	return nil, fmt.Errorf("缓存未命中")
}

// setUserCache 设置用户缓存
func (s *userService) setUserCache(_ context.Context, cacheKey string, user *models.User, _ time.Duration) {
	// 简化实现，实际项目中可以使用 JSON 序列化存储完整用户对象
	// 这里只做占位符实现
}

// getStorageStatsFromCache 从缓存获取存储统计
func (s *userService) getStorageStatsFromCache(_ context.Context, _ string) (*UserStorageStats, error) {
	// 简化实现，实际项目中可以使用 JSON 序列化
	return nil, fmt.Errorf("缓存未命中")
}

// setStorageStatsCache 设置存储统计缓存
func (s *userService) setStorageStatsCache(_ context.Context, _ string, stats *UserStorageStats, _ time.Duration) {
	// 简化实现，实际项目中可以使用 JSON 序列化存储统计对象
}

// parseIntFromString 从字符串解析整数
func parseIntFromString(str string) int64 {
	var result int64
	if _, err := fmt.Sscanf(str, "%d", &result); err != nil {
		// 解析失败，返回默认值
		return 0
	}
	return result
}
