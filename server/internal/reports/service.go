// Package reports 提供巡检报告查询与 PDF 生成能力。报告以已完成的巡检任务为基础，
// 聚合任务、节点、执行结果、缺陷和附件信息，供后台管理端展示和导出。
package reports

import (
	"strings"
	"time"

	"aiglasses/server/internal/platform/database"
	"gorm.io/gorm"
)

// Service 提供巡检报告相关的数据查询与 PDF 生成能力。
type Service struct {
	db *gorm.DB
}

// NewService 创建报告服务。
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ReportListItem 报告列表中的一条记录。
type ReportListItem struct {
	database.InspectionTask
	TemplateName  string `json:"template_name"`
	AssigneeName  string `json:"assignee_name"`
	ExecutorName  string `json:"executor_name"`
	NodeCount     int    `json:"node_count"`
}

// ListQuery 报告列表查询条件。
type ListQuery struct {
	Keyword  string
	Page     int
	PageSize int
}

// ListResult 报告列表分页结果。
type ListResult struct {
	Items []ReportListItem `json:"items"`
	Total int64            `json:"total"`
}

// List 查询已完成任务的报告列表，支持关键词搜索和分页。
func (s *Service) List(q ListQuery) (ListResult, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 20
	}

	db := s.db.Model(&database.InspectionTask{}).Where("status = ?", "completed")
	if keyword := strings.TrimSpace(q.Keyword); keyword != "" {
		like := "%" + escapeLike(keyword) + "%"
		db = db.Where("task_name LIKE ? OR point_name LIKE ? OR equipment_name LIKE ?", like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return ListResult{}, err
	}

	var tasks []database.InspectionTask
	if err := db.Order("completed_at DESC, id DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&tasks).Error; err != nil {
		return ListResult{}, err
	}

	// 批量查询模板名称
	templateIDs := make(map[uint64]bool)
	userIDs := make(map[uint64]bool)
	nodeCounts := make(map[uint64]int)

	for _, t := range tasks {
		templateIDs[t.TemplateID] = true
		if t.AssigneeID > 0 {
			userIDs[t.AssigneeID] = true
		}
		if t.ExecutorID != nil && *t.ExecutorID > 0 {
			userIDs[*t.ExecutorID] = true
		}
	}

	// 批量查询节点数
	if len(tasks) > 0 {
		taskIDs := make([]uint64, len(tasks))
		for i, t := range tasks {
			taskIDs[i] = t.ID
		}
		type countRow struct {
			TaskID uint64 `gorm:"column:task_id"`
			Count  int    `gorm:"column:count"`
		}
		var counts []countRow
		s.db.Table("inspection_task_nodes").
			Select("task_id, COUNT(*) as count").
			Where("task_id IN ?", taskIDs).
			Group("task_id").
			Find(&counts)
		for _, c := range counts {
			nodeCounts[c.TaskID] = c.Count
		}
	}

	// 批量查询用户名
	userNames := make(map[uint64]string)
	if len(userIDs) > 0 {
		ids := make([]uint64, 0, len(userIDs))
		for id := range userIDs {
			ids = append(ids, id)
		}
		var users []database.User
		s.db.Where("id IN ?", ids).Find(&users)
		for _, u := range users {
			name := u.DisplayName
			if name == "" {
				name = u.Name
			}
			if name == "" {
				name = u.Username
			}
			userNames[u.ID] = name
		}
	}

	// 批量查询模板名称
	templateNames := make(map[uint64]string)
	if len(templateIDs) > 0 {
		ids := make([]uint64, 0, len(templateIDs))
		for id := range templateIDs {
			ids = append(ids, id)
		}
		var templates []database.InspectionTemplate
		s.db.Where("id IN ?", ids).Find(&templates)
		for _, t := range templates {
			templateNames[t.ID] = t.Name
		}
	}

	items := make([]ReportListItem, len(tasks))
	for i, t := range tasks {
		items[i] = ReportListItem{
			InspectionTask: t,
			TemplateName:   templateNames[t.TemplateID],
			AssigneeName:   userNames[t.AssigneeID],
			ExecutorName:   userNames[ptrVal(t.ExecutorID)],
			NodeCount:      nodeCounts[t.ID],
		}
	}

	return ListResult{Items: items, Total: total}, nil
}

