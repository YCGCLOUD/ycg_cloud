# Gorm ORM 集成文档

## 概述

本模块完成了Gorm ORM的深度集成，提供了完整的数据库操作功能，包括事务管理、分页查询、软删除、插件系统等高级特性。

## 主要功能

### 1. 数据库连接管理
- MySQL连接池配置和优化
- 连接健康检查和监控
- 优雅的连接关闭

### 2. 事务管理
- 支持手动事务控制
- 带上下文的事务管理
- 自动回滚和提交

### 3. 分页查询
- 灵活的分页参数配置
- 自动计算总页数和总记录数
- 支持排序和过滤

### 4. 模型基础
- 统一的基础模型（BaseModel）
- 审计模型（AuditModel）
- 状态模型（StatusModel）
- 软删除支持

### 5. 插件系统
- 审计插件：记录数据变更
- 性能监控插件：监控慢查询
- 链路追踪插件：支持分布式追踪

### 6. 迁移管理
- 自动模型注册
- 批量迁移执行
- 迁移状态检查

## 使用示例

### 基本操作

```go
package main

import (
    "cloudpan/internal/pkg/database"
    "cloudpan/internal/pkg/database/models"
)

// 获取数据库连接
db := database.GetDB()

// 使用基础模型
type User struct {
    models.BaseModel
    Username string `gorm:"uniqueIndex;size:50" json:"username"`
    Email    string `gorm:"uniqueIndex;size:100" json:"email"`
    Password string `gorm:"size:255" json:"-"`
}
```

### 事务操作

```go
// 简单事务
err := database.Transaction(func(tx *gorm.DB) error {
    user := &User{Username: "test", Email: "test@example.com"}
    if err := tx.Create(user).Error; err != nil {
        return err
    }
    
    // 其他操作...
    return nil
})

// 带上下文的事务
ctx := context.WithTimeout(context.Background(), 10*time.Second)
err := database.TransactionWithContext(ctx, func(tx *gorm.DB) error {
    // 事务操作
    return nil
})
```

### 分页查询

```go
var users []User
opts := &database.QueryOptions{
    Page: 1,
    Size: 20,
    Sort: "created_at",
    Order: "desc",
    Filters: map[string]interface{}{
        "status": "active",
    },
    Preloads: []string{"Profile"},
}

result, err := database.Paginate(db.Model(&User{}), &users, opts)
if err != nil {
    // 处理错误
}

fmt.Printf("总记录数: %d, 当前页: %d, 总页数: %d\n", 
    result.Total, result.Page, result.TotalPages)
```

### 批量操作

```go
// 批量创建
users := []User{
    {Username: "user1", Email: "user1@example.com"},
    {Username: "user2", Email: "user2@example.com"},
}
err := database.BatchCreate(db, &users, 100)

// 批量更新
updates := map[string]interface{}{
    "status": "inactive",
    "updated_at": time.Now(),
}
err := database.BatchUpdate(db, &User{}, updates, "created_at < ?", time.Now().AddDate(0, -1, 0))

// 批量软删除
err := database.BatchDelete(db, &User{}, "status = ?", "disabled")
```

### 高级查询操作

```go
// 检查记录是否存在
exists, err := database.Exists(db, &User{}, "email = ?", "test@example.com")

// 获取或创建记录
user := &User{Username: "newuser", Email: "new@example.com"}
err := database.GetOrCreate(db, user, map[string]interface{}{
    "email": user.Email,
})

// 乐观锁更新
err := database.OptimisticLocking(db, &user, user.Version, map[string]interface{}{
    "username": "updated_username",
})

// 悲观锁查询
err := database.PessimisticLocking(db, &user, "id = ?", 1)
```

### 软删除操作

```go
// 软删除
err := database.BatchDelete(db, &User{}, "id = ?", userID)

// 恢复软删除的记录
err := database.Restore(db, &User{}, "id = ?", userID)

// 物理删除
err := database.ForceDelete(db, &User{}, "id = ?", userID)
```

### 上下文管理

