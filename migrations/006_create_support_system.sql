-- =============================================================
-- 006_create_support_system.sql
-- 创建系统支撑功能相关表
-- 包含：回收站表、审计日志表、系统设置表、密码重置令牌表、通知表、验证码表等
-- =============================================================

-- 回收站表 (recycle_bin)
CREATE TABLE `recycle_bin` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '回收站ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '回收站UUID',
  `file_id` int unsigned NOT NULL COMMENT '文件ID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件名',
  `file_path` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '原路径',
  `file_size` bigint DEFAULT '0' COMMENT '文件大小',
  `mime_type` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'MIME类型',
  `is_folder` tinyint(1) DEFAULT '0' COMMENT '是否文件夹',
  `parent_path` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '父目录路径',
  `deletion_reason` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '删除原因',
  `metadata` json DEFAULT NULL COMMENT '文件元数据',
  `deleted_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '删除时间',
  `auto_delete_at` timestamp NOT NULL COMMENT '自动删除时间',
  `restored_at` timestamp NULL DEFAULT NULL COMMENT '恢复时间',
  `restored_by` int unsigned DEFAULT NULL COMMENT '恢复者ID',
  `restore_path` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '恢复路径',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_recycle_bin_uuid` (`uuid`),
  KEY `idx_recycle_bin_file_id` (`file_id`),
  KEY `idx_recycle_bin_user_id` (`user_id`),
  KEY `idx_recycle_bin_deleted_at` (`deleted_at`),
  KEY `idx_recycle_bin_auto_delete_at` (`auto_delete_at`),
  KEY `idx_recycle_bin_restored_by` (`restored_by`),
  CONSTRAINT `fk_recycle_bin_file_id` FOREIGN KEY (`file_id`) REFERENCES `files` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_recycle_bin_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_recycle_bin_restored_by` FOREIGN KEY (`restored_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='回收站表'
PARTITION BY RANGE (YEAR(deleted_at)) (
  PARTITION p2023 VALUES LESS THAN (2024),
  PARTITION p2024 VALUES LESS THAN (2025),
  PARTITION p2025 VALUES LESS THAN (2026),
  PARTITION p_future VALUES LESS THAN MAXVALUE
);

-- 审计日志表 (audit_logs)
CREATE TABLE `audit_logs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '日志UUID',
  `user_id` int unsigned DEFAULT NULL COMMENT '操作用户ID',
  `session_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '会话ID',
  `action` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '操作类型',
  `resource_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源类型',
  `resource_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '资源ID',
  `resource_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '资源名称',
  `details` json DEFAULT NULL COMMENT '操作详情',
  `changes` json DEFAULT NULL COMMENT '变更内容',
  `ip_address` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户代理',
  `location` json DEFAULT NULL COMMENT '地理位置',
  `request_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '请求ID',
  `request_method` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '请求方法',
  `request_uri` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '请求URI',
  `response_status` int DEFAULT NULL COMMENT '响应状态码',
  `result` enum('success','failure','error','warning') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'success' COMMENT '操作结果',
  `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '错误信息',
  `duration` int DEFAULT NULL COMMENT '执行耗时(毫秒)',
  `severity` enum('low','medium','high','critical') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'low' COMMENT '严重程度',
  `category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '日志分类',
  `tags` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '标签',
  `metadata` json DEFAULT NULL COMMENT '扩展元数据',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_audit_logs_uuid` (`uuid`),
  KEY `idx_audit_logs_user_id` (`user_id`),
  KEY `idx_audit_logs_action` (`action`),
  KEY `idx_audit_logs_resource_type` (`resource_type`),
  KEY `idx_audit_logs_resource_id` (`resource_id`),
  KEY `idx_audit_logs_result` (`result`),
  KEY `idx_audit_logs_severity` (`severity`),
  KEY `idx_audit_logs_category` (`category`),
  KEY `idx_audit_logs_ip_address` (`ip_address`),
  KEY `idx_audit_logs_created_at` (`created_at`),
  CONSTRAINT `fk_audit_logs_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志表'
PARTITION BY RANGE (YEAR(created_at)) (
  PARTITION p2023 VALUES LESS THAN (2024),
  PARTITION p2024 VALUES LESS THAN (2025),
  PARTITION p2025 VALUES LESS THAN (2026),
  PARTITION p_future VALUES LESS THAN MAXVALUE
);

