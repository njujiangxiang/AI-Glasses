# PDM 与 Go 模型对比分析报告

---

## 1. 摘要

- PDM 中表数量: 25
- Go 模型数量: 28
- 共同表数量: 0
- PDM 独有表数量: 25
- Go 独有表数量: 28

## 2. 表对比

### 2.1 PDM 中有但代码中没有的表

| 表名 | 注释 |
|------|------|
| algorithm_config | algorithm_config |
| attachments | 附件表：保存眼镜端照片、音频等证据文件的元数据和绑定关系 |
| audit_logs | 审计日志表：记录后台关键业务操作，用于追踪和审计 |
| defects | 缺陷表：保存巡检过程中发现并上报的异常缺陷 |
| device_audit_logs | 设备审计日志表：记录设备绑定、撤销、丢失禁用等操作 |
| device_sessions | 设备会话表：保存眼镜端登录和刷新令牌状态 |
| devices | 设备表：保存智能眼镜设备登记、绑定和状态信息 |
| inspection_task_nodes | 巡检任务节点表：保存具体任务下每个巡检步骤的执行要求和状态 |
| inspection_tasks | 巡检任务表：保存由计划生成并由眼镜端执行的具体任务 |
| inspection_template_nodes | 巡检模板节点表：定义模板下的巡检步骤、顺序和提交要求 |
| inspection_templates | 巡检模板表：定义可生成巡检任务的模板主信息 |
| organizations | 组织表：保存单位编码、单位名称和上下级单位关系，用于系统管理和用户归属 |
| outbox_events | Outbox事件表：保存待发布的异步业务事件，用于任务分配等消息集成 |
| permissions | 权限表：定义系统接口和功能权限编码 |
| realtime_query_config | realtime_query_config |
| role_permissions | 角色权限关联表：维护角色和权限的多对多关系 |
| roles | 角色表：定义系统内可分配给用户的角色 |
| task_node_results | 任务节点结果表：保存眼镜端提交的节点执行结果 |
| task_plans | 任务计划表：按模板和周期规则生成具体巡检任务 |
| task_type_dict | task_type_dict |
| team_members | 班组成员表：维护班组和用户的成员关系 |
| teams | 班组表：保存巡检任务可分配的班组 |
| template_nodes_config | 巡检模板节点-配置明细表 |
| user_roles | 用户角色关联表：维护用户和角色的多对多关系 |
| users | 用户表：保存后台管理员、任务人员和眼镜端巡检员的基础账号信息 |

### 2.2 代码中有但 PDM 中没有的表

| 表名 | Struct 名 |
|------|-----------|
| algorithmconfig | AlgorithmConfig |
| attachment | Attachment |
| auditlog | AuditLog |
| businesscode | BusinessCode |
| defect | Defect |
| device | Device |
| deviceauditlog | DeviceAuditLog |
| devicesession | DeviceSession |
| inspectiontask | InspectionTask |
| inspectiontasknode | InspectionTaskNode |
| inspectiontemplate | InspectionTemplate |
| inspectiontemplatenode | InspectionTemplateNode |
| organization | Organization |
| outboxevent | OutboxEvent |
| permission | Permission |
| realtimequeryconfig | RealtimeQueryConfig |
| role | Role |
| rolepermission | RolePermission |
| tasknoderesult | TaskNodeResult |
| taskplan | TaskPlan |
| tasktypedict | TaskTypeDict |
| team | Team |
| teammember | TeamMember |
| templatenodeconfig | TemplateNodeConfig |
| user | User |
| userrole | UserRole |
| workflow | Workflow |
| workflowstep | WorkflowStep |

## 3. 字段对比（共同表）

## 4. 更新建议

### 4.1 需要同步到代码的 PDM 表

