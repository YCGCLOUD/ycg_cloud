package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// Language 语言表结构
type Language struct {
	basemodels.BaseModel
	// 基本信息
	UUID string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 语言唯一标识符
	Code string `gorm:"type:varchar(10);not null;unique" json:"code"`   // 语言代码（如zh-CN、en-US）
	Name string `gorm:"type:varchar(100);not null" json:"name"`         // 语言名称（如中文、English）

	// 语言信息
	NativeName string  `gorm:"type:varchar(100);not null" json:"native_name"`         // 本地名称（如中文、English）
	Direction  string  `gorm:"type:enum('ltr','rtl');default:'ltr'" json:"direction"` // 文字方向
	Region     *string `gorm:"type:varchar(10)" json:"region,omitempty"`              // 地区代码
	Country    *string `gorm:"type:varchar(100)" json:"country,omitempty"`            // 国家/地区

	// 状态信息
	IsActive    bool `gorm:"default:true" json:"is_active"`     // 是否启用
	IsDefault   bool `gorm:"default:false" json:"is_default"`   // 是否默认语言
	IsSystem    bool `gorm:"default:false" json:"is_system"`    // 是否系统语言
	IsCompleted bool `gorm:"default:false" json:"is_completed"` // 翻译是否完成

	// 统计信息
	TotalTexts      int     `gorm:"default:0" json:"total_texts"`      // 总文本数
	TranslatedTexts int     `gorm:"default:0" json:"translated_texts"` // 已翻译文本数
	Progress        float64 `gorm:"default:0" json:"progress"`         // 翻译进度

	// 排序和优先级
	SortOrder int `gorm:"default:0" json:"sort_order"` // 排序顺序

	// 文化设置
	DateFormat     *string `gorm:"type:varchar(50)" json:"date_format,omitempty"`     // 日期格式
	TimeFormat     *string `gorm:"type:varchar(50)" json:"time_format,omitempty"`     // 时间格式
	NumberFormat   *string `gorm:"type:varchar(50)" json:"number_format,omitempty"`   // 数字格式
	CurrencyFormat *string `gorm:"type:varchar(50)" json:"currency_format,omitempty"` // 货币格式

	// 更新信息
	LastUpdatedBy *uint      `json:"last_updated_by,omitempty"` // 最后更新者
	LastSyncAt    *time.Time `json:"last_sync_at,omitempty"`    // 最后同步时间

	// 关联关系
	LastUpdater *User          `gorm:"foreignKey:LastUpdatedBy" json:"last_updater,omitempty"`
	Texts       []LanguageText `gorm:"foreignKey:LanguageID" json:"texts,omitempty"`
}

// TableName 语言表名
func (Language) TableName() string {
	return "languages"
}

// BeforeCreate 创建前钩子
func (l *Language) BeforeCreate(tx *gorm.DB) error {
	if l.UUID == "" {
		l.UUID = basemodels.GenerateUUID()
	}
	return l.BaseModel.BeforeCreate(tx)
}

// UpdateProgress 更新翻译进度
func (l *Language) UpdateProgress() {
	if l.TotalTexts > 0 {
		l.Progress = float64(l.TranslatedTexts) / float64(l.TotalTexts) * 100
	}
}

// IsRTL 检查是否为从右到左的语言
func (l *Language) IsRTL() bool {
	return l.Direction == "rtl"
}

// GetLocale 获取完整的locale代码
func (l *Language) GetLocale() string {
	if l.Region != nil {
		return l.Code + "-" + *l.Region
	}
	return l.Code
}

// LanguageText 语言文本表结构
type LanguageText struct {
	basemodels.BaseModel
	// 基本信息
	UUID       string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 文本唯一标识符
	LanguageID uint   `gorm:"not null;index" json:"language_id"`              // 语言ID
	Key        string `gorm:"type:varchar(255);not null;index" json:"key"`    // 文本键

	// 文本内容
	Value       string  `gorm:"type:text;not null" json:"value"`              // 文本值
	Description *string `gorm:"type:text" json:"description,omitempty"`       // 描述
	Context     *string `gorm:"type:varchar(255)" json:"context,omitempty"`   // 上下文
	Namespace   *string `gorm:"type:varchar(100)" json:"namespace,omitempty"` // 命名空间

	// 翻译状态
	Status         string     `gorm:"type:enum('pending','translated','reviewed','approved');default:'pending'" json:"status"` // 翻译状态
	IsPlural       bool       `gorm:"default:false" json:"is_plural"`                                                          // 是否为复数形式
	PluralForms    *string    `gorm:"type:text" json:"plural_forms,omitempty"`                                                 // 复数形式规则
	LastReviewedAt *time.Time `json:"last_reviewed_at,omitempty"`                                                              // 最后审核时间

	// 版本信息
	Version     int    `gorm:"default:1" json:"version"`              // 版本号
	OriginalKey string `gorm:"type:varchar(255)" json:"original_key"` // 原始键（用于版本追踪）

	// 标签和分类
	Tags       *string `gorm:"type:varchar(500)" json:"tags,omitempty"`     // 标签（逗号分隔）
	Category   *string `gorm:"type:varchar(100)" json:"category,omitempty"` // 分类
	Priority   int     `gorm:"default:0" json:"priority"`                   // 优先级
	IsSystem   bool    `gorm:"default:false" json:"is_system"`              // 是否系统文本
	IsRequired bool    `gorm:"default:false" json:"is_required"`            // 是否必需

	// 更新信息
	TranslatedBy   *uint      `json:"translated_by,omitempty"`    // 翻译者
	ReviewedBy     *uint      `json:"reviewed_by,omitempty"`      // 审核者
	ApprovedBy     *uint      `json:"approved_by,omitempty"`      // 批准者
	LastModifiedBy *uint      `json:"last_modified_by,omitempty"` // 最后修改者
	LastModifiedAt *time.Time `json:"last_modified_at,omitempty"` // 最后修改时间

	// 关联关系
	Language     Language `gorm:"foreignKey:LanguageID" json:"language,omitempty"`
	Translator   *User    `gorm:"foreignKey:TranslatedBy" json:"translator,omitempty"`
	Reviewer     *User    `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
	Approver     *User    `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
	LastModifier *User    `gorm:"foreignKey:LastModifiedBy" json:"last_modifier,omitempty"`
}

