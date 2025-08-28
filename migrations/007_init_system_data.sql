-- =============================================================
-- 007_init_system_data.sql
-- 系统初始化数据脚本
-- 包含：默认角色、权限、系统设置、邮件模板等基础数据
-- =============================================================

-- 插入默认系统角色
INSERT INTO `roles` (`uuid`, `name`, `slug`, `description`, `is_system`, `is_default`, `priority`, `permissions`, `status`, `created_at`, `updated_at`) VALUES
(UUID(), '系统管理员', 'admin', '拥有系统所有权限的管理员角色', 1, 0, 100, '{"*": ["*"]}', 'active', NOW(), NOW()),
(UUID(), '普通用户', 'user', '普通用户默认角色', 1, 1, 10, '{"file": ["read", "write", "delete"], "team": ["create", "join"], "message": ["send", "read"]}', 'active', NOW(), NOW()),
(UUID(), '团队管理员', 'team_admin', '团队管理员角色', 1, 0, 50, '{"team": ["*"], "file": ["*"], "message": ["*"]}', 'active', NOW(), NOW()),
(UUID(), '团队编辑者', 'team_editor', '团队编辑者角色', 1, 0, 30, '{"team": ["read", "write"], "file": ["read", "write"], "message": ["send", "read"]}', 'active', NOW(), NOW()),
(UUID(), '团队查看者', 'team_viewer', '团队只读权限角色', 1, 0, 20, '{"team": ["read"], "file": ["read"], "message": ["read"]}', 'active', NOW(), NOW());

-- 插入系统权限
INSERT INTO `permissions` (`uuid`, `name`, `slug`, `description`, `resource_type`, `action`, `is_system`, `category`, `status`, `created_at`, `updated_at`) VALUES
-- 用户权限
(UUID(), '用户创建', 'user.create', '创建用户账号', 'user', 'create', 1, 'user', 'active', NOW(), NOW()),
(UUID(), '用户查看', 'user.read', '查看用户信息', 'user', 'read', 1, 'user', 'active', NOW(), NOW()),
(UUID(), '用户更新', 'user.update', '更新用户信息', 'user', 'update', 1, 'user', 'active', NOW(), NOW()),
(UUID(), '用户删除', 'user.delete', '删除用户账号', 'user', 'delete', 1, 'user', 'active', NOW(), NOW()),
(UUID(), '用户管理', 'user.manage', '管理用户权限', 'user', 'manage', 1, 'user', 'active', NOW(), NOW()),

-- 文件权限
(UUID(), '文件上传', 'file.upload', '上传文件', 'file', 'upload', 1, 'file', 'active', NOW(), NOW()),
(UUID(), '文件下载', 'file.download', '下载文件', 'file', 'download', 1, 'file', 'active', NOW(), NOW()),
(UUID(), '文件预览', 'file.preview', '预览文件内容', 'file', 'preview', 1, 'file', 'active', NOW(), NOW()),
(UUID(), '文件分享', 'file.share', '分享文件给他人', 'file', 'share', 1, 'file', 'active', NOW(), NOW()),
(UUID(), '文件删除', 'file.delete', '删除文件', 'file', 'delete', 1, 'file', 'active', NOW(), NOW()),
(UUID(), '文件管理', 'file.manage', '管理文件系统', 'file', 'manage', 1, 'file', 'active', NOW(), NOW()),

-- 团队权限
(UUID(), '团队创建', 'team.create', '创建团队', 'team', 'create', 1, 'team', 'active', NOW(), NOW()),
(UUID(), '团队加入', 'team.join', '加入团队', 'team', 'join', 1, 'team', 'active', NOW(), NOW()),
(UUID(), '团队管理', 'team.manage', '管理团队', 'team', 'manage', 1, 'team', 'active', NOW(), NOW()),
(UUID(), '团队邀请', 'team.invite', '邀请团队成员', 'team', 'invite', 1, 'team', 'active', NOW(), NOW()),

-- 消息权限
(UUID(), '消息发送', 'message.send', '发送消息', 'message', 'send', 1, 'message', 'active', NOW(), NOW()),
(UUID(), '消息读取', 'message.read', '读取消息', 'message', 'read', 1, 'message', 'active', NOW(), NOW()),
(UUID(), '消息删除', 'message.delete', '删除消息', 'message', 'delete', 1, 'message', 'active', NOW(), NOW()),
(UUID(), '消息管理', 'message.manage', '管理消息系统', 'message', 'manage', 1, 'message', 'active', NOW(), NOW()),