- **algorithm_config** (algorithm_config)
  ```go
  type algorithm_config struct {
      CallMethod *string `gorm:"size:20"`
      CreatedAt *time.Time
      Id int64 `gorm:"primaryKey;not null"`
      InputParams *string `gorm:"size:512"`
      IsEnable *int8 `gorm:"size:1"`
      Name *string `gorm:"size:100"`
      OutputParams *string `gorm:"size:512"`
      Remark *string `gorm:"size:255"`
      ServiceUrl *string `gorm:"size:255"`
      Type *string `gorm:"size:50"`
      UpdatedAt *time.Time
      Version *string `gorm:"size:20"`
  }
  ```

- **attachments** (附件表：保存眼镜端照片、音频等证据文件的元数据和绑定关系)
  ```go
  type attachments struct {
      BindStatus string `gorm:"size:32;not null"`
      CaptureTime *time.Time
      ContentType string `gorm:"size:128;not null"`
      CreatedAt *time.Time
      DeviceId *int64
      FileName *string `gorm:"size:255"`
      GpsLat *float64
      GpsLng *float64
      Id int64 `gorm:"primaryKey;not null;comment:附件ID，系统内部主键"`
      NodeId *int64
      ObjectKey string `gorm:"size:255;not null"`
      ResultId *int64
      Sha256 *string `gorm:"size:64"`
      SizeBytes int64 `gorm:"not null"`
      TaskId *int64
      UpdatedAt *time.Time
      UploadTime *time.Time
      UserId int64 `gorm:"not null"`
  }
  ```

- **audit_logs** (审计日志表：记录后台关键业务操作，用于追踪和审计)
  ```go
  type audit_logs struct {
      Action string `gorm:"size:128;not null"`
      ActorId int64 `gorm:"not null"`
      CreatedAt *time.Time
      Detail *string
      Id int64 `gorm:"primaryKey;not null;comment:审计日志ID，系统内部主键"`
      Target string `gorm:"size:128;not null"`
      TargetId int64 `gorm:"not null"`
  }
  ```

- **defects** (缺陷表：保存巡检过程中发现并上报的异常缺陷)
  ```go
  type defects struct {
      CloseReason *string
      ClosedAt *time.Time
      ConfirmedAt *time.Time
      CreatedAt *time.Time
      Description *string
      Id int64 `gorm:"primaryKey;not null;comment:缺陷ID，系统内部主键"`
      InsId *int64 `gorm:"comment:巡检任务ID，系统内部主键"`
      NodeId int64 `gorm:"not null"`
      ReporterId int64 `gorm:"not null"`
      Status string `gorm:"size:32;not null"`
      TaskId int64 `gorm:"not null"`
      UpdatedAt *time.Time
  }
  ```

- **device_audit_logs** (设备审计日志表：记录设备绑定、撤销、丢失禁用等操作)
  ```go
  type device_audit_logs struct {
      Action string `gorm:"size:64;not null"`
      ActorId int64 `gorm:"not null"`
      CreatedAt *time.Time
      DeviceId int64 `gorm:"not null"`
      Id int64 `gorm:"primaryKey;not null;comment:设备审计日志ID，系统内部主键"`
      Reason *string `gorm:"size:255"`
  }
  ```

- **device_sessions** (设备会话表：保存眼镜端登录和刷新令牌状态)
  ```go
  type device_sessions struct {
      CreatedAt *time.Time
      DeviceId int64 `gorm:"not null"`
      Id int64 `gorm:"primaryKey;not null;comment:设备会话ID，系统内部主键"`
      RefreshJti string `gorm:"size:64;not null"`
      RefreshUntil *time.Time
      Status string `gorm:"size:32;not null"`
      UpdatedAt *time.Time
      UserId int64 `gorm:"not null"`
  }
  ```

- **devices** (设备表：保存智能眼镜设备登记、绑定和状态信息)
  ```go
  type devices struct {
      BoundAt *time.Time
      BoundUserId *int64
      CreatedAt *time.Time
      Id int64 `gorm:"primaryKey;not null;comment:设备ID，系统内部主键"`
      Name *string `gorm:"size:128"`
      SerialNo string `gorm:"size:128;not null"`
      Status string `gorm:"size:32;not null"`
      UpdatedAt *time.Time
  }
  ```

