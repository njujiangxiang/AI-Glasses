// Package workflows 工作流配置管理，支持多步骤、多类型输入、异常触发规则配置。
package workflows

import (
	"encoding/json"
	"errors"
	"strings"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"gorm.io/gorm"
)

const (
	StatusDraft     = "draft"     // 草稿
	StatusPublished = "published" // 已发布
	StatusArchived  = "archived"  // 已归档

	StepTypeText   = "text"   // 文本输入
	StepTypeNumber = "number" // 数值输入
	StepTypeSelect = "select" // 选择清单
	StepTypePhoto  = "photo"  // 拍照
	StepTypeVideo  = "video"  // 录像
	StepTypeAudio  = "audio"  // 录音
)

var validStepTypes = map[string]bool{
	StepTypeText:   true,
	StepTypeNumber: true,
	StepTypeSelect: true,
	StepTypePhoto:  true,
	StepTypeVideo:  true,
	StepTypeAudio:  true,
}

// SelectOption 选择项
type SelectOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// WorkflowInput 创建和更新工作流的输入
type WorkflowInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// StepInput 步骤创建和更新输入
type StepInput struct {
	Name                    string         `json:"name"`
	Description             string         `json:"description"`
	Type                    string         `json:"type"`
	Required                bool           `json:"required"`
	Options                 []SelectOption `json:"options"`
	AbnormalEnabled         bool           `json:"abnormal_enabled"`
	AbnormalRequirePhoto    bool           `json:"abnormal_require_photo"`
	AbnormalRequireVideo    bool           `json:"abnormal_require_video"`
	AbnormalRequireNote     bool           `json:"abnormal_require_note"`
	AbnormalRequireSignature bool          `json:"abnormal_require_signature"`
}

// WorkflowDetail 工作流详情（包含步骤）
type WorkflowDetail struct {
	Workflow database.Workflow  `json:"workflow"`
	Steps    []database.WorkflowStep `json:"steps"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Create 创建工作流
func (s *Service) Create(createdBy uint64, input WorkflowInput) (database.Workflow, error) {
	model := database.Workflow{
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Status:      StatusDraft,
		CreatedBy:   createdBy,
	}
	if err := s.validateWorkflow(model, 0); err != nil {
		return database.Workflow{}, err
	}
	return model, s.db.Create(&model).Error
}

// Update 更新工作流基本信息
func (s *Service) Update(id uint64, input WorkflowInput) (database.Workflow, error) {
	var model database.Workflow
	if err := s.db.First(&model, id).Error; err != nil {
		return database.Workflow{}, notFound(err, "工作流不存在")
	}
	model.Name = strings.TrimSpace(input.Name)
	model.Description = strings.TrimSpace(input.Description)
	if err := s.validateWorkflow(model, id); err != nil {
		return database.Workflow{}, err
	}
	return model, s.db.Save(&model).Error
}

// Get 获取工作流详情
func (s *Service) Get(id uint64) (database.Workflow, error) {
	var model database.Workflow
	err := s.db.First(&model, id).Error
	return model, notFound(err, "工作流不存在")
}

// GetDetail 获取工作流详情（包含步骤）
func (s *Service) GetDetail(id uint64) (WorkflowDetail, error) {
	var workflow database.Workflow
	if err := s.db.First(&workflow, id).Error; err != nil {
		return WorkflowDetail{}, notFound(err, "工作流不存在")
	}
	var steps []database.WorkflowStep
	err := s.db.Where("workflow_id = ?", id).Order("sort_order asc").Find(&steps).Error
	return WorkflowDetail{Workflow: workflow, Steps: steps}, err
}

// List 查询工作流列表
func (s *Service) List(keyword string, status string) ([]database.Workflow, error) {
	var items []database.Workflow
	query := s.db.Order("updated_at desc")

	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		like := "%" + escapeLike(keyword) + "%"
		query = query.Where(`name LIKE ? ESCAPE '\' OR description LIKE ? ESCAPE '\'`, like, like)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	return items, query.Find(&items).Error
}

// Delete 删除工作流（仅草稿状态可删除）
func (s *Service) Delete(id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var model database.Workflow
		if err := tx.First(&model, id).Error; err != nil {
			return notFound(err, "工作流不存在")
		}
		if model.Status == StatusPublished {
			return httperr.New(httperr.ValidationFailed, "已发布的工作流不能删除，请先取消发布")
		}
		if err := tx.Where("workflow_id = ?", id).Delete(&database.WorkflowStep{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model).Error
	})
}

// Publish 发布工作流
func (s *Service) Publish(id uint64) error {
	var steps []database.WorkflowStep
	if err := s.db.Where("workflow_id = ?", id).Find(&steps).Error; err != nil {
		return err
	}
	if len(steps) == 0 {
		return httperr.New(httperr.ValidationFailed, "请先添加至少一个步骤后再发布")
	}
	result := s.db.Model(&database.Workflow{}).Where("id = ?", id).Update("status", StatusPublished)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "工作流不存在")
	}
	return nil
}

// Unpublish 取消发布
func (s *Service) Unpublish(id uint64) error {
	result := s.db.Model(&database.Workflow{}).Where("id = ?", id).Update("status", StatusDraft)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "工作流不存在")
	}
	return nil
}

