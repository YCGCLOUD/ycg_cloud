package database

import (
	"cloudpan/internal/repository/models"
)

// RegisterAllModels 注册所有数据模型
func RegisterAllModels() {
	// 用户相关模型
	RegisterModel("User", &models.User{})
	RegisterModel("UserSession", &models.UserSession{})
	RegisterModel("UserLoginHistory", &models.UserLoginHistory{})
	RegisterModel("UserPreference", &models.UserPreference{})

	// 文件相关模型
	RegisterModel("File", &models.File{})
	RegisterModel("FileVersion", &models.FileVersion{})
	RegisterModel("FileShare", &models.FileShare{})
	RegisterModel("FileTag", &models.FileTag{})
	RegisterModel("FileUploadChunk", &models.FileUploadChunk{})

	// 团队相关模型
	RegisterModel("Team", &models.Team{})
	RegisterModel("TeamMember", &models.TeamMember{})
	RegisterModel("TeamFile", &models.TeamFile{})
	RegisterModel("TeamInvitation", &models.TeamInvitation{})

	// 消息相关模型
	RegisterModel("Conversation", &models.Conversation{})
	RegisterModel("Message", &models.Message{})
	RegisterModel("ConversationMember", &models.ConversationMember{})
	RegisterModel("MessageReadStatus", &models.MessageReadStatus{})

	// 权限相关模型
	RegisterModel("Role", &models.Role{})
	RegisterModel("Permission", &models.Permission{})
	RegisterModel("UserRole", &models.UserRole{})
	RegisterModel("RolePermission", &models.RolePermission{})

	// 系统相关模型
	RegisterModel("RecycleBin", &models.RecycleBin{})
	RegisterModel("AuditLog", &models.AuditLog{})
	RegisterModel("SystemSetting", &models.SystemSetting{})
	RegisterModel("PasswordResetToken", &models.PasswordResetToken{})

	// 新增模型
	RegisterModel("FileComment", &models.FileComment{})
	RegisterModel("CommentLike", &models.CommentLike{})
	RegisterModel("Notification", &models.Notification{})
	RegisterModel("VerificationCode", &models.VerificationCode{})
	RegisterModel("EmailTemplate", &models.EmailTemplate{})
	RegisterModel("Tag", &models.Tag{})
	RegisterModel("FileTagV2", &models.FileTagV2{})

	// 离线操作与同步模型
	RegisterModel("OfflineOperation", &models.OfflineOperation{})
	RegisterModel("OfflineFile", &models.OfflineFile{})
	RegisterModel("SyncDevice", &models.SyncDevice{})

	// 文件自动分类规则模型
	RegisterModel("AutoClassifyRule", &models.AutoClassifyRule{})
	RegisterModel("AutoClassifyLog", &models.AutoClassifyLog{})
	RegisterModel("FileClassifyTemplate", &models.FileClassifyTemplate{})

	// OSS存储配置模型
	RegisterModel("StorageProvider", &models.StorageProvider{})
	RegisterModel("StoragePolicy", &models.StoragePolicy{})
	RegisterModel("StorageMigrationTask", &models.StorageMigrationTask{})

	// 版本与灰度管理模型
	RegisterModel("SystemVersion", &models.SystemVersion{})
	RegisterModel("GrayReleaseConfig", &models.GrayReleaseConfig{})
	RegisterModel("VersionDeployment", &models.VersionDeployment{})
	RegisterModel("FeatureFlag", &models.FeatureFlag{})

	// API开放平台模型
	RegisterModel("APIApp", &models.APIApp{})
	RegisterModel("APIToken", &models.APIToken{})
	RegisterModel("Webhook", &models.Webhook{})
	RegisterModel("APILog", &models.APILog{})

	// 多语言支持模型
	RegisterModel("Language", &models.Language{})
	RegisterModel("LanguageText", &models.LanguageText{})

	// 服务器资源监控模型
	RegisterModel("SystemMetric", &models.SystemMetric{})
	RegisterModel("AlertRule", &models.AlertRule{})
	RegisterModel("AlertRecord", &models.AlertRecord{})
}

// GetAllModels 获取所有模型列表（用于手动迁移）
func GetAllModels() []interface{} {
	return []interface{}{
		// 用户相关模型
		&models.User{},
		&models.UserSession{},
		&models.UserLoginHistory{},
		&models.UserPreference{},

		// 文件相关模型
		&models.File{},
		&models.FileVersion{},
		&models.FileShare{},
		&models.FileTag{},
		&models.FileUploadChunk{},

		// 团队相关模型
		&models.Team{},
		&models.TeamMember{},
		&models.TeamFile{},
		&models.TeamInvitation{},

		// 消息相关模型
		&models.Conversation{},
		&models.Message{},
		&models.ConversationMember{},
		&models.MessageReadStatus{},

		// 权限相关模型
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},

		// 系统相关模型
		&models.RecycleBin{},
		&models.AuditLog{},
		&models.SystemSetting{},
		&models.PasswordResetToken{},

		// 新增模型
		&models.FileComment{},
		&models.CommentLike{},
		&models.Notification{},
		&models.VerificationCode{},
		&models.EmailTemplate{},
		&models.Tag{},
		&models.FileTagV2{},

		// 离线操作与同步模型
		&models.OfflineOperation{},
		&models.OfflineFile{},
		&models.SyncDevice{},

		// 文件自动分类规则模型
		&models.AutoClassifyRule{},
		&models.AutoClassifyLog{},
		&models.FileClassifyTemplate{},

		// OSS存储配置模型
		&models.StorageProvider{},
		&models.StoragePolicy{},
		&models.StorageMigrationTask{},

		// 版本与灰度管理模型
		&models.SystemVersion{},
		&models.GrayReleaseConfig{},
		&models.VersionDeployment{},
		&models.FeatureFlag{},

		// API开放平台模型
		&models.APIApp{},
		&models.APIToken{},
		&models.Webhook{},
		&models.APILog{},

		// 多语言支持模型
		&models.Language{},
		&models.LanguageText{},

		// 服务器资源监控模型
		&models.SystemMetric{},
		&models.AlertRule{},
		&models.AlertRecord{},
	}
}

// MigrateAllModels 迁移所有模型
func MigrateAllModels(config ...*MigrationConfig) error {
	// 注册所有模型
	RegisterAllModels()

	// 执行自动迁移
	return AutoMigrate(config...)
}
