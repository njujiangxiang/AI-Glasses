CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(128) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  INDEX idx_users_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS roles (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS permissions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(128) NOT NULL UNIQUE,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_roles (user_id BIGINT UNSIGNED NOT NULL, role_id BIGINT UNSIGNED NOT NULL, PRIMARY KEY (user_id, role_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS role_permissions (role_id BIGINT UNSIGNED NOT NULL, permission_id BIGINT UNSIGNED NOT NULL, PRIMARY KEY (role_id, permission_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS teams (id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, name VARCHAR(128) NOT NULL, created_at DATETIME(3) NULL, updated_at DATETIME(3) NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS team_members (team_id BIGINT UNSIGNED NOT NULL, user_id BIGINT UNSIGNED NOT NULL, PRIMARY KEY (team_id, user_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS devices (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  serial_no VARCHAR(128) NOT NULL UNIQUE,
  name VARCHAR(128) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL,
  bound_user_id BIGINT UNSIGNED NULL,
  bound_at DATETIME(3) NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  INDEX idx_devices_status_bound_user_id (status, bound_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS device_sessions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  device_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  refresh_jti VARCHAR(64) NOT NULL UNIQUE,
  status VARCHAR(32) NOT NULL,
  refresh_until DATETIME(3) NOT NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  INDEX idx_device_sessions_device_id (device_id),
  INDEX idx_device_sessions_user_id (user_id),
  INDEX idx_device_sessions_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS device_audit_logs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  device_id BIGINT UNSIGNED NOT NULL,
  actor_id BIGINT UNSIGNED NOT NULL,
  action VARCHAR(64) NOT NULL,
  reason VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME(3) NULL,
  INDEX idx_device_audit_logs_device_id (device_id),
  INDEX idx_device_audit_logs_actor_id (actor_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS inspection_templates (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  description VARCHAR(512) NOT NULL DEFAULT '',
  applicable_roles VARCHAR(255) NOT NULL DEFAULT '',
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  INDEX idx_inspection_templates_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS inspection_template_nodes (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  template_id BIGINT UNSIGNED NOT NULL,
  sort_order INT NOT NULL,
  name VARCHAR(128) NOT NULL,
  description VARCHAR(512) NOT NULL DEFAULT '',
  node_type VARCHAR(32) NOT NULL,
  min_photos INT NOT NULL DEFAULT 0,
  require_text BOOLEAN NOT NULL DEFAULT FALSE,
  allow_abnormal BOOLEAN NOT NULL DEFAULT FALSE,
  require_live_capture BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  UNIQUE KEY idx_template_sort (template_id, sort_order),
  INDEX idx_inspection_template_nodes_template_id (template_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS task_plans (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  template_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(128) NOT NULL,
  cron_expr VARCHAR(64) NOT NULL,
  timezone VARCHAR(64) NOT NULL,
  start_at DATETIME(3) NOT NULL,
  due_duration_minutes INT NOT NULL,
  assignee_type VARCHAR(16) NOT NULL,
  assignee_id BIGINT UNSIGNED NOT NULL,
  point_name VARCHAR(128) NOT NULL DEFAULT '',
  equipment_name VARCHAR(128) NOT NULL DEFAULT '',
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  INDEX idx_task_plans_template_id (template_id),
  INDEX idx_task_plans_assignee_id (assignee_id),
  INDEX idx_task_plans_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS inspection_tasks (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  plan_id BIGINT UNSIGNED NOT NULL,
  template_id BIGINT UNSIGNED NOT NULL,
  scheduled_at DATETIME(3) NOT NULL,
  due_at DATETIME(3) NOT NULL,
  status VARCHAR(32) NOT NULL,
  assignee_type VARCHAR(16) NOT NULL,
  assignee_id BIGINT UNSIGNED NOT NULL,
  executor_id BIGINT UNSIGNED NULL,
  point_name VARCHAR(128) NOT NULL DEFAULT '',
  equipment_name VARCHAR(128) NOT NULL DEFAULT '',
  started_at DATETIME(3) NULL,
  submitted_at DATETIME(3) NULL,
  completed_at DATETIME(3) NULL,
  cancelled_at DATETIME(3) NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  UNIQUE KEY idx_task_plan_schedule_assignee (plan_id, scheduled_at, assignee_type, assignee_id),
  INDEX idx_inspection_tasks_assignee_id_status_due_id (assignee_id, status, due_at, id),
  INDEX idx_task_status_due_id (status, due_at, id),
  INDEX idx_inspection_tasks_executor_id (executor_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS inspection_task_nodes (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  task_id BIGINT UNSIGNED NOT NULL,
  template_node_id BIGINT UNSIGNED NOT NULL,
  sort_order INT NOT NULL,
  name VARCHAR(128) NOT NULL,
  node_type VARCHAR(32) NOT NULL,
  min_photos INT NOT NULL DEFAULT 0,
  require_text BOOLEAN NOT NULL DEFAULT FALSE,
  allow_abnormal BOOLEAN NOT NULL DEFAULT FALSE,
  status VARCHAR(32) NOT NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  UNIQUE KEY idx_task_node_sort (task_id, sort_order),
  INDEX idx_inspection_task_nodes_task_id (task_id),
  INDEX idx_inspection_task_nodes_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS task_node_results (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  task_id BIGINT UNSIGNED NOT NULL,
  node_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL,
  text_note TEXT NULL,
  idempotency_key VARCHAR(128) NOT NULL,
  completed_at DATETIME(3) NOT NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  UNIQUE KEY idx_result_task_node (task_id, node_id),
  INDEX idx_task_node_results_user_id (user_id),
  INDEX idx_task_node_results_idempotency_key (idempotency_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS attachments (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  object_key VARCHAR(255) NOT NULL UNIQUE,
  file_name VARCHAR(255) NOT NULL DEFAULT '',
  content_type VARCHAR(128) NOT NULL,
  size_bytes BIGINT NOT NULL,
  sha256 VARCHAR(64) NOT NULL DEFAULT '',
  bind_status VARCHAR(32) NOT NULL,
  task_id BIGINT UNSIGNED NULL,
  node_id BIGINT UNSIGNED NULL,
  result_id BIGINT UNSIGNED NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  device_id BIGINT UNSIGNED NULL,
  capture_time DATETIME(3) NULL,
  upload_time DATETIME(3) NULL,
  gps_lat DOUBLE NULL,
  gps_lng DOUBLE NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  INDEX idx_attachment_bind_created (bind_status, created_at),
  INDEX idx_attachments_task_id (task_id),
  INDEX idx_attachments_node_id (node_id),
  INDEX idx_attachments_result_id (result_id),
  INDEX idx_attachments_user_id (user_id),
  INDEX idx_attachments_device_id (device_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS defects (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  task_id BIGINT UNSIGNED NOT NULL,
  node_id BIGINT UNSIGNED NOT NULL,
  reporter_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL,
  description TEXT NULL,
  close_reason TEXT NULL,
  confirmed_at DATETIME(3) NULL,
  closed_at DATETIME(3) NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  INDEX idx_defects_status_created_id (status, created_at, id),
  INDEX idx_defects_task_id (task_id),
  INDEX idx_defects_node_id (node_id),
  INDEX idx_defects_reporter_id (reporter_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  actor_id BIGINT UNSIGNED NOT NULL,
  action VARCHAR(128) NOT NULL,
  target VARCHAR(128) NOT NULL,
  target_id BIGINT UNSIGNED NOT NULL,
  detail TEXT NULL,
  created_at DATETIME(3) NULL,
  INDEX idx_audit_logs_actor_id (actor_id),
  INDEX idx_audit_logs_target_id (target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS outbox_events (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  event_key VARCHAR(160) NOT NULL UNIQUE,
  topic VARCHAR(128) NOT NULL,
  payload JSON NOT NULL,
  published_at DATETIME(3) NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  INDEX idx_outbox_events_topic (topic)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
