package models

import (
	"database/sql"
	"testing"
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // 使用纯Go的SQLite驱动
)

// 为SQLite测试创建的兼容模型
type FileTest struct {
	basemodels.BaseModel
	UUID     string `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	UserID   uint   `gorm:"not null;index" json:"user_id"`
	ParentID *uint  `gorm:"index" json:"parent_id,omitempty"`
	Name     string `gorm:"type:varchar(255);not null" json:"name"`
	Path     string `gorm:"type:varchar(2000);not null;index" json:"path"`

	IsFolder  bool    `gorm:"default:false;index" json:"is_folder"`
	MimeType  *string `gorm:"type:varchar(255)" json:"mime_type,omitempty"`
	Extension *string `gorm:"type:varchar(50)" json:"extension,omitempty"`
	Size      int64   `gorm:"default:0" json:"size"`
	Hash      *string `gorm:"type:varchar(255);index" json:"hash,omitempty"`
	HashType  *string `gorm:"type:varchar(20);default:'md5'" json:"hash_type,omitempty"`

	StorageType   string  `gorm:"type:varchar(20);default:'local'" json:"storage_type"`
	StoragePath   *string `gorm:"type:varchar(2000)" json:"storage_path,omitempty"`
	StorageBucket *string `gorm:"type:varchar(255)" json:"storage_bucket,omitempty"`

	IsEncrypted   bool    `gorm:"default:false" json:"is_encrypted"`
	EncryptionKey *string `gorm:"type:varchar(255)" json:"-"`
	AccessLevel   string  `gorm:"type:varchar(20);default:'private'" json:"access_level"`

	Status       string  `gorm:"type:varchar(20);default:'active'" json:"status"`
	UploadStatus string  `gorm:"type:varchar(20);default:'completed'" json:"upload_status"`
	ThumbnailURL *string `gorm:"type:varchar(500)" json:"thumbnail_url,omitempty"`
	PreviewURL   *string `gorm:"type:varchar(500)" json:"preview_url,omitempty"`

	Metadata    *basemodels.JSONMap `gorm:"type:text" json:"metadata,omitempty"`
	Tags        *string             `gorm:"type:varchar(1000)" json:"tags,omitempty"`
	Description *string             `gorm:"type:text" json:"description,omitempty"`

	DownloadCount int64 `gorm:"default:0" json:"download_count"`
	ViewCount     int64 `gorm:"default:0" json:"view_count"`
	ShareCount    int64 `gorm:"default:0" json:"share_count"`

	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
}

func (FileTest) TableName() string {
	return "files"
}

func (f *FileTest) BeforeCreate(tx *gorm.DB) error {
	if f.UUID == "" {
		f.UUID = basemodels.GenerateUUID()
	}
	return f.BaseModel.BeforeCreate(tx)
}

func (f *FileTest) IsActive() bool {
	return f.Status == "active"
}

func (f *FileTest) IsImage() bool {
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

func (f *FileTest) IsVideo() bool {
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

func (f *FileTest) GetFullPath() string {
	if f.Path == "" {
		return f.Name
	}
	if f.Path == "/" {
		return "/" + f.Name
	}
	return f.Path + "/" + f.Name
}

type FileVersionTest struct {
	basemodels.BaseModel
	FileID        uint                `gorm:"not null;index" json:"file_id"`
	VersionNumber int                 `gorm:"not null" json:"version_number"`
	Name          string              `gorm:"type:varchar(255);not null" json:"name"`
	Size          int64               `gorm:"default:0" json:"size"`
	Hash          string              `gorm:"type:varchar(255);not null" json:"hash"`
	StoragePath   string              `gorm:"type:varchar(2000);not null" json:"storage_path"`
	MimeType      *string             `gorm:"type:varchar(255)" json:"mime_type,omitempty"`
	Metadata      *basemodels.JSONMap `gorm:"type:text" json:"metadata,omitempty"`
	ChangeLog     *string             `gorm:"type:text" json:"change_log,omitempty"`
	CreatedBy     uint                `gorm:"not null" json:"created_by"`
}

func (FileVersionTest) TableName() string {
	return "file_versions"
}

type FileShareTest struct {
	basemodels.BaseModel
	FileID    uint   `gorm:"not null;index" json:"file_id"`
	SharerID  uint   `gorm:"not null;index" json:"sharer_id"`
	ShareCode string `gorm:"type:varchar(100);uniqueIndex;not null" json:"share_code"`
	ShareURL  string `gorm:"type:varchar(500);not null" json:"share_url"`

	Permission  string  `gorm:"type:varchar(20);default:'view'" json:"permission"`
	Password    *string `gorm:"type:varchar(255)" json:"-"`
	HasPassword bool    `gorm:"default:false" json:"has_password"`

	MaxAccess     *int `json:"max_access,omitempty"`
	AccessCount   int  `gorm:"default:0" json:"access_count"`
	MaxDownload   *int `json:"max_download,omitempty"`
	DownloadCount int  `gorm:"default:0" json:"download_count"`

	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`

	Status string `gorm:"type:varchar(20);default:'active'" json:"status"`

	Settings *basemodels.JSONMap `gorm:"type:text" json:"settings,omitempty"`
}

