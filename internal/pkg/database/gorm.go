package database

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TransactionOptions 事务选项
type TransactionOptions struct {
	Isolation sql.IsolationLevel
	ReadOnly  bool
	Timeout   time.Duration
}

// DefaultTransactionOptions 默认事务选项
var DefaultTransactionOptions = &TransactionOptions{
	Timeout: 30 * time.Second,
}

// PaginationResult 分页查询结果
type PaginationResult struct {
	Records    interface{} `json:"records"`     // 查询结果
	Total      int64       `json:"total"`       // 总记录数
	Page       int         `json:"page"`        // 当前页码
	Size       int         `json:"size"`        // 每页大小
	TotalPages int         `json:"total_pages"` // 总页数
}

// QueryOptions 查询选项
type QueryOptions struct {
	Page     int                    `json:"page"`     // 页码，从1开始
	Size     int                    `json:"size"`     // 每页大小
	Sort     string                 `json:"sort"`     // 排序字段
	Order    string                 `json:"order"`    // 排序方向：asc/desc
	Filters  map[string]interface{} `json:"filters"`  // 过滤条件
	Preloads []string               `json:"preloads"` // 预加载关联
}

// DefaultQueryOptions 默认查询选项
var DefaultQueryOptions = &QueryOptions{
	Page:  1,
	Size:  20,
	Order: "desc",
}