- **inspection_task_nodes** (巡检任务节点表：保存具体任务下每个巡检步骤的执行要求和状态)
  ```go
  type inspection_task_nodes struct {
      ActualExecTime *time.Time
      AlgorithmId *string `gorm:"size:32"`
      AllowAbnormal int8 `gorm:"size:1;not null"`
      CreatedAt *time.Time
      Id int64 `gorm:"primaryKey;not null;comment:任务节点ID，系统内部主键"`
      InsId *int64 `gorm:"comment:模板节点ID，系统内部主键"`
      IsMandatory *int8 `gorm:"size:1"`
      IsRequired *int8 `gorm:"size:1"`
      MinPhotos int64 `gorm:"not null"`
      Name string `gorm:"size:128;not null"`
      NodeType string `gorm:"size:32;not null"`
      QueryId *string `gorm:"size:32"`
      Remark *string `gorm:"size:128"`
      RequireText int8 `gorm:"size:1;not null"`
      SortOrder int64 `gorm:"not null"`
      Status string `gorm:"size:32;not null"`
      TaskId int64 `gorm:"not null"`
      TaskTypeCode *string `gorm:"size:30"`
      TemplateNodeId int64 `gorm:"not null"`
      UpdatedAt *time.Time
  }
  ```

- **inspection_tasks** (巡检任务表：保存由计划生成并由眼镜端执行的具体任务)
  ```go
  type inspection_tasks struct {
      AssignTime *time.Time
      AssignUser *string `gorm:"size:50"`
      AssigneeId int64 `gorm:"not null"`
      AssigneeType string `gorm:"size:16;not null"`
      CancelledAt *time.Time
      CompletedAt *time.Time
      CreatedAt *time.Time
      DueAt *time.Time
      EquipmentName *string `gorm:"size:128"`
      ExecutorId *int64
      GlassesSn *string `gorm:"size:50"`
      Id int64 `gorm:"primaryKey;not null;comment:巡检任务ID，系统内部主键"`
      InspectArea *string `gorm:"size:255"`
      PlanId int64 `gorm:"not null"`
      PointName *string `gorm:"size:128"`
      ScheduledAt *time.Time
      StartedAt *time.Time
      Status string `gorm:"size:32;not null"`
      SubmittedAt *time.Time
      TaskName *string `gorm:"size:255"`
      TemplateId int64 `gorm:"not null"`
      UpdatedAt *time.Time
  }
  ```

- **inspection_template_nodes** (巡检模板节点表：定义模板下的巡检步骤、顺序和提交要求)
  ```go
  type inspection_template_nodes struct {
      AlgorithmId *string `gorm:"size:32"`
      AllowAbnormal int8 `gorm:"size:1;not null"`
      CreatedAt *time.Time
      Description *string `gorm:"size:512"`
      Id int64 `gorm:"primaryKey;not null;comment:模板节点ID，系统内部主键"`
      IsMandatory int8 `gorm:"size:1;not null"`
      IsRequired int8 `gorm:"size:1;not null"`
      MinPhotos int64 `gorm:"not null"`
      Name string `gorm:"size:128;not null"`
      NodeDesc string `gorm:"size:512;not null"`
      NodeType string `gorm:"size:32;not null"`
      QueryId *string `gorm:"size:32"`
      Remark *string `gorm:"size:128"`
      RequireLiveCapture int8 `gorm:"size:1;not null"`
      RequireText int8 `gorm:"size:1;not null"`
      SortOrder int64 `gorm:"not null"`
      TaskTypeId *string `gorm:"size:32"`
      TemplateId int64 `gorm:"not null"`
      TimeoutSecond *int64
      UpdatedAt *time.Time
  }
  ```

