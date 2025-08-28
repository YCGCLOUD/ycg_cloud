package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// AutoClassifyRule 自动分类规则表结构
type AutoClassifyRule struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 规则唯一标识符
	UserID *uint  `gorm:"index" json:"user_id,omitempty"`                 // 用户ID（全局规则为null）
	Name   string `gorm:"type:varchar(255);not null" json:"name"`         // 规则名称

	// 规则信息
	Description *string `gorm:"type:text" json:"description,omitempty"` // 规则描述
	Priority    int     `gorm:"default:0" json:"priority"`              // 优先级（数字越大优先级越高）
	IsActive    bool    `gorm:"default:true;index" json:"is_active"`    // 是否启用

	// 规则类型
	RuleType string `gorm:"type:enum('user','system','global');default:'user'" json:"rule_type"` // 规则类型
	IsSystem bool   `gorm:"default:false" json:"is_system"`                                      // 是否为系统规则

	// 触发条件
	TriggerEvent string `gorm:"type:enum('upload','create','rename','move','share');not null" json:"trigger_event"` // 触发事件

	// 匹配条件（JSON格式存储复杂条件）
	Conditions *basemodels.JSONMap `gorm:"type:json;not null" json:"conditions"` // 匹配条件

	// 执行动作（JSON格式存储多个动作）
	Actions *basemodels.JSONMap `gorm:"type:json;not null" json:"actions"` // 执行动作

	// 规则限制
	MaxExecutions *int       `json:"max_executions,omitempty"`                          // 最大执行次数
	ValidUntil    *time.Time `json:"valid_until,omitempty"`                             // 有效期至
	ApplyToPath   *string    `gorm:"type:varchar(2000)" json:"apply_to_path,omitempty"` // 应用路径

	// 统计信息
	ExecutionCount int        `gorm:"default:0" json:"execution_count"`      // 执行次数
	SuccessCount   int        `gorm:"default:0" json:"success_count"`        // 成功次数
	LastExecutedAt *time.Time `json:"last_executed_at,omitempty"`            // 最后执行时间
	LastError      *string    `gorm:"type:text" json:"last_error,omitempty"` // 最后错误信息

	// 关联关系
	User *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Logs []AutoClassifyLog `gorm:"foreignKey:RuleID" json:"logs,omitempty"`
}

// TableName 自动分类规则表名
func (AutoClassifyRule) TableName() string {
	return "auto_classify_rules"
}

// BeforeCreate 创建前钩子
func (r *AutoClassifyRule) BeforeCreate(tx *gorm.DB) error {
	if r.UUID == "" {
		r.UUID = basemodels.GenerateUUID()
	}
	return r.BaseModel.BeforeCreate(tx)
}

// IsGlobal 检查是否为全局规则
func (r *AutoClassifyRule) IsGlobal() bool {
	return r.UserID == nil
}

// IsExpired 检查是否过期
func (r *AutoClassifyRule) IsExpired() bool {
	if r.ValidUntil == nil {
		return false
	}
	return time.Now().After(*r.ValidUntil)
}

// CanExecute 检查是否可以执行
func (r *AutoClassifyRule) CanExecute() bool {
	if !r.IsActive || r.IsExpired() {
		return false
	}
	if r.MaxExecutions != nil && r.ExecutionCount >= *r.MaxExecutions {
		return false
	}
	return true
}

// IncrementExecution 增加执行次数
func (r *AutoClassifyRule) IncrementExecution(success bool) {
	r.ExecutionCount++
	if success {
		r.SuccessCount++
	}
	now := time.Now()
	r.LastExecutedAt = &now
}

// SetLastError 设置最后错误
func (r *AutoClassifyRule) SetLastError(errorMsg string) {
	r.LastError = &errorMsg
}

