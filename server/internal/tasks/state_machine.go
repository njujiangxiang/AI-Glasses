// Package tasks 保存后台操作、眼镜执行和调度维护共用的巡检任务状态规则。将状态迁移判断
// 集中在这里，可以避免领取、开始、提交、取消和逾期等行为在不同入口产生分歧。
package tasks

import "aiglasses/server/internal/platform/httperr"

const (
	StatusPending    = "pending"
	StatusAssigned   = "assigned"
	StatusInProgress = "in_progress"
	StatusSubmitted  = "submitted"
	StatusCompleted  = "completed"
	StatusOverdue    = "overdue"
	StatusCancelled  = "cancelled"

	NodePending   = "pending"
	NodeCompleted = "completed"
	NodeSkipped   = "skipped"
	NodeAbnormal  = "abnormal"
)

// CanClaim 判断当前任务状态是否允许班组成员领取。
func CanClaim(status string) bool { return status == StatusPending }

// CanStart 判断当前任务状态是否允许巡检员开始执行。
func CanStart(status string) bool { return status == StatusAssigned }

// CanSubmitNode 判断当前任务状态是否允许提交节点结果。
func CanSubmitNode(status string) bool { return status == StatusInProgress || status == StatusOverdue }

// CanSubmitTask 判断当前任务状态是否允许提交整单巡检结果。
func CanSubmitTask(status string) bool { return status == StatusInProgress || status == StatusOverdue }

// CanComplete 判断当前任务状态是否允许后台确认完成。
func CanComplete(status string) bool { return status == StatusSubmitted }

// CanCancel 判断当前任务状态是否允许后台取消。
func CanCancel(status string) bool {
	return status == StatusPending || status == StatusAssigned || status == StatusInProgress
}

// CanOverdue 判断当前任务状态是否允许被调度器标记为逾期。
func CanOverdue(status string) bool { return status == StatusAssigned || status == StatusInProgress }

// Ensure 将状态判断结果转换为统一状态冲突错误。
func Ensure(ok bool) error {
	if !ok {
		return httperr.New(httperr.TaskStateConflict, "task state transition is not allowed")
	}
	return nil
}
