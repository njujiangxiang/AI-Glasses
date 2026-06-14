package templates

import (
	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"gorm.io/gorm"
)

const (
	NodeCheckin  = "checkin"
	NodePhoto    = "photo"
	NodeText     = "text"
	NodeNumber   = "number"
	NodeAbnormal = "abnormal"
	NodeConfirm  = "confirm"
)

type NodeInput struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	NodeType           string `json:"node_type"`
	MinPhotos          int    `json:"min_photos"`
	RequireText        bool   `json:"require_text"`
	AllowAbnormal      bool   `json:"allow_abnormal"`
	RequireLiveCapture bool   `json:"require_live_capture"`
}

type CreateInput struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	ApplicableRoles string      `json:"applicable_roles"`
	Enabled         bool        `json:"enabled"`
	Nodes           []NodeInput `json:"nodes"`
}

type Service struct {
	db *gorm.DB
}

// NewService 创建巡检模板服务，注入数据库访问能力。
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// Create 创建巡检模板和模板节点，并校验节点类型是否合法。
func (s *Service) Create(input CreateInput) (database.InspectionTemplate, error) {
	if input.Name == "" || len(input.Nodes) == 0 {
		return database.InspectionTemplate{}, httperr.New(httperr.TaskStateConflict, "template name and nodes are required")
	}
	for _, node := range input.Nodes {
		if node.Name == "" || !validNodeType(node.NodeType) || node.MinPhotos < 0 {
			return database.InspectionTemplate{}, httperr.New(httperr.TaskStateConflict, "invalid template node")
		}
	}
	template := database.InspectionTemplate{Name: input.Name, Description: input.Description, ApplicableRoles: input.ApplicableRoles, Enabled: input.Enabled}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&template).Error; err != nil {
			return err
		}
		for index, node := range input.Nodes {
			requireLiveCapture := node.RequireLiveCapture || node.MinPhotos > 0
			model := database.InspectionTemplateNode{
				TemplateID:         template.ID,
				SortOrder:          index + 1,
				Name:               node.Name,
				Description:        node.Description,
				NodeType:           node.NodeType,
				MinPhotos:          node.MinPhotos,
				RequireText:        node.RequireText,
				AllowAbnormal:      node.AllowAbnormal,
				RequireLiveCapture: requireLiveCapture,
			}
			if err := tx.Create(&model).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return template, err
}

// List 查询巡检模板列表，供后台模板管理页面展示。
func (s *Service) List() ([]database.InspectionTemplate, error) {
	var templates []database.InspectionTemplate
	return templates, s.db.Order("id desc").Find(&templates).Error
}

// validNodeType 判断模板节点类型是否属于系统允许范围。
func validNodeType(nodeType string) bool {
	switch nodeType {
	case NodeCheckin, NodePhoto, NodeText, NodeNumber, NodeAbnormal, NodeConfirm:
		return true
	default:
		return false
	}
}
