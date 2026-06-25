
// Package database 定义 MVP 使用的持久化数据模型。模型覆盖用户、角色、设备生命周期、
// 巡检模板、生成任务、证据附件、缺陷、审计日志以及调度/事件流水线使用的 outbox 事件。
package database

import "time"

type User struct {
	ID                uint64     `gorm:"primaryKey;comment:'用户ID，系统内部主键'" json:"id"`
	Username          string     `gorm:"size:64;uniqueIndex;not null;comment:'用户名，用于后台或眼镜端登录，系统内唯一'" json:"username"`
	PasswordHash      string     `gorm:"size:255;not null;comment:'密码哈希；当前MVP为开发态占位值'" json:"-"`
	DisplayName       string     `gorm:"size:128;not null;default:'';comment:'显示名称，用于页面展示'" json:"display_name"`
	Name              string     `gorm:"size:64;not null;default:'';comment:'用户姓名'" json:"name"`
	Gender            string     `gorm:"size:8;not null;default:'';comment:'性别'" json:"gender"`
	AvatarData        []byte     `gorm:"type:longblob;comment:'头像二进制数据'" json:"-"`
	AvatarContentType string     `gorm:"size:64;not null;default:'';comment:'头像内容类型'" json:"avatar_content_type"`
	AvatarSize        int64      `gorm:"not null;default:0;comment:'头像大小（字节）'" json:"avatar_size"`
	BirthYear         int        `gorm:"not null;default:0;comment:'出生年份'" json:"birth_year"`
	BirthMonth        int        `gorm:"not null;default:0;comment:'出生月份'" json:"birth_month"`
	IDCardNo          string     `gorm:"size:32;index;not null;default:'';comment:'身份证号'" json:"id_card_no"`
	OrgCode           string     `gorm:"size:64;index;not null;default:'';comment:'所属组织编码'" json:"org_code"`
	Status            string     `gorm:"size:32;index;not null;comment:'用户状态：active启用，disabled停用'" json:"status"`
	Phone             string     `gorm:"size:11;comment:'手机号'" json:"phone"`
	Email             string     `gorm:"size:128;comment:'邮箱'" json:"email"`
	IsDeleted         bool       `gorm:"type:tinyint(1);default:0;comment:'是否已删除'" json:"is_deleted"`
	Role              int        `gorm:"comment:'角色标识（兼容旧版）'" json:"role"`
	UserType          int        `gorm:"type:tinyint(1);comment:'用户类型：1长期，2临时'" json:"user_type"`
	UserStatus        int        `gorm:"type:tinyint(1);comment:'用户状态：1正常'" json:"user_status"`
	LastLoginTime     *time.Time `gorm:"comment:'最后登录时间'" json:"last_login_time"`
	UserLock          bool       `gorm:"type:tinyint(1);comment:'账号是否锁定'" json:"user_lock"`
	PwdExpireTime     *time.Time `gorm:"comment:'密码过期时间'" json:"pwd_expire_time"`
	CreatedAt         time.Time  `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"comment:'更新时间'" json:"updated_at"`
}

type Organization struct {
	ID         uint64    `gorm:"primaryKey;comment:'组织ID，系统内部主键'" json:"id"`
	Code       string    `gorm:"size:64;uniqueIndex;not null;comment:'组织编码，唯一'" json:"code"`
	Name       string    `gorm:"size:128;not null;comment:'组织名称'" json:"name"`
	ParentCode string    `gorm:"size:64;index;not null;default:'';comment:'上级组织编码'" json:"parent_code"`
	Status     string    `gorm:"size:32;index;not null;comment:'组织状态：active启用，disabled停用'" json:"status"`
	CreatedAt  time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt  time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
}

type BusinessCode struct {
	ID           uint64    `gorm:"primaryKey;comment:'业务编码配置ID，系统内部主键'" json:"id"`
	Name         string    `gorm:"size:128;not null;comment:'编码名称，用于后台展示规则用途'" json:"name"`
	Code         string    `gorm:"size:64;uniqueIndex;not null;comment:'业务代码，系统内唯一，例如TK'" json:"code"`
	DateFormat   string    `gorm:"size:32;not null;comment:'日期格式，首版仅支持yyyyMMdd'" json:"date_format"`
	SeqPadding   int       `gorm:"not null;comment:'流水号位数，例如4表示0001'" json:"seq_padding"`
	Separator    string    `gorm:"size:8;not null;default:'';comment:'分隔符，例如-，不使用时为空'" json:"separator"`
	UseSeparator bool      `gorm:"not null;default:false;comment:'是否在代码、日期、流水号之间使用分隔符'" json:"use_separator"`
	UseDate      *bool     `gorm:"not null;default:true;comment:'是否按日生成流水号'" json:"use_date"`
	UseOrgCode   bool      `gorm:"not null;default:false;comment:'是否使用组织编码'" json:"use_org_code"`
	OrgSource    string    `gorm:"size:32;not null;default:'fixed';comment:'组织编码来源：fixed/current'" json:"org_source"`
	OrgCode      string    `gorm:"size:64;not null;default:'';comment:'组织机构编码'" json:"org_code"`
	Status       string    `gorm:"size:32;not null;default:'active';comment:'编码状态：active启用，disabled停用'" json:"status"`
	CreatedAt    time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt    time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
}

