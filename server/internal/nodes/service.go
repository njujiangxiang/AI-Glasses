// Package nodes 实现巡检节点管理。节点是巡检模板的组成单元，可以独立创建后分配给模板使用。
package nodes

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

type CreateInput struct {
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

type UpdateInput struct {
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

type ListQuery struct {
	Keyword    string
	NodeType   string
	Assigned   *bool // nil=全部, true=已分配, false=未分配
	Page       int
	PageSize   int
}

type ListResult struct {
	Items []database.InspectionTemplateNode `json:"items"`
	Total int64                             `json:"total"`
}

type Service struct {
	db *gorm.DB
}

// NewService 创建节点管理服务，注入数据库访问能力。
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// List 查询节点列表，支持关键词搜索、类型筛选和分配状态筛选。
func (s *Service) List(query ListQuery) (ListResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	db := s.db.Model(&database.InspectionTemplateNode{})

	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + escapeLike(keyword) + "%"
		db = db.Where("name LIKE ? OR description LIKE ?", like, like)
	}
	if nodeType := strings.TrimSpace(query.NodeType); nodeType != "" {
		db = db.Where("node_type = ?", nodeType)
	}
	if query.Assigned != nil {
		if *query.Assigned {
			db = db.Where("template_id IS NOT NULL")
		} else {
			db = db.Where("template_id IS NULL")
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return ListResult{}, err
	}

	var items []database.InspectionTemplateNode
	if err := db.Order("id desc").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error; err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, Total: total}, nil
}

// Get 查询单个节点详情。
func (s *Service) Get(id uint64) (database.InspectionTemplateNode, error) {
	var node database.InspectionTemplateNode
	if err := s.db.First(&node, id).Error; err != nil {
		return node, notFound(err, "node not found")
	}
	return node, nil
}

// Create 创建巡检节点，节点创建后处于未分配状态（template_id 为空）。
func (s *Service) Create(input CreateInput) (database.InspectionTemplateNode, error) {
	if input.Name == "" {
		return database.InspectionTemplateNode{}, httperr.New(httperr.TaskStateConflict, "node name is required")
	}
	if !validNodeType(input.NodeType) {
		return database.InspectionTemplateNode{}, httperr.New(httperr.TaskStateConflict, "invalid node type")
	}
	if input.MinPhotos < 0 {
		return database.InspectionTemplateNode{}, httperr.New(httperr.TaskStateConflict, "min_photos must be non-negative")
	}

	isMandatory := input.IsMandatory
	if isMandatory == "" {
		isMandatory = "1"
	}
	isRequired := input.IsRequired
	if isRequired == "" {
		isRequired = "1"
	}
	requireLiveCapture := input.RequireLiveCapture || input.MinPhotos > 0

	node := database.InspectionTemplateNode{
		TemplateID:         nil, // 新创建的节点未分配
		SortOrder:          0,
		Name:               input.Name,
		Description:        input.Description,
		NodeDesc:           input.NodeDesc,
		NodeType:           input.NodeType,
		MinPhotos:          input.MinPhotos,
		RequireText:        input.RequireText,
		AllowAbnormal:      input.AllowAbnormal,
		RequireLiveCapture: requireLiveCapture,
		NodesConfigID:      input.NodesConfigID,
		TaskTypeID:         input.TaskTypeID,
		IsMandatory:        isMandatory,
		IsRequired:         isRequired,
		AlgorithmID:        input.AlgorithmID,
		QueryID:            input.QueryID,
		TimeoutSecond:      input.TimeoutSecond,
		Remark:             input.Remark,
	}
	return node, s.db.Create(&node).Error
}

// Update 更新节点信息，允许重复分配，不限制已分配节点的编辑。
func (s *Service) Update(id uint64, input UpdateInput) (database.InspectionTemplateNode, error) {
	var node database.InspectionTemplateNode
	if err := s.db.First(&node, id).Error; err != nil {
		return node, notFound(err, "node not found")
	}
	if input.Name == "" {
		return database.InspectionTemplateNode{}, httperr.New(httperr.TaskStateConflict, "node name is required")
	}
	if !validNodeType(input.NodeType) {
		return database.InspectionTemplateNode{}, httperr.New(httperr.TaskStateConflict, "invalid node type")
	}

	isMandatory := input.IsMandatory
	if isMandatory == "" {
		isMandatory = "1"
	}
	isRequired := input.IsRequired
	if isRequired == "" {
		isRequired = "1"
	}
	requireLiveCapture := input.RequireLiveCapture || input.MinPhotos > 0

	node.Name = input.Name
	node.Description = input.Description
	node.NodeDesc = input.NodeDesc
	node.NodeType = input.NodeType
	node.MinPhotos = input.MinPhotos
	node.RequireText = input.RequireText
	node.AllowAbnormal = input.AllowAbnormal
	node.RequireLiveCapture = requireLiveCapture
	node.NodesConfigID = input.NodesConfigID
	node.TaskTypeID = input.TaskTypeID
	node.IsMandatory = isMandatory
	node.IsRequired = isRequired
	node.AlgorithmID = input.AlgorithmID
	node.QueryID = input.QueryID
	node.TimeoutSecond = input.TimeoutSecond
	node.Remark = input.Remark

	return node, s.db.Save(&node).Error
}

// Delete 删除节点，允许重复分配，不限制已分配节点的删除。
func (s *Service) Delete(id uint64) error {
	var node database.InspectionTemplateNode
	if err := s.db.First(&node, id).Error; err != nil {
		return notFound(err, "node not found")
	}
	return s.db.Delete(&node).Error
}

// AssignToTemplate 批量将节点分配到指定模板。
func (s *Service) AssignToTemplate(nodeIDs []uint64, templateID uint64) error {
	if len(nodeIDs) == 0 {
		return nil
	}
	// 验证模板存在
	var template database.InspectionTemplate
	if err := s.db.First(&template, templateID).Error; err != nil {
		return notFound(err, "template not found")
	}
	// 允许重复分配，不限制已分配节点
	return s.db.Model(&database.InspectionTemplateNode{}).
		Where("id IN ?", nodeIDs).
		Update("template_id", templateID).Error
}

// UnassignFromTemplate 批量将节点从模板中释放（设为未分配）。
func (s *Service) UnassignFromTemplate(nodeIDs []uint64) error {
	if len(nodeIDs) == 0 {
		return nil
	}
	return s.db.Model(&database.InspectionTemplateNode{}).
		Where("id IN ?", nodeIDs).
		Update("template_id", nil).Error
}

// ListByTemplate 查询指定模板下的节点列表。
func (s *Service) ListByTemplate(templateID uint64) ([]database.InspectionTemplateNode, error) {
	var nodes []database.InspectionTemplateNode
	err := s.db.Where("template_id = ?", templateID).Order("sort_order asc, id asc").Find(&nodes).Error
	return nodes, err
}

// ListUnassigned 查询所有未分配的节点（供模板选择）。
func (s *Service) ListUnassigned() ([]database.InspectionTemplateNode, error) {
	var nodes []database.InspectionTemplateNode
	err := s.db.Where("template_id IS NULL").Order("id desc").Find(&nodes).Error
	return nodes, err
}

// validNodeType 判断节点类型是否属于系统允许范围。
func validNodeType(nodeType string) bool {
	switch nodeType {
	case NodeText, NodeRead, NodeCheck, NodePhoto, NodeVideo, NodeAudio:
		return true
	default:
		return false
	}
}

// notFound 将 gorm.ErrRecordNotFound 转换为 httperr.ResourceNotFound。
func notFound(err error, message string) error {
	if err == gorm.ErrRecordNotFound {
		return httperr.New(httperr.ResourceNotFound, message)
	}
	return err
}

// escapeLike 转义 SQL LIKE 通配符。
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
