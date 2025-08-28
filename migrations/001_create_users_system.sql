-- =============================================================
-- 001_create_users_system.sql
-- 创建用户体系相关表
-- 包含：用户表、用户会话表、用户登录历史表、用户偏好表
-- =============================================================

-- 用户表 (users)
CREATE TABLE `users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户UUID',
  `email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邮箱地址',
  `username` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户名',
  `password_hash` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密码哈希',
  `password_salt` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密码盐值',
  `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '手机号',
  `avatar_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '头像URL',
  `status` enum('active','inactive','suspended','deleted') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'active' COMMENT '账号状态',
  `profile` json DEFAULT NULL COMMENT '用户配置信息',
  `storage_quota` bigint DEFAULT '10737418240' COMMENT '存储配额(字节)',
  `storage_used` bigint DEFAULT '0' COMMENT '已用存储(字节)',
  `email_verified` tinyint(1) DEFAULT '0' COMMENT '邮箱是否验证',
  `email_verified_at` timestamp NULL DEFAULT NULL COMMENT '邮箱验证时间',
  `phone_verified` tinyint(1) DEFAULT '0' COMMENT '手机是否验证',
  `phone_verified_at` timestamp NULL DEFAULT NULL COMMENT '手机验证时间',
  `mfa_enabled` tinyint(1) DEFAULT '0' COMMENT '是否启用MFA',
  `mfa_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'MFA类型',
  `mfa_secret` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'MFA密钥',
  `backup_codes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'MFA备用码',
  `last_login_at` timestamp NULL DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '最后登录IP',
  `password_changed_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '密码修改时间',
  `login_attempts` int DEFAULT '0' COMMENT '登录尝试次数',
  `locked_until` timestamp NULL DEFAULT NULL COMMENT '锁定到期时间',
  `is_admin` tinyint(1) DEFAULT '0' COMMENT '是否管理员',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_uuid` (`uuid`),
  UNIQUE KEY `uk_users_email` (`email`),
  UNIQUE KEY `uk_users_username` (`username`),
  KEY `idx_users_status` (`status`),
  KEY `idx_users_created_at` (`created_at`),
  KEY `idx_users_last_login_at` (`last_login_at`),
  KEY `idx_users_phone` (`phone`),
  KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 用户会话表 (user_sessions)
CREATE TABLE `user_sessions` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '会话ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '会话UUID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `session_token` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '会话令牌',
  `refresh_token` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '刷新令牌',
  `device_info` json DEFAULT NULL COMMENT '设备信息',
  `ip_address` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户代理',
  `location` json DEFAULT NULL COMMENT '登录位置信息',
  `expires_at` timestamp NOT NULL COMMENT '过期时间',
  `last_accessed_at` timestamp NULL DEFAULT NULL COMMENT '最后访问时间',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否活跃',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_sessions_uuid` (`uuid`),
  UNIQUE KEY `uk_user_sessions_session_token` (`session_token`),
  UNIQUE KEY `uk_user_sessions_refresh_token` (`refresh_token`),
  KEY `idx_user_sessions_user_id` (`user_id`),
  KEY `idx_user_sessions_expires_at` (`expires_at`),
  KEY `idx_user_sessions_ip_address` (`ip_address`),
  KEY `idx_user_sessions_is_active` (`is_active`),
  CONSTRAINT `fk_user_sessions_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表';

-- 用户登录历史表 (user_login_history)
CREATE TABLE `user_login_history` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '记录UUID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `ip_address` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户代理',
  `location` json DEFAULT NULL COMMENT '位置信息',
  `login_type` enum('password','mfa','social','remember') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'password' COMMENT '登录类型',
  `status` enum('success','failed','blocked','suspicious') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'success' COMMENT '登录状态',
  `failure_reason` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '失败原因',
  `device_fingerprint` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '设备指纹',
  `session_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '会话ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_login_history_uuid` (`uuid`),
  KEY `idx_user_login_history_user_id` (`user_id`),
  KEY `idx_user_login_history_ip_address` (`ip_address`),
  KEY `idx_user_login_history_status` (`status`),
  KEY `idx_user_login_history_created_at` (`created_at`),
  KEY `idx_user_login_history_login_type` (`login_type`),
  CONSTRAINT `fk_user_login_history_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录历史表'
PARTITION BY RANGE (YEAR(created_at)) (
  PARTITION p2023 VALUES LESS THAN (2024),
  PARTITION p2024 VALUES LESS THAN (2025),
  PARTITION p2025 VALUES LESS THAN (2026),
  PARTITION p_future VALUES LESS THAN MAXVALUE
);

-- 用户偏好表 (user_preferences)
CREATE TABLE `user_preferences` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '偏好ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '偏好UUID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '偏好分类',
  `key_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '偏好键名',
  `value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '偏好值',
  `value_type` enum('string','number','boolean','json','array') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'string' COMMENT '值类型',
  `is_encrypted` tinyint(1) DEFAULT '0' COMMENT '是否加密',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '偏好描述',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_preferences_uuid` (`uuid`),
  UNIQUE KEY `uk_user_preferences_user_key` (`user_id`,`category`,`key_name`),
  KEY `idx_user_preferences_user_id` (`user_id`),
  KEY `idx_user_preferences_category` (`category`),
  CONSTRAINT `fk_user_preferences_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户偏好表';

-- 创建索引
CREATE INDEX `idx_users_composite_status_created` ON `users` (`status`, `created_at`);
CREATE INDEX `idx_user_sessions_composite_user_expires` ON `user_sessions` (`user_id`, `expires_at`);
CREATE INDEX `idx_user_login_history_composite_user_time` ON `user_login_history` (`user_id`, `created_at`);

-- 添加注释
ALTER TABLE `users` COMMENT = '用户基本信息表，存储用户账号、认证、状态等核心信息';
ALTER TABLE `user_sessions` COMMENT = '用户会话管理表，记录用户登录会话和设备信息';  
ALTER TABLE `user_login_history` COMMENT = '用户登录历史表，记录登录行为和安全审计信息';
ALTER TABLE `user_preferences` COMMENT = '用户个性化偏好配置表';