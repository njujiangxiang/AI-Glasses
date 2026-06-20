-- 智镜巡检系统近期数据库结构更新脚本。
-- 适用场景：已有本地/测试/生产 MySQL 数据库拉取新代码后，补齐近期新增字段、索引和默认数据。
-- 安全性：脚本按“缺什么补什么”设计，可重复执行；不会删除业务数据。
-- 执行前建议：先备份数据库，再在目标数据库上执行。
-- 执行命令：
--   mysql -u <user> -p <database> < server/scripts/update_recent_schema.sql
-- 执行后建议：重启后端 API，让新字段和新接口一起生效。

DELIMITER //

-- 工具过程：仅当指定表缺少指定字段时，才执行 ALTER TABLE ADD COLUMN。
-- 这样可以让脚本重复执行，不会因为字段已存在而中断。
DROP PROCEDURE IF EXISTS add_column_if_missing//
CREATE PROCEDURE add_column_if_missing(
  IN p_table_name VARCHAR(64),
  IN p_column_name VARCHAR(64),
  IN p_column_definition TEXT
)
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = p_table_name
      AND COLUMN_NAME = p_column_name
  ) THEN
    SET @sql = CONCAT('ALTER TABLE `', p_table_name, '` ADD COLUMN `', p_column_name, '` ', p_column_definition);
    PREPARE stmt FROM @sql;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
  END IF;
END//

-- 工具过程：仅当指定表缺少指定索引时，才执行 ALTER TABLE ADD INDEX。
-- 注意：这里只补普通索引，避免旧数据存在重复值时添加唯一索引失败。
DROP PROCEDURE IF EXISTS add_index_if_missing//
CREATE PROCEDURE add_index_if_missing(
  IN p_table_name VARCHAR(64),
  IN p_index_name VARCHAR(64),
  IN p_index_definition TEXT
)
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = p_table_name
      AND INDEX_NAME = p_index_name
  ) THEN
    SET @sql = CONCAT('ALTER TABLE `', p_table_name, '` ADD ', p_index_definition);
    PREPARE stmt FROM @sql;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
  END IF;
END//

DELIMITER ;

START TRANSACTION;

-- users 表：补齐个人中心、头像、所属组织、角色关联需要的字段。
-- 背景：近期个人中心、用户管理、数据范围权限都依赖这些字段。
CALL add_column_if_missing('users', 'name', "VARCHAR(64) NOT NULL DEFAULT '' COMMENT '姓名' AFTER `display_name`");
CALL add_column_if_missing('users', 'gender', "VARCHAR(8) NOT NULL DEFAULT 'unknown' COMMENT '性别：male/female/unknown' AFTER `name`");
CALL add_column_if_missing('users', 'avatar_data', "LONGBLOB NULL COMMENT '头像二进制数据' AFTER `gender`");
CALL add_column_if_missing('users', 'avatar_content_type', "VARCHAR(64) NOT NULL DEFAULT '' COMMENT '头像MIME类型' AFTER `avatar_data`");
CALL add_column_if_missing('users', 'avatar_size', "BIGINT NOT NULL DEFAULT 0 COMMENT '头像字节数' AFTER `avatar_content_type`");
CALL add_column_if_missing('users', 'birth_year', "INT NOT NULL DEFAULT 0 COMMENT '出生年份' AFTER `avatar_size`");
CALL add_column_if_missing('users', 'birth_month', "INT NOT NULL DEFAULT 0 COMMENT '出生月份' AFTER `birth_year`");
CALL add_column_if_missing('users', 'id_card_no', "VARCHAR(32) NOT NULL DEFAULT '' COMMENT '身份证号' AFTER `birth_month`");
CALL add_column_if_missing('users', 'org_code', "VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属组织编码' AFTER `id_card_no`");
CALL add_column_if_missing('users', 'role_id', "BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '角色ID，关联roles.id' AFTER `org_code`");

CALL add_index_if_missing('users', 'idx_users_id_card_no', 'INDEX `idx_users_id_card_no` (`id_card_no`)');
CALL add_index_if_missing('users', 'idx_users_org_code', 'INDEX `idx_users_org_code` (`org_code`)');
CALL add_index_if_missing('users', 'idx_users_role_id', 'INDEX `idx_users_role_id` (`role_id`)');

UPDATE users
SET name = display_name
WHERE (name = '' OR name IS NULL)
  AND display_name <> '';

UPDATE users
SET gender = 'unknown'
WHERE gender = '' OR gender IS NULL;

-- organizations 表：补齐组织树父级编码，用于“本组织及下级”的数据范围过滤。
CALL add_column_if_missing('organizations', 'parent_code', "VARCHAR(64) NOT NULL DEFAULT '' COMMENT '父级组织编码' AFTER `name`");
CALL add_index_if_missing('organizations', 'idx_organizations_parent_code', 'INDEX `idx_organizations_parent_code` (`parent_code`)');

