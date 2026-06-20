-- 权限表迁移脚本 - 添加缺失字段
-- 执行此脚本来更新 permissions 表结构以匹配新的 RBAC 模型

-- 1. 添加 pid 字段（父级ID）
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS pid BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父级菜单ID' AFTER id;

-- 2. 添加 type 字段（菜单类型）
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS type VARCHAR(16) NOT NULL DEFAULT 'menu' COMMENT '类型: M=目录, C=菜单, A=按钮' AFTER pid;

-- 3. 添加 icon 字段
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS icon VARCHAR(64) NOT NULL DEFAULT '' COMMENT '图标' AFTER code;

-- 4. 添加 path 字段
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS path VARCHAR(255) NOT NULL DEFAULT '' COMMENT '路由路径' AFTER icon;

-- 5. 添加 component 字段
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS component VARCHAR(255) NOT NULL DEFAULT '' COMMENT '前端组件' AFTER path;

-- 6. 添加 sort 字段
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS sort INT NOT NULL DEFAULT 0 COMMENT '排序' AFTER component;

-- 7. 添加 perms 字段
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS perms VARCHAR(255) NOT NULL DEFAULT '' COMMENT '权限标识' AFTER sort;

-- 8. 添加 visible 字段
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS visible TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否显示: 0=隐藏, 1=显示' AFTER perms;

-- 9. 添加 is_cache 字段
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS is_cache TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否缓存' AFTER visible;

-- 10. 确保 status 字段存在
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '状态' AFTER is_cache;

-- 11. 添加索引
ALTER TABLE permissions ADD INDEX IF NOT EXISTS idx_pid (pid);
ALTER TABLE permissions ADD INDEX IF NOT EXISTS idx_status (status);

-- ================================
-- 执行完成后再次运行初始化脚本:
-- go run cmd/init_rbac/main.go
-- ================================
