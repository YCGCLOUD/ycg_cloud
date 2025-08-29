package user

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"cloudpan/internal/repository/models"
)

// userRepository 用户数据仓库实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户数据仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if user == nil {
		return fmt.Errorf("用户数据不能为空")
	}

	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	if id == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByUUID 根据UUID获取用户
func (r *userRepository) GetByUUID(ctx context.Context, uuid string) (*models.User, error) {
	if uuid == "" {
		return nil, fmt.Errorf("用户UUID不能为空")
	}

	var user models.User
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}

	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}

	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update 更新用户信息
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	if user == nil || user.ID == 0 {
		return fmt.Errorf("用户数据不能为空")
	}

	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, fmt.Errorf("邮箱不能为空")
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, fmt.Errorf("用户名不能为空")
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByID 检查用户ID是否存在
func (r *userRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	if id == 0 {
		return false, fmt.Errorf("用户ID不能为空")
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ValidatePassword 验证密码
func (r *userRepository) ValidatePassword(ctx context.Context, hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// List 获取用户列表
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Search 搜索用户
func (r *userRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{})

	// 构建搜索条件
	if keyword != "" {
		searchPattern := "%" + keyword + "%"
		query = query.Where("email LIKE ? OR username LIKE ? OR display_name LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetActiveUsersCount 获取活跃用户数量
func (r *userRepository) GetActiveUsersCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("status = ?", "active").
		Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

// UpdateStorageUsed 更新用户存储使用量
func (r *userRepository) UpdateStorageUsed(ctx context.Context, userID uint, size int64) error {
	if userID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("storage_used", gorm.Expr("storage_used + ?", size)).Error
}

// GetUserFileCount 获取用户文件数量
func (r *userRepository) GetUserFileCount(ctx context.Context, userID uint) (int64, error) {
	if userID == 0 {
		return 0, fmt.Errorf("用户ID不能为空")
	}

	var count int64
	// 注意：这里假设有files表，实际实现时需要根据文件模型调整
	err := r.db.WithContext(ctx).Table("files").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error
	if err != nil {
		// 如果files表不存在，返回0而不是错误
		return 0, nil
	}

	return count, nil
}

// GetUserPreferences 获取用户偏好设置
func (r *userRepository) GetUserPreferences(ctx context.Context, userID uint, category string) ([]*models.UserPreference, error) {
	if userID == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	var preferences []*models.UserPreference
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	err := query.Find(&preferences).Error
	if err != nil {
		return nil, err
	}

	return preferences, nil
}

// SetUserPreference 设置用户偏好
func (r *userRepository) SetUserPreference(ctx context.Context, userID uint, category, key, value string) error {
	if userID == 0 || category == "" || key == "" {
		return fmt.Errorf("用户ID、分类和键不能为空")
	}

	preference := &models.UserPreference{
		UserID:    userID,
		Category:  category,
		Key:       key,
		Value:     &value,
		ValueType: "string",
	}

	// 使用 ON DUPLICATE KEY UPDATE 或 UPSERT 逻辑
	return r.db.WithContext(ctx).
		Where("user_id = ? AND category = ? AND key = ?", userID, category, key).
		Assign(models.UserPreference{Value: &value}).
		FirstOrCreate(preference).Error
}

// DeleteUserPreference 删除用户偏好
func (r *userRepository) DeleteUserPreference(ctx context.Context, userID uint, category, key string) error {
	if userID == 0 || category == "" || key == "" {
		return fmt.Errorf("用户ID、分类和键不能为空")
	}

	return r.db.WithContext(ctx).
		Where("user_id = ? AND category = ? AND key = ?", userID, category, key).
		Delete(&models.UserPreference{}).Error
}

// GetTotalUsersCount 获取用户总数
func (r *userRepository) GetTotalUsersCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetUsersByStatus 根据状态获取用户列表
func (r *userRepository) GetUsersByStatus(ctx context.Context, status string, limit, offset int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
