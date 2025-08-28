package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// Role 角色表结构
type Role struct {
	basemodels.BaseModel
	// 基本信息
	UUID        string  `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`     // 角色唯一标识符
	Name        string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"` // 角色名称
	DisplayName string  `gorm:"type:varchar(255);not null" json:"display_name"`     // 显示名称
	Description *string `gorm:"type:text" json:"description,omitempty"`             // 角色描述

	// 类型和分类
	Type     string  `gorm:"type:enum('system','custom','template');default:'custom'" json:"type"` // 角色类型
	Category *string `gorm:"type:varchar(100)" json:"category,omitempty"`                          // 角色分类

	// 状态信息
	IsSystem  bool `gorm:"default:false" json:"is_system"`  // 是否为系统角色
	IsDefault bool `gorm:"default:false" json:"is_default"` // 是否为默认角色
	IsActive  bool `gorm:"default:true" json:"is_active"`   // 是否激活

	// 权限缓存
	PermissionCache *string    `gorm:"type:text" json:"-"`         // 权限缓存(逗号分隔)
	CacheUpdatedAt  *time.Time `json:"cache_updated_at,omitempty"` // 缓存更新时间

	// 排序和分组
	Sort  int `gorm:"default:0" json:"sort"`  // 排序权重
	Level int `gorm:"default:1" json:"level"` // 角色级别

	// 关联关系
	UserRoles       []UserRole       `gorm:"foreignKey:RoleID" json:"user_roles,omitempty"`
	RolePermissions []RolePermission `gorm:"foreignKey:RoleID" json:"role_permissions,omitempty"`
	Permissions     []Permission     `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

// TableName 角色表名
func (Role) TableName() string {
	return "roles"
}

// BeforeCreate 创建前钩子
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.UUID == "" {
		r.UUID = basemodels.GenerateUUID()
	}
	return r.BaseModel.BeforeCreate(tx)
}

// IsSystemRole 检查是否为系统角色
func (r *Role) IsSystemRole() bool {
	return r.IsSystem
}

// CanDelete 检查是否可以删除
func (r *Role) CanDelete() bool {
	return !r.IsSystem && !r.IsDefault
}

// UpdatePermissionCache 更新权限缓存
func (r *Role) UpdatePermissionCache(permissions []string) {
	if len(permissions) == 0 {
		r.PermissionCache = nil
	} else {
		cache := ""
		for i, perm := range permissions {
			if i > 0 {
				cache += ","
			}
			cache += perm
		}
		r.PermissionCache = &cache
	}
	now := time.Now()
	r.CacheUpdatedAt = &now
}

// Permission 权限表结构
type Permission struct {
	basemodels.BaseModel
	// 基本信息
	UUID        string  `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`     // 权限唯一标识符
	Name        string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"` // 权限名称
	DisplayName string  `gorm:"type:varchar(255);not null" json:"display_name"`     // 显示名称
	Description *string `gorm:"type:text" json:"description,omitempty"`             // 权限描述

	// 资源和操作
	ResourceType string  `gorm:"type:varchar(100);not null;index" json:"resource_type"` // 资源类型
	Action       string  `gorm:"type:varchar(100);not null;index" json:"action"`        // 操作类型
	Resource     *string `gorm:"type:varchar(255)" json:"resource,omitempty"`           // 具体资源

	// 分类和分组
	Category string  `gorm:"type:varchar(100);not null;index" json:"category"` // 权限分类
	Group    *string `gorm:"type:varchar(100)" json:"group,omitempty"`         // 权限分组

	// 状态信息
	IsSystem bool `gorm:"default:false" json:"is_system"` // 是否为系统权限
	IsActive bool `gorm:"default:true" json:"is_active"`  // 是否激活

	// 条件和约束
	Conditions  *basemodels.JSONMap `gorm:"type:json" json:"conditions,omitempty"`  // 权限条件
	Constraints *basemodels.JSONMap `gorm:"type:json" json:"constraints,omitempty"` // 权限约束

	// 排序和分组
	Sort  int `gorm:"default:0" json:"sort"`  // 排序权重
	Level int `gorm:"default:1" json:"level"` // 权限级别

	// 关联关系
	RolePermissions []RolePermission `gorm:"foreignKey:PermissionID" json:"role_permissions,omitempty"`
	Roles           []Role           `gorm:"many2many:role_permissions" json:"roles,omitempty"`
}

