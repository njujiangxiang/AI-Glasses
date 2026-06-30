CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '用户ID，系统内部主键' PRIMARY KEY,
  username VARCHAR(64) NOT NULL UNIQUE COMMENT '用户名，用于后台或眼镜端登录，系统内唯一',
  password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希；当前MVP为开发态占位值',
  display_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '显示名称，用于页面展示',
  status VARCHAR(32) NOT NULL COMMENT '用户状态：active启用，disabled停用',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  INDEX idx_users_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表：保存后台管理员、任务人员和眼镜端巡检员的基础账号信息';

CREATE TABLE IF NOT EXISTS roles (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '角色ID，系统内部主键' PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE COMMENT '角色名称，例如系统管理员、任务管理员、班组长、巡检员',
  code VARCHAR(64) NOT NULL DEFAULT '' UNIQUE COMMENT '角色编码，例如super_admin、user、inspector',
  description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '角色说明',
  data_scope VARCHAR(32) NOT NULL DEFAULT 'org_only' COMMENT '数据范围：all全部，org_and_sub本组织及下级，org_only本组织，self_only仅自己',
  sort INT NOT NULL DEFAULT 0 COMMENT '排序值，数值越小越靠前',
  status VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '角色状态：active启用，disabled停用',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  INDEX idx_roles_data_scope (data_scope),
  INDEX idx_roles_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表：定义系统内可分配给用户的角色';

CREATE TABLE IF NOT EXISTS permissions (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '权限ID，系统内部主键' PRIMARY KEY,
  code VARCHAR(128) NOT NULL UNIQUE COMMENT '权限编码，例如admin:*、admin:tasks、glasses:tasks',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表：定义系统接口和功能权限编码';

CREATE TABLE IF NOT EXISTS user_roles (
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID，关联users.id',
  role_id BIGINT UNSIGNED NOT NULL COMMENT '角色ID，关联roles.id',
  PRIMARY KEY (user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表：维护用户和角色的多对多关系';

CREATE TABLE IF NOT EXISTS role_permissions (
  role_id BIGINT UNSIGNED NOT NULL COMMENT '角色ID，关联roles.id',
  permission_id BIGINT UNSIGNED NOT NULL COMMENT '权限ID，关联permissions.id',
  PRIMARY KEY (role_id, permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表：维护角色和权限的多对多关系';

CREATE TABLE IF NOT EXISTS teams (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '班组ID，系统内部主键' PRIMARY KEY,
  name VARCHAR(128) NOT NULL COMMENT '班组名称，例如A区巡检班组',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='班组表：保存巡检任务可分配的班组';

CREATE TABLE IF NOT EXISTS team_members (
  team_id BIGINT UNSIGNED NOT NULL COMMENT '班组ID，关联teams.id',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID，关联users.id',
  PRIMARY KEY (team_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='班组成员表：维护班组和用户的成员关系';

CREATE TABLE IF NOT EXISTS devices (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '设备ID，系统内部主键' PRIMARY KEY,
  serial_no VARCHAR(128) NOT NULL UNIQUE COMMENT '智能眼镜设备序列号，系统内唯一',
  name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '设备名称或备注名称',
  status VARCHAR(32) NOT NULL COMMENT '设备状态：pending待绑定，active启用，revoked撤销，lost_disabled丢失禁用',
  bound_user_id BIGINT UNSIGNED NULL COMMENT '当前绑定用户ID，关联users.id',
  bound_at DATETIME(3) NULL COMMENT '绑定时间',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  INDEX idx_devices_status_bound_user_id (status, bound_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备表：保存智能眼镜设备登记、绑定和状态信息';

CREATE TABLE IF NOT EXISTS device_sessions (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '设备会话ID，系统内部主键' PRIMARY KEY,
  device_id BIGINT UNSIGNED NOT NULL COMMENT '设备ID，关联devices.id',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID，关联users.id',
  refresh_jti VARCHAR(64) NOT NULL UNIQUE COMMENT '刷新令牌唯一标识，用于设备会话续期和撤销',
  status VARCHAR(32) NOT NULL COMMENT '会话状态：active有效，revoked已撤销',
  refresh_until DATETIME(3) NOT NULL COMMENT '刷新令牌有效截止时间',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  INDEX idx_device_sessions_device_id (device_id),
  INDEX idx_device_sessions_user_id (user_id),
  INDEX idx_device_sessions_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备会话表：保存眼镜端登录和刷新令牌状态';

CREATE TABLE IF NOT EXISTS device_audit_logs (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '设备审计日志ID，系统内部主键' PRIMARY KEY,
  device_id BIGINT UNSIGNED NOT NULL COMMENT '设备ID，关联devices.id',
  actor_id BIGINT UNSIGNED NOT NULL COMMENT '操作人用户ID，关联users.id',
  action VARCHAR(64) NOT NULL COMMENT '操作动作，例如bind、revoke、disable_lost',
  reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作原因或备注',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  INDEX idx_device_audit_logs_device_id (device_id),
  INDEX idx_device_audit_logs_actor_id (actor_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备审计日志表：记录设备绑定、撤销、丢失禁用等操作';

CREATE TABLE IF NOT EXISTS inspection_templates (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '巡检模板ID，系统内部主键' PRIMARY KEY,
  name VARCHAR(128) NOT NULL COMMENT '模板名称',
  description VARCHAR(512) NOT NULL DEFAULT '' COMMENT '模板说明',
  applicable_roles VARCHAR(255) NOT NULL DEFAULT '' COMMENT '适用角色，使用逗号或文本记录角色范围',
  enabled BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否启用模板',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  INDEX idx_inspection_templates_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检模板表：定义可生成巡检任务的模板主信息';

CREATE TABLE IF NOT EXISTS inspection_template_nodes (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '模板节点ID，系统内部主键' PRIMARY KEY,
  template_id BIGINT UNSIGNED NULL COMMENT '巡检模板ID，关联inspection_templates.id，NULL表示未分配',
  sort_order INT NOT NULL COMMENT '节点排序号，数值越小越靠前',
  name VARCHAR(128) NOT NULL COMMENT '节点名称',
  description VARCHAR(512) NOT NULL DEFAULT '' COMMENT '节点说明',
  node_type VARCHAR(32) NOT NULL COMMENT '节点类型：checkin签到，photo拍照，text文本，number数值，abnormal异常，confirm确认',
  min_photos INT NOT NULL DEFAULT 0 COMMENT '最少照片数量要求',
  require_text BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否要求填写文本说明',
  allow_abnormal BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否允许在该节点上报异常',
  require_live_capture BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否要求现场实时拍摄',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  UNIQUE KEY idx_template_sort (template_id, sort_order),
  INDEX idx_inspection_template_nodes_template_id (template_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检模板节点表：定义模板下的巡检步骤、顺序和提交要求';

CREATE TABLE IF NOT EXISTS task_plans (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '任务计划ID，系统内部主键' PRIMARY KEY,
  template_id BIGINT UNSIGNED NOT NULL COMMENT '巡检模板ID，关联inspection_templates.id',
  name VARCHAR(128) NOT NULL COMMENT '计划名称',
  cron_expr VARCHAR(64) NOT NULL COMMENT 'Cron表达式，用于周期性生成巡检任务',
  timezone VARCHAR(64) NOT NULL COMMENT '计划执行时区，例如Asia/Shanghai',
  start_at DATETIME(3) NOT NULL COMMENT '计划开始生效时间',
  due_duration_minutes INT NOT NULL COMMENT '任务生成后多少分钟内应完成',
  assignee_type VARCHAR(16) NOT NULL COMMENT '指派类型：user用户，team班组',
  assignee_id BIGINT UNSIGNED NOT NULL COMMENT '指派对象ID，根据assignee_type关联users.id或teams.id',
  point_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '巡检点位名称',
  equipment_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '巡检设备名称',
  enabled BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否启用计划',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  INDEX idx_task_plans_template_id (template_id),
  INDEX idx_task_plans_assignee_id (assignee_id),
  INDEX idx_task_plans_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务计划表：按模板和周期规则生成具体巡检任务';

CREATE TABLE IF NOT EXISTS inspection_tasks (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '巡检任务ID，系统内部主键' PRIMARY KEY,
  plan_id BIGINT UNSIGNED NOT NULL COMMENT '任务计划ID，关联task_plans.id',
  template_id BIGINT UNSIGNED NOT NULL COMMENT '巡检模板ID，关联inspection_templates.id',
  scheduled_at DATETIME(3) NOT NULL COMMENT '计划执行时间',
  due_at DATETIME(3) NOT NULL COMMENT '任务截止时间',
  status VARCHAR(32) NOT NULL COMMENT '任务状态：pending待领取，assigned已分配，in_progress执行中，submitted已提交，completed已完成，overdue已逾期，cancelled已取消',
  assignee_type VARCHAR(16) NOT NULL COMMENT '指派类型：user用户，team班组',
  assignee_id BIGINT UNSIGNED NOT NULL COMMENT '指派对象ID，根据assignee_type关联users.id或teams.id',
  executor_id BIGINT UNSIGNED NULL COMMENT '实际执行人用户ID，领取班组任务后写入',
  point_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '巡检点位名称，生成任务时从计划复制',
  equipment_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '巡检设备名称，生成任务时从计划复制',
  started_at DATETIME(3) NULL COMMENT '开始执行时间',
  submitted_at DATETIME(3) NULL COMMENT '提交时间',
  completed_at DATETIME(3) NULL COMMENT '后台确认完成时间',
  cancelled_at DATETIME(3) NULL COMMENT '取消时间',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  UNIQUE KEY idx_task_plan_schedule_assignee (plan_id, scheduled_at, assignee_type, assignee_id),
  INDEX idx_inspection_tasks_assignee_id_status_due_id (assignee_id, status, due_at, id),
  INDEX idx_task_status_due_id (status, due_at, id),
  INDEX idx_inspection_tasks_executor_id (executor_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检任务表：保存由计划生成并由眼镜端执行的具体任务';

CREATE TABLE IF NOT EXISTS inspection_task_nodes (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '任务节点ID，系统内部主键' PRIMARY KEY,
  task_id BIGINT UNSIGNED NOT NULL COMMENT '巡检任务ID，关联inspection_tasks.id',
  template_node_id BIGINT UNSIGNED NOT NULL COMMENT '模板节点ID，关联inspection_template_nodes.id',
  sort_order INT NOT NULL COMMENT '节点排序号，数值越小越靠前',
  name VARCHAR(128) NOT NULL COMMENT '任务节点名称，生成任务时从模板节点复制',
  node_type VARCHAR(32) NOT NULL COMMENT '节点类型：checkin签到，photo拍照，text文本，number数值，abnormal异常，confirm确认',
  min_photos INT NOT NULL DEFAULT 0 COMMENT '最少照片数量要求',
  require_text BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否要求填写文本说明',
  allow_abnormal BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否允许在该节点上报异常',
  status VARCHAR(32) NOT NULL COMMENT '节点状态：pending待提交，completed已完成，abnormal异常',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  UNIQUE KEY idx_task_node_sort (task_id, sort_order),
  INDEX idx_inspection_task_nodes_task_id (task_id),
  INDEX idx_inspection_task_nodes_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检任务节点表：保存具体任务下每个巡检步骤的执行要求和状态';

CREATE TABLE IF NOT EXISTS task_node_results (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '节点结果ID，系统内部主键' PRIMARY KEY,
  task_id BIGINT UNSIGNED NOT NULL COMMENT '巡检任务ID，关联inspection_tasks.id',
  node_id BIGINT UNSIGNED NOT NULL COMMENT '任务节点ID，关联inspection_task_nodes.id',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '提交人用户ID，关联users.id',
  status VARCHAR(32) NOT NULL COMMENT '结果状态：completed正常完成，abnormal异常',
  text_note TEXT NULL COMMENT '节点文字说明或备注',
  idempotency_key VARCHAR(128) NOT NULL COMMENT '幂等键，用于防止弱网重复提交',
  completed_at DATETIME(3) NOT NULL COMMENT '节点完成时间',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  UNIQUE KEY idx_result_task_node (task_id, node_id),
  INDEX idx_task_node_results_user_id (user_id),
  INDEX idx_task_node_results_idempotency_key (idempotency_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务节点结果表：保存眼镜端提交的节点执行结果';

CREATE TABLE IF NOT EXISTS attachments (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '附件ID，系统内部主键' PRIMARY KEY,
  object_key VARCHAR(255) NOT NULL UNIQUE COMMENT '对象存储Key，用于定位MinIO/S3中的文件',
  file_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原始文件名',
  content_type VARCHAR(128) NOT NULL COMMENT '文件MIME类型，例如image/jpeg、audio/mpeg',
  size_bytes BIGINT NOT NULL COMMENT '文件大小，单位字节',
  sha256 VARCHAR(64) NOT NULL DEFAULT '' COMMENT '文件SHA-256摘要，用于完整性校验',
  bind_status VARCHAR(32) NOT NULL COMMENT '绑定状态：uploaded已上传，bound已绑定，orphaned孤立',
  task_id BIGINT UNSIGNED NULL COMMENT '关联巡检任务ID，关联inspection_tasks.id',
  node_id BIGINT UNSIGNED NULL COMMENT '关联任务节点ID，关联inspection_task_nodes.id',
  result_id BIGINT UNSIGNED NULL COMMENT '关联节点结果ID，关联task_node_results.id',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '上传人用户ID，关联users.id',
  device_id BIGINT UNSIGNED NULL COMMENT '上传设备ID，关联devices.id',
  capture_time DATETIME(3) NULL COMMENT '现场采集时间',
  upload_time DATETIME(3) NULL COMMENT '上传完成时间',
  gps_lat DOUBLE NULL COMMENT '采集位置纬度',
  gps_lng DOUBLE NULL COMMENT '采集位置经度',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  INDEX idx_attachment_bind_created (bind_status, created_at),
  INDEX idx_attachments_task_id (task_id),
  INDEX idx_attachments_node_id (node_id),
  INDEX idx_attachments_result_id (result_id),
  INDEX idx_attachments_user_id (user_id),
  INDEX idx_attachments_device_id (device_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='附件表：保存眼镜端照片、音频等证据文件的元数据和绑定关系';

CREATE TABLE IF NOT EXISTS defects (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '缺陷ID，系统内部主键' PRIMARY KEY,
  task_id BIGINT UNSIGNED NOT NULL COMMENT '巡检任务ID，关联inspection_tasks.id',
  node_id BIGINT UNSIGNED NOT NULL COMMENT '任务节点ID，关联inspection_task_nodes.id',
  reporter_id BIGINT UNSIGNED NOT NULL COMMENT '上报人用户ID，关联users.id',
  status VARCHAR(32) NOT NULL COMMENT '缺陷状态：reported已上报，confirmed已确认，closed已关闭',
  description TEXT NULL COMMENT '缺陷描述',
  close_reason TEXT NULL COMMENT '关闭原因',
  confirmed_at DATETIME(3) NULL COMMENT '确认时间',
  closed_at DATETIME(3) NULL COMMENT '关闭时间',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  INDEX idx_defects_status_created_id (status, created_at, id),
  INDEX idx_defects_task_id (task_id),
  INDEX idx_defects_node_id (node_id),
  INDEX idx_defects_reporter_id (reporter_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='缺陷表：保存巡检过程中发现并上报的异常缺陷';

CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '审计日志ID，系统内部主键' PRIMARY KEY,
  actor_id BIGINT UNSIGNED NOT NULL COMMENT '操作人用户ID，关联users.id',
  action VARCHAR(128) NOT NULL COMMENT '操作动作，例如create、update、delete、submit',
  target VARCHAR(128) NOT NULL COMMENT '操作对象类型，例如task、device、template',
  target_id BIGINT UNSIGNED NOT NULL COMMENT '操作对象ID',
  detail TEXT NULL COMMENT '操作详情，通常为JSON或文本说明',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  INDEX idx_audit_logs_actor_id (actor_id),
  INDEX idx_audit_logs_target_id (target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审计日志表：记录后台关键业务操作，用于追踪和审计';

CREATE TABLE IF NOT EXISTS outbox_events (
  id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '事件ID，系统内部主键' PRIMARY KEY,
  event_key VARCHAR(160) NOT NULL UNIQUE COMMENT '事件唯一键，用于保证事件幂等写入和发布',
  topic VARCHAR(128) NOT NULL COMMENT '事件主题或队列名，例如task.assigned',
  payload JSON NOT NULL COMMENT '事件载荷JSON，保存异步消息内容',
  published_at DATETIME(3) NULL COMMENT '发布时间，未发布时为空',
  created_at DATETIME(3) NULL COMMENT '创建时间',
  updated_at DATETIME(3) NULL COMMENT '更新时间',
  INDEX idx_outbox_events_topic (topic)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Outbox事件表：保存待发布的异步业务事件，用于任务分配等消息集成';
