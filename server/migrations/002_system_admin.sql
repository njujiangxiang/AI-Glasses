CREATE TABLE IF NOT EXISTS organizations (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '组织ID，系统内部主键' PRIMARY KEY,
  code VARCHAR(64) NOT NULL UNIQUE COMMENT '单位编码，系统内唯一，用于组织树和用户所属单位关联',
  name VARCHAR(128) NOT NULL COMMENT '单位名称，用于页面展示和登录用户公司名称展示',
  parent_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '上级单位编码，空字符串表示顶级单位',
  status VARCHAR(32) NOT NULL COMMENT '组织状态：active启用，disabled停用',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  INDEX idx_organizations_parent_code (parent_code),
  INDEX idx_organizations_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='组织表：保存单位编码、单位名称和上下级单位关系，用于系统管理和用户归属';

ALTER TABLE users
  ADD COLUMN name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '姓名，用户真实姓名',
  ADD COLUMN gender VARCHAR(8) NOT NULL DEFAULT '' COMMENT '性别：male男，female女，unknown未知',
  ADD COLUMN avatar_data LONGBLOB NULL COMMENT '用户头像二进制数据，直接保存到数据库',
  ADD COLUMN avatar_content_type VARCHAR(64) NOT NULL DEFAULT '' COMMENT '头像MIME类型，例如image/jpeg、image/png、image/webp',
  ADD COLUMN avatar_size BIGINT NOT NULL DEFAULT 0 COMMENT '头像大小，单位字节；0表示未上传头像',
  ADD COLUMN birth_year INT NOT NULL DEFAULT 0 COMMENT '出生年份，0表示未填写',
  ADD COLUMN birth_month INT NOT NULL DEFAULT 0 COMMENT '出生月份，1-12，0表示未填写',
  ADD COLUMN id_card_no VARCHAR(32) NOT NULL DEFAULT '' COMMENT '身份证号码，非空时由服务层校验格式和唯一性',
  ADD COLUMN org_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属单位编码，关联organizations.code',
  ADD INDEX idx_users_id_card_no (id_card_no),
  ADD INDEX idx_users_org_code (org_code);
