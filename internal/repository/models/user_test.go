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
type UserTest struct {
	basemodels.BaseModel
	UUID         string  `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	Email        string  `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Username     string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	PasswordHash string  `gorm:"type:varchar(255);not null" json:"-"`
	Phone        *string `gorm:"type:varchar(20);index" json:"phone,omitempty"`
	AvatarURL    *string `gorm:"type:varchar(500)" json:"avatar_url,omitempty"`
	DisplayName  *string `gorm:"type:varchar(100)" json:"display_name,omitempty"`

	Status          string     `gorm:"type:varchar(20);default:'active';index" json:"status"`
	EmailVerified   bool       `gorm:"default:false" json:"email_verified"`
	PhoneVerified   bool       `gorm:"default:false" json:"phone_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty"`

	StorageQuota int64 `gorm:"default:10737418240" json:"storage_quota"`
	StorageUsed  int64 `gorm:"default:0" json:"storage_used"`

	MFAEnabled     bool    `gorm:"default:false" json:"mfa_enabled"`
	MFASecret      *string `gorm:"type:varchar(255)" json:"-"`
	MFAType        string  `gorm:"type:varchar(20);default:'totp'" json:"mfa_type"`
	MFABackupCodes *string `gorm:"type:text" json:"-"`

	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP       *string    `gorm:"type:varchar(45)" json:"last_login_ip,omitempty"`
	PasswordUpdatedAt *time.Time `json:"password_updated_at,omitempty"`

	Profile  *basemodels.JSONMap `gorm:"type:text" json:"profile,omitempty"`
	Settings *basemodels.JSONMap `gorm:"type:text" json:"settings,omitempty"`
}

func (UserTest) TableName() string {
	return "users"
}

func (u *UserTest) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		u.UUID = basemodels.GenerateUUID()
	}
	if u.PasswordUpdatedAt == nil {
		now := time.Now()
		u.PasswordUpdatedAt = &now
	}
	return u.BaseModel.BeforeCreate(tx)
}

func (u *UserTest) IsActive() bool {
	return u.Status == "active"
}

func (u *UserTest) IsSuspended() bool {
	return u.Status == "suspended"
}

func (u *UserTest) GetStorageUsagePercent() float64 {
	if u.StorageQuota == 0 {
		return 0
	}
	return float64(u.StorageUsed) / float64(u.StorageQuota) * 100
}

func (u *UserTest) HasStorageSpace(size int64) bool {
	return u.StorageUsed+size <= u.StorageQuota
}

