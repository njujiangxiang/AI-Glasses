package plans

import (
	"time"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

const (
	AssigneeUser = "user"
	AssigneeTeam = "team"
)

type CreateInput struct {
	TemplateID         uint64    `json:"template_id"`
	Name               string    `json:"name"`
	CronExpr           string    `json:"cron_expr"`
	Timezone           string    `json:"timezone"`
	StartAt            time.Time `json:"start_at"`
	DueDurationMinutes int       `json:"due_duration_minutes"`
	AssigneeType       string    `json:"assignee_type"`
	AssigneeID         uint64    `json:"assignee_id"`
	PointName          string    `json:"point_name"`
	EquipmentName      string    `json:"equipment_name"`
}

type Service struct {
	db *gorm.DB
}

// NewService 创建任务计划服务，注入数据库访问能力。
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// Create 创建巡检任务计划，包含模板、指派对象、点位设备和 cron 配置。
func (s *Service) Create(input CreateInput) (database.TaskPlan, error) {
	if input.Name == "" || input.TemplateID == 0 || input.AssigneeID == 0 || input.DueDurationMinutes <= 0 {
		return database.TaskPlan{}, httperr.New(httperr.TaskStateConflict, "plan required fields missing")
	}
	if input.AssigneeType != AssigneeUser && input.AssigneeType != AssigneeTeam {
		return database.TaskPlan{}, httperr.New(httperr.TaskStateConflict, "invalid assignee type")
	}
	if input.Timezone == "" {
		input.Timezone = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(input.Timezone)
	if err != nil {
		return database.TaskPlan{}, httperr.New(httperr.TaskStateConflict, "invalid timezone")
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	if _, err := parser.Parse(input.CronExpr); err != nil {
		return database.TaskPlan{}, httperr.New(httperr.TaskStateConflict, "invalid cron expression")
	}
	plan := database.TaskPlan{
		TemplateID:         input.TemplateID,
		Name:               input.Name,
		CronExpr:           input.CronExpr,
		Timezone:           loc.String(),
		StartAt:            input.StartAt.UTC(),
		DueDurationMinutes: input.DueDurationMinutes,
		AssigneeType:       input.AssigneeType,
		AssigneeID:         input.AssigneeID,
		PointName:          input.PointName,
		EquipmentName:      input.EquipmentName,
		Enabled:            false,
	}
	return plan, s.db.Create(&plan).Error
}

// Enable 启用指定任务计划，使调度器开始按计划生成任务。
func (s *Service) Enable(id uint64) error {
	return s.db.Model(&database.TaskPlan{}).Where("id = ?", id).Update("enabled", true).Error
}

// List 查询任务计划列表，供后台计划管理页面展示。
func (s *Service) List() ([]database.TaskPlan, error) {
	var plans []database.TaskPlan
	return plans, s.db.Order("id desc").Find(&plans).Error
}
