package models

import (
	"time"

	basemodels "cloudpan/internal/pkg/database/models"

	"gorm.io/gorm"
)

// SystemVersion 系统版本表结构
type SystemVersion struct {
	basemodels.BaseModel
	// 基本信息
	UUID    string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`   // 版本唯一标识符
	Version string `gorm:"type:varchar(100);not null;unique" json:"version"` // 版本号
	Name    string `gorm:"type:varchar(255);not null" json:"name"`           // 版本名称

	// 版本信息
	Description  *string `gorm:"type:text" json:"description,omitempty"`                                   // 版本描述
	ChangeLog    *string `gorm:"type:text" json:"change_log,omitempty"`                                    // 更新日志
	ReleaseNotes *string `gorm:"type:text" json:"release_notes,omitempty"`                                 // 发布说明
	VersionType  string  `gorm:"type:enum('major','minor','patch','hotfix');not null" json:"version_type"` // 版本类型

	// 状态信息
	Status       string     `gorm:"type:enum('development','testing','staging','released','deprecated','rollback');default:'development'" json:"status"` // 版本状态
	IsActive     bool       `gorm:"default:false" json:"is_active"`                                                                                      // 是否为当前激活版本
	IsCurrent    bool       `gorm:"default:false" json:"is_current"`                                                                                     // 是否为当前版本
	ReleasedAt   *time.Time `json:"released_at,omitempty"`                                                                                               // 发布时间
	DeprecatedAt *time.Time `json:"deprecated_at,omitempty"`                                                                                             // 弃用时间

	// 版本兼容性
	MinCompatibleVersion *string `gorm:"type:varchar(100)" json:"min_compatible_version,omitempty"` // 最小兼容版本
	RequiredMigrations   *string `gorm:"type:text" json:"required_migrations,omitempty"`            // 需要的数据迁移
	BreakingChanges      bool    `gorm:"default:false" json:"breaking_changes"`                     // 是否包含破坏性更改

	// 文件信息
	DownloadURL    *string `gorm:"type:varchar(500)" json:"download_url,omitempty"`    // 下载链接
	FileSize       int64   `gorm:"default:0" json:"file_size"`                         // 文件大小
	FileHash       *string `gorm:"type:varchar(255)" json:"file_hash,omitempty"`       // 文件哈希
	ChecksumMD5    *string `gorm:"type:varchar(255)" json:"checksum_md5,omitempty"`    // MD5校验
	ChecksumSHA256 *string `gorm:"type:varchar(255)" json:"checksum_sha256,omitempty"` // SHA256校验

	// 统计信息
	DownloadCount int64   `gorm:"default:0" json:"download_count"` // 下载次数
	InstallCount  int64   `gorm:"default:0" json:"install_count"`  // 安装次数
	ErrorReports  int64   `gorm:"default:0" json:"error_reports"`  // 错误报告数
	SuccessRate   float64 `gorm:"default:0" json:"success_rate"`   // 成功率

	// 创建者信息
	CreatedBy   uint  `gorm:"not null" json:"created_by"` // 创建者ID
	PublishedBy *uint `json:"published_by,omitempty"`     // 发布者ID

	// 关联关系
	Creator            User                `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Publisher          *User               `gorm:"foreignKey:PublishedBy" json:"publisher,omitempty"`
	GrayReleaseConfigs []GrayReleaseConfig `gorm:"foreignKey:VersionID" json:"gray_release_configs,omitempty"`
	VersionDeployments []VersionDeployment `gorm:"foreignKey:VersionID" json:"version_deployments,omitempty"`
	FeatureFlags       []FeatureFlag       `gorm:"foreignKey:VersionID" json:"feature_flags,omitempty"`
}

// TableName 系统版本表名
func (SystemVersion) TableName() string {
	return "system_versions"
}

// BeforeCreate 创建前钩子
func (sv *SystemVersion) BeforeCreate(tx *gorm.DB) error {
	if sv.UUID == "" {
		sv.UUID = basemodels.GenerateUUID()
	}
	return sv.BaseModel.BeforeCreate(tx)
}

// MarkAsReleased 标记为已发布
func (sv *SystemVersion) MarkAsReleased(publisherID uint) {
	sv.Status = "released"
	sv.PublishedBy = &publisherID
	now := time.Now()
	sv.ReleasedAt = &now
}