// fieldNameRegex 字段名验证正则表达式（仅允许字母、数字、下划线和点号）
var fieldNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.]*$`)

// isValidFieldName 验证字段名是否安全
func isValidFieldName(field string) bool {
	if field == "" || len(field) > 64 {
		return false
	}
	return fieldNameRegex.MatchString(field)
}

// Transaction 执行事务
func Transaction(fn func(tx *gorm.DB) error, opts ...*TransactionOptions) error {
	db := GetDB()

	// 获取事务选项
	var options *TransactionOptions
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	} else {
		options = DefaultTransactionOptions
	}

	// 设置超时上下文
	ctx := context.Background()
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// 开始事务
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// 确保事务回滚
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 执行事务逻辑
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TransactionWithContext 带上下文的事务执行
func TransactionWithContext(ctx context.Context, fn func(tx *gorm.DB) error) error {
	db := GetDB()

	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// validatePaginationOptions 验证分页参数
func validatePaginationOptions(opts *QueryOptions) *QueryOptions {
	if opts == nil {
		opts = DefaultQueryOptions
	}

	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.Size < 1 || opts.Size > 1000 {
		opts.Size = 20
	}

	return opts
}

// applyFilters 应用过滤条件
func applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if filters == nil {
		return query
	}

	for field, value := range filters {
		if value != nil && isValidFieldName(field) {
			// 使用GORM的Where方法，更安全且性能更好
			query = query.Where(field+" = ?", value)
		}
	}

	return query
}

// applySorting 应用排序
func applySorting(query *gorm.DB, sort, order string) *gorm.DB {
	if sort == "" || !isValidFieldName(sort) {
		return query
	}

	// 验证排序方向
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// 直接拼接，避免使用fmt.Sprintf
	orderClause := sort + " " + order
	return query.Order(orderClause)
}

// applyPreloads 应用预加载
func applyPreloads(query *gorm.DB, preloads []string) *gorm.DB {
	for _, preload := range preloads {
		if preload != "" {
			query = query.Preload(preload)
		}
	}

	return query
}

// calculateTotalPages 计算总页数
func calculateTotalPages(total int64, size int) int {
	return int((total + int64(size) - 1) / int64(size))
}

// Paginate 分页查询
func Paginate(db *gorm.DB, result interface{}, opts *QueryOptions) (*PaginationResult, error) {
	// 验证参数
	opts = validatePaginationOptions(opts)

	// 创建查询副本并应用条件
	query := db
	query = applyFilters(query, opts.Filters)

	// 获取总数
	var total int64
	if err := query.Model(result).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// 应用排序和预加载
	query = applySorting(query, opts.Sort, opts.Order)
	query = applyPreloads(query, opts.Preloads)

	// 计算偏移量并执行查询
	offset := (opts.Page - 1) * opts.Size
	if err := query.Offset(offset).Limit(opts.Size).Find(result).Error; err != nil {
		return nil, fmt.Errorf("failed to query records: %w", err)
	}

	return &PaginationResult{
		Records:    result,
		Total:      total,
		Page:       opts.Page,
		Size:       opts.Size,
		TotalPages: calculateTotalPages(total, opts.Size),
	}, nil
}

// BatchCreate 批量创建
func BatchCreate(db *gorm.DB, data interface{}, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}

	return db.CreateInBatches(data, batchSize).Error
}

// BatchUpdate 批量更新
func BatchUpdate(db *gorm.DB, model interface{}, updates map[string]interface{}, where ...interface{}) error {
	query := db.Model(model)

	if len(where) > 0 {
		query = query.Where(where[0], where[1:]...)
	}

	return query.Updates(updates).Error
}

// BatchDelete 批量删除（软删除）
func BatchDelete(db *gorm.DB, model interface{}, where ...interface{}) error {
	query := db.Model(model)

	if len(where) > 0 {
		query = query.Where(where[0], where[1:]...)
	}

	return query.Delete(model).Error
}

// ForceDelete 强制删除（物理删除）
func ForceDelete(db *gorm.DB, model interface{}, where ...interface{}) error {
	query := db.Unscoped().Model(model)

	if len(where) > 0 {
		query = query.Where(where[0], where[1:]...)
	}

	return query.Delete(model).Error
}

// Restore 恢复软删除的记录
func Restore(db *gorm.DB, model interface{}, where ...interface{}) error {
	query := db.Unscoped().Model(model)

	if len(where) > 0 {
		query = query.Where(where[0], where[1:]...)
	}

	return query.Update("deleted_at", nil).Error
}

// Upsert 插入或更新（使用ON DUPLICATE KEY UPDATE）
func Upsert(db *gorm.DB, model interface{}, columns ...string) error {
	if len(columns) == 0 {
		return db.Clauses(clause.OnConflict{UpdateAll: true}).Create(model).Error
	}

	updateColumns := make([]clause.Column, len(columns))
	for i, col := range columns {
		updateColumns[i] = clause.Column{Name: col}
	}

	return db.Clauses(clause.OnConflict{
		Columns:   updateColumns,
		UpdateAll: true,
	}).Create(model).Error
}

// GetOrCreate 获取或创建记录
func GetOrCreate(db *gorm.DB, model interface{}, where map[string]interface{}) error {
	// 先尝试查找
	query := db
	for field, value := range where {
		if isValidFieldName(field) {
			query = query.Where(field+" = ?", value)
		}
	}

	if err := query.First(model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 记录不存在，创建新记录
			return db.Create(model).Error
		}
		return err
	}

	return nil
}

// Exists 检查记录是否存在
func Exists(db *gorm.DB, model interface{}, where ...interface{}) (bool, error) {
	query := db.Model(model)

	if len(where) > 0 {
		query = query.Where(where[0], where[1:]...)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// FirstOrCreate 查找第一个匹配的记录，如果没有找到则创建
func FirstOrCreate(db *gorm.DB, model interface{}, where map[string]interface{}) error {
	query := db
	for field, value := range where {
		if isValidFieldName(field) {
			query = query.Where(field+" = ?", value)
		}
	}

	return query.FirstOrCreate(model).Error
}

// OptimisticLocking 乐观锁更新
func OptimisticLocking(db *gorm.DB, model interface{}, version int64, updates map[string]interface{}) error {
	// 添加版本检查条件
	updates["version"] = version + 1

	result := db.Model(model).Where("version = ?", version).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	// 检查是否有记录被更新
	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic locking failed: record has been modified by another process")
	}

	return nil
}

// PessimisticLocking 悲观锁查询
func PessimisticLocking(db *gorm.DB, model interface{}, where ...interface{}) error {
	query := db.Set("gorm:query_option", "FOR UPDATE")

	if len(where) > 0 {
		query = query.Where(where[0], where[1:]...)
	}

	return query.First(model).Error
}

// rawQuery 执行原生SQL查询
func RawQuery(db *gorm.DB, sql string, values ...interface{}) (*gorm.DB, error) {
	return db.Raw(sql, values...), nil
}

// RawExec 执行原生SQL命令
func RawExec(db *gorm.DB, sql string, values ...interface{}) error {
	return db.Exec(sql, values...).Error
}