- **inspection_templates** (巡检模板表：定义可生成巡检任务的模板主信息)
  ```go
  type inspection_templates struct {
      ApplicableRoles *string `gorm:"size:255"`
      CreatedAt *time.Time
      Creator string `gorm:"size:50;not null"`
      Description string `gorm:"size:512;not null"`
      Enabled int8 `gorm:"size:1;not null"`
      Id int64 `gorm:"primaryKey;not null;comment:巡检模板ID，系统内部主键"`
      IsEnable int8 `gorm:"size:1;not null"`
      Name string `gorm:"size:128;not null"`
      Remark *string `gorm:"size:255"`
      Scene string `gorm:"size:100;not null"`
      Type string `gorm:"size:50;not null"`
      UpdatedAt *time.Time
      Version string `gorm:"size:20;not null"`
  }
  ```

- **organizations** (组织表：保存单位编码、单位名称和上下级单位关系，用于系统管理和用户归属)
  ```go
  type organizations struct {
      Code string `gorm:"size:64;not null"`
      CreatedAt *time.Time
      Id int64 `gorm:"primaryKey;not null;comment:组织ID，系统内部主键"`
      Name string `gorm:"size:128;not null"`
      ParentCode string `gorm:"size:64;not null"`
      Status string `gorm:"size:32;not null"`
      UpdatedAt *time.Time
  }
  ```

- **outbox_events** (Outbox事件表：保存待发布的异步业务事件，用于任务分配等消息集成)
  ```go
  type outbox_events struct {
      CreatedAt *time.Time
      EventKey string `gorm:"size:160;not null"`
      Id int64 `gorm:"primaryKey;not null;comment:事件ID，系统内部主键"`
      Payload string `gorm:"not null"`
      PublishedAt *time.Time
      Topic string `gorm:"size:128;not null"`
      UpdatedAt *time.Time
  }
  ```

- **permissions** (权限表：定义系统接口和功能权限编码)
  ```go
  type permissions struct {
      Code string `gorm:"size:128;not null"`
      CreatedAt *time.Time
      Id int64 `gorm:"primaryKey;not null;comment:权限ID，系统内部主键"`
      PermissionId *int64 `gorm:"comment:权限ID，关联permissions.id"`
      RoleId *int64 `gorm:"comment:角色ID，关联roles.id"`
      UpdatedAt *time.Time
  }
  ```

- **realtime_query_config** (realtime_query_config)
  ```go
  type realtime_query_config struct {
      ApiUrl *string `gorm:"size:255"`
      ApplyScene *string `gorm:"size:100"`
      AuthType *string `gorm:"size:30"`
      CreatedAt *time.Time
      Id int64 `gorm:"primaryKey;not null"`
      Name *string `gorm:"size:100"`
      Remark *string `gorm:"size:255"`
      RequestMethod *string `gorm:"size:10"`
      RequestParams *string `gorm:"size:255"`
      ResponseParams *string `gorm:"size:255"`
      TimeoutSecond *int64
      UpdatedAt *time.Time
  }
  ```

- **role_permissions** (角色权限关联表：维护角色和权限的多对多关系)
  ```go
  type role_permissions struct {
      PermissionId int64 `gorm:"primaryKey;not null;comment:权限ID，关联permissions.id"`
      RoleId int64 `gorm:"primaryKey;not null;comment:角色ID，关联roles.id"`
  }
  ```

- **roles** (角色表：定义系统内可分配给用户的角色)
  ```go
  type roles struct {
      CreatedAt *time.Time
      Id int64 `gorm:"primaryKey;not null;comment:角色ID，系统内部主键"`
      Name string `gorm:"size:64;not null"`
      PermissionId *int64 `gorm:"comment:权限ID，关联permissions.id"`
      RoleId *int64 `gorm:"comment:角色ID，关联roles.id"`
      UpdatedAt *time.Time
      UseRoleId *int64 `gorm:"comment:角色ID，关联roles.id"`
      UserId *int64 `gorm:"comment:用户ID，关联users.id"`
  }
  ```

