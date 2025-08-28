package models

import (
	"crypto/md5" // #nosec G501 - 仅用于文件校验，非安全用途
	"crypto/rand"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
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

// 通用数据库索引名称常量
const (
	IndexTypeUnique    = "unique"
	IndexTypeComposite = "composite"
	IndexTypePartial   = "partial"
	IndexTypeGIN       = "gin"
	IndexTypeBTree     = "btree"
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

// GetSort 获取排序权重
func (m *SortModel) GetSort() int {
	return m.Sort
}

// SetSort 设置排序权重
func (m *SortModel) SetSort(sort int) {
	m.Sort = sort
}

// TrackableModel 可追踪模型，记录创建者和更新者
type TrackableModel struct {
	BaseModel
	CreatedBy *uint `json:"created_by,omitempty"` // 创建者ID
	UpdatedBy *uint `json:"updated_by,omitempty"` // 更新者ID
}

// SetCreatedBy 设置创建者
func (m *TrackableModel) SetCreatedBy(userID uint) {
	m.CreatedBy = &userID
}

// SetUpdatedBy 设置更新者
func (m *TrackableModel) SetUpdatedBy(userID uint) {
	m.UpdatedBy = &userID
}

// BeforeCreate GORM钩子：创建前
func (m *TrackableModel) BeforeCreate(tx *gorm.DB) error {
	if m.Version == 0 {
		m.Version = 1
	}
	// 如果没有设置创建者，尝试从上下文获取
	if m.CreatedBy == nil {
		if userID, ok := tx.Get("current_user_id"); ok {
			if uid, valid := userID.(uint); valid {
				m.CreatedBy = &uid
			}
		}
	}
	m.UpdatedBy = m.CreatedBy
	return nil
}

// BeforeUpdate GORM钩子：更新前
func (m *TrackableModel) BeforeUpdate(tx *gorm.DB) error {
	m.Version++
	// 如果没有设置更新者，尝试从上下文获取
	if m.UpdatedBy == nil {
		if userID, ok := tx.Get("current_user_id"); ok {
			if uid, valid := userID.(uint); valid {
				m.UpdatedBy = &uid
			}
		}
	}
	return nil
}

// JSONMap 自定义JSON类型
type JSONMap map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	return json.Unmarshal(bytes, j)
}

// 辅助函数

// GenerateUUID 生成UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateShareCode 生成分享码
func GenerateShareCode() string {
	return GenerateRandomString(8)
}

// GenerateInviteCode 生成邀请码
func GenerateInviteCode() string {
	return GenerateRandomString(12)
}

// GenerateResetToken 生成重置令牌
func GenerateResetToken() string {
	return GenerateRandomString(32)
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// HashToken 对令牌进行哈希
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// HashMD5 计算MD5哈希（仅用于文件校验，非安全用途）
func HashMD5(data string) string {
	// #nosec G401 - MD5仅用于文件校验，不用于安全用途
	h := md5.Sum([]byte(data))
	return hex.EncodeToString(h[:])
}

// GenerateSalt 生成盐值
func GenerateSalt() string {
	return GenerateRandomString(16)
}

// HashWithSalt 使用盐值进行哈希
func HashWithSalt(data, salt string) string {
	h := sha256.Sum256([]byte(data + salt))
	return hex.EncodeToString(h[:])
}

// VerifyHashWithSalt 验证哈希值
func VerifyHashWithSalt(data, hash, salt string) bool {
	return HashWithSalt(data, salt) == hash
}

// GenerateSessionID 生成会话ID
func GenerateSessionID() string {
	return GenerateRandomString(64)
}

// GenerateVerificationCode 生成验证码
func GenerateVerificationCode() string {
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code[i] = byte('0' + n.Int64())
	}
	return string(code)
}

// GenerateNumericCode 生成数字验证码
func GenerateNumericCode(length int) string {
	code := make([]byte, length)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code[i] = byte('0' + n.Int64())
	}
	return string(code)
}

// NormalizeEmail 规范化邮箱地址
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// FormatFloat 格式化浮点数
func FormatFloat(value float64, precision int) string {
	return fmt.Sprintf("%."+fmt.Sprintf("%d", precision)+"f", value)
}

// SplitString 分割字符串
func SplitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}

// JoinString 连接字符串
func JoinString(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// ModelRegistry 模型注册表
type ModelRegistry struct {
	models []interface{}
}

// NewModelRegistry 创建新的模型注册表
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		models: make([]interface{}, 0),
	}
}

// Register 注册模型
func (r *ModelRegistry) Register(model interface{}) {
	r.models = append(r.models, model)
}

// GetModels 获取所有注册的模型
func (r *ModelRegistry) GetModels() []interface{} {
	return r.models
}
