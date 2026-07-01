// Package tasks 实现基于数据库的任务执行流程。服务在状态变更操作中使用行级锁，避免
// 眼镜端提交、后台操作和调度更新之间发生静默覆盖。
package tasks

import (
	"strings"
	"time"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	AssigneeUser = "user"
	AssigneeTeam = "team"
)

type AdminCreateInput struct {
	TemplateID    uint64     `json:"template_id"`
	TaskCode      string     `json:"task_code"`
	TaskName      string     `json:"task_name"`
	PointName     string     `json:"point_name"`
	EquipmentName string     `json:"equipment_name"`
	InspectArea   string     `json:"inspect_area"`
	AssigneeType  string     `json:"assignee_type"`
	AssigneeID    uint64     `json:"assignee_id"`
	DueAt         time.Time  `json:"due_at"`
	GlassesSN     string     `json:"glasses_sn"`
	AssignUser    string     `json:"assign_user"`
}

type AdminListQuery struct {
	Keyword    string
	Status     string
	TemplateID uint64
	Page       int
	PageSize   int
}

type TaskListItem struct {
	database.InspectionTask
	TemplateName string `json:"template_name"`
}

type AdminListResult struct {
	Items []TaskListItem `json:"items"`
	Total int64          `json:"total"`
}

type Service struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewService 创建任务服务，注入数据库和 Redis 客户端。
func NewService(db *gorm.DB, redisClient *redis.Client) *Service {
	return &Service{db: db, redis: redisClient}
}

