package glass_api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/tasks"
)

type taskCard struct {
	PlanType    string         `json:"plan_type"`
	TotalCount  int            `json:"total_count"`
	UndoneCount int            `json:"undone_count"`
	Tasks       []taskCardItem `json:"tasks"`
}

type taskCardItem struct {
	TaskID         string  `json:"task_id"`
	TaskName       string  `json:"task_name"`
	SubstationName string  `json:"substation_name"`
	InspectArea    string  `json:"inspect_area"`
	Status         string  `json:"status"`
	ScheduledAt    string  `json:"scheduled_at"`
	DueAt          string  `json:"due_at"`
	Progress       float64 `json:"progress"`
	GlassesSN      string  `json:"glasses_sn"`
}

type taskDetailResponse struct {
	TaskInfo    taskInfo   `json:"task_info"`
	Nodes       []nodeInfo `json:"nodes"`
	DefectCount int        `json:"defect_count"`
}

type taskInfo struct {
	TaskID         string  `json:"task_id"`
	TaskName       string  `json:"task_name"`
	PlanType       string  `json:"plan_type"`
	SubstationName string  `json:"substation_name"`
	InspectArea    string  `json:"inspect_area"`
	Status         string  `json:"status"`
	ScheduledAt    string  `json:"scheduled_at"`
	DueAt          string  `json:"due_at"`
	ExecutorName   string  `json:"executor_name"`
	ExecutorUnit   string  `json:"executor_unit"`
	GlassesSN      string  `json:"glasses_sn"`
	StartedAt      string  `json:"started_at"`
	Progress       float64 `json:"progress"`
}

type nodeInfo struct {
	NodeID             string       `json:"node_id"`
	TemplateNodeID     string       `json:"template_node_id"`
	Name               string       `json:"name"`
	NodeDesc           string       `json:"node_desc"`
	SortOrder          int          `json:"sort_order"`
	NodeType           string       `json:"node_type"`
	MinPhotos          int          `json:"min_photos"`
	RequireText        string       `json:"require_text"`
	AllowAbnormal      string       `json:"allow_abnormal"`
	RequireLiveCapture string       `json:"require_live_capture"`
	IsMandatory        string       `json:"is_mandatory"`
	IsRequired         string       `json:"is_required"`
	Status             string       `json:"status"`
	HasResult          bool         `json:"has_result"`
	TimeoutSecond      int          `json:"timeout_second"`
	Configs            []nodeConfig `json:"configs"`
}

type nodeConfig struct {
	ConfigID    string `json:"config_id"`
	ConfigCode  string `json:"config_code"`
	ConfigName  string `json:"config_name"`
	ConfigValue string `json:"config_value"`
	IsDefault   string `json:"is_default"`
}

type submitNodeRequest struct {
	IdempotencyKey string  `json:"idempotency_key"`
	TaskID         string  `json:"task_id"`
	TaskTypeCode   string  `json:"task_type_code"`
	Feedback       *string `json:"feedback_content"`
	TextNote       *string `json:"text_note"`
	LocationGPS    string  `json:"location_gps"`
	AttachmentIDs  string  `json:"attachment_ids"`
	IsAbnormal     string  `json:"is_abnormal"`
	AbnormalDesc   *string `json:"abnormal_desc"`
	Remark         string  `json:"remark"`
}

type submitNodeResponse struct {
	ResultID        string  `json:"result_id"`
	AlgorithmResult *string `json:"algorithm_result"`
	QueryResult     *string `json:"query_result"`
	IsAbnormal      string  `json:"is_abnormal"`
	NextNodeID      *string `json:"next_node_id"`
	TaskProgress    float64 `json:"task_progress"`
	DefectID        *string `json:"defect_id"`
}

type progressRequest struct {
	NodeID   string  `json:"node_id"`
	Progress float64 `json:"progress"`
}

type skipRequest struct {
	Reason string `json:"reason"`
}

type myTasksResponse struct {
	Total int64        `json:"total"`
	List  []myTaskItem `json:"list"`
}

type myTaskItem struct {
	TaskID              string  `json:"task_id"`
	TaskName            string  `json:"task_name"`
	PlanType            string  `json:"plan_type"`
	SubstationName      string  `json:"substation_name"`
	Status              string  `json:"status"`
	ScheduledAt         string  `json:"scheduled_at"`
	Progress            float64 `json:"progress"`
	UnfinishedNodeCount int     `json:"unfinished_node_count"`
	DefectCount         int     `json:"defect_count"`
}

