# 配置系统使用指南

## 🏗️ 分层配置架构

### 配置加载优先级（从低到高）
```
1. 默认配置 (config.yaml) - 通用基础配置
2. 环境配置 (config.dev.yaml) - 环境差异配置  
3. 环境变量文件 (.env.dev) - 敏感信息
4. 系统环境变量 (CLOUDPAN_*) - 最高优先级
```

## 📁 配置文件结构

### 默认配置 (config.yaml)
```yaml
# 只包含各环境通用的基础配置
app:
  name: "cloudpan"
  version: "1.0.0"
  
# 通用业务规则
user:
  password:
    min_length: 8
    require_number: true
```

### 环境配置 (config.dev.yaml)
```yaml
# 只包含开发环境特有的差异
app:
  env: "development"
  debug: true

server:
  host: "0.0.0.0"
  port: 8080
```

### 环境变量文件 (.env.dev)
```bash
# 敏感信息，不提交到Git
CLOUDPAN_DATABASE_MYSQL_USERNAME=cloudpan_dev
CLOUDPAN_DATABASE_MYSQL_PASSWORD=dev_password_123
CLOUDPAN_JWT_SECRET=dev-jwt-secret-key
```

## 💻 在代码中使用配置

### 1. 加载配置
```go
import "cloudpan/internal/pkg/config"

func main() {
    // 加载配置
    if err := config.Load(); err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // 获取配置
    cfg := config.GetConfig()
}
```

### 2. 类型安全的配置访问
```go
// 数据库配置
dsn := config.GetDSN()
dbHost := config.GetConfig().Database.MySQL.Host

// 服务器配置  
serverAddr := config.GetServerAddr()
port := config.GetConfig().Server.Port

// JWT配置
secret := config.GetConfig().JWT.Secret
expireHours := config.GetConfig().JWT.ExpireHours

// 环境判断
if config.IsDevelopment() {
    // 开发环境逻辑
}
```

### 3. 配置验证
```go
// 系统启动时自动验证
// 必填字段检查
// 数据类型验证
// 取值范围检查
```

## 🔧 环境变量规范

### 命名规则
```bash
# 格式：CLOUDPAN_配置路径（用下划线分隔）
CLOUDPAN_DATABASE_MYSQL_HOST=localhost
CLOUDPAN_DATABASE_MYSQL_PORT=3306
CLOUDPAN_JWT_SECRET=your-secret
CLOUDPAN_STORAGE_LOCAL_ROOT_PATH=/data/storage
```

### 敏感信息管理
```bash
# 开发环境 (.env.dev)
CLOUDPAN_DATABASE_MYSQL_PASSWORD=dev_password
CLOUDPAN_JWT_SECRET=dev-secret

# 生产环境 (系统环境变量)
export CLOUDPAN_DATABASE_MYSQL_PASSWORD=prod_password
export CLOUDPAN_JWT_SECRET=super-secret-key
```

## 🌍 环境切换

### 开发环境
```bash
# 方式1：环境变量
export GO_ENV=development
./cloudpan

# 方式2：PowerShell
$env:GO_ENV="development"
.\cloudpan.exe
```

### 生产环境
```bash
export GO_ENV=production
export CLOUDPAN_DATABASE_MYSQL_PASSWORD=prod_password
export CLOUDPAN_JWT_SECRET=production-secret
./cloudpan
```

## 🔐 安全最佳实践

### 1. 敏感信息隔离
- ❌ 配置文件中不存储密码、密钥
- ✅ 使用环境变量或配置中心
- ✅ .env文件添加到.gitignore

### 2. 配置验证
- ✅ 启动时验证必填配置
- ✅ 验证JWT密钥长度（≥32字符）
- ✅ 验证端口范围（1-65535）

### 3. 环境隔离
- ✅ 不同环境使用不同配置文件
- ✅ 生产环境禁用调试功能
- ✅ 开发环境降低安全要求（便于调试）

## ⚡ 配置热重载

```go
// TODO: 后续可以添加配置热重载功能
// 监听配置文件变化，动态重新加载配置
```

## 🚀 使用示例

### main.go中的用法
```go
package main

import (
    "log"
    "cloudpan/internal/pkg/config"
    "github.com/gin-gonic/gin"
)

func main() {
    // 加载配置
    if err := config.Load(); err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    cfg := config.GetConfig()
    
    // 根据环境设置Gin模式
    if config.IsProduction() {
        gin.SetMode(gin.ReleaseMode)
    }
    
    // 创建路由
    r := gin.Default()
    
    // 启动服务器
    addr := config.GetServerAddr()
    if err := r.Run(addr); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

这样的配置系统实现了：
- ✅ 类型安全的配置访问
- ✅ 分层配置管理  
- ✅ 敏感信息隔离
- ✅ 环境变量支持
- ✅ 配置验证
- ✅ 开发友好的.env文件支持