-- 系统权限
(UUID(), '系统配置', 'system.config', '系统配置管理', 'system', 'config', 1, 'system', 'active', NOW(), NOW()),
(UUID(), '系统监控', 'system.monitor', '系统监控查看', 'system', 'monitor', 1, 'system', 'active', NOW(), NOW()),
(UUID(), '系统审计', 'system.audit', '审计日志查看', 'system', 'audit', 1, 'system', 'active', NOW(), NOW()),
(UUID(), 'OSS配置', 'system.oss', 'OSS存储配置', 'system', 'oss', 1, 'system', 'active', NOW(), NOW());

-- 插入系统设置
INSERT INTO `system_settings` (`uuid`, `key_name`, `value`, `default_value`, `description`, `type`, `category`, `group_name`, `is_public`, `validation_rules`, `display_order`, `created_at`, `updated_at`) VALUES
-- 基础配置
(UUID(), 'system.name', '个人版网络云盘', '个人版网络云盘', '系统名称', 'string', 'general', 'basic', 1, '{"required": true, "max_length": 100}', 1, NOW(), NOW()),
(UUID(), 'system.version', '1.0.0', '1.0.0', '系统版本号', 'string', 'general', 'basic', 1, '{"required": true}', 2, NOW(), NOW()),
(UUID(), 'system.description', '安全高效的个人云盘系统', '安全高效的个人云盘系统', '系统描述', 'string', 'general', 'basic', 1, '{"max_length": 500}', 3, NOW(), NOW()),
(UUID(), 'system.timezone', 'Asia/Shanghai', 'Asia/Shanghai', '系统时区', 'string', 'general', 'basic', 1, '{"required": true}', 4, NOW(), NOW()),
(UUID(), 'system.language', 'zh-CN', 'zh-CN', '默认语言', 'string', 'general', 'basic', 1, '{"required": true}', 5, NOW(), NOW()),

-- 用户配置
(UUID(), 'user.default_quota', '10737418240', '10737418240', '用户默认存储配额(10GB)', 'number', 'user', 'quota', 0, '{"min": 1048576, "max": 1099511627776}', 10, NOW(), NOW()),
(UUID(), 'user.max_quota', '107374182400', '107374182400', '用户最大存储配额(100GB)', 'number', 'user', 'quota', 0, '{"min": 1048576}', 11, NOW(), NOW()),
(UUID(), 'user.registration_enabled', 'true', 'true', '是否允许用户注册', 'boolean', 'user', 'registration', 1, NULL, 12, NOW(), NOW()),
(UUID(), 'user.email_verification_required', 'true', 'true', '注册时是否需要邮箱验证', 'boolean', 'user', 'registration', 1, NULL, 13, NOW(), NOW()),
(UUID(), 'user.admin_approval_required', 'false', 'false', '注册时是否需要管理员审核', 'boolean', 'user', 'registration', 0, NULL, 14, NOW(), NOW()),

-- 文件配置
(UUID(), 'file.max_size', '536870912', '536870912', '单文件最大大小(512MB)', 'number', 'file', 'upload', 1, '{"min": 1024, "max": 10737418240}', 20, NOW(), NOW()),
(UUID(), 'file.chunk_size', '5242880', '5242880', '分片上传块大小(5MB)', 'number', 'file', 'upload', 0, '{"min": 1048576, "max": 104857600}', 21, NOW(), NOW()),
(UUID(), 'file.chunk_threshold', '52428800', '52428800', '分片上传阈值(50MB)', 'number', 'file', 'upload', 0, '{"min": 1048576}', 22, NOW(), NOW()),
(UUID(), 'file.allowed_extensions', 'jpg,jpeg,png,gif,bmp,pdf,doc,docx,xls,xlsx,ppt,pptx,txt,zip,rar', 'jpg,jpeg,png,gif,bmp,pdf,doc,docx,xls,xlsx,ppt,pptx,txt,zip,rar', '允许上传的文件扩展名', 'string', 'file', 'upload', 1, NULL, 23, NOW(), NOW()),
(UUID(), 'file.blocked_extensions', 'exe,bat,cmd,scr,vbs', 'exe,bat,cmd,scr,vbs', '禁止上传的文件扩展名', 'string', 'file', 'upload', 1, NULL, 24, NOW(), NOW()),
(UUID(), 'file.version_limit', '10', '10', '文件版本保留数量', 'number', 'file', 'version', 0, '{"min": 1, "max": 100}', 25, NOW(), NOW()),
(UUID(), 'file.enable_preview', 'true', 'true', '是否启用文件预览', 'boolean', 'file', 'preview', 1, NULL, 26, NOW(), NOW()),
(UUID(), 'file.enable_thumbnail', 'true', 'true', '是否启用缩略图', 'boolean', 'file', 'preview', 1, NULL, 27, NOW(), NOW()),

