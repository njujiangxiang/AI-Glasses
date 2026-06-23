package templates

import (
	"strings"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"gorm.io/gorm"
)

const (
	NodeText  = "text"
	NodeRead  = "read"
	NodeCheck = "check"
	NodePhoto = "photo"
	NodeVideo = "video"
	NodeAudio = "audio"
)

type NodeInput struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	NodeDesc           string `json:"node_desc"`
	NodeType           string `json:"node_type"`
	MinPhotos          int    `json:"min_photos"`
	RequireText        bool   `json:"require_text"`
	AllowAbnormal      bool   `json:"allow_abnormal"`
	RequireLiveCapture bool   `json:"require_live_capture"`
	NodesConfigID      string `json:"nodes_config_id"`
	TaskTypeID         string `json:"task_type_id"`
	IsMandatory        string `json:"is_mandatory"`
	IsRequired         string `json:"is_required"`
	AlgorithmID        string `json:"algorithm_id"`
	QueryID            string `json:"query_id"`
	TimeoutSecond      int    `json:"timeout_second"`
	Remark             string `json:"remark"`
}

type CreateInput struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	ApplicableRoles string      `json:"applicable_roles"`
	Enabled         bool        `json:"enabled"`
	Type            string      `json:"type"`
	Scene           string      `json:"scene"`
	Version         string      `json:"version"`
	Creator         string      `json:"creator"`
	Remark          string      `json:"remark"`
	NodeIDs         []uint64    `json:"node_ids"`         // 选择已有节点
	Nodes           []NodeInput `json:"nodes"`            // 向后兼容：内联创建节点
}

type UpdateInput struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	ApplicableRoles string      `json:"applicable_roles"`
	Enabled         *bool       `json:"enabled"`
	Type            string      `json:"type"`
	Scene           string      `json:"scene"`
	Version         string      `json:"version"`
	Remark          string      `json:"remark"`
	NodeIDs         []uint64    `json:"node_ids"`
}

type ListQuery struct {
	Keyword  string
	Page     int
	PageSize int
}

type ListResult struct {
	Items []database.InspectionTemplate `json:"items"`
	Total int64                         `json:"total"`
}

type Service struct {
	db *gorm.DB
}

// NewService 创建巡检模板服务，注入数据库访问能力。
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// Create 创建巡检模板。支持两种方式：选择已有节点（node_ids）或内联创建节点（nodes）。
func (s *Service) Create(input CreateInput) (database.InspectionTemplate, error) {
	if input.Name == "" {
		return database.InspectionTemplate{}, httperr.New(httperr.TaskStateConflict, "template name is required")
	}
	if len(input.NodeIDs) == 0 && len(input.Nodes) == 0 {
		return database.InspectionTemplate{}, httperr.New(httperr.TaskStateConflict, "template must have at least one node")
	}

	// 校验内联节点
	for _, node := range input.Nodes {
		if node.Name == "" || !validNodeType(node.NodeType) || node.MinPhotos < 0 {
			return database.InspectionTemplate{}, httperr.New(httperr.TaskStateConflict, "invalid template node")
		}
	}

	template := database.InspectionTemplate{
		Name:            input.Name,
		Description:     input.Description,
		ApplicableRoles: input.ApplicableRoles,
		Enabled:         input.Enabled,
		Type:            input.Type,
		Scene:           input.Scene,
		Version:         input.Version,
		Creator:         input.Creator,
		Remark:          input.Remark,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&template).Error; err != nil {
			return err
		}

		// 方式1：分配已有节点到模板
		if len(input.NodeIDs) > 0 {
			// 获取节点当前最大 sort_order
			var maxSort int
			tx.Model(&database.InspectionTemplateNode{}).
				Where("template_id = ?", template.ID).
				Select("COALESCE(MAX(sort_order), 0)").
				Scan(&maxSort)

			for i, nodeID := range input.NodeIDs {
				if err := tx.Model(&database.InspectionTemplateNode{}).
					Where("id = ? AND template_id IS NULL", nodeID).
					Updates(map[string]any{
						"template_id": template.ID,
						"sort_order":  maxSort + i + 1,
					}).Error; err != nil {
					return err
				}
			}
		}

		// 方式2：内联创建节点（向后兼容）
		for index, node := range input.Nodes {
			requireLiveCapture := node.RequireLiveCapture || node.MinPhotos > 0
			isMandatory := node.IsMandatory
			if isMandatory == "" {
				isMandatory = "1"
			}
			isRequired := node.IsRequired
			if isRequired == "" {
				isRequired = "1"
			}
			model := database.InspectionTemplateNode{
				TemplateID:         &template.ID,
				SortOrder:          index + 1,
				Name:               node.Name,
				Description:        node.Description,
				NodeDesc:           node.NodeDesc,
				NodeType:           node.NodeType,
				MinPhotos:          node.MinPhotos,
				RequireText:        node.RequireText,
				AllowAbnormal:      node.AllowAbnormal,
				RequireLiveCapture: requireLiveCapture,
				NodesConfigID:      node.NodesConfigID,
				TaskTypeID:         node.TaskTypeID,
				IsMandatory:        isMandatory,
				IsRequired:         isRequired,
				AlgorithmID:        node.AlgorithmID,
				QueryID:            node.QueryID,
				TimeoutSecond:      node.TimeoutSecond,
				Remark:             node.Remark,
			}
			if err := tx.Create(&model).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return template, err
}

// List 查询巡检模板列表，支持分页和关键词搜索。
func (s *Service) List(query ListQuery) (ListResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	db := s.db.Model(&database.InspectionTemplate{})
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + escapeLike(keyword) + "%"
		db = db.Where("name LIKE ? OR description LIKE ?", like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return ListResult{}, err
	}

	var items []database.InspectionTemplate
	if err := db.Order("id desc").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error; err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, Total: total}, nil
}

