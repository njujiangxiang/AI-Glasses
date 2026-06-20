// Package tasks 实现基于数据库的任务执行流程。服务在状态变更操作中使用行级锁，避免
// 眼镜端提交、后台操作和调度更新之间发生静默覆盖。
package tasks

import (
	"time"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewService 创建任务服务，注入数据库和 Redis 客户端。
func NewService(db *gorm.DB, redisClient *redis.Client) *Service {
	return &Service{db: db, redis: redisClient}
}

// AdminList 按状态查询后台任务列表，并限制最大返回数量。
func (s *Service) AdminList(status string, limit int) ([]database.InspectionTask, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := s.db.Order("due_at asc, id asc").Limit(limit)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var tasks []database.InspectionTask
	return tasks, query.Find(&tasks).Error
}

// AdminListWithScope 带数据范围过滤查询后台任务列表
// scope: 当前用户的数据范围信息
func (s *Service) AdminListWithScope(status string, limit int, scope interface{}) ([]database.InspectionTask, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := s.db.Model(&database.InspectionTask{}).Order("due_at asc, id asc").Limit(limit)

	// 应用数据范围过滤
	if scope != nil {
		// 尝试转换为 ScopeInfo 接口
		if scopeInfo, ok := scope.(interface {
			IsAll() bool
			IsSelfOnly() bool
			GetUserID() uint64
			GetOrgCodes() []string
		}); ok {
			if scopeInfo.IsAll() {
				// 全部数据 - 不限制
			} else if scopeInfo.IsSelfOnly() {
				// 仅自己创建或分配给自己的任务
				query = query.Where("executor_id = ? OR assignee_id = ?", scopeInfo.GetUserID(), scopeInfo.GetUserID())
			} else if len(scopeInfo.GetOrgCodes()) > 0 {
				// 组织范围 - 通过 JOIN users 表过滤
				// 任务通过 executor_id 关联用户，用户通过 org_code 关联组织
				query = query.Joins("LEFT JOIN users AS executor_users ON executor_users.id = inspection_tasks.executor_id").
					Where("executor_users.org_code IN ?", scopeInfo.GetOrgCodes())
			} else {
				// 无权限 - 返回空结果
				return []database.InspectionTask{}, nil
			}
		}
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var tasks []database.InspectionTask
	return tasks, query.Find(&tasks).Error
}

// GlassesList 查询指定眼镜用户可见的个人任务和班组任务。
func (s *Service) GlassesList(userID uint64, teamIDs []uint64, limit int) ([]database.InspectionTask, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	statuses := []string{StatusPending, StatusAssigned, StatusInProgress, StatusOverdue}
	visibility := s.db.Where("assignee_type = ? AND assignee_id = ?", "user", userID).Or("executor_id = ?", userID)
	if len(teamIDs) > 0 {
		visibility = visibility.Or("assignee_type = ? AND assignee_id IN ? AND executor_id IS NULL", "team", teamIDs)
	}
	var tasks []database.InspectionTask
	return tasks, s.db.Where("status IN ?", statuses).Where(visibility).Order("due_at asc, id asc").Limit(limit).Find(&tasks).Error
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
	TextNote      string   `json:"text_note"`
	AttachmentIDs []uint64 `json:"attachment_ids"`
	Abnormal      bool     `json:"abnormal"`
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
		if input.Abnormal {
			status = NodeAbnormal
		}
		result = database.TaskNodeResult{TaskID: taskID, NodeID: nodeID, UserID: userID, Status: status, TextNote: input.TextNote, IdempotencyKey: idemKey, CompletedAt: time.Now().UTC()}
		if err := tx.Where(database.TaskNodeResult{TaskID: taskID, NodeID: nodeID}).Assign(result).FirstOrCreate(&result).Error; err != nil {
			return err
		}
		if err := tx.Model(&database.Attachment{}).Where("id IN ?", input.AttachmentIDs).Updates(map[string]any{"bind_status": "bound", "task_id": taskID, "node_id": nodeID, "result_id": result.ID}).Error; err != nil {
			return err
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
		if err := tx.Model(&database.InspectionTaskNode{}).Where("task_id = ? AND status = ? AND (min_photos > 0 OR require_text = ?)", taskID, NodePending, true).Count(&missing).Error; err != nil {
			return err
		}
		if missing > 0 {
			return httperr.New(httperr.NodeRequiredPhotoMissing, "required nodes are incomplete")
		}
		return tx.Model(&task).Updates(map[string]any{"status": StatusSubmitted, "submitted_at": now}).Error
	})
}

// Complete 由后台确认已提交任务完成。
func (s *Service) Complete(taskID uint64) error {
	now := time.Now().UTC()
	result := s.db.Model(&database.InspectionTask{}).Where("id = ? AND status = ?", taskID, StatusSubmitted).Updates(map[string]any{"status": StatusCompleted, "completed_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.TaskStateConflict, "task cannot be completed")
	}
	return nil
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

// ownsTask 判断当前用户是否是任务执行人。
func ownsTask(task database.InspectionTask, userID uint64) bool {
	return (task.ExecutorID != nil && *task.ExecutorID == userID) || (task.AssigneeType == "user" && task.AssigneeID == userID)
}
