-- 为业务编码规则增加可选组织机构编码段。
-- 运行时 schema 由 GORM AutoMigrate 同步；本文件用于手工部署和 schema 说明。

ALTER TABLE business_codes
  ADD COLUMN use_org_code TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在编码中添加组织机构编码' AFTER use_separator,
  ADD COLUMN org_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '组织机构编码，启用组织编码时写入生成结果' AFTER use_org_code;

CREATE INDEX idx_business_codes_org_code ON business_codes (org_code);
