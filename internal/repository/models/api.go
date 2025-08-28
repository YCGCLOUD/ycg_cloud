package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// APIApp API应用表结构
type APIApp struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`  // 应用唯一标识符
	UserID uint   `gorm:"not null;index" json:"user_id"`                   // 应用所有者用户ID
	Name   string `gorm:"type:varchar(255);not null" json:"name"`          // 应用名称
	AppID  string `gorm:"type:varchar(100);not null;unique" json:"app_id"` // 应用ID

	// 应用信息
	Description *string `gorm:"type:text" json:"description,omitempty"`                            // 应用描述
	Website     *string `gorm:"type:varchar(500)" json:"website,omitempty"`                        // 应用网站
	Logo        *string `gorm:"type:varchar(500)" json:"logo,omitempty"`                           // 应用Logo
	Category    string  `gorm:"type:varchar(100)" json:"category"`                                 // 应用分类
	Type        string  `gorm:"type:enum('web','mobile','desktop','server');not null" json:"type"` // 应用类型

	// 认证信息
	AppKey    string `gorm:"type:varchar(255);not null;unique" json:"app_key"` // 应用密钥
	AppSecret string `gorm:"type:varchar(255);not null" json:"-"`              // 应用秘钥（加密存储）

	// 状态信息
	Status       string     `gorm:"type:enum('pending','approved','rejected','suspended','deleted');default:'pending'" json:"status"` // 审核状态
	IsActive     bool       `gorm:"default:false" json:"is_active"`                                                                   // 是否激活
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`                                                                            // 审核通过时间
	ApprovedBy   *uint      `json:"approved_by,omitempty"`                                                                            // 审核人ID
	RejectedAt   *time.Time `json:"rejected_at,omitempty"`                                                                            // 拒绝时间
	RejectReason *string    `gorm:"type:text" json:"reject_reason,omitempty"`                                                         // 拒绝原因

	// 权限配置
	Permissions *basemodels.JSONMap `gorm:"type:json" json:"permissions,omitempty"` // 权限列表
	Scopes      *string             `gorm:"type:text" json:"scopes,omitempty"`      // 授权范围

	// 限流配置
	RateLimit       int `gorm:"default:1000" json:"rate_limit"`      // 每分钟请求限制
	DailyLimit      int `gorm:"default:10000" json:"daily_limit"`    // 每日请求限制
	MonthlyLimit    int `gorm:"default:100000" json:"monthly_limit"` // 每月请求限制
	ConcurrentLimit int `gorm:"default:10" json:"concurrent_limit"`  // 并发请求限制

	// 回调配置
	CallbackURLs   *string `gorm:"type:text" json:"callback_urls,omitempty"`   // 回调URL列表
	AllowedOrigins *string `gorm:"type:text" json:"allowed_origins,omitempty"` // 允许的来源
	AllowedIPs     *string `gorm:"type:text" json:"allowed_ips,omitempty"`     // 允许的IP地址

	// 统计信息
	TotalRequests   int64      `gorm:"default:0" json:"total_requests"`   // 总请求数
	SuccessRequests int64      `gorm:"default:0" json:"success_requests"` // 成功请求数
	FailedRequests  int64      `gorm:"default:0" json:"failed_requests"`  // 失败请求数
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`            // 最后使用时间

	// 关联关系
	Owner    User       `gorm:"foreignKey:UserID" json:"owner,omitempty"`
	Approver *User      `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
	Tokens   []APIToken `gorm:"foreignKey:AppID;references:AppID" json:"tokens,omitempty"`
	Logs     []APILog   `gorm:"foreignKey:AppID;references:AppID" json:"logs,omitempty"`
	Webhooks []Webhook  `gorm:"foreignKey:AppID;references:AppID" json:"webhooks,omitempty"`
}

// TableName API应用表名
func (APIApp) TableName() string {
	return "api_apps"
}

// BeforeCreate 创建前钩子
func (app *APIApp) BeforeCreate(tx *gorm.DB) error {
	if app.UUID == "" {
		app.UUID = basemodels.GenerateUUID()
	}
	if app.AppID == "" {
		app.AppID = basemodels.GenerateRandomString(20)
	}
	if app.AppKey == "" {
		app.AppKey = basemodels.GenerateRandomString(32)
	}
	if app.AppSecret == "" {
		app.AppSecret = basemodels.GenerateRandomString(64)
	}
	return app.BaseModel.BeforeCreate(tx)
}

