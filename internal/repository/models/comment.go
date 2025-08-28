package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// FileComment 文件评论表结构
type FileComment struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 评论唯一标识符
	FileID uint   `gorm:"not null;index" json:"file_id"`                  // 文件ID
	UserID uint   `gorm:"not null;index" json:"user_id"`                  // 评论者ID

	// 评论内容
	Content string `gorm:"type:text;not null" json:"content"` // 评论内容

	// 回复关系
	ParentID *uint `gorm:"index" json:"parent_id,omitempty"` // 父评论ID（用于回复）

	// 提及功能
	MentionedUsers *string             `gorm:"type:text" json:"mentioned_users,omitempty"` // 提及的用户ID(逗号分隔)
	Mentions       *basemodels.JSONMap `gorm:"type:json" json:"mentions,omitempty"`        // 提及信息详情

	// 编辑信息
	IsEdited bool       `gorm:"default:false" json:"is_edited"` // 是否已编辑
	EditedAt *time.Time `json:"edited_at,omitempty"`            // 编辑时间

	// 状态信息
	IsDeleted bool       `gorm:"default:false" json:"is_deleted"` // 是否已删除（软删除）
	DeletedAt *time.Time `json:"deleted_at,omitempty"`            // 删除时间

	// 点赞统计
	LikeCount int `gorm:"default:0" json:"like_count"` // 点赞数

	// 关联关系
	File    File          `gorm:"foreignKey:FileID" json:"file,omitempty"`
	User    User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Parent  *FileComment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies []FileComment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TableName 文件评论表名
func (FileComment) TableName() string {
	return "file_comments"
}

// BeforeCreate 创建前钩子
func (c *FileComment) BeforeCreate(tx *gorm.DB) error {
	if c.UUID == "" {
		c.UUID = basemodels.GenerateUUID()
	}
	return c.BaseModel.BeforeCreate(tx)
}

// IsReply 检查是否为回复评论
func (c *FileComment) IsReply() bool {
	return c.ParentID != nil
}

// CanEdit 检查是否可以编辑（评论者本人或文件所有者）
func (c *FileComment) CanEdit(userID uint, fileOwnerID uint) bool {
	return c.UserID == userID || fileOwnerID == userID
}

// CanDelete 检查是否可以删除（评论者本人或文件所有者）
func (c *FileComment) CanDelete(userID uint, fileOwnerID uint) bool {
	return c.UserID == userID || fileOwnerID == userID
}

// SoftDelete 软删除评论
func (c *FileComment) SoftDelete() {
	c.IsDeleted = true
	now := time.Now()
	c.DeletedAt = &now
}

// Edit 编辑评论
func (c *FileComment) Edit(content string) {
	c.Content = content
	c.IsEdited = true
	now := time.Now()
	c.EditedAt = &now
}

// CommentLike 评论点赞表结构
type CommentLike struct {
	basemodels.BaseModel
	CommentID uint `gorm:"not null;index" json:"comment_id"` // 评论ID
	UserID    uint `gorm:"not null;index" json:"user_id"`    // 用户ID

	// 关联关系
	Comment FileComment `gorm:"foreignKey:CommentID" json:"comment,omitempty"`
	User    User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 评论点赞表名
func (CommentLike) TableName() string {
	return "comment_likes"
}

// BeforeCreate 创建前钩子
func (cl *CommentLike) BeforeCreate(tx *gorm.DB) error {
	// 确保同一用户对同一评论唯一点赞
	var count int64
	tx.Model(&CommentLike{}).Where("comment_id = ? AND user_id = ?",
		cl.CommentID, cl.UserID).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return cl.BaseModel.BeforeCreate(tx)
}
