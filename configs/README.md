# configs 目录

## 目录说明
配置文件目录，存放不同环境的配置文件。

## 功能描述
- 存放YAML配置文件
- 多环境配置管理
- 配置模板文件
- 默认配置值
- 环境变量支持
- 配置验证和热重载

## 文件结构
```
configs/
├── config.yaml         # 默认配置文件（包含所有配置项）
├── config.dev.yaml     # 开发环境配置（覆盖默认值）
├── config.test.yaml    # 测试环境配置（用于单元测试）
├── config.prod.yaml    # 生产环境配置（使用环境变量）
└── config.example.yaml # 配置模板文件（部署参考）
```

## 配置内容
### 核心配置
- **应用配置**: 应用名称、版本、环境、调试模式
- **服务器配置**: 监听地址、端口、超时设置
- **数据库配置**: MySQL连接、连接池、字符集等
- **Redis配置**: 连接信息、连接池、协议版本

### 业务配置
- **JWT配置**: 密钥、过期时间、签发者
- **存储配置**: 本地存储、OSS云存储、自动切换策略
- **用户配置**: 存储配额、头像设置、密码策略
- **邮件配置**: SMTP服务器、模板、验证码设置

### 安全配置
- **CORS配置**: 跨域设置、允许的域名和方法
- **限流配置**: 请求频率限制、突发请求处理
- **病毒扫描**: ClamAV集成、扫描超时设置

### 运维配置
- **日志配置**: 日志级别、格式、轮转策略
- **缓存配置**: TTL设置、不同类型缓存策略
- **监控配置**: 指标收集、健康检查、性能分析
- **消息队列**: Redis Stream配置、队列名称

## 使用说明

### 环境选择
配置文件根据 `GO_ENV` 环境变量自动选择：
- `development` -> config.dev.yaml
- `testing` -> config.test.yaml
- `production` -> config.prod.yaml
- 默认 -> config.yaml

### 环境变量覆盖
支持使用环境变量覆盖配置文件中的值：
```bash
# 直接使用环境变量
DB_HOST=localhost
DB_PORT=3306
JWT_SECRET=your-secret-key

# 或使用CLOUDPAN前缀
CLOUDPAN_DATABASE_MYSQL_HOST=localhost
CLOUDPAN_JWT_SECRET=your-secret-key
```

### 配置验证
系统启动时会自动验证配置的有效性：
- 必填字段检查
- 数据类型验证
- 取值范围检查
- 依赖关系验证

### 配置热重载
支持在运行时重新加载配置文件，无需重启服务。

## 安全注意事项
1. **生产环境**: 敏感信息（密码、密钥）必须使用环境变量
2. **权限控制**: 配置文件应设置适当的文件权限（600或644）
3. **版本控制**: 不要将包含敏感信息的配置文件提交到Git
4. **密钥管理**: JWT密钥长度至少32字符，生产环境使用随机生成

## 部署指南
1. 复制 `config.example.yaml` 到目标环境
2. 根据实际环境修改配置值
3. 设置环境变量覆盖敏感配置
4. 验证配置有效性
5. 启动应用服务

## 配置示例

### 开发环境启动
```bash
export GO_ENV=development
export DB_PASSWORD=dev_password
./cloudpan
```

### 生产环境启动
```bash
export GO_ENV=production
export DB_HOST=prod-db-server
export DB_PASSWORD=prod_password
export JWT_SECRET=super-secret-key
./cloudpan
```