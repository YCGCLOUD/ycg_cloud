package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// File 文件表结构
type File struct {
	basemodels.BaseModel
	// 基本信息
	UUID     string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 文件唯一标识符
	UserID   uint   `gorm:"not null;index" json:"user_id"`                  // 所属用户ID
	ParentID *uint  `gorm:"index" json:"parent_id,omitempty"`               // 父文件夹ID
	Name     string `gorm:"type:varchar(255);not null" json:"name"`         // 文件名
	Path     string `gorm:"type:varchar(2000);not null;index" json:"path"`  // 文件路径

	// 文件类型和内容信息
	IsFolder  bool    `gorm:"default:false;index" json:"is_folder"`                                      // 是否为文件夹
	MimeType  *string `gorm:"type:varchar(255)" json:"mime_type,omitempty"`                              // MIME类型
	Extension *string `gorm:"type:varchar(50)" json:"extension,omitempty"`                               // 文件扩展名
	Size      int64   `gorm:"default:0" json:"size"`                                                     // 文件大小(字节)
	Hash      *string `gorm:"type:varchar(255);index" json:"hash,omitempty"`                             // 文件哈希值(MD5/SHA256)
	HashType  *string `gorm:"type:enum('md5','sha1','sha256');default:'md5'" json:"hash_type,omitempty"` // 哈希类型

	// 存储信息
	StorageType   string  `gorm:"type:enum('local','oss','s3','minio');default:'local'" json:"storage_type"` // 存储类型
	StoragePath   *string `gorm:"type:varchar(2000)" json:"storage_path,omitempty"`                          // 实际存储路径
	StorageBucket *string `gorm:"type:varchar(255)" json:"storage_bucket,omitempty"`                         // 存储桶名称

	// 安全和权限
	IsEncrypted   bool    `gorm:"default:false" json:"is_encrypted"`                                            // 是否加密
	EncryptionKey *string `gorm:"type:varchar(255)" json:"-"`                                                   // 加密密钥(不返回)
	AccessLevel   string  `gorm:"type:enum('private','public','shared');default:'private'" json:"access_level"` // 访问级别

	// 状态信息
	Status       string  `gorm:"type:enum('uploading','processing','active','error','deleted');default:'active'" json:"status"`  // 文件状态
	UploadStatus string  `gorm:"type:enum('pending','uploading','completed','failed');default:'completed'" json:"upload_status"` // 上传状态
	ThumbnailURL *string `gorm:"type:varchar(500)" json:"thumbnail_url,omitempty"`                                               // 缩略图URL
	PreviewURL   *string `gorm:"type:varchar(500)" json:"preview_url,omitempty"`                                                 // 预览URL

	// 元数据
	Metadata    *basemodels.JSONMap `gorm:"type:json" json:"metadata,omitempty"`      // 文件元数据
	Tags        *string             `gorm:"type:varchar(1000)" json:"tags,omitempty"` // 标签(逗号分隔)
	Description *string             `gorm:"type:text" json:"description,omitempty"`   // 文件描述

	// 统计信息
	DownloadCount int64 `gorm:"default:0" json:"download_count"` // 下载次数
	ViewCount     int64 `gorm:"default:0" json:"view_count"`     // 查看次数
	ShareCount    int64 `gorm:"default:0" json:"share_count"`    // 分享次数

	// 时间信息
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"` // 最后访问时间

	// 关联关系
	Owner        User              `gorm:"foreignKey:UserID" json:"owner,omitempty"`
	Parent       *File             `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children     []File            `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Versions     []FileVersion     `gorm:"foreignKey:FileID" json:"versions,omitempty"`
	Shares       []FileShare       `gorm:"foreignKey:FileID" json:"shares,omitempty"`
	FileTags     []FileTag         `gorm:"foreignKey:FileID" json:"file_tags,omitempty"`
	UploadChunks []FileUploadChunk `gorm:"foreignKey:FileID" json:"upload_chunks,omitempty"`
}

// TableName 文件表名
func (File) TableName() string {
	return "files"
}

// BeforeCreate 创建前钩子
func (f *File) BeforeCreate(tx *gorm.DB) error {
	if f.UUID == "" {
		f.UUID = basemodels.GenerateUUID()
	}
	return f.BaseModel.BeforeCreate(tx)
}

// IsActive 检查文件是否活动
func (f *File) IsActive() bool {
	return f.Status == "active"
}

// IsImage 检查是否为图片文件
func (f *File) IsImage() bool {
	if f.MimeType == nil {
		return false
	}
	imageTypes := []string{"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp", "image/bmp"}
	for _, t := range imageTypes {
		if *f.MimeType == t {
			return true
		}
	}
	return false
}

