-- 为业务编码规则增加是否添加日期开关。
-- 运行时 schema 由 GORM AutoMigrate 同步；本文件用于手工部署和 schema 说明。

ALTER TABLE business_codes
  ADD COLUMN use_date TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否在编码中添加日期' AFTER code;
