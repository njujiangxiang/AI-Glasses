// Package database 定义 MVP 使用的持久化数据模型。模型覆盖用户、角色、设备生命周期、
// 巡检模板、生成任务、证据附件、缺陷、审计日志以及调度/事件流水线使用的 outbox 事件。
package database

import "time"

type User struct {
	ID                uint64 `gorm:"primaryKey" json:"id"`
	Username          string `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash      string `gorm:"size:255;not null" json:"-"`
	DisplayName       string `gorm:"size:128" json:"display_name"`
	Name              string `gorm:"size:64;not null;default:''" json:"name"`
	Gender            string `gorm:"size:8;not null;default:''" json:"gender"`
	AvatarData        []byte `gorm:"type:longblob" json:"-"`
	AvatarContentType string `gorm:"size:64;not null;default:''" json:"avatar_content_type"`
	AvatarSize        int64  `gorm:"not null;default:0" json:"avatar_size"`
	BirthYear         int    `gorm:"not null;default:0" json:"birth_year"`
	BirthMonth        int    `gorm:"not null;default:0" json:"birth_month"`
	IDCardNo          string `gorm:"size:32;index;not null;default:''" json:"id_card_no"`
	OrgCode           string `gorm:"size:64;index;not null;default:''" json:"org_code"`
	Status            string `gorm:"size:32;index;not null" json:"status"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Organization struct {
	ID         uint64 `gorm:"primaryKey" json:"id"`
	Code       string `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name       string `gorm:"size:128;not null" json:"name"`
	ParentCode string `gorm:"size:64;index;not null;default:''" json:"parent_code"`
	Status     string `gorm:"size:32;index;not null" json:"status"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
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

type TaskTypeDict struct {
	ID               uint64 `gorm:"primaryKey" json:"id"`
	TypeCode         string `gorm:"size:64;uniqueIndex;not null" json:"type_code"`
	TypeName         string `gorm:"size:128;not null" json:"type_name"`
	TypeDesc         string `gorm:"size:512" json:"type_desc"`
	SupportAlgorithm bool   `gorm:"not null;default:false" json:"support_algorithm"`
	SupportQuery     bool   `gorm:"not null;default:false" json:"support_query"`
	SupportMandatory bool   `gorm:"not null;default:false" json:"support_mandatory"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type AlgorithmConfig struct {
	ID           uint64 `gorm:"primaryKey" json:"id"`
	Name         string `gorm:"size:128;not null" json:"name"`
	ServiceURL   string `gorm:"size:255" json:"service_url"`
	InputParams  string `gorm:"type:text" json:"input_params"`
	OutputParams string `gorm:"type:text" json:"output_params"`
	IsEnable     bool   `gorm:"index;not null;default:true" json:"is_enable"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RealtimeQueryConfig struct {
	ID             uint64 `gorm:"primaryKey" json:"id"`
	Name           string `gorm:"size:128;not null" json:"name"`
	APIURL         string `gorm:"size:255" json:"api_url"`
	RequestParams  string `gorm:"type:text" json:"request_params"`
	ResponseParams string `gorm:"type:text" json:"response_params"`
	IsEnable       bool   `gorm:"index;not null;default:true" json:"is_enable"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TemplateNodeConfig struct {
	ID             uint64 `gorm:"primaryKey" json:"id"`
	TemplateNodeID uint64 `gorm:"index;not null" json:"template_node_id"`
	ConfigCode     string `gorm:"size:64;not null" json:"config_code"`
	ConfigName     string `gorm:"size:128;not null" json:"config_name"`
	ConfigValue    string `gorm:"size:255" json:"config_value"`
	IsDefault      bool   `gorm:"not null;default:false" json:"is_default"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Role struct {
	ID        uint64 `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"size:64;uniqueIndex;not null" json:"name"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Permission struct {
	ID        uint64 `gorm:"primaryKey" json:"id"`
	Code      string `gorm:"size:128;uniqueIndex;not null" json:"code"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRole struct {
	UserID uint64 `gorm:"primaryKey"`
	RoleID uint64 `gorm:"primaryKey"`
}

type RolePermission struct {
	RoleID       uint64 `gorm:"primaryKey"`
	PermissionID uint64 `gorm:"primaryKey"`
}

type Team struct {
	ID        uint64 `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"size:128;not null" json:"name"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TeamMember struct {
	TeamID uint64 `gorm:"primaryKey"`
	UserID uint64 `gorm:"primaryKey"`
}

type Device struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	SerialNo    string     `gorm:"size:128;uniqueIndex;not null" json:"serial_no"`
	Name        string     `gorm:"size:128" json:"name"`
	Status      string     `gorm:"size:32;index;not null" json:"status"`
	BoundUserID *uint64    `gorm:"index" json:"bound_user_id"`
	BoundAt     *time.Time `json:"bound_at"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DeviceSession struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	DeviceID     uint64    `gorm:"index;not null" json:"device_id"`
	UserID       uint64    `gorm:"index;not null" json:"user_id"`
	RefreshJTI   string    `gorm:"size:64;uniqueIndex;not null" json:"refresh_jti"`
	Status       string    `gorm:"size:32;index;not null" json:"status"`
	RefreshUntil time.Time `json:"refresh_until"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type DeviceAuditLog struct {
	ID        uint64 `gorm:"primaryKey" json:"id"`
	DeviceID  uint64 `gorm:"index;not null" json:"device_id"`
	ActorID   uint64 `gorm:"index;not null" json:"actor_id"`
	Action    string `gorm:"size:64;not null" json:"action"`
	Reason    string `gorm:"size:255" json:"reason"`
	CreatedAt time.Time
}

type InspectionTemplate struct {
	ID              uint64 `gorm:"primaryKey" json:"id"`
	Name            string `gorm:"size:128;not null" json:"name"`
	Description     string `gorm:"size:512" json:"description"`
	ApplicableRoles string `gorm:"size:255" json:"applicable_roles"`
	Enabled         bool   `gorm:"index;not null" json:"enabled"`
	Type            string `gorm:"size:50;not null;default:''" json:"type"`
	Scene           string `gorm:"size:100;not null;default:''" json:"scene"`
	Version         string `gorm:"size:20;not null;default:'v1'" json:"version"`
	IsEnable        bool   `gorm:"not null;default:true" json:"is_enable"`
	Creator         string `gorm:"size:50;not null;default:''" json:"creator"`
	Remark          string `gorm:"size:255" json:"remark"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type InspectionTemplateNode struct {
	ID                 uint64 `gorm:"primaryKey" json:"id"`
	TemplateID         uint64 `gorm:"uniqueIndex:idx_template_sort;index;not null" json:"template_id"`
	SortOrder          int    `gorm:"uniqueIndex:idx_template_sort;not null" json:"sort_order"`
	Name               string `gorm:"size:128;not null" json:"name"`
	Description        string `gorm:"size:512" json:"description"`
	NodeType           string `gorm:"size:32;not null" json:"node_type"`
	MinPhotos          int    `gorm:"not null" json:"min_photos"`
	RequireText        bool   `gorm:"not null" json:"require_text"`
	AllowAbnormal      bool   `gorm:"not null" json:"allow_abnormal"`
	RequireLiveCapture bool   `gorm:"not null" json:"require_live_capture"`
	NodeDesc           string `gorm:"size:512;not null;default:''" json:"node_desc"`
	NodesConfigID      string `gorm:"size:32" json:"nodes_config_id"`
	TaskTypeID         string `gorm:"size:32" json:"task_type_id"`
	IsMandatory        bool   `gorm:"not null;default:false" json:"is_mandatory"`
	IsRequired         bool   `gorm:"not null;default:false" json:"is_required"`
	AlgorithmID        string `gorm:"size:32" json:"algorithm_id"`
	QueryID            string `gorm:"size:32" json:"query_id"`
	TimeoutSecond      int    `json:"timeout_second"`
	Remark             string `gorm:"size:128" json:"remark"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

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
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type InspectionTask struct {
	ID            uint64     `gorm:"primaryKey" json:"id"`
	PlanID        uint64     `gorm:"uniqueIndex:idx_task_plan_schedule_assignee;index;not null" json:"plan_id"`
	TemplateID    uint64     `gorm:"index;not null" json:"template_id"`
	ScheduledAt   time.Time  `gorm:"uniqueIndex:idx_task_plan_schedule_assignee;index;not null" json:"scheduled_at"`
	DueAt         time.Time  `gorm:"index;not null" json:"due_at"`
	Status        string     `gorm:"size:32;index:idx_task_status_due_id,priority:1;not null" json:"status"`
	AssigneeType  string     `gorm:"size:16;uniqueIndex:idx_task_plan_schedule_assignee;not null" json:"assignee_type"`
	AssigneeID    uint64     `gorm:"uniqueIndex:idx_task_plan_schedule_assignee;index;not null" json:"assignee_id"`
	ExecutorID    *uint64    `gorm:"index" json:"executor_id"`
	PointName     string     `gorm:"size:128" json:"point_name"`
	EquipmentName string     `gorm:"size:128" json:"equipment_name"`
	StartedAt     *time.Time `json:"started_at"`
	SubmittedAt   *time.Time `json:"submitted_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	CancelledAt   *time.Time `json:"cancelled_at"`
	CreatedAt     time.Time  `gorm:"index:idx_task_status_due_id,priority:3"`
	UpdatedAt     time.Time
}

type InspectionTaskNode struct {
	ID             uint64 `gorm:"primaryKey" json:"id"`
	TaskID         uint64 `gorm:"uniqueIndex:idx_task_node_sort;index;not null" json:"task_id"`
	TemplateNodeID uint64 `gorm:"index;not null" json:"template_node_id"`
	SortOrder      int    `gorm:"uniqueIndex:idx_task_node_sort;not null" json:"sort_order"`
	Name           string `gorm:"size:128;not null" json:"name"`
	NodeType       string `gorm:"size:32;not null" json:"node_type"`
	MinPhotos      int    `gorm:"not null" json:"min_photos"`
	RequireText    bool   `gorm:"not null" json:"require_text"`
	AllowAbnormal  bool   `gorm:"not null" json:"allow_abnormal"`
	Status         string `gorm:"size:32;index;not null" json:"status"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TaskNodeResult struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	TaskID         uint64    `gorm:"uniqueIndex:idx_result_task_node;index;not null" json:"task_id"`
	NodeID         uint64    `gorm:"uniqueIndex:idx_result_task_node;index;not null" json:"node_id"`
	UserID         uint64    `gorm:"index;not null" json:"user_id"`
	Status         string    `gorm:"size:32;not null" json:"status"`
	TextNote       string    `gorm:"type:text" json:"text_note"`
	IdempotencyKey string    `gorm:"size:128;index;not null" json:"idempotency_key"`
	CompletedAt    time.Time `json:"completed_at"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
	CreatedAt   time.Time  `gorm:"index:idx_attachment_bind_created,priority:2"`
	UpdatedAt   time.Time
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
	CreatedAt   time.Time  `gorm:"index"`
	UpdatedAt   time.Time
}

type AuditLog struct {
	ID        uint64 `gorm:"primaryKey" json:"id"`
	ActorID   uint64 `gorm:"index;not null" json:"actor_id"`
	Action    string `gorm:"size:128;not null" json:"action"`
	Target    string `gorm:"size:128;not null" json:"target"`
	TargetID  uint64 `gorm:"index;not null" json:"target_id"`
	Detail    string `gorm:"type:text" json:"detail"`
	CreatedAt time.Time
}

type OutboxEvent struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	EventKey    string     `gorm:"size:160;uniqueIndex;not null" json:"event_key"`
	Topic       string     `gorm:"size:128;index;not null" json:"topic"`
	Payload     string     `gorm:"type:json;not null" json:"payload"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
