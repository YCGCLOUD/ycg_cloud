package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestValidateFieldName 测试字段名验证
func TestValidateFieldName(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{
			name:     "valid field name",
			field:    "user_name",
			expected: true,
		},
		{
			name:     "valid field with dots",
			field:    "user.profile.name",
			expected: true,
		},
		{
			name:     "empty field",
			field:    "",
			expected: false,
		},
		{
			name:     "field too long",
			field:    "this_is_a_very_long_field_name_that_exceeds_the_maximum_allowed_length_limit",
			expected: false,
		},
		{
			name:     "field with invalid characters",
			field:    "user-name",
			expected: false,
		},
		{
			name:     "field starting with number",
			field:    "1user_name",
			expected: false,
		},
		{
			name:     "field with spaces",
			field:    "user name",
			expected: false,
		},
		{
			name:     "field with special characters",
			field:    "user@name",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidFieldName(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidatePaginationOptions 测试分页参数验证
func TestValidatePaginationOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    *QueryOptions
		expected *QueryOptions
	}{
		{
			name:  "nil options",
			input: nil,
			expected: &QueryOptions{
				Page:  1,
				Size:  20,
				Order: "desc",
			},
		},
		{
			name: "negative page",
			input: &QueryOptions{
				Page: -1,
				Size: 10,
			},
			expected: &QueryOptions{
				Page: 1,
				Size: 10,
			},
		},
		{
			name: "zero page",
			input: &QueryOptions{
				Page: 0,
				Size: 10,
			},
			expected: &QueryOptions{
				Page: 1,
				Size: 10,
			},
		},
		{
			name: "negative size",
			input: &QueryOptions{
				Page: 1,
				Size: -1,
			},
			expected: &QueryOptions{
				Page: 1,
				Size: 20,
			},
		},
		{
			name: "zero size",
			input: &QueryOptions{
				Page: 1,
				Size: 0,
			},
			expected: &QueryOptions{
				Page: 1,
				Size: 20,
			},
		},
		{
			name: "size too large",
			input: &QueryOptions{
				Page: 1,
				Size: 2000,
			},
			expected: &QueryOptions{
				Page: 1,
				Size: 20,
			},
		},
		{
			name: "valid options",
			input: &QueryOptions{
				Page: 2,
				Size: 50,
			},
			expected: &QueryOptions{
				Page: 2,
				Size: 50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validatePaginationOptions(tt.input)
			assert.Equal(t, tt.expected.Page, result.Page)
			assert.Equal(t, tt.expected.Size, result.Size)
		})
	}
}

// TestCalculateTotalPages 测试总页数计算
func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		size     int
		expected int
	}{
		{
			name:     "exact division",
			total:    100,
			size:     10,
			expected: 10,
		},
		{
			name:     "with remainder",
			total:    105,
			size:     10,
			expected: 11,
		},
		{
			name:     "zero total",
			total:    0,
			size:     10,
			expected: 0,
		},
		{
			name:     "one record",
			total:    1,
			size:     10,
			expected: 1,
		},
		{
			name:     "large numbers",
			total:    999999,
			size:     100,
			expected: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTotalPages(tt.total, tt.size)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTransactionOptions 测试事务选项
func TestTransactionOptions(t *testing.T) {
	// 测试默认事务选项
	assert.NotNil(t, DefaultTransactionOptions)
	assert.Equal(t, 30*time.Second, DefaultTransactionOptions.Timeout)

	// 测试自定义事务选项
	customOpts := &TransactionOptions{
		Timeout: 10 * time.Second,
	}
	assert.Equal(t, 10*time.Second, customOpts.Timeout)
}

// TestQueryOptions 测试查询选项结构
func TestQueryOptions(t *testing.T) {
	// 测试默认查询选项
	assert.NotNil(t, DefaultQueryOptions)
	assert.Equal(t, 1, DefaultQueryOptions.Page)
	assert.Equal(t, 20, DefaultQueryOptions.Size)
	assert.Equal(t, "desc", DefaultQueryOptions.Order)

	// 测试自定义查询选项
	customOpts := &QueryOptions{
		Page:  2,
		Size:  50,
		Sort:  "created_at",
		Order: "asc",
		Filters: map[string]interface{}{
			"status": "active",
		},
		Preloads: []string{"User", "Category"},
	}

	assert.Equal(t, 2, customOpts.Page)
	assert.Equal(t, 50, customOpts.Size)
	assert.Equal(t, "created_at", customOpts.Sort)
	assert.Equal(t, "asc", customOpts.Order)
	assert.Equal(t, "active", customOpts.Filters["status"])
	assert.Contains(t, customOpts.Preloads, "User")
	assert.Contains(t, customOpts.Preloads, "Category")
}

// TestPaginationResult 测试分页结果结构
func TestPaginationResult(t *testing.T) {
	records := []string{"item1", "item2", "item3"}
	result := &PaginationResult{
		Records:    records,
		Total:      100,
		Page:       2,
		Size:       20,
		TotalPages: 5,
	}

	assert.Equal(t, records, result.Records)
	assert.Equal(t, int64(100), result.Total)
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 20, result.Size)
	assert.Equal(t, 5, result.TotalPages)
}

// MockDB 模拟数据库结构体，用于测试
type MockDB struct {
	*gorm.DB
}

// TestApplyFilters 测试过滤器应用
func TestApplyFilters(t *testing.T) {
	// 注意：这个测试不需要真实的数据库连接
	// 我们主要测试过滤器逻辑

	tests := []struct {
		name    string
		filters map[string]interface{}
	}{
		{
			name:    "nil filters",
			filters: nil,
		},
		{
			name:    "empty filters",
			filters: map[string]interface{}{},
		},
		{
			name: "valid filters",
			filters: map[string]interface{}{
				"status":  "active",
				"user_id": 123,
			},
		},
		{
			name: "filters with nil values",
			filters: map[string]interface{}{
				"status":     "active",
				"deleted_at": nil,
			},
		},
		{
			name: "filters with invalid field names",
			filters: map[string]interface{}{
				"valid_field":   "value",
				"invalid-field": "value",
				"2invalid":      "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 这个测试主要验证函数不会panic
			// 在真实环境中需要GORM实例，这里我们测试逻辑
			if DB != nil {
				query := DB.Model(&struct{}{})
				result := applyFilters(query, tt.filters)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestApplySorting 测试排序应用
func TestApplySorting(t *testing.T) {
	tests := []struct {
		name  string
		sort  string
		order string
	}{
		{
			name:  "empty sort field",
			sort:  "",
			order: "asc",
		},
		{
			name:  "valid sort field",
			sort:  "created_at",
			order: "desc",
		},
		{
			name:  "invalid sort field",
			sort:  "invalid-field",
			order: "asc",
		},
		{
			name:  "invalid order",
			sort:  "created_at",
			order: "invalid",
		},
		{
			name:  "uppercase order",
			sort:  "updated_at",
			order: "ASC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 这个测试主要验证函数不会panic
			if DB != nil {
				query := DB.Model(&struct{}{})
				result := applySorting(query, tt.sort, tt.order)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestApplyPreloads 测试预加载应用
func TestApplyPreloads(t *testing.T) {
	tests := []struct {
		name     string
		preloads []string
	}{
		{
			name:     "nil preloads",
			preloads: nil,
		},
		{
			name:     "empty preloads",
			preloads: []string{},
		},
		{
			name:     "single preload",
			preloads: []string{"User"},
		},
		{
			name:     "multiple preloads",
			preloads: []string{"User", "Category", "Tags"},
		},
		{
			name:     "preloads with empty strings",
			preloads: []string{"User", "", "Category"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 这个测试主要验证函数不会panic
			if DB != nil {
				query := DB.Model(&struct{}{})
				result := applyPreloads(query, tt.preloads)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestTransactionWithoutDB 测试事务函数在没有数据库时的行为
func TestTransactionWithoutDB(t *testing.T) {
	// 保存原始DB
	originalDB := DB
	defer func() { DB = originalDB }()

	// 设置DB为nil
	DB = nil

	// 先检查GetDB()的行为
	db := GetDB()
	assert.Nil(t, db)

	// 由于GetDB()返回nil，Transaction函数应该在内部检查并返回错误
	// 而不是panic
	err := Transaction(func(tx *gorm.DB) error {
		return nil
	})

	// 应该返回错误，因为DB为nil
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")
}

// TestTransactionWithContextWithoutDB 测试带上下文的事务函数在没有数据库时的行为
func TestTransactionWithContextWithoutDB(t *testing.T) {
	// 保存原始DB
	originalDB := DB
	defer func() { DB = originalDB }()

	// 设置DB为nil
	DB = nil

	ctx := context.Background()
	err := TransactionWithContext(ctx, func(tx *gorm.DB) error {
		return nil
	})

	// 应该返回错误，因为DB为nil
	assert.Error(t, err)
}

// BenchmarkValidateFieldName 基准测试字段名验证
func BenchmarkValidateFieldName(b *testing.B) {
	testFields := []string{
		"valid_field",
		"user.profile.name",
		"invalid-field",
		"2invalid",
		"",
		"very_long_field_name_that_might_exceed_limits",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		field := testFields[i%len(testFields)]
		isValidFieldName(field)
	}
}

// BenchmarkValidatePaginationOptions 基准测试分页参数验证
func BenchmarkValidatePaginationOptions(b *testing.B) {
	opts := &QueryOptions{
		Page: 1,
		Size: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validatePaginationOptions(opts)
	}
}

// BenchmarkCalculateTotalPages 基准测试总页数计算
func BenchmarkCalculateTotalPages(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateTotalPages(1000, 20)
	}
}