// MarkAsDeprecated 标记为已弃用
func (sv *SystemVersion) MarkAsDeprecated() {
	sv.Status = "deprecated"
	sv.IsActive = false
	now := time.Now()
	sv.DeprecatedAt = &now
}

// IncrementDownload 增加下载次数
func (sv *SystemVersion) IncrementDownload() {
	sv.DownloadCount++
}

// IncrementInstall 增加安装次数
func (sv *SystemVersion) IncrementInstall() {
	sv.InstallCount++
	if sv.DownloadCount > 0 {
		sv.SuccessRate = float64(sv.InstallCount) / float64(sv.DownloadCount) * 100
	}
}

// GrayReleaseConfig 灰度发布配置表结构
type GrayReleaseConfig struct {
	basemodels.BaseModel
	// 基本信息
	UUID      string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 配置唯一标识符
	VersionID uint   `gorm:"not null;index" json:"version_id"`               // 版本ID
	Name      string `gorm:"type:varchar(255);not null" json:"name"`         // 配置名称

	// 灰度策略
	Strategy    string `gorm:"type:enum('percentage','user_list','user_group','region','device');not null" json:"strategy"` // 灰度策略
	TargetValue string `gorm:"type:text;not null" json:"target_value"`                                                      // 目标值
	Percentage  *int   `json:"percentage,omitempty"`                                                                        // 灰度百分比
	MaxUsers    *int   `json:"max_users,omitempty"`                                                                         // 最大用户数

	// 状态信息
	Status    string     `gorm:"type:enum('pending','active','paused','completed','stopped');default:'pending'" json:"status"` // 状态
	IsActive  bool       `gorm:"default:false" json:"is_active"`                                                               // 是否激活
	StartedAt *time.Time `json:"started_at,omitempty"`                                                                         // 开始时间
	EndedAt   *time.Time `json:"ended_at,omitempty"`                                                                           // 结束时间

	// 条件配置
	Conditions *basemodels.JSONMap `gorm:"type:json" json:"conditions,omitempty"` // 条件配置
	Rules      *basemodels.JSONMap `gorm:"type:json" json:"rules,omitempty"`      // 规则配置

	// 统计信息
	TotalUsers     int64   `gorm:"default:0" json:"total_users"`     // 总用户数
	ActiveUsers    int64   `gorm:"default:0" json:"active_users"`    // 活跃用户数
	SuccessCount   int64   `gorm:"default:0" json:"success_count"`   // 成功数
	ErrorCount     int64   `gorm:"default:0" json:"error_count"`     // 错误数
	ConversionRate float64 `gorm:"default:0" json:"conversion_rate"` // 转换率
	ErrorRate      float64 `gorm:"default:0" json:"error_rate"`      // 错误率

	// 监控配置
	MonitorMetrics     *basemodels.JSONMap `gorm:"type:json" json:"monitor_metrics,omitempty"`     // 监控指标
	AlertThresholds    *basemodels.JSONMap `gorm:"type:json" json:"alert_thresholds,omitempty"`    // 告警阈值
	RollbackConditions *basemodels.JSONMap `gorm:"type:json" json:"rollback_conditions,omitempty"` // 回滚条件

	// 创建者信息
	CreatedBy uint `gorm:"not null" json:"created_by"` // 创建者ID

	// 关联关系
	Version *SystemVersion   `gorm:"foreignKey:VersionID" json:"version,omitempty"`
	Creator User             `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Logs    []GrayReleaseLog `gorm:"foreignKey:ConfigID" json:"logs,omitempty"`
}

// TableName 灰度发布配置表名
func (GrayReleaseConfig) TableName() string {
	return "gray_release_configs"
}

// BeforeCreate 创建前钩子
func (grc *GrayReleaseConfig) BeforeCreate(tx *gorm.DB) error {
	if grc.UUID == "" {
		grc.UUID = basemodels.GenerateUUID()
	}
	return grc.BaseModel.BeforeCreate(tx)
}

// Start 启动灰度发布
func (grc *GrayReleaseConfig) Start() {
	grc.Status = "active"
	grc.IsActive = true
	now := time.Now()
	grc.StartedAt = &now
}

// Stop 停止灰度发布
func (grc *GrayReleaseConfig) Stop() {
	grc.Status = "stopped"
	grc.IsActive = false
	now := time.Now()
	grc.EndedAt = &now
}

// UpdateMetrics 更新指标
func (grc *GrayReleaseConfig) UpdateMetrics() {
	total := grc.SuccessCount + grc.ErrorCount
	if total > 0 {
		grc.ErrorRate = float64(grc.ErrorCount) / float64(total) * 100
	}
	if grc.TotalUsers > 0 {
		grc.ConversionRate = float64(grc.ActiveUsers) / float64(grc.TotalUsers) * 100
	}
}

// GrayReleaseLog 灰度发布日志表结构
type GrayReleaseLog struct {
	basemodels.BaseModel
	// 基本信息
	UUID     string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 日志唯一标识符
	ConfigID uint   `gorm:"not null;index" json:"config_id"`                // 配置ID
	UserID   *uint  `gorm:"index" json:"user_id,omitempty"`                 // 用户ID

	// 日志信息
	Action     string              `gorm:"type:varchar(100);not null" json:"action"`       // 操作
	Status     string              `gorm:"type:varchar(50);not null" json:"status"`        // 状态
	Message    *string             `gorm:"type:text" json:"message,omitempty"`             // 消息
	Details    *basemodels.JSONMap `gorm:"type:json" json:"details,omitempty"`             // 详细信息
	UserAgent  *string             `gorm:"type:varchar(1000)" json:"user_agent,omitempty"` // 用户代理
	IPAddress  *string             `gorm:"type:varchar(45)" json:"ip_address,omitempty"`   // IP地址
	DeviceInfo *string             `gorm:"type:varchar(500)" json:"device_info,omitempty"` // 设备信息

	// 执行结果
	Duration     int64   `gorm:"default:0" json:"duration"`                    // 执行时长（毫秒）
	ErrorCode    *string `gorm:"type:varchar(50)" json:"error_code,omitempty"` // 错误代码
	ErrorMessage *string `gorm:"type:text" json:"error_message,omitempty"`     // 错误信息

	// 关联关系
	Config *GrayReleaseConfig `gorm:"foreignKey:ConfigID" json:"config,omitempty"`
	User   *User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 灰度发布日志表名
func (GrayReleaseLog) TableName() string {
	return "gray_release_logs"
}

// BeforeCreate 创建前钩子
func (grl *GrayReleaseLog) BeforeCreate(tx *gorm.DB) error {
	if grl.UUID == "" {
		grl.UUID = basemodels.GenerateUUID()
	}
	return grl.BaseModel.BeforeCreate(tx)
}

// VersionDeployment 版本部署表结构
type VersionDeployment struct {
	basemodels.BaseModel
	// 基本信息
	UUID        string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 部署唯一标识符
	VersionID   uint   `gorm:"not null;index" json:"version_id"`               // 版本ID
	Environment string `gorm:"type:varchar(100);not null" json:"environment"`  // 部署环境

	// 部署信息
	Status       string     `gorm:"type:enum('deploying','deployed','failed','rollback');not null" json:"status"`      // 部署状态
	DeployedAt   *time.Time `json:"deployed_at,omitempty"`                                                             // 部署时间
	RollbackAt   *time.Time `json:"rollback_at,omitempty"`                                                             // 回滚时间
	HealthStatus string     `gorm:"type:enum('healthy','unhealthy','unknown');default:'unknown'" json:"health_status"` // 健康状态

	// 部署配置
	Config    *basemodels.JSONMap `gorm:"type:json" json:"config,omitempty"`    // 部署配置
	Variables *basemodels.JSONMap `gorm:"type:json" json:"variables,omitempty"` // 环境变量

	// 监控信息
	Metrics      *basemodels.JSONMap `gorm:"type:json" json:"metrics,omitempty"`       // 监控指标
	Logs         *string             `gorm:"type:text" json:"logs,omitempty"`          // 部署日志
	ErrorMessage *string             `gorm:"type:text" json:"error_message,omitempty"` // 错误信息

	// 创建者信息
	DeployedBy uint `gorm:"not null" json:"deployed_by"` // 部署者ID

	// 关联关系
	Version  *SystemVersion `gorm:"foreignKey:VersionID" json:"version,omitempty"`
	Deployer User           `gorm:"foreignKey:DeployedBy" json:"deployer,omitempty"`
}

// TableName 版本部署表名
func (VersionDeployment) TableName() string {
	return "version_deployments"
}

// BeforeCreate 创建前钩子
func (vd *VersionDeployment) BeforeCreate(tx *gorm.DB) error {
	if vd.UUID == "" {
		vd.UUID = basemodels.GenerateUUID()
	}
	return vd.BaseModel.BeforeCreate(tx)
}

// FeatureFlag 功能特性标记表结构
type FeatureFlag struct {
	basemodels.BaseModel
	// 基本信息
	UUID      string `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"` // 特性唯一标识符
	VersionID *uint  `gorm:"index" json:"version_id,omitempty"`              // 版本ID
	Name      string `gorm:"type:varchar(255);not null;unique" json:"name"`  // 特性名称
	Key       string `gorm:"type:varchar(255);not null;unique" json:"key"`   // 特性键

	// 特性信息
	Description *string `gorm:"type:text" json:"description,omitempty"` // 特性描述
	Category    string  `gorm:"type:varchar(100)" json:"category"`      // 特性分类
	IsEnabled   bool    `gorm:"default:false" json:"is_enabled"`        // 是否启用
	IsDefault   bool    `gorm:"default:false" json:"is_default"`        // 是否默认启用

	// 目标配置
	TargetUsers   *string `gorm:"type:text" json:"target_users,omitempty"`  // 目标用户
	TargetGroups  *string `gorm:"type:text" json:"target_groups,omitempty"` // 目标用户组
	TargetPercent *int    `json:"target_percent,omitempty"`                 // 目标百分比

	// 条件配置
	Conditions *basemodels.JSONMap `gorm:"type:json" json:"conditions,omitempty"` // 启用条件

	// 时间配置
	StartTime *time.Time `json:"start_time,omitempty"` // 开始时间
	EndTime   *time.Time `json:"end_time,omitempty"`   // 结束时间

	// 统计信息
	UsageCount int64 `gorm:"default:0" json:"usage_count"` // 使用次数

	// 创建者信息
	CreatedBy uint `gorm:"not null" json:"created_by"` // 创建者ID

	// 关联关系
	Version *SystemVersion `gorm:"foreignKey:VersionID" json:"version,omitempty"`
	Creator User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName 功能特性标记表名
func (FeatureFlag) TableName() string {
	return "feature_flags"
}

// BeforeCreate 创建前钩子
func (ff *FeatureFlag) BeforeCreate(tx *gorm.DB) error {
	if ff.UUID == "" {
		ff.UUID = basemodels.GenerateUUID()
	}
	return ff.BaseModel.BeforeCreate(tx)
}

// IsActive 检查特性是否激活
func (ff *FeatureFlag) IsActive() bool {
	if !ff.IsEnabled {
		return false
	}
	now := time.Now()
	if ff.StartTime != nil && now.Before(*ff.StartTime) {
		return false
	}
	if ff.EndTime != nil && now.After(*ff.EndTime) {
		return false
	}
	return true
}

// IncrementUsage 增加使用次数
func (ff *FeatureFlag) IncrementUsage() {
	ff.UsageCount++
}

// 版本类型常量
const (
	VersionTypeMajor  = "major"  // 主版本
	VersionTypeMinor  = "minor"  // 次版本
	VersionTypePatch  = "patch"  // 补丁版本
	VersionTypeHotfix = "hotfix" // 热修复版本
)

// 版本状态常量
const (
	VersionStatusDevelopment = "development" // 开发中
	VersionStatusTesting     = "testing"     // 测试中
	VersionStatusStaging     = "staging"     // 预发布
	VersionStatusReleased    = "released"    // 已发布
	VersionStatusDeprecated  = "deprecated"  // 已弃用
	VersionStatusRollback    = "rollback"    // 已回滚
)

// 灰度策略常量
const (
	GrayStrategyPercentage = "percentage" // 按百分比
	GrayStrategyUserList   = "user_list"  // 按用户列表
	GrayStrategyUserGroup  = "user_group" // 按用户组
	GrayStrategyRegion     = "region"     // 按地区
	GrayStrategyDevice     = "device"     // 按设备
)

// 部署环境常量
const (
	DeployEnvironmentDev     = "development" // 开发环境
	DeployEnvironmentTest    = "testing"     // 测试环境
	DeployEnvironmentStaging = "staging"     // 预发布环境
	DeployEnvironmentProd    = "production"  // 生产环境
)

// 特性分类常量
const (
	FeatureCategoryUI           = "ui"           // 界面特性
	FeatureCategoryAPI          = "api"          // API特性
	FeatureCategoryStorage      = "storage"      // 存储特性
	FeatureCategorySecurity     = "security"     // 安全特性
	FeatureCategoryPerformance  = "performance"  // 性能特性
	FeatureCategoryExperimental = "experimental" // 实验特性
)