// ListAll 查询所有模板（不分页，用于下拉选择）。
func (s *Service) ListAll() ([]database.InspectionTemplate, error) {
	var templates []database.InspectionTemplate
	return templates, s.db.Order("id desc").Find(&templates).Error
}

// GetDetail 查询模板详情（含节点列表）。
func (s *Service) GetDetail(id uint64) (database.InspectionTemplate, []database.InspectionTemplateNode, error) {
	var template database.InspectionTemplate
	if err := s.db.First(&template, id).Error; err != nil {
		return template, nil, notFound(err, "template not found")
	}
	var nodes []database.InspectionTemplateNode
	if err := s.db.Where("template_id = ?", id).Order("sort_order asc, id asc").Find(&nodes).Error; err != nil {
		return template, nil, err
	}
	return template, nodes, nil
}

// Update 更新模板信息和节点关联。
func (s *Service) Update(id uint64, input UpdateInput) (database.InspectionTemplate, error) {
	var template database.InspectionTemplate
	if err := s.db.First(&template, id).Error; err != nil {
		return template, notFound(err, "template not found")
	}

	if input.Name == "" {
		return database.InspectionTemplate{}, httperr.New(httperr.TaskStateConflict, "template name is required")
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 更新基本信息
		template.Name = input.Name
		template.Description = input.Description
		template.ApplicableRoles = input.ApplicableRoles
		template.Type = input.Type
		template.Scene = input.Scene
		template.Version = input.Version
		template.Remark = input.Remark
		if input.Enabled != nil {
			template.Enabled = *input.Enabled
		}
		if err := tx.Save(&template).Error; err != nil {
			return err
		}

		// 如果提供了 node_ids，重新分配节点
		if input.NodeIDs != nil {
			// 先释放当前模板下的节点
			if err := tx.Model(&database.InspectionTemplateNode{}).
				Where("template_id = ?", id).
				Update("template_id", nil).Error; err != nil {
				return err
			}
			// 再分配新选择的节点
			for i, nodeID := range input.NodeIDs {
				if err := tx.Model(&database.InspectionTemplateNode{}).
					Where("id = ? AND template_id IS NULL", nodeID).
					Updates(map[string]any{
						"template_id": id,
						"sort_order":  i + 1,
					}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	return template, err
}

// Enable 启用模板。
func (s *Service) Enable(id uint64) error {
	result := s.db.Model(&database.InspectionTemplate{}).Where("id = ?", id).Update("enabled", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "template not found")
	}
	return nil
}

// Disable 停用模板。
func (s *Service) Disable(id uint64) error {
	result := s.db.Model(&database.InspectionTemplate{}).Where("id = ?", id).Update("enabled", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "template not found")
	}
	return nil
}

// Delete 删除模板，同时释放关联的节点（将 template_id 设为 NULL）。
func (s *Service) Delete(id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 释放节点
		if err := tx.Model(&database.InspectionTemplateNode{}).
			Where("template_id = ?", id).
			Update("template_id", nil).Error; err != nil {
			return err
		}
		// 删除模板
		result := tx.Delete(&database.InspectionTemplate{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return httperr.New(httperr.ResourceNotFound, "template not found")
		}
		return nil
	})
}

// GetNodeCount 查询模板下的节点数量。
func (s *Service) GetNodeCount(templateID uint64) (int64, error) {
	var count int64
	err := s.db.Model(&database.InspectionTemplateNode{}).Where("template_id = ?", templateID).Count(&count).Error
	return count, err
}

// validNodeType 判断模板节点类型是否属于系统允许范围。
func validNodeType(nodeType string) bool {
	switch nodeType {
	case NodeText, NodeRead, NodeCheck, NodePhoto, NodeVideo, NodeAudio:
		return true
	default:
		return false
	}
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
