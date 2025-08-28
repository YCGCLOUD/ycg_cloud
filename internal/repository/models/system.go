package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// RecycleBin 回收站表结构
type RecycleBin struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 回收站项目唯一标识符
	UserID uint   `gorm:"not null;index" json:"user_id"`                  // 用户ID
	FileID uint   `gorm:"not null;index" json:"file_id"`                  // 文件ID

	// 原始信息
	OriginalName     string `gorm:"type:varchar(255);not null" json:"original_name"`  // 原始文件名
	OriginalPath     string `gorm:"type:varchar(2000);not null" json:"original_path"` // 原始路径
	OriginalParentID *uint  `json:"original_parent_id,omitempty"`                     // 原始父文件夹ID

	// 删除信息
	DeletedBy    uint      `gorm:"not null" json:"deleted_by"`                       // 删除者ID
	DeletedAt    time.Time `gorm:"not null" json:"deleted_at"`                       // 删除时间
	DeleteReason *string   `gorm:"type:varchar(255)" json:"delete_reason,omitempty"` // 删除原因

	// 文件信息
	FileSize int64 `gorm:"default:0" json:"file_size"`     // 文件大小
	IsFolder bool  `gorm:"default:false" json:"is_folder"` // 是否为文件夹

	// 恢复信息
	IsRestored  bool       `gorm:"default:false" json:"is_restored"`                 // 是否已恢复
	RestoredBy  *uint      `json:"restored_by,omitempty"`                            // 恢复者ID
	RestoredAt  *time.Time `json:"restored_at,omitempty"`                            // 恢复时间
	RestorePath *string    `gorm:"type:varchar(2000)" json:"restore_path,omitempty"` // 恢复路径

	// 清理信息
	AutoDeleteAt time.Time `gorm:"not null;index" json:"auto_delete_at"`  // 自动删除时间
	IsExpired    bool      `gorm:"default:false;index" json:"is_expired"` // 是否已过期

	// 元数据
	Metadata *basemodels.JSONMap `gorm:"type:json" json:"metadata,omitempty"` // 附加元数据

	// 关联关系
	User     User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	File     File  `gorm:"foreignKey:FileID" json:"file,omitempty"`
	Deleter  User  `gorm:"foreignKey:DeletedBy" json:"deleter,omitempty"`
	Restorer *User `gorm:"foreignKey:RestoredBy" json:"restorer,omitempty"`
}

// TableName 回收站表名
func (RecycleBin) TableName() string {
	return "recycle_bin"
}

// BeforeCreate 创建前钩子
func (r *RecycleBin) BeforeCreate(tx *gorm.DB) error {
	if r.UUID == "" {
		r.UUID = basemodels.GenerateUUID()
	}

	if r.DeletedAt.IsZero() {
		r.DeletedAt = time.Now()
	}

	if r.AutoDeleteAt.IsZero() {
		// 默认30天后自动删除
		r.AutoDeleteAt = r.DeletedAt.Add(30 * 24 * time.Hour)
	}

	return r.BaseModel.BeforeCreate(tx)
}

// IsExpiredForDeletion 检查是否过期需要删除
func (r *RecycleBin) IsExpiredForDeletion() bool {
	return time.Now().After(r.AutoDeleteAt)
}

// CanRestore 检查是否可以恢复
func (r *RecycleBin) CanRestore() bool {
	return !r.IsRestored && !r.IsExpiredForDeletion()
}

// Restore 恢复文件
func (r *RecycleBin) Restore(userID uint, restorePath string) {
	r.IsRestored = true
	r.RestoredBy = &userID
	now := time.Now()
	r.RestoredAt = &now
	r.RestorePath = &restorePath
}

