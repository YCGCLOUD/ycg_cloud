package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// StorageProvider 存储提供商表结构
type StorageProvider struct {
	basemodels.BaseModel
	// 基本信息
	UUID string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 提供商唯一标识符
	Name string `gorm:"type:varchar(100);not null;unique" json:"name"`  // 提供商名称
	Type string `gorm:"type:varchar(50);not null" json:"type"`          // 提供商类型

	// 配置信息
	Config      *basemodels.JSONMap `gorm:"type:json;not null" json:"-"`            // 配置信息（加密存储）
	DisplayName string              `gorm:"type:varchar(255)" json:"display_name"`  // 显示名称
	Description *string             `gorm:"type:text" json:"description,omitempty"` // 描述

	// 状态信息
	IsActive   bool       `gorm:"default:true" json:"is_active"`                                                   // 是否启用
	IsDefault  bool       `gorm:"default:false" json:"is_default"`                                                 // 是否默认
	Status     string     `gorm:"type:enum('active','inactive','error','testing');default:'active'" json:"status"` // 状态
	LastTestAt *time.Time `json:"last_test_at,omitempty"`                                                          // 最后测试时间
	TestResult *string    `gorm:"type:text" json:"test_result,omitempty"`                                          // 测试结果

	// 容量信息
	TotalCapacity int64 `gorm:"default:0" json:"total_capacity"` // 总容量
	UsedCapacity  int64 `gorm:"default:0" json:"used_capacity"`  // 已用容量
	MaxFileSize   int64 `gorm:"default:0" json:"max_file_size"`  // 最大文件大小

	// 性能配置
	MaxConcurrency int `gorm:"default:10" json:"max_concurrency"` // 最大并发数
	ChunkSize      int `gorm:"default:5242880" json:"chunk_size"` // 分片大小（5MB）
	RetryCount     int `gorm:"default:3" json:"retry_count"`      // 重试次数
	Timeout        int `gorm:"default:30" json:"timeout"`         // 超时时间（秒）

	// 安全配置
	EnableEncryption bool    `gorm:"default:false" json:"enable_encryption"` // 启用加密
	EncryptionKey    *string `gorm:"type:varchar(255)" json:"-"`             // 加密密钥

	// 访问控制
	AllowedMimeTypes  *string `gorm:"type:text" json:"allowed_mime_types,omitempty"` // 允许的MIME类型
	BlockedMimeTypes  *string `gorm:"type:text" json:"blocked_mime_types,omitempty"` // 禁止的MIME类型
	AllowedExtensions *string `gorm:"type:text" json:"allowed_extensions,omitempty"` // 允许的扩展名
	BlockedExtensions *string `gorm:"type:text" json:"blocked_extensions,omitempty"` // 禁止的扩展名

	// 创建者信息
	CreatedBy uint `gorm:"not null" json:"created_by"` // 创建者ID
	UpdatedBy uint `gorm:"not null" json:"updated_by"` // 更新者ID

	// 关联关系
	Creator  User            `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater  User            `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	Policies []StoragePolicy `gorm:"foreignKey:ProviderID" json:"policies,omitempty"`
	Files    []File          `gorm:"foreignKey:StorageProviderID" json:"files,omitempty"`
}

// TableName 存储提供商表名
func (StorageProvider) TableName() string {
	return "storage_providers"
}

// BeforeCreate 创建前钩子
func (sp *StorageProvider) BeforeCreate(tx *gorm.DB) error {
	if sp.UUID == "" {
		sp.UUID = basemodels.GenerateUUID()
	}
	return sp.BaseModel.BeforeCreate(tx)
}

// GetCapacityUsagePercent 获取容量使用百分比
func (sp *StorageProvider) GetCapacityUsagePercent() float64 {
	if sp.TotalCapacity == 0 {
		return 0
	}
	return float64(sp.UsedCapacity) / float64(sp.TotalCapacity) * 100
}