// AdminCreate 后台手动创建巡检任务，不依赖计划。
func (s *Service) AdminCreate(input AdminCreateInput) (database.InspectionTask, error) {
	if input.TemplateID == 0 {
		return database.InspectionTask{}, httperr.New(httperr.TaskStateConflict, "template_id is required")
	}
	if input.TaskName == "" {
		return database.InspectionTask{}, httperr.New(httperr.TaskStateConflict, "task_name is required")
	}
	if input.AssigneeType != AssigneeUser && input.AssigneeType != AssigneeTeam {
		return database.InspectionTask{}, httperr.New(httperr.TaskStateConflict, "invalid assignee type")
	}
	if input.AssigneeID == 0 {
		return database.InspectionTask{}, httperr.New(httperr.TaskStateConflict, "assignee_id is required")
	}
	if input.DueAt.IsZero() {
		return database.InspectionTask{}, httperr.New(httperr.TaskStateConflict, "due_at is required")
	}

	// 验证模板存在
	var template database.InspectionTemplate
	if err := s.db.First(&template, input.TemplateID).Error; err != nil {
		return database.InspectionTask{}, httperr.New(httperr.ResourceNotFound, "template not found")
	}

	// 查询模板节点
	var templateNodes []database.InspectionTemplateNode
	if err := s.db.Where("template_id = ?", input.TemplateID).Order("sort_order asc").Find(&templateNodes).Error; err != nil {
		return database.InspectionTask{}, err
	}
	if len(templateNodes) == 0 {
		return database.InspectionTask{}, httperr.New(httperr.TaskStateConflict, "template has no nodes")
	}

	task := database.InspectionTask{
		PlanID:        nil, // 手动创建，无计划
		TaskCode:      input.TaskCode,
		TemplateID:    input.TemplateID,
		ScheduledAt:   nil, // 手动创建，无计划时间
		DueAt:         input.DueAt.UTC(),
		Status:        StatusPending,
		AssigneeType:  input.AssigneeType,
		AssigneeID:    input.AssigneeID,
		PointName:     input.PointName,
		EquipmentName: input.EquipmentName,
		TaskName:      input.TaskName,
		InspectArea:   input.InspectArea,
		GlassesSN:     input.GlassesSN,
		AssignUser:    input.AssignUser,
		AssignTime:    timePtr(time.Now().UTC()),
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		// 从模板复制节点
		for _, tn := range templateNodes {
			taskNode := database.InspectionTaskNode{
				TaskID:         task.ID,
				TemplateNodeID: tn.ID,
				SortOrder:      tn.SortOrder,
				Name:           tn.Name,
				NodeType:       tn.NodeType,
				MinPhotos:      tn.MinPhotos,
				RequireText:    tn.RequireText,
				AllowAbnormal:  tn.AllowAbnormal,
				Status:         NodePending,
				NodesConfigID:  tn.NodesConfigID,
				TaskTypeCode:   tn.TaskTypeID,
				IsMandatory:    tn.IsMandatory,
				IsRequired:     tn.IsRequired,
				AlgorithmID:    tn.AlgorithmID,
				QueryID:        tn.QueryID,
				Remark:         tn.Remark,
			}
			if err := tx.Create(&taskNode).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return task, err
}

// AdminList 按条件查询后台任务列表，支持分页。
func (s *Service) AdminList(query AdminListQuery) (AdminListResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	db := s.db.Model(&database.InspectionTask{})
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + escapeLike(keyword) + "%"
		db = db.Where("task_name LIKE ? OR point_name LIKE ? OR equipment_name LIKE ?", like, like, like)
	}
	if status := strings.TrimSpace(query.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	if query.TemplateID > 0 {
		db = db.Where("template_id = ?", query.TemplateID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return AdminListResult{}, err
	}

	var tasks []database.InspectionTask
	if err := db.Order("due_at asc, id asc").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&tasks).Error; err != nil {
		return AdminListResult{}, err
	}

	// 批量查询模板名称
	templateIDs := make(map[uint64]bool)
	for _, t := range tasks {
		templateIDs[t.TemplateID] = true
	}
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

	items := make([]TaskListItem, len(tasks))
	for i, t := range tasks {
		items[i] = TaskListItem{
			InspectionTask: t,
			TemplateName:   templateNames[t.TemplateID],
		}
	}

	return AdminListResult{Items: items, Total: total}, nil
}

// GlassesList 查询指定眼镜用户可见的个人任务和班组任务。
func (s *Service) GlassesList(userID uint64, teamIDs []uint64, limit int) ([]database.InspectionTask, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	visibility := s.visibleQuery(userID, "")
	if len(teamIDs) > 0 {
		visibility = visibility.Or("assignee_type = ? AND assignee_id IN ? AND executor_id IS NULL", "team", teamIDs)
	}
	var tasks []database.InspectionTask
	return tasks, visibility.Order("due_at asc, id asc").Limit(limit).Find(&tasks).Error
}

// GlassesPage 分页查询指定眼镜用户可见的任务，支持文档接口中的状态过滤。
func (s *Service) visibleQuery(userID uint64, status string) *gorm.DB {
	statuses := []string{StatusPending, StatusAssigned, StatusInProgress, StatusOverdue}
	if status != "" {
		statuses = []string{status}
	}
	visibility := s.db.Where("assignee_type = ? AND assignee_id = ?", "user", userID).Or("executor_id = ?", userID).Or("assignee_type = ? AND executor_id IS NULL", "team")
	return s.db.Where("status IN ?", statuses).Where(visibility)
}

func (s *Service) GlassesPage(userID uint64, status string, page, pageSize int) ([]database.InspectionTask, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 500 {
		pageSize = 50
	}
	query := s.visibleQuery(userID, status)
	var total int64
	if err := query.Model(&database.InspectionTask{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tasks []database.InspectionTask
	err := query.Order("due_at asc, id asc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&tasks).Error
	return tasks, total, err
}

// CurrentNode 查询指定任务当前应执行的节点，优先返回待执行节点。
func (s *Service) CurrentNode(taskID, userID uint64) (database.InspectionTaskNode, error) {
	if err := s.EnsureAccessible(taskID, userID); err != nil {
		return database.InspectionTaskNode{}, err
	}
	var node database.InspectionTaskNode
	if err := s.db.Where("task_id = ? AND status = ?", taskID, NodePending).Order("sort_order asc, id asc").First(&node).Error; err == nil {
		return node, nil
	} else if err != gorm.ErrRecordNotFound {
		return database.InspectionTaskNode{}, err
	}
	if err := s.db.Where("task_id = ?", taskID).Order("sort_order asc, id asc").First(&node).Error; err != nil {
		return database.InspectionTaskNode{}, err
	}
	return node, nil
}

// EnsureAccessible 校验任务是否对当前眼镜用户可见。
func (s *Service) EnsureAccessible(taskID, userID uint64) error {
	var task database.InspectionTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return err
	}
	if !ownsTask(task, userID) && !(task.AssigneeType == "team" && task.ExecutorID == nil) {
		return httperr.New(httperr.TaskNotAssigned, "task is not assigned to user")
	}
	return nil
}

// AcceptNodeProgress 校验节点进度上报。当前版本不持久化节点内临时进度。
func (s *Service) AcceptNodeProgress(taskID, nodeID, userID uint64, progress float64) error {
	if progress < 0 || progress > 1 {
		return httperr.New(httperr.ValidationFailed, "progress must be between 0 and 1")
	}
	if err := s.EnsureAccessible(taskID, userID); err != nil {
		return err
	}
	var count int64
	if err := s.db.Model(&database.InspectionTaskNode{}).Where("task_id = ? AND id = ?", taskID, nodeID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return httperr.New(httperr.ResourceNotFound, "task node not found")
	}
	return nil
}

// SkipNode 将当前用户可执行任务中的节点标记为跳过。
func (s *Service) SkipNode(taskID, nodeID, userID uint64, reason string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var task database.InspectionTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&task, taskID).Error; err != nil {
			return err
		}
		if !ownsTask(task, userID) {
			return httperr.New(httperr.TaskNotAssigned, "task is not assigned to user")
		}
		if err := Ensure(CanSubmitNode(task.Status)); err != nil {
			return err
		}
		var node database.InspectionTaskNode
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("task_id = ? AND id = ?", taskID, nodeID).First(&node).Error; err != nil {
			return err
		}
		if node.Status != NodePending {
			return httperr.New(httperr.TaskStateConflict, "node cannot be skipped from current status")
		}
		return tx.Model(&node).Update("status", NodeSkipped).Error
	})
}

// Detail 查询任务详情，包括节点、节点结果、附件和关联缺陷。
func (s *Service) Detail(taskID uint64) (database.InspectionTask, []database.InspectionTaskNode, []database.TaskNodeResult, []database.Attachment, []database.Defect, error) {
	var task database.InspectionTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return task, nil, nil, nil, nil, err
	}
	var nodes []database.InspectionTaskNode
	var results []database.TaskNodeResult
	var attachments []database.Attachment
	var defects []database.Defect
	if err := s.db.Where("task_id = ?", taskID).Order("sort_order asc").Find(&nodes).Error; err != nil {
		return task, nil, nil, nil, nil, err
	}
	if err := s.db.Where("task_id = ?", taskID).Find(&results).Error; err != nil {
		return task, nil, nil, nil, nil, err
	}
	if err := s.db.Where("task_id = ?", taskID).Find(&attachments).Error; err != nil {
		return task, nil, nil, nil, nil, err
	}
	if err := s.db.Where("task_id = ?", taskID).Find(&defects).Error; err != nil {
		return task, nil, nil, nil, nil, err
	}
	return task, nodes, results, attachments, defects, nil
}

