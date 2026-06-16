CREATE TABLE IF NOT EXISTS business_codes (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '业务编码配置ID，系统内部主键' PRIMARY KEY,
  name VARCHAR(128) NOT NULL COMMENT '编码名称，用于后台展示规则用途',
  code VARCHAR(64) NOT NULL UNIQUE COMMENT '业务代码，系统内唯一，例如TK',
  date_format VARCHAR(32) NOT NULL COMMENT '日期格式，首版仅支持yyyyMMdd',
  seq_padding BIGINT NOT NULL COMMENT '流水号位数，例如4表示0001',
  separator VARCHAR(8) NOT NULL DEFAULT '' COMMENT '分隔符，例如-，不使用时为空',
  use_separator BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否在代码、日期、流水号之间使用分隔符',
  status VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '编码状态：active启用，disabled停用',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务编码配置表：保存按日生成业务编号所需的代码、日期格式、流水号位数和启停状态';
