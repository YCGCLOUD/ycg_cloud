package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// Conversation 会话表结构
type Conversation struct {
	basemodels.BaseModel
	// 基本信息
	UUID        string  `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`                             // 会话唯一标识符
	Type        string  `gorm:"type:enum('private','group','team','system');default:'private'" json:"type"` // 会话类型
	Name        *string `gorm:"type:varchar(255)" json:"name,omitempty"`                                    // 会话名称(群聊)
	Description *string `gorm:"type:text" json:"description,omitempty"`                                     // 会话描述
	Avatar      *string `gorm:"type:varchar(500)" json:"avatar,omitempty"`                                  // 会话头像URL

	// 创建者信息
	CreatorID uint `gorm:"not null;index" json:"creator_id"` // 创建者ID

	// 状态信息
	Status      string `gorm:"type:enum('active','archived','deleted','muted');default:'active'" json:"status"` // 会话状态
	IsEncrypted bool   `gorm:"default:false" json:"is_encrypted"`                                               // 是否加密

	// 设置信息
	MaxMembers   *int `json:"max_members,omitempty"`              // 最大成员数(群聊)
	JoinApproval bool `gorm:"default:false" json:"join_approval"` // 是否需要审批加入
	AllowInvite  bool `gorm:"default:true" json:"allow_invite"`   // 是否允许邀请成员

	// 统计信息
	MemberCount  int   `gorm:"default:0" json:"member_count"`  // 成员数量
	MessageCount int64 `gorm:"default:0" json:"message_count"` // 消息数量

	// 时间信息
	LastMessageAt *time.Time `json:"last_message_at,omitempty"` // 最后消息时间
	LastActiveAt  *time.Time `json:"last_active_at,omitempty"`  // 最后活跃时间

	// 设置参数
	Settings *basemodels.JSONMap `gorm:"type:json" json:"settings,omitempty"` // 会话设置

	// 关联信息
	TeamID *uint `gorm:"index" json:"team_id,omitempty"` // 关联团队ID(团队会话)

	// 关联关系
	Creator  User                 `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Team     *Team                `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Members  []ConversationMember `gorm:"foreignKey:ConversationID" json:"members,omitempty"`
	Messages []Message            `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

// TableName 会话表名
func (Conversation) TableName() string {
	return "conversations"
}

// BeforeCreate 创建前钩子
func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.UUID == "" {
		c.UUID = basemodels.GenerateUUID()
	}
	return c.BaseModel.BeforeCreate(tx)
}

// IsActive 检查会话是否活跃
func (c *Conversation) IsActive() bool {
	return c.Status == "active"
}

// IsPrivate 检查是否为私聊
func (c *Conversation) IsPrivate() bool {
	return c.Type == "private"
}

// IsGroup 检查是否为群聊
func (c *Conversation) IsGroup() bool {
	return c.Type == "group"
}

// CanAddMember 检查是否可以添加成员
func (c *Conversation) CanAddMember() bool {
	if c.MaxMembers == nil {
		return true
	}
	return c.MemberCount < *c.MaxMembers
}

// Message 消息表结构
type Message struct {
	basemodels.BaseModel
	// 基本信息
	UUID           string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 消息唯一标识符
	ConversationID uint   `gorm:"not null;index" json:"conversation_id"`          // 会话ID
	SenderID       uint   `gorm:"not null;index" json:"sender_id"`                // 发送者ID

	// 消息内容
	Type       string  `gorm:"type:enum('text','image','video','audio','file','system','recall');default:'text'" json:"type"` // 消息类型
	Content    *string `gorm:"type:text" json:"content,omitempty"`                                                            // 消息内容
	RawContent *string `gorm:"type:text" json:"raw_content,omitempty"`                                                        // 原始内容(加密前)

	// 回复信息
	ReplyToID *uint `gorm:"index" json:"reply_to_id,omitempty"` // 回复的消息ID

	// 附件信息
	Attachments *basemodels.JSONMap `gorm:"type:json" json:"attachments,omitempty"` // 附件信息
	FileID      *uint               `gorm:"index" json:"file_id,omitempty"`         // 关联文件ID(文件消息)

	// 消息属性
	IsEncrypted bool `gorm:"default:false" json:"is_encrypted"` // 是否加密
	IsPinned    bool `gorm:"default:false" json:"is_pinned"`    // 是否置顶
	IsEdited    bool `gorm:"default:false" json:"is_edited"`    // 是否已编辑
	IsRecalled  bool `gorm:"default:false" json:"is_recalled"`  // 是否已撤回

	// 时间信息
	EditedAt   *time.Time `json:"edited_at,omitempty"`   // 编辑时间
	RecalledAt *time.Time `json:"recalled_at,omitempty"` // 撤回时间

	// 统计信息
	ReadCount int `gorm:"default:0" json:"read_count"` // 已读人数

	// 提及信息
	MentionedUsers *string `gorm:"type:text" json:"mentioned_users,omitempty"` // 提及的用户ID(逗号分隔)
	MentionAll     bool    `gorm:"default:false" json:"mention_all"`           // 是否@全体成员

	// 元数据
	Metadata *basemodels.JSONMap `gorm:"type:json" json:"metadata,omitempty"` // 消息元数据

	// 关联关系
	Conversation Conversation        `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
	Sender       User                `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ReplyTo      *Message            `gorm:"foreignKey:ReplyToID" json:"reply_to,omitempty"`
	File         *File               `gorm:"foreignKey:FileID" json:"file,omitempty"`
	ReadStatus   []MessageReadStatus `gorm:"foreignKey:MessageID" json:"read_status,omitempty"`
}

// TableName 消息表名
func (Message) TableName() string {
	return "messages"
}

// BeforeCreate 创建前钩子
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.UUID == "" {
		m.UUID = basemodels.GenerateUUID()
	}
	return m.BaseModel.BeforeCreate(tx)
}

// IsText 检查是否为文本消息
func (m *Message) IsText() bool {
	return m.Type == "text"
}

// IsFile 检查是否为文件消息
func (m *Message) IsFile() bool {
	return m.Type == "file"
}

// IsImage 检查是否为图片消息
func (m *Message) IsImage() bool {
	return m.Type == "image"
}

// IsSystem 检查是否为系统消息
func (m *Message) IsSystem() bool {
	return m.Type == "system"
}

// CanRecall 检查是否可以撤回(5分钟内)
func (m *Message) CanRecall() bool {
	if m.IsRecalled {
		return false
	}
	return time.Since(m.CreatedAt) <= 5*time.Minute
}

// Recall 撤回消息
func (m *Message) Recall() {
	m.IsRecalled = true
	now := time.Now()
	m.RecalledAt = &now
}

// ConversationMember 会话成员表结构
type ConversationMember struct {
	basemodels.BaseModel
	ConversationID uint `gorm:"not null;index" json:"conversation_id"` // 会话ID
	UserID         uint `gorm:"not null;index" json:"user_id"`         // 用户ID

	// 角色和权限
	Role string `gorm:"type:enum('owner','admin','member');default:'member'" json:"role"` // 成员角色

	// 状态信息
	Status   string     `gorm:"type:enum('active','muted','left','kicked','banned');default:'active'" json:"status"` // 成员状态
	JoinedAt time.Time  `gorm:"not null" json:"joined_at"`                                                           // 加入时间
	LeftAt   *time.Time `json:"left_at,omitempty"`                                                                   // 离开时间

	// 消息状态
	LastReadAt    *time.Time `json:"last_read_at,omitempty"`                 // 最后已读时间
	LastMessageID *uint      `gorm:"index" json:"last_message_id,omitempty"` // 最后已读消息ID
	UnreadCount   int64      `gorm:"default:0" json:"unread_count"`          // 未读消息数

	// 设置信息
	Nickname        *string `gorm:"type:varchar(100)" json:"nickname,omitempty"` // 会话内昵称
	IsMuted         bool    `gorm:"default:false" json:"is_muted"`               // 是否静音
	IsNotifyEnabled bool    `gorm:"default:true" json:"is_notify_enabled"`       // 是否启用通知
	IsPinned        bool    `gorm:"default:false" json:"is_pinned"`              // 是否置顶会话

	// 操作者信息
	InvitedBy *uint `gorm:"index" json:"invited_by,omitempty"` // 邀请人ID
	RemovedBy *uint `gorm:"index" json:"removed_by,omitempty"` // 移除人ID

	// 关联关系
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Inviter      *User        `gorm:"foreignKey:InvitedBy" json:"inviter,omitempty"`
	Remover      *User        `gorm:"foreignKey:RemovedBy" json:"remover,omitempty"`
	LastMessage  *Message     `gorm:"foreignKey:LastMessageID" json:"last_message,omitempty"`
}

// TableName 会话成员表名
func (ConversationMember) TableName() string {
	return "conversation_members"
}

// BeforeCreate 创建前钩子
func (m *ConversationMember) BeforeCreate(tx *gorm.DB) error {
	// 确保同一用户在同一会话中唯一
	var count int64
	tx.Model(&ConversationMember{}).Where("conversation_id = ? AND user_id = ?",
		m.ConversationID, m.UserID).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}

	if m.JoinedAt.IsZero() {
		m.JoinedAt = time.Now()
	}
	return m.BaseModel.BeforeCreate(tx)
}

// IsActive 检查成员是否活跃
func (m *ConversationMember) IsActive() bool {
	return m.Status == "active"
}

// IsOwner 检查是否为所有者
func (m *ConversationMember) IsOwner() bool {
	return m.Role == "owner"
}

// IsAdmin 检查是否为管理员
func (m *ConversationMember) IsAdmin() bool {
	return m.Role == "admin" || m.Role == "owner"
}

// CanManageMembers 检查是否可以管理成员
func (m *ConversationMember) CanManageMembers() bool {
	return m.IsAdmin()
}

// Leave 离开会话
func (m *ConversationMember) Leave() {
	m.Status = "left"
	now := time.Now()
	m.LeftAt = &now
}

// MessageReadStatus 消息已读状态表结构
type MessageReadStatus struct {
	basemodels.BaseModel
	MessageID      uint `gorm:"not null;index" json:"message_id"`      // 消息ID
	UserID         uint `gorm:"not null;index" json:"user_id"`         // 用户ID
	ConversationID uint `gorm:"not null;index" json:"conversation_id"` // 会话ID

	// 状态信息
	IsRead bool       `gorm:"default:false" json:"is_read"` // 是否已读
	ReadAt *time.Time `json:"read_at,omitempty"`            // 已读时间

	// 关联关系
	Message      Message      `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
}

