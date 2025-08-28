-- =============================================================
-- 003_create_file_system.sql
-- 创建文件管理系统相关表
-- 包含：文件表、文件版本表、文件分享表、文件标签表、文件上传分片表
-- =============================================================

-- 文件表 (files)
CREATE TABLE `files` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '文件ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件UUID',
  `user_id` int unsigned NOT NULL COMMENT '所属用户ID',
  `parent_id` int unsigned DEFAULT NULL COMMENT '父文件夹ID',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件名',
  `path` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件路径',
  `full_path` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '完整路径',
  `mime_type` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'MIME类型',
  `extension` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件扩展名',
  `size` bigint DEFAULT '0' COMMENT '文件大小(字节)',
  `hash` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件哈希值(SHA256)',
  `md5_hash` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'MD5哈希值',
  `storage_type` enum('local','oss','both') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'local' COMMENT '存储类型',
  `storage_path` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储路径',
  `storage_provider_id` int unsigned DEFAULT NULL COMMENT '存储提供商ID',
  `metadata` json DEFAULT NULL COMMENT '元数据信息',
  `is_encrypted` tinyint(1) DEFAULT '0' COMMENT '是否加密',
  `encryption_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加密密钥',
  `is_folder` tinyint(1) DEFAULT '0' COMMENT '是否文件夹',
  `is_public` tinyint(1) DEFAULT '0' COMMENT '是否公开',
  `is_shared` tinyint(1) DEFAULT '0' COMMENT '是否已分享',
  `is_favorite` tinyint(1) DEFAULT '0' COMMENT '是否收藏',
  `download_count` int DEFAULT '0' COMMENT '下载次数',
  `view_count` int DEFAULT '0' COMMENT '查看次数',
  `share_count` int DEFAULT '0' COMMENT '分享次数',
  `version_count` int DEFAULT '1' COMMENT '版本数量',
  `current_version` int DEFAULT '1' COMMENT '当前版本号',
  `status` enum('active','processing','failed','deleted') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'active' COMMENT '文件状态',
  `last_accessed_at` timestamp NULL DEFAULT NULL COMMENT '最后访问时间',
  `last_modified_at` timestamp NULL DEFAULT NULL COMMENT '最后修改时间',
  `thumbnail_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '缩略图URL',
  `preview_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '预览URL',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_files_uuid` (`uuid`),
  UNIQUE KEY `uk_files_user_path` (`user_id`,`full_path`(500)),
  KEY `idx_files_user_id` (`user_id`),
  KEY `idx_files_parent_id` (`parent_id`),
  KEY `idx_files_path` (`path`(255)),
  KEY `idx_files_hash` (`hash`),
  KEY `idx_files_md5_hash` (`md5_hash`),
  KEY `idx_files_mime_type` (`mime_type`),
  KEY `idx_files_extension` (`extension`),
  KEY `idx_files_size` (`size`),
  KEY `idx_files_is_folder` (`is_folder`),
  KEY `idx_files_is_public` (`is_public`),
  KEY `idx_files_status` (`status`),
  KEY `idx_files_storage_type` (`storage_type`),
  KEY `idx_files_storage_provider_id` (`storage_provider_id`),
  KEY `idx_files_created_at` (`created_at`),
  KEY `idx_files_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_files_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_files_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `files` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件表';

-- 文件版本表 (file_versions)
CREATE TABLE `file_versions` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '版本ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '版本UUID',
  `file_id` int unsigned NOT NULL COMMENT '文件ID',
  `version_number` int NOT NULL COMMENT '版本号',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '版本文件名',
  `size` bigint NOT NULL COMMENT '文件大小',
  `hash` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件哈希值',
  `md5_hash` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'MD5哈希值',
  `storage_path` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '存储路径',
  `comment` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '版本说明',
  `change_type` enum('create','update','rename','move') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'update' COMMENT '变更类型',
  `metadata` json DEFAULT NULL COMMENT '版本元数据',
  `is_current` tinyint(1) DEFAULT '0' COMMENT '是否当前版本',
  `created_by` int unsigned NOT NULL COMMENT '创建者ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_file_versions_uuid` (`uuid`),
  UNIQUE KEY `uk_file_versions_file_version` (`file_id`,`version_number`),
  KEY `idx_file_versions_file_id` (`file_id`),
  KEY `idx_file_versions_version_number` (`version_number`),
  KEY `idx_file_versions_hash` (`hash`),
  KEY `idx_file_versions_created_by` (`created_by`),
  KEY `idx_file_versions_created_at` (`created_at`),
  KEY `idx_file_versions_is_current` (`is_current`),
  CONSTRAINT `fk_file_versions_file_id` FOREIGN KEY (`file_id`) REFERENCES `files` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_file_versions_created_by` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件版本表';

-- 文件分享表 (file_shares)
CREATE TABLE `file_shares` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '分享ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分享UUID',
  `file_id` int unsigned NOT NULL COMMENT '文件ID',
  `shared_by` int unsigned NOT NULL COMMENT '分享者ID',
  `shared_with` int unsigned DEFAULT NULL COMMENT '分享给用户ID',
  `share_token` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分享令牌',
  `share_type` enum('private','public','password','link') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'private' COMMENT '分享类型',
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '访问密码',
  `permissions` json DEFAULT NULL COMMENT '权限配置',
  `download_limit` int DEFAULT NULL COMMENT '下载次数限制',
  `download_count` int DEFAULT '0' COMMENT '已下载次数',
  `view_limit` int DEFAULT NULL COMMENT '查看次数限制',
  `view_count` int DEFAULT '0' COMMENT '已查看次数',
  `expires_at` timestamp NULL DEFAULT NULL COMMENT '过期时间',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否激活',
  `allow_download` tinyint(1) DEFAULT '1' COMMENT '允许下载',
  `allow_preview` tinyint(1) DEFAULT '1' COMMENT '允许预览',
  `allow_upload` tinyint(1) DEFAULT '0' COMMENT '允许上传',
  `require_login` tinyint(1) DEFAULT '0' COMMENT '需要登录',
  `metadata` json DEFAULT NULL COMMENT '分享元数据',
  `last_accessed_at` timestamp NULL DEFAULT NULL COMMENT '最后访问时间',
  `last_accessed_ip` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '最后访问IP',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_file_shares_uuid` (`uuid`),
  UNIQUE KEY `uk_file_shares_token` (`share_token`),
  KEY `idx_file_shares_file_id` (`file_id`),
  KEY `idx_file_shares_shared_by` (`shared_by`),
  KEY `idx_file_shares_shared_with` (`shared_with`),
  KEY `idx_file_shares_share_type` (`share_type`),
  KEY `idx_file_shares_expires_at` (`expires_at`),
  KEY `idx_file_shares_is_active` (`is_active`),
  KEY `idx_file_shares_created_at` (`created_at`),
  KEY `idx_file_shares_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_file_shares_file_id` FOREIGN KEY (`file_id`) REFERENCES `files` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_file_shares_shared_by` FOREIGN KEY (`shared_by`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_file_shares_shared_with` FOREIGN KEY (`shared_with`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件分享表';

-- 标签表 (tags) - 独立标签管理
CREATE TABLE `tags` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '标签ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '标签UUID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '标签名称',
  `color` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '#1890ff' COMMENT '标签颜色',
  `icon` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '标签图标',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '标签描述',
  `file_count` int DEFAULT '0' COMMENT '关联文件数量',
  `usage_count` int DEFAULT '0' COMMENT '使用次数',
  `is_system` tinyint(1) DEFAULT '0' COMMENT '是否系统标签',
  `category` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '标签分类',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tags_uuid` (`uuid`),
  UNIQUE KEY `uk_tags_user_name` (`user_id`,`name`),
  KEY `idx_tags_user_id` (`user_id`),
  KEY `idx_tags_name` (`name`),
  KEY `idx_tags_category` (`category`),
  KEY `idx_tags_is_system` (`is_system`),
  KEY `idx_tags_usage_count` (`usage_count`),
  KEY `idx_tags_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_tags_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签表';

-- 文件标签关联表 (file_tag_relations)
CREATE TABLE `file_tag_relations` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `file_id` int unsigned NOT NULL COMMENT '文件ID',
  `tag_id` int unsigned NOT NULL COMMENT '标签ID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_file_tag_relations_file_tag` (`file_id`,`tag_id`),
  KEY `idx_file_tag_relations_file_id` (`file_id`),
  KEY `idx_file_tag_relations_tag_id` (`tag_id`),
  KEY `idx_file_tag_relations_user_id` (`user_id`),
  CONSTRAINT `fk_file_tag_relations_file_id` FOREIGN KEY (`file_id`) REFERENCES `files` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_file_tag_relations_tag_id` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_file_tag_relations_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件标签关联表';

-- 文件上传分片表 (file_upload_chunks)
CREATE TABLE `file_upload_chunks` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '分片ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分片UUID',
  `upload_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '上传ID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `filename` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件名',
  `chunk_number` int NOT NULL COMMENT '分片序号',
  `chunk_size` bigint NOT NULL COMMENT '分片大小',
  `total_chunks` int NOT NULL COMMENT '总分片数',
  `total_size` bigint NOT NULL COMMENT '文件总大小',
  `hash` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分片哈希值',
  `file_hash` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件总哈希值',
  `storage_path` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分片存储路径',
  `status` enum('uploading','completed','failed','expired') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'uploading' COMMENT '上传状态',
  `retry_count` int DEFAULT '0' COMMENT '重试次数',
  `expires_at` timestamp NOT NULL COMMENT '过期时间',
  `metadata` json DEFAULT NULL COMMENT '分片元数据',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_file_upload_chunks_uuid` (`uuid`),
  UNIQUE KEY `uk_file_upload_chunks_upload_chunk` (`upload_id`,`chunk_number`),
  KEY `idx_file_upload_chunks_upload_id` (`upload_id`),
  KEY `idx_file_upload_chunks_user_id` (`user_id`),
  KEY `idx_file_upload_chunks_status` (`status`),
  KEY `idx_file_upload_chunks_expires_at` (`expires_at`),
  KEY `idx_file_upload_chunks_created_at` (`created_at`),
  CONSTRAINT `fk_file_upload_chunks_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件上传分片表';

-- 创建复合索引
CREATE INDEX `idx_files_composite_user_folder_name` ON `files` (`user_id`, `is_folder`, `name`);
CREATE INDEX `idx_files_composite_user_status_created` ON `files` (`user_id`, `status`, `created_at`);
CREATE INDEX `idx_files_composite_hash_size` ON `files` (`hash`, `size`);
CREATE INDEX `idx_file_versions_composite_file_current` ON `file_versions` (`file_id`, `is_current`);
CREATE INDEX `idx_file_shares_composite_active_expires` ON `file_shares` (`is_active`, `expires_at`);

-- 添加注释
ALTER TABLE `files` COMMENT = '文件基本信息表，存储文件元数据、路径、存储信息等';
ALTER TABLE `file_versions` COMMENT = '文件版本历史表，记录文件的版本变化';
ALTER TABLE `file_shares` COMMENT = '文件分享管理表，控制文件的分享权限和访问';
ALTER TABLE `tags` COMMENT = '标签定义表，支持用户自定义文件标签';
ALTER TABLE `file_tag_relations` COMMENT = '文件标签关联表，建立文件与标签的多对多关系';
ALTER TABLE `file_upload_chunks` COMMENT = '文件分片上传表，支持大文件断点续传';