type Role struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Status    string    `gorm:"size:32;not null;default:'active'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Permission struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"size:128;uniqueIndex;not null" json:"code"`
	Perms     string    `gorm:"size:128;not null;default:''" json:"perms"`
	Status    string    `gorm:"size:32;not null;default:'active'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRole struct {
	UserID uint64 `gorm:"primaryKey" json:"user_id"`
	RoleID uint64 `gorm:"primaryKey" json:"role_id"`
}

type RolePermission struct {
	RoleID       uint64 `gorm:"primaryKey" json:"role_id"`
	PermissionID uint64 `gorm:"primaryKey" json:"permission_id"`
}

type Team struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TeamMember struct {
	TeamID uint64 `gorm:"primaryKey" json:"team_id"`
	UserID uint64 `gorm:"primaryKey" json:"user_id"`
}

type Device struct {
	ID          uint64     `gorm:"primaryKey;comment:'设备ID，系统内部主键'" json:"id"`
	SerialNo    string     `gorm:"size:128;uniqueIndex;not null;comment:'智能眼镜设备序列号，系统内唯一'" json:"serial_no"`
	Name        string     `gorm:"size:128;not null;default:'';comment:'设备名称或备注名称'" json:"name"`
	OrgCode     string     `gorm:"size:64;index;not null;default:'';comment:'所属组织编码'" json:"org_code"`
	Status      string     `gorm:"size:32;index:idx_devices_status_bound_user_id,priority:1;not null;comment:'设备状态：pending待绑定，active启用，revoked撤销，lost_disabled丢失禁用'" json:"status"`
	BoundUserID *uint64    `gorm:"index:idx_devices_status_bound_user_id,priority:2;comment:'当前绑定用户ID，关联users.id'" json:"bound_user_id"`
	BoundAt     *time.Time `gorm:"comment:'绑定时间'" json:"bound_at"`
	CreatedAt   time.Time  `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"comment:'更新时间'" json:"updated_at"`
}

type DeviceSession struct {
	ID           uint64    `gorm:"primaryKey;comment:'设备会话ID，系统内部主键'" json:"id"`
	DeviceID     uint64    `gorm:"index;not null;comment:'设备ID，关联devices.id'" json:"device_id"`
	UserID       uint64    `gorm:"index;not null;comment:'用户ID，关联users.id'" json:"user_id"`
	RefreshJTI   string    `gorm:"size:64;uniqueIndex;not null;comment:'刷新令牌唯一标识，用于设备会话续期和撤销'" json:"refresh_jti"`
	Status       string    `gorm:"size:32;index;not null;comment:'会话状态：active有效，revoked已撤销'" json:"status"`
	RefreshUntil time.Time `gorm:"comment:'刷新令牌有效截止时间'" json:"refresh_until"`
	CreatedAt    time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt    time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
}

type DeviceAuditLog struct {
	ID        uint64    `gorm:"primaryKey;comment:'设备审计日志ID，系统内部主键'" json:"id"`
	DeviceID  uint64    `gorm:"index;not null;comment:'设备ID，关联devices.id'" json:"device_id"`
	ActorID   uint64    `gorm:"index;not null;comment:'操作人用户ID，关联users.id'" json:"actor_id"`
	Action    string    `gorm:"size:64;not null;comment:'操作动作，例如bind、revoke、disable_lost'" json:"action"`
	Reason    string    `gorm:"size:255;not null;default:'';comment:'操作原因或备注'" json:"reason"`
	CreatedAt time.Time `gorm:"comment:'创建时间'" json:"created_at"`
}

// ========== 新增配置表 ==========

type TaskTypeDict struct {
	ID               uint64    `gorm:"primaryKey;comment:'任务类型字典ID，系统内部主键'" json:"id"`
	TypeCode         string    `gorm:"size:30;comment:'类型编码：text/read/check/photo/video/audio'" json:"type_code"`
	TypeName         string    `gorm:"size:50;comment:'类型名称'" json:"type_name"`
	TypeDesc         string    `gorm:"size:255;comment:'类型描述'" json:"type_desc"`
	SupportAlgorithm string    `gorm:"size:1;comment:'是否支持算法：1是，0否'" json:"support_algorithm"`
	SupportQuery     string    `gorm:"size:1;comment:'是否支持实时查询：1是，0否'" json:"support_query"`
	SupportMandatory string    `gorm:"size:1;comment:'是否支持强制配置：1是，0否'" json:"support_mandatory"`
	CreatedAt        time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt        time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
	Remark           string    `gorm:"size:256;comment:'备注'" json:"remark"`
}

type AlgorithmConfig struct {
	ID            uint64    `gorm:"primaryKey;comment:'算法配置ID，系统内部主键'" json:"id"`
	Name          string    `gorm:"size:100;comment:'算法名称'" json:"name"`
	Type          string    `gorm:"size:50;comment:'算法类型'" json:"type"`
	ServiceURL    string    `gorm:"size:255;comment:'算法服务URL'" json:"service_url"`
	CallMethod    string    `gorm:"size:20;comment:'调用方法：GET/POST'" json:"call_method"`
	InputParams   string    `gorm:"size:512;comment:'输入参数JSON'" json:"input_params"`
	OutputParams  string    `gorm:"size:512;comment:'输出参数JSON'" json:"output_params"`
	Version       string    `gorm:"size:20;comment:'版本号'" json:"version"`
	IsEnable      bool      `gorm:"type:tinyint(1);comment:'是否启用'" json:"is_enable"`
	CreatedAt     time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt     time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
	Remark        string    `gorm:"size:255;comment:'备注'" json:"remark"`
}

type RealtimeQueryConfig struct {
	ID             uint64    `gorm:"primaryKey;comment:'实时查询配置ID，系统内部主键'" json:"id"`
	Name           string    `gorm:"size:100;comment:'查询配置名称'" json:"name"`
	ApiURL         string    `gorm:"size:255;comment:'API接口URL'" json:"api_url"`
	RequestMethod  string    `gorm:"size:10;comment:'请求方法：GET/POST'" json:"request_method"`
	AuthType       string    `gorm:"size:30;comment:'认证方式'" json:"auth_type"`
	RequestParams  string    `gorm:"size:255;comment:'请求参数JSON'" json:"request_params"`
	ResponseParams string    `gorm:"size:255;comment:'响应参数JSON'" json:"response_params"`
	TimeoutSecond  int       `gorm:"comment:'超时时间（秒）'" json:"timeout_second"`
	ApplyScene     string    `gorm:"size:100;comment:'适用场景'" json:"apply_scene"`
	CreatedAt      time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt      time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
	Remark         string    `gorm:"size:255;comment:'备注'" json:"remark"`
}

// ========== 模板相关 ==========

type InspectionTemplate struct {
	ID              uint64    `gorm:"primaryKey;comment:'巡检模板ID，系统内部主键'" json:"id"`
	Name            string    `gorm:"size:128;not null;comment:'模板名称'" json:"name"`
	Description     string    `gorm:"size:512;not null;default:'';comment:'模板说明'" json:"description"`
	ApplicableRoles string    `gorm:"size:255;not null;default:'';comment:'适用角色，使用逗号或文本记录角色范围'" json:"applicable_roles"`
	Enabled         bool      `gorm:"type:tinyint(1);index;not null;default:0;comment:'是否启用模板'" json:"enabled"`
	Type            string    `gorm:"size:50;not null;default:'';comment:'业务类型：设备巡检/缺陷复查/安全交底/保电特巡'" json:"type"`
	Scene           string    `gorm:"size:100;not null;default:'';comment:'适用业务场景：变电巡视/配电巡检/输电线路'" json:"scene"`
	Version         string    `gorm:"size:20;not null;default:'v1';comment:'版本号'" json:"version"`
	IsEnable        bool      `gorm:"type:tinyint(1);not null;default:true;comment:'是否启用（兼容字段）'" json:"is_enable"`
	Creator         string    `gorm:"size:50;not null;default:'';comment:'创建人'" json:"creator"`
	Remark          string    `gorm:"size:255;comment:'备注'" json:"remark"`
	CreatedAt       time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt       time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
}

type InspectionTemplateNode struct {
	ID                 uint64    `gorm:"primaryKey;comment:'模板节点ID，系统内部主键'" json:"id"`
	TemplateID         uint64    `gorm:"uniqueIndex:idx_template_sort;index;not null;comment:'巡检模板ID，关联inspection_templates.id'" json:"template_id"`
	SortOrder          int       `gorm:"uniqueIndex:idx_template_sort;not null;default:0;comment:'节点排序号，数值越小越靠前'" json:"sort_order"`
	Name               string    `gorm:"size:128;not null;comment:'节点名称'" json:"name"`
	Description        string    `gorm:"size:512;not null;default:'';comment:'节点说明'" json:"description"`
	NodeDesc           string    `gorm:"size:512;not null;default:'';comment:'节点简短提示（AR眼镜端展示）'" json:"node_desc"`
	NodeType           string    `gorm:"size:32;not null;comment:'节点类型：text/read/check/photo/video/audio'" json:"node_type"`
	MinPhotos          int       `gorm:"not null;default:0;comment:'最少照片数量要求'" json:"min_photos"`
	RequireText        bool      `gorm:"type:tinyint(1);not null;default:0;comment:'是否要求填写文本说明'" json:"require_text"`
	AllowAbnormal      bool      `gorm:"type:tinyint(1);not null;default:0;comment:'是否允许在该节点上报异常'" json:"allow_abnormal"`
	RequireLiveCapture bool      `gorm:"type:tinyint(1);not null;default:1;comment:'是否要求现场实时拍摄'" json:"require_live_capture"`
	NodesConfigID      string    `gorm:"size:32;comment:'节点配置ID，关联template_nodes_config'" json:"nodes_config_id"`
	TaskTypeID         string    `gorm:"size:32;comment:'任务类型ID，关联task_type_dict'" json:"task_type_id"`
	IsMandatory        bool      `gorm:"not null;default:false;comment:'是否强制执行'" json:"is_mandatory"`
	IsRequired         bool      `gorm:"not null;default:false;comment:'是否必做节点'" json:"is_required"`
	AlgorithmID        string    `gorm:"size:32;comment:'绑定AI算法ID'" json:"algorithm_id"`
	QueryID            string    `gorm:"size:32;comment:'绑定实时查询接口ID'" json:"query_id"`
	TimeoutSecond      int       `gorm:"comment:'节点超时时间（秒）'" json:"timeout_second"`
	Remark             string    `gorm:"size:128;comment:'备注'" json:"remark"`
	CreatedAt          time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt          time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
}

type TemplateNodeConfig struct {
	ID          uint64    `gorm:"primaryKey;comment:'配置ID，系统内部主键'" json:"id"`
	NodeID      uint64    `gorm:"index;not null;comment:'模板节点ID，关联inspection_template_nodes.id'" json:"node_id"`
	ConfigCode  string    `gorm:"size:50;comment:'配置项编码，例如NORMAL/ABNORMAL/TEMP_MAX'" json:"config_code"`
	ConfigName  string    `gorm:"size:100;comment:'配置项名称'" json:"config_name"`
	ConfigValue string    `gorm:"size:256;comment:'配置值'" json:"config_value"`
	ConfigType  string    `gorm:"size:30;comment:'配置类型：枚举选项/数值阈值/业务参数'" json:"config_type"`
	Sort        int       `gorm:"comment:'排序号，默认0'" json:"sort"`
	IsDefault   bool      `gorm:"type:tinyint(1);comment:'是否默认值：1是，0否'" json:"is_default"`
	Remark      string    `gorm:"size:128;comment:'备注'" json:"remark"`
	CreatedAt   time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt   time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
}

// Workflow 工作流配置，定义巡检步骤和异常触发规则。
// Status: draft-草稿, published-已发布, archived-已归档
type Workflow struct {
	ID          uint64 `gorm:"primaryKey;comment:工作流ID" json:"id"`
	Name        string `gorm:"size:128;not null;comment:工作流名称" json:"name"`
	Description string `gorm:"size:512;comment:工作流描述" json:"description"`
	Status      string `gorm:"size:32;index;not null;default:'draft';comment:状态：draft草稿，published已发布，archived已归档" json:"status"`
	CreatedBy   uint64 `gorm:"index;not null;comment:创建人ID" json:"created_by"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// WorkflowStep 工作流步骤，支持多种输入类型和异常触发配置。
