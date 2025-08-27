# file repository 目录

## 目录说明
文件数据访问模块，处理文件相关的数据库操作。

## 功能描述
- 文件元数据管理
- 文件夹层级关系
- 文件版本管理
- 文件统计查询
- 分享链接管理

## 主要文件
- **file_repository.go** - 文件数据访问接口
- **file_repository_impl.go** - 文件数据访问实现
- **folder_repository.go** - 文件夹数据访问
- **file_version_repository.go** - 文件版本数据访问
- **file_share_repository.go** - 文件分享数据访问

## 核心功能
- 文件元数据存储和查询
- 文件夹树形结构管理
- 文件版本历史记录
- 文件搜索和过滤
- 存储使用量统计