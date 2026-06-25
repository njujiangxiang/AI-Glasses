package glass_api

import (
	"errors"
	"net/http"

	"aiglasses/server/internal/auth"
	"aiglasses/server/internal/config"
	"aiglasses/server/internal/defects"
	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"aiglasses/server/internal/tasks"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	auth    *auth.Service
	tasks   *tasks.Service
	defects *defects.Service
	db      *gorm.DB
	cfg     config.Config
}

// NewHandler 创建 AR 眼镜端 v1 文档兼容接口处理器。
func NewHandler(authSvc *auth.Service, taskSvc *tasks.Service, defectSvc *defects.Service) *Handler {
	return &Handler{auth: authSvc, tasks: taskSvc, defects: defectSvc}
}

// NewHandlerWithRuntime 创建包含数据库与配置能力的 AR 眼镜端 v1 接口处理器。
func NewHandlerWithRuntime(authSvc *auth.Service, taskSvc *tasks.Service, defectSvc *defects.Service, db *gorm.DB, cfg config.Config) *Handler {
	return &Handler{auth: authSvc, tasks: taskSvc, defects: defectSvc, db: db, cfg: cfg}
}

// Register 挂载模块五接口。所有接口均要求 glasses scope Bearer token。
func (h *Handler) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	v1.POST("/auth/refresh", h.refreshToken)
	v1.POST("/devices/register", h.registerDevice)

	v1Auth := v1.Group("", auth.Middleware(h.auth, auth.ScopeGlasses))
	v1Auth.POST("/auth/logout", h.logout)
	v1Auth.GET("/auth/userinfo", h.userInfo)
	v1Auth.POST("/devices/status/report", h.reportDeviceStatus)
	v1Auth.POST("/defects", h.createDefect)
	v1Auth.POST("/attachments/upload", h.uploadAttachment)
	v1Auth.POST("/attachments/upload/batch", h.uploadAttachmentBatch)
	v1Auth.GET("/attachments/:attachment_id", h.attachmentDetail)
	v1Auth.POST("/realtime/query", h.realtimeQuery)
	v1Auth.POST("/algorithm/invoke", h.invokeAlgorithm)
	v1Auth.GET("/dicts/task-types", h.taskTypes)

	group := v1Auth.Group("/tasks")
	group.GET("/cards", h.taskCards)
	group.GET("/my", h.myTasks)
	group.GET("/:task_id", h.taskDetail)
	group.POST("/:task_id/start", h.startTask)
	group.GET("/:task_id/current-node", h.currentNode)
	group.POST("/nodes/:node_id/results", h.submitNodeResult)
	group.POST("/:task_id/nodes/:node_id/progress", h.reportProgress)
	group.POST("/:task_id/submit", h.submitTask)
	group.POST("/:task_id/nodes/:node_id/skip", h.skipNode)
}

func (h *Handler) taskCards(c *gin.Context) {
	items, _, err := h.tasks.GlassesPage(auth.UserID(c), internalStatus(c.Query("status")), 1, 500)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	cards := map[string]*taskCard{}
	for _, item := range items {
		_, nodes, _, _, _, err := h.tasks.Detail(item.ID)
		if err != nil {
			httperr.Respond(c, err)
			return
		}
		key := planType(item)
		card := cards[key]
		if card == nil {
			card = &taskCard{PlanType: key}
			cards[key] = card
		}
		card.TotalCount++
		if item.Status != tasks.StatusSubmitted && item.Status != tasks.StatusCompleted && item.Status != tasks.StatusCancelled {
			card.UndoneCount++
		}
		card.Tasks = append(card.Tasks, taskCardItem{
			TaskID:         idString(item.ID),
			TaskName:       taskName(item),
			SubstationName: item.PointName,
			InspectArea:    item.EquipmentName,
			Status:         apiStatus(item.Status),
			ScheduledAt:    formatTimePtr(item.ScheduledAt),
			DueAt:          formatTime(item.DueAt),
			Progress:       progressOf(nodes),
		})
	}
	result := make([]taskCard, 0, len(cards))
	for _, card := range cards {
		result = append(result, *card)
	}
	httperr.OK(c, result)
}

func (h *Handler) myTasks(c *gin.Context) {
	page := intQuery(c, "page", 1)
	pageSize := intQuery(c, "page_size", 50)
	items, total, err := h.tasks.GlassesPage(auth.UserID(c), internalStatus(c.Query("status")), page, pageSize)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	list := make([]myTaskItem, 0, len(items))
	for _, item := range items {
		_, nodes, _, _, defects, err := h.tasks.Detail(item.ID)
		if err != nil {
			httperr.Respond(c, err)
			return
		}
		unfinished := 0
		for _, node := range nodes {
			if node.Status == tasks.NodePending {
				unfinished++
			}
		}
		list = append(list, myTaskItem{
			TaskID:              idString(item.ID),
			TaskName:            taskName(item),
			PlanType:            planType(item),
			SubstationName:      item.PointName,
			Status:              apiStatus(item.Status),
			ScheduledAt:         formatTimePtr(item.ScheduledAt),
			Progress:            progressOf(nodes),
			UnfinishedNodeCount: unfinished,
			DefectCount:         len(defects),
		})
	}
	httperr.OK(c, myTasksResponse{Total: total, List: list})
}

func (h *Handler) taskDetail(c *gin.Context) {
	taskID, err := parseID(c.Param("task_id"))
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid task_id"))
		return
	}
	if err := h.tasks.EnsureAccessible(taskID, auth.UserID(c)); err != nil {
		httperr.Respond(c, err)
		return
	}
	task, nodes, results, _, defects, err := h.tasks.Detail(taskID)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	resultByNode := resultMap(results)
	nodeItems := make([]nodeInfo, 0, len(nodes))
	for _, node := range nodes {
		nodeItems = append(nodeItems, toNodeInfo(node, resultByNode))
	}
	httperr.OK(c, taskDetailResponse{TaskInfo: toTaskInfo(task, progressOf(nodes)), Nodes: nodeItems, DefectCount: len(defects)})
}