// IsVideo 检查是否为视频文件
func (f *File) IsVideo() bool {
	if f.MimeType == nil {
		return false
	}
	videoTypes := []string{"video/mp4", "video/avi", "video/mkv", "video/mov", "video/wmv", "video/flv"}
	for _, t := range videoTypes {
		if *f.MimeType == t {
			return true
		}
	}
	return false
}

// GetFullPath 获取完整路径
func (f *File) GetFullPath() string {
	if f.Path == "" {
		return f.Name
	}
	if f.Path == "/" {
		return "/" + f.Name
	}
	return f.Path + "/" + f.Name
}

// FileVersion 文件版本表结构
type FileVersion struct {
	basemodels.BaseModel
	FileID        uint                `gorm:"not null;index" json:"file_id"`                   // 文件ID
	VersionNumber int                 `gorm:"not null" json:"version_number"`                  // 版本号
	Name          string              `gorm:"type:varchar(255);not null" json:"name"`          // 版本名称
	Size          int64               `gorm:"default:0" json:"size"`                           // 文件大小
	Hash          string              `gorm:"type:varchar(255);not null" json:"hash"`          // 文件哈希值
	StoragePath   string              `gorm:"type:varchar(2000);not null" json:"storage_path"` // 存储路径
	MimeType      *string             `gorm:"type:varchar(255)" json:"mime_type,omitempty"`    // MIME类型
	Metadata      *basemodels.JSONMap `gorm:"type:json" json:"metadata,omitempty"`             // 版本元数据
	ChangeLog     *string             `gorm:"type:text" json:"change_log,omitempty"`           // 变更日志
	CreatedBy     uint                `gorm:"not null" json:"created_by"`                      // 创建者ID

	// 关联关系
	File    File `gorm:"foreignKey:FileID" json:"file,omitempty"`
	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName 文件版本表名
func (FileVersion) TableName() string {
	return "file_versions"
}

// FileShare 文件分享表结构
type FileShare struct {
	basemodels.BaseModel
	FileID    uint   `gorm:"not null;index" json:"file_id"`                            // 文件ID
	SharerID  uint   `gorm:"not null;index" json:"sharer_id"`                          // 分享者ID
	ShareCode string `gorm:"type:varchar(100);uniqueIndex;not null" json:"share_code"` // 分享码
	ShareURL  string `gorm:"type:varchar(500);not null" json:"share_url"`              // 分享链接

	// 权限设置
	Permission  string  `gorm:"type:enum('view','download','edit');default:'view'" json:"permission"` // 权限类型
	Password    *string `gorm:"type:varchar(255)" json:"-"`                                           // 分享密码(加密存储)
	HasPassword bool    `gorm:"default:false" json:"has_password"`                                    // 是否设置密码

	// 访问控制
	MaxAccess     *int `json:"max_access,omitempty"`            // 最大访问次数
	AccessCount   int  `gorm:"default:0" json:"access_count"`   // 已访问次数
	MaxDownload   *int `json:"max_download,omitempty"`          // 最大下载次数
	DownloadCount int  `gorm:"default:0" json:"download_count"` // 已下载次数

	// 时间控制
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`       // 过期时间
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"` // 最后访问时间

	// 状态
	Status string `gorm:"type:enum('active','expired','disabled','deleted');default:'active'" json:"status"` // 分享状态

	// 元数据
	Settings *basemodels.JSONMap `gorm:"type:json" json:"settings,omitempty"` // 分享设置

	// 关联关系
	File   File `gorm:"foreignKey:FileID" json:"file,omitempty"`
	Sharer User `gorm:"foreignKey:SharerID" json:"sharer,omitempty"`
}

// TableName 文件分享表名
func (FileShare) TableName() string {
	return "file_shares"
}

// BeforeCreate 创建前钩子
func (s *FileShare) BeforeCreate(tx *gorm.DB) error {
	if s.ShareCode == "" {
		s.ShareCode = basemodels.GenerateShareCode()
	}
	return s.BaseModel.BeforeCreate(tx)
}

// IsExpired 检查是否过期
func (s *FileShare) IsExpired() bool {
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}

// IsAccessible 检查是否可访问
func (s *FileShare) IsAccessible() bool {
	if s.Status != "active" {
		return false
	}
	if s.IsExpired() {
		return false
	}
	if s.MaxAccess != nil && s.AccessCount >= *s.MaxAccess {
		return false
	}
	return true
}

// FileTag 文件标签表结构
type FileTag struct {
	basemodels.BaseModel
	FileID      uint    `gorm:"not null;index" json:"file_id"`                  // 文件ID
	UserID      uint    `gorm:"not null;index" json:"user_id"`                  // 用户ID
	Tag         string  `gorm:"type:varchar(100);not null" json:"tag"`          // 标签名称
	Color       *string `gorm:"type:varchar(20)" json:"color,omitempty"`        // 标签颜色
	Description *string `gorm:"type:varchar(255)" json:"description,omitempty"` // 标签描述

	// 关联关系
	File File `gorm:"foreignKey:FileID" json:"file,omitempty"`
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 文件标签表名
func (FileTag) TableName() string {
	return "file_tags"
}

// BeforeCreate 创建前钩子
func (t *FileTag) BeforeCreate(tx *gorm.DB) error {
	// 确保同一用户对同一文件的同一标签唯一
	var count int64
	tx.Model(&FileTag{}).Where("file_id = ? AND user_id = ? AND tag = ?",
		t.FileID, t.UserID, t.Tag).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return t.BaseModel.BeforeCreate(tx)
}

// FileUploadChunk 文件分片上传表结构
type FileUploadChunk struct {
	basemodels.BaseModel
	FileID   *uint  `gorm:"index" json:"file_id,omitempty"`                    // 文件ID(上传完成后设置)
	UploadID string `gorm:"type:varchar(100);not null;index" json:"upload_id"` // 上传任务ID
	UserID   uint   `gorm:"not null;index" json:"user_id"`                     // 用户ID

	// 文件信息
	FileName string  `gorm:"type:varchar(255);not null" json:"file_name"`  // 原始文件名
	FileSize int64   `gorm:"not null" json:"file_size"`                    // 文件总大小
	FileHash string  `gorm:"type:varchar(255);not null" json:"file_hash"`  // 文件哈希值
	MimeType *string `gorm:"type:varchar(255)" json:"mime_type,omitempty"` // MIME类型

	// 分片信息
	ChunkIndex  int    `gorm:"not null" json:"chunk_index"`                  // 分片索引(从0开始)
	ChunkSize   int64  `gorm:"not null" json:"chunk_size"`                   // 分片大小
	ChunkHash   string `gorm:"type:varchar(255);not null" json:"chunk_hash"` // 分片哈希值
	TotalChunks int    `gorm:"not null" json:"total_chunks"`                 // 总分片数

	// 存储信息
	StoragePath string `gorm:"type:varchar(2000);not null" json:"storage_path"`                           // 分片存储路径
	StorageType string `gorm:"type:enum('local','oss','s3','minio');default:'local'" json:"storage_type"` // 存储类型

	// 状态信息
	Status string `gorm:"type:enum('uploading','completed','failed','merged');default:'uploading'" json:"status"` // 分片状态

	// 时间信息
	ExpiresAt   time.Time  `gorm:"not null;index" json:"expires_at"` // 过期时间(24小时)
	CompletedAt *time.Time `json:"completed_at,omitempty"`           // 完成时间

	// 关联关系
	File *File `gorm:"foreignKey:FileID" json:"file,omitempty"`
	User User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 文件分片上传表名
func (FileUploadChunk) TableName() string {
	return "file_upload_chunks"
}

// BeforeCreate 创建前钩子
func (c *FileUploadChunk) BeforeCreate(tx *gorm.DB) error {
	if c.ExpiresAt.IsZero() {
		c.ExpiresAt = time.Now().Add(24 * time.Hour) // 默认24小时过期
	}
	return c.BaseModel.BeforeCreate(tx)
}

// IsExpired 检查是否过期
func (c *FileUploadChunk) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IsCompleted 检查是否完成
func (c *FileUploadChunk) IsCompleted() bool {
	return c.Status == "completed"
}

// 文件状态常量
const (
	FileStatusUploading  = "uploading"  // 上传中
	FileStatusProcessing = "processing" // 处理中
	FileStatusActive     = "active"     // 活动
	FileStatusError      = "error"      // 错误
	FileStatusDeleted    = "deleted"    // 已删除
)

// 上传状态常量
const (
	UploadStatusPending   = "pending"   // 待上传
	UploadStatusUploading = "uploading" // 上传中
	UploadStatusCompleted = "completed" // 已完成
	UploadStatusFailed    = "failed"    // 上传失败
)

// 存储类型常量
const (
	StorageTypeLocal = "local" // 本地存储
	StorageTypeOSS   = "oss"   // 阿里云OSS
	StorageTypeS3    = "s3"    // Amazon S3
	StorageTypeMinio = "minio" // MinIO
)

// 访问级别常量
const (
	AccessLevelPrivate = "private" // 私有
	AccessLevelPublic  = "public"  // 公开
	AccessLevelShared  = "shared"  // 已分享
)

// 分享权限常量
const (
	SharePermissionView     = "view"     // 仅查看
	SharePermissionDownload = "download" // 可下载
	SharePermissionEdit     = "edit"     // 可编辑
)
