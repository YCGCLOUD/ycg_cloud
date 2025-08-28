# database 目录

## 目录说明
数据库连接管理模块，负责MySQL连接池配置、连接管理和健康检查。

## 功能描述
- MySQL连接池初始化和配置
- 数据库连接生命周期管理
- 连接池性能优化
- 数据库健康检查
- 连接池统计信息

## 主要文件
- **mysql.go** - MySQL连接池实现和配置管理

## 核心功能

### 1. 连接池配置
- **最大连接数控制**：防止连接数过多导致数据库压力
- **空闲连接管理**：保持适量空闲连接提高响应速度
- **连接生命周期**：自动管理连接的创建和销毁
- **连接超时设置**：防止长时间无效连接

### 2. 性能优化
- **预编译语句缓存**：提高SQL执行效率
- **跳过默认事务**：减少不必要的事务开销
- **连接复用**：减少连接建立的开销

### 3. 监控和诊断
- **健康检查**：提供数据库连接状态检查
- **连接池统计**：实时监控连接池使用情况
- **连接测试**：验证数据库连接的有效性

## 配置参数

### MySQL连接池配置
```yaml
database:
  mysql:
    host: "localhost"                    # 数据库主机
    port: 3306                          # 数据库端口
    username: "username"                # 用户名
    password: "password"                # 密码
    dbname: "database_name"             # 数据库名
    charset: "utf8mb4"                  # 字符集
    parse_time: true                    # 解析时间类型
    loc: "Local"                        # 时区
    max_idle_conns: 10                  # 最大空闲连接数
    max_open_conns: 100                 # 最大打开连接数
    conn_max_lifetime: "1h"             # 连接最大生存时间
    conn_max_idle_time: "10m"           # 连接最大空闲时间
    timezone: "+08:00"                  # 数据库时区
```

## 使用方法

### 初始化连接池
```go
import "cloudpan/internal/pkg/database"

// 初始化数据库连接池
if err := database.InitMySQL(); err != nil {
    log.Fatal("Failed to initialize MySQL:", err)
}
```

### 获取数据库连接
```go
// 获取GORM数据库实例
db := database.GetDB()

// 执行数据库操作
var user User
db.First(&user, 1)
```

### 健康检查
```go
// 检查数据库连接健康状态
if err := database.HealthCheck(); err != nil {
    log.Printf("Database health check failed: %v", err)
}
```

### 获取连接池统计
```go
// 获取连接池统计信息
stats := database.GetConnectionStats()
fmt.Printf("Connection pool stats: %+v\n", stats)
```

### 关闭连接
```go
// 应用程序退出时关闭数据库连接
defer func() {
    if err := database.Close(); err != nil {
        log.Printf("Failed to close database: %v", err)
    }
}()
```

## 设计原则

### 1. 连接池优化
- **默认连接数**：最大100个连接，空闲10个连接
- **连接生命周期**：1小时最大生存时间，10分钟最大空闲时间
- **性能调优**：启用预编译语句缓存，跳过默认事务

### 2. 兼容性保证
- **MySQL 8.0.31支持**：添加`allowNativePasswords=true`参数
- **字符集配置**：使用`utf8mb4`支持完整Unicode
- **时区处理**：支持自定义时区配置

### 3. 监控和诊断
- **连接状态监控**：实时跟踪连接池使用情况
- **健康检查接口**：提供外部监控集成
- **详细日志**：记录连接建立和配置信息

## 错误处理

### 常见错误和解决方案
1. **连接失败**：检查主机、端口、用户名、密码配置
2. **连接池耗尽**：调整`max_open_conns`参数或优化查询
3. **连接超时**：检查网络连接或调整超时参数
4. **字符集问题**：确保使用`utf8mb4`字符集

### 监控指标
- `open_connections`：当前打开的连接数
- `in_use`：正在使用的连接数
- `idle`：空闲连接数
- `wait_count`：等待连接的请求数
- `wait_duration`：平均等待时间

## 注意事项
1. 确保在应用启动时调用`InitMySQL()`初始化连接池
2. 应用退出时调用`Close()`优雅关闭连接
3. 定期监控连接池统计信息，优化配置参数
4. 开发环境可启用详细SQL日志，生产环境建议关闭