- **task_node_results** (任务节点结果表：保存眼镜端提交的节点执行结果)
  ```go
  type task_node_results struct {
      AbnormalDesc *string `gorm:"size:128"`
      AlgorithmResult *string `gorm:"size:512"`
      AttachmentIds *string `gorm:"size:256"`
      CompletedAt *time.Time
      CreatedAt *time.Time
      FeedbackContent *string `gorm:"size:512"`
      Id int64 `gorm:"primaryKey;not null;comment:节点结果ID，系统内部主键"`
      IdempotencyKey string `gorm:"size:128;not null"`
      InsId *int64 `gorm:"comment:任务节点ID，系统内部主键"`
      IsAbnormal *int8 `gorm:"size:1"`
      LocationGps *string `gorm:"size:50"`
      NodeId int64 `gorm:"not null"`
      QueryResult *string `gorm:"size:512"`
      Remark *string `gorm:"size:128"`
      Status string `gorm:"size:32;not null"`
      TaskId int64 `gorm:"not null"`
      TaskTypeCode *string `gorm:"size:32"`
      TextNote *string
      UpdatedAt *time.Time
      UserId int64 `gorm:"not null"`
  }
  ```

- **task_plans** (任务计划表：按模板和周期规则生成具体巡检任务)
  ```go
  type task_plans struct {
      AssigneeId int64 `gorm:"not null"`
      AssigneeType string `gorm:"size:16;not null"`
      BelongUnit *string `gorm:"size:100"`
      CreatedAt *time.Time
      Creator *string `gorm:"size:50"`
      CronExpr string `gorm:"size:64;not null"`
      DueDurationMinutes int64 `gorm:"not null"`
      Enabled int8 `gorm:"size:1;not null"`
      EquipmentName *string `gorm:"size:128"`
      Guardian *string `gorm:"size:50"`
      Id int64 `gorm:"primaryKey;not null;comment:任务计划ID，系统内部主键"`
      InspectArea *string `gorm:"size:256"`
      Name string `gorm:"size:128;not null"`
      OperatorUnit *string `gorm:"size:100"`
      OperatorUser *string `gorm:"size:50"`
      PlanDesc *string `gorm:"size:256"`
      PlanEndTime *time.Time
      PlanPrincipal *string `gorm:"size:50"`
      PlanStartTime *time.Time
      PlanStatus *string `gorm:"size:20"`
      PlanType *string `gorm:"size:50"`
      PointName *string `gorm:"size:128"`
      StartAt *time.Time
      SubstationName *string `gorm:"size:100"`
      TemplateId int64 `gorm:"not null"`
      Timezone string `gorm:"size:64;not null"`
      UpdatedAt *time.Time
  }
  ```

- **task_type_dict** (task_type_dict)
  ```go
  type task_type_dict struct {
      CreatedAt *time.Time
      Id int64 `gorm:"primaryKey;not null"`
      Remark *string `gorm:"size:256"`
      SupportAlgorithm *int8 `gorm:"size:1"`
      SupportMandatory *int8 `gorm:"size:1"`
      SupportQuery *int8 `gorm:"size:1"`
      TypeCode *string `gorm:"size:30"`
      TypeDesc *string `gorm:"size:255"`
      TypeName *string `gorm:"size:50"`
      UpdatedAt *time.Time
  }
  ```

- **team_members** (班组成员表：维护班组和用户的成员关系)
  ```go
  type team_members struct {
      TeamId int64 `gorm:"primaryKey;not null;comment:班组ID，关联teams.id"`
      UserId int64 `gorm:"primaryKey;not null;comment:用户ID，关联users.id"`
  }
  ```

- **teams** (班组表：保存巡检任务可分配的班组)
  ```go
  type teams struct {
      CreatedAt *time.Time
      Id int64 `gorm:"primaryKey;not null;comment:班组ID，系统内部主键"`
      Name string `gorm:"size:128;not null"`
      TeamId *int64 `gorm:"comment:班组ID，关联teams.id"`
      UpdatedAt *time.Time
      UserId *int64 `gorm:"comment:用户ID，关联users.id"`
  }
  ```