// AuditLog 审计日志表结构
type AuditLog struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 日志唯一标识符
	UserID *uint  `gorm:"index" json:"user_id,omitempty"`                 // 用户ID(系统操作可为空)

	// 操作信息
	Action       string  `gorm:"type:varchar(100);not null;index" json:"action"`        // 操作类型
	Module       string  `gorm:"type:varchar(100);not null;index" json:"module"`        // 模块名称
	ResourceType string  `gorm:"type:varchar(100);not null;index" json:"resource_type"` // 资源类型
	ResourceID   *string `gorm:"type:varchar(100);index" json:"resource_id,omitempty"`  // 资源ID
	ResourceName *string `gorm:"type:varchar(255)" json:"resource_name,omitempty"`      // 资源名称

	// 请求信息
	Method    string  `gorm:"type:varchar(20);not null" json:"method"`           // HTTP方法
	URL       string  `gorm:"type:varchar(2000);not null" json:"url"`            // 请求URL
	UserAgent *string `gorm:"type:varchar(1000)" json:"user_agent,omitempty"`    // 用户代理
	IPAddress string  `gorm:"type:varchar(45);not null;index" json:"ip_address"` // IP地址
	Location  *string `gorm:"type:varchar(200)" json:"location,omitempty"`       // 地理位置

	// 结果信息
	Status       string  `gorm:"type:enum('success','failed','error','warning');default:'success'" json:"status"` // 操作状态
	StatusCode   int     `gorm:"default:200" json:"status_code"`                                                  // HTTP状态码
	ErrorMessage *string `gorm:"type:text" json:"error_message,omitempty"`                                        // 错误信息

	// 数据信息
	RequestData  *basemodels.JSONMap `gorm:"type:json" json:"request_data,omitempty"`  // 请求数据
	ResponseData *basemodels.JSONMap `gorm:"type:json" json:"response_data,omitempty"` // 响应数据
	Changes      *basemodels.JSONMap `gorm:"type:json" json:"changes,omitempty"`       // 数据变更

	// 时间信息
	Duration  int64     `gorm:"default:0" json:"duration"`        // 执行时长(毫秒)
	CreatedAt time.Time `gorm:"not null;index" json:"created_at"` // 创建时间

	// 风险评估
	RiskLevel   string `gorm:"type:enum('low','medium','high','critical');default:'low'" json:"risk_level"` // 风险级别
	IsAnonymous bool   `gorm:"default:false" json:"is_anonymous"`                                           // 是否匿名操作

	// 关联关系
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 审计日志表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate 创建前钩子
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.UUID == "" {
		a.UUID = basemodels.GenerateUUID()
	}

	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}

	return a.BaseModel.BeforeCreate(tx)
}

// IsSuccessful 检查操作是否成功
func (a *AuditLog) IsSuccessful() bool {
	return a.Status == "success"
}

// IsHighRisk 检查是否为高风险操作
func (a *AuditLog) IsHighRisk() bool {
	return a.RiskLevel == "high" || a.RiskLevel == "critical"
}

