package models

import (
	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// Tag 标签表结构
type Tag struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 标签唯一标识符
	UserID uint   `gorm:"not null;index" json:"user_id"`                  // 用户ID
	Name   string `gorm:"type:varchar(100);not null" json:"name"`         // 标签名称

	// 外观设置
	Color       string  `gorm:"type:varchar(20);default:'#1890ff'" json:"color"` // 标签颜色
	Icon        *string `gorm:"type:varchar(50)" json:"icon,omitempty"`          // 标签图标
	Description *string `gorm:"type:varchar(255)" json:"description,omitempty"`  // 标签描述

	// 统计信息
	FileCount  int `gorm:"default:0" json:"file_count"`  // 关联文件数量
	UsageCount int `gorm:"default:0" json:"usage_count"` // 使用次数

	// 系统标签
	IsSystem bool `gorm:"default:false" json:"is_system"` // 是否为系统标签

	// 分类信息
	Category *string `gorm:"type:varchar(100)" json:"category,omitempty"` // 标签分类

	// 关联关系
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	FileTags []FileTag `gorm:"foreignKey:TagID" json:"file_tags,omitempty"`
}

// TableName 标签表名
func (Tag) TableName() string {
	return "tags"
}

// BeforeCreate 创建前钩子
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	if t.UUID == "" {
		t.UUID = basemodels.GenerateUUID()
	}

	// 确保同一用户的标签名称唯一
	var count int64
	tx.Model(&Tag{}).Where("user_id = ? AND name = ?", t.UserID, t.Name).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}

	return t.BaseModel.BeforeCreate(tx)
}

// IncrementUsage 增加使用次数
func (t *Tag) IncrementUsage() {
	t.UsageCount++
}

// IncrementFileCount 增加文件关联数
func (t *Tag) IncrementFileCount() {
	t.FileCount++
}

// DecrementFileCount 减少文件关联数
func (t *Tag) DecrementFileCount() {
	if t.FileCount > 0 {
		t.FileCount--
	}
}

// FileTagV2 文件标签关联表结构（新版本）
type FileTagV2 struct {
	basemodels.BaseModel
	FileID uint `gorm:"not null;index" json:"file_id"` // 文件ID
	TagID  uint `gorm:"not null;index" json:"tag_id"`  // 标签ID
	UserID uint `gorm:"not null;index" json:"user_id"` // 用户ID

	// 关联关系
	File File `gorm:"foreignKey:FileID" json:"file,omitempty"`
	Tag  Tag  `gorm:"foreignKey:TagID" json:"tag,omitempty"`
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 文件标签关联表名
func (FileTagV2) TableName() string {
	return "file_tag_relations"
}

// BeforeCreate 创建前钩子
func (ft *FileTagV2) BeforeCreate(tx *gorm.DB) error {
	// 确保同一文件的同一标签唯一
	var count int64
	tx.Model(&FileTagV2{}).Where("file_id = ? AND tag_id = ?", ft.FileID, ft.TagID).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return ft.BaseModel.BeforeCreate(tx)
}

// 系统标签常量
const (
	SystemTagWork      = "工作" // 工作相关
	SystemTagPersonal  = "个人" // 个人相关
	SystemTagImportant = "重要" // 重要文件
	SystemTagFavorite  = "收藏" // 收藏文件
	SystemTagTemp      = "临时" // 临时文件
	SystemTagProject   = "项目" // 项目文件
	SystemTagArchive   = "归档" // 归档文件
	SystemTagPublic    = "公开" // 公开文件
	SystemTagShared    = "共享" // 共享文件
	SystemTagRecent    = "最近" // 最近使用
)

// 系统标签颜色常量
const (
	TagColorBlue   = "#1890ff" // 蓝色
	TagColorGreen  = "#52c41a" // 绿色
	TagColorRed    = "#ff4d4f" // 红色
	TagColorOrange = "#fa8c16" // 橙色
	TagColorPurple = "#722ed1" // 紫色
	TagColorGold   = "#faad14" // 金色
	TagColorCyan   = "#13c2c2" // 青色
	TagColorGray   = "#8c8c8c" // 灰色
)

// 标签分类常量
const (
	TagCategoryWork     = "work"     // 工作
	TagCategoryPersonal = "personal" // 个人
	TagCategoryProject  = "project"  // 项目
	TagCategoryDocument = "document" // 文档
	TagCategoryMedia    = "media"    // 媒体
	TagCategoryOther    = "other"    // 其他
)