- **template_nodes_config** (巡检模板节点-配置明细表)
  ```go
  type template_nodes_config struct {
      ConfigCode string `gorm:"size:50;not null;comment:配置项编码"`
      ConfigName string `gorm:"size:100;not null;comment:配置项名称"`
      ConfigType string `gorm:"size:50;not null;comment:配置项类型"`
      ConfigValue string `gorm:"size:50;not null;comment:配置项值"`
      CreatedAt *time.Time
      Id int64 `gorm:"primaryKey;not null;comment:配置项唯一编号，系统内部主键"`
      IsDefault *int8 `gorm:"size:1;comment:是否默认值"`
      NodeId int64 `gorm:"not null;comment:所属节点ID"`
      Remark *string `gorm:"size:255;comment:配置说明"`
      Sort int `gorm:"not null;comment:排序号"`
      UpdatedAt *time.Time
  }
  ```

- **user_roles** (用户角色关联表：维护用户和角色的多对多关系)
  ```go
  type user_roles struct {
      RoleId int64 `gorm:"primaryKey;not null;comment:角色ID，关联roles.id"`
      UserId int64 `gorm:"primaryKey;not null;comment:用户ID，关联users.id"`
  }
  ```

- **users** (用户表：保存后台管理员、任务人员和眼镜端巡检员的基础账号信息)
  ```go
  type users struct {
      AudId *int64 `gorm:"comment:审计日志ID，系统内部主键"`
      AvatarContentType string `gorm:"size:64;not null"`
      AvatarData []byte
      AvatarSize int64 `gorm:"not null"`
      BirthMonth int64 `gorm:"not null"`
      BirthYear int64 `gorm:"not null"`
      CreatedAt *time.Time
      DisplayName *string `gorm:"size:128"`
      Email *string `gorm:"size:128"`
      Gender string `gorm:"size:8;not null"`
      Id int64 `gorm:"primaryKey;not null;comment:用户ID，系统内部主键"`
      IdCardNo string `gorm:"size:32;not null"`
      IsDeleted *int8 `gorm:"size:1"`
      LastLoginTime *time.Time
      Name string `gorm:"size:64;not null"`
      OrgCode string `gorm:"size:64;not null"`
      PasswordHash string `gorm:"size:255;not null"`
      Phone *string `gorm:"size:11"`
      PwdExpireTime *time.Time
      Role *int64
      RoleId *int64 `gorm:"comment:角色ID，关联roles.id"`
      Status string `gorm:"size:32;not null"`
      TeamId *int64 `gorm:"comment:班组ID，关联teams.id"`
      UpdatedAt *time.Time
      UseUserId *int64 `gorm:"comment:用户ID，关联users.id"`
      UserId *int64 `gorm:"comment:用户ID，关联users.id"`
      UserLock *int8 `gorm:"size:1"`
      UserStatus *int8 `gorm:"size:1"`
      UserType *int8 `gorm:"size:1"`
      Username string `gorm:"size:64;not null"`
  }
  ```

### 4.2 需要同步到 PDM 的 Go 模型表

- **algorithmconfig** (AlgorithmConfig)
- **attachment** (Attachment)
- **auditlog** (AuditLog)
- **businesscode** (BusinessCode)
- **defect** (Defect)
- **device** (Device)
- **deviceauditlog** (DeviceAuditLog)
- **devicesession** (DeviceSession)
- **inspectiontask** (InspectionTask)
- **inspectiontasknode** (InspectionTaskNode)
- **inspectiontemplate** (InspectionTemplate)
- **inspectiontemplatenode** (InspectionTemplateNode)
- **organization** (Organization)
- **outboxevent** (OutboxEvent)
- **permission** (Permission)
- **realtimequeryconfig** (RealtimeQueryConfig)
- **role** (Role)
- **rolepermission** (RolePermission)
- **tasknoderesult** (TaskNodeResult)
- **taskplan** (TaskPlan)
- **tasktypedict** (TaskTypeDict)
- **team** (Team)
- **teammember** (TeamMember)
- **templatenodeconfig** (TemplateNodeConfig)
- **user** (User)
- **userrole** (UserRole)
- **workflow** (Workflow)
- **workflowstep** (WorkflowStep)