func toTaskInfo(task database.InspectionTask, progress float64) taskInfo {
	return taskInfo{
		TaskID:         idString(task.ID),
		TaskName:       taskName(task),
		PlanType:       planType(task),
		SubstationName: task.PointName,
		InspectArea:    task.EquipmentName,
		Status:         apiStatus(task.Status),
		ScheduledAt:    formatTimePtr(task.ScheduledAt),
		DueAt:          formatTime(task.DueAt),
		StartedAt:      formatTimePtr(task.StartedAt),
		Progress:       progress,
	}
}

func toNodeInfo(node database.InspectionTaskNode, resultByNode map[uint64]database.TaskNodeResult) nodeInfo {
	_, hasResult := resultByNode[node.ID]
	return nodeInfo{
		NodeID:             idString(node.ID),
		TemplateNodeID:     idString(node.TemplateNodeID),
		Name:               node.Name,
		SortOrder:          node.SortOrder,
		NodeType:           node.NodeType,
		MinPhotos:          node.MinPhotos,
		RequireText:        boolString(node.RequireText),
		AllowAbnormal:      boolString(node.AllowAbnormal),
		RequireLiveCapture: "0",
		IsMandatory:        boolString(node.MinPhotos > 0 || node.RequireText),
		IsRequired:         boolString(node.MinPhotos > 0 || node.RequireText),
		Status:             apiNodeStatus(node.Status),
		HasResult:          hasResult,
		Configs:            []nodeConfig{},
	}
}

func progressOf(nodes []database.InspectionTaskNode) float64 {
	if len(nodes) == 0 {
		return 0
	}
	done := 0
	for _, node := range nodes {
		switch node.Status {
		case tasks.NodeCompleted, tasks.NodeAbnormal, tasks.NodeSkipped:
			done++
		}
	}
	return float64(done) / float64(len(nodes))
}

func resultMap(results []database.TaskNodeResult) map[uint64]database.TaskNodeResult {
	items := make(map[uint64]database.TaskNodeResult, len(results))
	for _, result := range results {
		items[result.NodeID] = result
	}
	return items
}

func parseID(value string) (uint64, error) {
	return strconv.ParseUint(strings.TrimSpace(value), 10, 64)
}

func parseAttachmentIDs(value string) ([]uint64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	ids := make([]uint64, 0, len(parts))
	for _, part := range parts {
		id, err := parseID(part)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func combineText(input submitNodeRequest) string {
	parts := []string{}
	if input.Feedback != nil && strings.TrimSpace(*input.Feedback) != "" {
		parts = append(parts, strings.TrimSpace(*input.Feedback))
	}
	if input.TextNote != nil && strings.TrimSpace(*input.TextNote) != "" {
		parts = append(parts, strings.TrimSpace(*input.TextNote))
	}
	if input.Remark != "" {
		parts = append(parts, strings.TrimSpace(input.Remark))
	}
	if input.LocationGPS != "" {
		parts = append(parts, fmt.Sprintf("GPS:%s", strings.TrimSpace(input.LocationGPS)))
	}
	return strings.Join(parts, "\n")
}

func taskName(task database.InspectionTask) string {
	if task.EquipmentName != "" {
		return task.EquipmentName
	}
	if task.PointName != "" {
		return task.PointName
	}
	return fmt.Sprintf("巡检任务 #%d", task.ID)
}

func planType(task database.InspectionTask) string {
	if task.PointName != "" {
		return task.PointName
	}
	return "巡检任务"
}

func apiStatus(status string) string {
	switch status {
	case tasks.StatusAssigned, tasks.StatusPending:
		return "pending"
	case tasks.StatusInProgress, tasks.StatusOverdue:
		return "executing"
	case tasks.StatusCompleted:
		return "completed"
	case tasks.StatusCancelled:
		return "cancelled"
	default:
		return status
	}
}

func internalStatus(status string) string {
	switch status {
	case "executing":
		return tasks.StatusInProgress
	case "completed":
		return tasks.StatusCompleted
	default:
		return status
	}
}

func apiNodeStatus(status string) string {
	if status == tasks.NodeSkipped {
		return "skiped"
	}
	return status
}

func boolString(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func idString(id uint64) string { return strconv.FormatUint(id, 10) }

func formatTime(value time.Time) string { return value.UTC().Format(time.RFC3339) }

func formatTimePtr(value *time.Time) string {
	if value == nil {
		return ""
	}
	return formatTime(*value)
}