// TableName 权限表名
func (Permission) TableName() string {
	return "permissions"
}

// BeforeCreate 创建前钩子
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.UUID == "" {
		p.UUID = basemodels.GenerateUUID()
	}
	return p.BaseModel.BeforeCreate(tx)
}

// GetFullName 获取完整权限名称
func (p *Permission) GetFullName() string {
	if p.Resource != nil && *p.Resource != "" {
		return p.ResourceType + ":" + p.Action + ":" + *p.Resource
	}
	return p.ResourceType + ":" + p.Action
}

// IsSystemPermission 检查是否为系统权限
func (p *Permission) IsSystemPermission() bool {
	return p.IsSystem
}

// CanDelete 检查是否可以删除
func (p *Permission) CanDelete() bool {
	return !p.IsSystem
}

// UserRole 用户角色关联表结构
type UserRole struct {
	basemodels.BaseModel
	UserID uint `gorm:"not null;index" json:"user_id"` // 用户ID
	RoleID uint `gorm:"not null;index" json:"role_id"` // 角色ID

	// 授权信息
	GrantedBy uint      `gorm:"not null" json:"granted_by"` // 授权人ID
	GrantedAt time.Time `gorm:"not null" json:"granted_at"` // 授权时间

	// 有效期
	ExpiresAt *time.Time `json:"expires_at,omitempty"`          // 过期时间
	IsActive  bool       `gorm:"default:true" json:"is_active"` // 是否激活

	// 范围限制
	Scope   *string             `gorm:"type:varchar(255)" json:"scope,omitempty"` // 权限范围
	Context *basemodels.JSONMap `gorm:"type:json" json:"context,omitempty"`       // 上下文信息

	// 条件约束
	Conditions *basemodels.JSONMap `gorm:"type:json" json:"conditions,omitempty"` // 附加条件

	// 关联关系
	User    User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role    Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Granter User `gorm:"foreignKey:GrantedBy" json:"granter,omitempty"`
}

// TableName 用户角色表名
func (UserRole) TableName() string {
	return "user_roles"
}

// BeforeCreate 创建前钩子
func (ur *UserRole) BeforeCreate(tx *gorm.DB) error {
	// 确保同一用户的同一角色唯一
	var count int64
	tx.Model(&UserRole{}).Where("user_id = ? AND role_id = ? AND is_active = ?",
		ur.UserID, ur.RoleID, true).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}

	if ur.GrantedAt.IsZero() {
		ur.GrantedAt = time.Now()
	}
	return ur.BaseModel.BeforeCreate(tx)
}

// IsExpired 检查是否过期
func (ur *UserRole) IsExpired() bool {
	if ur.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*ur.ExpiresAt)
}

// IsValid 检查是否有效
func (ur *UserRole) IsValid() bool {
	return ur.IsActive && !ur.IsExpired()
}

// Revoke 撤销角色
func (ur *UserRole) Revoke() {
	ur.IsActive = false
}

// RolePermission 角色权限关联表结构
type RolePermission struct {
	basemodels.BaseModel
	RoleID       uint `gorm:"not null;index" json:"role_id"`       // 角色ID
	PermissionID uint `gorm:"not null;index" json:"permission_id"` // 权限ID

	// 授权信息
	GrantedBy uint      `gorm:"not null" json:"granted_by"` // 授权人ID
	GrantedAt time.Time `gorm:"not null" json:"granted_at"` // 授权时间

	// 状态信息
	IsActive bool `gorm:"default:true" json:"is_active"` // 是否激活

	// 条件约束
	Conditions  *basemodels.JSONMap `gorm:"type:json" json:"conditions,omitempty"`  // 权限条件
	Constraints *basemodels.JSONMap `gorm:"type:json" json:"constraints,omitempty"` // 权限约束

	// 关联关系
	Role       Role       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Permission Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
	Granter    User       `gorm:"foreignKey:GrantedBy" json:"granter,omitempty"`
}