// AddStep 添加步骤
func (s *Service) AddStep(workflowID uint64, input StepInput) (database.WorkflowStep, error) {
	var maxSort int
	s.db.Model(&database.WorkflowStep{}).Where("workflow_id = ?", workflowID).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxSort)

	model := database.WorkflowStep{
		WorkflowID:               workflowID,
		SortOrder:                maxSort + 1,
		Name:                     strings.TrimSpace(input.Name),
		Description:              strings.TrimSpace(input.Description),
		Type:                     input.Type,
		Required:                 input.Required,
		AbnormalEnabled:          input.AbnormalEnabled,
		AbnormalRequirePhoto:     input.AbnormalRequirePhoto,
		AbnormalRequireVideo:     input.AbnormalRequireVideo,
		AbnormalRequireNote:      input.AbnormalRequireNote,
		AbnormalRequireSignature: input.AbnormalRequireSignature,
	}
	// 仅 select 类型保存 options，其他类型设为 nil（MySQL JSON 不能存空字符串）
	if input.Type == StepTypeSelect && len(input.Options) > 0 {
		optionsJSON, _ := json.Marshal(input.Options)
		s := string(optionsJSON)
		model.OptionsJSON = &s
	} else {
		model.OptionsJSON = nil
	}
	if err := s.validateStep(model, 0, input.Options); err != nil {
		return database.WorkflowStep{}, err
	}
	return model, s.db.Create(&model).Error
}

// UpdateStep 更新步骤
func (s *Service) UpdateStep(id uint64, input StepInput) (database.WorkflowStep, error) {
	var model database.WorkflowStep
	if err := s.db.First(&model, id).Error; err != nil {
		return database.WorkflowStep{}, notFound(err, "步骤不存在")
	}
	model.Name = strings.TrimSpace(input.Name)
	model.Description = strings.TrimSpace(input.Description)
	model.Type = input.Type
	model.Required = input.Required
	model.AbnormalEnabled = input.AbnormalEnabled
	model.AbnormalRequirePhoto = input.AbnormalRequirePhoto
	model.AbnormalRequireVideo = input.AbnormalRequireVideo
	model.AbnormalRequireNote = input.AbnormalRequireNote
	model.AbnormalRequireSignature = input.AbnormalRequireSignature
	// 仅 select 类型保存 options，其他类型设为 nil（MySQL JSON 不能存空字符串）
	if input.Type == StepTypeSelect && len(input.Options) > 0 {
		optionsJSON, _ := json.Marshal(input.Options)
		s := string(optionsJSON)
		model.OptionsJSON = &s
	} else {
		model.OptionsJSON = nil
	}
	if err := s.validateStep(model, id, input.Options); err != nil {
		return database.WorkflowStep{}, err
	}
	return model, s.db.Save(&model).Error
}

// DeleteStep 删除步骤
func (s *Service) DeleteStep(id uint64) error {
	result := s.db.Delete(&database.WorkflowStep{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "步骤不存在")
	}
	return nil
}

// ReorderSteps 重新排序步骤
func (s *Service) ReorderSteps(workflowID uint64, stepIDs []uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i, stepID := range stepIDs {
			err := tx.Model(&database.WorkflowStep{}).
				Where("id = ? AND workflow_id = ?", stepID, workflowID).
				Update("sort_order", i+1).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// DuplicateStep 复制步骤
func (s *Service) DuplicateStep(id uint64) (database.WorkflowStep, error) {
	var step database.WorkflowStep
	if err := s.db.First(&step, id).Error; err != nil {
		return database.WorkflowStep{}, notFound(err, "步骤不存在")
	}
	var maxSort int
	s.db.Model(&database.WorkflowStep{}).Where("workflow_id = ?", step.WorkflowID).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxSort)

	newStep := database.WorkflowStep{
		WorkflowID:               step.WorkflowID,
		SortOrder:                maxSort + 1,
		Name:                     step.Name + " (副本)",
		Description:              step.Description,
		Type:                     step.Type,
		Required:                 step.Required,
		OptionsJSON:              step.OptionsJSON,
		AbnormalEnabled:          step.AbnormalEnabled,
		AbnormalRequirePhoto:     step.AbnormalRequirePhoto,
		AbnormalRequireVideo:     step.AbnormalRequireVideo,
		AbnormalRequireNote:      step.AbnormalRequireNote,
		AbnormalRequireSignature: step.AbnormalRequireSignature,
	}
	return newStep, s.db.Create(&newStep).Error
}

func (s *Service) validateWorkflow(model database.Workflow, currentID uint64) error {
	if model.Name == "" {
		return httperr.New(httperr.ValidationFailed, "工作流名称不能为空")
	}
	if len([]rune(model.Name)) > 128 {
		return httperr.New(httperr.ValidationFailed, "工作流名称不能超过128个字符")
	}
	if len([]rune(model.Description)) > 512 {
		return httperr.New(httperr.ValidationFailed, "工作流描述不能超过512个字符")
	}
	if model.Status != StatusDraft && model.Status != StatusPublished && model.Status != StatusArchived {
		return httperr.New(httperr.ValidationFailed, "工作流状态不正确")
	}
	return nil
}

func (s *Service) validateStep(model database.WorkflowStep, currentID uint64, options []SelectOption) error {
	if model.Name == "" {
		return httperr.New(httperr.ValidationFailed, "步骤名称不能为空")
	}
	if len([]rune(model.Name)) > 128 {
		return httperr.New(httperr.ValidationFailed, "步骤名称不能超过128个字符")
	}
	if !validStepTypes[model.Type] {
		return httperr.New(httperr.ValidationFailed, "不支持的步骤类型")
	}
	return nil
}

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

func notFound(err error, message string) error {
	if err == gorm.ErrRecordNotFound {
		return httperr.New(httperr.ResourceNotFound, message)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return httperr.New(httperr.ResourceNotFound, message)
	}
	return err
}
