-- RBAC 权限系统初始化数据
-- 执行此脚本以初始化菜单和角色数据
-- 注意: 确保数据库表使用 GORM 默认的复数蛇形命名

-- ================================
-- 1. 清理现有权限数据
-- ================================

-- 先删除角色权限关联（避免外键约束）
DELETE FROM role_permissions;

-- 删除菜单/权限表（保留系统数据）
-- 如果是新系统，可以清空后重建；如果已有数据，建议跳过删除直接添加
-- DELETE FROM permissions;
-- DELETE FROM roles;

-- ================================
-- 2. 初始化菜单/权限数据（如果不存在）
-- ================================

-- 使用 INSERT IGNORE 避免重复插入（MySQL 语法）
-- 如果使用其他数据库，请调整语法

-- 一级菜单
INSERT IGNORE INTO permissions (id, pid, type, name, code, icon, path, component, sort, perms, visible, status, is_cache, created_at, updated_at) VALUES
(1, 0, 'M', '工作台', 'workbench', 'DataAnalysis', '/workbench', '', 1, '', 1, 'active', 0, NOW(), NOW()),
(2, 0, 'M', '巡检模板', 'templates', 'Document', '/templates', '', 2, '', 1, 'active', 0, NOW(), NOW()),
(3, 0, 'M', '工作流管理', 'workflows', 'Operation', '/workflows', '', 3, '', 1, 'active', 0, NOW(), NOW()),
(4, 0, 'M', '任务计划', 'plans', 'Calendar', '/plans', '', 4, '', 1, 'active', 0, NOW(), NOW()),
(5, 0, 'M', '任务管理', 'tasks', 'Tickets', '/tasks', '', 5, '', 1, 'active', 0, NOW(), NOW()),
(6, 0, 'M', '作业任务单', 'tasksheets', 'Document', '/tasksheets', '', 6, '', 1, 'active', 0, NOW(), NOW()),
(7, 0, 'M', '缺陷管理', 'defects', 'Bell', '/defects', '', 7, '', 1, 'active', 0, NOW(), NOW()),
(8, 0, 'M', '台账和主数据管理', 'master_data', 'Collection', '', '', 8, '', 1, 'active', 0, NOW(), NOW()),
(9, 0, 'M', '系统管理', 'system', 'Setting', '', '', 99, '', 1, 'active', 0, NOW(), NOW());

-- 二级菜单 - 台账和主数据管理
INSERT IGNORE INTO permissions (id, pid, type, name, code, icon, path, component, sort, perms, visible, status, is_cache, created_at, updated_at) VALUES
(81, 8, 'C', '设备管理', 'devices', 'Monitor', '/devices', '', 1, '', 1, 'active', 0, NOW(), NOW());

-- 二级菜单 - 系统管理
INSERT IGNORE INTO permissions (id, pid, type, name, code, icon, path, component, sort, perms, visible, status, is_cache, created_at, updated_at) VALUES
(91, 9, 'C', '组织管理', 'organizations', 'OfficeBuilding', '/organizations', '', 1, '', 1, 'active', 0, NOW(), NOW()),
(92, 9, 'C', '用户管理', 'users', 'User', '/users', '', 2, '', 1, 'active', 0, NOW(), NOW()),
(93, 9, 'C', '角色管理', 'roles', 'Lock', '/roles', '', 3, '', 1, 'active', 0, NOW(), NOW()),
(94, 9, 'C', '菜单权限', 'menus', 'Setting', '/menus', '', 4, '', 1, 'active', 0, NOW(), NOW()),
(95, 9, 'C', '业务编码配置', 'business_codes', 'Key', '/business-codes', '', 5, '', 1, 'active', 0, NOW(), NOW()),
(96, 9, 'C', '实时监控', 'monitoring_logs', 'Monitor', '/monitoring/logs', '', 6, '', 1, 'active', 0, NOW(), NOW());

-- 按钮级权限 - 实时监控
INSERT IGNORE INTO permissions (id, pid, type, name, code, icon, path, component, sort, perms, visible, status, is_cache, created_at, updated_at) VALUES
(961, 96, 'A', '实时监控查看', 'monitor:view', '', '', '', 1, 'system:monitor:view', 1, 'active', 0, NOW(), NOW());

-- 按钮级权限 - 用户管理
INSERT IGNORE INTO permissions (id, pid, type, name, code, icon, path, component, sort, perms, visible, status, is_cache, created_at, updated_at) VALUES
(921, 92, 'A', '用户查询', 'user:list', '', '', '', 1, 'system:user:list', 1, 'active', 0, NOW(), NOW()),
(922, 92, 'A', '用户新增', 'user:add', '', '', '', 2, 'system:user:add', 1, 'active', 0, NOW(), NOW()),
(923, 92, 'A', '用户编辑', 'user:edit', '', '', '', 3, 'system:user:edit', 1, 'active', 0, NOW(), NOW()),
(924, 92, 'A', '用户删除', 'user:delete', '', '', '', 4, 'system:user:delete', 1, 'active', 0, NOW(), NOW()),
(925, 92, 'A', '用户启用', 'user:enable', '', '', '', 5, 'system:user:enable', 1, 'active', 0, NOW(), NOW()),
(926, 92, 'A', '用户停用', 'user:disable', '', '', '', 6, 'system:user:disable', 1, 'active', 0, NOW(), NOW());

