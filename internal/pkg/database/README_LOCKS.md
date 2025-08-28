# 数据库同步锁机制使用文档

## 概述

本模块实现了完整的数据库同步锁机制，包括：

- **事务隔离级别控制**：支持MySQL的4种隔离级别
- **数据库锁机制**：悲观锁（行锁、表锁）和乐观锁（版本控制）
- **Redis分布式锁**：支持跨进程、跨服务器的分布式锁
- **并发控制管理器**：统一的并发控制接口

## 主要组件

### 1. 事务管理器 (TransactionManager)

#### 功能特性
- 支持4种事务隔离级别
- 只读事务支持
- 自定义超时控制

#### 使用示例
```go
// 获取事务管理器
tm := database.NewTransactionManager(db)

// 使用特定隔离级别执行事务
err := tm.WithIsolationLevel(database.RepeatableRead, func(tx *gorm.DB) error {
    // 在REPEATABLE READ隔离级别下执行业务逻辑
    var user User
    tx.First(&user, 1)
    user.Name = "Updated Name"
    return tx.Save(&user).Error
})

// 只读事务
err := tm.WithReadOnlyTransaction(func(tx *gorm.DB) error {
    // 只读操作，性能更好
    var users []User
    return tx.Find(&users).Error
})
```

### 2. 数据库锁管理器 (DatabaseLockManager)

#### 悲观锁
```go
dlm := database.NewDatabaseLockManager(db)

// 在事务中获取悲观锁
err := db.Transaction(func(tx *gorm.DB) error {
    // 获取排他锁
    var user User
    err := dlm.PessimisticLockQuery(tx, &user, database.ExclusiveLock, "id = ?", userID)
    if err != nil {
        return err
    }
    
    // 执行需要锁保护的操作
    user.Balance += 100
    return tx.Save(&user).Error
})

// 手动指定锁定条件
err := database.Transaction(func(tx *gorm.DB) error {
    ctx := context.Background()
    // 锁定特定记录
    err := dlm.AcquirePessimisticLock(ctx, tx, "users", database.ExclusiveLock, "id = ?", userID)
    if err != nil {
        return err
    }
    
    // 执行业务逻辑
    return tx.Model(&User{}).Where("id = ?", userID).Update("status", "locked").Error
})
```

#### 乐观锁
```go
// 乐观锁更新
err := dlm.OptimisticLockUpdate(tx, &user, user.Version, map[string]interface{}{
    "name":   "New Name",
    "status": "active",
})
if err != nil {
    if strings.Contains(err.Error(), "optimistic lock conflict") {
        // 处理并发冲突
        return handleConcurrencyConflict(user)
    }
    return err
}

// 乐观锁删除
err := dlm.OptimisticLockDelete(tx, &user, user.Version)
```

### 3. Redis分布式锁 (RedisDistributedLock)

#### 基本用法
```go
// 创建锁管理器
rlm := database.NewRedisLockManager(redisClient)

// 创建分布式锁
lock, err := rlm.NewLock("user:123:update", 30*time.Second)
if err != nil {
    return err
}

// 尝试获取锁
ctx := context.Background()
acquired, err := lock.TryLock(ctx)
if err != nil {
    return err
}

if !acquired {
    return errors.New("resource is busy")
}

// 确保释放锁
defer lock.Unlock(ctx)

// 执行需要锁保护的操作
// ...
```

#### 阻塞获取锁
```go
// 阻塞等待直到获取锁
err := lock.Lock(ctx, 100*time.Millisecond) // 重试间隔
if err != nil {
    return err
}
defer lock.Unlock(ctx)

// 执行业务逻辑
```

#### 锁延期
```go
// 获取锁
acquired, err := lock.TryLock(ctx)
if !acquired {
    return errors.New("failed to acquire lock")
}

// 延长锁的过期时间
err = lock.Extend(ctx, 60*time.Second)
if err != nil {
    return err
}

// 带自动续期的锁
err = lock.LockWithAutoRenewal(ctx, 10*time.Second) // 每10秒续期一次
```

#### 检查锁状态
```go
// 检查锁是否仍然有效
isLocked, err := lock.IsLocked(ctx)
if err != nil {
    return err
}

if isLocked {
    log.Println("Lock is still active")
}
```

### 4. 并发控制管理器 (ConcurrencyControlManager)

#### 统一接口
```go
// 获取全局并发控制管理器
ccm := database.GetConcurrencyManager()

// 使用分布式锁执行操作
err := ccm.WithDistributedLock(ctx, "critical-section", 30*time.Second, func() error {
    // 需要分布式锁保护的操作
    return performCriticalOperation()
})

// 使用悲观锁执行事务
err := ccm.WithPessimisticLock(ctx, "users", database.ExclusiveLock, "id = ?", func(tx *gorm.DB) error {
    // 在悲观锁保护下的事务操作
    var user User
    tx.First(&user, userID)
    user.Balance += amount
    return tx.Save(&user).Error
}, userID)

// 使用乐观锁执行更新
err := ccm.WithOptimisticLock(&user, user.Version, map[string]interface{}{
    "status": "updated",
})
```

