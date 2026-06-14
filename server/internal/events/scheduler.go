// Package events 提供后台任务基础设施，用于把启用的巡检计划转换为具体巡检任务。调度器
// 依赖数据库唯一约束和有界回看窗口，避免重复 tick 或短暂宕机后生成重复任务。
package events

import (
	"encoding/json"
	"fmt"
	"time"

	"aiglasses/server/internal/config"
	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/tasks"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Scheduler struct {
	db  *gorm.DB
	cfg config.Config
}

// NewScheduler 创建巡检计划调度器。
func NewScheduler(db *gorm.DB, cfg config.Config) *Scheduler { return &Scheduler{db: db, cfg: cfg} }

// Tick 扫描启用计划并尝试生成当前时间窗口内应生成的任务。
func (s *Scheduler) Tick(now time.Time) error {
	var plans []database.TaskPlan
	if err := s.db.Where("enabled = ? AND start_at <= ?", true, now.UTC()).Find(&plans).Error; err != nil {
		return err
	}
	for _, plan := range plans {
		if err := s.generateForPlan(plan, now.UTC()); err != nil {
			return err
		}
	}
	return nil
}

// generateForPlan 根据单个计划的 cron 和时区计算本次应生成的任务时间点。
func (s *Scheduler) generateForPlan(plan database.TaskPlan, now time.Time) error {
	loc, err := time.LoadLocation(plan.Timezone)
	if err != nil {
		return err
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(plan.CronExpr)
	if err != nil {
		return err
	}
	windowStart := now.Add(-s.cfg.SchedulerLookback)
	candidate := schedule.Next(windowStart.In(loc)).UTC()
	for !candidate.After(now) {
		if candidate.Before(plan.StartAt.UTC()) {
			candidate = schedule.Next(candidate.In(loc)).UTC()
			continue
		}
		if err := s.insertTask(plan, candidate); err != nil {
			return err
		}
		candidate = schedule.Next(candidate.In(loc)).UTC()
	}
	return nil
}

// insertTask 插入巡检任务、任务节点和 outbox 事件，依赖唯一索引保证幂等。
func (s *Scheduler) insertTask(plan database.TaskPlan, scheduledAt time.Time) error {
	dueAt := scheduledAt.Add(time.Duration(plan.DueDurationMinutes) * time.Minute)
	return s.db.Transaction(func(tx *gorm.DB) error {
		task := database.InspectionTask{
			PlanID:        plan.ID,
			TemplateID:    plan.TemplateID,
			ScheduledAt:   scheduledAt,
			DueAt:         dueAt,
			Status:        initialStatus(plan.AssigneeType),
			AssigneeType:  plan.AssigneeType,
			AssigneeID:    plan.AssigneeID,
			PointName:     plan.PointName,
			EquipmentName: plan.EquipmentName,
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&task)
		if result.Error != nil || result.RowsAffected == 0 {
			return result.Error
		}
		var templateNodes []database.InspectionTemplateNode
		if err := tx.Where("template_id = ?", plan.TemplateID).Order("sort_order asc").Find(&templateNodes).Error; err != nil {
			return err
		}
		for _, node := range templateNodes {
			model := database.InspectionTaskNode{TaskID: task.ID, TemplateNodeID: node.ID, SortOrder: node.SortOrder, Name: node.Name, NodeType: node.NodeType, MinPhotos: node.MinPhotos, RequireText: node.RequireText, AllowAbnormal: node.AllowAbnormal, Status: tasks.NodePending}
			if err := tx.Create(&model).Error; err != nil {
				return err
			}
		}
		payload, err := json.Marshal(map[string]any{"task_id": task.ID, "plan_id": plan.ID, "scheduled_at": scheduledAt.Format(time.RFC3339)})
		if err != nil {
			return err
		}
		outbox := database.OutboxEvent{EventKey: fmt.Sprintf("task.assigned:%d", task.ID), Topic: "task.assigned", Payload: string(payload)}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&outbox).Error
	})
}

// initialStatus 根据任务指派类型计算生成后的初始状态。
func initialStatus(assigneeType string) string {
	if assigneeType == "team" {
		return tasks.StatusPending
	}
	return tasks.StatusAssigned
}