// Claim 将班组待领取任务分配给当前巡检员。
func (s *Service) Claim(taskID, userID uint64) error {
	now := time.Now().UTC()
	return s.db.Transaction(func(tx *gorm.DB) error {
		var task database.InspectionTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&task, taskID).Error; err != nil {
			return err
		}
		if !CanClaim(task.Status) || task.ExecutorID != nil {
			return httperr.New(httperr.TaskAlreadyClaimed, "task already claimed")
		}
		return tx.Model(&task).Updates(map[string]any{"status": StatusAssigned, "executor_id": userID, "updated_at": now}).Error
	})
}

// Start 将已分配任务切换为执行中。
func (s *Service) Start(taskID, userID uint64) error {
	now := time.Now().UTC()
	return s.db.Transaction(func(tx *gorm.DB) error {
		var task database.InspectionTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&task, taskID).Error; err != nil {
			return err
		}
		if !ownsTask(task, userID) {
			return httperr.New(httperr.TaskNotAssigned, "task is not assigned to user")
		}
		if err := Ensure(CanStart(task.Status)); err != nil {
			return err
		}
		return tx.Model(&task).Updates(map[string]any{"status": StatusInProgress, "started_at": now}).Error
	})
}

type NodeResultInput struct {
	TaskTypeCode     string   `json:"task_type_code"`
	FeedbackContent  string   `json:"feedback_content"`
	TextNote         string   `json:"text_note"`
	LocationGPS      string   `json:"location_gps"`
	AttachmentIDs    []uint64 `json:"attachment_ids"`
	IsAbnormal       bool     `json:"is_abnormal"`
	AbnormalDesc     string   `json:"abnormal_desc"`
	Remark           string   `json:"remark"`
}

