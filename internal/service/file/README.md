# file service 目录

## 目录说明
文件相关业务逻辑处理模块。

## 功能描述
- 文件上传、下载、删除
- 分片上传和断点续传
- 文件秒传和去重
- 文件预览和转换
- 文件版本管理
- 存储策略管理

## 主要文件
- **file_service.go** - 文件服务接口定义
- **file_service_impl.go** - 文件服务实现
- **upload_service.go** - 上传服务（分片、秒传）
- **storage_service.go** - 存储策略服务
- **preview_service.go** - 文件预览服务

## 核心功能
- 多种存储后端支持（本地、OSS）
- 大文件分片处理
- 文件去重和秒传
- 文件安全扫描（ClamAV）
- 存储配额控制