## 5. 完整清单

### 5.1 PDM 表完整清单

- algorithm_config: algorithm_config (12 列)
- attachments: 附件表：保存眼镜端照片、音频等证据文件的元数据和绑定关系 (18 列)
- audit_logs: 审计日志表：记录后台关键业务操作，用于追踪和审计 (7 列)
- defects: 缺陷表：保存巡检过程中发现并上报的异常缺陷 (12 列)
- device_audit_logs: 设备审计日志表：记录设备绑定、撤销、丢失禁用等操作 (6 列)
- device_sessions: 设备会话表：保存眼镜端登录和刷新令牌状态 (8 列)
- devices: 设备表：保存智能眼镜设备登记、绑定和状态信息 (8 列)
- inspection_task_nodes: 巡检任务节点表：保存具体任务下每个巡检步骤的执行要求和状态 (20 列)
- inspection_tasks: 巡检任务表：保存由计划生成并由眼镜端执行的具体任务 (22 列)
- inspection_template_nodes: 巡检模板节点表：定义模板下的巡检步骤、顺序和提交要求 (20 列)
- inspection_templates: 巡检模板表：定义可生成巡检任务的模板主信息 (13 列)
- organizations: 组织表：保存单位编码、单位名称和上下级单位关系，用于系统管理和用户归属 (7 列)
- outbox_events: Outbox事件表：保存待发布的异步业务事件，用于任务分配等消息集成 (7 列)
- permissions: 权限表：定义系统接口和功能权限编码 (6 列)
- realtime_query_config: realtime_query_config (12 列)
- role_permissions: 角色权限关联表：维护角色和权限的多对多关系 (2 列)
- roles: 角色表：定义系统内可分配给用户的角色 (8 列)
- task_node_results: 任务节点结果表：保存眼镜端提交的节点执行结果 (20 列)
- task_plans: 任务计划表：按模板和周期规则生成具体巡检任务 (27 列)
- task_type_dict: task_type_dict (10 列)
- team_members: 班组成员表：维护班组和用户的成员关系 (2 列)
- teams: 班组表：保存巡检任务可分配的班组 (6 列)
- template_nodes_config: 巡检模板节点-配置明细表 (11 列)
- user_roles: 用户角色关联表：维护用户和角色的多对多关系 (2 列)
- users: 用户表：保存后台管理员、任务人员和眼镜端巡检员的基础账号信息 (30 列)

### 5.2 Go 模型完整清单

- algorithmconfig: AlgorithmConfig (10 字段)
- attachment: Attachment (14 字段)
- auditlog: AuditLog (6 字段)
- businesscode: BusinessCode (8 字段)
- defect: Defect (7 字段)
- device: Device (5 字段)
- deviceauditlog: DeviceAuditLog (5 字段)
- devicesession: DeviceSession (5 字段)
- inspectiontask: InspectionTask (13 字段)
- inspectiontasknode: InspectionTaskNode (16 字段)
- inspectiontemplate: InspectionTemplate (11 字段)
- inspectiontemplatenode: InspectionTemplateNode (17 字段)
- organization: Organization (5 字段)
- outboxevent: OutboxEvent (4 字段)
- permission: Permission (2 字段)
- realtimequeryconfig: RealtimeQueryConfig (10 字段)
- role: Role (2 字段)
- rolepermission: RolePermission (2 字段)
- tasknoderesult: TaskNodeResult (16 字段)
- taskplan: TaskPlan (22 字段)
- tasktypedict: TaskTypeDict (8 字段)
- team: Team (2 字段)
- teammember: TeamMember (2 字段)
- templatenodeconfig: TemplateNodeConfig (9 字段)
- user: User (21 字段)
- userrole: UserRole (2 字段)
- workflow: Workflow (5 字段)
- workflowstep: WorkflowStep (13 字段)
