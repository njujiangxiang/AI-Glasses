
// Package database 定义 MVP 使用的持久化数据模型。模型覆盖用户、角色、设备生命周期、
// 巡检模板、生成任务、证据附件、缺陷、审计日志以及调度/事件流水线使用的 outbox 事件。
package database

import "time"

type User struct {
	ID                uint64    `gorm:"primaryKey" json:"id"`
	Username          string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash      string    `gorm:"size:255;not null" json:"-"`
	DisplayName       string    `gorm:"size:128" json:"display_name"`
	Name              string    `gorm:"size:64;not null;default:''" json:"name"`
	Gender            string    `gorm:"size:8;not null;default:''" json:"gender"`
	AvatarData        []byte    `gorm:"type:longblob" json:"-"`
	AvatarContentType string    `gorm:"size:64;not null;default:''" json:"avatar_content_type"`
	AvatarSize        int64     `gorm:"not null;default:0" json:"avatar_size"`
	BirthYear         int       `gorm:"not null;default:0" json:"birth_year"`
	BirthMonth        int       `gorm:"not null;default:0" json:"birth_month"`
	IDCardNo          string    `gorm:"size:32;index;not null;default:''" json:"id_card_no"`
	OrgCode           string    `gorm:"size:64;index;not null;default:''" json:"org_code"`
	Status            string    `gorm:"size:32;index;not null" json:"status"`
	Phone             string    `gorm:"size:11" json:"phone"`
	Email             string    `gorm:"size:128" json:"email"`
	IsDeleted         bool      `gorm:"type:tinyint(1);default:0" json:"is_deleted"`
	Role              int       `json:"role"`
	UserType          int       `gorm:"type:tinyint(1)" json:"user_type"`
	UserStatus        int       `gorm:"type:tinyint(1)" json:"user_status"`
	LastLoginTime     *time.Time `json:"last_login_time"`
	UserLock          bool      `gorm:"type:tinyint(1)" json:"user_lock"`
	PwdExpireTime     *time.Time `json:"pwd_expire_time"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Organization struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	Code       string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name       string    `gorm:"size:128;not null" json:"name"`
	ParentCode string    `gorm:"size:64;index;not null;default:''" json:"parent_code"`
	Status     string    `gorm:"size:32;index;not null" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type BusinessCode struct {
	ID           uint64 `gorm:"primaryKey;comment:业务编码配置ID，系统内部主键" json:"id"`
	Name         string `gorm:"size:128;not null;comment:编码名称，用于后台展示规则用途" json:"name"`
	Code         string `gorm:"size:64;uniqueIndex;not null;comment:业务代码，系统内唯一，例如TK" json:"code"`
	DateFormat   string `gorm:"size:32;not null;comment:日期格式，首版仅支持yyyyMMdd" json:"date_format"`
	SeqPadding   int    `gorm:"not null;comment:流水号位数，例如4表示0001" json:"seq_padding"`
	Separator    string `gorm:"size:8;not null;default:'';comment:分隔符，例如-，不使用时为空" json:"separator"`
	UseSeparator bool   `gorm:"not null;default:false;comment:是否在代码、日期、流水号之间使用分隔符" json:"use_separator"`
	Status       string `gorm:"size:32;not null;default:'active';comment:编码状态：active启用，disabled停用" json:"status"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Role struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Permission struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"size:128;uniqueIndex;not null" json:"code"`
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
	ID         uint64     `gorm:"primaryKey" json:"id"`
	SerialNo   string     `gorm:"size:128;uniqueIndex;not null" json:"serial_no"`
	Name       string     `gorm:"size:128" json:"name"`
	Status     string     `gorm:"size:32;index;not null" json:"status"`
	BoundUserID *uint64   `gorm:"index" json:"bound_user_id"`
	BoundAt    *time.Time `json:"bound_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type DeviceSession struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	DeviceID      uint64    `gorm:"index;not null" json:"device_id"`
	UserID        uint64    `gorm:"index;not null" json:"user_id"`
	RefreshJTI    string    `gorm:"size:64;uniqueIndex;not null" json:"refresh_jti"`
	Status        string    `gorm:"size:32;index;not null" json:"status"`
	RefreshUntil  time.Time `json:"refresh_until"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DeviceAuditLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	DeviceID  uint64    `gorm:"index;not null" json:"device_id"`
	ActorID   uint64    `gorm:"index;not null" json:"actor_id"`
	Action    string    `gorm:"size:64;not null" json:"action"`
	Reason    string    `gorm:"size:255" json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

// ========== 新增配置表 ==========

type TaskTypeDict struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	TypeCode         string    `gorm:"size:30" json:"type_code"`
	TypeName         string    `gorm:"size:50" json:"type_name"`
	TypeDesc         string    `gorm:"size:255" json:"type_desc"`
	SupportAlgorithm bool      `gorm:"type:tinyint(1)" json:"support_algorithm"`
	SupportQuery     bool      `gorm:"type:tinyint(1)" json:"support_query"`
	SupportMandatory bool      `gorm:"type:tinyint(1)" json:"support_mandatory"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Remark           string    `gorm:"size:256" json:"remark"`
}

type AlgorithmConfig struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:100" json:"name"`
	Type          string    `gorm:"size:50" json:"type"`
	ServiceURL    string    `gorm:"size:255" json:"service_url"`
	CallMethod    string    `gorm:"size:20" json:"call_method"`
	InputParams   string    `gorm:"size:512" json:"input_params"`
	OutputParams  string    `gorm:"size:512" json:"output_params"`
	Version       string    `gorm:"size:20" json:"version"`
	IsEnable      bool      `gorm:"type:tinyint(1)" json:"is_enable"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Remark        string    `gorm:"size:255" json:"remark"`
}

type RealtimeQueryConfig struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"size:100" json:"name"`
	ApiURL         string    `gorm:"size:255" json:"api_url"`
	RequestMethod  string    `gorm:"size:10" json:"request_method"`
	AuthType       string    `gorm:"size:30" json:"auth_type"`
	RequestParams  string    `gorm:"size:255" json:"request_params"`
	ResponseParams string    `gorm:"size:255" json:"response_params"`
	TimeoutSecond  int       `json:"timeout_second"`
	ApplyScene     string    `gorm:"size:100" json:"apply_scene"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Remark         string    `gorm:"size:255" json:"remark"`
}

// ========== 模板相关 ==========

type InspectionTemplate struct {
	ID              uint64    `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"size:128;not null" json:"name"`
	Description     string    `gorm:"size:512" json:"description"`
	ApplicableRoles string    `gorm:"size:255" json:"applicable_roles"`
	Enabled         bool      `gorm:"type:tinyint(1);index;not null" json:"enabled"`
	Type            string    `gorm:"size:50" json:"type"`
	Scene           string    `gorm:"size:100" json:"scene"`
	Version         string    `gorm:"size:20" json:"version"`
	IsEnable        bool      `gorm:"type:tinyint(1)" json:"is_enable"`
	Creator         string    `gorm:"size:50" json:"creator"`
	Remark          string    `gorm:"size:255" json:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type InspectionTemplateNode struct {
	ID                 uint64    `gorm:"primaryKey" json:"id"`
	TemplateID         uint64    `gorm:"uniqueIndex:idx_template_sort;index;not null" json:"template_id"`
	SortOrder          int       `gorm:"uniqueIndex:idx_template_sort;not null" json:"sort_order"`
	Name               string    `gorm:"size:128;not null" json:"name"`
	Description        string    `gorm:"size:512" json:"description"`
	NodeType           string    `gorm:"size:32;not null" json:"node_type"`
	MinPhotos          int       `gorm:"not null" json:"min_photos"`
	RequireText        bool      `gorm:"type:tinyint(1);not null" json:"require_text"`
	AllowAbnormal      bool      `gorm:"type:tinyint(1);not null" json:"allow_abnormal"`
	RequireLiveCapture bool      `gorm:"type:tinyint(1);not null" json:"require_live_capture"`
	TaskTypeID         string    `gorm:"size:32" json:"task_type_id"`
	IsMandatory        bool      `gorm:"type:tinyint(1)" json:"is_mandatory"`
	IsRequired         bool      `gorm:"type:tinyint(1)" json:"is_required"`
	AlgorithmID        string    `gorm:"size:32" json:"algorithm_id"`
	QueryID            string    `gorm:"size:32" json:"query_id"`
	TimeoutSecond      int       `json:"timeout_second"`
	Remark             string    `gorm:"size:128" json:"remark"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type TemplateNodeConfig struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	NodeID      uint64    `gorm:"index;not null" json:"node_id"`
	ConfigCode  string    `gorm:"size:50" json:"config_code"`
	ConfigName  string    `gorm:"size:100" json:"config_name"`
	ConfigValue string    `gorm:"size:256" json:"config_value"`
	ConfigType  string    `gorm:"size:30" json:"config_type"`
	Sort        int       `json:"sort"`
	IsDefault   bool      `gorm:"type:tinyint(1)" json:"is_default"`
	Remark      string    `gorm:"size:128" json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
	ID                    uint64  `gorm:"primaryKey;comment:步骤ID" json:"id"`
	WorkflowID            uint64  `gorm:"uniqueIndex:idx_workflow_sort;index;not null;comment:所属工作流ID" json:"workflow_id"`
	SortOrder             int     `gorm:"uniqueIndex:idx_workflow_sort;not null;comment:排序序号" json:"sort_order"`
	Name                  string  `gorm:"size:128;not null;comment:步骤名称" json:"name"`
	Description           string  `gorm:"size:512;comment:步骤描述" json:"description"`
	Type                  string  `gorm:"size:32;not null;comment:步骤类型：text,number,select,photo,video,audio" json:"type"`
	Required              bool    `gorm:"not null;default:true;comment:是否必填" json:"required"`
	OptionsJSON           *string `gorm:"type:json;comment:选择项配置JSON，仅select类型使用" json:"options_json"`
	AbnormalEnabled       bool    `gorm:"not null;default:false;comment:是否启用异常触发" json:"abnormal_enabled"`
	AbnormalRequirePhoto  bool    `gorm:"not null;default:true;comment:异常时必须拍照" json:"abnormal_require_photo"`
	AbnormalRequireVideo  bool    `gorm:"not null;default:false;comment:异常时必须录像" json:"abnormal_require_video"`
	AbnormalRequireNote   bool    `gorm:"not null;default:true;comment:异常时必须填写备注" json:"abnormal_require_note"`
	AbnormalRequireSignature bool `gorm:"not null;default:false;comment:异常时必须签字确认" json:"abnormal_require_signature"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// ========== 任务计划 ==========

type TaskPlan struct {
	ID                 uint64    `gorm:"primaryKey" json:"id"`
	TemplateID         uint64    `gorm:"index;not null" json:"template_id"`
	Name               string    `gorm:"size:128;not null" json:"name"`
	CronExpr           string    `gorm:"size:64;not null" json:"cron_expr"`
	Timezone           string    `gorm:"size:64;not null" json:"timezone"`
	StartAt            time.Time `gorm:"index;not null" json:"start_at"`
	DueDurationMinutes int       `gorm:"not null" json:"due_duration_minutes"`
	AssigneeType       string    `gorm:"size:16;not null" json:"assignee_type"`
	AssigneeID         uint64    `gorm:"index;not null" json:"assignee_id"`
	PointName          string    `gorm:"size:128" json:"point_name"`
	EquipmentName      string    `gorm:"size:128" json:"equipment_name"`
	Enabled            bool      `gorm:"index;not null" json:"enabled"`
	PlanType           string    `gorm:"size:50" json:"plan_type"`
	BelongUnit         string    `gorm:"size:100" json:"belong_unit"`
	OperatorUnit       string    `gorm:"size:100" json:"operator_unit"`
	SubstationName     string    `gorm:"size:100" json:"substation_name"`
	InspectArea        string    `gorm:"size:256" json:"inspect_area"`
	PlanStartTime      *time.Time `json:"plan_start_time"`
	PlanEndTime        *time.Time `json:"plan_end_time"`
	PlanPrincipal      string    `gorm:"size:50" json:"plan_principal"`
	OperatorUser       string    `gorm:"size:50" json:"operator_user"`
	Guardian           string    `gorm:"size:50" json:"guardian"`
	PlanDesc           string    `gorm:"size:256" json:"plan_desc"`
	PlanStatus         string    `gorm:"size:20" json:"plan_status"`
	Creator            string    `gorm:"size:50" json:"creator"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type InspectionTask struct {
	ID             uint64     `gorm:"primaryKey" json:"id"`
	PlanID         uint64     `gorm:"uniqueIndex:idx_task_plan_schedule_assignee;index;not null" json:"plan_id"`
	TemplateID     uint64     `gorm:"index;not null" json:"template_id"`
	ScheduledAt    time.Time  `gorm:"uniqueIndex:idx_task_plan_schedule_assignee;index;not null" json:"scheduled_at"`
	DueAt          time.Time  `gorm:"index;not null" json:"due_at"`
	Status         string     `gorm:"size:32;index:idx_task_status_due_id,priority:1;not null" json:"status"`
	AssigneeType   string     `gorm:"size:16;uniqueIndex:idx_task_plan_schedule_assignee;not null" json:"assignee_type"`
	AssigneeID     uint64     `gorm:"uniqueIndex:idx_task_plan_schedule_assignee;index;not null" json:"assignee_id"`
	ExecutorID     *uint64    `gorm:"index" json:"executor_id"`
	PointName      string     `gorm:"size:128" json:"point_name"`
	EquipmentName  string     `gorm:"size:128" json:"equipment_name"`
	StartedAt      *time.Time `json:"started_at"`
	SubmittedAt    *time.Time `json:"submitted_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	CancelledAt    *time.Time `json:"cancelled_at"`
	TaskName       string     `gorm:"size:255" json:"task_name"`
	InspectArea    string     `gorm:"size:255" json:"inspect_area"`
	GlassesSN      string     `gorm:"size:50" json:"glasses_sn"`
	AssignUser     string     `gorm:"size:50" json:"assign_user"`
	AssignTime     *time.Time `json:"assign_time"`
	CreatedAt      time.Time  `gorm:"index:idx_task_status_due_id,priority:3" json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type InspectionTaskNode struct {
	ID             uint64     `gorm:"primaryKey" json:"id"`
	TaskID         uint64     `gorm:"uniqueIndex:idx_task_node_sort;index;not null" json:"task_id"`
	TemplateNodeID uint64     `gorm:"index;not null" json:"template_node_id"`
	SortOrder      int        `gorm:"uniqueIndex:idx_task_node_sort;not null" json:"sort_order"`
	Name           string     `gorm:"size:128;not null" json:"name"`
	NodeType       string     `gorm:"size:32;not null" json:"node_type"`
	MinPhotos      int        `gorm:"not null" json:"min_photos"`
	RequireText    bool       `gorm:"type:tinyint(1);not null" json:"require_text"`
	AllowAbnormal  bool       `gorm:"type:tinyint(1);not null" json:"allow_abnormal"`
	Status         string     `gorm:"size:32;index;not null" json:"status"`
	TaskTypeCode   string     `gorm:"size:30" json:"task_type_code"`
	IsMandatory    string     `gorm:"size:1" json:"is_mandatory"`
	IsRequired     string     `gorm:"size:1" json:"is_required"`
	AlgorithmID    string     `gorm:"size:32" json:"algorithm_id"`
	QueryID        string     `gorm:"size:32" json:"query_id"`
	ActualExecTime *time.Time `json:"actual_exec_time"`
	Remark         string     `gorm:"size:128" json:"remark"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type TaskNodeResult struct {
	ID               uint64     `gorm:"primaryKey" json:"id"`
	TaskID           uint64     `gorm:"uniqueIndex:idx_result_task_node;index;not null" json:"task_id"`
	NodeID           uint64     `gorm:"uniqueIndex:idx_result_task_node;index;not null" json:"node_id"`
	UserID           uint64     `gorm:"index;not null" json:"user_id"`
	Status           string     `gorm:"size:32;not null" json:"status"`
	TextNote         string     `gorm:"type:text" json:"text_note"`
	IdempotencyKey   string     `gorm:"size:128;index;not null" json:"idempotency_key"`
	CompletedAt      time.Time  `json:"completed_at"`
	TaskTypeCode     string     `gorm:"size:32" json:"task_type_code"`
	FeedbackContent  string     `gorm:"size:512" json:"feedback_content"`
	AlgorithmResult  string     `gorm:"size:512" json:"algorithm_result"`
	QueryResult      string     `gorm:"size:512" json:"query_result"`
	LocationGPS      string     `gorm:"size:50" json:"location_gps"`
	AttachmentIDs    string     `gorm:"size:256" json:"attachment_ids"`
	IsAbnormal       bool       `gorm:"type:tinyint(1)" json:"is_abnormal"`
	AbnormalDesc     string     `gorm:"size:128" json:"abnormal_desc"`
	Remark           string     `gorm:"size:128" json:"remark"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Attachment struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	ObjectKey   string     `gorm:"size:255;uniqueIndex;not null" json:"object_key"`
	FileName    string     `gorm:"size:255" json:"file_name"`
	ContentType string     `gorm:"size:128;not null" json:"content_type"`
	SizeBytes   int64      `gorm:"not null" json:"size_bytes"`
	SHA256      string     `gorm:"size:64" json:"sha256"`
	BindStatus  string     `gorm:"size:32;index:idx_attachment_bind_created,priority:1;not null" json:"bind_status"`
	TaskID      *uint64    `gorm:"index" json:"task_id"`
	NodeID      *uint64    `gorm:"index" json:"node_id"`
	ResultID    *uint64    `gorm:"index" json:"result_id"`
	UserID      uint64     `gorm:"index;not null" json:"user_id"`
	DeviceID    *uint64    `gorm:"index" json:"device_id"`
	CaptureTime *time.Time `json:"capture_time"`
	UploadTime  *time.Time `json:"upload_time"`
	GPSLat      *float64   `json:"gps_lat"`
	GPSLng      *float64   `json:"gps_lng"`
	CreatedAt   time.Time  `gorm:"index:idx_attachment_bind_created,priority:2" json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Defect struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	TaskID      uint64     `gorm:"index;not null" json:"task_id"`
	NodeID      uint64     `gorm:"index;not null" json:"node_id"`
	ReporterID  uint64     `gorm:"index;not null" json:"reporter_id"`
	Status      string     `gorm:"size:32;index;not null" json:"status"`
	Description string     `gorm:"type:text" json:"description"`
	CloseReason string     `gorm:"type:text" json:"close_reason"`
	ConfirmedAt *time.Time `json:"confirmed_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	CreatedAt   time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type AuditLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	ActorID   uint64    `gorm:"index;not null" json:"actor_id"`
	Action    string    `gorm:"size:128;not null" json:"action"`
	Target    string    `gorm:"size:128;not null" json:"target"`
	TargetID  uint64    `gorm:"index;not null" json:"target_id"`
	Detail    string    `gorm:"type:text" json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

type OutboxEvent struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	EventKey    string     `gorm:"size:160;uniqueIndex;not null" json:"event_key"`
	Topic       string     `gorm:"size:128;index;not null" json:"topic"`
	Payload     string     `gorm:"type:json;not null" json:"payload"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
