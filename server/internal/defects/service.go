// Package defects 实现轻量级异常结果处理流程。缺陷从待确认流转到已确认或已关闭，并通过
// 关闭原因保存管理员判断，但不引入完整维修工单系统。
package defects

import (
	"time"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"gorm.io/gorm"
)

const (
	StatusPendingConfirm = "pending_confirm"
	StatusConfirmed      = "confirmed"
	StatusClosed         = "closed"
)

type Service struct {
	db *gorm.DB
}

// NewService 创建缺陷服务，注入数据库访问能力。
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// Create 根据异常节点结果创建待确认缺陷。
func (s *Service) Create(taskID, nodeID, reporterID uint64, description string) (database.Defect, error) {
	defect := database.Defect{TaskID: taskID, NodeID: nodeID, ReporterID: reporterID, Description: description, Status: StatusPendingConfirm}
	return defect, s.db.Create(&defect).Error
}

// Confirm 将待确认缺陷标记为已确认。
func (s *Service) Confirm(id uint64) error {
	now := time.Now().UTC()
	result := s.db.Model(&database.Defect{}).Where("id = ? AND status = ?", id, StatusPendingConfirm).Updates(map[string]any{
		"status":       StatusConfirmed,
		"confirmed_at": now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.TaskStateConflict, "defect cannot be confirmed from current status")
	}
	return nil
}

// Close 关闭缺陷并记录关闭原因。
func (s *Service) Close(id uint64, reason string) error {
	now := time.Now().UTC()
	result := s.db.Model(&database.Defect{}).Where("id = ? AND status IN ?", id, []string{StatusPendingConfirm, StatusConfirmed}).Updates(map[string]any{
		"status":       StatusClosed,
		"close_reason": reason,
		"closed_at":    now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.TaskStateConflict, "defect cannot be closed from current status")
	}
	return nil
}

// List 按状态查询缺陷列表，限制返回数量用于后台分页/列表展示。
func (s *Service) List(status string, limit int) ([]database.Defect, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := s.db.Order("created_at desc, id desc").Limit(limit)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var defects []database.Defect
	return defects, query.Find(&defects).Error
}
