# HXLOS Cloud 云盘系统 API 文档

## 概述

本文档提供了HXLOS Cloud云盘系统所有核心模块的完整API文档，包括配置管理、数据库操作、缓存系统、错误处理等。

## 目录

- [配置管理 API](#配置管理-api)
- [数据库操作 API](#数据库操作-api)
- [缓存系统 API](#缓存系统-api)
- [错误处理 API](#错误处理-api)
- [分布式锁 API](#分布式锁-api)

---

## 配置管理 API

### 初始化配置

```go
package main

import (
    "cloudpan/internal/pkg/config"
)

func main() {
    // 初始化配置
    if err := config.Init(); err != nil {
        log.Fatalf("配置初始化失败: %v", err)
    }
    
    // 获取配置实例
    cfg := config.AppConfig
}
```

### 配置验证

```go
// 验证配置完整性
if err := config.ValidateConfig(); err != nil {
    log.Printf("配置验证失败: %v", err)
}

// 获取配置摘要
summary := config.GetConfigSummary()
fmt.Printf("配置摘要: %+v\n", summary)
```

### 环境管理

```go
// 获取当前环境
env := config.GetCurrentEnv()
fmt.Printf("当前环境: %s\n", env)

// 检查是否为开发环境
if config.IsDevelopment() {
    // 开发环境特殊逻辑
}

// 检查是否为生产环境
if config.IsProduction() {
    // 生产环境特殊逻辑
}
```

### 辅助函数

```go
// 获取用户存储目录
userStoragePath := config.GetUserStoragePath(userID)

// 获取数据库DSN
dsn := config.GetDatabaseDSN()

// 获取Redis地址
redisAddr := config.GetRedisAddr()

// 获取服务器地址
serverAddr := config.GetServerAddr()
```

---

## 数据库操作 API

### 初始化数据库

```go
package main

import (
    "cloudpan/internal/pkg/database"
    "cloudpan/internal/pkg/config"
)

func main() {
    // 初始化配置
    config.Init()
    
    // 初始化数据库
    if err := database.InitMySQL(); err != nil {
        log.Fatalf("数据库初始化失败: %v", err)
    }
    
    // 获取数据库实例
    db := database.GetDB()
}
```

### 基础模型定义

```go
import "cloudpan/internal/pkg/database/models"

// 使用基础模型
type User struct {
    models.BaseModel
    Username string `gorm:"uniqueIndex;size:50" json:"username"`
    Email    string `gorm:"uniqueIndex;size:100" json:"email"`
    Password string `gorm:"size:255" json:"-"`
}

// 使用审计模型
type Document struct {
    models.AuditModel
    Title   string `gorm:"size:200" json:"title"`
    Content string `gorm:"type:text" json:"content"`
}

// 使用状态模型
type Campaign struct {
    models.StatusModel
    Name      string    `gorm:"size:100" json:"name"`
    StartTime time.Time `json:"start_time"`
    EndTime   time.Time `json:"end_time"`
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
    
    // 其他事务操作...
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
    return err
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

### 高级查询

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

### 软删除管理

```go
// 软删除
err := database.BatchDelete(db, &User{}, "id = ?", userID)

// 恢复软删除的记录
err := database.Restore(db, &User{}, "id = ?", userID)

// 物理删除
err := database.ForceDelete(db, &User{}, "id = ?", userID)
```

---

## 缓存系统 API

### 初始化缓存

```go
package main

import (
    "cloudpan/internal/pkg/cache"
    "cloudpan/internal/pkg/config"
)

func main() {
    // 初始化配置
    config.Init()
    
    // 初始化Redis
    if err := cache.InitRedis(); err != nil {
        log.Fatalf("Redis初始化失败: %v", err)
    }
}
```

### 基础缓存操作

```go
// 设置缓存
err := cache.Cache.Set("user:123", userData, 1*time.Hour)

// 获取缓存
var user User
err := cache.Cache.Get("user:123", &user)
if err != nil {
    if errors.Is(err, cache.ErrCacheNotFound) {
        // 缓存未找到，从数据库获取
    }
}

// 删除缓存
err := cache.Cache.Delete("user:123")

// 检查缓存是否存在
exists, err := cache.Cache.Exists("user:123")
```

### 缓存键构建

```go
// 使用键构建器
userSessionKey := cache.Keys.UserSession(token)
userPermissionsKey := cache.Keys.UserPermissions(userID)
fileInfoKey := cache.Keys.FileInfo(fileID)
teamMembersKey := cache.Keys.TeamMembers(teamID)

// 验证码相关键
verifyCodeKey := cache.Keys.VerifyCode("email", "user@example.com")
verifyAttemptKey := cache.Keys.VerifyAttempt("email", "user@example.com")

// 限流相关键
rateLimitKey := cache.Keys.RateLimit(clientIP, "/api/upload")
userRateLimitKey := cache.Keys.UserRateLimit(userID, "download")
```

### 高级缓存操作

```go
// Hash操作
err := cache.Cache.HSet("user:profile:123", "name", "John Doe")
var name string
err := cache.Cache.HGet("user:profile:123", "name", &name)

// 集合操作
err := cache.Cache.SAdd("online:users", userID)
isMember, err := cache.Cache.SIsMember("online:users", userID)
members, err := cache.Cache.SMembers("online:users")

// 有序集合操作
err := cache.Cache.ZAdd("user:scores", 95.5, userID)
topUsers, err := cache.Cache.ZRange("user:scores", 0, 9)

// 原子操作
count, err := cache.Cache.Increment("api:calls:count")
newValue, err := cache.Cache.IncrementBy("user:downloads", 5)
```

### 批量操作

```go
// 批量操作
batch := cache.Cache.Batch()
batch.Set("key1", "value1", time.Hour).
      Set("key2", "value2", time.Hour).
      Delete("key3", "key4")

err := batch.Execute()
```

### TTL缓存

```go
// 设置带TTL的缓存
cache.UserSession.Set(token, sessionData, 24*time.Hour)
cache.FileUpload.Set(uploadID, uploadInfo, 2*time.Hour)
cache.VerifyCode.Set("email", email, codeData, 5*time.Minute)

// 获取TTL缓存
var sessionData SessionData
err := cache.UserSession.Get(token, &sessionData)

// 删除TTL缓存
err := cache.UserSession.Delete(token)

// 清理过期缓存
cache.UserSession.CleanupExpired()
cache.FileUpload.CleanupExpired()
```

---

## 错误处理 API

### 预定义错误

```go
import "cloudpan/internal/pkg/errors"

// 缓存相关错误
if errors.Is(err, errors.ErrCacheNotFound) {
    // 处理缓存未找到
}

// 配置相关错误
if errors.Is(err, errors.ErrConfigNotFound) {
    // 处理配置文件未找到
}

// 数据库相关错误
if errors.Is(err, errors.ErrDatabaseNotInitialized) {
    // 处理数据库未初始化
}

// 业务逻辑错误
if errors.Is(err, errors.ErrResourceNotFound) {
    // 处理资源未找到
}
```

### 错误包装

```go
// 简单错误包装
err := errors.WrapError(originalErr, "failed to create user")

// 格式化错误包装
err := errors.WrapErrorf(originalErr, "failed to create user %s", username)

// 创建特定类型错误
err := errors.NewValidationError("email", "invalid email format")
err := errors.NewResourceError("user", "create", originalErr)
```

### 错误类型检查

```go
// 检查错误类型
if errors.IsNotFoundError(err) {
    // 处理未找到错误
}

if errors.IsPermissionError(err) {
    // 处理权限错误
}

if errors.IsValidationError(err) {
    // 处理验证错误
}

if errors.IsRetryableError(err) {
    // 可重试的错误，执行重试逻辑
}
```

---

## 分布式锁 API

### 数据库锁

```go
import "cloudpan/internal/pkg/database"

// 获取悲观锁
ctx := context.Background()
lock, err := database.GetPessimisticLock(ctx, "users", "id = ?", userID)
if err != nil {
    return err
}
defer lock.Release() // 确保释放锁

// 在锁保护下执行操作
err = database.Transaction(func(tx *gorm.DB) error {
    // 执行需要锁保护的操作
    return nil
})
```

### Redis分布式锁

```go
import "cloudpan/internal/pkg/database"

// 获取分布式锁
ctx := context.Background()
lockKey := "file:process:" + fileID
ttl := 30 * time.Second

lock, err := database.AcquireDistributedLock(ctx, lockKey, ttl)
if err != nil {
    return err
}
defer lock.Release() // 确保释放锁

// 在锁保护下执行操作
// 执行文件处理逻辑...
```

### 锁使用模式

```go
// 简单锁模式
err := database.WithLock(ctx, lockKey, ttl, func() error {
    // 执行需要锁保护的操作
    return processFile(fileID)
})

// 可重入锁模式
err := database.WithReentrantLock(ctx, lockKey, ttl, func() error {
    // 可以在此处再次获取相同的锁
    return complexOperation()
})
```

---

## 健康检查和监控

### 系统健康检查

```go
// 数据库健康检查
if err := database.HealthCheck(); err != nil {
    log.Printf("数据库健康检查失败: %v", err)
}

// Redis健康检查
if err := cache.HealthCheck(); err != nil {
    log.Printf("Redis健康检查失败: %v", err)
}

// 获取连接池统计
stats := database.GetConnectionStats()
fmt.Printf("数据库连接池状态: %+v\n", stats)
```

### 性能监控

```go
// 获取缓存统计
cacheStats := cache.GetCacheStats()
fmt.Printf("缓存统计: %+v\n", cacheStats)

// 获取系统状态
systemStatus := database.Status()
fmt.Printf("系统状态: %+v\n", systemStatus)
```

---

## 注意事项

1. **错误处理**: 始终检查和处理返回的错误
2. **资源释放**: 使用defer确保资源正确释放
3. **上下文传递**: 在可能的情况下传递context进行超时控制
4. **并发安全**: 在并发环境下使用适当的锁机制
5. **性能考虑**: 合理设置连接池参数和缓存TTL
6. **安全考虑**: 验证输入参数，防止SQL注入和缓存穿透

---

## 版本信息

- **版本**: v1.0.0
- **最后更新**: 2024年
- **Go版本**: 1.23+
- **兼容性**: 兼容MySQL 8.0+, Redis 7.0+