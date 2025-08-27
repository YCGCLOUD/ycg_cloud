# docs 目录

## 目录说明
项目文档目录，包含API文档和开发文档。

## 功能描述
- API接口文档
- 开发指南
- 部署文档
- 架构设计文档

## 文件结构
```
docs/
├── api/               # API文档
│   ├── swagger.yaml   # Swagger文档
│   └── postman/       # Postman集合
├── dev/               # 开发文档
│   ├── setup.md       # 环境搭建
│   └── coding.md      # 编码规范
└── deploy/            # 部署文档
    ├── docker.md      # Docker部署
    └── production.md  # 生产环境部署
```

## 文档类型
- **API文档**: Swagger自动生成的接口文档
- **开发文档**: 环境搭建、编码规范、调试指南
- **部署文档**: Docker部署、生产环境配置
- **架构文档**: 系统架构、数据库设计

## 维护说明
- API文档通过代码注释自动生成
- 开发文档需手动维护
- 部署文档与实际环境保持同步