// HasAvailableSpace 检查是否有可用空间
func (sp *StorageProvider) HasAvailableSpace(size int64) bool {
	if sp.TotalCapacity == 0 {
		return true // 无限容量
	}
	return sp.UsedCapacity+size <= sp.TotalCapacity
}

// CanStoreFileType 检查是否可以存储该文件类型
func (sp *StorageProvider) CanStoreFileType(mimeType, extension string) bool {
	// TODO: 实现MIME类型和扩展名检查逻辑
	return true
}

// StoragePolicy 存储策略表结构
type StoragePolicy struct {
	basemodels.BaseModel
	// 基本信息
	UUID       string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 策略唯一标识符
	ProviderID uint   `gorm:"not null;index" json:"provider_id"`              // 存储提供商ID
	Name       string `gorm:"type:varchar(255);not null" json:"name"`         // 策略名称

	// 策略信息
	Description *string `gorm:"type:text" json:"description,omitempty"` // 策略描述
	Priority    int     `gorm:"default:0" json:"priority"`              // 优先级
	IsActive    bool    `gorm:"default:true" json:"is_active"`          // 是否启用
	IsDefault   bool    `gorm:"default:false" json:"is_default"`        // 是否默认策略

	// 应用条件
	Conditions *basemodels.JSONMap `gorm:"type:json;not null" json:"conditions"` // 应用条件

	// 策略配置
	Config *basemodels.JSONMap `gorm:"type:json" json:"config,omitempty"` // 策略配置

	// 生效范围
	ApplyToUsers   *string `gorm:"type:text" json:"apply_to_users,omitempty"`   // 应用到用户（ID列表）
	ApplyToGroups  *string `gorm:"type:text" json:"apply_to_groups,omitempty"`  // 应用到用户组
	ApplyToFolders *string `gorm:"type:text" json:"apply_to_folders,omitempty"` // 应用到文件夹

	// 统计信息
	AppliedCount  int        `gorm:"default:0" json:"applied_count"` // 应用次数
	LastAppliedAt *time.Time `json:"last_applied_at,omitempty"`      // 最后应用时间

	// 创建者信息
	CreatedBy uint `gorm:"not null" json:"created_by"` // 创建者ID

	// 关联关系
	Provider *StorageProvider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	Creator  User             `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName 存储策略表名
func (StoragePolicy) TableName() string {
	return "storage_policies"
}

// BeforeCreate 创建前钩子
func (sp *StoragePolicy) BeforeCreate(tx *gorm.DB) error {
	if sp.UUID == "" {
		sp.UUID = basemodels.GenerateUUID()
	}
	return sp.BaseModel.BeforeCreate(tx)
}

// IncrementApplied 增加应用次数
func (sp *StoragePolicy) IncrementApplied() {
	sp.AppliedCount++
	now := time.Now()
	sp.LastAppliedAt = &now
}

// StorageMigrationTask 存储迁移任务表结构
type StorageMigrationTask struct {
	basemodels.BaseModel
	// 基本信息
	UUID           string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 任务唯一标识符
	Name           string `gorm:"type:varchar(255);not null" json:"name"`         // 任务名称
	SourceProvider uint   `gorm:"not null;index" json:"source_provider"`          // 源存储提供商ID
	TargetProvider uint   `gorm:"not null;index" json:"target_provider"`          // 目标存储提供商ID

	// 任务状态
	Status      string     `gorm:"type:enum('pending','running','paused','completed','failed','cancelled');default:'pending'" json:"status"` // 任务状态
	Progress    int        `gorm:"default:0" json:"progress"`                                                                                // 进度百分比
	StartedAt   *time.Time `json:"started_at,omitempty"`                                                                                     // 开始时间
	CompletedAt *time.Time `json:"completed_at,omitempty"`                                                                                   // 完成时间

	// 迁移统计
	TotalFiles      int64 `gorm:"default:0" json:"total_files"`      // 总文件数
	ProcessedFiles  int64 `gorm:"default:0" json:"processed_files"`  // 已处理文件数
	SuccessfulFiles int64 `gorm:"default:0" json:"successful_files"` // 成功文件数
	FailedFiles     int64 `gorm:"default:0" json:"failed_files"`     // 失败文件数
	TotalSize       int64 `gorm:"default:0" json:"total_size"`       // 总大小
	ProcessedSize   int64 `gorm:"default:0" json:"processed_size"`   // 已处理大小
	TransferredSize int64 `gorm:"default:0" json:"transferred_size"` // 已传输大小

	// 错误信息
	ErrorMessage   *string `gorm:"type:text" json:"error_message,omitempty"`    // 错误信息
	FailedFileList *string `gorm:"type:text" json:"failed_file_list,omitempty"` // 失败文件列表

	// 任务配置
	Config  *basemodels.JSONMap `gorm:"type:json" json:"config,omitempty"`  // 任务配置
	Filters *basemodels.JSONMap `gorm:"type:json" json:"filters,omitempty"` // 过滤条件
	Options *basemodels.JSONMap `gorm:"type:json" json:"options,omitempty"` // 任务选项

	// 创建者信息
	CreatedBy uint `gorm:"not null" json:"created_by"` // 创建者ID

	// 关联关系
	SourceStorageProvider *StorageProvider `gorm:"foreignKey:SourceProvider" json:"source_storage_provider,omitempty"`
	TargetStorageProvider *StorageProvider `gorm:"foreignKey:TargetProvider" json:"target_storage_provider,omitempty"`
	Creator               User             `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName 存储迁移任务表名
func (StorageMigrationTask) TableName() string {
	return "storage_migration_tasks"
}

// BeforeCreate 创建前钩子
func (smt *StorageMigrationTask) BeforeCreate(tx *gorm.DB) error {
	if smt.UUID == "" {
		smt.UUID = basemodels.GenerateUUID()
	}
	return smt.BaseModel.BeforeCreate(tx)
}

// UpdateProgress 更新进度
func (smt *StorageMigrationTask) UpdateProgress() {
	if smt.TotalFiles > 0 {
		smt.Progress = int((smt.ProcessedFiles * 100) / smt.TotalFiles)
	}
}

// MarkAsStarted 标记为开始
func (smt *StorageMigrationTask) MarkAsStarted() {
	smt.Status = "running"
	now := time.Now()
	smt.StartedAt = &now
}

// MarkAsCompleted 标记为完成
func (smt *StorageMigrationTask) MarkAsCompleted() {
	smt.Status = "completed"
	smt.Progress = 100
	now := time.Now()
	smt.CompletedAt = &now
}

// MarkAsFailed 标记为失败
func (smt *StorageMigrationTask) MarkAsFailed(errorMsg string) {
	smt.Status = "failed"
	smt.ErrorMessage = &errorMsg
}

// 存储策略条件类型常量
const (
	PolicyConditionFileSize   = "file_size"   // 文件大小
	PolicyConditionMimeType   = "mime_type"   // MIME类型
	PolicyConditionExtension  = "extension"   // 文件扩展名
	PolicyConditionUser       = "user"        // 用户
	PolicyConditionUserGroup  = "user_group"  // 用户组
	PolicyConditionFolder     = "folder"      // 文件夹
	PolicyConditionUploadTime = "upload_time" // 上传时间
	PolicyConditionAccessFreq = "access_freq" // 访问频率
	PolicyConditionAge        = "age"         // 文件年龄
)

// 迁移任务状态常量
const (
	MigrationStatusPending   = "pending"   // 待处理
	MigrationStatusRunning   = "running"   // 运行中
	MigrationStatusPaused    = "paused"    // 暂停
	MigrationStatusCompleted = "completed" // 完成
	MigrationStatusFailed    = "failed"    // 失败
	MigrationStatusCancelled = "cancelled" // 取消
)
