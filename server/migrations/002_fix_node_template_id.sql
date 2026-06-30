-- 修复节点模板ID：将 template_id=0 的记录更新为 NULL（表示未分配）
-- 这是为了支持独立节点管理功能

-- 1. 修改表结构：允许 template_id 为 NULL
ALTER TABLE inspection_template_nodes
  MODIFY COLUMN template_id BIGINT UNSIGNED NULL
  COMMENT '巡检模板ID，关联inspection_templates.id，NULL表示未分配';

-- 2. 修复现有数据：将 template_id=0 的记录更新为 NULL
UPDATE inspection_template_nodes
SET template_id = NULL
WHERE template_id = 0;
