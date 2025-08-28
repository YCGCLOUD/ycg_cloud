package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"cloudpan/internal/pkg/config"
	"cloudpan/internal/pkg/database"
)

func main() {
	// 定义命令行参数
	var (
		action      = flag.String("action", "migrate", "Action to perform: migrate, status, validate, drop")
		configPath  = flag.String("config", "configs/config.yaml", "Path to config file")
		dropFirst   = flag.Bool("drop", false, "Drop tables before migration")
		createIndex = flag.Bool("index", true, "Create indexes after migration")
	)
	flag.Parse()

	// 初始化系统
	if err := initSystem(*configPath); err != nil {
		log.Fatalf("Failed to initialize system: %v", err)
	}
	defer database.Close()

	// 执行操作
	if err := executeAction(*action, *dropFirst, *createIndex); err != nil {
		log.Fatalf("Operation failed: %v", err)
	}
}

// initSystem 初始化系统
func initSystem(configPath string) error {
	// 初始化配置
	if err := config.LoadFromFile(configPath); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// 初始化数据库
	if err := database.Init(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	return nil
}

// executeAction 执行操作
func executeAction(action string, dropFirst, createIndex bool) error {
	switch action {
	case "migrate":
		return handleMigration(dropFirst, createIndex)
	case "status":
		return handleStatus()
	case "validate":
		return handleValidation()
	case "drop":
		return handleDrop()
	default:
		return handleUnknownAction(action)
	}
}

// handleMigration 处理迁移操作
func handleMigration(dropFirst, createIndex bool) error {
	if err := runMigration(dropFirst, createIndex); err != nil {
		return err
	}
	fmt.Println("Migration completed successfully")
	return nil
}

// handleStatus 处理状态查询
func handleStatus() error {
	return showMigrationStatus()
}

// handleValidation 处理模式验证
func handleValidation() error {
	if err := validateSchema(); err != nil {
		return err
	}
	fmt.Println("Schema validation passed")
	return nil
}

// handleDrop 处理删除操作
func handleDrop() error {
	if err := dropAllTables(); err != nil {
		return err
	}
	fmt.Println("All tables dropped successfully")
	return nil
}

// handleUnknownAction 处理未知操作
func handleUnknownAction(action string) error {
	fmt.Printf("Unknown action: %s\n", action)
	fmt.Println("Available actions: migrate, status, validate, drop")
	os.Exit(1)
	return nil
}

// runMigration 执行迁移
func runMigration(dropFirst, createIndex bool) error {
	migrationConfig := &database.MigrationConfig{
		AutoMigrate: true,
		DropFirst:   dropFirst,
		CreateIndex: createIndex,
	}

	return database.MigrateAllModels(migrationConfig)
}

// showMigrationStatus 显示迁移状态
func showMigrationStatus() error {
	// 注册所有模型
	database.RegisterAllModels()

	status := database.CheckMigrationStatus()

	fmt.Println("Migration Status:")
	fmt.Println("================")

	for modelName, info := range status {
		if infoMap, ok := info.(map[string]interface{}); ok {
			fmt.Printf("\nModel: %s\n", modelName)

			if exists, ok := infoMap["exists"].(bool); ok {
				fmt.Printf("  Table exists: %t\n", exists)

				if exists {
					if columns, ok := infoMap["columns"].([]string); ok {
						fmt.Printf("  Columns: %v\n", columns)
					}
					if indexes, ok := infoMap["indexes"].([]string); ok && len(indexes) > 0 {
						fmt.Printf("  Indexes: %v\n", indexes)
					}
				}
			}

			if err, ok := infoMap["error"]; ok {
				fmt.Printf("  Error: %v\n", err)
			}
		}
	}

	return nil
}

// validateSchema 验证数据库模式
func validateSchema() error {
	// 注册所有模型
	database.RegisterAllModels()

	return database.ValidateSchema()
}

// dropAllTables 删除所有表
func dropAllTables() error {
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	models := database.GetAllModels()

	fmt.Printf("Dropping %d tables...\n", len(models))

	for _, model := range models {
		if db.Migrator().HasTable(model) {
			if err := db.Migrator().DropTable(model); err != nil {
				return fmt.Errorf("failed to drop table for model %T: %w", model, err)
			}
			fmt.Printf("Dropped table for model: %T\n", model)
		}
	}

	return nil
}