// Approve 审核通过
func (app *APIApp) Approve(approverID uint) {
	app.Status = "approved"
	app.IsActive = true
	app.ApprovedBy = &approverID
	now := time.Now()
	app.ApprovedAt = &now
}

// Reject 拒绝审核
func (app *APIApp) Reject(reason string) {
	app.Status = "rejected"
	app.IsActive = false
	app.RejectReason = &reason
	now := time.Now()
	app.RejectedAt = &now
}

// UpdateUsage 更新使用统计
func (app *APIApp) UpdateUsage(success bool) {
	app.TotalRequests++
	if success {
		app.SuccessRequests++
	} else {
		app.FailedRequests++
	}
	now := time.Now()
	app.LastUsedAt = &now
}

// APIToken API访问令牌表结构
type APIToken struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 令牌唯一标识符
	AppID  string `gorm:"type:varchar(100);not null;index" json:"app_id"` // 应用ID
	UserID uint   `gorm:"not null;index" json:"user_id"`                  // 用户ID
	Name   string `gorm:"type:varchar(255);not null" json:"name"`         // 令牌名称

	// 令牌信息
	Token        string  `gorm:"type:varchar(255);not null;unique" json:"token"`   // 访问令牌
	TokenHash    string  `gorm:"type:varchar(255);not null" json:"-"`              // 令牌哈希
	RefreshToken *string `gorm:"type:varchar(255)" json:"refresh_token,omitempty"` // 刷新令牌

	// 权限配置
	Scopes      *string             `gorm:"type:text" json:"scopes,omitempty"`      // 授权范围
	Permissions *basemodels.JSONMap `gorm:"type:json" json:"permissions,omitempty"` // 权限配置

	// 状态信息
	IsActive   bool       `gorm:"default:true" json:"is_active"` // 是否激活
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`          // 过期时间
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`        // 最后使用时间

	// 使用限制
	UsageCount int64  `gorm:"default:0" json:"usage_count"` // 使用次数
	UsageLimit *int64 `json:"usage_limit,omitempty"`        // 使用限制

	// IP限制
	AllowedIPs *string `gorm:"type:text" json:"allowed_ips,omitempty"` // 允许的IP地址

	// 关联关系
	App  *APIApp `gorm:"foreignKey:AppID;references:AppID" json:"app,omitempty"`
	User User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName API访问令牌表名
func (APIToken) TableName() string {
	return "api_tokens"
}

// BeforeCreate 创建前钩子
func (token *APIToken) BeforeCreate(tx *gorm.DB) error {
	if token.UUID == "" {
		token.UUID = basemodels.GenerateUUID()
	}
	if token.Token == "" {
		token.Token = basemodels.GenerateRandomString(64)
	}
	token.TokenHash = basemodels.HashToken(token.Token)
	return token.BaseModel.BeforeCreate(tx)
}

// IsExpired 检查是否过期
func (token *APIToken) IsExpired() bool {
	if token.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*token.ExpiresAt)
}

// IsValid 检查令牌是否有效
func (token *APIToken) IsValid() bool {
	if !token.IsActive || token.IsExpired() {
		return false
	}
	if token.UsageLimit != nil && token.UsageCount >= *token.UsageLimit {
		return false
	}
	return true
}

// IncrementUsage 增加使用次数
func (token *APIToken) IncrementUsage() {
	token.UsageCount++
	now := time.Now()
	token.LastUsedAt = &now
}