-- roles 表：补齐角色编码、描述、数据范围、排序和状态。
-- 背景：角色管理页面和数据范围中间件会读取这些字段。
CALL add_column_if_missing('roles', 'code', "VARCHAR(64) NOT NULL DEFAULT '' COMMENT '角色编码' AFTER `name`");
CALL add_column_if_missing('roles', 'description', "VARCHAR(255) NOT NULL DEFAULT '' COMMENT '角色说明' AFTER `code`");
CALL add_column_if_missing('roles', 'data_scope', "VARCHAR(32) NOT NULL DEFAULT 'org_only' COMMENT '数据范围：all/org_and_sub/org_only/self_only' AFTER `description`");
CALL add_column_if_missing('roles', 'sort', "INT NOT NULL DEFAULT 0 COMMENT '排序值' AFTER `data_scope`");
CALL add_column_if_missing('roles', 'status', "VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '角色状态：active/disabled' AFTER `sort`");

UPDATE roles SET code = CONCAT('role_', id) WHERE code = '' OR code IS NULL;
UPDATE roles SET data_scope = 'org_only' WHERE data_scope = '' OR data_scope IS NULL;
UPDATE roles SET status = 'active' WHERE status = '' OR status IS NULL;

UPDATE roles SET code = 'admin', description = '系统管理员', data_scope = 'all', sort = 1, status = 'active' WHERE name = '系统管理员';
UPDATE roles SET code = 'task_manager', description = '任务管理员', data_scope = 'org_and_sub', sort = 2, status = 'active' WHERE name = '任务管理员';
UPDATE roles SET code = 'team_leader', description = '班组长', data_scope = 'org_only', sort = 3, status = 'active' WHERE name = '班组长';
UPDATE roles SET code = 'inspector', description = '巡检员', data_scope = 'self_only', sort = 4, status = 'active' WHERE name = '巡检员';
UPDATE roles SET code = 'super_admin', description = '系统超级管理员，拥有所有权限', data_scope = 'all', sort = 1, status = 'active' WHERE name = '超级管理员';
UPDATE roles SET code = 'user', description = '普通业务用户', data_scope = 'org_only', sort = 2, status = 'active' WHERE name = '普通用户';
UPDATE roles SET code = 'inspector', description = '负责执行巡检任务', data_scope = 'self_only', sort = 3, status = 'active' WHERE name = '巡检人员';

CALL add_index_if_missing('roles', 'idx_roles_data_scope', 'INDEX `idx_roles_data_scope` (`data_scope`)');
CALL add_index_if_missing('roles', 'idx_roles_status', 'INDEX `idx_roles_status` (`status`)');

-- permissions 表：补齐菜单权限管理需要的菜单层级、图标、路由、组件、权限标识等字段。
-- 背景：菜单权限页面新增/编辑菜单，以及左侧动态菜单都依赖这些字段。
CALL add_column_if_missing('permissions', 'pid', "BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父级权限ID' AFTER `id`");
CALL add_column_if_missing('permissions', 'type', "VARCHAR(16) NOT NULL DEFAULT 'menu' COMMENT '权限类型：M目录/C菜单/A按钮' AFTER `pid`");
CALL add_column_if_missing('permissions', 'name', "VARCHAR(64) NOT NULL DEFAULT '' COMMENT '菜单或权限名称' AFTER `type`");
CALL add_column_if_missing('permissions', 'icon', "VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Element Plus 图标名称' AFTER `code`");
CALL add_column_if_missing('permissions', 'path', "VARCHAR(255) NOT NULL DEFAULT '' COMMENT '前端路由地址' AFTER `icon`");
CALL add_column_if_missing('permissions', 'component', "VARCHAR(255) NOT NULL DEFAULT '' COMMENT '前端组件路径' AFTER `path`");
CALL add_column_if_missing('permissions', 'sort', "INT NOT NULL DEFAULT 0 COMMENT '排序值' AFTER `component`");
CALL add_column_if_missing('permissions', 'perms', "VARCHAR(255) NOT NULL DEFAULT '' COMMENT '按钮权限标识' AFTER `sort`");
CALL add_column_if_missing('permissions', 'visible', "BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否显示' AFTER `perms`");
CALL add_column_if_missing('permissions', 'is_cache', "BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否缓存' AFTER `visible`");
CALL add_column_if_missing('permissions', 'status', "VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '权限状态：active/disabled' AFTER `is_cache`");

UPDATE permissions SET name = code WHERE (name = '' OR name IS NULL) AND code <> '';
UPDATE permissions SET type = 'menu' WHERE type = '' OR type IS NULL;
UPDATE permissions SET status = 'active' WHERE status = '' OR status IS NULL;

CALL add_index_if_missing('permissions', 'idx_permissions_status', 'INDEX `idx_permissions_status` (`status`)');

-- 默认用户 role_id 回填：只处理 role_id 仍为 0 的内置账号，不覆盖人工维护过的角色。
UPDATE users u
JOIN roles r ON r.name = '系统管理员'
SET u.role_id = r.id
WHERE u.username = 'admin' AND u.role_id = 0;

UPDATE users u
JOIN roles r ON r.name = '任务管理员'
SET u.role_id = r.id
WHERE u.username = 'manager' AND u.role_id = 0;

UPDATE users u
JOIN roles r ON r.name = '班组长'
SET u.role_id = r.id
WHERE u.username = 'leader' AND u.role_id = 0;

UPDATE users u
JOIN roles r ON r.name IN ('巡检员', '巡检人员')
SET u.role_id = r.id
WHERE u.username = 'inspector' AND u.role_id = 0;

COMMIT;

DROP PROCEDURE IF EXISTS add_index_if_missing;
DROP PROCEDURE IF EXISTS add_column_if_missing;