type UserSessionTest struct {
	basemodels.BaseModel
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	SessionToken   string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"session_token"`
	RefreshToken   *string    `gorm:"type:varchar(255);index" json:"refresh_token,omitempty"`
	DeviceInfo     *string    `gorm:"type:varchar(500)" json:"device_info,omitempty"`
	UserAgent      *string    `gorm:"type:varchar(1000)" json:"user_agent,omitempty"`
	IPAddress      *string    `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	Location       *string    `gorm:"type:varchar(200)" json:"location,omitempty"`
	ExpiresAt      time.Time  `gorm:"not null;index" json:"expires_at"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
}

func (UserSessionTest) TableName() string {
	return "user_sessions"
}

func (s *UserSessionTest) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *UserSessionTest) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

type UserLoginHistoryTest struct {
	basemodels.BaseModel
	UserID      uint    `gorm:"not null;index" json:"user_id"`
	IPAddress   string  `gorm:"type:varchar(45);not null" json:"ip_address"`
	UserAgent   *string `gorm:"type:varchar(1000)" json:"user_agent,omitempty"`
	DeviceInfo  *string `gorm:"type:varchar(500)" json:"device_info,omitempty"`
	Location    *string `gorm:"type:varchar(200)" json:"location,omitempty"`
	LoginMethod string  `gorm:"type:varchar(20);default:'password'" json:"login_method"`
	Status      string  `gorm:"type:varchar(20);default:'success'" json:"status"`
	FailReason  *string `gorm:"type:varchar(255)" json:"fail_reason,omitempty"`
	SessionID   *uint   `gorm:"index" json:"session_id,omitempty"`
}

func (UserLoginHistoryTest) TableName() string {
	return "user_login_history"
}

func (h *UserLoginHistoryTest) IsSuccessful() bool {
	return h.Status == "success"
}

type UserPreferenceTest struct {
	basemodels.BaseModel
	UserID      uint    `gorm:"not null;index" json:"user_id"`
	Category    string  `gorm:"type:varchar(100);not null" json:"category"`
	Key         string  `gorm:"type:varchar(100);not null" json:"key"`
	Value       *string `gorm:"type:text" json:"value,omitempty"`
	ValueType   string  `gorm:"type:varchar(20);default:'string'" json:"value_type"`
	Description *string `gorm:"type:varchar(255)" json:"description,omitempty"`
	IsPublic    bool    `gorm:"default:false" json:"is_public"`
}

func (UserPreferenceTest) TableName() string {
	return "user_preferences"
}

func (p *UserPreferenceTest) BeforeCreate(tx *gorm.DB) error {
	var count int64
	tx.Model(&UserPreferenceTest{}).Where("user_id = ? AND category = ? AND key = ?",
		p.UserID, p.Category, p.Key).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return p.BaseModel.BeforeCreate(tx)
}

func (p *UserPreferenceTest) GetBoolValue() bool {
	if p.Value == nil || p.ValueType != "boolean" {
		return false
	}
	return *p.Value == "true"
}

func (p *UserPreferenceTest) GetStringValue() string {
	if p.Value == nil {
		return ""
	}
	return *p.Value
}

// setupTestDB 设置测试数据库
func setupTestDB() (*gorm.DB, error) {
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
		&UserSessionTest{},
		&UserLoginHistoryTest{},
		&UserPreferenceTest{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestUser_TableName(t *testing.T) {
	user := &UserTest{}
	if user.TableName() != "users" {
		t.Errorf("Expected table name 'users', got '%s'", user.TableName())
	}
}

func TestUser_BeforeCreate(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	user := &UserTest{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
	}

	// 测试创建前UUID生成
	if err := user.BeforeCreate(db); err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}

	if user.UUID == "" {
		t.Error("UUID should be generated in BeforeCreate")
	}

	if user.PasswordUpdatedAt == nil {
		t.Error("PasswordUpdatedAt should be set in BeforeCreate")
	}
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"active", true},
		{"inactive", false},
		{"suspended", false},
		{"deleted", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			user := &UserTest{Status: tt.status}
			if user.IsActive() != tt.expected {
				t.Errorf("IsActive() for status '%s' = %v, want %v", tt.status, user.IsActive(), tt.expected)
			}
		})
	}
}

func TestUser_IsSuspended(t *testing.T) {
	user := &UserTest{Status: "suspended"}
	if !user.IsSuspended() {
		t.Error("Expected user to be suspended")
	}

	user.Status = "active"
	if user.IsSuspended() {
		t.Error("Expected user not to be suspended")
	}
}

func TestUser_StorageMethods(t *testing.T) {
	user := &UserTest{
		StorageQuota: 1000,
		StorageUsed:  300,
	}

	// Test GetStorageUsagePercent
	expectedPercent := 30.0
	if percent := user.GetStorageUsagePercent(); percent != expectedPercent {
		t.Errorf("GetStorageUsagePercent() = %v, want %v", percent, expectedPercent)
	}

	// Test HasStorageSpace
	if !user.HasStorageSpace(500) {
		t.Error("Should have enough storage space for 500 bytes")
	}

	if user.HasStorageSpace(800) {
		t.Error("Should not have enough storage space for 800 bytes")
	}

	// Test with zero quota
	user.StorageQuota = 0
	if user.GetStorageUsagePercent() != 0 {
		t.Error("Should return 0% when quota is 0")
	}
}

func TestUserSession_TableName(t *testing.T) {
	session := &UserSessionTest{}
	if session.TableName() != "user_sessions" {
		t.Errorf("Expected table name 'user_sessions', got '%s'", session.TableName())
	}
}

func TestUserSession_IsExpired(t *testing.T) {
	now := time.Now()

	// Test expired session
	expiredSession := &UserSessionTest{
		ExpiresAt: now.Add(-time.Hour),
	}
	if !expiredSession.IsExpired() {
		t.Error("Session should be expired")
	}

	// Test valid session
	validSession := &UserSessionTest{
		ExpiresAt: now.Add(time.Hour),
	}
	if validSession.IsExpired() {
		t.Error("Session should not be expired")
	}
}

func TestUserSession_IsValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		session  *UserSessionTest
		expected bool
	}{
		{
			name: "valid and active session",
			session: &UserSessionTest{
				IsActive:  true,
				ExpiresAt: now.Add(time.Hour),
			},
			expected: true,
		},
		{
			name: "inactive session",
			session: &UserSessionTest{
				IsActive:  false,
				ExpiresAt: now.Add(time.Hour),
			},
			expected: false,
		},
		{
			name: "expired session",
			session: &UserSessionTest{
				IsActive:  true,
				ExpiresAt: now.Add(-time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.session.IsValid() != tt.expected {
				t.Errorf("IsValid() = %v, want %v", tt.session.IsValid(), tt.expected)
			}
		})
	}
}

func TestUserLoginHistory_TableName(t *testing.T) {
	history := &UserLoginHistoryTest{}
	if history.TableName() != "user_login_history" {
		t.Errorf("Expected table name 'user_login_history', got '%s'", history.TableName())
	}
}

func TestUserLoginHistory_IsSuccessful(t *testing.T) {
	successHistory := &UserLoginHistoryTest{Status: "success"}
	if !successHistory.IsSuccessful() {
		t.Error("Login history should be successful")
	}

	failedHistory := &UserLoginHistoryTest{Status: "failed"}
	if failedHistory.IsSuccessful() {
		t.Error("Login history should not be successful")
	}
}

func TestUserPreference_TableName(t *testing.T) {
	pref := &UserPreferenceTest{}
	if pref.TableName() != "user_preferences" {
		t.Errorf("Expected table name 'user_preferences', got '%s'", pref.TableName())
	}
}

func TestUserPreference_ValueMethods(t *testing.T) {
	// Test GetBoolValue
	boolValue := "true"
	pref := &UserPreferenceTest{
		Value:     &boolValue,
		ValueType: "boolean",
	}
	if !pref.GetBoolValue() {
		t.Error("Expected GetBoolValue to return true")
	}

	// Test GetStringValue
	stringValue := "test value"
	stringPref := &UserPreferenceTest{
		Value: &stringValue,
	}
	if pref := stringPref.GetStringValue(); pref != "test value" {
		t.Errorf("Expected GetStringValue to return 'test value', got '%s'", pref)
	}

	// Test with nil value
	nilPref := &UserPreferenceTest{}
	if pref := nilPref.GetStringValue(); pref != "" {
		t.Errorf("Expected GetStringValue to return empty string for nil value, got '%s'", pref)
	}
}

func TestUserPreference_BeforeCreate(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create first preference
	pref1 := &UserPreferenceTest{
		UserID:    1,
		Category:  "ui",
		Key:       "theme",
		ValueType: "string",
	}

	if err := pref1.BeforeCreate(db); err != nil {
		t.Fatalf("First preference creation should succeed: %v", err)
	}

	// Actually save it to database
	result := db.Create(pref1)
	if result.Error != nil {
		t.Fatalf("Failed to save first preference: %v", result.Error)
	}

	// Try to create another preference with same user, category, and key
	pref2 := &UserPreferenceTest{
		UserID:    1,
		Category:  "ui",
		Key:       "theme",
		ValueType: "string",
	}

	if err := pref2.BeforeCreate(db); err == nil {
		t.Error("Should fail to create duplicate preference")
	}
}

func TestCreateUserWithDatabase(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	user := &UserTest{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Status:       "active",
	}

	// Test creating user in database
	result := db.Create(user)
	if result.Error != nil {
		t.Fatalf("Failed to create user: %v", result.Error)
	}

	// Verify user was created with UUID
	if user.UUID == "" {
		t.Error("User UUID should be generated")
	}

	if user.ID == 0 {
		t.Error("User ID should be assigned")
	}

	// Test retrieving user
	var retrievedUser UserTest
	result = db.First(&retrievedUser, user.ID)
	if result.Error != nil {
		t.Fatalf("Failed to retrieve user: %v", result.Error)
	}

	if retrievedUser.Email != user.Email {
		t.Errorf("Retrieved user email = %v, want %v", retrievedUser.Email, user.Email)
	}

	if !retrievedUser.IsActive() {
		t.Error("Retrieved user should be active")
	}
}