func (h *Handler) startTask(c *gin.Context) {
	taskID, err := parseID(c.Param("task_id"))
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid task_id"))
		return
	}
	if err := h.tasks.Start(taskID, auth.UserID(c)); err != nil {
		httperr.Respond(c, err)
		return
	}
	task, _, _, _, _, err := h.tasks.Detail(taskID)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"started_at": formatTimePtr(task.StartedAt), "status": apiStatus(task.Status)})
}

func (h *Handler) currentNode(c *gin.Context) {
	taskID, err := parseID(c.Param("task_id"))
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid task_id"))
		return
	}
	node, err := h.tasks.CurrentNode(taskID, auth.UserID(c))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	data := toNodeInfo(node, nil)
	httperr.OK(c, gin.H{
		"node_id":              data.NodeID,
		"template_node_id":     data.TemplateNodeID,
		"name":                 data.Name,
		"node_desc":            data.NodeDesc,
		"node_type":            data.NodeType,
		"min_photos":           data.MinPhotos,
		"require_text":         data.RequireText,
		"allow_abnormal":       data.AllowAbnormal,
		"require_live_capture": data.RequireLiveCapture,
		"is_mandatory":         data.IsMandatory,
		"configs":              data.Configs,
		"timeout_remain":       0,
	})
}

func (h *Handler) submitNodeResult(c *gin.Context) {
	nodeID, err := parseID(c.Param("node_id"))
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid node_id"))
		return
	}
	var body submitNodeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	taskID, err := parseID(body.TaskID)
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid task_id"))
		return
	}
	attachments, err := parseAttachmentIDs(body.AttachmentIDs)
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid attachment_ids"))
		return
	}
	abnormal := body.IsAbnormal == "1"
	result, err := h.tasks.SubmitNode(taskID, nodeID, auth.UserID(c), body.IdempotencyKey, tasks.NodeResultInput{TextNote: combineText(body), AttachmentIDs: attachments, IsAbnormal: abnormal})
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	var defectID *string
	if abnormal && h.defects != nil {
		desc := "节点异常"
		if body.AbnormalDesc != nil && *body.AbnormalDesc != "" {
			desc = *body.AbnormalDesc
		}
		defect, err := h.defects.Create(taskID, nodeID, auth.UserID(c), desc)
		if err != nil {
			httperr.Respond(c, err)
			return
		}
		id := idString(defect.ID)
		defectID = &id
	}
	_, nodes, _, _, _, err := h.tasks.Detail(taskID)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	next := nextNodeID(nodes, nodeID)
	httperr.OK(c, submitNodeResponse{ResultID: idString(result.ID), IsAbnormal: boolString(abnormal), NextNodeID: next, TaskProgress: progressOf(nodes), DefectID: defectID})
}

func (h *Handler) reportProgress(c *gin.Context) {
	taskID, err := parseID(c.Param("task_id"))
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid task_id"))
		return
	}
	nodeID, err := parseID(c.Param("node_id"))
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid node_id"))
		return
	}
	var body progressRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	if body.NodeID != "" {
		bodyNodeID, err := parseID(body.NodeID)
		if err != nil || bodyNodeID != nodeID {
			httperr.Respond(c, httperr.New(httperr.ValidationFailed, "node_id mismatch"))
			return
		}
	}
	if err := h.tasks.AcceptNodeProgress(taskID, nodeID, auth.UserID(c), body.Progress); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"accepted": true, "progress": body.Progress})
}

func (h *Handler) submitTask(c *gin.Context) {
	taskID, err := parseID(c.Param("task_id"))
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid task_id"))
		return
	}
	if err := h.tasks.SubmitTask(taskID, auth.UserID(c)); err != nil {
		httperr.Respond(c, err)
		return
	}
	task, nodes, _, _, defects, err := h.tasks.Detail(taskID)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"status": apiStatus(task.Status), "submitted_at": formatTimePtr(task.SubmittedAt), "unfinished_mandatory_nodes": unfinishedNodes(nodes), "defect_summary": gin.H{"total": len(defects)}})
}

func (h *Handler) skipNode(c *gin.Context) {
	taskID, err := parseID(c.Param("task_id"))
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid task_id"))
		return
	}
	nodeID, err := parseID(c.Param("node_id"))
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "invalid node_id"))
		return
	}
	var body skipRequest
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, http.ErrBodyNotAllowed) {
		httperr.Respond(c, err)
		return
	}
	if err := h.tasks.SkipNode(taskID, nodeID, auth.UserID(c), body.Reason); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"skipped": true})
}

func intQuery(c *gin.Context, key string, fallback int) int {
	value, err := parseID(c.Query(key))
	if err != nil || value == 0 {
		return fallback
	}
	return int(value)
}

func nextNodeID(nodes []database.InspectionTaskNode, current uint64) *string {
	for _, node := range nodes {
		if node.ID != current && node.Status == tasks.NodePending {
			id := idString(node.ID)
			return &id
		}
	}
	return nil
}

func unfinishedNodes(nodes []database.InspectionTaskNode) []gin.H {
	items := []gin.H{}
	for _, node := range nodes {
		if node.Status == tasks.NodePending && (node.MinPhotos > 0 || node.RequireText) {
			items = append(items, gin.H{"node_id": idString(node.ID), "name": node.Name})
		}
	}
	return items
}
