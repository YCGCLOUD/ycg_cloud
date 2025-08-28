package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型，包含通用字段
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Version   int64          `gorm:"default:1" json:"version"` // 乐观锁版本号
}

// BaseModelWithoutSoftDelete 不包含软删除的基础模型
type BaseModelWithoutSoftDelete struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int64     `gorm:"default:1" json:"version"`
}

// TimeModel 仅包含时间字段的模型
type TimeModel struct {
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// VersionedModel 带版本控制的模型接口
type VersionedModel interface {
	GetVersion() int64
	SetVersion(version int64)
}

// GetVersion 获取版本号
func (m *BaseModel) GetVersion() int64 {
	return m.Version
}

// SetVersion 设置版本号
func (m *BaseModel) SetVersion(version int64) {
	m.Version = version
}

// GetVersion 获取版本号
func (m *BaseModelWithoutSoftDelete) GetVersion() int64 {
	return m.Version
}

// SetVersion 设置版本号
func (m *BaseModelWithoutSoftDelete) SetVersion(version int64) {
	m.Version = version
}

// BeforeCreate GORM钩子：创建前
func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if m.Version == 0 {
		m.Version = 1
	}
	return nil
}

// BeforeUpdate GORM钩子：更新前，自动增加版本号
func (m *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	m.Version++
	return nil
}

// BeforeCreate GORM钩子：创建前
func (m *BaseModelWithoutSoftDelete) BeforeCreate(tx *gorm.DB) error {
	if m.Version == 0 {
		m.Version = 1
	}
	return nil
}

// BeforeUpdate GORM钩子：更新前，自动增加版本号
func (m *BaseModelWithoutSoftDelete) BeforeUpdate(tx *gorm.DB) error {
	m.Version++
	return nil
}

// TableName 表名接口
type TableNamer interface {
	TableName() string
}

// SoftDeleteModel 软删除模型接口
type SoftDeleteModel interface {
	IsDeleted() bool
	SoftDelete() error
	Restore() error
}

// IsDeleted 检查是否已软删除
func (m *BaseModel) IsDeleted() bool {
	return m.DeletedAt.Valid
}

// GetDeletedAt 获取删除时间
func (m *BaseModel) GetDeletedAt() *time.Time {
	if m.DeletedAt.Valid {
		return &m.DeletedAt.Time
	}
	return nil
}

// AuditModel 审计模型，记录创建者和更新者
type AuditModel struct {
	BaseModel
	CreatedBy uint `json:"created_by"` // 创建者ID
	UpdatedBy uint `json:"updated_by"` // 更新者ID
}

// SetCreatedBy 设置创建者
func (m *AuditModel) SetCreatedBy(userID uint) {
	m.CreatedBy = userID
}

// SetUpdatedBy 设置更新者
func (m *AuditModel) SetUpdatedBy(userID uint) {
	m.UpdatedBy = userID
}

// BeforeCreate GORM钩子：创建前
func (m *AuditModel) BeforeCreate(tx *gorm.DB) error {
	if m.Version == 0 {
		m.Version = 1
	}
	// 如果没有设置创建者，尝试从上下文获取
	if m.CreatedBy == 0 {
		if userID, ok := tx.Get("current_user_id"); ok {
			if uid, valid := userID.(uint); valid {
				m.CreatedBy = uid
			}
		}
	}
	m.UpdatedBy = m.CreatedBy
	return nil
}

// BeforeUpdate GORM钩子：更新前
func (m *AuditModel) BeforeUpdate(tx *gorm.DB) error {
	m.Version++
	// 如果没有设置更新者，尝试从上下文获取
	if m.UpdatedBy == 0 {
		if userID, ok := tx.Get("current_user_id"); ok {
			if uid, valid := userID.(uint); valid {
				m.UpdatedBy = uid
			}
		}
	}
	return nil
}

// StatusModel 带状态的模型
type StatusModel struct {
	BaseModel
	Status string `gorm:"type:varchar(50);default:'active';index" json:"status"`
}

// 常用状态常量
const (
	StatusActive   = "active"   // 激活
	StatusInactive = "inactive" // 未激活
	StatusDisabled = "disabled" // 禁用
	StatusPending  = "pending"  // 待处理
	StatusDeleted  = "deleted"  // 已删除
)

// IsActive 检查是否激活
func (m *StatusModel) IsActive() bool {
	return m.Status == StatusActive
}

// Activate 激活
func (m *StatusModel) Activate() {
	m.Status = StatusActive
}

// Deactivate 停用
func (m *StatusModel) Deactivate() {
	m.Status = StatusInactive
}

// Disable 禁用
func (m *StatusModel) Disable() {
	m.Status = StatusDisabled
}

// SortModel 可排序模型
type SortModel struct {
	BaseModel
	Sort int `gorm:"default:0;index" json:"sort"` // 排序权重
}

// SetSort 设置排序权重
func (m *SortModel) SetSort(sort int) {
	m.Sort = sort
}

// GetSort 获取排序权重
func (m *SortModel) GetSort() int {
	return m.Sort
}
