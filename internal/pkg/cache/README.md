# Cache 缓存模块

本模块提供Redis缓存系统的封装，实现了完整的缓存操作、键管理和TTL策略。

## 🚀 功能特性

### ✅ 已实现功能

- **Redis连接管理**：支持连接池、健康检查、连接统计
- **缓存操作封装**：提供完整的缓存CRUD操作
- **键命名规范**：统一的缓存键命名标准和构建器
- **TTL管理**：智能的缓存过期时间管理
- **数据结构支持**：String、Hash、Set、ZSet等Redis数据结构
- **批量操作**：支持批量设置和删除操作
- **类型安全**：JSON序列化/反序列化支持
- **错误处理**：完善的错误定义和处理机制

## 📁 文件结构

```
cache/
├── redis.go        # Redis连接管理
├── manager.go      # 缓存操作管理器
├── keys.go         # 缓存键命名规范
├── ttl.go          # TTL管理和缓存包装器
├── cache_test.go   # 完整的单元测试
└── README.md       # 本文档
```

## ⚙️ 使用方法

### 1. 初始化Redis连接
```go
if err := cache.InitRedis(); err != nil {
    log.Fatal(err)
}
defer cache.CloseRedis()
```

### 2. 基础缓存操作
```go
manager := cache.NewCacheManager()

// 设置缓存
err := manager.Set("user:123", userData)

// 获取缓存
var user User
err = manager.Get("user:123", &user)
```

### 3. 使用缓存包装器
```go
// 用户会话管理（自动TTL）
err := cache.Cache.SetUserSession(token, sessionData)

// 文件信息缓存
err = cache.Cache.SetFileInfo(fileID, fileInfo)

// 限流控制
count, err := cache.Cache.IncrementRateLimit(ip, endpoint)
```