func (FileShareTest) TableName() string {
	return "file_shares"
}

func (s *FileShareTest) BeforeCreate(tx *gorm.DB) error {
	if s.ShareCode == "" {
		s.ShareCode = basemodels.GenerateShareCode()
	}
	return s.BaseModel.BeforeCreate(tx)
}

func (s *FileShareTest) IsExpired() bool {
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}

func (s *FileShareTest) IsAccessible() bool {
	return s.Status == "active" && !s.IsExpired() && (s.MaxAccess == nil || s.AccessCount < *s.MaxAccess)
}

type FileTagTest struct {
	basemodels.BaseModel
	FileID uint   `gorm:"not null;index" json:"file_id"`
	UserID uint   `gorm:"not null;index" json:"user_id"`
	Tag    string `gorm:"type:varchar(100);not null" json:"tag"`
}

func (FileTagTest) TableName() string {
	return "file_tags"
}

func (t *FileTagTest) BeforeCreate(tx *gorm.DB) error {
	var count int64
	tx.Model(&FileTagTest{}).Where("file_id = ? AND user_id = ? AND tag = ?",
		t.FileID, t.UserID, t.Tag).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return t.BaseModel.BeforeCreate(tx)
}

type FileUploadChunkTest struct {
	basemodels.BaseModel
	UploadID    string    `gorm:"type:varchar(255);not null;index" json:"upload_id"`
	FileID      *uint     `gorm:"index" json:"file_id,omitempty"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	FileName    string    `gorm:"type:varchar(255);not null" json:"file_name"`
	FileSize    int64     `gorm:"not null" json:"file_size"`
	FileHash    string    `gorm:"type:varchar(255);not null" json:"file_hash"`
	ChunkIndex  int       `gorm:"not null" json:"chunk_index"`
	ChunkSize   int64     `gorm:"not null" json:"chunk_size"`
	ChunkHash   string    `gorm:"type:varchar(255);not null" json:"chunk_hash"`
	TotalChunks int       `gorm:"not null" json:"total_chunks"`
	StoragePath string    `gorm:"type:varchar(2000);not null" json:"storage_path"`
	Status      string    `gorm:"type:varchar(20);default:'uploading'" json:"status"`
	ExpiresAt   time.Time `gorm:"not null;index" json:"expires_at"`
}

func (FileUploadChunkTest) TableName() string {
	return "file_upload_chunks"
}

func (c *FileUploadChunkTest) BeforeCreate(tx *gorm.DB) error {
	if c.ExpiresAt.IsZero() {
		c.ExpiresAt = time.Now().Add(24 * time.Hour)
	}
	return c.BaseModel.BeforeCreate(tx)
}

func (c *FileUploadChunkTest) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

func (c *FileUploadChunkTest) IsCompleted() bool {
	return c.Status == "completed"
}

// setupFileTestDB 设置文件测试数据库
func setupFileTestDB() (*gorm.DB, error) {
	// 直接使用modernc.org/sqlite驱动
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}

	// 使用GORM打开已存在的数据库连接
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移测试模型
	err = db.AutoMigrate(
		&UserTest{},
		&FileTest{},
		&FileVersionTest{},
		&FileShareTest{},
		&FileTagTest{},
		&FileUploadChunkTest{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestFile_TableName(t *testing.T) {
	file := &FileTest{}
	if file.TableName() != "files" {
		t.Errorf("Expected table name 'files', got '%s'", file.TableName())
	}
}

func TestFile_BeforeCreate(t *testing.T) {
	db, err := setupFileTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	file := &FileTest{
		Name:   "test.txt",
		Path:   "/test",
		UserID: 1,
	}

	// 测试创建前UUID生成
	if err := file.BeforeCreate(db); err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}

	if file.UUID == "" {
		t.Error("UUID should be generated in BeforeCreate")
	}
}

func TestFile_IsActive(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"active", true},
		{"uploading", false},
		{"processing", false},
		{"error", false},
		{"deleted", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			file := &FileTest{Status: tt.status}
			if file.IsActive() != tt.expected {
				t.Errorf("IsActive() for status '%s' = %v, want %v", tt.status, file.IsActive(), tt.expected)
			}
		})
	}
}

func TestFile_FileTypeMethods(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		isImage  bool
		isVideo  bool
	}{
		{"jpeg image", "image/jpeg", true, false},
		{"png image", "image/png", true, false},
		{"gif image", "image/gif", true, false},
		{"mp4 video", "video/mp4", false, true},
		{"avi video", "video/avi", false, true},
		{"text file", "text/plain", false, false},
		{"pdf file", "application/pdf", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &FileTest{MimeType: &tt.mimeType}

			if file.IsImage() != tt.isImage {
				t.Errorf("IsImage() = %v, want %v", file.IsImage(), tt.isImage)
			}

			if file.IsVideo() != tt.isVideo {
				t.Errorf("IsVideo() = %v, want %v", file.IsVideo(), tt.isVideo)
			}
		})
	}

	// Test with nil mime type
	fileWithNilMime := &FileTest{}
	if fileWithNilMime.IsImage() {
		t.Error("IsImage() should return false for nil mime type")
	}
	if fileWithNilMime.IsVideo() {
		t.Error("IsVideo() should return false for nil mime type")
	}
}

func TestFile_GetFullPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		fileName string
		expected string
	}{
		{"with path", "/documents", "test.txt", "/documents/test.txt"},
		{"empty path", "", "test.txt", "test.txt"},
		{"root path", "/", "test.txt", "/test.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &FileTest{
				Path: tt.path,
				Name: tt.fileName,
			}

			if fullPath := file.GetFullPath(); fullPath != tt.expected {
				t.Errorf("GetFullPath() = %v, want %v", fullPath, tt.expected)
			}
		})
	}
}

func TestFileVersion_TableName(t *testing.T) {
	version := &FileVersionTest{}
	if version.TableName() != "file_versions" {
		t.Errorf("Expected table name 'file_versions', got '%s'", version.TableName())
	}
}

func TestFileShare_TableName(t *testing.T) {
	share := &FileShareTest{}
	if share.TableName() != "file_shares" {
		t.Errorf("Expected table name 'file_shares', got '%s'", share.TableName())
	}
}

func TestFileShare_BeforeCreate(t *testing.T) {
	db, err := setupFileTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	share := &FileShareTest{
		FileID:   1,
		SharerID: 1,
		ShareURL: "https://example.com/share/123",
	}

	// 测试创建前ShareCode生成
	if err := share.BeforeCreate(db); err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}

	if share.ShareCode == "" {
		t.Error("ShareCode should be generated in BeforeCreate")
	}
}

func TestFileShare_IsExpired(t *testing.T) {
	now := time.Now()

	// Test with no expiration
	neverExpireShare := &FileShareTest{}
	if neverExpireShare.IsExpired() {
		t.Error("Share with no expiration should not be expired")
	}

	// Test expired share
	expiredTime := now.Add(-time.Hour)
	expiredShare := &FileShareTest{ExpiresAt: &expiredTime}
	if !expiredShare.IsExpired() {
		t.Error("Share should be expired")
	}

	// Test valid share
	validTime := now.Add(time.Hour)
	validShare := &FileShareTest{ExpiresAt: &validTime}
	if validShare.IsExpired() {
		t.Error("Share should not be expired")
	}
}

func TestFileShare_IsAccessible(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		share    *FileShareTest
		expected bool
	}{
		{
			name: "active and not expired",
			share: &FileShareTest{
				Status: "active",
			},
			expected: true,
		},
		{
			name: "inactive share",
			share: &FileShareTest{
				Status: "inactive",
			},
			expected: false,
		},
		{
			name: "expired share",
			share: &FileShareTest{
				Status:    "active",
				ExpiresAt: &[]time.Time{now.Add(-time.Hour)}[0],
			},
			expected: false,
		},
		{
			name: "max access reached",
			share: &FileShareTest{
				Status:      "active",
				MaxAccess:   &[]int{5}[0],
				AccessCount: 5,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.share.IsAccessible() != tt.expected {
				t.Errorf("IsAccessible() = %v, want %v", tt.share.IsAccessible(), tt.expected)
			}
		})
	}
}

func TestFileTag_TableName(t *testing.T) {
	tag := &FileTagTest{}
	if tag.TableName() != "file_tags" {
		t.Errorf("Expected table name 'file_tags', got '%s'", tag.TableName())
	}
}

func TestFileTag_BeforeCreate(t *testing.T) {
	db, err := setupFileTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create first tag
	tag1 := &FileTagTest{
		FileID: 1,
		UserID: 1,
		Tag:    "important",
	}

	if err := tag1.BeforeCreate(db); err != nil {
		t.Fatalf("First tag creation should succeed: %v", err)
	}

	// Actually save it to database
	result := db.Create(tag1)
	if result.Error != nil {
		t.Fatalf("Failed to save first tag: %v", result.Error)
	}

	// Try to create duplicate tag
	tag2 := &FileTagTest{
		FileID: 1,
		UserID: 1,
		Tag:    "important",
	}

	if err := tag2.BeforeCreate(db); err == nil {
		t.Error("Should fail to create duplicate tag")
	}
}

func TestFileUploadChunk_TableName(t *testing.T) {
	chunk := &FileUploadChunkTest{}
	if chunk.TableName() != "file_upload_chunks" {
		t.Errorf("Expected table name 'file_upload_chunks', got '%s'", chunk.TableName())
	}
}

func TestFileUploadChunk_BeforeCreate(t *testing.T) {
	db, err := setupFileTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	chunk := &FileUploadChunkTest{
		UploadID:    "upload123",
		UserID:      1,
		FileName:    "test.txt",
		FileSize:    1000,
		FileHash:    "abcd1234",
		ChunkIndex:  0,
		ChunkSize:   100,
		ChunkHash:   "efgh5678",
		TotalChunks: 10,
		StoragePath: "/tmp/chunk0",
	}

	// 测试创建前过期时间设置
	if err := chunk.BeforeCreate(db); err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}

	if chunk.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should be set in BeforeCreate")
	}

	// 检查默认过期时间是否为24小时后
	expectedTime := time.Now().Add(24 * time.Hour)
	timeDiff := chunk.ExpiresAt.Sub(expectedTime)
	if timeDiff > time.Minute || timeDiff < -time.Minute {
		t.Error("ExpiresAt should be approximately 24 hours from now")
	}
}

func TestFileUploadChunk_IsExpired(t *testing.T) {
	now := time.Now()

	// Test expired chunk
	expiredChunk := &FileUploadChunkTest{
		ExpiresAt: now.Add(-time.Hour),
	}
	if !expiredChunk.IsExpired() {
		t.Error("Chunk should be expired")
	}

	// Test valid chunk
	validChunk := &FileUploadChunkTest{
		ExpiresAt: now.Add(time.Hour),
	}
	if validChunk.IsExpired() {
		t.Error("Chunk should not be expired")
	}
}

func TestFileUploadChunk_IsCompleted(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"completed", true},
		{"uploading", false},
		{"failed", false},
		{"merged", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			chunk := &FileUploadChunkTest{Status: tt.status}
			if chunk.IsCompleted() != tt.expected {
				t.Errorf("IsCompleted() for status '%s' = %v, want %v", tt.status, chunk.IsCompleted(), tt.expected)
			}
		})
	}
}

func TestCreateFileWithDatabase(t *testing.T) {
	db, err := setupFileTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create user first
	user := &UserTest{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
	}
	result := db.Create(user)
	if result.Error != nil {
		t.Fatalf("Failed to create user: %v", result.Error)
	}

	// Create file
	mimeType := "text/plain"
	file := &FileTest{
		UserID:      user.ID,
		Name:        "test.txt",
		Path:        "/documents",
		IsFolder:    false,
		MimeType:    &mimeType,
		Size:        1024,
		StorageType: "local",
		Status:      "active",
	}

	// Test creating file in database
	result = db.Create(file)
	if result.Error != nil {
		t.Fatalf("Failed to create file: %v", result.Error)
	}

	// Verify file was created with UUID
	if file.UUID == "" {
		t.Error("File UUID should be generated")
	}

	if file.ID == 0 {
		t.Error("File ID should be assigned")
	}

	// Test retrieving file
	var retrievedFile FileTest
	result = db.First(&retrievedFile, file.ID)
	if result.Error != nil {
		t.Fatalf("Failed to retrieve file: %v", result.Error)
	}

	if retrievedFile.Name != file.Name {
		t.Errorf("Retrieved file name = %v, want %v", retrievedFile.Name, file.Name)
	}

	if !retrievedFile.IsActive() {
		t.Error("Retrieved file should be active")
	}
}
