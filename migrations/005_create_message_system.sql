-- =============================================================
-- 005_create_message_system.sql
-- 创建即时通讯系统相关表
-- 包含：会话表、消息表、会话成员表、消息已读状态表
-- =============================================================

-- 会话表 (conversations)
CREATE TABLE `conversations` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '会话ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '会话UUID',
  `type` enum('private','group','team','system') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'private' COMMENT '会话类型',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '会话名称',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '会话描述',
  `creator_id` int unsigned NOT NULL COMMENT '创建者ID',
  `avatar_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '会话头像URL',
  `settings` json DEFAULT NULL COMMENT '会话设置',
  `last_message_id` int unsigned DEFAULT NULL COMMENT '最后消息ID',
  `last_message_at` timestamp NULL DEFAULT NULL COMMENT '最后消息时间',
  `last_message_preview` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '最后消息预览',
  `member_count` int DEFAULT '0' COMMENT '成员数量',
  `message_count` int DEFAULT '0' COMMENT '消息总数',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否激活',
  `is_archived` tinyint(1) DEFAULT '0' COMMENT '是否归档',
  `is_muted` tinyint(1) DEFAULT '0' COMMENT '是否静音',
  `is_pinned` tinyint(1) DEFAULT '0' COMMENT '是否置顶',
  `pin_order` int DEFAULT '0' COMMENT '置顶排序',
  `auto_delete_after` int DEFAULT NULL COMMENT '自动删除时间(天)',
  `max_members` int DEFAULT '500' COMMENT '最大成员数',
  `require_approval` tinyint(1) DEFAULT '0' COMMENT '需要审核加入',
  `allow_invite` tinyint(1) DEFAULT '1' COMMENT '允许邀请成员',
  `visibility` enum('public','private','secret') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'private' COMMENT '可见性',
  `metadata` json DEFAULT NULL COMMENT '会话元数据',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_conversations_uuid` (`uuid`),
  KEY `idx_conversations_creator_id` (`creator_id`),
  KEY `idx_conversations_type` (`type`),
  KEY `idx_conversations_last_message_at` (`last_message_at`),
  KEY `idx_conversations_is_active` (`is_active`),
  KEY `idx_conversations_visibility` (`visibility`),
  KEY `idx_conversations_created_at` (`created_at`),
  KEY `idx_conversations_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_conversations_creator_id` FOREIGN KEY (`creator_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话表';

-- 消息表 (messages)
CREATE TABLE `messages` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '消息UUID',
  `conversation_id` int unsigned NOT NULL COMMENT '会话ID',
  `sender_id` int unsigned NOT NULL COMMENT '发送者ID',
  `reply_to_id` int unsigned DEFAULT NULL COMMENT '回复消息ID',
  `forward_from_id` int unsigned DEFAULT NULL COMMENT '转发来源消息ID',
  `message_type` enum('text','image','file','audio','video','system','location','contact','sticker','link') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'text' COMMENT '消息类型',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '消息内容',
  `content_html` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'HTML格式内容',
  `attachments` json DEFAULT NULL COMMENT '附件信息',
  `mentions` json DEFAULT NULL COMMENT '提及用户',
  `metadata` json DEFAULT NULL COMMENT '消息元数据',
  `file_id` int unsigned DEFAULT NULL COMMENT '关联文件ID',
  `file_size` bigint DEFAULT NULL COMMENT '文件大小',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件名称',
  `thumbnail_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '缩略图URL',
  `preview_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '预览URL',
  `duration` int DEFAULT NULL COMMENT '媒体时长(秒)',
  `is_edited` tinyint(1) DEFAULT '0' COMMENT '是否已编辑',
  `is_deleted` tinyint(1) DEFAULT '0' COMMENT '是否已删除',
  `is_pinned` tinyint(1) DEFAULT '0' COMMENT '是否置顶',
  `is_system` tinyint(1) DEFAULT '0' COMMENT '是否系统消息',
  `is_broadcast` tinyint(1) DEFAULT '0' COMMENT '是否广播消息',
  `delivery_status` enum('pending','sent','delivered','read','failed') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'pending' COMMENT '投递状态',
  `read_count` int DEFAULT '0' COMMENT '已读人数',
  `reaction_count` int DEFAULT '0' COMMENT '反应数量',
  `reply_count` int DEFAULT '0' COMMENT '回复数量',
  `edit_count` int DEFAULT '0' COMMENT '编辑次数',
  `last_edited_at` timestamp NULL DEFAULT NULL COMMENT '最后编辑时间',
  `expires_at` timestamp NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_messages_uuid` (`uuid`),
  KEY `idx_messages_conversation_id` (`conversation_id`),
  KEY `idx_messages_sender_id` (`sender_id`),
  KEY `idx_messages_reply_to_id` (`reply_to_id`),
  KEY `idx_messages_forward_from_id` (`forward_from_id`),
  KEY `idx_messages_message_type` (`message_type`),
  KEY `idx_messages_file_id` (`file_id`),
  KEY `idx_messages_delivery_status` (`delivery_status`),
  KEY `idx_messages_is_deleted` (`is_deleted`),
  KEY `idx_messages_is_pinned` (`is_pinned`),
  KEY `idx_messages_created_at` (`created_at`),
  KEY `idx_messages_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_messages_conversation_id` FOREIGN KEY (`conversation_id`) REFERENCES `conversations` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_messages_sender_id` FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_messages_reply_to_id` FOREIGN KEY (`reply_to_id`) REFERENCES `messages` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_messages_forward_from_id` FOREIGN KEY (`forward_from_id`) REFERENCES `messages` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_messages_file_id` FOREIGN KEY (`file_id`) REFERENCES `files` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表'
PARTITION BY RANGE (YEAR(created_at)) (
  PARTITION p2023 VALUES LESS THAN (2024),
  PARTITION p2024 VALUES LESS THAN (2025),
  PARTITION p2025 VALUES LESS THAN (2026),
  PARTITION p_future VALUES LESS THAN MAXVALUE
);

-- 会话成员表 (conversation_members)
CREATE TABLE `conversation_members` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '成员关系ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '关系UUID',
  `conversation_id` int unsigned NOT NULL COMMENT '会话ID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `role` enum('owner','admin','member','guest') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'member' COMMENT '成员角色',
  `permissions` json DEFAULT NULL COMMENT '成员权限',
  `status` enum('active','inactive','kicked','left','banned','muted') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'active' COMMENT '成员状态',
  `last_read_message_id` int unsigned DEFAULT NULL COMMENT '最后已读消息ID',
  `last_read_at` timestamp NULL DEFAULT NULL COMMENT '最后阅读时间',
  `last_active_at` timestamp NULL DEFAULT NULL COMMENT '最后活跃时间',
  `unread_count` int DEFAULT '0' COMMENT '未读消息数',
  `mention_count` int DEFAULT '0' COMMENT '提及消息数',
  `is_muted` tinyint(1) DEFAULT '0' COMMENT '是否静音',
  `is_pinned` tinyint(1) DEFAULT '0' COMMENT '是否置顶会话',
  `is_archived` tinyint(1) DEFAULT '0' COMMENT '是否归档会话',
  `is_favorite` tinyint(1) DEFAULT '0' COMMENT '是否收藏会话',
  `muted_until` timestamp NULL DEFAULT NULL COMMENT '静音到期时间',
  `joined_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
  `left_at` timestamp NULL DEFAULT NULL COMMENT '离开时间',
  `invited_by` int unsigned DEFAULT NULL COMMENT '邀请者ID',
  `message_count` int DEFAULT '0' COMMENT '发送消息数',
  `notification_settings` json DEFAULT NULL COMMENT '通知设置',
  `custom_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '自定义会话名称',
  `metadata` json DEFAULT NULL COMMENT '成员元数据',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_conversation_members_uuid` (`uuid`),
  UNIQUE KEY `uk_conversation_members_conv_user` (`conversation_id`,`user_id`),
  KEY `idx_conversation_members_conversation_id` (`conversation_id`),
  KEY `idx_conversation_members_user_id` (`user_id`),
  KEY `idx_conversation_members_role` (`role`),
  KEY `idx_conversation_members_status` (`status`),
  KEY `idx_conversation_members_last_read_at` (`last_read_at`),
  KEY `idx_conversation_members_invited_by` (`invited_by`),
  KEY `idx_conversation_members_joined_at` (`joined_at`),
  CONSTRAINT `fk_conversation_members_conversation_id` FOREIGN KEY (`conversation_id`) REFERENCES `conversations` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_conversation_members_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_conversation_members_invited_by` FOREIGN KEY (`invited_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话成员表';

-- 消息已读状态表 (message_read_status)
CREATE TABLE `message_read_status` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '已读状态ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态UUID',
  `message_id` int unsigned NOT NULL COMMENT '消息ID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `conversation_id` int unsigned NOT NULL COMMENT '会话ID',
  `read_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '阅读时间',
  `device_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '设备类型',
  `platform` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '平台',
  `ip_address` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'IP地址',
  `metadata` json DEFAULT NULL COMMENT '阅读元数据',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_message_read_status_uuid` (`uuid`),
  UNIQUE KEY `uk_message_read_status_message_user` (`message_id`,`user_id`),
  KEY `idx_message_read_status_message_id` (`message_id`),
  KEY `idx_message_read_status_user_id` (`user_id`),
  KEY `idx_message_read_status_conversation_id` (`conversation_id`),
  KEY `idx_message_read_status_read_at` (`read_at`),
  CONSTRAINT `fk_message_read_status_message_id` FOREIGN KEY (`message_id`) REFERENCES `messages` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_message_read_status_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_message_read_status_conversation_id` FOREIGN KEY (`conversation_id`) REFERENCES `conversations` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息已读状态表'
PARTITION BY RANGE (YEAR(read_at)) (
  PARTITION p2023 VALUES LESS THAN (2024),
  PARTITION p2024 VALUES LESS THAN (2025),
  PARTITION p2025 VALUES LESS THAN (2026),
  PARTITION p_future VALUES LESS THAN MAXVALUE
);

-- 创建复合索引
CREATE INDEX `idx_conversations_composite_type_active_updated` ON `conversations` (`type`, `is_active`, `last_message_at`);
CREATE INDEX `idx_messages_composite_conv_created_deleted` ON `messages` (`conversation_id`, `created_at`, `is_deleted`);
CREATE INDEX `idx_messages_composite_sender_type_created` ON `messages` (`sender_id`, `message_type`, `created_at`);
CREATE INDEX `idx_conversation_members_composite_user_status_active` ON `conversation_members` (`user_id`, `status`, `last_active_at`);
CREATE INDEX `idx_conversation_members_composite_conv_role_status` ON `conversation_members` (`conversation_id`, `role`, `status`);
CREATE INDEX `idx_message_read_status_composite_user_read` ON `message_read_status` (`user_id`, `read_at`);

-- 添加注释
ALTER TABLE `conversations` COMMENT = '会话基本信息表，管理私聊、群聊和团队会话';
ALTER TABLE `messages` COMMENT = '消息内容表，存储各类型消息和附件信息';
ALTER TABLE `conversation_members` COMMENT = '会话成员管理表，控制成员权限和状态';
ALTER TABLE `message_read_status` COMMENT = '消息阅读状态表，跟踪消息的已读状态';