## 隔离级别说明

### READ UNCOMMITTED
- 最低隔离级别
- 可能出现脏读、不可重复读、幻读
- 性能最好，但数据一致性最差

### READ COMMITTED
- 避免脏读
- 可能出现不可重复读、幻读
- 适合大多数应用场景

### REPEATABLE READ (MySQL默认)
- 避免脏读、不可重复读
- 可能出现幻读（MySQL通过间隙锁避免）
- 平衡性能和一致性

### SERIALIZABLE
- 最高隔离级别
- 避免所有并发问题
- 性能最差，适合要求强一致性的场景

## 锁类型说明

### 共享锁 (LOCK IN SHARE MODE)
- 多个事务可以同时持有共享锁
- 防止其他事务修改数据
- 适合读多写少的场景

### 排他锁 (FOR UPDATE)
- 只有一个事务可以持有排他锁
- 防止其他事务读取或修改数据
- 适合需要独占访问的场景

## 最佳实践

### 1. 锁的选择原则
```go
// 选择合适的锁类型
func chooseAppropiateLock(operation string) {
    switch operation {
    case "read-heavy":
        // 使用共享锁或乐观锁
        dlm.PessimisticLockQuery(tx, model, database.SharedLock)
        
    case "write-heavy":
        // 使用排他锁
        dlm.PessimisticLockQuery(tx, model, database.ExclusiveLock)
        
    case "cross-service":
        // 使用分布式锁
        ccm.WithDistributedLock(ctx, lockKey, ttl, operation)
        
    case "high-concurrency":
        // 使用乐观锁
        dlm.OptimisticLockUpdate(tx, model, version, updates)
    }
}
```

### 2. 避免死锁
```go
// 按照固定顺序获取锁
func transferMoney(fromUserID, toUserID uint, amount float64) error {
    // 确保锁的获取顺序一致
    userIDs := []uint{fromUserID, toUserID}
    sort.Slice(userIDs, func(i, j int) bool {
        return userIDs[i] < userIDs[j]
    })
    
    return database.Transaction(func(tx *gorm.DB) error {
        for _, userID := range userIDs {
            var user User
            err := dlm.PessimisticLockQuery(tx, &user, database.ExclusiveLock, "id = ?", userID)
            if err != nil {
                return err
            }
        }
        
        // 执行转账逻辑
        return performTransfer(tx, fromUserID, toUserID, amount)
    })
}
```

### 3. 锁超时处理
```go
// 设置合理的锁超时时间
func withLockTimeout(operation func() error) error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    return ccm.WithDistributedLock(ctx, "resource-key", 60*time.Second, operation)
}
```

### 4. 错误处理
```go
func handleLockErrors(err error) error {
    switch {
    case strings.Contains(err.Error(), "optimistic lock conflict"):
        // 乐观锁冲突，可以重试
        return retryWithBackoff(operation)
        
    case strings.Contains(err.Error(), "context deadline exceeded"):
        // 超时，记录日志并返回用户友好错误
        log.Error("Lock acquisition timeout")
        return errors.New("operation timed out, please try again")
        
    case strings.Contains(err.Error(), "failed to acquire lock"):
        // 锁获取失败，资源忙
        return errors.New("resource is busy, please try again later")
        
    default:
        return err
    }
}
```

## 性能考虑

### 1. 锁粒度
- **表锁**：影响整个表，并发性差但实现简单
- **行锁**：只影响特定行，并发性好但可能导致死锁
- **页锁**：介于表锁和行锁之间

### 2. 锁持有时间
```go
// 尽量缩短锁持有时间
func optimizeLockDuration() error {
    lock, err := rlm.NewLock("resource", 30*time.Second)
    if err != nil {
        return err
    }
    
    ctx := context.Background()
    if err := lock.Lock(ctx, 100*time.Millisecond); err != nil {
        return err
    }
    
    // 在锁外准备数据
    data := prepareData()
    
    // 获取锁后快速执行
    defer lock.Unlock(ctx)
    return quickUpdate(data)
}
```

### 3. 监控和调试
```go
// 添加锁监控
func monitorLockPerformance() {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        if duration > 1*time.Second {
            log.Warnf("Lock operation took %v", duration)
        }
    }()
    
    // 执行锁操作
}
```

## 配置说明

### Redis配置
```yaml
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
  pool_timeout: 4s
  idle_timeout: 300s
```

### 锁相关配置建议
```go
const (
    // 分布式锁默认TTL
    DefaultLockTTL = 30 * time.Second
    
    // 锁重试间隔
    LockRetryInterval = 100 * time.Millisecond
    
    // 锁自动续期间隔
    LockRenewalInterval = 10 * time.Second
    
    // 最大锁等待时间
    MaxLockWaitTime = 60 * time.Second
)
```

通过这套完整的数据库同步锁机制，可以有效解决高并发场景下的数据一致性问题，确保系统的稳定性和可靠性。