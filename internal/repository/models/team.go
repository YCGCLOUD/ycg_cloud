package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// Team 团队表结构
type Team struct {
	basemodels.BaseModel
	// 基本信息
	UUID        string  `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 团队唯一标识符
	Name        string  `gorm:"type:varchar(255);not null" json:"name"`         // 团队名称
	Description *string `gorm:"type:text" json:"description,omitempty"`         // 团队描述
	Avatar      *string `gorm:"type:varchar(500)" json:"avatar,omitempty"`      // 团队头像URL

	// 所有者信息
	OwnerID uint `gorm:"not null;index" json:"owner_id"` // 团队所有者ID

	// 设置信息
	IsPublic     bool `gorm:"default:false" json:"is_public"`    // 是否公开团队
	JoinApproval bool `gorm:"default:true" json:"join_approval"` // 是否需要审批加入
	MaxMembers   int  `gorm:"default:50" json:"max_members"`     // 最大成员数量

	// 存储配额
	StorageQuota int64 `gorm:"default:53687091200" json:"storage_quota"` // 存储配额(50GB)
	StorageUsed  int64 `gorm:"default:0" json:"storage_used"`            // 已使用存储

	// 状态信息
	Status string `gorm:"type:enum('active','inactive','suspended','archived');default:'active'" json:"status"` // 团队状态

	// 统计信息
	MemberCount int   `gorm:"default:1" json:"member_count"` // 成员数量
	FileCount   int64 `gorm:"default:0" json:"file_count"`   // 文件数量

	// 设置参数
	Settings *basemodels.JSONMap `gorm:"type:json" json:"settings,omitempty"` // 团队设置

	// 关联关系
	Owner       User             `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Members     []TeamMember     `gorm:"foreignKey:TeamID" json:"members,omitempty"`
	Files       []TeamFile       `gorm:"foreignKey:TeamID" json:"files,omitempty"`
	Invitations []TeamInvitation `gorm:"foreignKey:TeamID" json:"invitations,omitempty"`
}

// TableName 团队表名
func (Team) TableName() string {
	return "teams"
}

// BeforeCreate 创建前钩子
func (t *Team) BeforeCreate(tx *gorm.DB) error {
	if t.UUID == "" {
		t.UUID = basemodels.GenerateUUID()
	}
	return t.BaseModel.BeforeCreate(tx)
}

// IsActive 检查团队是否活动
func (t *Team) IsActive() bool {
	return t.Status == "active"
}

// HasStorageSpace 检查是否有足够存储空间
func (t *Team) HasStorageSpace(size int64) bool {
	return t.StorageUsed+size <= t.StorageQuota
}

// GetStorageUsagePercent 获取存储使用百分比
func (t *Team) GetStorageUsagePercent() float64 {
	if t.StorageQuota == 0 {
		return 0
	}
	return float64(t.StorageUsed) / float64(t.StorageQuota) * 100
}

// CanAddMember 检查是否可以添加成员
func (t *Team) CanAddMember() bool {
	return t.MemberCount < t.MaxMembers
}

