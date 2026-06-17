-- 工作流配置表
CREATE TABLE IF NOT EXISTS workflows (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '工作流ID' PRIMARY KEY,
  name VARCHAR(128) NOT NULL COMMENT '工作流名称',
  description VARCHAR(512) COMMENT '工作流描述',
  status VARCHAR(32) NOT NULL DEFAULT 'draft' COMMENT '状态：draft草稿 published已发布 archived已归档',
  created_by BIGINT UNSIGNED NOT NULL COMMENT '创建人ID',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工作流配置表：定义巡检检查步骤和异常触发规则';

-- 工作流步骤表
CREATE TABLE IF NOT EXISTS workflow_steps (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '步骤ID' PRIMARY KEY,
  workflow_id BIGINT UNSIGNED NOT NULL COMMENT '所属工作流ID',
  sort_order INT NOT NULL COMMENT '排序序号',
  name VARCHAR(128) NOT NULL COMMENT '步骤名称',
  description VARCHAR(512) COMMENT '步骤描述',
  type VARCHAR(32) NOT NULL COMMENT '步骤类型：text文本输入 number数值输入 select选择清单 photo拍照 video录像 audio录音',
  required BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否必填',
  options_json JSON COMMENT '选择项配置，仅select类型使用',
  abnormal_enabled BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否启用异常触发',
  abnormal_require_photo BOOLEAN NOT NULL DEFAULT TRUE COMMENT '异常时必须拍照',
  abnormal_require_video BOOLEAN NOT NULL DEFAULT FALSE COMMENT '异常时必须录像',
  abnormal_require_note BOOLEAN NOT NULL DEFAULT TRUE COMMENT '异常时必须填写备注',
  abnormal_require_signature BOOLEAN NOT NULL DEFAULT FALSE COMMENT '异常时必须签字确认',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  UNIQUE KEY idx_workflow_sort (workflow_id, sort_order),
  INDEX idx_workflow_id (workflow_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工作流步骤表：定义每个检查步骤的输入类型、是否必填和异常触发规则';