-- 安全配置
(UUID(), 'security.password_min_length', '8', '8', '密码最小长度', 'number', 'security', 'password', 1, '{"min": 6, "max": 128}', 30, NOW(), NOW()),
(UUID(), 'security.password_require_uppercase', 'true', 'true', '密码是否需要大写字母', 'boolean', 'security', 'password', 1, NULL, 31, NOW(), NOW()),
(UUID(), 'security.password_require_lowercase', 'true', 'true', '密码是否需要小写字母', 'boolean', 'security', 'password', 1, NULL, 32, NOW(), NOW()),
(UUID(), 'security.password_require_number', 'true', 'true', '密码是否需要数字', 'boolean', 'security', 'password', 1, NULL, 33, NOW(), NOW()),
(UUID(), 'security.password_require_symbol', 'false', 'false', '密码是否需要特殊字符', 'boolean', 'security', 'password', 1, NULL, 34, NOW(), NOW()),
(UUID(), 'security.session_timeout', '7200', '7200', '会话超时时间(秒)', 'number', 'security', 'session', 0, '{"min": 300, "max": 86400}', 35, NOW(), NOW()),
(UUID(), 'security.max_login_attempts', '5', '5', '最大登录尝试次数', 'number', 'security', 'login', 0, '{"min": 3, "max": 20}', 36, NOW(), NOW()),
(UUID(), 'security.lockout_duration', '1800', '1800', '账号锁定时长(秒)', 'number', 'security', 'login', 0, '{"min": 300, "max": 86400}', 37, NOW(), NOW()),
(UUID(), 'security.enable_mfa', 'true', 'true', '是否启用多因素认证', 'boolean', 'security', 'mfa', 1, NULL, 38, NOW(), NOW()),

-- 回收站配置
(UUID(), 'recycle.retention_days', '30', '30', '回收站文件保留天数', 'number', 'file', 'recycle', 1, '{"min": 1, "max": 365}', 40, NOW(), NOW()),
(UUID(), 'recycle.auto_clean_enabled', 'true', 'true', '是否自动清理过期文件', 'boolean', 'file', 'recycle', 1, NULL, 41, NOW(), NOW()),

-- 通知配置
(UUID(), 'notification.email_enabled', 'true', 'true', '是否启用邮件通知', 'boolean', 'notification', 'email', 1, NULL, 50, NOW(), NOW()),
(UUID(), 'notification.sms_enabled', 'false', 'false', '是否启用短信通知', 'boolean', 'notification', 'sms', 1, NULL, 51, NOW(), NOW()),
(UUID(), 'notification.push_enabled', 'true', 'true', '是否启用推送通知', 'boolean', 'notification', 'push', 1, NULL, 52, NOW(), NOW()),

-- 存储配置
(UUID(), 'storage.default_type', 'local', 'local', '默认存储类型', 'string', 'storage', 'basic', 0, '{"enum": ["local", "oss"]}', 60, NOW(), NOW()),
(UUID(), 'storage.oss_enabled', 'false', 'false', '是否启用OSS存储', 'boolean', 'storage', 'oss', 0, NULL, 61, NOW(), NOW()),
(UUID(), 'storage.oss_threshold', '524288000', '524288000', 'OSS存储阈值(500MB)', 'number', 'storage', 'oss', 0, '{"min": 1048576}', 62, NOW(), NOW());

-- 插入邮件模板
INSERT INTO `email_templates` (`uuid`, `type`, `name`, `subject`, `content`, `text_content`, `variables`, `is_active`, `language`, `created_at`, `updated_at`) VALUES
-- 注册验证邮件
(UUID(), 'register', '注册验证邮件', '欢迎注册{{system_name}} - 请验证您的邮箱', 
'<html><body><h2>欢迎注册{{system_name}}</h2><p>您好 {{username}}，</p><p>感谢您注册我们的服务！请点击下面的链接验证您的邮箱地址：</p><p><a href="{{verify_url}}" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">验证邮箱</a></p><p>验证码：<strong>{{verification_code}}</strong></p><p>此验证码将在{{expires_minutes}}分钟后过期。</p><p>如果您没有注册此账号，请忽略此邮件。</p><p>祝好，<br/>{{system_name}}团队</p></body></html>',
'欢迎注册{{system_name}}\n\n您好 {{username}}，\n\n感谢您注册我们的服务！请使用以下验证码验证您的邮箱地址：\n\n验证码：{{verification_code}}\n\n此验证码将在{{expires_minutes}}分钟后过期。\n\n验证链接：{{verify_url}}\n\n如果您没有注册此账号，请忽略此邮件。\n\n祝好，\n{{system_name}}团队',
'{"system_name": "系统名称", "username": "用户名", "verification_code": "验证码", "verify_url": "验证链接", "expires_minutes": "过期分钟数"}', 1, 'zh-CN', NOW(), NOW()),