// TableName 角色权限表名
func (RolePermission) TableName() string {
	return "role_permissions"
}

// BeforeCreate 创建前钩子
func (rp *RolePermission) BeforeCreate(tx *gorm.DB) error {
	// 确保同一角色的同一权限唯一
	var count int64
	tx.Model(&RolePermission{}).Where("role_id = ? AND permission_id = ? AND is_active = ?",
		rp.RoleID, rp.PermissionID, true).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}

	if rp.GrantedAt.IsZero() {
		rp.GrantedAt = time.Now()
	}
	return rp.BaseModel.BeforeCreate(tx)
}

// Revoke 撤销权限
func (rp *RolePermission) Revoke() {
	rp.IsActive = false
}

// 角色类型常量
const (
	RoleTypeSystem   = "system"   // 系统角色
	RoleTypeCustom   = "custom"   // 自定义角色
	RoleTypeTemplate = "template" // 模板角色
)

// 系统默认角色常量
const (
	RoleNameSuperAdmin = "super_admin" // 超级管理员
	RoleNameAdmin      = "admin"       // 管理员
	RoleNameUser       = "user"        // 普通用户
	RoleNameGuest      = "guest"       // 访客
)

// 权限分类常量
const (
	PermissionCategoryUser    = "user"    // 用户管理
	PermissionCategoryFile    = "file"    // 文件管理
	PermissionCategoryTeam    = "team"    // 团队管理
	PermissionCategoryMessage = "message" // 消息管理
	PermissionCategorySystem  = "system"  // 系统管理
	PermissionCategoryAudit   = "audit"   // 审计管理
)

// 资源类型常量
const (
	ResourceTypeUser         = "user"         // 用户
	ResourceTypeFile         = "file"         // 文件
	ResourceTypeFolder       = "folder"       // 文件夹
	ResourceTypeTeam         = "team"         // 团队
	ResourceTypeConversation = "conversation" // 会话
	ResourceTypeMessage      = "message"      // 消息
	ResourceTypeSystem       = "system"       // 系统
	ResourceTypeAPI          = "api"          // API
)

// 操作类型常量
const (
	ActionCreate = "create" // 创建
	ActionRead   = "read"   // 读取
	ActionUpdate = "update" // 更新
	ActionDelete = "delete" // 删除
	ActionList   = "list"   // 列表
	ActionManage = "manage" // 管理
	ActionShare  = "share"  // 分享
	ActionInvite = "invite" // 邀请
	ActionJoin   = "join"   // 加入
	ActionLeave  = "leave"  // 离开
)

// 常用权限组合常量
const (
	// 用户权限
	PermissionUserCreate = "user:create"
	PermissionUserRead   = "user:read"
	PermissionUserUpdate = "user:update"
	PermissionUserDelete = "user:delete"
	PermissionUserList   = "user:list"
	PermissionUserManage = "user:manage"

	// 文件权限
	PermissionFileCreate = "file:create"
	PermissionFileRead   = "file:read"
	PermissionFileUpdate = "file:update"
	PermissionFileDelete = "file:delete"
	PermissionFileList   = "file:list"
	PermissionFileShare  = "file:share"
	PermissionFileManage = "file:manage"

	// 团队权限
	PermissionTeamCreate = "team:create"
	PermissionTeamRead   = "team:read"
	PermissionTeamUpdate = "team:update"
	PermissionTeamDelete = "team:delete"
	PermissionTeamList   = "team:list"
	PermissionTeamJoin   = "team:join"
	PermissionTeamLeave  = "team:leave"
	PermissionTeamInvite = "team:invite"
	PermissionTeamManage = "team:manage"

	// 消息权限
	PermissionMessageCreate = "message:create"
	PermissionMessageRead   = "message:read"
	PermissionMessageUpdate = "message:update"
	PermissionMessageDelete = "message:delete"
	PermissionMessageManage = "message:manage"

	// 系统权限
	PermissionSystemConfig = "system:config"
	PermissionSystemAudit  = "system:audit"
	PermissionSystemManage = "system:manage"
)
