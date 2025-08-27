# configs 目录

## 目录说明
配置文件目录，存放不同环境的配置文件。

## 功能描述
- 存放YAML/JSON配置文件
- 多环境配置管理
- 配置模板文件
- 默认配置值

## 文件结构
```
configs/
├── config.yaml        # 默认配置文件
├── config.dev.yaml    # 开发环境配置
├── config.test.yaml   # 测试环境配置
├── config.prod.yaml   # 生产环境配置
└── config.example.yaml # 配置模板文件
```

## 配置内容
- 数据库连接配置
- Redis连接配置
- 服务器端口配置
- 日志级别配置
- OSS存储配置
- 第三方服务配置

## 使用说明
- 根据环境变量自动选择配置文件
- 支持配置参数验证
- 支持配置热重载