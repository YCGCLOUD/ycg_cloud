# docker 目录

## 目录说明
Docker容器化配置目录，包含Docker相关的配置文件。

## 功能描述
- Dockerfile配置
- Docker Compose配置
- 容器环境配置
- 多阶段构建配置

## 文件结构
```
docker/
├── Dockerfile          # 应用容器配置
├── docker-compose.yaml # 服务编排配置
├── docker-compose.dev.yaml  # 开发环境配置
├── docker-compose.prod.yaml # 生产环境配置
└── nginx/              # Nginx配置
    └── nginx.conf
```

## 容器服务
- **应用容器**: Go应用程序
- **数据库容器**: MySQL 8.4
- **缓存容器**: Redis 7.2
- **代理容器**: Nginx反向代理
- **存储容器**: MinIO对象存储

## 配置特性
- 多阶段构建优化镜像大小
- 环境变量配置
- 数据卷持久化
- 网络隔离
- 健康检查

## 使用说明
- 支持一键启动开发环境
- 生产环境优化配置
- 自动化部署支持