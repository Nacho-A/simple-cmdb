-- Cursor CMDB 数据库初始化
-- MySQL 8.0, charset utf8mb4

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 用户表
-- ----------------------------
DROP TABLE IF EXISTS `user_roles`;
DROP TABLE IF EXISTS `role_menus`;
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL,
  `password` varchar(255) NOT NULL,
  `nickname` varchar(64) DEFAULT '',
  `email` varchar(128) DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '0禁用 1启用',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------
-- 角色表
-- ----------------------------
DROP TABLE IF EXISTS `roles`;
CREATE TABLE `roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `description` varchar(255) DEFAULT '',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_roles_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------
-- 用户-角色关联
-- ----------------------------
CREATE TABLE `user_roles` (
  `user_id` bigint unsigned NOT NULL,
  `role_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`user_id`,`role_id`),
  KEY `idx_user_roles_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------
-- 菜单表
-- ----------------------------
DROP TABLE IF EXISTS `menus`;
CREATE TABLE `menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `path` varchar(255) NOT NULL,
  `component` varchar(255) DEFAULT '',
  `icon` varchar(64) DEFAULT '',
  `parent_id` bigint unsigned DEFAULT NULL,
  `order` int NOT NULL DEFAULT 0,
  `hidden` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_menus_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------
-- 角色-菜单关联
-- ----------------------------
CREATE TABLE `role_menus` (
  `role_id` bigint unsigned NOT NULL,
  `menu_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`role_id`,`menu_id`),
  KEY `idx_role_menus_menu_id` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------
-- CMDB 资产表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb_assets`;
CREATE TABLE `cmdb_assets` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `service_name` varchar(100) NOT NULL,
  `private_ip` varchar(50) DEFAULT '',
  `public_ip` varchar(50) DEFAULT '',
  `labels` json DEFAULT NULL,
  `tags` varchar(200) DEFAULT '',
  `owner` varchar(50) DEFAULT '',
  `cloud_provider` varchar(50) DEFAULT '',
  `region` varchar(50) DEFAULT '',
  `instance_type` varchar(50) DEFAULT '',
  `status` varchar(20) DEFAULT '',
  `remark` text,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_cmdb_assets_service_name` (`service_name`),
  KEY `idx_cmdb_assets_owner` (`owner`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET FOREIGN_KEY_CHECKS = 1;

-- ----------------------------
-- 初始角色（id 1=admin, 2=operator, 3=viewer）
-- ----------------------------
INSERT INTO `roles` (`id`, `name`, `description`, `created_at`, `updated_at`) VALUES
(1, 'admin', '系统管理员', NOW(3), NOW(3)),
(2, 'operator', '运维/操作员', NOW(3), NOW(3)),
(3, 'viewer', '只读访客', NOW(3), NOW(3));

-- ----------------------------
-- 默认菜单（一级 + 二级）
-- 仪表盘=1, 资产管理(目录)=2, CMDB资产=3, 系统管理=4, 用户管理=5, 角色管理=6, 菜单管理=7
-- ----------------------------
INSERT INTO `menus` (`id`, `name`, `path`, `component`, `icon`, `parent_id`, `order`, `hidden`, `created_at`, `updated_at`) VALUES
(1, '仪表盘', '/dashboard', 'views/dashboard/index.vue', 'Odometer', NULL, 1, 0, NOW(3), NOW(3)),
(2, '资产管理', '/cmdb', '', 'Box', NULL, 2, 0, NOW(3), NOW(3)),
(3, 'CMDB资产', '/cmdb/assets', 'views/cmdb/assets/index.vue', 'Monitor', 2, 1, 0, NOW(3), NOW(3)),
(4, '系统管理', '/system', '', 'Setting', NULL, 9, 0, NOW(3), NOW(3)),
(5, '用户管理', '/system/user', 'views/system/user/index.vue', 'User', 4, 1, 0, NOW(3), NOW(3)),
(6, '角色管理', '/system/role', 'views/system/role/index.vue', 'Avatar', 4, 2, 0, NOW(3), NOW(3)),
(7, '菜单管理', '/system/menu', 'views/system/menu/index.vue', 'Menu', 4, 3, 0, NOW(3), NOW(3));

-- ----------------------------
-- 角色-菜单绑定
-- admin: 全部 1,2,3,4,5,6,7
-- operator / viewer: 仪表盘+资产管理+CMDB资产 1,2,3
-- ----------------------------
INSERT INTO `role_menus` (`role_id`, `menu_id`) VALUES
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7),
(2, 1), (2, 2), (2, 3),
(3, 1), (3, 2), (3, 3);

-- ----------------------------
-- 默认管理员：不在此插入，由后端首次启动时自动创建（app.bootstrap=true）
-- 用户名 admin，密码 admin123
-- ----------------------------

-- Casbin 策略表由 GORM Adapter 自动创建，策略数据由后端 bootstrap 时写入（p, admin, /api/v1/*, * 等）。
