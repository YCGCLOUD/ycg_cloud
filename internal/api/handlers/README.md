# handlers 目录

## 目录说明
HTTP请求处理器目录，包含各个业务模块的请求处理逻辑。

## 功能描述
- 解析HTTP请求参数
- 调用service层处理业务逻辑
- 格式化响应数据
- 处理HTTP状态码

## 文件组织
按业务模块组织处理器文件：
- user_handler.go - 用户相关请求处理
- file_handler.go - 文件相关请求处理
- team_handler.go - 团队相关请求处理
- message_handler.go - 消息相关请求处理

## 开发规范
- 每个handler只处理一个业务领域
- 不包含业务逻辑，只负责参数转换
- 统一的错误处理和响应格式
- 支持参数验证和数据绑定