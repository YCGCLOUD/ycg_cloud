# user repository 目录

## 目录说明
用户数据访问模块，处理用户相关的数据库操作。

## 功能描述
- 用户CRUD操作
- 用户认证数据查询
- 用户权限数据管理
- 用户统计信息

## 主要文件
- **user_repository.go** - 用户数据访问接口
- **user_repository_impl.go** - 用户数据访问实现
- **role_repository.go** - 角色权限数据访问
- **user_cache.go** - 用户缓存管理

## 核心功能
- 用户基本信息管理
- 登录凭证验证
- 权限数据查询
- 用户活跃度统计
- Redis缓存集成