// TableName 消息已读状态表名
func (MessageReadStatus) TableName() string {
	return "message_read_status"
}

// BeforeCreate 创建前钩子
func (s *MessageReadStatus) BeforeCreate(tx *gorm.DB) error {
	// 确保同一用户对同一消息的状态唯一
	var count int64
	tx.Model(&MessageReadStatus{}).Where("message_id = ? AND user_id = ?",
		s.MessageID, s.UserID).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return s.BaseModel.BeforeCreate(tx)
}

// MarkAsRead 标记为已读
func (s *MessageReadStatus) MarkAsRead() {
	s.IsRead = true
	now := time.Now()
	s.ReadAt = &now
}

// 会话类型常量
const (
	ConversationTypePrivate = "private" // 私聊
	ConversationTypeGroup   = "group"   // 群聊
	ConversationTypeTeam    = "team"    // 团队会话
	ConversationTypeSystem  = "system"  // 系统会话
)

// 会话状态常量
const (
	ConversationStatusActive   = "active"   // 活跃
	ConversationStatusArchived = "archived" // 归档
	ConversationStatusDeleted  = "deleted"  // 已删除
	ConversationStatusMuted    = "muted"    // 静音
)

// 消息类型常量
const (
	MessageTypeText   = "text"   // 文本消息
	MessageTypeImage  = "image"  // 图片消息
	MessageTypeVideo  = "video"  // 视频消息
	MessageTypeAudio  = "audio"  // 音频消息
	MessageTypeFile   = "file"   // 文件消息
	MessageTypeSystem = "system" // 系统消息
	MessageTypeRecall = "recall" // 撤回消息
)

// 会话成员角色常量
const (
	ConversationMemberRoleOwner  = "owner"  // 所有者
	ConversationMemberRoleAdmin  = "admin"  // 管理员
	ConversationMemberRoleMember = "member" // 普通成员
)

// 会话成员状态常量
const (
	ConversationMemberStatusActive = "active" // 活跃
	ConversationMemberStatusMuted  = "muted"  // 静音
	ConversationMemberStatusLeft   = "left"   // 已离开
	ConversationMemberStatusKicked = "kicked" // 被踢出
	ConversationMemberStatusBanned = "banned" // 被禁言
)