```go
// 设置用户上下文（用于审计）
dbWithUser := database.WithUserContext(db, currentUserID)

// 设置追踪上下文
dbWithTrace := database.WithTraceContext(db, "trace-id-123")

// 设置超时上下文
dbWithTimeout := database.WithTimeout(db, 30*time.Second)
```

## 模型定义

### 基础模型

```go
// 使用基础模型（包含软删除）
type Product struct {
    models.BaseModel
    Name        string  `gorm:"size:100;not null" json:"name"`
    Price       float64 `gorm:"type:decimal(10,2)" json:"price"`
    Description string  `gorm:"type:text" json:"description"`
}

// 使用审计模型（包含创建者和更新者）
type Order struct {
    models.AuditModel
    OrderNo     string    `gorm:"uniqueIndex;size:50" json:"order_no"`
    TotalAmount float64   `gorm:"type:decimal(10,2)" json:"total_amount"`
    Status      string    `gorm:"size:20;default:'pending'" json:"status"`
}

// 使用状态模型
type Campaign struct {
    models.StatusModel
    Name        string    `gorm:"size:100" json:"name"`
    StartTime   time.Time `json:"start_time"`
    EndTime     time.Time `json:"end_time"`
}
```

### 自定义表名

```go
func (User) TableName() string {
    return "users"
}

func (Product) TableName() string {
    return "products"
}
```

## 模型注册和迁移

```go
// 注册模型
database.RegisterModel("User", &User{})
database.RegisterModel("Product", &Product{})
database.RegisterModel("Order", &Order{})

// 执行自动迁移
err := database.AutoMigrate()

// 检查迁移状态
status := database.CheckMigrationStatus()
fmt.Printf("迁移状态: %+v\n", status)

// 验证数据库模式
err := database.ValidateSchema()
```

## 插件配置

```go
// 安装默认插件
plugins := database.GetDefaultPlugins()
err := database.InstallPlugins(db, plugins...)

// 自定义插件配置
metricsPlugin := &database.MetricsPlugin{
    SlowQueryThreshold: 500 * time.Millisecond,
}
err := database.InstallPlugins(db, metricsPlugin)
```

## 性能监控

```go
// 获取连接池统计
stats := database.GetConnectionStats()
fmt.Printf("连接池统计: %+v\n", stats)

// 健康检查
err := database.HealthCheck()
if err != nil {
    log.Printf("数据库健康检查失败: %v", err)
}

// 完整状态检查
status := database.Status()
fmt.Printf("数据库状态: %+v\n", status)
```

## 最佳实践

### 1. 事务使用建议
- 对于复杂的业务操作，使用事务确保数据一致性
- 避免长时间持有事务，可能导致锁等待
- 在事务中避免调用外部服务

### 2. 分页查询优化
- 合理设置页大小，避免单次查询过多数据
- 使用索引优化排序字段
- 对于大数据集，考虑使用游标分页

### 3. 模型设计建议
- 继承适当的基础模型减少重复代码
- 合理使用索引提高查询性能
- 遵循数据库设计范式

### 4. 插件使用建议
- 生产环境中适当配置慢查询阈值
- 使用审计插件记录重要数据变更
- 结合监控系统收集性能指标

## 注意事项

1. **测试环境**: 大部分功能测试需要真实的数据库连接
2. **性能考虑**: 在高并发场景下合理配置连接池参数
3. **错误处理**: 始终检查和处理数据库操作的错误
4. **安全考虑**: 使用参数化查询防止SQL注入
5. **版本控制**: 使用乐观锁处理并发更新冲突

## 配置参数

数据库相关配置在 `config/types.go` 中的 `MySQLConfig` 结构体中定义：

```yaml
database:
  mysql:
    max_idle_conns: 10      # 最大空闲连接数
    max_open_conns: 100     # 最大打开连接数
    conn_max_lifetime: "1h" # 连接最大生存时间
    conn_max_idle_time: "10m" # 连接最大空闲时间
```

通过这个完整的Gorm集成，我们提供了一个功能丰富、性能优化、易于使用的数据库操作层。