// SubmitNode 提交单个巡检节点结果，并通过幂等键避免弱网重复提交。
func (s *Service) SubmitNode(taskID, nodeID, userID uint64, idemKey string, input NodeResultInput) (database.TaskNodeResult, error) {
	if idemKey == "" {
		return database.TaskNodeResult{}, httperr.New(httperr.IdempotencyConflict, "idempotency key is required")
	}
	var result database.TaskNodeResult
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var task database.InspectionTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&task, taskID).Error; err != nil {
			return err
		}
		if !ownsTask(task, userID) {
			return httperr.New(httperr.TaskNotAssigned, "task is not assigned to user")
		}
		if err := Ensure(CanSubmitNode(task.Status)); err != nil {
			return err
		}
		var node database.InspectionTaskNode
		if err := tx.Where("task_id = ? AND id = ?", taskID, nodeID).First(&node).Error; err != nil {
			return err
		}
		if node.RequireText && input.TextNote == "" {
			return httperr.New(httperr.NodeRequiredTextMissing, "node text note is required")
		}
		if node.MinPhotos > 0 && len(input.AttachmentIDs) < node.MinPhotos {
			return httperr.New(httperr.NodeRequiredPhotoMissing, "required photos are missing")
		}
		if len(input.AttachmentIDs) > 0 {
			var count int64
			if err := tx.Model(&database.Attachment{}).Where("id IN ? AND bind_status = ?", input.AttachmentIDs, "uploaded").Count(&count).Error; err != nil {
				return err
			}
			if int(count) != len(input.AttachmentIDs) {
				return httperr.New(httperr.AttachmentNotUploaded, "attachment is not uploaded")
			}
		}
		status := NodeCompleted
		if input.IsAbnormal {
			status = NodeAbnormal
		}

		// 构建附件ID字符串
		var attachmentIDStr string
		if len(input.AttachmentIDs) > 0 {
			var idStrs []string
			for _, id := range input.AttachmentIDs {
				idStrs = append(idStrs, string(rune(id)))
			}
			attachmentIDStr = strings.Join(idStrs, ",")
		}

		result = database.TaskNodeResult{
			TaskID:           taskID,
			NodeID:           nodeID,
			UserID:           userID,
			Status:           status,
			TaskTypeCode:     input.TaskTypeCode,
			FeedbackContent:  input.FeedbackContent,
			TextNote:         input.TextNote,
			LocationGPS:      input.LocationGPS,
			AttachmentIDs:    attachmentIDStr,
			IsAbnormal:       input.IsAbnormal,
			AbnormalDesc:     input.AbnormalDesc,
			Remark:           input.Remark,
			IdempotencyKey:   idemKey,
			CompletedAt:      time.Now().UTC(),
		}
		if err := tx.Where(database.TaskNodeResult{TaskID: taskID, NodeID: nodeID}).Assign(result).FirstOrCreate(&result).Error; err != nil {
			return err
		}
		if len(input.AttachmentIDs) > 0 {
			if err := tx.Model(&database.Attachment{}).Where("id IN ?", input.AttachmentIDs).Updates(map[string]any{"bind_status": "bound", "task_id": taskID, "node_id": nodeID, "result_id": result.ID}).Error; err != nil {
				return err
			}
		}
		return tx.Model(&node).Updates(map[string]any{"status": status}).Error
	})
	return result, err
}

// SubmitTask 在所有必填节点满足要求后提交整单巡检任务。
func (s *Service) SubmitTask(taskID, userID uint64) error {
	now := time.Now().UTC()
	return s.db.Transaction(func(tx *gorm.DB) error {
		var task database.InspectionTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&task, taskID).Error; err != nil {
			return err
		}
		if !ownsTask(task, userID) {
			return httperr.New(httperr.TaskNotAssigned, "task is not assigned to user")
		}
		if err := Ensure(CanSubmitTask(task.Status)); err != nil {
			return err
		}
		var missing int64
		if err := tx.Model(&database.InspectionTaskNode{}).Where("task_id = ? AND status = ? AND is_required = ?", taskID, NodePending, "1").Count(&missing).Error; err != nil {
			return err
		}
		if missing > 0 {
			return httperr.New(httperr.NodeRequiredPhotoMissing, "required nodes are incomplete")
		}
		return tx.Model(&task).Updates(map[string]any{"status": StatusCompleted, "completed_at": now}).Error
	})
}

// Cancel 由后台取消尚未完成的任务。
func (s *Service) Cancel(taskID uint64) error {
	now := time.Now().UTC()
	result := s.db.Model(&database.InspectionTask{}).Where("id = ? AND status IN ?", taskID, []string{StatusPending, StatusAssigned, StatusInProgress}).Updates(map[string]any{"status": StatusCancelled, "cancelled_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.TaskStateConflict, "task cannot be cancelled")
	}
	return nil
}

// AdminDelete 后台删除巡检任务及其关联的节点、节点结果、附件和缺陷记录。
func (s *Service) AdminDelete(taskID uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var task database.InspectionTask
		if err := tx.First(&task, taskID).Error; err != nil {
			return err
		}
		// 仅允许删除未在执行中的任务（pending / assigned / completed / cancelled / overdue）
		if task.Status == StatusInProgress {
			return httperr.New(httperr.TaskStateConflict, "cannot delete an in-progress task")
		}
		if err := tx.Where("task_id = ?", taskID).Delete(&database.InspectionTaskNode{}).Error; err != nil {
			return err
		}
		if err := tx.Where("task_id = ?", taskID).Delete(&database.TaskNodeResult{}).Error; err != nil {
			return err
		}
		if err := tx.Where("task_id = ?", taskID).Delete(&database.Attachment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("task_id = ?", taskID).Delete(&database.Defect{}).Error; err != nil {
			return err
		}
		return tx.Delete(&task).Error
	})
}

// ownsTask 判断当前用户是否是任务执行人。
func ownsTask(task database.InspectionTask, userID uint64) bool {
	return (task.ExecutorID != nil && *task.ExecutorID == userID) || (task.AssigneeType == "user" && task.AssigneeID == userID)
}

func timePtr(t time.Time) *time.Time { return &t }

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