// Type: text-文本输入, number-数值输入, select-选择清单, photo-拍照, video-录像, audio-录音
type WorkflowStep struct {
	ID                    uint64 `gorm:"primaryKey;comment:步骤ID" json:"id"`
	WorkflowID            uint64 `gorm:"uniqueIndex:idx_workflow_sort;index;not null;comment:所属工作流ID" json:"workflow_id"`
	SortOrder             int `gorm:"uniqueIndex:idx_workflow_sort;not null;comment:排序序号" json:"sort_order"`
	Name                  string `gorm:"size:128;not null;comment:步骤名称" json:"name"`
	Description           string `gorm:"size:512;comment:步骤描述" json:"description"`
	Type                  string `gorm:"size:32;not null;comment:步骤类型：text,number,select,photo,video,audio" json:"type"`
	Required              bool `gorm:"not null;default:true;comment:是否必填" json:"required"`
	OptionsJSON           *string `gorm:"type:json;comment:选择项配置JSON，仅select类型使用" json:"options_json"`
	AbnormalEnabled       bool `gorm:"not null;default:false;comment:是否启用异常触发" json:"abnormal_enabled"`
	AbnormalRequirePhoto  bool `gorm:"not null;default:true;comment:异常时必须拍照" json:"abnormal_require_photo"`
	AbnormalRequireVideo  bool `gorm:"not null;default:false;comment:异常时必须录像" json:"abnormal_require_video"`
	AbnormalRequireNote   bool `gorm:"not null;default:true;comment:异常时必须填写备注" json:"abnormal_require_note"`
	AbnormalRequireSignature bool `gorm:"not null;default:false;comment:异常时必须签字确认" json:"abnormal_require_signature"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// ========== 任务计划 ==========

type TaskPlan struct {
	ID                 uint64     `gorm:"primaryKey;comment:'任务计划ID，系统内部主键'" json:"id"`
	TemplateID         uint64     `gorm:"index;not null;comment:'巡检模板ID，关联inspection_templates.id'" json:"template_id"`
	Name               string     `gorm:"size:128;not null;comment:'计划名称'" json:"name"`
	CronExpr           string     `gorm:"size:64;not null;comment:'Cron表达式，用于周期性生成巡检任务'" json:"cron_expr"`
	Timezone           string     `gorm:"size:64;not null;comment:'计划执行时区，例如Asia/Shanghai'" json:"timezone"`
	StartAt            time.Time  `gorm:"index;not null;comment:'计划开始生效时间'" json:"start_at"`
	DueDurationMinutes int        `gorm:"not null;comment:'任务生成后多少分钟内应完成'" json:"due_duration_minutes"`
	AssigneeType       string     `gorm:"size:16;not null;comment:'指派类型：user用户，team班组'" json:"assignee_type"`
	AssigneeID         uint64     `gorm:"index;not null;comment:'指派对象ID，根据assignee_type关联users.id或teams.id'" json:"assignee_id"`
	PointName          string     `gorm:"size:128;not null;default:'';comment:'巡检点位名称'" json:"point_name"`
	EquipmentName      string     `gorm:"size:128;not null;default:'';comment:'巡检设备名称'" json:"equipment_name"`
	Enabled            bool       `gorm:"index;not null;default:0;comment:'是否启用计划'" json:"enabled"`
	PlanType           string     `gorm:"size:50;comment:'计划类型：日常例行/专项防雷/缺陷复查/保电特巡'" json:"plan_type"`
	BelongUnit         string     `gorm:"size:100;comment:'计划所属单位'" json:"belong_unit"`
	OperatorUnit       string     `gorm:"size:100;comment:'作业人员所属单位'" json:"operator_unit"`
	SubstationName     string     `gorm:"size:100;comment:'作业地点（变电站）'" json:"substation_name"`
	InspectArea        string     `gorm:"size:256;comment:'作业区域'" json:"inspect_area"`
	PlanStartTime      *time.Time `gorm:"comment:'计划开始时间（手动模式）'" json:"plan_start_time"`
	PlanEndTime        *time.Time `gorm:"comment:'计划结束时间（手动模式）'" json:"plan_end_time"`
	PlanPrincipal      string     `gorm:"size:50;comment:'计划负责人'" json:"plan_principal"`
	OperatorUser       string     `gorm:"size:50;comment:'作业人员'" json:"operator_user"`
	Guardian           string     `gorm:"size:50;comment:'现场监护人'" json:"guardian"`
	PlanDesc           string     `gorm:"size:256;comment:'工作内容概述'" json:"plan_desc"`
	PlanStatus         string     `gorm:"size:20;comment:'计划状态：待下发/已下发/执行中/已完成/已取消'" json:"plan_status"`
	Creator            string     `gorm:"size:50;comment:'创建人'" json:"creator"`
	CreatedAt          time.Time  `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"comment:'更新时间'" json:"updated_at"`
}

type InspectionTask struct {
	ID             uint64     `gorm:"primaryKey;comment:'巡检任务ID，系统内部主键'" json:"id"`
	PlanID         *uint64    `gorm:"index;uniqueIndex:idx_plan_scheduled,priority:1;comment:'任务计划ID，关联task_plans.id，手动创建时为空'" json:"plan_id"`
	TemplateID     uint64     `gorm:"index;not null;comment:'巡检模板ID，关联inspection_templates.id'" json:"template_id"`
	ScheduledAt    *time.Time `gorm:"uniqueIndex:idx_plan_scheduled,priority:2;comment:'计划执行时间，手动创建时可为空'" json:"scheduled_at"`
	DueAt          time.Time  `gorm:"index;not null;comment:'任务截止时间'" json:"due_at"`
	Status         string     `gorm:"size:32;index:idx_task_status_due_id,priority:1;not null;comment:'任务状态：pending待领取，assigned已分配，in_progress执行中，submitted已提交，completed已完成，overdue已逾期，cancelled已取消'" json:"status"`
	AssigneeType   string     `gorm:"size:16;not null;comment:'指派类型：user用户，team班组'" json:"assignee_type"`
	AssigneeID     uint64     `gorm:"index;not null;comment:'指派对象ID，根据assignee_type关联users.id或teams.id'" json:"assignee_id"`
	ExecutorID     *uint64    `gorm:"index;comment:'实际执行人用户ID，领取班组任务后写入'" json:"executor_id"`
	PointName      string     `gorm:"size:128;not null;default:'';comment:'巡检点位名称'" json:"point_name"`
	EquipmentName  string     `gorm:"size:128;not null;default:'';comment:'巡检设备名称'" json:"equipment_name"`
	StartedAt      *time.Time `gorm:"comment:'开始执行时间'" json:"started_at"`
	SubmittedAt    *time.Time `gorm:"comment:'提交时间'" json:"submitted_at"`
	CompletedAt    *time.Time `gorm:"comment:'后台确认完成时间'" json:"completed_at"`
	CancelledAt    *time.Time `gorm:"comment:'取消时间'" json:"cancelled_at"`
	TaskName       string     `gorm:"size:255;comment:'任务名称'" json:"task_name"`
	InspectArea    string     `gorm:"size:255;comment:'巡检区域'" json:"inspect_area"`
	GlassesSN      string     `gorm:"size:50;comment:'指定的AR眼镜序列号'" json:"glasses_sn"`
	AssignUser     string     `gorm:"size:50;comment:'下发人'" json:"assign_user"`
	AssignTime     *time.Time `gorm:"comment:'下发时间'" json:"assign_time"`
	CreatedAt      time.Time  `gorm:"index:idx_task_status_due_id,priority:3;comment:'创建时间'" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"comment:'更新时间'" json:"updated_at"`
}

type InspectionTaskNode struct {
	ID             uint64     `gorm:"primaryKey;comment:'任务节点ID，系统内部主键'" json:"id"`
	TaskID         uint64     `gorm:"uniqueIndex:idx_task_node_sort;index;not null;comment:'巡检任务ID，关联inspection_tasks.id'" json:"task_id"`
	TemplateNodeID uint64     `gorm:"index;not null;comment:'模板节点ID，关联inspection_template_nodes.id'" json:"template_node_id"`
	SortOrder      int        `gorm:"uniqueIndex:idx_task_node_sort;not null;comment:'节点排序号，数值越小越靠前'" json:"sort_order"`
	Name           string     `gorm:"size:128;not null;comment:'任务节点名称，生成任务时从模板节点复制'" json:"name"`
	NodeType       string     `gorm:"size:32;not null;comment:'节点类型：text/read/check/photo/video/audio'" json:"node_type"`
	MinPhotos      int        `gorm:"not null;default:0;comment:'最少照片数量要求'" json:"min_photos"`
	RequireText    bool       `gorm:"type:tinyint(1);not null;default:0;comment:'是否要求填写文本说明'" json:"require_text"`
	AllowAbnormal  bool       `gorm:"type:tinyint(1);not null;default:0;comment:'是否允许在该节点上报异常'" json:"allow_abnormal"`
	Status         string     `gorm:"size:32;index;not null;comment:'节点状态：pending待提交，completed已完成，abnormal异常'" json:"status"`
	NodesConfigID  string     `gorm:"size:32;comment:'节点配置ID，关联template_nodes_config'" json:"nodes_config_id"`
	TaskTypeCode   string     `gorm:"size:30;comment:'任务类型编码：text/read/check/photo/video/audio'" json:"task_type_code"`
	IsMandatory    bool       `gorm:"not null;default:false;comment:'是否强制执行'" json:"is_mandatory"`
	IsRequired     bool       `gorm:"not null;default:false;comment:'是否必做节点'" json:"is_required"`
	AlgorithmID    string     `gorm:"size:32;comment:'绑定AI算法ID'" json:"algorithm_id"`
	QueryID        string     `gorm:"size:32;comment:'绑定实时查询接口ID'" json:"query_id"`
	ActualExecTime *time.Time `gorm:"comment:'实际执行时间'" json:"actual_exec_time"`
	Remark         string     `gorm:"size:128;comment:'备注'" json:"remark"`
	CreatedAt      time.Time  `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"comment:'更新时间'" json:"updated_at"`
}

type TaskNodeResult struct {
	ID               uint64     `gorm:"primaryKey;comment:'节点结果ID，系统内部主键'" json:"id"`
	TaskID           uint64     `gorm:"uniqueIndex:idx_result_task_node;index;not null;comment:'巡检任务ID，关联inspection_tasks.id'" json:"task_id"`
	NodeID           uint64     `gorm:"uniqueIndex:idx_result_task_node;index;not null;comment:'任务节点ID，关联inspection_task_nodes.id'" json:"node_id"`
	UserID           uint64     `gorm:"index;not null;comment:'提交人用户ID，关联users.id'" json:"user_id"`
	Status           string     `gorm:"size:32;not null;comment:'结果状态：completed正常完成，abnormal异常'" json:"status"`
	TextNote         string     `gorm:"type:text;comment:'节点文字说明或备注'" json:"text_note"`
	IdempotencyKey   string     `gorm:"size:128;index;not null;comment:'幂等键，用于防止弱网重复提交'" json:"idempotency_key"`
	CompletedAt      time.Time  `gorm:"comment:'节点完成时间'" json:"completed_at"`
	TaskTypeCode     string     `gorm:"size:32;comment:'任务类型编码：text/read/check/photo/video/audio'" json:"task_type_code"`
	FeedbackContent  string     `gorm:"size:512;comment:'核心反馈内容：check类传选项值，read类传读数值，text类传文本'" json:"feedback_content"`
	AlgorithmResult  string     `gorm:"size:512;comment:'AI算法分析结果'" json:"algorithm_result"`
	QueryResult      string     `gorm:"size:512;comment:'实时查询数据'" json:"query_result"`
	LocationGPS      string     `gorm:"size:50;comment:'GPS坐标，格式lng,lat'" json:"location_gps"`
	AttachmentIDs    string     `gorm:"size:256;comment:'附件ID列表，多个用逗号分隔'" json:"attachment_ids"`
	IsAbnormal       bool       `gorm:"type:tinyint(1);comment:'是否异常：1是，0否'" json:"is_abnormal"`
	AbnormalDesc     string     `gorm:"size:128;comment:'异常描述'" json:"abnormal_desc"`
	Remark           string     `gorm:"size:128;comment:'备注'" json:"remark"`
	CreatedAt        time.Time  `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"comment:'更新时间'" json:"updated_at"`
}

type Attachment struct {
	ID          uint64     `gorm:"primaryKey;comment:'附件ID，系统内部主键'" json:"id"`
	ObjectKey   string     `gorm:"size:255;uniqueIndex;not null;comment:'对象存储Key，用于定位MinIO/S3中的文件'" json:"object_key"`
	FileName    string     `gorm:"size:255;not null;default:'';comment:'原始文件名'" json:"file_name"`
	ContentType string     `gorm:"size:128;not null;comment:'文件MIME类型，例如image/jpeg、audio/mpeg'" json:"content_type"`
	SizeBytes   int64      `gorm:"not null;comment:'文件大小，单位字节'" json:"size_bytes"`
	SHA256      string     `gorm:"size:64;not null;default:'';comment:'文件SHA-256摘要，用于完整性校验'" json:"sha256"`
	BindStatus  string     `gorm:"size:32;index:idx_attachment_bind_created,priority:1;not null;comment:'绑定状态：uploaded已上传，bound已绑定，orphaned孤立'" json:"bind_status"`
	TaskID      *uint64    `gorm:"index;comment:'关联巡检任务ID，关联inspection_tasks.id'" json:"task_id"`
	NodeID      *uint64    `gorm:"index;comment:'关联任务节点ID，关联inspection_task_nodes.id'" json:"node_id"`
	ResultID    *uint64    `gorm:"index;comment:'关联节点结果ID，关联task_node_results.id'" json:"result_id"`
	UserID      uint64     `gorm:"index;not null;comment:'上传人用户ID，关联users.id'" json:"user_id"`
	DeviceID    *uint64    `gorm:"index;comment:'上传设备ID，关联devices.id'" json:"device_id"`
	CaptureTime *time.Time `gorm:"comment:'现场采集时间'" json:"capture_time"`
	UploadTime  *time.Time `gorm:"comment:'上传完成时间'" json:"upload_time"`
	GPSLat      *float64   `gorm:"comment:'采集位置纬度'" json:"gps_lat"`
	GPSLng      *float64   `gorm:"comment:'采集位置经度'" json:"gps_lng"`
	CreatedAt   time.Time  `gorm:"index:idx_attachment_bind_created,priority:2;comment:'创建时间'" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"comment:'更新时间'" json:"updated_at"`
}

type Defect struct {
	ID          uint64     `gorm:"primaryKey;comment:'缺陷ID，系统内部主键'" json:"id"`
	TaskID      uint64     `gorm:"index;not null;comment:'巡检任务ID，关联inspection_tasks.id'" json:"task_id"`
	NodeID      uint64     `gorm:"index;not null;comment:'任务节点ID，关联inspection_task_nodes.id'" json:"node_id"`
	ReporterID  uint64     `gorm:"index;not null;comment:'上报人用户ID，关联users.id'" json:"reporter_id"`
	Status      string     `gorm:"size:32;index:idx_defects_status_created_id,priority:1;not null;comment:'缺陷状态：reported已上报，confirmed已确认，closed已关闭'" json:"status"`
	Description string     `gorm:"type:text;comment:'缺陷描述'" json:"description"`
	CloseReason string     `gorm:"type:text;comment:'关闭原因'" json:"close_reason"`
	ConfirmedAt *time.Time `gorm:"comment:'确认时间'" json:"confirmed_at"`
	ClosedAt    *time.Time `gorm:"comment:'关闭时间'" json:"closed_at"`
	CreatedAt   time.Time  `gorm:"index:idx_defects_status_created_id,priority:2;comment:'创建时间'" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"comment:'更新时间'" json:"updated_at"`
}

type AuditLog struct {
	ID        uint64    `gorm:"primaryKey;comment:'审计日志ID，系统内部主键'" json:"id"`
	ActorID   uint64    `gorm:"index;not null;comment:'操作人用户ID，关联users.id'" json:"actor_id"`
	Action    string    `gorm:"size:128;not null;comment:'操作动作，例如create、update、delete、submit'" json:"action"`
	Target    string    `gorm:"size:128;not null;comment:'操作对象类型，例如task、device、template'" json:"target"`
	TargetID  uint64    `gorm:"index;not null;comment:'操作对象ID'" json:"target_id"`
	Detail    string    `gorm:"type:text;comment:'操作详情，通常为JSON或文本说明'" json:"detail"`
	CreatedAt time.Time `gorm:"comment:'创建时间'" json:"created_at"`
}

type OutboxEvent struct {
	ID          uint64     `gorm:"primaryKey;comment:'事件ID，系统内部主键'" json:"id"`
	EventKey    string     `gorm:"size:160;uniqueIndex;not null;comment:'事件唯一键，用于保证事件幂等写入和发布'" json:"event_key"`
	Topic       string     `gorm:"size:128;index;not null;comment:'事件主题或队列名，例如task.assigned'" json:"topic"`
	Payload     string     `gorm:"type:json;not null;comment:'事件载荷JSON，保存异步消息内容'" json:"payload"`
	PublishedAt *time.Time `gorm:"comment:'发布时间，未发布时为空'" json:"published_at"`
	CreatedAt   time.Time  `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"comment:'更新时间'" json:"updated_at"`
}

type InspectionPoint struct {
	ID            uint64    `gorm:"primaryKey;comment:'巡检点位ID，系统内部主键'" json:"id"`
	Name          string    `gorm:"size:128;not null;comment:'巡检点位名称'" json:"name"`
	EquipmentName string    `gorm:"size:128;not null;default:'';comment:'关联设备名称'" json:"equipment_name"`
	Location      string    `gorm:"size:256;comment:'点位位置描述'" json:"location"`
	Area          string    `gorm:"size:128;comment:'所属区域'" json:"area"`
	Substation    string    `gorm:"size:128;comment:'所属变电站'" json:"substation"`
	Description   string    `gorm:"size:256;comment:'点位说明'" json:"description"`
	Enabled       bool      `gorm:"not null;default:1;comment:'是否启用'" json:"enabled"`
	CreatedAt     time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt     time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
}