-- 系统设置表 (system_settings)
CREATE TABLE `system_settings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '设置ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '设置UUID',
  `key_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '设置键名',
  `value` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '设置值',
  `default_value` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '默认值',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '设置描述',
  `type` enum('string','number','boolean','json','array','file','password') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'string' COMMENT '值类型',
  `category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'general' COMMENT '设置分类',
  `group_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分组名称',
  `is_public` tinyint(1) DEFAULT '0' COMMENT '是否公开',
  `is_readonly` tinyint(1) DEFAULT '0' COMMENT '是否只读',
  `is_encrypted` tinyint(1) DEFAULT '0' COMMENT '是否加密存储',
  `validation_rules` json DEFAULT NULL COMMENT '验证规则',
  `display_order` int DEFAULT '0' COMMENT '显示顺序',
  `updated_by` int unsigned DEFAULT NULL COMMENT '更新者ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_system_settings_uuid` (`uuid`),
  UNIQUE KEY `uk_system_settings_key_name` (`key_name`),
  KEY `idx_system_settings_category` (`category`),
  KEY `idx_system_settings_group_name` (`group_name`),
  KEY `idx_system_settings_is_public` (`is_public`),
  KEY `idx_system_settings_updated_by` (`updated_by`),
  CONSTRAINT `fk_system_settings_updated_by` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统设置表';

-- 密码重置令牌表 (password_reset_tokens)
CREATE TABLE `password_reset_tokens` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '令牌ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '令牌UUID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邮箱地址',
  `token` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '重置令牌',
  `token_hash` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '令牌哈希值',
  `salt` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '盐值',
  `is_used` tinyint(1) DEFAULT '0' COMMENT '是否已使用',
  `used_at` timestamp NULL DEFAULT NULL COMMENT '使用时间',
  `expires_at` timestamp NOT NULL COMMENT '过期时间',
  `attempt_count` int DEFAULT '0' COMMENT '尝试次数',
  `max_attempts` int DEFAULT '3' COMMENT '最大尝试次数',
  `ip_address` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '请求IP',
  `user_agent` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户代理',
  `metadata` json DEFAULT NULL COMMENT '令牌元数据',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_password_reset_tokens_uuid` (`uuid`),
  UNIQUE KEY `uk_password_reset_tokens_token` (`token`),
  KEY `idx_password_reset_tokens_user_id` (`user_id`),
  KEY `idx_password_reset_tokens_email` (`email`),
  KEY `idx_password_reset_tokens_token_hash` (`token_hash`),
  KEY `idx_password_reset_tokens_expires_at` (`expires_at`),
  KEY `idx_password_reset_tokens_is_used` (`is_used`),
  CONSTRAINT `fk_password_reset_tokens_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='密码重置令牌表';

-- 通知表 (notifications)
CREATE TABLE `notifications` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '通知ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '通知UUID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '通知类型',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '通知标题',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '通知内容',
  `summary` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '通知摘要',
  `data` json DEFAULT NULL COMMENT '通知数据',
  `action_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '操作链接',
  `action_text` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '操作按钮文本',
  `priority` enum('low','normal','high','urgent') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'normal' COMMENT '优先级',
  `category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '通知分类',
  `source_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '来源类型',
  `source_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '来源ID',
  `channels` json DEFAULT NULL COMMENT '发送渠道',
  `is_read` tinyint(1) DEFAULT '0' COMMENT '是否已读',
  `is_sent` tinyint(1) DEFAULT '0' COMMENT '是否已发送',
  `is_clicked` tinyint(1) DEFAULT '0' COMMENT '是否已点击',
  `read_at` timestamp NULL DEFAULT NULL COMMENT '阅读时间',
  `sent_at` timestamp NULL DEFAULT NULL COMMENT '发送时间',
  `clicked_at` timestamp NULL DEFAULT NULL COMMENT '点击时间',
  `expires_at` timestamp NULL DEFAULT NULL COMMENT '过期时间',
  `metadata` json DEFAULT NULL COMMENT '通知元数据',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_notifications_uuid` (`uuid`),
  KEY `idx_notifications_user_id` (`user_id`),
  KEY `idx_notifications_type` (`type`),
  KEY `idx_notifications_category` (`category`),
  KEY `idx_notifications_priority` (`priority`),
  KEY `idx_notifications_is_read` (`is_read`),
  KEY `idx_notifications_is_sent` (`is_sent`),
  KEY `idx_notifications_source` (`source_type`,`source_id`),
  KEY `idx_notifications_expires_at` (`expires_at`),
  KEY `idx_notifications_created_at` (`created_at`),
  CONSTRAINT `fk_notifications_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知表'
