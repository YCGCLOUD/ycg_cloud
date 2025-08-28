package models

import (
	"testing"
	"time"

	. "cloudpan/internal/pkg/database/models"
)

// TestModelStructures 测试模型结构体的基本功能（不依赖数据库）
func TestModelStructures(t *testing.T) {
	t.Run("User Model", testUserModel)
	t.Run("File Model", testFileModel)
	t.Run("UserSession Model", testUserSessionModel)
	t.Run("FileShare Model", testFileShareModel)
	t.Run("UserPreference Model", testUserPreferenceModel)
	t.Run("FileUploadChunk Model", testFileUploadChunkModel)
}

func testUserModel(t *testing.T) {
	user := &User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Status:       "active",
		StorageQuota: 1000,
		StorageUsed:  300,
	}

	// Test table name
	if user.TableName() != "users" {
		t.Errorf("Expected table name 'users', got '%s'", user.TableName())
	}

	// Test status methods
	if !user.IsActive() {
		t.Error("User should be active")
	}

	user.Status = "suspended"
	if !user.IsSuspended() {
		t.Error("User should be suspended")
	}

	// Test storage methods
	if user.GetStorageUsagePercent() != 30.0 {
		t.Errorf("Expected storage usage 30%%, got %f", user.GetStorageUsagePercent())
	}

	if !user.HasStorageSpace(500) {
		t.Error("User should have enough storage space")
	}

	if user.HasStorageSpace(800) {
		t.Error("User should not have enough storage space")
	}
}

func testFileModel(t *testing.T) {
	mimeType := "image/jpeg"
	file := &File{
		Name:     "test.jpg",
		Path:     "/images",
		MimeType: &mimeType,
		Status:   "active",
	}

	// Test table name
	if file.TableName() != "files" {
		t.Errorf("Expected table name 'files', got '%s'", file.TableName())
	}

	// Test status
	if !file.IsActive() {
		t.Error("File should be active")
	}

	// Test file type methods
	if !file.IsImage() {
		t.Error("File should be an image")
	}

	if file.IsVideo() {
		t.Error("File should not be a video")
	}

	// Test path methods
	expectedPath := "/images/test.jpg"
	if file.GetFullPath() != expectedPath {
		t.Errorf("Expected full path '%s', got '%s'", expectedPath, file.GetFullPath())
	}

	// Test root path
	file.Path = "/"
	expectedRootPath := "/test.jpg"
	if file.GetFullPath() != expectedRootPath {
		t.Errorf("Expected root path '%s', got '%s'", expectedRootPath, file.GetFullPath())
	}

	// Test empty path
	file.Path = ""
	if file.GetFullPath() != "test.jpg" {
		t.Errorf("Expected name only, got '%s'", file.GetFullPath())
	}
}

func testUserSessionModel(t *testing.T) {
	now := time.Now()
	session := &UserSession{
		UserID:    1,
		IsActive:  true,
		ExpiresAt: now.Add(time.Hour),
	}

	// Test table name
	if session.TableName() != "user_sessions" {
		t.Errorf("Expected table name 'user_sessions', got '%s'", session.TableName())
	}

	// Test validity
	if !session.IsValid() {
		t.Error("Session should be valid")
	}

	if session.IsExpired() {
		t.Error("Session should not be expired")
	}

	// Test expired session
	session.ExpiresAt = now.Add(-time.Hour)
	if !session.IsExpired() {
		t.Error("Session should be expired")
	}

	if session.IsValid() {
		t.Error("Expired session should not be valid")
	}

	// Test inactive session
	session.ExpiresAt = now.Add(time.Hour)
	session.IsActive = false
	if session.IsValid() {
		t.Error("Inactive session should not be valid")
	}
}

func testFileShareModel(t *testing.T) {
	now := time.Now()
	share := &FileShare{
		Status: "active",
	}

	// Test table name
	if share.TableName() != "file_shares" {
		t.Errorf("Expected table name 'file_shares', got '%s'", share.TableName())
	}

	// Test accessibility without expiration
	if !share.IsAccessible() {
		t.Error("Share should be accessible")
	}

	// Test with expiration
	expiredTime := now.Add(-time.Hour)
	share.ExpiresAt = &expiredTime
	if share.IsAccessible() {
		t.Error("Expired share should not be accessible")
	}

	// Test with max access limit
	share.ExpiresAt = nil
	maxAccess := 5
	share.MaxAccess = &maxAccess
	share.AccessCount = 5
	if share.IsAccessible() {
		t.Error("Share with max access reached should not be accessible")
	}

	// Test inactive status
	share.AccessCount = 0
	share.Status = "inactive"
	if share.IsAccessible() {
		t.Error("Inactive share should not be accessible")
	}
}

