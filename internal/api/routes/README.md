# routes 目录

## 目录说明
路由定义目录，负责API路由的注册和管理。

## 功能描述
- 定义RESTful API路由
- 配置路由中间件
- 实现API版本管理
- 组织路由分组

## 文件组织
- **router.go** - 主路由配置文件
- **api_v1.go** - API v1版本路由
- **api_v2.go** - API v2版本路由（预留）
- **auth_routes.go** - 认证相关路由
- **file_routes.go** - 文件相关路由
- **user_routes.go** - 用户相关路由
- **team_routes.go** - 团队相关路由

## 路由设计
```
/api/v1/
├── auth/          # 认证相关
├── users/         # 用户管理
├── files/         # 文件管理
├── teams/         # 团队协作
└── messages/      # 即时通讯
```

## 开发规范
- 遵循RESTful API设计原则
- 支持API版本管理
- 合理使用HTTP方法和状态码
- 统一的路由中间件配置