// SystemSetting 系统设置表结构
type SystemSetting struct {
	basemodels.BaseModel
	// 基本信息
	Category     string  `gorm:"type:varchar(100);not null;index" json:"category"` // 设置分类
	Key          string  `gorm:"type:varchar(100);not null" json:"key"`            // 设置键
	Value        *string `gorm:"type:text" json:"value,omitempty"`                 // 设置值
	DefaultValue *string `gorm:"type:text" json:"default_value,omitempty"`         // 默认值

	// 类型和验证
	ValueType  string              `gorm:"type:enum('string','number','boolean','json','array');default:'string'" json:"value_type"` // 值类型
	Validation *string             `gorm:"type:varchar(500)" json:"validation,omitempty"`                                            // 验证规则
	Options    *basemodels.JSONMap `gorm:"type:json" json:"options,omitempty"`                                                       // 可选值

	// 描述信息
	Name        string  `gorm:"type:varchar(255);not null" json:"name"`   // 设置名称
	Description *string `gorm:"type:text" json:"description,omitempty"`   // 设置描述
	Group       *string `gorm:"type:varchar(100)" json:"group,omitempty"` // 设置分组

	// 权限和可见性
	IsPublic           bool    `gorm:"default:false" json:"is_public"`                         // 是否公开
	IsEditable         bool    `gorm:"default:true" json:"is_editable"`                        // 是否可编辑
	IsSystem           bool    `gorm:"default:false" json:"is_system"`                         // 是否系统设置
	RequiredPermission *string `gorm:"type:varchar(100)" json:"required_permission,omitempty"` // 所需权限

	// 排序和分组
	Sort int `gorm:"default:0" json:"sort"` // 排序权重

	// 更新信息
	UpdatedBy *uint `json:"updated_by,omitempty"` // 更新者ID

	// 关联关系
	Updater *User `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

// TableName 系统设置表名
func (SystemSetting) TableName() string {
	return "system_settings"
}

// BeforeCreate 创建前钩子
func (s *SystemSetting) BeforeCreate(tx *gorm.DB) error {
	// 确保同一分类下的键唯一
	var count int64
	tx.Model(&SystemSetting{}).Where("category = ? AND key = ?",
		s.Category, s.Key).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return s.BaseModel.BeforeCreate(tx)
}

// GetBoolValue 获取布尔值
func (s *SystemSetting) GetBoolValue() bool {
	if s.Value == nil || s.ValueType != "boolean" {
		return false
	}
	return *s.Value == "true" || *s.Value == "1"
}

// GetStringValue 获取字符串值
func (s *SystemSetting) GetStringValue() string {
	if s.Value == nil {
		if s.DefaultValue != nil {
			return *s.DefaultValue
		}
		return ""
	}
	return *s.Value
}

// SetBoolValue 设置布尔值
func (s *SystemSetting) SetBoolValue(value bool) {
	s.ValueType = "boolean"
	if value {
		str := "true"
		s.Value = &str
	} else {
		str := "false"
		s.Value = &str
	}
}

// SetStringValue 设置字符串值
func (s *SystemSetting) SetStringValue(value string) {
	s.ValueType = "string"
	s.Value = &value
}

// CanEdit 检查是否可以编辑
func (s *SystemSetting) CanEdit() bool {
	return s.IsEditable && !s.IsSystem
}

// PasswordResetToken 密码重置令牌表结构
type PasswordResetToken struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 令牌唯一标识符
	UserID uint   `gorm:"not null;index" json:"user_id"`                  // 用户ID
	Email  string `gorm:"type:varchar(255);not null;index" json:"email"`  // 邮箱地址

	// 令牌信息
	Token     string  `gorm:"type:varchar(255);uniqueIndex;not null" json:"token"` // 重置令牌
	TokenHash string  `gorm:"type:varchar(255);not null" json:"-"`                 // 令牌哈希值
	Code      *string `gorm:"type:varchar(20)" json:"code,omitempty"`              // 验证码(6位数字)

	// 状态信息
	IsUsed    bool `gorm:"default:false" json:"is_used"`    // 是否已使用
	IsExpired bool `gorm:"default:false" json:"is_expired"` // 是否已过期

	// 时间信息
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"` // 过期时间
	UsedAt    *time.Time `json:"used_at,omitempty"`                // 使用时间

	// 请求信息
	IPAddress string  `gorm:"type:varchar(45);not null" json:"ip_address"`    // 请求IP地址
	UserAgent *string `gorm:"type:varchar(1000)" json:"user_agent,omitempty"` // 用户代理

	// 尝试次数
	AttemptCount int `gorm:"default:0" json:"attempt_count"` // 尝试次数
	MaxAttempts  int `gorm:"default:5" json:"max_attempts"`  // 最大尝试次数

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 密码重置令牌表名
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// BeforeCreate 创建前钩子
func (p *PasswordResetToken) BeforeCreate(tx *gorm.DB) error {
	if p.UUID == "" {
		p.UUID = basemodels.GenerateUUID()
	}

	if p.Token == "" {
		p.Token = basemodels.GenerateResetToken()
	}

	if p.TokenHash == "" {
		p.TokenHash = basemodels.HashToken(p.Token)
	}

	if p.ExpiresAt.IsZero() {
		p.ExpiresAt = time.Now().Add(1 * time.Hour) // 默认1小时过期
	}

	return p.BaseModel.BeforeCreate(tx)
}

// IsExpiredToken 检查令牌是否过期
func (p *PasswordResetToken) IsExpiredToken() bool {
	return time.Now().After(p.ExpiresAt) || p.IsExpired
}

// IsValid 检查令牌是否有效
func (p *PasswordResetToken) IsValid() bool {
	return !p.IsUsed && !p.IsExpiredToken() && p.AttemptCount < p.MaxAttempts
}

// Use 使用令牌
func (p *PasswordResetToken) Use() {
	p.IsUsed = true
	now := time.Now()
	p.UsedAt = &now
}

// IncrementAttempt 增加尝试次数
func (p *PasswordResetToken) IncrementAttempt() {
	p.AttemptCount++
	if p.AttemptCount >= p.MaxAttempts {
		p.IsExpired = true
	}
}

// 系统设置分类常量
const (
	SettingCategoryGeneral      = "general"      // 通用设置
	SettingCategorySecurity     = "security"     // 安全设置
	SettingCategoryStorage      = "storage"      // 存储设置
	SettingCategoryEmail        = "email"        // 邮件设置
	SettingCategoryNotification = "notification" // 通知设置
	SettingCategoryFile         = "file"         // 文件设置
	SettingCategoryTeam         = "team"         // 团队设置
	SettingCategoryMessage      = "message"      // 消息设置
	SettingCategoryUI           = "ui"           // 界面设置
)

// 常用系统设置键常量
const (
	// 通用设置
	SettingKeySystemName        = "system_name"        // 系统名称
	SettingKeySystemVersion     = "system_version"     // 系统版本
	SettingKeySystemMaintenance = "system_maintenance" // 系统维护模式

	// 安全设置
	SettingKeyPasswordMinLength  = "password_min_length" // 密码最小长度
	SettingKeyPasswordComplexity = "password_complexity" // 密码复杂度要求
	SettingKeySessionTimeout     = "session_timeout"     // 会话超时时间
	SettingKeyMFARequired        = "mfa_required"        // 强制MFA
	SettingKeyLoginAttemptLimit  = "login_attempt_limit" // 登录尝试限制

	// 存储设置
	SettingKeyDefaultStorageQuota = "default_storage_quota" // 默认存储配额
	SettingKeyMaxFileSize         = "max_file_size"         // 最大文件大小
	SettingKeyAllowedFileTypes    = "allowed_file_types"    // 允许的文件类型
	SettingKeyStorageType         = "storage_type"          // 默认存储类型

	// 文件设置
	SettingKeyRecycleBinRetention = "recycle_bin_retention" // 回收站保留天数
	SettingKeyFileVersionLimit    = "file_version_limit"    // 文件版本限制
	SettingKeyThumbnailGeneration = "thumbnail_generation"  // 缩略图生成

	// 团队设置
	SettingKeyMaxTeamMembers      = "max_team_members"      // 最大团队成员数
	SettingKeyTeamCreationAllowed = "team_creation_allowed" // 允许创建团队
)

// 审计日志操作类型常量
const (
	AuditActionLogin    = "login"    // 登录
	AuditActionLogout   = "logout"   // 登出
	AuditActionCreate   = "create"   // 创建
	AuditActionRead     = "read"     // 读取
	AuditActionUpdate   = "update"   // 更新
	AuditActionDelete   = "delete"   // 删除
	AuditActionShare    = "share"    // 分享
	AuditActionDownload = "download" // 下载
	AuditActionUpload   = "upload"   // 上传
	AuditActionInvite   = "invite"   // 邀请
	AuditActionJoin     = "join"     // 加入
	AuditActionLeave    = "leave"    // 离开
	AuditActionConfig   = "config"   // 配置
)

// 审计日志模块常量
const (
	AuditModuleAuth    = "auth"    // 认证模块
	AuditModuleUser    = "user"    // 用户模块
	AuditModuleFile    = "file"    // 文件模块
	AuditModuleTeam    = "team"    // 团队模块
	AuditModuleMessage = "message" // 消息模块
	AuditModuleSystem  = "system"  // 系统模块
)

// 审计日志风险级别常量
const (
	AuditRiskLevelLow      = "low"      // 低风险
	AuditRiskLevelMedium   = "medium"   // 中风险
	AuditRiskLevelHigh     = "high"     // 高风险
	AuditRiskLevelCritical = "critical" // 严重风险
)
