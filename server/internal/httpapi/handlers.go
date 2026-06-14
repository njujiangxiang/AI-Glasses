// Package httpapi 将领域服务适配为 REST 接口。它组织公开登录路由、后台受保护路由和
// 眼镜端受保护路由，并保持 Vue 后台与 Android 客户端共用的响应结构和错误码契约。
package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"aiglasses/server/internal/attachments"
	"aiglasses/server/internal/auth"
	"aiglasses/server/internal/defects"
	"aiglasses/server/internal/devices"
	"aiglasses/server/internal/events"
	"aiglasses/server/internal/plans"
	"aiglasses/server/internal/platform/httperr"
	"aiglasses/server/internal/tasks"
	"aiglasses/server/internal/templates"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	auth        *auth.Service
	attachments *attachments.Service
	defects     *defects.Service
	devices     *devices.Service
	plans       *plans.Service
	tasks       *tasks.Service
	templates   *templates.Service
	scheduler   *events.Scheduler
}

// NewHandler 创建 HTTP 处理器集合，并注入所有业务服务。
func NewHandler(authSvc *auth.Service, attachmentSvc *attachments.Service, defectSvc *defects.Service, deviceSvc *devices.Service, planSvc *plans.Service, taskSvc *tasks.Service, templateSvc *templates.Service, scheduler *events.Scheduler) *Handler {
	return &Handler{auth: authSvc, attachments: attachmentSvc, defects: defectSvc, devices: deviceSvc, plans: planSvc, tasks: taskSvc, templates: templateSvc, scheduler: scheduler}
}

// Register 注册公开接口、后台接口和眼镜端接口。
func (h *Handler) Register(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	api := r.Group("/api")
	api.POST("/admin/auth/login", h.adminLogin)
	api.POST("/glasses/auth/token", h.glassesLogin)
	admin := api.Group("/admin", auth.Middleware(h.auth, auth.ScopeAdmin))
	admin.POST("/templates", h.createTemplate)
	admin.GET("/templates", h.listTemplates)
	admin.POST("/plans", h.createPlan)
	admin.GET("/plans", h.listPlans)
	admin.POST("/plans/:id/enable", h.enablePlan)
	admin.POST("/plans/:id/generate-now", h.generateNow)
	admin.GET("/tasks", h.adminTasks)
	admin.GET("/tasks/:id", h.taskDetail)
	admin.POST("/tasks/:id/cancel", h.cancelTask)
	admin.POST("/tasks/:id/complete", h.completeTask)
	admin.GET("/defects", h.listDefects)
	admin.POST("/defects/:id/confirm", h.confirmDefect)
	admin.POST("/defects/:id/close", h.closeDefect)
	admin.GET("/devices", h.listDevices)
	admin.POST("/devices", h.registerDevice)
	admin.POST("/devices/:id/revoke", h.revokeDevice)
	admin.POST("/devices/:id/disable-lost", h.disableLostDevice)

	glasses := api.Group("/glasses", auth.Middleware(h.auth, auth.ScopeGlasses))
	glasses.GET("/tasks", h.glassesTasks)
	glasses.GET("/tasks/:id", h.taskDetail)
	glasses.POST("/tasks/:id/claim", h.claimTask)
	glasses.POST("/tasks/:id/start", h.startTask)
	glasses.POST("/tasks/:id/nodes/:nodeId/result", h.submitNode)
	glasses.POST("/tasks/:id/submit", h.submitTask)
	glasses.POST("/tasks/:id/nodes/:nodeId/abnormal", h.reportAbnormal)
	glasses.POST("/attachments/presign", h.presignAttachment)
}

// adminLogin 处理后台管理员登录并返回 admin scope token。
func (h *Handler) adminLogin(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	token, user, err := h.auth.Login(body.Username, auth.ScopeAdmin, nil)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"access_token": token, "user": user})
}

