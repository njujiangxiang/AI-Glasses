CREATE TABLE IF NOT EXISTS organizations (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(64) NOT NULL UNIQUE,
  name VARCHAR(128) NOT NULL,
  parent_code VARCHAR(64) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  INDEX idx_organizations_parent_code (parent_code),
  INDEX idx_organizations_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE users
  ADD COLUMN name VARCHAR(64) NOT NULL DEFAULT '',
  ADD COLUMN gender VARCHAR(8) NOT NULL DEFAULT '',
  ADD COLUMN avatar_data LONGBLOB NULL,
  ADD COLUMN avatar_content_type VARCHAR(64) NOT NULL DEFAULT '',
  ADD COLUMN avatar_size BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN birth_year INT NOT NULL DEFAULT 0,
  ADD COLUMN birth_month INT NOT NULL DEFAULT 0,
  ADD COLUMN id_card_no VARCHAR(32) NOT NULL DEFAULT '',
  ADD COLUMN org_code VARCHAR(64) NOT NULL DEFAULT '',
  ADD INDEX idx_users_id_card_no (id_card_no),
  ADD INDEX idx_users_org_code (org_code);