-- 按钮级权限 - 角色管理
INSERT IGNORE INTO permissions (id, pid, type, name, code, icon, path, component, sort, perms, visible, status, is_cache, created_at, updated_at) VALUES
(931, 93, 'A', '角色查询', 'role:list', '', '', '', 1, 'system:role:list', 1, 'active', 0, NOW(), NOW()),
(932, 93, 'A', '角色新增', 'role:add', '', '', '', 2, 'system:role:add', 1, 'active', 0, NOW(), NOW()),
(933, 93, 'A', '角色编辑', 'role:edit', '', '', '', 3, 'system:role:edit', 1, 'active', 0, NOW(), NOW()),
(934, 93, 'A', '角色删除', 'role:delete', '', '', '', 4, 'system:role:delete', 1, 'active', 0, NOW(), NOW()),
(935, 93, 'A', '分配权限', 'role:assign', '', '', '', 5, 'system:role:assign', 1, 'active', 0, NOW(), NOW());

-- 按钮级权限 - 菜单管理
INSERT IGNORE INTO permissions (id, pid, type, name, code, icon, path, component, sort, perms, visible, status, is_cache, created_at, updated_at) VALUES
(941, 94, 'A', '菜单查询', 'menu:list', '', '', '', 1, 'system:menu:list', 1, 'active', 0, NOW(), NOW()),
(942, 94, 'A', '菜单新增', 'menu:add', '', '', '', 2, 'system:menu:add', 1, 'active', 0, NOW(), NOW()),
(943, 94, 'A', '菜单编辑', 'menu:edit', '', '', '', 3, 'system:menu:edit', 1, 'active', 0, NOW(), NOW()),
(944, 94, 'A', '菜单删除', 'menu:delete', '', '', '', 4, 'system:menu:delete', 1, 'active', 0, NOW(), NOW());

-- 按钮级权限 - 组织管理
INSERT IGNORE INTO permissions (id, pid, type, name, code, icon, path, component, sort, perms, visible, status, is_cache, created_at, updated_at) VALUES
(911, 91, 'A', '组织查询', 'org:list', '', '', '', 1, 'system:org:list', 1, 'active', 0, NOW(), NOW()),
(912, 91, 'A', '组织新增', 'org:add', '', '', '', 2, 'system:org:add', 1, 'active', 0, NOW(), NOW()),
(913, 91, 'A', '组织编辑', 'org:edit', '', '', '', 3, 'system:org:edit', 1, 'active', 0, NOW(), NOW()),
(914, 91, 'A', '组织删除', 'org:delete', '', '', '', 4, 'system:org:delete', 1, 'active', 0, NOW(), NOW());

-- ================================
-- 3. 初始化角色数据
-- ================================

INSERT IGNORE INTO roles (id, name, code, description, sort, status, created_at, updated_at) VALUES
(1, '超级管理员', 'super_admin', '系统超级管理员，拥有所有权限', 1, 'active', NOW(), NOW()),
(2, '普通用户', 'user', '普通业务用户', 2, 'active', NOW(), NOW()),
(3, '巡检人员', 'inspector', '负责执行巡检任务', 3, 'active', NOW(), NOW());

-- ================================
-- 4. 为超级管理员分配所有菜单权限
-- ================================

-- 先删除旧的关联
DELETE FROM role_permissions WHERE role_id = 1;

-- 超级管理员拥有所有菜单和目录权限，并额外拥有实时监控查看权限。
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions WHERE type IN ('M', 'C') OR perms = 'system:monitor:view' ORDER BY id;

-- 如果需要也包含按钮级权限，使用下面的语句
-- INSERT INTO role_permissions (role_id, permission_id)
-- SELECT 1, id FROM permissions ORDER BY id;

-- ================================
-- 5. 更新 admin 用户的角色为超级管理员
-- ================================

UPDATE users SET role_id = 1 WHERE username = 'admin';

-- ================================
-- 执行完成说明
-- ================================
-- 执行完成后：
-- 1. 使用 admin 账号登录，应该能看到所有系统菜单
-- 2. 进入"系统管理 → 角色管理"可以创建新角色并分配权限
-- 3. 进入"系统管理 → 菜单权限"可以添加新菜单或按钮权限
-- 4. 进入"系统管理 → 用户管理"可以为用户分配角色
