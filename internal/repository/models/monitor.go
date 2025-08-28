package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// SystemMetric 系统监控指标表结构
type SystemMetric struct {
	basemodels.BaseModel
	// 基本信息
	UUID     string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`    // 指标唯一标识符
	NodeName string `gorm:"type:varchar(100);not null;index" json:"node_name"` // 节点名称
	NodeIP   string `gorm:"type:varchar(45);not null;index" json:"node_ip"`    // 节点IP

	// 指标信息
	MetricType  string  `gorm:"type:varchar(50);not null;index" json:"metric_type"`  // 指标类型
	MetricName  string  `gorm:"type:varchar(100);not null;index" json:"metric_name"` // 指标名称
	MetricValue float64 `gorm:"not null" json:"metric_value"`                        // 指标值
	MetricUnit  string  `gorm:"type:varchar(20)" json:"metric_unit"`                 // 指标单位

	// 时间戳
	Timestamp   time.Time `gorm:"not null;index" json:"timestamp"`    // 数据时间戳
	CollectedAt time.Time `gorm:"not null;index" json:"collected_at"` // 采集时间

	// 标签和维度
	Labels     *basemodels.JSONMap `gorm:"type:json" json:"labels,omitempty"`     // 标签信息
	Dimensions *basemodels.JSONMap `gorm:"type:json" json:"dimensions,omitempty"` // 维度信息

	// 状态信息
	Status     string   `gorm:"type:enum('normal','warning','critical','unknown');default:'normal'" json:"status"` // 状态
	Threshold  *float64 `json:"threshold,omitempty"`                                                               // 阈值
	IsAlert    bool     `gorm:"default:false" json:"is_alert"`                                                     // 是否告警
	AlertLevel *string  `gorm:"type:varchar(20)" json:"alert_level,omitempty"`                                     // 告警级别

	// 聚合信息
	AggregationType   *string  `gorm:"type:varchar(20)" json:"aggregation_type,omitempty"` // 聚合类型
	AggregationPeriod *int     `json:"aggregation_period,omitempty"`                       // 聚合周期（秒）
	MinValue          *float64 `json:"min_value,omitempty"`                                // 最小值
	MaxValue          *float64 `json:"max_value,omitempty"`                                // 最大值
	AvgValue          *float64 `json:"avg_value,omitempty"`                                // 平均值
	SampleCount       *int64   `json:"sample_count,omitempty"`                             // 样本数

	// 关联关系
	AlertRules []AlertRule `gorm:"many2many:metric_alert_rules;" json:"alert_rules,omitempty"`
}

// TableName 系统监控指标表名
func (SystemMetric) TableName() string {
	return "system_metrics"
}

// BeforeCreate 创建前钩子
func (sm *SystemMetric) BeforeCreate(tx *gorm.DB) error {
	if sm.UUID == "" {
		sm.UUID = basemodels.GenerateUUID()
	}

	if sm.Timestamp.IsZero() {
		sm.Timestamp = time.Now()
	}

	if sm.CollectedAt.IsZero() {
		sm.CollectedAt = time.Now()
	}

	return sm.BaseModel.BeforeCreate(tx)
}

// IsAbnormal 检查指标是否异常
func (sm *SystemMetric) IsAbnormal() bool {
	return sm.Status == "warning" || sm.Status == "critical"
}

// CheckThreshold 检查是否超过阈值
func (sm *SystemMetric) CheckThreshold() bool {
	if sm.Threshold == nil {
		return false
	}
	return sm.MetricValue > *sm.Threshold
}

// GetFormattedValue 获取格式化的值
func (sm *SystemMetric) GetFormattedValue() string {
	if sm.MetricUnit != "" {
		return basemodels.FormatFloat(sm.MetricValue, 2) + " " + sm.MetricUnit
	}
	return basemodels.FormatFloat(sm.MetricValue, 2)
}

// AlertRule 告警规则表结构
type AlertRule struct {
	basemodels.BaseModel
	// 基本信息
	UUID string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 规则唯一标识符
	Name string `gorm:"type:varchar(255);not null" json:"name"`         // 规则名称

	// 规则配置
	Description *string `gorm:"type:text" json:"description,omitempty"`             // 规则描述
	MetricType  string  `gorm:"type:varchar(50);not null;index" json:"metric_type"` // 监控指标类型
	MetricName  *string `gorm:"type:varchar(100)" json:"metric_name,omitempty"`     // 指标名称

	// 条件配置
	Operator  string  `gorm:"type:enum('gt','gte','lt','lte','eq','ne');not null" json:"operator"` // 操作符
	Threshold float64 `gorm:"not null" json:"threshold"`                                           // 阈值
	Duration  int     `gorm:"default:300" json:"duration"`                                         // 持续时间（秒）

	// 告警配置
	Severity      string `gorm:"type:enum('info','warning','critical','emergency');default:'warning'" json:"severity"` // 严重程度
	IsEnabled     bool   `gorm:"default:true" json:"is_enabled"`                                                       // 是否启用
	IsAutoResolve bool   `gorm:"default:true" json:"is_auto_resolve"`                                                  // 是否自动恢复

	// 过滤条件
	NodeFilter  *string             `gorm:"type:varchar(255)" json:"node_filter,omitempty"` // 节点过滤
	LabelFilter *basemodels.JSONMap `gorm:"type:json" json:"label_filter,omitempty"`        // 标签过滤
	TimeFilter  *string             `gorm:"type:varchar(100)" json:"time_filter,omitempty"` // 时间过滤

	// 通知配置
	NotifyChannels *string `gorm:"type:varchar(500)" json:"notify_channels,omitempty"` // 通知渠道
	NotifyUsers    *string `gorm:"type:text" json:"notify_users,omitempty"`            // 通知用户
	NotifyGroups   *string `gorm:"type:text" json:"notify_groups,omitempty"`           // 通知组
	NotifyTemplate *string `gorm:"type:varchar(255)" json:"notify_template,omitempty"` // 通知模板

	// 抑制配置
	SuppressDuration int        `gorm:"default:3600" json:"suppress_duration"` // 抑制时长（秒）
	SuppressCount    int        `gorm:"default:1" json:"suppress_count"`       // 抑制次数
	LastTriggeredAt  *time.Time `json:"last_triggered_at,omitempty"`           // 最后触发时间
	TriggerCount     int64      `gorm:"default:0" json:"trigger_count"`        // 触发次数

	// 状态信息
	Status      string     `gorm:"type:enum('active','suppressed','resolved','disabled');default:'active'" json:"status"` // 规则状态
	LastStatus  *string    `gorm:"type:varchar(20)" json:"last_status,omitempty"`                                         // 上次状态
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`                                                                 // 恢复时间
	NextCheckAt *time.Time `json:"next_check_at,omitempty"`                                                               // 下次检查时间

	// 创建者信息
	CreatedBy uint `gorm:"not null" json:"created_by"` // 创建者ID
	UpdatedBy uint `gorm:"not null" json:"updated_by"` // 更新者ID

	// 关联关系
	Creator User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater User           `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	Alerts  []AlertRecord  `gorm:"foreignKey:RuleID" json:"alerts,omitempty"`
	Metrics []SystemMetric `gorm:"many2many:metric_alert_rules;" json:"metrics,omitempty"`
}

// TableName 告警规则表名
func (AlertRule) TableName() string {
	return "alert_rules"
}

// BeforeCreate 创建前钩子
func (ar *AlertRule) BeforeCreate(tx *gorm.DB) error {
	if ar.UUID == "" {
		ar.UUID = basemodels.GenerateUUID()
	}
	return ar.BaseModel.BeforeCreate(tx)
}

// IsActive 检查规则是否激活
func (ar *AlertRule) IsActive() bool {
	return ar.IsEnabled && ar.Status == "active"
}

// ShouldTrigger 检查是否应该触发告警
func (ar *AlertRule) ShouldTrigger(value float64) bool {
	if !ar.IsActive() {
		return false
	}

	switch ar.Operator {
	case "gt":
		return value > ar.Threshold
	case "gte":
		return value >= ar.Threshold
	case "lt":
		return value < ar.Threshold
	case "lte":
		return value <= ar.Threshold
	case "eq":
		return value == ar.Threshold
	case "ne":
		return value != ar.Threshold
	default:
		return false
	}
}

// IncrementTrigger 增加触发次数
func (ar *AlertRule) IncrementTrigger() {
	ar.TriggerCount++
	now := time.Now()
	ar.LastTriggeredAt = &now
}

// MarkAsResolved 标记为已恢复
func (ar *AlertRule) MarkAsResolved() {
	ar.LastStatus = &ar.Status
	ar.Status = "resolved"
	now := time.Now()
	ar.ResolvedAt = &now
}

// AlertRecord 告警记录表结构
type AlertRecord struct {
	basemodels.BaseModel
	// 基本信息
	UUID   string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 告警唯一标识符
	RuleID uint   `gorm:"not null;index" json:"rule_id"`                  // 规则ID

	// 告警信息
	Title       string  `gorm:"type:varchar(255);not null" json:"title"` // 告警标题
	Message     string  `gorm:"type:text;not null" json:"message"`       // 告警消息
	MetricValue float64 `gorm:"not null" json:"metric_value"`            // 触发时的指标值
	Threshold   float64 `gorm:"not null" json:"threshold"`               // 阈值

	// 状态信息
	Status     string     `gorm:"type:enum('firing','resolved','suppressed','acknowledged');default:'firing'" json:"status"` // 告警状态
	Severity   string     `gorm:"type:enum('info','warning','critical','emergency');default:'warning'" json:"severity"`      // 严重程度
	FiredAt    time.Time  `gorm:"not null;index" json:"fired_at"`                                                            // 触发时间
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`                                                                     // 恢复时间
	Duration   *int       `json:"duration,omitempty"`                                                                        // 持续时间（秒）

	// 上下文信息
	NodeName   string              `gorm:"type:varchar(100);not null" json:"node_name"`   // 节点名称
	NodeIP     string              `gorm:"type:varchar(45);not null" json:"node_ip"`      // 节点IP
	MetricName string              `gorm:"type:varchar(100);not null" json:"metric_name"` // 指标名称
	Labels     *basemodels.JSONMap `gorm:"type:json" json:"labels,omitempty"`             // 标签信息
	Context    *basemodels.JSONMap `gorm:"type:json" json:"context,omitempty"`            // 上下文信息

	// 通知信息
	NotificationSent   bool       `gorm:"default:false" json:"notification_sent"`                // 是否已发送通知
	NotificationStatus *string    `gorm:"type:varchar(50)" json:"notification_status,omitempty"` // 通知状态
	NotifiedAt         *time.Time `json:"notified_at,omitempty"`                                 // 通知时间
	NotifiedUsers      *string    `gorm:"type:text" json:"notified_users,omitempty"`             // 已通知用户

	// 确认信息
	IsAcknowledged  bool       `gorm:"default:false" json:"is_acknowledged"`        // 是否已确认
	AcknowledgedBy  *uint      `json:"acknowledged_by,omitempty"`                   // 确认者ID
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty"`                   // 确认时间
	AcknowledgeNote *string    `gorm:"type:text" json:"acknowledge_note,omitempty"` // 确认备注

	// 关联关系
	Rule         AlertRule `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	Acknowledger *User     `gorm:"foreignKey:AcknowledgedBy" json:"acknowledger,omitempty"`
}

// TableName 告警记录表名
func (AlertRecord) TableName() string {
	return "alert_records"
}

// BeforeCreate 创建前钩子
func (ar *AlertRecord) BeforeCreate(tx *gorm.DB) error {
	if ar.UUID == "" {
		ar.UUID = basemodels.GenerateUUID()
	}

	if ar.FiredAt.IsZero() {
		ar.FiredAt = time.Now()
	}

	return ar.BaseModel.BeforeCreate(tx)
}

// IsActive 检查告警是否激活
func (ar *AlertRecord) IsActive() bool {
	return ar.Status == "firing"
}

// Resolve 恢复告警
func (ar *AlertRecord) Resolve() {
	ar.Status = "resolved"
	now := time.Now()
	ar.ResolvedAt = &now

	if !ar.FiredAt.IsZero() {
		duration := int(now.Sub(ar.FiredAt).Seconds())
		ar.Duration = &duration
	}
}

// Acknowledge 确认告警
func (ar *AlertRecord) Acknowledge(userID uint, note string) {
	ar.IsAcknowledged = true
	ar.Status = "acknowledged"
	ar.AcknowledgedBy = &userID
	now := time.Now()
	ar.AcknowledgedAt = &now
	if note != "" {
		ar.AcknowledgeNote = &note
	}
}

// MarkNotificationSent 标记通知已发送
func (ar *AlertRecord) MarkNotificationSent(status string, users []string) {
	ar.NotificationSent = true
	ar.NotificationStatus = &status
	now := time.Now()
	ar.NotifiedAt = &now
	if len(users) > 0 {
		userStr := basemodels.JoinString(users, ",")
		ar.NotifiedUsers = &userStr
	}
}

// 监控指标类型常量
const (
	MetricTypeCPU        = "cpu"        // CPU使用率
	MetricTypeMemory     = "memory"     // 内存使用率
	MetricTypeDisk       = "disk"       // 磁盘使用率
	MetricTypeNetwork    = "network"    // 网络流量
	MetricTypeLoad       = "load"       // 系统负载
	MetricTypeProcess    = "process"    // 进程数量
	MetricTypeConnection = "connection" // 连接数
	MetricTypeStorage    = "storage"    // 存储使用量
	MetricTypeUptime     = "uptime"     // 运行时间
	MetricTypeLatency    = "latency"    // 延迟
	MetricTypeThroughput = "throughput" // 吞吐量
	MetricTypeError      = "error"      // 错误率
	MetricTypeCustom     = "custom"     // 自定义指标
)

// 告警严重程度常量
const (
	SeverityInfo      = "info"      // 信息
	SeverityWarning   = "critical"  // 警告
	SeverityCritical  = "critical"  // 严重
	SeverityEmergency = "emergency" // 紧急
)

// 告警状态常量
const (
	AlertStatusFiring       = "firing"       // 触发中
	AlertStatusResolved     = "resolved"     // 已恢复
	AlertStatusSuppressed   = "suppressed"   // 已抑制
	AlertStatusAcknowledged = "acknowledged" // 已确认
)

// 操作符常量
const (
	OperatorGreaterThan      = "gt"  // 大于
	OperatorGreaterThanEqual = "gte" // 大于等于
	OperatorLessThan         = "lt"  // 小于
	OperatorLessThanEqual    = "lte" // 小于等于
	OperatorEqual            = "eq"  // 等于
	OperatorNotEqual         = "ne"  // 不等于
)

// 通知渠道常量
const (
	NotifyChannelEmail    = "email"    // 邮件
	NotifyChannelSMS      = "sms"      // 短信
	NotifyChannelWebhook  = "webhook"  // Webhook
	NotifyChannelSlack    = "slack"    // Slack
	NotifyChannelDingTalk = "dingtalk" // 钉钉
	NotifyChannelWeChat   = "wechat"   // 企业微信
	NotifyChannelTelegram = "telegram" // Telegram
)
