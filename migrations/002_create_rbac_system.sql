-- =============================================================
-- 002_create_rbac_system.sql
-- 创建RBAC权限系统相关表
-- 包含：角色表、权限表、用户角色关联表、角色权限关联表
-- =============================================================

-- 角色表 (roles)
CREATE TABLE `roles` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色UUID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色名称',
  `slug` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色标识',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '角色描述',
  `is_system` tinyint(1) DEFAULT '0' COMMENT '是否系统角色',
  `is_default` tinyint(1) DEFAULT '0' COMMENT '是否默认角色',
  `priority` int DEFAULT '0' COMMENT '优先级',
  `permissions` json DEFAULT NULL COMMENT '权限配置',
  `settings` json DEFAULT NULL COMMENT '角色设置',
  `status` enum('active','inactive','deleted') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'active' COMMENT '状态',
  `created_by` int unsigned DEFAULT NULL COMMENT '创建者ID',
  `updated_by` int unsigned DEFAULT NULL COMMENT '更新者ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_roles_uuid` (`uuid`),
  UNIQUE KEY `uk_roles_name` (`name`),
  UNIQUE KEY `uk_roles_slug` (`slug`),
  KEY `idx_roles_is_system` (`is_system`),
  KEY `idx_roles_status` (`status`),
  KEY `idx_roles_created_by` (`created_by`),
  KEY `idx_roles_priority` (`priority`),
  KEY `idx_roles_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_roles_created_by` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_roles_updated_by` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 权限表 (permissions)
CREATE TABLE `permissions` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '权限ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '权限UUID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '权限名称',
  `slug` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '权限标识',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '权限描述',
  `resource_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源类型',
  `action` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '操作类型',
  `conditions` json DEFAULT NULL COMMENT '权限条件',
  `parent_id` int unsigned DEFAULT NULL COMMENT '父权限ID',
  `level` int DEFAULT '0' COMMENT '权限层级',
  `is_system` tinyint(1) DEFAULT '0' COMMENT '是否系统权限',
  `category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '权限分类',
  `sort_order` int DEFAULT '0' COMMENT '排序顺序',
  `status` enum('active','inactive','deleted') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'active' COMMENT '状态',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_permissions_uuid` (`uuid`),
  UNIQUE KEY `uk_permissions_name` (`name`),
  UNIQUE KEY `uk_permissions_slug` (`slug`),
  UNIQUE KEY `uk_permissions_resource_action` (`resource_type`,`action`),
  KEY `idx_permissions_resource_type` (`resource_type`),
  KEY `idx_permissions_action` (`action`),
  KEY `idx_permissions_parent_id` (`parent_id`),
  KEY `idx_permissions_category` (`category`),
  KEY `idx_permissions_is_system` (`is_system`),
  KEY `idx_permissions_status` (`status`),
  KEY `idx_permissions_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_permissions_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `permissions` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';

-- 用户角色关联表 (user_roles)
CREATE TABLE `user_roles` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '关联UUID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `role_id` int unsigned NOT NULL COMMENT '角色ID',
  `granted_by` int unsigned DEFAULT NULL COMMENT '授予者ID',
  `granted_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授予时间',
  `expires_at` timestamp NULL DEFAULT NULL COMMENT '过期时间',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否激活',
  `source` enum('direct','inherited','temporary') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'direct' COMMENT '角色来源',
  `metadata` json DEFAULT NULL COMMENT '元数据信息',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_roles_uuid` (`uuid`),
  UNIQUE KEY `uk_user_roles_user_role` (`user_id`,`role_id`),
  KEY `idx_user_roles_user_id` (`user_id`),
  KEY `idx_user_roles_role_id` (`role_id`),
  KEY `idx_user_roles_granted_by` (`granted_by`),
  KEY `idx_user_roles_expires_at` (`expires_at`),
  KEY `idx_user_roles_is_active` (`is_active`),
  CONSTRAINT `fk_user_roles_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_roles_role_id` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_roles_granted_by` FOREIGN KEY (`granted_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- 角色权限关联表 (role_permissions)
CREATE TABLE `role_permissions` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `uuid` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '关联UUID',
  `role_id` int unsigned NOT NULL COMMENT '角色ID',
  `permission_id` int unsigned NOT NULL COMMENT '权限ID',
  `granted_by` int unsigned DEFAULT NULL COMMENT '授予者ID',
  `granted_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授予时间',
  `conditions` json DEFAULT NULL COMMENT '权限条件',
  `is_denied` tinyint(1) DEFAULT '0' COMMENT '是否拒绝权限',
  `metadata` json DEFAULT NULL COMMENT '元数据信息',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` bigint DEFAULT '1' COMMENT '版本号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_permissions_uuid` (`uuid`),
  UNIQUE KEY `uk_role_permissions_role_permission` (`role_id`,`permission_id`),
  KEY `idx_role_permissions_role_id` (`role_id`),
  KEY `idx_role_permissions_permission_id` (`permission_id`),
  KEY `idx_role_permissions_granted_by` (`granted_by`),
  KEY `idx_role_permissions_is_denied` (`is_denied`),
  CONSTRAINT `fk_role_permissions_role_id` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_role_permissions_permission_id` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_role_permissions_granted_by` FOREIGN KEY (`granted_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';

-- 创建复合索引
CREATE INDEX `idx_roles_composite_status_system` ON `roles` (`status`, `is_system`);
CREATE INDEX `idx_permissions_composite_resource_action_status` ON `permissions` (`resource_type`, `action`, `status`);
CREATE INDEX `idx_user_roles_composite_user_active_expires` ON `user_roles` (`user_id`, `is_active`, `expires_at`);
CREATE INDEX `idx_role_permissions_composite_role_denied` ON `role_permissions` (`role_id`, `is_denied`);

-- 添加注释
ALTER TABLE `roles` COMMENT = '系统角色定义表，包含角色名称、权限配置和状态信息';
ALTER TABLE `permissions` COMMENT = '系统权限定义表，定义资源操作权限和层级关系';
ALTER TABLE `user_roles` COMMENT = '用户角色关联表，建立用户与角色的多对多关系';
ALTER TABLE `role_permissions` COMMENT = '角色权限关联表，建立角色与权限的多对多关系';