// NodeDefectInfo 节点关联的缺陷信息。
type NodeDefectInfo struct {
	ID          uint64     `json:"id"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CloseReason string     `json:"close_reason"`
	CreatedAt   time.Time  `json:"created_at"`
	ConfirmedAt *time.Time `json:"confirmed_at"`
	ClosedAt    *time.Time `json:"closed_at"`
}

// NodeAttachmentInfo 节点关联的附件信息。
type NodeAttachmentInfo struct {
	ID          uint64 `json:"id"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

// NodeResultInfo 节点执行结果。
type NodeResultInfo struct {
	ID              uint64             `json:"id"`
	Status          string             `json:"status"`
	FeedbackContent string             `json:"feedback_content"`
	TextNote        string             `json:"text_note"`
	AlgorithmResult string             `json:"algorithm_result"`
	QueryResult     string             `json:"query_result"`
	IsAbnormal      bool               `json:"is_abnormal"`
	AbnormalDesc    string             `json:"abnormal_desc"`
	CompletedAt     time.Time          `json:"completed_at"`
	Attachments     []NodeAttachmentInfo `json:"attachments"`
}

// ReportNode 报告中的单个节点详情。
type ReportNode struct {
	ID             uint64          `json:"id"`
	SortOrder      int             `json:"sort_order"`
	Name           string          `json:"name"`
	NodeType       string          `json:"node_type"`
	Status         string          `json:"status"`
	IsMandatory    bool            `json:"is_mandatory"`
	IsRequired     bool            `json:"is_required"`
	ActualExecTime *time.Time      `json:"actual_exec_time"`
	Result         *NodeResultInfo `json:"result"`
	Defects        []NodeDefectInfo `json:"defects"`
}

// ReportDetail 报告详情，包含任务信息、节点执行结果和缺陷。
type ReportDetail struct {
	Task         database.InspectionTask `json:"task"`
	TemplateName string                  `json:"template_name"`
	AssigneeName string                  `json:"assignee_name"`
	ExecutorName string                  `json:"executor_name"`
	Nodes        []ReportNode            `json:"nodes"`
	Defects      []NodeDefectInfo        `json:"defects"`
	NodeCount    int                     `json:"node_count"`
}

// Detail 查询单个已完成任务的报告详情。
func (s *Service) Detail(taskID uint64) (ReportDetail, error) {
	var task database.InspectionTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ReportDetail{}, err
		}
		return ReportDetail{}, err
	}

	// 查询节点
	var nodes []database.InspectionTaskNode
	if err := s.db.Where("task_id = ?", taskID).Order("sort_order ASC").Find(&nodes).Error; err != nil {
		return ReportDetail{}, err
	}

	// 查询结果
	var results []database.TaskNodeResult
	if err := s.db.Where("task_id = ?", taskID).Find(&results).Error; err != nil {
		return ReportDetail{}, err
	}

	// 查询缺陷
	var defects []database.Defect
	if err := s.db.Where("task_id = ?", taskID).Order("id ASC").Find(&defects).Error; err != nil {
		return ReportDetail{}, err
	}

	// 查询附件
	var attachments []database.Attachment
	if err := s.db.Where("task_id = ?", taskID).Find(&attachments).Error; err != nil {
		return ReportDetail{}, err
	}

	// 构建结果索引 (node_id -> result)
	resultByNode := make(map[uint64]*database.TaskNodeResult)
	for i := range results {
		resultByNode[results[i].NodeID] = &results[i]
	}

	// 构建附件索引 (node_id -> attachments)
	attachmentsByNode := make(map[uint64][]database.Attachment)
	for _, a := range attachments {
		if a.NodeID != nil {
			attachmentsByNode[*a.NodeID] = append(attachmentsByNode[*a.NodeID], a)
		}
	}

	// 构建缺陷索引 (node_id -> defects)
	defectsByNode := make(map[uint64][]database.Defect)
	for _, d := range defects {
		defectsByNode[d.NodeID] = append(defectsByNode[d.NodeID], d)
	}

	// 组装节点详情
	reportNodes := make([]ReportNode, len(nodes))
	for i, n := range nodes {
		rn := ReportNode{
			ID:             n.ID,
			SortOrder:      n.SortOrder,
			Name:           n.Name,
			NodeType:       n.NodeType,
			Status:         n.Status,
			IsMandatory:    n.IsMandatory,
			IsRequired:     n.IsRequired,
			ActualExecTime: n.ActualExecTime,
		}

		// 挂载结果
		if r, ok := resultByNode[n.ID]; ok {
			atts := make([]NodeAttachmentInfo, 0)
			for _, a := range attachmentsByNode[n.ID] {
				atts = append(atts, NodeAttachmentInfo{
					ID:          a.ID,
					FileName:    a.FileName,
					ContentType: a.ContentType,
					SizeBytes:   a.SizeBytes,
				})
			}
			rn.Result = &NodeResultInfo{
				ID:              r.ID,
				Status:          r.Status,
				FeedbackContent: r.FeedbackContent,
				TextNote:        r.TextNote,
				AlgorithmResult: r.AlgorithmResult,
				QueryResult:     r.QueryResult,
				IsAbnormal:      r.IsAbnormal,
				AbnormalDesc:    r.AbnormalDesc,
				CompletedAt:     r.CompletedAt,
				Attachments:     atts,
			}
		}

		// 挂载缺陷
		nodeDefects := make([]NodeDefectInfo, 0)
		for _, d := range defectsByNode[n.ID] {
			nodeDefects = append(nodeDefects, NodeDefectInfo{
				ID:          d.ID,
				Description: d.Description,
				Status:      d.Status,
				CloseReason: d.CloseReason,
				CreatedAt:   d.CreatedAt,
				ConfirmedAt: d.ConfirmedAt,
				ClosedAt:    d.ClosedAt,
			})
		}
		rn.Defects = nodeDefects
		reportNodes[i] = rn
	}

	// 汇总所有缺陷
	allDefects := make([]NodeDefectInfo, 0, len(defects))
	for _, d := range defects {
		allDefects = append(allDefects, NodeDefectInfo{
			ID:          d.ID,
			Description: d.Description,
			Status:      d.Status,
			CloseReason: d.CloseReason,
			CreatedAt:   d.CreatedAt,
			ConfirmedAt: d.ConfirmedAt,
			ClosedAt:    d.ClosedAt,
		})
	}

	// 查询模板名称
	templateName := ""
	var tmpl database.InspectionTemplate
	if err := s.db.First(&tmpl, task.TemplateID).Error; err == nil {
		templateName = tmpl.Name
	}

	// 查询用户名
	assigneeName := s.userName(task.AssigneeID)
	executorName := s.userName(ptrVal(task.ExecutorID))

	return ReportDetail{
		Task:         task,
		TemplateName: templateName,
		AssigneeName: assigneeName,
		ExecutorName: executorName,
		Nodes:        reportNodes,
		Defects:      allDefects,
		NodeCount:    len(nodes),
	}, nil
}

// userName 根据用户 ID 查询显示名称。
func (s *Service) userName(userID uint64) string {
	if userID == 0 {
		return ""
	}
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return ""
	}
	if user.DisplayName != "" {
		return user.DisplayName
	}
	if user.Name != "" {
		return user.Name
	}
	return user.Username
}

// ptrVal 安全地解引用 uint64 指针。
func ptrVal(p *uint64) uint64 {
	if p == nil {
		return 0
	}
	return *p
}

// escapeLike 转义 LIKE 查询中的特殊字符。
func escapeLike(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `%`, `\%`)
	v = strings.ReplaceAll(v, `_`, `\_`)
	return v
}
