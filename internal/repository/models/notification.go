package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// Notification 通知表结构
type Notification struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 通知唯一标识符
	UserID uint   `gorm:"not null;index" json:"user_id"`                  // 接收者用户ID

	// 通知内容
	Type    string              `gorm:"type:varchar(50);not null;index" json:"type"` // 通知类型
	Title   string              `gorm:"type:varchar(255);not null" json:"title"`     // 通知标题
	Content string              `gorm:"type:text" json:"content"`                    // 通知内容
	Data    *basemodels.JSONMap `gorm:"type:json" json:"data,omitempty"`             // 通知数据

	// 状态信息
	IsRead bool       `gorm:"default:false;index" json:"is_read"` // 是否已读
	ReadAt *time.Time `json:"read_at,omitempty"`                  // 已读时间

	// 发送信息
	SenderID *uint  `gorm:"index" json:"sender_id,omitempty"`                                           // 发送者ID（系统通知为空）
	Channel  string `gorm:"type:varchar(50);default:'app'" json:"channel"`                              // 通知渠道
	Priority string `gorm:"type:enum('low','normal','high','urgent');default:'normal'" json:"priority"` // 优先级

	// 关联信息
	RelatedType string `gorm:"type:varchar(100)" json:"related_type,omitempty"` // 关联资源类型
	RelatedID   *uint  `json:"related_id,omitempty"`                            // 关联资源ID

	// 过期时间
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // 过期时间

	// 关联关系
	User   User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Sender *User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}

// TableName 通知表名
func (Notification) TableName() string {
	return "notifications"
}

// BeforeCreate 创建前钩子
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.UUID == "" {
		n.UUID = basemodels.GenerateUUID()
	}
	return n.BaseModel.BeforeCreate(tx)
}

// MarkAsRead 标记为已读
func (n *Notification) MarkAsRead() {
	n.IsRead = true
	now := time.Now()
	n.ReadAt = &now
}

// IsExpired 检查是否过期
func (n *Notification) IsExpired() bool {
	if n.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*n.ExpiresAt)
}

// IsHighPriority 检查是否高优先级
func (n *Notification) IsHighPriority() bool {
	return n.Priority == "high" || n.Priority == "urgent"
}

// IsSystemNotification 检查是否为系统通知
func (n *Notification) IsSystemNotification() bool {
	return n.SenderID == nil
}

// 通知类型常量
const (
	NotificationTypeFileShare         = "file_share"         // 文件分享
	NotificationTypeFileComment       = "file_comment"       // 文件评论
	NotificationTypeFileMention       = "file_mention"       // 文件评论@提及
	NotificationTypeTeamInvite        = "team_invite"        // 团队邀请
	NotificationTypeTeamJoin          = "team_join"          // 加入团队
	NotificationTypeTeamFileShare     = "team_file_share"    // 团队文件分享
	NotificationTypeMessageMention    = "message_mention"    // 消息@提及
	NotificationTypeMessageReply      = "message_reply"      // 消息回复
	NotificationTypeStorageWarning    = "storage_warning"    // 存储空间警告
	NotificationTypeSecurityAlert     = "security_alert"     // 安全警告
	NotificationTypeSystemUpdate      = "system_update"      // 系统更新
	NotificationTypePasswordChanged   = "password_changed"   // 密码修改
	NotificationTypeLoginAlert        = "login_alert"        // 登录警告
	NotificationTypeFileUpload        = "file_upload"        // 文件上传完成
	NotificationTypeFileVersion       = "file_version"       // 文件版本更新
	NotificationTypeTeamMemberJoin    = "team_member_join"   // 团队成员加入
	NotificationTypeTeamMemberLeave   = "team_member_leave"  // 团队成员离开
	NotificationTypeSystemMaintenance = "system_maintenance" // 系统维护
)

// 通知渠道常量
const (
	NotificationChannelApp   = "app"   // 应用内通知
	NotificationChannelEmail = "email" // 邮件通知
	NotificationChannelSMS   = "sms"   // 短信通知
	NotificationChannelPush  = "push"  // 推送通知
)

// 通知优先级常量
const (
	NotificationPriorityLow    = "low"    // 低优先级
	NotificationPriorityNormal = "normal" // 普通优先级
	NotificationPriorityHigh   = "high"   // 高优先级
	NotificationPriorityUrgent = "urgent" // 紧急优先级
)

// 关联资源类型常量
const (
	NotificationRelatedTypeFile         = "file"         // 文件
	NotificationRelatedTypeFolder       = "folder"       // 文件夹
	NotificationRelatedTypeTeam         = "team"         // 团队
	NotificationRelatedTypeComment      = "comment"      // 评论
	NotificationRelatedTypeMessage      = "message"      // 消息
	NotificationRelatedTypeConversation = "conversation" // 会话
	NotificationRelatedTypeUser         = "user"         // 用户
)