-- 密码重置邮件
(UUID(), 'reset_password', '密码重置邮件', '{{system_name}} - 重置您的密码', 
'<html><body><h2>重置您的密码</h2><p>您好 {{username}}，</p><p>我们收到了重置您密码的请求。请点击下面的链接重置您的密码：</p><p><a href="{{reset_url}}" style="background-color: #dc3545; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">重置密码</a></p><p>重置码：<strong>{{reset_code}}</strong></p><p>此重置链接将在{{expires_minutes}}分钟后过期。</p><p>如果您没有请求重置密码，请忽略此邮件，您的密码不会被更改。</p><p>祝好，<br/>{{system_name}}团队</p></body></html>',
'重置您的密码\n\n您好 {{username}}，\n\n我们收到了重置您密码的请求。请使用以下重置码重置您的密码：\n\n重置码：{{reset_code}}\n\n重置链接：{{reset_url}}\n\n此重置链接将在{{expires_minutes}}分钟后过期。\n\n如果您没有请求重置密码，请忽略此邮件，您的密码不会被更改。\n\n祝好，\n{{system_name}}团队',
'{"system_name": "系统名称", "username": "用户名", "reset_code": "重置码", "reset_url": "重置链接", "expires_minutes": "过期分钟数"}', 1, 'zh-CN', NOW(), NOW()),

-- 团队邀请邮件
(UUID(), 'team_invite', '团队邀请邮件', '{{inviter_name}} 邀请您加入 {{team_name}} 团队', 
'<html><body><h2>团队邀请</h2><p>您好，</p><p>{{inviter_name}} 邀请您加入 <strong>{{team_name}}</strong> 团队。</p><p>团队描述：{{team_description}}</p><p>您的角色：{{role_name}}</p><p>请点击下面的链接接受邀请：</p><p><a href="{{invite_url}}" style="background-color: #28a745; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">接受邀请</a></p><p>邀请码：<strong>{{invite_code}}</strong></p><p>此邀请将在{{expires_date}}过期。</p><p>祝好，<br/>{{system_name}}团队</p></body></html>',
'团队邀请\n\n您好，\n\n{{inviter_name}} 邀请您加入 {{team_name}} 团队。\n\n团队描述：{{team_description}}\n您的角色：{{role_name}}\n\n邀请码：{{invite_code}}\n邀请链接：{{invite_url}}\n\n此邀请将在{{expires_date}}过期。\n\n祝好，\n{{system_name}}团队',
'{"system_name": "系统名称", "inviter_name": "邀请者姓名", "team_name": "团队名称", "team_description": "团队描述", "role_name": "角色名称", "invite_code": "邀请码", "invite_url": "邀请链接", "expires_date": "过期日期"}', 1, 'zh-CN', NOW(), NOW()),

-- 文件分享邮件
(UUID(), 'file_share', '文件分享邮件', '{{sharer_name}} 与您分享了文件', 
'<html><body><h2>文件分享</h2><p>您好，</p><p>{{sharer_name}} 与您分享了以下文件：</p><p><strong>{{file_name}}</strong></p><p>文件大小：{{file_size}}</p><p>分享说明：{{share_message}}</p><p>请点击下面的链接查看文件：</p><p><a href="{{share_url}}" style="background-color: #17a2b8; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">查看文件</a></p><p>访问密码：<strong>{{access_password}}</strong></p><p>此分享将在{{expires_date}}过期。</p><p>祝好，<br/>{{system_name}}团队</p></body></html>',
'文件分享\n\n您好，\n\n{{sharer_name}} 与您分享了以下文件：\n\n文件名：{{file_name}}\n文件大小：{{file_size}}\n分享说明：{{share_message}}\n\n访问密码：{{access_password}}\n分享链接：{{share_url}}\n\n此分享将在{{expires_date}}过期。\n\n祝好，\n{{system_name}}团队',
'{"system_name": "系统名称", "sharer_name": "分享者姓名", "file_name": "文件名", "file_size": "文件大小", "share_message": "分享说明", "access_password": "访问密码", "share_url": "分享链接", "expires_date": "过期日期"}', 1, 'zh-CN', NOW(), NOW()),

