-- 为业务编码配置表添加 org_source 字段，支持选择使用当前登录人所在机构
ALTER TABLE business_codes ADD COLUMN org_source VARCHAR(32) NOT NULL DEFAULT 'fixed' COMMENT '组织编码来源：fixed固定编码，current当前登录人组织' AFTER org_code;
