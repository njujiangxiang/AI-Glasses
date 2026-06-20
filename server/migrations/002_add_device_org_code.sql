-- 为设备表添加 org_code 字段，用于关联设备所属组织机构
ALTER TABLE devices ADD COLUMN org_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '设备所属组织机构编码' AFTER name;
ALTER TABLE devices ADD INDEX idx_devices_org_code (org_code);