PARTITION BY RANGE (YEAR(created_at)) (
  PARTITION p2023 VALUES LESS THAN (2024),
  PARTITION p2024 VALUES LESS THAN (2025),
  PARTITION p2025 VALUES LESS THAN (2026),
  PARTITION p_future VALUES LESS THAN MAXVALUE
);

-- 验证码表 (verification_codes)
CREATE TABLE `verification_codes` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '验证码ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '验证码UUID',
  `target` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '目标(邮箱/手机)',
  `type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '验证码类型',
  `code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '验证码',
  `code_hash` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '验证码哈希值',
  `salt` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '盐值',
  `is_used` tinyint(1) DEFAULT '0' COMMENT '是否已使用',
  `used_at` timestamp NULL DEFAULT NULL COMMENT '使用时间',
  `expires_at` timestamp NOT NULL COMMENT '过期时间',
  `attempt_count` int DEFAULT '0' COMMENT '尝试次数',
  `max_attempts` int DEFAULT '5' COMMENT '最大尝试次数',
  `ip_address` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '请求IP',
  `user_agent` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户代理',
  `user_id` int unsigned DEFAULT NULL COMMENT '关联用户ID',
  `metadata` json DEFAULT NULL COMMENT '验证码元数据',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_verification_codes_uuid` (`uuid`),
  KEY `idx_verification_codes_target` (`target`),
  KEY `idx_verification_codes_type` (`type`),
  KEY `idx_verification_codes_code_hash` (`code_hash`),
  KEY `idx_verification_codes_expires_at` (`expires_at`),
  KEY `idx_verification_codes_is_used` (`is_used`),
  KEY `idx_verification_codes_user_id` (`user_id`),
  KEY `idx_verification_codes_ip_address` (`ip_address`),
  CONSTRAINT `fk_verification_codes_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='验证码表';

-- 邮件模板表 (email_templates)
CREATE TABLE `email_templates` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '模板ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模板UUID',
  `type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模板类型',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模板名称',
  `subject` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邮件主题',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邮件内容(HTML)',
  `text_content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '纯文本内容',
  `variables` json DEFAULT NULL COMMENT '模板变量定义',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否启用',
  `language` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'zh-CN' COMMENT '语言代码',
  `updated_by` int unsigned DEFAULT NULL COMMENT '更新者ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_email_templates_uuid` (`uuid`),
  UNIQUE KEY `uk_email_templates_type_language` (`type`,`language`),
  KEY `idx_email_templates_type` (`type`),
  KEY `idx_email_templates_is_active` (`is_active`),
  KEY `idx_email_templates_language` (`language`),
  KEY `idx_email_templates_updated_by` (`updated_by`),
  CONSTRAINT `fk_email_templates_updated_by` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邮件模板表';

-- 创建复合索引
CREATE INDEX `idx_recycle_bin_composite_user_deleted` ON `recycle_bin` (`user_id`, `deleted_at`);
CREATE INDEX `idx_audit_logs_composite_user_action_created` ON `audit_logs` (`user_id`, `action`, `created_at`);
CREATE INDEX `idx_audit_logs_composite_resource_result` ON `audit_logs` (`resource_type`, `resource_id`, `result`);
CREATE INDEX `idx_notifications_composite_user_read_created` ON `notifications` (`user_id`, `is_read`, `created_at`);
CREATE INDEX `idx_verification_codes_composite_target_type_expires` ON `verification_codes` (`target`, `type`, `expires_at`);

-- 添加注释
ALTER TABLE `recycle_bin` COMMENT = '回收站文件管理表，支持文件恢复和自动清理';
ALTER TABLE `audit_logs` COMMENT = '系统审计日志表，记录用户操作和系统事件';
ALTER TABLE `system_settings` COMMENT = '系统配置参数表，存储全局系统设置';
ALTER TABLE `password_reset_tokens` COMMENT = '密码重置令牌表，管理密码找回流程';
ALTER TABLE `notifications` COMMENT = '系统通知表，支持多渠道消息推送';
ALTER TABLE `verification_codes` COMMENT = '验证码管理表，支持邮箱和短信验证';
ALTER TABLE `email_templates` COMMENT = '邮件模板管理表，支持多语言邮件模板';