// Webhook WebHook表结构
type Webhook struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // WebHook唯一标识符
	AppID  string `gorm:"type:varchar(100);not null;index" json:"app_id"` // 应用ID
	UserID uint   `gorm:"not null;index" json:"user_id"`                  // 用户ID
	Name   string `gorm:"type:varchar(255);not null" json:"name"`         // WebHook名称

	// URL配置
	URL    string  `gorm:"type:varchar(500);not null" json:"url"`                // 回调URL
	Secret *string `gorm:"type:varchar(255)" json:"secret,omitempty"`            // 签名秘钥
	Method string  `gorm:"type:enum('POST','PUT');default:'POST'" json:"method"` // HTTP方法

	// 事件配置
	Events      string              `gorm:"type:text;not null" json:"events"`                                 // 监听的事件列表
	Filters     *basemodels.JSONMap `gorm:"type:json" json:"filters,omitempty"`                               // 事件过滤器
	ContentType string              `gorm:"type:varchar(100);default:'application/json'" json:"content_type"` // 内容类型

	// 状态信息
	IsActive    bool       `gorm:"default:true" json:"is_active"`                 // 是否激活
	LastTrigger *time.Time `json:"last_trigger,omitempty"`                        // 最后触发时间
	LastStatus  string     `gorm:"type:varchar(50)" json:"last_status,omitempty"` // 最后状态

	// 重试配置
	RetryCount int `gorm:"default:3" json:"retry_count"`  // 重试次数
	RetryDelay int `gorm:"default:60" json:"retry_delay"` // 重试延迟（秒）
	Timeout    int `gorm:"default:30" json:"timeout"`     // 超时时间（秒）

	// 统计信息
	TotalTriggers   int64 `gorm:"default:0" json:"total_triggers"`   // 总触发次数
	SuccessTriggers int64 `gorm:"default:0" json:"success_triggers"` // 成功触发次数
	FailedTriggers  int64 `gorm:"default:0" json:"failed_triggers"`  // 失败触发次数

	// 关联关系
	App  *APIApp      `gorm:"foreignKey:AppID;references:AppID" json:"app,omitempty"`
	User User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Logs []WebhookLog `gorm:"foreignKey:WebhookID" json:"logs,omitempty"`
}

// TableName WebHook表名
func (Webhook) TableName() string {
	return "webhooks"
}

// BeforeCreate 创建前钩子
func (w *Webhook) BeforeCreate(tx *gorm.DB) error {
	if w.UUID == "" {
		w.UUID = basemodels.GenerateUUID()
	}
	return w.BaseModel.BeforeCreate(tx)
}

// UpdateTriggerStats 更新触发统计
func (w *Webhook) UpdateTriggerStats(success bool) {
	w.TotalTriggers++
	if success {
		w.SuccessTriggers++
		w.LastStatus = "success"
	} else {
		w.FailedTriggers++
		w.LastStatus = "failed"
	}
	now := time.Now()
	w.LastTrigger = &now
}

// WebhookLog WebHook日志表结构
type WebhookLog struct {
	basemodels.BaseModel
	// 基本信息
	UUID      string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 日志唯一标识符
	WebhookID uint   `gorm:"not null;index" json:"webhook_id"`               // WebHook ID
	Event     string `gorm:"type:varchar(100);not null" json:"event"`        // 事件类型

	// 请求信息
	RequestURL     string  `gorm:"type:varchar(500);not null" json:"request_url"`   // 请求URL
	RequestMethod  string  `gorm:"type:varchar(10);not null" json:"request_method"` // 请求方法
	RequestHeaders *string `gorm:"type:text" json:"request_headers,omitempty"`      // 请求头
	RequestBody    *string `gorm:"type:text" json:"request_body,omitempty"`         // 请求体

	// 响应信息
	ResponseStatus  int     `gorm:"default:0" json:"response_status"`            // 响应状态码
	ResponseHeaders *string `gorm:"type:text" json:"response_headers,omitempty"` // 响应头
	ResponseBody    *string `gorm:"type:text" json:"response_body,omitempty"`    // 响应体

	// 执行信息
	Status       string  `gorm:"type:enum('pending','success','failed','timeout','retry');not null" json:"status"` // 执行状态
	Duration     int64   `gorm:"default:0" json:"duration"`                                                        // 执行时长（毫秒）
	RetryCount   int     `gorm:"default:0" json:"retry_count"`                                                     // 重试次数
	ErrorMessage *string `gorm:"type:text" json:"error_message,omitempty"`                                         // 错误信息

	// 触发信息
	TriggerData *basemodels.JSONMap `gorm:"type:json" json:"trigger_data,omitempty"` // 触发数据

	// 关联关系
	Webhook *Webhook `gorm:"foreignKey:WebhookID" json:"webhook,omitempty"`
}

// TableName WebHook日志表名
func (WebhookLog) TableName() string {
	return "webhook_logs"
}

