# Utils Package

基础工具函数库，提供字符串处理、时间处理、HTTP响应格式化等常用功能。

## 功能模块

### string.go - 字符串处理工具
- **随机字符串生成**: 支持多种字符集的随机字符串生成
- **字符串转换**: 驼峰、蛇形、帕斯卡命名转换
- **字符串验证**: 邮箱、用户名格式验证
- **字符串操作**: 截断、填充、反转等
- **数据脱敏**: 邮箱、电话号码脱敏
- **编码转换**: Base64、十六进制编码

### time.go - 时间处理工具
- **时间格式化**: 支持多种时间格式的格式化和解析
- **时区处理**: 时区转换和时区信息获取
- **时间计算**: 时间差计算、工作日计算
- **时间判断**: 今天、昨天、周末等时间判断
- **人性化显示**: 多少时间前、还有多长时间等

### response.go - HTTP响应工具
- **统一响应格式**: 标准化的API响应结构
- **错误码管理**: 完整的业务错误码定义
- **分页支持**: 标准分页信息和响应
- **响应封装**: 成功、错误、列表等响应的快速封装

## 使用示例

### 字符串工具使用

```go
import "cloudpan/internal/pkg/utils"

// 生成随机字符串
randomStr, _ := utils.GenerateAlphanumeric(10) // 生成10位字母数字字符串
token, _ := utils.GenerateSecureToken(32)      // 生成安全令牌

// 字符串转换
camelCase := utils.ToCamelCase("hello_world")     // helloWorld
snakeCase := utils.ToSnakeCase("HelloWorld")      // hello_world
pascalCase := utils.ToPascalCase("hello_world")   // HelloWorld

// 字符串验证
isValid := utils.IsValidEmail("user@example.com") // true
isValid = utils.IsValidUsername("user123")        // true

// 字符串操作
truncated := utils.TruncateWithEllipsis("很长的字符串", 10) // 很长的字符串...
padded := utils.PadLeft("123", 6, '0')                   // 000123

// 数据脱敏
masked := utils.MaskEmail("user@example.com")     // u***@example.com
masked = utils.MaskPhone("13800138000")           // 138****8000

// 类型转换
intVal := utils.StringToInt("123", 0)             // 123
boolVal := utils.StringToBool("true", false)      // true
```

### 时间工具使用

```go
import "cloudpan/internal/pkg/utils"

// 时间格式化
now := time.Now()
dateStr := utils.FormatDate(now)                    // 2024-01-01
dateTimeStr := utils.FormatDateTime(now)            // 2024-01-01 15:04:05
chineseDate := utils.FormatChineseDate(now)         // 2024年01月01日

// 时间解析
date, _ := utils.ParseDate("2024-01-01")
dateTime, _ := utils.ParseDateTime("2024-01-01 15:04:05")
parsed, _ := utils.TryParseTime("2024/1/1 3:4:5")   // 尝试多种格式解析

// 时区处理
beijingTime := utils.ToBeijingTime(now)
utcTime := utils.ToUTC(now)

// 时间判断
isToday := utils.IsToday(now)                       // true
isWeekend := utils.IsWeekend(now)                   // 根据具体日期
isSame := utils.IsSameDay(time1, time2)

// 时间计算
startOfDay := utils.StartOfDay(now)                 // 当天开始时间
endOfMonth := utils.EndOfMonth(now)                 // 当月结束时间
daysBetween := utils.DaysBetween(start, end)       // 计算天数差

// 人性化显示
timeAgo := utils.TimeAgo(lastWeek)                  // 7天前
timeUntil := utils.TimeUntil(tomorrow)              // 1天

// 工作日计算
nextWorkday := utils.NextBusinessDay(now)          // 下一个工作日
workdayAfter := utils.AddBusinessDays(now, 5)      // 5个工作日后
```

### HTTP响应工具使用

```go
import (
    "cloudpan/internal/pkg/utils"
    "github.com/gin-gonic/gin"
)

func UserController(c *gin.Context) {
    // 成功响应
    user := &User{ID: 1, Name: "张三"}
    utils.Success(c, user)
    
    // 自定义成功消息
    utils.SuccessWithMessage(c, "用户创建成功", user)
    
    // 错误响应
    utils.Error(c, utils.CodeValidationError)
    utils.ErrorWithMessage(c, utils.CodeNotFound, "用户不存在")
    
    // 常用错误响应
    utils.ValidationError(c, validationErrors)
    utils.Unauthorized(c)
    utils.Forbidden(c)
    utils.NotFound(c)
    utils.InternalError(c)
    
    // 列表响应
    users := []User{...}
    pagination := utils.NewPagination(1, 20, 100)
    utils.SuccessList(c, users, pagination)
    
    // 创建/更新/删除响应
    utils.Created(c, user)
    utils.Updated(c, user)
    utils.Deleted(c)
}

func ListUsers(c *gin.Context) {
    // 解析分页参数
    pageReq := utils.ParsePageRequest(c)
    
    // 验证排序字段
    allowedFields := []string{"id", "name", "created_at"}
    if !pageReq.ValidateSortField(allowedFields) {
        utils.ValidationError(c, "无效的排序字段")
        return
    }
    
    // 数据库查询
    offset := pageReq.GetOffset()
    limit := pageReq.GetLimit()
    orderBy := pageReq.GetOrderBy()
    
    users, total := getUsersFromDB(offset, limit, orderBy)
    pagination := utils.NewPagination(pageReq.Page, pageReq.PageSize, total)
    
    utils.SuccessList(c, users, pagination)
}
```

## 错误码定义

### HTTP状态码映射
- `200`: 成功
- `400-499`: 客户端错误
- `500-599`: 服务端错误

### 自定义业务错误码
- `1001`: 数据验证失败
- `1002`: 数据重复  
- `1003`: 数据不存在
- `1004`: 操作失败
- `1005`: 配额超出
- `1006-1014`: 认证相关错误
- `1015-1019`: 文件相关错误
- `1020-1023`: 系统相关错误

## 响应格式

### 标准响应
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {...},
  "request_id": "req-123456",
  "timestamp": 1704067200
}
```

### 列表响应
```json
{
  "code": 200,
  "message": "操作成功",
  "data": [...],
  "pagination": {
    "current_page": 1,
    "page_size": 20,
    "total_count": 100,
    "total_pages": 5,
    "has_previous": false,
    "has_next": true,
    "next_page": 2
  },
  "request_id": "req-123456",
  "timestamp": 1704067200
}
```

### 错误响应
```json
{
  "code": 1001,
  "message": "数据验证失败",
  "data": {
    "field_errors": {...}
  },
  "request_id": "req-123456",
  "timestamp": 1704067200
}
```

## 最佳实践

1. **字符串处理**:
   - 生成敏感数据使用`GenerateSecureToken`
   - 处理用户输入前先验证格式
   - 存储前进行适当的清理和转换

2. **时间处理**:
   - 统一使用UTC时间存储
   - 显示时转换为用户时区
   - 避免直接字符串拼接时间格式

3. **HTTP响应**:
   - 统一使用响应工具函数
   - 错误码与HTTP状态码正确映射
   - 分页参数要进行验证和限制

4. **性能考虑**:
   - 大量字符串操作使用`strings.Builder`
   - 时间格式化缓存常用格式
   - 响应数据避免过度嵌套

## 文件组织

```
utils/
├── string.go      # 字符串处理工具
├── time.go        # 时间处理工具  
├── response.go    # HTTP响应工具
└── README.md      # 说明文档
```