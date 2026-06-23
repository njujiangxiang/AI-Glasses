package plans

import (
	"strings"
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
	TemplateID         uint64     `json:"template_id"`
	Name               string     `json:"name"`
	CronExpr           string     `json:"cron_expr"`
	Timezone           string     `json:"timezone"`
	StartAt            time.Time  `json:"start_at"`
	DueDurationMinutes int        `json:"due_duration_minutes"`
	AssigneeType       string     `json:"assignee_type"`
	AssigneeID         uint64     `json:"assignee_id"`
	PointName          string     `json:"point_name"`
	EquipmentName      string     `json:"equipment_name"`
	PlanType           string     `json:"plan_type"`
	BelongUnit         string     `json:"belong_unit"`
	OperatorUnit       string     `json:"operator_unit"`
	SubstationName     string     `json:"substation_name"`
	InspectArea        string     `json:"inspect_area"`
	PlanStartTime      *time.Time `json:"plan_start_time"`
	PlanEndTime        *time.Time `json:"plan_end_time"`
	PlanPrincipal      string     `json:"plan_principal"`
	OperatorUser       string     `json:"operator_user"`
	Guardian           string     `json:"guardian"`
	PlanDesc           string     `json:"plan_desc"`
	Creator            string     `json:"creator"`
}

type UpdateInput struct {
	TemplateID         uint64     `json:"template_id"`
	Name               string     `json:"name"`
	CronExpr           string     `json:"cron_expr"`
	Timezone           string     `json:"timezone"`
	StartAt            *time.Time `json:"start_at"`
	DueDurationMinutes int        `json:"due_duration_minutes"`
	AssigneeType       string     `json:"assignee_type"`
	AssigneeID         uint64     `json:"assignee_id"`
	PointName          string     `json:"point_name"`
	EquipmentName      string     `json:"equipment_name"`
	PlanType           string     `json:"plan_type"`
	BelongUnit         string     `json:"belong_unit"`
	OperatorUnit       string     `json:"operator_unit"`
	SubstationName     string     `json:"substation_name"`
	InspectArea        string     `json:"inspect_area"`
	PlanStartTime      *time.Time `json:"plan_start_time"`
	PlanEndTime        *time.Time `json:"plan_end_time"`
	PlanPrincipal      string     `json:"plan_principal"`
	OperatorUser       string     `json:"operator_user"`
	Guardian           string     `json:"guardian"`
	PlanDesc           string     `json:"plan_desc"`
}

type ListQuery struct {
	Keyword  string
	Status   string
	Page     int
	PageSize int
}

type ListResult struct {
	Items []database.TaskPlan `json:"items"`
	Total int64               `json:"total"`
}

type Service struct {
	db *gorm.DB
}

// NewService 创建任务计划服务，注入数据库访问能力。
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// Create 创建巡检任务计划。
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

	// 验证模板存在
	var template database.InspectionTemplate
	if err := s.db.First(&template, input.TemplateID).Error; err != nil {
		return database.TaskPlan{}, httperr.New(httperr.ResourceNotFound, "template not found")
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
		PlanType:           input.PlanType,
		BelongUnit:         input.BelongUnit,
		OperatorUnit:       input.OperatorUnit,
		SubstationName:     input.SubstationName,
		InspectArea:        input.InspectArea,
		PlanStartTime:      input.PlanStartTime,
		PlanEndTime:        input.PlanEndTime,
		PlanPrincipal:      input.PlanPrincipal,
		OperatorUser:       input.OperatorUser,
		Guardian:           input.Guardian,
		PlanDesc:           input.PlanDesc,
		PlanStatus:         "pending",
		Creator:            input.Creator,
	}
	return plan, s.db.Create(&plan).Error
}