func testUserPreferenceModel(t *testing.T) {
	pref := &UserPreference{}

	// Test table name
	if pref.TableName() != "user_preferences" {
		t.Errorf("Expected table name 'user_preferences', got '%s'", pref.TableName())
	}

	// Test boolean value methods
	pref.SetBoolValue(true)
	if !pref.GetBoolValue() {
		t.Error("Should return true for boolean value")
	}
	if pref.ValueType != "boolean" {
		t.Error("ValueType should be set to boolean")
	}

	pref.SetBoolValue(false)
	if pref.GetBoolValue() {
		t.Error("Should return false for boolean value")
	}

	// Test string value methods
	pref.SetStringValue("test value")
	if pref.GetStringValue() != "test value" {
		t.Error("Should return correct string value")
	}
	if pref.ValueType != "string" {
		t.Error("ValueType should be set to string")
	}

	// Test nil value
	nilPref := &UserPreference{}
	if nilPref.GetStringValue() != "" {
		t.Error("Should return empty string for nil value")
	}
	if nilPref.GetBoolValue() {
		t.Error("Should return false for nil value")
	}
}

func testFileUploadChunkModel(t *testing.T) {
	now := time.Now()
	chunk := &FileUploadChunk{
		Status:    "completed",
		ExpiresAt: now.Add(time.Hour),
	}

	// Test table name
	if chunk.TableName() != "file_upload_chunks" {
		t.Errorf("Expected table name 'file_upload_chunks', got '%s'", chunk.TableName())
	}

	// Test completion status
	if !chunk.IsCompleted() {
		t.Error("Chunk should be completed")
	}

	chunk.Status = "uploading"
	if chunk.IsCompleted() {
		t.Error("Chunk should not be completed")
	}

	// Test expiration
	if chunk.IsExpired() {
		t.Error("Chunk should not be expired")
	}

	chunk.ExpiresAt = now.Add(-time.Hour)
	if !chunk.IsExpired() {
		t.Error("Chunk should be expired")
	}
}

func TestConstants(t *testing.T) {
	t.Run("Status Constants", testStatusConstants)
	t.Run("File Constants", testFileConstants)
	t.Run("Storage Constants", testStorageConstants)
	t.Run("Access Constants", testAccessConstants)
	t.Run("Share Constants", testShareConstants)
}

func testStatusConstants(t *testing.T) {
	// Test status constants
	if StatusActive != "active" {
		t.Error("StatusActive constant mismatch")
	}
	if StatusInactive != "inactive" {
		t.Error("StatusInactive constant mismatch")
	}
}

func testFileConstants(t *testing.T) {
	// Test file status constants
	if FileStatusActive != "active" {
		t.Error("FileStatusActive constant mismatch")
	}
	if FileStatusUploading != "uploading" {
		t.Error("FileStatusUploading constant mismatch")
	}
}

func testStorageConstants(t *testing.T) {
	// Test storage type constants
	if StorageTypeLocal != "local" {
		t.Error("StorageTypeLocal constant mismatch")
	}
	if StorageTypeOSS != "oss" {
		t.Error("StorageTypeOSS constant mismatch")
	}
}

func testAccessConstants(t *testing.T) {
	// Test access level constants
	if AccessLevelPrivate != "private" {
		t.Error("AccessLevelPrivate constant mismatch")
	}
	if AccessLevelPublic != "public" {
		t.Error("AccessLevelPublic constant mismatch")
	}
}

func testShareConstants(t *testing.T) {
	// Test share permission constants
	if SharePermissionView != "view" {
		t.Error("SharePermissionView constant mismatch")
	}
	if SharePermissionDownload != "download" {
		t.Error("SharePermissionDownload constant mismatch")
	}
}

func TestPreferenceConstants(t *testing.T) {
	// Test preference category constants
	if PreferenceCategoryUI != "ui" {
		t.Error("PreferenceCategoryUI constant mismatch")
	}
	if PreferenceCategoryFile != "file" {
		t.Error("PreferenceCategoryFile constant mismatch")
	}

	// Test preference key constants
	if PreferenceKeyTheme != "theme" {
		t.Error("PreferenceKeyTheme constant mismatch")
	}
	if PreferenceKeyLanguage != "language" {
		t.Error("PreferenceKeyLanguage constant mismatch")
	}
}

func TestModelInterfaces(t *testing.T) {
	t.Run("TableNamer interface", func(t *testing.T) {
		models := []TableNamer{
			&User{},
			&UserSession{},
			&UserLoginHistory{},
			&UserPreference{},
			&File{},
			&FileVersion{},
			&FileShare{},
			&FileTag{},
			&FileUploadChunk{},
		}

		for _, model := range models {
			tableName := model.TableName()
			if tableName == "" {
				t.Errorf("Model %T should have a table name", model)
			}
		}
	})

	t.Run("VersionedModel interface", func(t *testing.T) {
		user := &User{}

		// Test initial version
		if user.GetVersion() != 0 {
			t.Error("Initial version should be 0")
		}

		// Test setting version
		user.SetVersion(5)
		if user.GetVersion() != 5 {
			t.Error("Version should be set to 5")
		}
	})
}

func TestBaseModelMethods(t *testing.T) {
	t.Run("User with BaseModel", func(t *testing.T) {
		user := &User{}

		// Test version methods
		if user.GetVersion() != 0 {
			t.Error("Initial version should be 0")
		}

		user.SetVersion(10)
		if user.GetVersion() != 10 {
			t.Error("Version should be updated")
		}

		// Test deletion status (without database)
		if user.IsDeleted() {
			t.Error("New user should not be marked as deleted")
		}

		// Test deletion time
		if user.GetDeletedAt() != nil {
			t.Error("New user should not have deletion time")
		}
	})
}
