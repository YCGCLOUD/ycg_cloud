package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// VerificationCode 验证码表结构
type VerificationCode struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 验证码唯一标识符
	Target string `gorm:"type:varchar(255);not null;index" json:"target"` // 目标（邮箱/手机号）
	Type   string `gorm:"type:varchar(50);not null;index" json:"type"`    // 验证码类型

	// 验证码信息
	Code     string `gorm:"type:varchar(20);not null" json:"-"`  // 验证码(不返回)
	CodeHash string `gorm:"type:varchar(255);not null" json:"-"` // 验证码哈希值
	Salt     string `gorm:"type:varchar(32);not null" json:"-"`  // 盐值

	// 状态信息
	IsUsed    bool       `gorm:"default:false" json:"is_used"`     // 是否已使用
	UsedAt    *time.Time `json:"used_at,omitempty"`                // 使用时间
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"` // 过期时间

	// 尝试限制
	AttemptCount int `gorm:"default:0" json:"attempt_count"` // 尝试次数
	MaxAttempts  int `gorm:"default:5" json:"max_attempts"`  // 最大尝试次数

	// 请求信息
	IPAddress string  `gorm:"type:varchar(45);not null" json:"ip_address"`    // 请求IP
	UserAgent *string `gorm:"type:varchar(1000)" json:"user_agent,omitempty"` // 用户代理

	// 关联信息
	UserID *uint `gorm:"index" json:"user_id,omitempty"` // 关联用户ID

	// 关联关系
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 验证码表名
func (VerificationCode) TableName() string {
	return "verification_codes"
}

// BeforeCreate 创建前钩子
func (v *VerificationCode) BeforeCreate(tx *gorm.DB) error {
	if v.UUID == "" {
		v.UUID = basemodels.GenerateUUID()
	}

	if v.ExpiresAt.IsZero() {
		// 默认5分钟过期
		v.ExpiresAt = time.Now().Add(5 * time.Minute)
	}

	if v.Salt == "" {
		v.Salt = basemodels.GenerateSalt()
	}

	if v.CodeHash == "" && v.Code != "" {
		v.CodeHash = basemodels.HashWithSalt(v.Code, v.Salt)
	}

	return v.BaseModel.BeforeCreate(tx)
}

// IsExpired 检查是否过期
func (v *VerificationCode) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

// IsValid 检查是否有效
func (v *VerificationCode) IsValid() bool {
	return !v.IsUsed && !v.IsExpired() && v.AttemptCount < v.MaxAttempts
}

// VerifyCode 验证验证码
func (v *VerificationCode) VerifyCode(code string) bool {
	if !v.IsValid() {
		return false
	}

	v.AttemptCount++
	return basemodels.VerifyHashWithSalt(code, v.CodeHash, v.Salt)
}

// Use 使用验证码
func (v *VerificationCode) Use() {
	v.IsUsed = true
	now := time.Now()
	v.UsedAt = &now
}

// 验证码类型常量
const (
	VerificationTypeRegister      = "register"       // 注册验证
	VerificationTypeLogin         = "login"          // 登录验证
	VerificationTypeResetPassword = "reset_password" // 重置密码
	VerificationTypeChangeEmail   = "change_email"   // 修改邮箱
	VerificationTypeChangePhone   = "change_phone"   // 修改手机
	VerificationTypeBindEmail     = "bind_email"     // 绑定邮箱
	VerificationTypeBindPhone     = "bind_phone"     // 绑定手机
	VerificationTypeMFA           = "mfa"            // 多因素认证
	VerificationTypeDelete        = "delete"         // 删除验证
	VerificationTypeFileShare     = "file_share"     // 文件分享验证
	VerificationTypeTeamInvite    = "team_invite"    // 团队邀请验证
)

// EmailTemplate 邮件模板表结构
type EmailTemplate struct {
	basemodels.BaseModel
	// 基本信息
	UUID string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 模板唯一标识符
	Type string `gorm:"type:varchar(50);not null;unique" json:"type"`   // 模板类型
	Name string `gorm:"type:varchar(100);not null" json:"name"`         // 模板名称

	// 模板内容
	Subject     string  `gorm:"type:varchar(255);not null" json:"subject"` // 邮件主题
	Content     string  `gorm:"type:text;not null" json:"content"`         // 邮件内容（HTML）
	TextContent *string `gorm:"type:text" json:"text_content,omitempty"`   // 纯文本内容

	// 模板变量
	Variables *basemodels.JSONMap `gorm:"type:json" json:"variables,omitempty"` // 模板变量定义

	// 状态信息
	IsActive bool `gorm:"default:true" json:"is_active"` // 是否启用

	// 语言设置
	Language string `gorm:"type:varchar(10);default:'zh-CN'" json:"language"` // 语言代码

	// 更新信息
	UpdatedBy *uint `json:"updated_by,omitempty"` // 更新者ID

	// 关联关系
	Updater *User `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

// TableName 邮件模板表名
func (EmailTemplate) TableName() string {
	return "email_templates"
}

// BeforeCreate 创建前钩子
func (e *EmailTemplate) BeforeCreate(tx *gorm.DB) error {
	if e.UUID == "" {
		e.UUID = basemodels.GenerateUUID()
	}
	return e.BaseModel.BeforeCreate(tx)
}

// 邮件模板类型常量
const (
	EmailTemplateRegister      = "register"       // 注册邮件
	EmailTemplateResetPassword = "reset_password" // 重置密码
	EmailTemplateChangeEmail   = "change_email"   // 修改邮箱
	EmailTemplateTeamInvite    = "team_invite"    // 团队邀请
	EmailTemplateFileShare     = "file_share"     // 文件分享
	EmailTemplateWelcome       = "welcome"        // 欢迎邮件
	EmailTemplateSecurityAlert = "security_alert" // 安全警告
	EmailTemplateStorageAlert  = "storage_alert"  // 存储警告
	EmailTemplateSystemUpdate  = "system_update"  // 系统更新
)