// List 查询任务计划列表，支持分页和筛选。
func (s *Service) List(query ListQuery) (ListResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	db := s.db.Model(&database.TaskPlan{})
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + escapeLike(keyword) + "%"
		db = db.Where("name LIKE ? OR point_name LIKE ? OR plan_desc LIKE ?", like, like, like)
	}
	if status := strings.TrimSpace(query.Status); status != "" {
		if status == "enabled" {
			db = db.Where("enabled = ?", true)
		} else if status == "disabled" {
			db = db.Where("enabled = ?", false)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return ListResult{}, err
	}

	var items []database.TaskPlan
	if err := db.Order("id desc").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error; err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, Total: total}, nil
}

// GetDetail 查询计划详情。
func (s *Service) GetDetail(id uint64) (database.TaskPlan, error) {
	var plan database.TaskPlan
	if err := s.db.First(&plan, id).Error; err != nil {
		return plan, notFound(err, "plan not found")
	}
	return plan, nil
}

// Update 更新计划。
func (s *Service) Update(id uint64, input UpdateInput) (database.TaskPlan, error) {
	var plan database.TaskPlan
	if err := s.db.First(&plan, id).Error; err != nil {
		return plan, notFound(err, "plan not found")
	}

	if input.Name == "" {
		return database.TaskPlan{}, httperr.New(httperr.TaskStateConflict, "plan name is required")
	}
	if input.TemplateID == 0 {
		return database.TaskPlan{}, httperr.New(httperr.TaskStateConflict, "template_id is required")
	}
	if input.AssigneeType != AssigneeUser && input.AssigneeType != AssigneeTeam {
		return database.TaskPlan{}, httperr.New(httperr.TaskStateConflict, "invalid assignee type")
	}

	// 验证模板存在
	var template database.InspectionTemplate
	if err := s.db.First(&template, input.TemplateID).Error; err != nil {
		return database.TaskPlan{}, httperr.New(httperr.ResourceNotFound, "template not found")
	}

	// 更新 cron 和 timezone
	if input.CronExpr != "" {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if _, err := parser.Parse(input.CronExpr); err != nil {
			return database.TaskPlan{}, httperr.New(httperr.TaskStateConflict, "invalid cron expression")
		}
		plan.CronExpr = input.CronExpr
	}
	if input.Timezone != "" {
		loc, err := time.LoadLocation(input.Timezone)
		if err != nil {
			return database.TaskPlan{}, httperr.New(httperr.TaskStateConflict, "invalid timezone")
		}
		plan.Timezone = loc.String()
	}
	if input.StartAt != nil {
		plan.StartAt = input.StartAt.UTC()
	}

	plan.TemplateID = input.TemplateID
	plan.Name = input.Name
	plan.DueDurationMinutes = input.DueDurationMinutes
	plan.AssigneeType = input.AssigneeType
	plan.AssigneeID = input.AssigneeID
	plan.PointName = input.PointName
	plan.EquipmentName = input.EquipmentName
	plan.PlanType = input.PlanType
	plan.BelongUnit = input.BelongUnit
	plan.OperatorUnit = input.OperatorUnit
	plan.SubstationName = input.SubstationName
	plan.InspectArea = input.InspectArea
	plan.PlanStartTime = input.PlanStartTime
	plan.PlanEndTime = input.PlanEndTime
	plan.PlanPrincipal = input.PlanPrincipal
	plan.OperatorUser = input.OperatorUser
	plan.Guardian = input.Guardian
	plan.PlanDesc = input.PlanDesc

	return plan, s.db.Save(&plan).Error
}

// Enable 启用计划。
func (s *Service) Enable(id uint64) error {
	result := s.db.Model(&database.TaskPlan{}).Where("id = ?", id).Update("enabled", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "plan not found")
	}
	return nil
}

// Disable 停用计划。
func (s *Service) Disable(id uint64) error {
	result := s.db.Model(&database.TaskPlan{}).Where("id = ?", id).Update("enabled", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "plan not found")
	}
	return nil
}

// Delete 删除计划。
func (s *Service) Delete(id uint64) error {
	result := s.db.Delete(&database.TaskPlan{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "plan not found")
	}
	return nil
}

func notFound(err error, message string) error {
	if err == gorm.ErrRecordNotFound {
		return httperr.New(httperr.ResourceNotFound, message)
	}
	return err
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
