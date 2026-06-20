-- 修复 roles 表 code 字段问题
-- 在迁移前执行此脚本

-- 1. 先为已有空 code 的数据填充默认值
UPDATE roles SET code = CONCAT('role_', id) WHERE code = '' OR code IS NULL;

-- 2. 如果存在重复的 name，也需要处理
UPDATE r1 SET r1.code = CONCAT(r1.code, '_', r1.id)
FROM roles r1
INNER JOIN roles r2 ON r1.name = r2.name AND r1.id > r2.id;

-- 执行完成后再运行迁移:
-- go run cmd/migrate/main.go
