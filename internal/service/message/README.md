# message service 目录

## 目录说明
即时通讯相关业务逻辑处理模块。

## 功能描述
- WebSocket连接管理
- 消息发送和接收
- 会话管理
- 离线消息处理
- 媒体文件消息

## 主要文件
- **message_service.go** - 消息服务接口定义
- **message_service_impl.go** - 消息服务实现
- **websocket_service.go** - WebSocket连接服务
- **conversation_service.go** - 会话管理服务
- **media_service.go** - 媒体消息服务

## 核心功能
- WebSocket连接池管理
- 消息路由和分发
- 离线消息存储和推送
- Redis Stream消息队列
- 消息状态管理（已发送/已读）