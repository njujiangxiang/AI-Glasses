package tasks

import "testing"

// TestTaskStateTransitions 验证任务状态机的核心允许/禁止迁移规则。
func TestTaskStateTransitions(t *testing.T) {
	if !CanClaim(StatusPending) || CanClaim(StatusAssigned) {
		t.Fatal("claim transition mismatch")
	}
	if !CanStart(StatusAssigned) || CanStart(StatusPending) {
		t.Fatal("start transition mismatch")
	}
	if !CanSubmitNode(StatusInProgress) || !CanSubmitNode(StatusOverdue) || CanSubmitNode(StatusCancelled) {
		t.Fatal("node submit transition mismatch")
	}
	if !CanCancel(StatusPending) || !CanCancel(StatusAssigned) || !CanCancel(StatusInProgress) || CanCancel(StatusCompleted) {
		t.Fatal("cancel transition mismatch")
	}
	if !CanOverdue(StatusAssigned) || !CanOverdue(StatusInProgress) || CanOverdue(StatusCompleted) {
		t.Fatal("overdue transition mismatch")
	}
}