// BeforeCreate 创建前钩子
func (wl *WebhookLog) BeforeCreate(tx *gorm.DB) error {
	if wl.UUID == "" {
		wl.UUID = basemodels.GenerateUUID()
	}
	return wl.BaseModel.BeforeCreate(tx)
}

// APILog API访问日志表结构
type APILog struct {
	basemodels.BaseModel
	// 基本信息
	UUID    string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 日志唯一标识符
	AppID   string `gorm:"type:varchar(100);not null;index" json:"app_id"` // 应用ID
	UserID  *uint  `gorm:"index" json:"user_id,omitempty"`                 // 用户ID
	TokenID *uint  `gorm:"index" json:"token_id,omitempty"`                // 令牌ID

	// 请求信息
	Method    string  `gorm:"type:varchar(10);not null" json:"method"`        // 请求方法
	Path      string  `gorm:"type:varchar(500);not null" json:"path"`         // 请求路径
	Query     *string `gorm:"type:text" json:"query,omitempty"`               // 查询参数
	Headers   *string `gorm:"type:text" json:"headers,omitempty"`             // 请求头
	Body      *string `gorm:"type:text" json:"body,omitempty"`                // 请求体
	IPAddress string  `gorm:"type:varchar(45);not null" json:"ip_address"`    // IP地址
	UserAgent *string `gorm:"type:varchar(1000)" json:"user_agent,omitempty"` // 用户代理

	// 响应信息
	StatusCode   int     `gorm:"not null" json:"status_code"`              // 状态码
	ResponseSize int64   `gorm:"default:0" json:"response_size"`           // 响应大小
	Duration     int64   `gorm:"default:0" json:"duration"`                // 执行时长（毫秒）
	ErrorMessage *string `gorm:"type:text" json:"error_message,omitempty"` // 错误信息

	// 业务信息
	Action       *string `gorm:"type:varchar(100)" json:"action,omitempty"`        // 操作类型
	ResourceType *string `gorm:"type:varchar(100)" json:"resource_type,omitempty"` // 资源类型
	ResourceID   *string `gorm:"type:varchar(100)" json:"resource_id,omitempty"`   // 资源ID

	// 关联关系
	App   *APIApp   `gorm:"foreignKey:AppID;references:AppID" json:"app,omitempty"`
	User  *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Token *APIToken `gorm:"foreignKey:TokenID" json:"token,omitempty"`
}

// TableName API访问日志表名
func (APILog) TableName() string {
	return "api_logs"
}

// BeforeCreate 创建前钩子
func (al *APILog) BeforeCreate(tx *gorm.DB) error {
	if al.UUID == "" {
		al.UUID = basemodels.GenerateUUID()
	}
	return al.BaseModel.BeforeCreate(tx)
}

// IsSuccess 检查是否成功
func (al *APILog) IsSuccess() bool {
	return al.StatusCode >= 200 && al.StatusCode < 300
}

// 应用类型常量
const (
	AppTypeWeb     = "web"     // Web应用
	AppTypeMobile  = "mobile"  // 移动应用
	AppTypeDesktop = "desktop" // 桌面应用
	AppTypeServer  = "server"  // 服务端应用
)

// 应用状态常量
const (
	AppStatusPending   = "pending"   // 待审核
	AppStatusApproved  = "approved"  // 已审核
	AppStatusRejected  = "rejected"  // 已拒绝
	AppStatusSuspended = "suspended" // 已暂停
	AppStatusDeleted   = "deleted"   // 已删除
)

// WebHook事件类型常量
const (
	WebhookEventFileUpload    = "file.upload"    // 文件上传
	WebhookEventFileDownload  = "file.download"  // 文件下载
	WebhookEventFileDelete    = "file.delete"    // 文件删除
	WebhookEventFileShare     = "file.share"     // 文件分享
	WebhookEventFileComment   = "file.comment"   // 文件评论
	WebhookEventTeamJoin      = "team.join"      // 加入团队
	WebhookEventTeamLeave     = "team.leave"     // 离开团队
	WebhookEventUserRegister  = "user.register"  // 用户注册
	WebhookEventUserLogin     = "user.login"     // 用户登录
	WebhookEventStorageAlert  = "storage.alert"  // 存储警告
	WebhookEventSecurityAlert = "security.alert" // 安全警告
)