// AutoClassifyLog 自动分类执行日志表结构
type AutoClassifyLog struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 日志唯一标识符
	RuleID uint   `gorm:"not null;index" json:"rule_id"`                  // 规则ID
	UserID *uint  `gorm:"index" json:"user_id,omitempty"`                 // 用户ID
	FileID *uint  `gorm:"index" json:"file_id,omitempty"`                 // 文件ID

	// 执行信息
	TriggerEvent string `gorm:"type:varchar(100);not null" json:"trigger_event"`                // 触发事件
	Status       string `gorm:"type:enum('success','failed','skipped');not null" json:"status"` // 执行状态

	// 执行数据
	InputData    *basemodels.JSONMap `gorm:"type:json" json:"input_data,omitempty"`    // 输入数据
	MatchedData  *basemodels.JSONMap `gorm:"type:json" json:"matched_data,omitempty"`  // 匹配数据
	ActionResult *basemodels.JSONMap `gorm:"type:json" json:"action_result,omitempty"` // 动作结果

	// 错误信息
	ErrorMessage *string `gorm:"type:text" json:"error_message,omitempty"`     // 错误信息
	ErrorCode    *string `gorm:"type:varchar(50)" json:"error_code,omitempty"` // 错误代码

	// 执行时间
	ExecutionTime int64 `gorm:"default:0" json:"execution_time"` // 执行耗时（毫秒）

	// 关联关系
	Rule *AutoClassifyRule `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	User *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	File *File             `gorm:"foreignKey:FileID" json:"file,omitempty"`
}

// TableName 自动分类执行日志表名
func (AutoClassifyLog) TableName() string {
	return "auto_classify_logs"
}

// BeforeCreate 创建前钩子
func (l *AutoClassifyLog) BeforeCreate(tx *gorm.DB) error {
	if l.UUID == "" {
		l.UUID = basemodels.GenerateUUID()
	}
	return l.BaseModel.BeforeCreate(tx)
}

// IsSuccess 检查是否执行成功
func (l *AutoClassifyLog) IsSuccess() bool {
	return l.Status == "success"
}

// FileClassifyTemplate 文件分类模板表结构
type FileClassifyTemplate struct {
	basemodels.BaseModel
	// 基本信息
	UUID string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 模板唯一标识符
	Name string `gorm:"type:varchar(255);not null" json:"name"`         // 模板名称

	// 模板信息
	Description *string `gorm:"type:text" json:"description,omitempty"`     // 模板描述
	Category    string  `gorm:"type:varchar(100);not null" json:"category"` // 模板分类
	IsBuiltIn   bool    `gorm:"default:false" json:"is_built_in"`           // 是否内置模板
	IsActive    bool    `gorm:"default:true" json:"is_active"`              // 是否启用

	// 模板配置
	Config *basemodels.JSONMap `gorm:"type:json;not null" json:"config"` // 模板配置

	// 使用统计
	UsageCount int `gorm:"default:0" json:"usage_count"` // 使用次数

	// 创建者信息
	CreatedBy *uint `json:"created_by,omitempty"` // 创建者ID

	// 关联关系
	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName 文件分类模板表名
func (FileClassifyTemplate) TableName() string {
	return "file_classify_templates"
}

// BeforeCreate 创建前钩子
func (t *FileClassifyTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.UUID == "" {
		t.UUID = basemodels.GenerateUUID()
	}
	return t.BaseModel.BeforeCreate(tx)
}

// IncrementUsage 增加使用次数
func (t *FileClassifyTemplate) IncrementUsage() {
	t.UsageCount++
}

// 规则类型常量
const (
	RuleTypeUser   = "user"   // 用户规则
	RuleTypeSystem = "system" // 系统规则
	RuleTypeGlobal = "global" // 全局规则
)

// 触发事件常量
const (
	TriggerEventUpload = "upload" // 上传时
	TriggerEventCreate = "create" // 创建时
	TriggerEventRename = "rename" // 重命名时
	TriggerEventMove   = "move"   // 移动时
	TriggerEventShare  = "share"  // 分享时
)

// 条件类型常量
const (
	ConditionTypeFileName    = "file_name"    // 文件名
	ConditionTypeFileExt     = "file_ext"     // 文件扩展名
	ConditionTypeFileSize    = "file_size"    // 文件大小
	ConditionTypeMimeType    = "mime_type"    // MIME类型
	ConditionTypeFilePath    = "file_path"    // 文件路径
	ConditionTypeFileContent = "file_content" // 文件内容
	ConditionTypeUploadTime  = "upload_time"  // 上传时间
	ConditionTypeKeyword     = "keyword"      // 关键词
	ConditionTypeTag         = "tag"          // 标签
	ConditionTypeUser        = "user"         // 用户
	ConditionTypeStorageType = "storage_type" // 存储类型
)

// 动作类型常量
const (
	ActionTypeMoveToFolder     = "move_to_folder"    // 移动到文件夹
	ActionTypeAddTag           = "add_tag"           // 添加标签
	ActionTypeRemoveTag        = "remove_tag"        // 移除标签
	ActionTypeSetPermission    = "set_permission"    // 设置权限
	ActionTypeEncrypt          = "encrypt"           // 加密文件
	ActionTypeCompress         = "compress"          // 压缩文件
	ActionTypeChangeStorage    = "change_storage"    // 更改存储类型
	ActionTypeAddComment       = "add_comment"       // 添加评论
	ActionTypeSendNotification = "send_notification" // 发送通知
	ActionTypeCreateBackup     = "create_backup"     // 创建备份
	ActionTypeSetExpiry        = "set_expiry"        // 设置过期时间
	ActionTypeBlockUpload      = "block_upload"      // 阻止上传
)

// 模板分类常量
const (
	TemplateCategoryDocument = "document" // 文档分类
	TemplateCategoryMedia    = "media"    // 媒体分类
	TemplateCategoryCode     = "code"     // 代码分类
	TemplateCategoryArchive  = "archive"  // 压缩包分类
	TemplateCategorySecurity = "security" // 安全分类
	TemplateCategoryProject  = "project"  // 项目分类
	TemplateCategoryPersonal = "personal" // 个人分类
	TemplateCategoryWork     = "work"     // 工作分类
)