-- 系统通知邮件
(UUID(), 'system_notification', '系统通知邮件', '{{system_name}} 系统通知', 
'<html><body><h2>系统通知</h2><p>您好 {{username}}，</p><p>{{notification_title}}</p><p>{{notification_content}}</p><p>如有疑问，请联系系统管理员。</p><p>祝好，<br/>{{system_name}}团队</p></body></html>',
'系统通知\n\n您好 {{username}}，\n\n{{notification_title}}\n\n{{notification_content}}\n\n如有疑问，请联系系统管理员。\n\n祝好，\n{{system_name}}团队',
'{"system_name": "系统名称", "username": "用户名", "notification_title": "通知标题", "notification_content": "通知内容"}', 1, 'zh-CN', NOW(), NOW());

-- 插入系统标签
INSERT INTO `tags` (`uuid`, `user_id`, `name`, `color`, `description`, `is_system`, `category`, `created_at`, `updated_at`) VALUES
(UUID(), 1, '工作', '#1890ff', '工作相关文件', 1, 'work', NOW(), NOW()),
(UUID(), 1, '个人', '#52c41a', '个人相关文件', 1, 'personal', NOW(), NOW()),
(UUID(), 1, '重要', '#ff4d4f', '重要文件标记', 1, 'document', NOW(), NOW()),
(UUID(), 1, '收藏', '#faad14', '收藏的文件', 1, 'other', NOW(), NOW()),
(UUID(), 1, '临时', '#8c8c8c', '临时文件', 1, 'other', NOW(), NOW()),
(UUID(), 1, '项目', '#722ed1', '项目相关文件', 1, 'project', NOW(), NOW()),
(UUID(), 1, '归档', '#13c2c2', '归档文件', 1, 'document', NOW(), NOW()),
(UUID(), 1, '公开', '#fa8c16', '公开共享文件', 1, 'other', NOW(), NOW());

-- 创建触发器：自动清理过期的验证码
DELIMITER $$
CREATE TRIGGER cleanup_expired_verification_codes
    BEFORE INSERT ON verification_codes
    FOR EACH ROW
BEGIN
    DELETE FROM verification_codes 
    WHERE expires_at < NOW() 
    AND created_at < DATE_SUB(NOW(), INTERVAL 1 DAY);
END$$
DELIMITER ;

-- 创建触发器：自动清理过期的密码重置令牌
DELIMITER $$
CREATE TRIGGER cleanup_expired_password_tokens
    BEFORE INSERT ON password_reset_tokens
    FOR EACH ROW
BEGIN
    DELETE FROM password_reset_tokens 
    WHERE expires_at < NOW() 
    AND created_at < DATE_SUB(NOW(), INTERVAL 1 DAY);
END$$
DELIMITER ;

-- 创建触发器：自动清理过期的用户会话
DELIMITER $$
CREATE TRIGGER cleanup_expired_sessions
    BEFORE INSERT ON user_sessions
    FOR EACH ROW
BEGIN
    DELETE FROM user_sessions 
    WHERE expires_at < NOW() 
    AND created_at < DATE_SUB(NOW(), INTERVAL 1 HOUR);
END$$
DELIMITER ;

-- 创建事件：定期清理任务
SET GLOBAL event_scheduler = ON;

-- 每日凌晨2点清理过期数据
CREATE EVENT IF NOT EXISTS daily_cleanup
ON SCHEDULE EVERY 1 DAY
STARTS (TIMESTAMP(CURRENT_DATE) + INTERVAL 2 HOUR)
DO
BEGIN
    -- 清理过期验证码
    DELETE FROM verification_codes WHERE expires_at < NOW();
    
    -- 清理过期密码重置令牌
    DELETE FROM password_reset_tokens WHERE expires_at < NOW();
    
    -- 清理过期会话
    DELETE FROM user_sessions WHERE expires_at < NOW();
    
    -- 清理过期通知(保留30天)
    DELETE FROM notifications WHERE created_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
    
    -- 清理过期文件分享
    UPDATE file_shares SET is_active = 0 WHERE expires_at < NOW() AND is_active = 1;
    
    -- 清理回收站过期文件
    DELETE rb FROM recycle_bin rb
    LEFT JOIN files f ON rb.file_id = f.id
    WHERE rb.auto_delete_at < NOW();
END;

-- 每周日凌晨3点清理审计日志(保留180天)
CREATE EVENT IF NOT EXISTS weekly_audit_cleanup
ON SCHEDULE EVERY 1 WEEK
STARTS (TIMESTAMP(CURRENT_DATE - INTERVAL WEEKDAY(CURRENT_DATE) DAY) + INTERVAL 3 HOUR)
DO
BEGIN
    DELETE FROM audit_logs WHERE created_at < DATE_SUB(NOW(), INTERVAL 180 DAY);
END;