// glassesLogin 处理眼镜端登录并返回携带设备 ID 的 glasses scope token。
func (h *Handler) glassesLogin(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		DeviceID uint64 `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	deviceID := body.DeviceID
	token, user, err := h.auth.Login(body.Username, auth.ScopeGlasses, &deviceID)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"access_token": token, "user": user, "device_id": deviceID})
}

// createTemplate 创建巡检模板。
func (h *Handler) createTemplate(c *gin.Context) {
	var input templates.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.templates.Create(input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// listTemplates 查询巡检模板列表。
func (h *Handler) listTemplates(c *gin.Context) {
	result, err := h.templates.List()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// createPlan 创建巡检任务计划。
func (h *Handler) createPlan(c *gin.Context) {
	var input plans.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.plans.Create(input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// listPlans 查询巡检任务计划列表。
func (h *Handler) listPlans(c *gin.Context) {
	result, err := h.plans.List()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// enablePlan 启用指定巡检任务计划。
func (h *Handler) enablePlan(c *gin.Context) {
	if err := h.plans.Enable(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"enabled": true})
}

// generateNow 手动触发一次计划调度生成任务。
func (h *Handler) generateNow(c *gin.Context) {
	if err := h.scheduler.Tick(time.Now().UTC()); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"generated": true})
}

// adminTasks 查询后台任务列表。
func (h *Handler) adminTasks(c *gin.Context) {
	result, err := h.tasks.AdminList(c.Query("status"), intQuery(c, "limit", 50))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// glassesTasks 查询当前眼镜用户可执行任务列表。
func (h *Handler) glassesTasks(c *gin.Context) {
	result, err := h.tasks.GlassesList(auth.UserID(c), nil, intQuery(c, "limit", 50))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// taskDetail 查询任务详情。
func (h *Handler) taskDetail(c *gin.Context) {
	task, nodes, results, attachments, defects, err := h.tasks.Detail(idParam(c, "id"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"task": task, "nodes": nodes, "results": results, "attachments": attachments, "defects": defects})
}

// claimTask 处理眼镜端领取班组任务。
func (h *Handler) claimTask(c *gin.Context) {
	if err := h.tasks.Claim(idParam(c, "id"), auth.UserID(c)); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"claimed": true})
}

// startTask 处理眼镜端开始任务。
func (h *Handler) startTask(c *gin.Context) {
	if err := h.tasks.Start(idParam(c, "id"), auth.UserID(c)); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"started": true})
}

// submitNode 处理眼镜端提交节点结果。
func (h *Handler) submitNode(c *gin.Context) {
	var input tasks.NodeResultInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.tasks.SubmitNode(idParam(c, "id"), idParam(c, "nodeId"), auth.UserID(c), c.GetHeader("Idempotency-Key"), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// submitTask 处理眼镜端提交整单任务。
func (h *Handler) submitTask(c *gin.Context) {
	if err := h.tasks.SubmitTask(idParam(c, "id"), auth.UserID(c)); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"submitted": true})
}

// cancelTask 处理后台取消任务。
func (h *Handler) cancelTask(c *gin.Context) {
	if err := h.tasks.Cancel(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"cancelled": true})
}

// completeTask 处理后台确认任务完成。
func (h *Handler) completeTask(c *gin.Context) {
	if err := h.tasks.Complete(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"completed": true})
}

// reportAbnormal 处理眼镜端异常上报并创建缺陷。
func (h *Handler) reportAbnormal(c *gin.Context) {
	var body struct {
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.defects.Create(idParam(c, "id"), idParam(c, "nodeId"), auth.UserID(c), body.Description)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// listDefects 查询缺陷列表。
func (h *Handler) listDefects(c *gin.Context) {
	result, err := h.defects.List(c.Query("status"), intQuery(c, "limit", 50))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// confirmDefect 处理后台确认缺陷。
func (h *Handler) confirmDefect(c *gin.Context) {
	if err := h.defects.Confirm(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"confirmed": true})
}

// closeDefect 处理后台关闭缺陷。
func (h *Handler) closeDefect(c *gin.Context) {
	var body struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	if err := h.defects.Close(idParam(c, "id"), body.Reason); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"closed": true})
}

// listDevices 查询设备列表。
func (h *Handler) listDevices(c *gin.Context) {
	result, err := h.devices.List()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// registerDevice 处理后台登记新设备。
func (h *Handler) registerDevice(c *gin.Context) {
	var body struct {
		SerialNo string `json:"serial_no"`
		Name     string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.devices.Register(body.SerialNo, body.Name)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// revokeDevice 处理后台撤销设备。
func (h *Handler) revokeDevice(c *gin.Context) {
	if err := h.devices.Revoke(auth.UserID(c), idParam(c, "id"), "admin revoke"); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"revoked": true})
}

// disableLostDevice 处理后台将设备标记为丢失禁用。
func (h *Handler) disableLostDevice(c *gin.Context) {
	if err := h.devices.DisableLost(auth.UserID(c), idParam(c, "id"), "lost device"); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"disabled": true})
}

// presignAttachment 处理证据附件预签名上传申请。
func (h *Handler) presignAttachment(c *gin.Context) {
	var input attachments.PresignInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	if input.DeviceID == nil {
		input.DeviceID = auth.DeviceID(c)
	}
	result, err := h.attachments.Presign(c.Request.Context(), auth.UserID(c), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// idParam 从路由参数读取无符号整数 ID。
func idParam(c *gin.Context, key string) uint64 {
	id, _ := strconv.ParseUint(c.Param(key), 10, 64)
	return id
}

// intQuery 从查询参数读取整数，不存在或非法时返回默认值。
func intQuery(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
