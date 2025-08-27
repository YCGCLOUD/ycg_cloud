# 个人版网络云盘后端项目

## 项目概述
基于Go语言开发的个人版网络云盘后端系统，提供文件管理、用户体系、团队协作、即时通讯等完整功能。

## 技术栈
- **编程语言**: Go 1.23+
- **Web框架**: Gin 1.10+
- **ORM框架**: GORM 1.25+
- **数据库**: MySQL 8.4+
- **缓存**: Redis 7.2+
- **消息队列**: Redis Stream
- **认证**: JWT 5.2+
- **对象存储**: MinIO/阿里云OSS
- **文件扫描**: ClamAV
- **国际化**: go-i18n

## 项目结构
```
cloudpan/
├── cmd/                    # 应用入口
├── internal/               # 内部模块
│   ├── api/               # API层
│   ├── service/           # 业务逻辑层
│   ├── repository/        # 数据访问层
│   └── pkg/               # 工具包
├── configs/                # 配置文件
├── docs/                   # 接口文档
├── scripts/                # 脚本文件
└── docker/                 # Docker配置
```

## 核心功能
- 📁 **文件管理**: 上传下载、分片续传、秒传、预览、版本管理
- 👥 **用户体系**: 认证授权、权限管理、用户信息
- 🤝 **团队协作**: 团队管理、文件共享、权限控制
- 💬 **即时通讯**: WebSocket、消息推送、媒体消息
- 🗑️ **回收站**: 文件恢复、自动清理、空间管理
- 💾 **存储管理**: 本地存储、OSS存储、存储策略

## 快速开始
```bash
# 克隆项目
git clone <repository-url>
cd cloudpan

# 初始化go.mod
go mod init cloudpan

# 安装依赖
go mod tidy

# 启动应用
go run cmd/main.go
```

## 开发规范
- 严格按照分层架构开发
- 遵循Go代码规范
- 单元测试覆盖率≥80%
- API文档自动生成
- 代码审查100%覆盖

## 部署方式
- Docker容器化部署
- docker-compose一键启动
- 支持多环境配置
- 自动化CI/CD流程