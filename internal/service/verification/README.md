# 验证码服务 (Verification Service)

验证码服务提供完整的验证码管理功能，支持密码重置、用户注册、登录验证等多种场景。

## 功能特性

### 1. 验证码生成
- **邮箱验证码**: 支持6位数字验证码生成
- **多种类型**: 注册、登录、密码重置、邮箱变更等
- **安全加密**: 使用盐值+SHA256哈希存储
- **过期管理**: 不同类型验证码支持不同过期时间

### 2. 验证码验证
- **格式验证**: 严格的6位数字格式检查
- **内容验证**: 哈希对比验证验证码正确性
- **状态检查**: 过期、使用状态、尝试次数检查
- **防暴力破解**: 最大尝试次数限制

### 3. 安全防护
- **频率限制**: 
  - 同一邮箱: 5分钟内最多3次
  - 同一IP: 1小时内最多10次
- **尝试限制**: 每个验证码最多尝试5次
- **自动失效**: 新验证码生成时旧验证码自动失效

### 4. 邮件集成
- **模板邮件**: 支持不同类型的邮件模板
- **异步发送**: 邮件发送失败不影响验证码生成
- **国际化**: 支持多语言邮件模板

## 使用示例

### 密码重置流程

```go
// 1. 生成密码重置验证码
verificationService := verification.NewVerificationService(db, emailService, logger)

code, err := verificationService.GeneratePasswordResetCode(
    ctx, 
    "user@example.com", 
    userID, 
    request.RemoteAddr,
)
if err != nil {
    // 处理错误
}

// 2. 验证密码重置验证码
verifiedCode, err := verificationService.VerifyPasswordResetCode(
    ctx,
    "user@example.com",
    userInputCode,
)
if err != nil {
    // 验证失败
}

// 3. 完成密码重置
err = verificationService.CompletePasswordReset(ctx, verifiedCode.ID)
```

### 邮箱验证流程

```go
// 1. 生成邮箱验证码
code, err := verificationService.GenerateEmailVerificationCode(
    ctx,
    "user@example.com",
    userID,
    request.RemoteAddr,
)

// 2. 验证邮箱验证码
verifiedCode, err := verificationService.VerifyEmailVerificationCode(
    ctx,
    "user@example.com",
    userInputCode,
)
```

## 验证码类型

| 类型 | 常量 | 过期时间 | 用途 |
|------|------|----------|------|
| 注册验证 | `VerificationTypeRegister` | 15分钟 | 用户注册邮箱验证 |
| 登录验证 | `VerificationTypeLogin` | 5分钟 | 安全登录验证 |
| 密码重置 | `VerificationTypeResetPassword` | 30分钟 | 忘记密码重置 |
| 邮箱变更 | `VerificationTypeChangeEmail` | 15分钟 | 修改邮箱地址 |

## 安全考虑

### 1. 存储安全
- 验证码不以明文存储
- 使用随机盐值增强安全性
- SHA256哈希算法加密

### 2. 防攻击措施
- 频率限制防止短信轰炸
- 尝试次数限制防止暴力破解
- IP级别限制防止恶意请求

### 3. 数据清理
- 自动清理过期验证码
- 定期清理已使用验证码
- 支持批量清理用户验证码

## 配置要求

### 1. 数据库表
确保 `verification_codes` 表已创建，包含以下字段：
- `id`, `uuid`, `target`, `type`
- `code_hash`, `salt`, `expires_at`
- `is_used`, `used_at`, `attempt_count`
- `ip_address`, `user_id`

### 2. 邮件服务
需要配置邮件服务支持：
- SMTP服务器配置
- 邮件模板配置
- 发送频率限制

### 3. 日志配置
推荐配置结构化日志：
- 验证码生成日志
- 验证尝试日志
- 安全事件日志

## 错误处理

服务使用统一的错误类型：
- `ValidationError`: 参数验证错误
- `InternalError`: 内部服务错误
- `RateLimitError`: 频率限制错误

## 性能优化

### 1. 数据库索引
- `target + type + is_used + expires_at` 复合索引
- `ip_address + created_at` 索引
- `user_id` 索引

### 2. 缓存策略
- 频率限制信息缓存
- 活跃验证码缓存
- IP黑名单缓存

### 3. 定期清理
建议设置定时任务清理过期数据：
```go
// 每小时清理一次过期验证码
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        verificationService.CleanupExpiredCodes(context.Background())
    }
}()
```