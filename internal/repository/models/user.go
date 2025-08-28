package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// User 用户表结构
type User struct {
	basemodels.BaseModel
	// 基本信息
	UUID         string  `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`         // 用户唯一标识符
	Email        string  `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`    // 邮箱地址
	Username     string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"` // 用户名
	PasswordHash string  `gorm:"type:varchar(255);not null" json:"-"`                    // 密码哈希值
	Phone        *string `gorm:"type:varchar(20);index" json:"phone,omitempty"`          // 手机号码
	AvatarURL    *string `gorm:"type:varchar(500)" json:"avatar_url,omitempty"`          // 头像URL
	DisplayName  *string `gorm:"type:varchar(100)" json:"display_name,omitempty"`        // 显示名称

	// 状态信息
	Status          string     `gorm:"type:enum('active','inactive','suspended','deleted');default:'active';index" json:"status"` // 用户状态
	EmailVerified   bool       `gorm:"default:false" json:"email_verified"`                                                       // 邮箱验证状态
	PhoneVerified   bool       `gorm:"default:false" json:"phone_verified"`                                                       // 手机验证状态
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`                                                               // 邮箱验证时间
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty"`                                                               // 手机验证时间

	// 存储配额
	StorageQuota int64 `gorm:"default:10737418240" json:"storage_quota"` // 存储配额(10GB)
	StorageUsed  int64 `gorm:"default:0" json:"storage_used"`            // 已使用存储

	// 安全信息
	MFAEnabled     bool    `gorm:"default:false" json:"mfa_enabled"`                               // 多因素认证启用状态
	MFASecret      *string `gorm:"type:varchar(255)" json:"-"`                                     // MFA密钥
	MFAType        string  `gorm:"type:enum('totp','sms','email');default:'totp'" json:"mfa_type"` // MFA类型
	MFABackupCodes *string `gorm:"type:text" json:"-"`                                             // MFA备用码

	// 时间信息
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`                         // 最后登录时间
	LastLoginIP       *string    `gorm:"type:varchar(45)" json:"last_login_ip,omitempty"` // 最后登录IP
	PasswordUpdatedAt *time.Time `json:"password_updated_at,omitempty"`                   // 密码最后更新时间

	// JSON字段
	Profile  *basemodels.JSONMap `gorm:"type:json" json:"profile,omitempty"`  // 用户配置信息
	Settings *basemodels.JSONMap `gorm:"type:json" json:"settings,omitempty"` // 用户设置

	// 关联关系
	Sessions     []UserSession      `gorm:"foreignKey:UserID" json:"sessions,omitempty"`
	LoginHistory []UserLoginHistory `gorm:"foreignKey:UserID" json:"login_history,omitempty"`
	Preferences  []UserPreference   `gorm:"foreignKey:UserID" json:"preferences,omitempty"`
	UserRoles    []UserRole         `gorm:"foreignKey:UserID" json:"user_roles,omitempty"`
}

// TableName 用户表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		u.UUID = basemodels.GenerateUUID()
	}
	if u.PasswordUpdatedAt == nil {
		now := time.Now()
		u.PasswordUpdatedAt = &now
	}
	return u.BaseModel.BeforeCreate(tx)
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status == "active"
}

// IsSuspended 检查用户是否被暂停
func (u *User) IsSuspended() bool {
	return u.Status == "suspended"
}

// GetStorageUsagePercent 获取存储使用百分比
func (u *User) GetStorageUsagePercent() float64 {
	if u.StorageQuota == 0 {
		return 0
	}
	return float64(u.StorageUsed) / float64(u.StorageQuota) * 100
}

// HasStorageSpace 检查是否有足够存储空间
func (u *User) HasStorageSpace(size int64) bool {
	return u.StorageUsed+size <= u.StorageQuota
}

// UserSession 用户会话表结构
type UserSession struct {
	basemodels.BaseModel
	UserID         uint       `gorm:"not null;index" json:"user_id"`                               // 用户ID
	SessionToken   string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"session_token"` // 会话令牌
	RefreshToken   *string    `gorm:"type:varchar(255);index" json:"refresh_token,omitempty"`      // 刷新令牌
	DeviceInfo     *string    `gorm:"type:varchar(500)" json:"device_info,omitempty"`              // 设备信息
	UserAgent      *string    `gorm:"type:varchar(1000)" json:"user_agent,omitempty"`              // 用户代理
	IPAddress      *string    `gorm:"type:varchar(45)" json:"ip_address,omitempty"`                // IP地址
	Location       *string    `gorm:"type:varchar(200)" json:"location,omitempty"`                 // 登录位置
	ExpiresAt      time.Time  `gorm:"not null;index" json:"expires_at"`                            // 过期时间
	IsActive       bool       `gorm:"default:true" json:"is_active"`                               // 是否激活
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`                                  // 最后访问时间

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 用户会话表名
func (UserSession) TableName() string {
	return "user_sessions"
}

// IsExpired 检查会话是否过期
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid 检查会话是否有效
func (s *UserSession) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

// UserLoginHistory 用户登录历史表结构
type UserLoginHistory struct {
	basemodels.BaseModel
	UserID      uint    `gorm:"not null;index" json:"user_id"`                                                            // 用户ID
	IPAddress   string  `gorm:"type:varchar(45);not null" json:"ip_address"`                                              // 登录IP地址
	UserAgent   *string `gorm:"type:varchar(1000)" json:"user_agent,omitempty"`                                           // 用户代理
	DeviceInfo  *string `gorm:"type:varchar(500)" json:"device_info,omitempty"`                                           // 设备信息
	Location    *string `gorm:"type:varchar(200)" json:"location,omitempty"`                                              // 登录位置
	LoginMethod string  `gorm:"type:enum('password','email_code','social','mfa');default:'password'" json:"login_method"` // 登录方式
	Status      string  `gorm:"type:enum('success','failed','blocked');default:'success'" json:"status"`                  // 登录状态
	FailReason  *string `gorm:"type:varchar(255)" json:"fail_reason,omitempty"`                                           // 失败原因
	SessionID   *uint   `gorm:"index" json:"session_id,omitempty"`                                                        // 关联会话ID

	// 关联关系
	User    User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Session *UserSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

// TableName 用户登录历史表名
func (UserLoginHistory) TableName() string {
	return "user_login_history"
}

// IsSuccessful 检查是否登录成功
func (h *UserLoginHistory) IsSuccessful() bool {
	return h.Status == "success"
}

// UserPreference 用户偏好设置表结构
type UserPreference struct {
	basemodels.BaseModel
	UserID      uint    `gorm:"not null;index" json:"user_id"`                                                    // 用户ID
	Category    string  `gorm:"type:varchar(100);not null" json:"category"`                                       // 设置分类
	Key         string  `gorm:"type:varchar(100);not null" json:"key"`                                            // 设置键
	Value       *string `gorm:"type:text" json:"value,omitempty"`                                                 // 设置值
	ValueType   string  `gorm:"type:enum('string','number','boolean','json');default:'string'" json:"value_type"` // 值类型
	Description *string `gorm:"type:varchar(255)" json:"description,omitempty"`                                   // 设置描述
	IsPublic    bool    `gorm:"default:false" json:"is_public"`                                                   // 是否公开

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 用户偏好设置表名
func (UserPreference) TableName() string {
	return "user_preferences"
}

// BeforeCreate 创建前钩子
func (p *UserPreference) BeforeCreate(tx *gorm.DB) error {
	// 确保同一用户的同一分类下的键值唯一
	var count int64
	tx.Model(&UserPreference{}).Where("user_id = ? AND category = ? AND key = ?",
		p.UserID, p.Category, p.Key).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return p.BaseModel.BeforeCreate(tx)
}

// GetBoolValue 获取布尔值
func (p *UserPreference) GetBoolValue() bool {
	if p.Value == nil || p.ValueType != "boolean" {
		return false
	}
	return *p.Value == "true"
}

// GetStringValue 获取字符串值
func (p *UserPreference) GetStringValue() string {
	if p.Value == nil {
		return ""
	}
	return *p.Value
}

// SetBoolValue 设置布尔值
func (p *UserPreference) SetBoolValue(value bool) {
	p.ValueType = "boolean"
	if value {
		str := "true"
		p.Value = &str
	} else {
		str := "false"
		p.Value = &str
	}
}

// SetStringValue 设置字符串值
func (p *UserPreference) SetStringValue(value string) {
	p.ValueType = "string"
	p.Value = &value
}

// 常用偏好设置类别常量
const (
	PreferenceCategoryUI       = "ui"       // 界面设置
	PreferenceCategoryFile     = "file"     // 文件设置
	PreferenceCategoryNotify   = "notify"   // 通知设置
	PreferenceCategorySecurity = "security" // 安全设置
	PreferenceCategoryPrivacy  = "privacy"  // 隐私设置
)

// 常用偏好设置键常量
const (
	// UI设置
	PreferenceKeyTheme    = "theme"     // 主题
	PreferenceKeyLanguage = "language"  // 语言
	PreferenceKeyTimezone = "timezone"  // 时区
	PreferenceKeyFileView = "file_view" // 文件视图模式

	// 文件设置
	PreferenceKeyAutoSync      = "auto_sync"      // 自动同步
	PreferenceKeyUploadQuality = "upload_quality" // 上传质量
	PreferenceKeyDownloadPath  = "download_path"  // 下载路径

	// 通知设置
	PreferenceKeyEmailNotify = "email_notify" // 邮件通知
	PreferenceKeyPushNotify  = "push_notify"  // 推送通知
	PreferenceKeySoundNotify = "sound_notify" // 声音通知

	// 安全设置
	PreferenceKeyMFAEnabled = "mfa_enabled" // MFA启用
	PreferenceKeyLoginAlert = "login_alert" // 登录提醒
)
