# Logger Package

基于Zap的高性能结构化日志系统，支持访问日志和应用日志的分离管理。

## 功能特性

- **高性能**: 基于Uber Zap日志库，零内存分配的结构化日志
- **多输出**: 支持控制台、文件、或两者同时输出
- **日志轮转**: 基于文件大小、时间、备份数量的自动轮转
- **结构化**: JSON和Console两种格式支持
- **上下文支持**: 自动提取请求ID、用户ID等上下文信息
- **分离管理**: 应用日志和访问日志独立配置和存储

## 基本使用

### 初始化日志系统

```go
import "cloudpan/internal/pkg/logger"

// 配置应用日志
logConfig := logger.LogConfig{
    Level:      "info",
    Format:     "json",
    Output:     "both",
    FilePath:   "logs/app.log",
    MaxSize:    100,    // 100MB
    MaxAge:     30,     // 30天
    MaxBackups: 5,      // 5个备份文件
    Compress:   true,
}

// 初始化
err := logger.InitLogger(logConfig)
if err != nil {
    panic(err)
}

// 配置访问日志
accessConfig := logger.AccessLogConfig{
    Enabled:  true,
    FilePath: "logs/access.log",
    Format:   "json",
}

err = logger.InitAccessLogger(accessConfig)
if err != nil {
    panic(err)
}
```

### 基本日志记录

```go
import (
    "go.uber.org/zap"
    "cloudpan/internal/pkg/logger"
)

// 使用全局Logger
logger.Info("用户登录成功", 
    zap.String("user_id", "12345"),
    zap.String("ip", "192.168.1.1"),
)

logger.Error("数据库连接失败", 
    zap.Error(err),
    zap.String("database", "mysql"),
)

// 使用SugaredLogger（支持格式化）
logger.SugaredLogger.Infof("用户 %s 上传了文件 %s", userID, filename)
```

### 上下文日志

```go
// 在HTTP Handler中使用
func HandleLogin(c *gin.Context) {
    ctx := c.Request.Context()
    
    // 添加请求ID到上下文
    ctx = context.WithValue(ctx, logger.RequestIDKey, "req-12345")
    
    // 从上下文创建Logger
    contextLogger := logger.WithContext(ctx)
    contextLogger.Info("处理登录请求",
        zap.String("username", username),
    )
    
    // 用户认证成功后添加用户ID
    userLogger := logger.WithUserID(ctx, userID)
    userLogger.Info("用户登录成功")
}
```

### 访问日志

```go
// 记录HTTP请求访问日志
entry := logger.AccessLogEntry{
    Timestamp:    time.Now(),
    RequestID:    "req-12345",
    UserID:       "user-789",
    Method:       "POST",
    Path:         "/api/v1/login",
    StatusCode:   200,
    ResponseTime: 150, // 毫秒
    IPAddress:    "192.168.1.100",
    UserAgent:    "CloudDisk/1.0.0",
}

logger.LogAccess(entry)
```

## 配置参数

### 应用日志配置 (LogConfig)

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| Level | string | 日志级别 (debug/info/warn/error/panic/fatal) | info |
| Format | string | 日志格式 (json/console) | json |
| Output | string | 输出方式 (file/console/both) | both |
| FilePath | string | 日志文件路径 | logs/app.log |
| MaxSize | int | 最大文件大小(MB) | 100 |
| MaxAge | int | 最大保留天数 | 30 |
| MaxBackups | int | 最大备份文件数 | 5 |
| Compress | bool | 是否压缩备份文件 | true |

### 访问日志配置 (AccessLogConfig)

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| Enabled | bool | 是否启用访问日志 | true |
| FilePath | string | 访问日志文件路径 | logs/access.log |
| Format | string | 日志格式 | json |

## 日志级别

- **Debug**: 详细的调试信息
- **Info**: 一般信息，如用户操作、系统状态
- **Warn**: 警告信息，可能的问题但不影响正常运行
- **Error**: 错误信息，需要关注但不会导致程序崩溃
- **Panic**: 严重错误，程序将panic
- **Fatal**: 致命错误，程序将退出

## 最佳实践

1. **合理使用日志级别**: 开发环境使用Debug，生产环境使用Info或Warn
2. **结构化字段**: 使用zap.Field而不是格式化字符串
3. **敏感信息**: 避免记录密码、令牌等敏感信息
4. **性能考虑**: 在高频调用的地方考虑使用Check()方法
5. **上下文传递**: 在请求处理链中传递上下文信息

## 文件组织

```
logger/
├── logger.go      # 主日志系统
├── access.go      # 访问日志管理
└── README.md      # 说明文档
```

## 依赖

- go.uber.org/zap: 高性能日志库
- gopkg.in/natefinch/lumberjack.v2: 日志轮转库