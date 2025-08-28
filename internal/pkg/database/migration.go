package database

import (
	"fmt"
	"log"
	"reflect"

	"gorm.io/gorm"

	"cloudpan/internal/pkg/database/models"
)

// MigrationConfig 迁移配置
type MigrationConfig struct {
	AutoMigrate bool `json:"auto_migrate"` // 是否自动迁移
	DropFirst   bool `json:"drop_first"`   // 是否先删除表
	CreateIndex bool `json:"create_index"` // 是否创建索引
}

// DefaultMigrationConfig 默认迁移配置
var DefaultMigrationConfig = &MigrationConfig{
	AutoMigrate: true,
	DropFirst:   false,
	CreateIndex: true,
}

// ModelRegistry 模型注册表
var ModelRegistry = make(map[string]interface{})

// RegisterModel 注册模型
func RegisterModel(name string, model interface{}) {
	ModelRegistry[name] = model
	log.Printf("Registered model: %s", name)
}

// GetRegisteredModels 获取所有注册的模型
func GetRegisteredModels() map[string]interface{} {
	return ModelRegistry
}

// getMigrationConfig 获取迁移配置
func getMigrationConfig(cfg []*MigrationConfig) *MigrationConfig {
	config := DefaultMigrationConfig
	if len(cfg) > 0 && cfg[0] != nil {
		config = cfg[0]
	}
	return config
}

// collectModels 收集所有模型
func collectModels() []interface{} {
	var models []interface{}
	for name, model := range ModelRegistry {
		// 如果是指针，获取其元素类型
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}

		// 创建模型实例
		modelInstance := reflect.New(modelType).Interface()
		models = append(models, modelInstance)

		log.Printf("Added model for migration: %s", name)
	}
	return models
}

// performMigration 执行迁移操作
func performMigration(db *gorm.DB, models []interface{}, config *MigrationConfig) error {
	// 如果需要先删除表
	if config.DropFirst {
		log.Println("Dropping existing tables...")
		if err := dropTables(db, models); err != nil {
			return fmt.Errorf("failed to drop tables: %w", err)
		}
	}

	// 执行迁移
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	// 创建索引
	if config.CreateIndex {
		if err := createIndexes(db, models); err != nil {
			log.Printf("Warning: failed to create some indexes: %v", err)
		}
	}

	return nil
}

// AutoMigrate 自动迁移所有注册的模型
func AutoMigrate(cfg ...*MigrationConfig) error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	config := getMigrationConfig(cfg)

	if !config.AutoMigrate {
		log.Println("Auto migration is disabled")
		return nil
	}

	log.Println("Starting database migration...")

	// 收集所有模型
	models := collectModels()

	// 执行迁移操作
	if err := performMigration(db, models, config); err != nil {
		return err
	}

	log.Printf("Database migration completed successfully, migrated %d models", len(models))
	return nil
}

// MigrateModel 迁移单个模型
func MigrateModel(model interface{}) error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	return db.AutoMigrate(model)
}

// dropTables 删除表
func dropTables(db *gorm.DB, models []interface{}) error {
	for _, model := range models {
		if db.Migrator().HasTable(model) {
			if err := db.Migrator().DropTable(model); err != nil {
				return fmt.Errorf("failed to drop table for model %T: %w", model, err)
			}
			log.Printf("Dropped table for model: %T", model)
		}
	}
	return nil
}

// createIndexes 创建索引
func createIndexes(db *gorm.DB, models []interface{}) error {
	for _, model := range models {
		if err := createModelIndexes(db, model); err != nil {
			log.Printf("Warning: failed to create indexes for model %T: %v", model, err)
		}
	}
	return nil
}

// createModelIndexes 为单个模型创建索引
func createModelIndexes(db *gorm.DB, model interface{}) error {
	// 这里可以根据模型的特定需求创建索引
	// 例如，为常用查询字段创建复合索引

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	switch modelType.Name() {
	case "User":
		// 为用户表创建特定索引
		return createUserIndexes(db, model)
	case "File":
		// 为文件表创建特定索引
		return createFileIndexes(db, model)
	case "Team":
		// 为团队表创建特定索引
		return createTeamIndexes(db, model)
	default:
		// 为所有模型创建通用索引
		return createCommonIndexes(db, model)
	}
}

// createCommonIndexes 创建通用索引
func createCommonIndexes(_ *gorm.DB, model interface{}) error {
	// 检查是否有BaseModel字段，如果有，创建通用索引
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}

	// 为created_at字段创建索引（如果存在）
	if modelValue.FieldByName("CreatedAt").IsValid() {
		// GORM会自动处理这些基础字段的索引
	}

	return nil
}

// createUserIndexes 创建用户表索引
func createUserIndexes(_ *gorm.DB, _ interface{}) error {
	// 用户表特定索引将在用户模型中定义
	return nil
}

// createFileIndexes 创建文件表索引
func createFileIndexes(_ *gorm.DB, _ interface{}) error {
	// 文件表特定索引将在文件模型中定义
	return nil
}

// createTeamIndexes 创建团队表索引
func createTeamIndexes(_ *gorm.DB, _ interface{}) error {
	// 团队表特定索引将在团队模型中定义
	return nil
}

// CheckMigrationStatus 检查迁移状态
func CheckMigrationStatus() map[string]interface{} {
	db := GetDB()
	if db == nil {
		return map[string]interface{}{
			"error": "database not initialized",
		}
	}

	status := make(map[string]interface{})
	migrator := db.Migrator()

	for name, model := range ModelRegistry {
		modelStatus := map[string]interface{}{
			"exists": migrator.HasTable(model),
		}

		if migrator.HasTable(model) {
			// 获取表的列信息
			columns, err := migrator.ColumnTypes(model)
			if err != nil {
				modelStatus["error"] = err.Error()
			} else {
				columnNames := make([]string, len(columns))
				for i, col := range columns {
					columnNames[i] = col.Name()
				}
				modelStatus["columns"] = columnNames
			}

			// 获取表的索引信息
			if indexes, err := migrator.GetIndexes(model); err == nil && len(indexes) > 0 {
				indexNames := make([]string, len(indexes))
				for i, idx := range indexes {
					indexNames[i] = idx.Name()
				}
				modelStatus["indexes"] = indexNames
			}
		}

		status[name] = modelStatus
	}

	return status
}

// ValidateSchema 验证数据库模式
func ValidateSchema() error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	var errors []string
	migrator := db.Migrator()

	for name, model := range ModelRegistry {
		// 检查表是否存在
		if !migrator.HasTable(model) {
			errors = append(errors, fmt.Sprintf("table for model %s does not exist", name))
			continue
		}

		// 这里可以添加更多的验证逻辑
		// 例如检查必需的列是否存在、数据类型是否正确等
	}

	if len(errors) > 0 {
		return fmt.Errorf("schema validation failed: %v", errors)
	}

	return nil
}

// init 初始化时注册基础模型
func init() {
	// 注册基础模型（这些主要用于测试和演示）
	RegisterModel("BaseModel", &models.BaseModel{})
	RegisterModel("AuditModel", &models.AuditModel{})
	RegisterModel("StatusModel", &models.StatusModel{})
	RegisterModel("SortModel", &models.SortModel{})
}
