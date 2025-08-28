-- =============================================================
-- 004_create_team_system.sql
-- 创建团队协作系统相关表
-- 包含：团队表、团队成员表、团队文件表、团队邀请表
-- =============================================================

-- 团队表 (teams)
CREATE TABLE `teams` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '团队ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '团队UUID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '团队名称',
  `slug` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '团队标识',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '团队描述',
  `owner_id` int unsigned NOT NULL COMMENT '团队所有者ID',
  `avatar_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '团队头像URL',
  `cover_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '团队封面URL',
  `settings` json DEFAULT NULL COMMENT '团队设置',
  `member_limit` int DEFAULT '50' COMMENT '成员数量限制',
  `member_count` int DEFAULT '1' COMMENT '当前成员数量',
  `storage_quota` bigint DEFAULT '107374182400' COMMENT '存储配额(100GB)',
  `storage_used` bigint DEFAULT '0' COMMENT '已用存储',
  `file_count` int DEFAULT '0' COMMENT '文件总数',
  `is_public` tinyint(1) DEFAULT '0' COMMENT '是否公开团队',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否激活',
  `allow_join_request` tinyint(1) DEFAULT '1' COMMENT '允许申请加入',
  `require_approval` tinyint(1) DEFAULT '1' COMMENT '需要审核加入',
  `visibility` enum('public','private','internal') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'private' COMMENT '可见性',
  `status` enum('active','suspended','deleted') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'active' COMMENT '团队状态',
  `category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '团队分类',
  `tags` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '团队标签',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_teams_uuid` (`uuid`),
  UNIQUE KEY `uk_teams_slug` (`slug`),
  KEY `idx_teams_name` (`name`),
  KEY `idx_teams_owner_id` (`owner_id`),
  KEY `idx_teams_is_public` (`is_public`),
  KEY `idx_teams_is_active` (`is_active`),
  KEY `idx_teams_status` (`status`),
  KEY `idx_teams_visibility` (`visibility`),
  KEY `idx_teams_category` (`category`),
  KEY `idx_teams_created_at` (`created_at`),
  KEY `idx_teams_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_teams_owner_id` FOREIGN KEY (`owner_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队表';

-- 团队成员表 (team_members)
CREATE TABLE `team_members` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '成员关系ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '关系UUID',
  `team_id` int unsigned NOT NULL COMMENT '团队ID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `role` enum('owner','admin','editor','viewer','guest') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'viewer' COMMENT '成员角色',
  `permissions` json DEFAULT NULL COMMENT '自定义权限',
  `status` enum('active','inactive','pending','rejected','left') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'active' COMMENT '成员状态',
  `invited_by` int unsigned DEFAULT NULL COMMENT '邀请者ID',
  `invited_at` timestamp NULL DEFAULT NULL COMMENT '邀请时间',
  `joined_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
  `last_active_at` timestamp NULL DEFAULT NULL COMMENT '最后活跃时间',
  `contribution_score` int DEFAULT '0' COMMENT '贡献分数',
  `file_upload_count` int DEFAULT '0' COMMENT '上传文件数',
  `file_download_count` int DEFAULT '0' COMMENT '下载文件数',
  `is_favorite` tinyint(1) DEFAULT '0' COMMENT '是否收藏团队',
  `notification_settings` json DEFAULT NULL COMMENT '通知设置',
  `metadata` json DEFAULT NULL COMMENT '成员元数据',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_team_members_uuid` (`uuid`),
  UNIQUE KEY `uk_team_members_team_user` (`team_id`,`user_id`),
  KEY `idx_team_members_team_id` (`team_id`),
  KEY `idx_team_members_user_id` (`user_id`),
  KEY `idx_team_members_role` (`role`),
  KEY `idx_team_members_status` (`status`),
  KEY `idx_team_members_invited_by` (`invited_by`),
  KEY `idx_team_members_joined_at` (`joined_at`),
  KEY `idx_team_members_last_active_at` (`last_active_at`),
  CONSTRAINT `fk_team_members_team_id` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_team_members_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_team_members_invited_by` FOREIGN KEY (`invited_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队成员表';

-- 团队文件表 (team_files)
CREATE TABLE `team_files` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '团队文件ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '关系UUID',
  `team_id` int unsigned NOT NULL COMMENT '团队ID',
  `file_id` int unsigned NOT NULL COMMENT '文件ID',
  `shared_by` int unsigned NOT NULL COMMENT '分享者ID',
  `permissions` json DEFAULT NULL COMMENT '文件权限',
  `access_level` enum('read','write','admin','owner') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'read' COMMENT '访问级别',
  `is_pinned` tinyint(1) DEFAULT '0' COMMENT '是否置顶',
  `pin_order` int DEFAULT '0' COMMENT '置顶排序',
  `category` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件分类',
  `tags` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件标签',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件描述',
  `view_count` int DEFAULT '0' COMMENT '查看次数',
  `download_count` int DEFAULT '0' COMMENT '下载次数',
  `comment_count` int DEFAULT '0' COMMENT '评论数量',
  `last_viewed_at` timestamp NULL DEFAULT NULL COMMENT '最后查看时间',
  `last_downloaded_at` timestamp NULL DEFAULT NULL COMMENT '最后下载时间',
  `shared_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '分享时间',
  `expires_at` timestamp NULL DEFAULT NULL COMMENT '过期时间',
  `metadata` json DEFAULT NULL COMMENT '扩展元数据',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_team_files_uuid` (`uuid`),
  UNIQUE KEY `uk_team_files_team_file` (`team_id`,`file_id`),
  KEY `idx_team_files_team_id` (`team_id`),
  KEY `idx_team_files_file_id` (`file_id`),
  KEY `idx_team_files_shared_by` (`shared_by`),
  KEY `idx_team_files_access_level` (`access_level`),
  KEY `idx_team_files_is_pinned` (`is_pinned`),
  KEY `idx_team_files_category` (`category`),
  KEY `idx_team_files_shared_at` (`shared_at`),
  KEY `idx_team_files_expires_at` (`expires_at`),
  CONSTRAINT `fk_team_files_team_id` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_team_files_file_id` FOREIGN KEY (`file_id`) REFERENCES `files` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_team_files_shared_by` FOREIGN KEY (`shared_by`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队文件表';

-- 团队邀请表 (team_invitations)
CREATE TABLE `team_invitations` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '邀请ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邀请UUID',
  `team_id` int unsigned NOT NULL COMMENT '团队ID',
  `invited_by` int unsigned NOT NULL COMMENT '邀请者ID',
  `invited_user_id` int unsigned DEFAULT NULL COMMENT '被邀请用户ID',
  `email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '邀请邮箱',
  `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '邀请手机号',
  `invitation_token` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邀请令牌',
  `invitation_code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '邀请码',
  `role` enum('admin','editor','viewer','guest') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'viewer' COMMENT '邀请角色',
  `permissions` json DEFAULT NULL COMMENT '预设权限',
  `message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '邀请消息',
  `status` enum('pending','accepted','rejected','expired','cancelled') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'pending' COMMENT '邀请状态',
  `invitation_type` enum('direct','email','phone','link','batch') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'direct' COMMENT '邀请类型',
  `expires_at` timestamp NOT NULL COMMENT '过期时间',
  `accepted_at` timestamp NULL DEFAULT NULL COMMENT '接受时间',
  `rejected_at` timestamp NULL DEFAULT NULL COMMENT '拒绝时间',
  `rejected_reason` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '拒绝原因',
  `attempts_count` int DEFAULT '0' COMMENT '尝试次数',
  `last_attempt_at` timestamp NULL DEFAULT NULL COMMENT '最后尝试时间',
  `metadata` json DEFAULT NULL COMMENT '邀请元数据',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_team_invitations_uuid` (`uuid`),
  UNIQUE KEY `uk_team_invitations_token` (`invitation_token`),
  KEY `idx_team_invitations_team_id` (`team_id`),
  KEY `idx_team_invitations_invited_by` (`invited_by`),
  KEY `idx_team_invitations_invited_user_id` (`invited_user_id`),
  KEY `idx_team_invitations_email` (`email`),
  KEY `idx_team_invitations_phone` (`phone`),
  KEY `idx_team_invitations_invitation_code` (`invitation_code`),
  KEY `idx_team_invitations_status` (`status`),
  KEY `idx_team_invitations_expires_at` (`expires_at`),
  KEY `idx_team_invitations_created_at` (`created_at`),
  CONSTRAINT `fk_team_invitations_team_id` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_team_invitations_invited_by` FOREIGN KEY (`invited_by`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_team_invitations_invited_user_id` FOREIGN KEY (`invited_user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队邀请表';

-- 创建复合索引
CREATE INDEX `idx_teams_composite_status_public` ON `teams` (`status`, `is_public`);
CREATE INDEX `idx_teams_composite_owner_status` ON `teams` (`owner_id`, `status`);
CREATE INDEX `idx_team_members_composite_team_role_status` ON `team_members` (`team_id`, `role`, `status`);
CREATE INDEX `idx_team_members_composite_user_status_joined` ON `team_members` (`user_id`, `status`, `joined_at`);
CREATE INDEX `idx_team_files_composite_team_pinned_shared` ON `team_files` (`team_id`, `is_pinned`, `shared_at`);
CREATE INDEX `idx_team_invitations_composite_status_expires` ON `team_invitations` (`status`, `expires_at`);

-- 添加注释
ALTER TABLE `teams` COMMENT = '团队基本信息表，存储团队配置、成员限制和存储配额';
ALTER TABLE `team_members` COMMENT = '团队成员关系表，管理成员角色、权限和活跃状态';
ALTER TABLE `team_files` COMMENT = '团队文件共享表，控制文件在团队中的访问权限';
ALTER TABLE `team_invitations` COMMENT = '团队邀请管理表，处理团队成员邀请流程';