package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// OfflineOperation 离线操作表结构
type OfflineOperation struct {
	basemodels.BaseModel
	// 基本信息
	UUID     string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`    // 离线操作唯一标识符
	UserID   uint   `gorm:"not null;index" json:"user_id"`                     // 用户ID
	DeviceID string `gorm:"type:varchar(255);not null;index" json:"device_id"` // 设备ID

	// 操作信息
	Operation    string  `gorm:"type:varchar(100);not null;index" json:"operation"` // 操作类型
	ResourceType string  `gorm:"type:varchar(50);not null" json:"resource_type"`    // 资源类型
	ResourceID   *string `gorm:"type:varchar(100)" json:"resource_id,omitempty"`    // 资源ID
	ResourcePath *string `gorm:"type:varchar(2000)" json:"resource_path,omitempty"` // 资源路径

	// 操作数据
	OperationData *basemodels.JSONMap `gorm:"type:json" json:"operation_data,omitempty"` // 操作数据
	OldData       *basemodels.JSONMap `gorm:"type:json" json:"old_data,omitempty"`       // 原始数据
	NewData       *basemodels.JSONMap `gorm:"type:json" json:"new_data,omitempty"`       // 新数据

	// 同步状态
	Status       string     `gorm:"type:enum('pending','synced','failed','conflict');default:'pending'" json:"status"` // 同步状态
	SyncedAt     *time.Time `json:"synced_at,omitempty"`                                                               // 同步时间
	ErrorMessage *string    `gorm:"type:text" json:"error_message,omitempty"`                                          // 错误信息

	// 冲突解决
	ConflictType       *string             `gorm:"type:varchar(100)" json:"conflict_type,omitempty"`       // 冲突类型
	ConflictResolution *string             `gorm:"type:varchar(100)" json:"conflict_resolution,omitempty"` // 冲突解决方式
	ConflictData       *basemodels.JSONMap `gorm:"type:json" json:"conflict_data,omitempty"`               // 冲突数据

	// 时间信息
	OperationTime time.Time `gorm:"not null" json:"operation_time"` // 操作发生时间
	RetryCount    int       `gorm:"default:0" json:"retry_count"`   // 重试次数
	MaxRetries    int       `gorm:"default:3" json:"max_retries"`   // 最大重试次数

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 离线操作表名
func (OfflineOperation) TableName() string {
	return "offline_operations"
}

// BeforeCreate 创建前钩子
func (o *OfflineOperation) BeforeCreate(tx *gorm.DB) error {
	if o.UUID == "" {
		o.UUID = basemodels.GenerateUUID()
	}
	if o.OperationTime.IsZero() {
		o.OperationTime = time.Now()
	}
	return o.BaseModel.BeforeCreate(tx)
}

// CanRetry 检查是否可以重试
func (o *OfflineOperation) CanRetry() bool {
	return o.Status == "failed" && o.RetryCount < o.MaxRetries
}

// MarkSynced 标记为已同步
func (o *OfflineOperation) MarkSynced() {
	o.Status = "synced"
	now := time.Now()
	o.SyncedAt = &now
}

// MarkFailed 标记为失败
func (o *OfflineOperation) MarkFailed(errorMsg string) {
	o.Status = "failed"
	o.ErrorMessage = &errorMsg
	o.RetryCount++
}

// MarkConflict 标记为冲突
func (o *OfflineOperation) MarkConflict(conflictType string, conflictData *basemodels.JSONMap) {
	o.Status = "conflict"
	o.ConflictType = &conflictType
	o.ConflictData = conflictData
}

// OfflineFile 离线文件表结构
type OfflineFile struct {
	basemodels.BaseModel
	// 基本信息
	UUID     string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`    // 离线文件唯一标识符
	UserID   uint   `gorm:"not null;index" json:"user_id"`                     // 用户ID
	FileID   uint   `gorm:"not null;index" json:"file_id"`                     // 文件ID
	DeviceID string `gorm:"type:varchar(255);not null;index" json:"device_id"` // 设备ID

	// 离线状态
	Status       string     `gorm:"type:enum('cached','downloading','downloaded','expired','error');default:'cached'" json:"status"` // 离线状态
	CachedAt     time.Time  `gorm:"not null" json:"cached_at"`                                                                       // 缓存时间
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`                                                                            // 过期时间
	LastAccessed *time.Time `json:"last_accessed_at,omitempty"`                                                                      // 最后访问时间

	// 缓存信息
	CacheSize     int64   `gorm:"default:0" json:"cache_size"`                    // 缓存大小
	CachePath     *string `gorm:"type:varchar(2000)" json:"cache_path,omitempty"` // 本地缓存路径
	CacheVersion  int     `gorm:"default:1" json:"cache_version"`                 // 缓存版本
	RemoteVersion int     `gorm:"default:1" json:"remote_version"`                // 远程版本

	// 同步标记
	NeedSync      bool       `gorm:"default:false" json:"need_sync"`  // 是否需要同步
	LastSyncAt    *time.Time `json:"last_sync_at,omitempty"`          // 最后同步时间
	SyncConflicts int        `gorm:"default:0" json:"sync_conflicts"` // 同步冲突次数

	// 离线编辑
	HasLocalChanges bool       `gorm:"default:false" json:"has_local_changes"` // 是否有本地修改
	LocalModifiedAt *time.Time `json:"local_modified_at,omitempty"`            // 本地修改时间

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	File File `gorm:"foreignKey:FileID" json:"file,omitempty"`
}

// TableName 离线文件表名
func (OfflineFile) TableName() string {
	return "offline_files"
}

// BeforeCreate 创建前钩子
func (of *OfflineFile) BeforeCreate(tx *gorm.DB) error {
	if of.UUID == "" {
		of.UUID = basemodels.GenerateUUID()
	}
	if of.CachedAt.IsZero() {
		of.CachedAt = time.Now()
	}
	return of.BaseModel.BeforeCreate(tx)
}

// IsExpired 检查是否过期
func (of *OfflineFile) IsExpired() bool {
	if of.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*of.ExpiresAt)
}

// NeedsUpdate 检查是否需要更新
func (of *OfflineFile) NeedsUpdate() bool {
	return of.CacheVersion < of.RemoteVersion
}

// MarkAsModified 标记为已修改
func (of *OfflineFile) MarkAsModified() {
	of.HasLocalChanges = true
	of.NeedSync = true
	now := time.Now()
	of.LocalModifiedAt = &now
}

// SyncDevice 同步设备表结构
type SyncDevice struct {
	basemodels.BaseModel
	// 基本信息
	UUID     string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`    // 设备唯一标识符
	UserID   uint   `gorm:"not null;index" json:"user_id"`                     // 用户ID
	DeviceID string `gorm:"type:varchar(255);not null;index" json:"device_id"` // 设备ID

	// 设备信息
	DeviceName string  `gorm:"type:varchar(255);not null" json:"device_name"`  // 设备名称
	DeviceType string  `gorm:"type:varchar(50);not null" json:"device_type"`   // 设备类型
	Platform   string  `gorm:"type:varchar(50)" json:"platform,omitempty"`     // 平台
	Version    *string `gorm:"type:varchar(100)" json:"version,omitempty"`     // 应用版本
	UserAgent  *string `gorm:"type:varchar(1000)" json:"user_agent,omitempty"` // 用户代理

	// 同步状态
	IsOnline    bool       `gorm:"default:false" json:"is_online"` // 是否在线
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty"`         // 最后在线时间
	LastSyncAt  *time.Time `json:"last_sync_at,omitempty"`         // 最后同步时间
	SyncVersion int64      `gorm:"default:0" json:"sync_version"`  // 同步版本号

	// 离线文件统计
	OfflineFileCount int64 `gorm:"default:0" json:"offline_file_count"` // 离线文件数量
	CacheSize        int64 `gorm:"default:0" json:"cache_size"`         // 缓存大小
	MaxCacheSize     int64 `gorm:"default:0" json:"max_cache_size"`     // 最大缓存大小

	// 设置选项
	AutoSync         bool   `gorm:"default:true" json:"auto_sync"`                                                      // 自动同步
	SyncOnWiFiOnly   bool   `gorm:"default:false" json:"sync_on_wifi_only"`                                             // 仅WiFi同步
	SyncConflictMode string `gorm:"type:enum('ask','server','client','merge');default:'ask'" json:"sync_conflict_mode"` // 冲突处理模式

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 同步设备表名
func (SyncDevice) TableName() string {
	return "sync_devices"
}

// BeforeCreate 创建前钩子
func (sd *SyncDevice) BeforeCreate(tx *gorm.DB) error {
	if sd.UUID == "" {
		sd.UUID = basemodels.GenerateUUID()
	}
	// 确保同一用户的同一设备ID唯一
	var count int64
	tx.Model(&SyncDevice{}).Where("user_id = ? AND device_id = ?", sd.UserID, sd.DeviceID).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return sd.BaseModel.BeforeCreate(tx)
}

// UpdateOnlineStatus 更新在线状态
func (sd *SyncDevice) UpdateOnlineStatus(isOnline bool) {
	sd.IsOnline = isOnline
	now := time.Now()
	if isOnline {
		sd.LastSeenAt = &now
	}
}

// UpdateSyncVersion 更新同步版本
func (sd *SyncDevice) UpdateSyncVersion() {
	sd.SyncVersion++
	now := time.Now()
	sd.LastSyncAt = &now
}

// 离线操作类型常量
const (
	OfflineOperationCreateFile    = "create_file"    // 创建文件
	OfflineOperationUpdateFile    = "update_file"    // 更新文件
	OfflineOperationDeleteFile    = "delete_file"    // 删除文件
	OfflineOperationMoveFile      = "move_file"      // 移动文件
	OfflineOperationRenameFile    = "rename_file"    // 重命名文件
	OfflineOperationCreateFolder  = "create_folder"  // 创建文件夹
	OfflineOperationUpdateFolder  = "update_folder"  // 更新文件夹
	OfflineOperationDeleteFolder  = "delete_folder"  // 删除文件夹
	OfflineOperationAddTag        = "add_tag"        // 添加标签
	OfflineOperationRemoveTag     = "remove_tag"     // 移除标签
	OfflineOperationAddComment    = "add_comment"    // 添加评论
	OfflineOperationUpdateComment = "update_comment" // 更新评论
	OfflineOperationDeleteComment = "delete_comment" // 删除评论
)

// 设备类型常量
const (
	DeviceTypeDesktop = "desktop" // 桌面端
	DeviceTypeMobile  = "mobile"  // 移动端
	DeviceTypeTablet  = "tablet"  // 平板端
	DeviceTypeWeb     = "web"     // 网页端
)

// 冲突类型常量
const (
	ConflictTypeFileModified     = "file_modified"     // 文件被修改
	ConflictTypeFileDeleted      = "file_deleted"      // 文件被删除
	ConflictTypePermissionDenied = "permission_denied" // 权限被拒绝
	ConflictTypeStorageFull      = "storage_full"      // 存储空间不足
	ConflictTypeNetworkError     = "network_error"     // 网络错误
)

// 冲突解决方式常量
const (
	ConflictResolutionKeepServer = "keep_server" // 保留服务器版本
	ConflictResolutionKeepClient = "keep_client" // 保留客户端版本
	ConflictResolutionMerge      = "merge"       // 合并
	ConflictResolutionCreateCopy = "create_copy" // 创建副本
	ConflictResolutionSkip       = "skip"        // 跳过
)