// TableName 语言文本表名
func (LanguageText) TableName() string {
	return "language_texts"
}

// BeforeCreate 创建前钩子
func (lt *LanguageText) BeforeCreate(tx *gorm.DB) error {
	if lt.UUID == "" {
		lt.UUID = basemodels.GenerateUUID()
	}

	// 确保同一语言下的键唯一
	var count int64
	tx.Model(&LanguageText{}).Where("language_id = ? AND key = ?",
		lt.LanguageID, lt.Key).Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}

	return lt.BaseModel.BeforeCreate(tx)
}

// BeforeUpdate 更新前钩子
func (lt *LanguageText) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	lt.LastModifiedAt = &now
	lt.Version++
	return nil
}

// MarkAsTranslated 标记为已翻译
func (lt *LanguageText) MarkAsTranslated(userID uint) {
	lt.Status = "translated"
	lt.TranslatedBy = &userID
	now := time.Now()
	lt.LastModifiedAt = &now
}

// MarkAsReviewed 标记为已审核
func (lt *LanguageText) MarkAsReviewed(userID uint) {
	lt.Status = "reviewed"
	lt.ReviewedBy = &userID
	now := time.Now()
	lt.LastReviewedAt = &now
	lt.LastModifiedAt = &now
}

// MarkAsApproved 标记为已批准
func (lt *LanguageText) MarkAsApproved(userID uint) {
	lt.Status = "approved"
	lt.ApprovedBy = &userID
	now := time.Now()
	lt.LastModifiedAt = &now
}

// IsTranslated 检查是否已翻译
func (lt *LanguageText) IsTranslated() bool {
	return lt.Status != "pending"
}

// GetTags 获取标签列表
func (lt *LanguageText) GetTags() []string {
	if lt.Tags == nil || *lt.Tags == "" {
		return []string{}
	}
	return basemodels.SplitString(*lt.Tags, ",")
}

// SetTags 设置标签
func (lt *LanguageText) SetTags(tags []string) {
	tagStr := basemodels.JoinString(tags, ",")
	lt.Tags = &tagStr
}

// 语言代码常量
const (
	LanguageCodeZhCN = "zh-CN" // 简体中文
	LanguageCodeZhTW = "zh-TW" // 繁体中文
	LanguageCodeEnUS = "en-US" // 美式英语
	LanguageCodeEnGB = "en-GB" // 英式英语
	LanguageCodeJaJP = "ja-JP" // 日语
	LanguageCodeKoKR = "ko-KR" // 韩语
	LanguageCodeFrFR = "fr-FR" // 法语
	LanguageCodeDeDE = "de-DE" // 德语
	LanguageCodeEsES = "es-ES" // 西班牙语
	LanguageCodeItIT = "it-IT" // 意大利语
	LanguageCodePtBR = "pt-BR" // 巴西葡萄牙语
	LanguageCodeRuRU = "ru-RU" // 俄语
	LanguageCodeArSA = "ar-SA" // 阿拉伯语
	LanguageCodeHiIN = "hi-IN" // 印地语
	LanguageCodeThTH = "th-TH" // 泰语
	LanguageCodeViVN = "vi-VN" // 越南语
)

// 翻译状态常量
const (
	TextStatusPending    = "pending"    // 待翻译
	TextStatusTranslated = "translated" // 已翻译
	TextStatusReviewed   = "reviewed"   // 已审核
	TextStatusApproved   = "approved"   // 已批准
)

// 文本分类常量
const (
	TextCategoryUI         = "ui"         // 用户界面
	TextCategoryMenu       = "menu"       // 菜单
	TextCategoryButton     = "button"     // 按钮
	TextCategoryMessage    = "message"    // 消息
	TextCategoryError      = "error"      // 错误信息
	TextCategoryWarning    = "warning"    // 警告信息
	TextCategorySuccess    = "success"    // 成功信息
	TextCategoryEmail      = "email"      // 邮件模板
	TextCategoryHelp       = "help"       // 帮助文档
	TextCategorySystem     = "system"     // 系统信息
	TextCategoryValidation = "validation" // 验证信息
)

// 文本命名空间常量
const (
	NamespaceAuth     = "auth"     // 认证模块
	NamespaceFile     = "file"     // 文件模块
	NamespaceTeam     = "team"     // 团队模块
	NamespaceMessage  = "message"  // 消息模块
	NamespaceSystem   = "system"   // 系统模块
	NamespaceAdmin    = "admin"    // 管理模块
	NamespaceSettings = "settings" // 设置模块
	NamespaceCommon   = "common"   // 通用模块
)