// TeamMember 团队成员表结构
type TeamMember struct {
	basemodels.BaseModel
	TeamID uint `gorm:"not null;index" json:"team_id"` // 团队ID
	UserID uint `gorm:"not null;index" json:"user_id"` // 用户ID

	// 角色和权限
	Role        string              `gorm:"type:enum('owner','admin','member','readonly');default:'member'" json:"role"` // 成员角色
	Permissions *basemodels.JSONMap `gorm:"type:json" json:"permissions,omitempty"`                                      // 自定义权限

	// 状态信息
	Status    string     `gorm:"type:enum('active','inactive','pending','suspended');default:'active'" json:"status"` // 成员状态
	JoinedAt  *time.Time `json:"joined_at,omitempty"`                                                                 // 加入时间
	InvitedBy *uint      `gorm:"index" json:"invited_by,omitempty"`                                                   // 邀请人ID

	// 设置信息
	Nickname        *string `gorm:"type:varchar(100)" json:"nickname,omitempty"` // 团队内昵称
	IsNotifyEnabled bool    `gorm:"default:true" json:"is_notify_enabled"`       // 是否启用通知

	// 统计信息
	FileContributions int64      `gorm:"default:0" json:"file_contributions"` // 文件贡献数
	LastActiveAt      *time.Time `json:"last_active_at,omitempty"`            // 最后活跃时间

	// 关联关系
	Team    Team  `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	User    User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Inviter *User `gorm:"foreignKey:InvitedBy" json:"inviter,omitempty"`
}

// TableName 团队成员表名
func (TeamMember) TableName() string {
	return "team_members"
}

// BeforeCreate 创建前钩子
func (m *TeamMember) BeforeCreate(tx *gorm.DB) error {
	// 确保同一用户在同一团队中唯一
	var count int64
	tx.Model(&TeamMember{}).Where("team_id = ? AND user_id = ?",
		m.TeamID, m.UserID).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}

	if m.JoinedAt == nil {
		now := time.Now()
		m.JoinedAt = &now
	}
	return m.BaseModel.BeforeCreate(tx)
}

// IsActive 检查成员是否活跃
func (m *TeamMember) IsActive() bool {
	return m.Status == "active"
}

// IsOwner 检查是否为团队所有者
func (m *TeamMember) IsOwner() bool {
	return m.Role == "owner"
}

// IsAdmin 检查是否为管理员
func (m *TeamMember) IsAdmin() bool {
	return m.Role == "admin" || m.Role == "owner"
}

// CanManageMembers 检查是否可以管理成员
func (m *TeamMember) CanManageMembers() bool {
	return m.Role == "owner" || m.Role == "admin"
}

// CanManageFiles 检查是否可以管理文件
func (m *TeamMember) CanManageFiles() bool {
	return m.Role != "readonly"
}

// TeamFile 团队文件表结构
type TeamFile struct {
	basemodels.BaseModel
	TeamID   uint `gorm:"not null;index" json:"team_id"`   // 团队ID
	FileID   uint `gorm:"not null;index" json:"file_id"`   // 文件ID
	SharedBy uint `gorm:"not null;index" json:"shared_by"` // 分享者ID

	// 权限设置
	Permission     string `gorm:"type:enum('view','download','edit','manage');default:'view'" json:"permission"` // 权限级别
	IsWritable     bool   `gorm:"default:false" json:"is_writable"`                                              // 是否可写
	IsDownloadable bool   `gorm:"default:true" json:"is_downloadable"`                                           // 是否可下载

	// 状态信息
	Status string `gorm:"type:enum('active','inactive','removed');default:'active'" json:"status"` // 共享状态

	// 时间信息
	SharedAt       time.Time  `gorm:"not null" json:"shared_at"`  // 分享时间
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`       // 过期时间
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"` // 最后访问时间

	// 统计信息
	ViewCount     int64 `gorm:"default:0" json:"view_count"`     // 查看次数
	DownloadCount int64 `gorm:"default:0" json:"download_count"` // 下载次数

	// 元数据
	ShareNote *string             `gorm:"type:text" json:"share_note,omitempty"` // 分享备注
	Metadata  *basemodels.JSONMap `gorm:"type:json" json:"metadata,omitempty"`   // 分享元数据

	// 关联关系
	Team   Team `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	File   File `gorm:"foreignKey:FileID" json:"file,omitempty"`
	Sharer User `gorm:"foreignKey:SharedBy" json:"sharer,omitempty"`
}

// TableName 团队文件表名
func (TeamFile) TableName() string {
	return "team_files"
}

// BeforeCreate 创建前钩子
func (f *TeamFile) BeforeCreate(tx *gorm.DB) error {
	// 确保同一文件在同一团队中唯一
	var count int64
	tx.Model(&TeamFile{}).Where("team_id = ? AND file_id = ?",
		f.TeamID, f.FileID).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}

	if f.SharedAt.IsZero() {
		f.SharedAt = time.Now()
	}
	return f.BaseModel.BeforeCreate(tx)
}

// IsActive 检查共享是否活跃
func (f *TeamFile) IsActive() bool {
	return f.Status == "active"
}

// IsExpired 检查是否过期
func (f *TeamFile) IsExpired() bool {
	if f.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*f.ExpiresAt)
}

// IsAccessible 检查是否可访问
func (f *TeamFile) IsAccessible() bool {
	return f.IsActive() && !f.IsExpired()
}

// TeamInvitation 团队邀请表结构
type TeamInvitation struct {
	basemodels.BaseModel
	TeamID    uint    `gorm:"not null;index" json:"team_id"`                  // 团队ID
	InviterID uint    `gorm:"not null;index" json:"inviter_id"`               // 邀请人ID
	InviteeID *uint   `gorm:"index" json:"invitee_id,omitempty"`              // 被邀请人ID(已注册用户)
	Email     *string `gorm:"type:varchar(255);index" json:"email,omitempty"` // 邀请邮箱(未注册用户)

	// 邀请信息
	InviteCode string `gorm:"type:varchar(100);uniqueIndex;not null" json:"invite_code"`           // 邀请码
	InviteURL  string `gorm:"type:varchar(500);not null" json:"invite_url"`                        // 邀请链接
	Role       string `gorm:"type:enum('admin','member','readonly');default:'member'" json:"role"` // 预设角色

	// 消息信息
	Message *string `gorm:"type:text" json:"message,omitempty"` // 邀请消息

	// 状态信息
	Status string `gorm:"type:enum('pending','accepted','declined','expired','cancelled');default:'pending'" json:"status"` // 邀请状态

	// 时间信息
	InvitedAt   time.Time  `gorm:"not null" json:"invited_at"`       // 邀请时间
	ExpiresAt   time.Time  `gorm:"not null;index" json:"expires_at"` // 过期时间
	RespondedAt *time.Time `json:"responded_at,omitempty"`           // 响应时间

	// 统计信息
	ViewCount int `gorm:"default:0" json:"view_count"` // 查看次数

	// 关联关系
	Team    Team  `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Inviter User  `gorm:"foreignKey:InviterID" json:"inviter,omitempty"`
	Invitee *User `gorm:"foreignKey:InviteeID" json:"invitee,omitempty"`
}

// TableName 团队邀请表名
func (TeamInvitation) TableName() string {
	return "team_invitations"
}

// BeforeCreate 创建前钩子
func (i *TeamInvitation) BeforeCreate(tx *gorm.DB) error {
	if i.InviteCode == "" {
		i.InviteCode = basemodels.GenerateInviteCode()
	}

	if i.InvitedAt.IsZero() {
		i.InvitedAt = time.Now()
	}

	if i.ExpiresAt.IsZero() {
		i.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 默认7天过期
	}

	return i.BaseModel.BeforeCreate(tx)
}

// IsExpired 检查是否过期
func (i *TeamInvitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsPending 检查是否待处理
func (i *TeamInvitation) IsPending() bool {
	return i.Status == "pending" && !i.IsExpired()
}

// CanAccept 检查是否可以接受
func (i *TeamInvitation) CanAccept() bool {
	return i.IsPending()
}

// Accept 接受邀请
func (i *TeamInvitation) Accept() {
	i.Status = "accepted"
	now := time.Now()
	i.RespondedAt = &now
}

// Decline 拒绝邀请
func (i *TeamInvitation) Decline() {
	i.Status = "declined"
	now := time.Now()
	i.RespondedAt = &now
}

// Cancel 取消邀请
func (i *TeamInvitation) Cancel() {
	i.Status = "cancelled"
	now := time.Now()
	i.RespondedAt = &now
}

// 团队状态常量
const (
	TeamStatusActive    = "active"    // 活跃
	TeamStatusInactive  = "inactive"  // 未活跃
	TeamStatusSuspended = "suspended" // 暂停
	TeamStatusArchived  = "archived"  // 归档
)

// 团队成员角色常量
const (
	TeamRoleOwner    = "owner"    // 所有者
	TeamRoleAdmin    = "admin"    // 管理员
	TeamRoleMember   = "member"   // 普通成员
	TeamRoleReadonly = "readonly" // 只读成员
)

// 团队成员状态常量
const (
	TeamMemberStatusActive    = "active"    // 活跃
	TeamMemberStatusInactive  = "inactive"  // 未活跃
	TeamMemberStatusPending   = "pending"   // 待处理
	TeamMemberStatusSuspended = "suspended" // 暂停
)

// 团队文件权限常量
const (
	TeamFilePermissionView     = "view"     // 查看
	TeamFilePermissionDownload = "download" // 下载
	TeamFilePermissionEdit     = "edit"     // 编辑
	TeamFilePermissionManage   = "manage"   // 管理
)

// 团队文件状态常量
const (
	TeamFileStatusActive   = "active"   // 活跃
	TeamFileStatusInactive = "inactive" // 未活跃
	TeamFileStatusRemoved  = "removed"  // 已移除
)

// 邀请状态常量
const (
	InvitationStatusPending   = "pending"   // 待处理
	InvitationStatusAccepted  = "accepted"  // 已接受
	InvitationStatusDeclined  = "declined"  // 已拒绝
	InvitationStatusExpired   = "expired"   // 已过期
	InvitationStatusCancelled = "cancelled" // 已取消
)
