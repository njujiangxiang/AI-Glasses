// Package httpapi 将领域服务适配为 REST 接口。它组织公开登录路由、后台受保护路由和
// 眼镜端受保护路由，并保持 Vue 后台与 Android 客户端共用的响应结构和错误码契约。
package httpapi

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"aiglasses/server/internal/attachments"
	"aiglasses/server/internal/auth"
	"aiglasses/server/internal/businesscodes"
	"aiglasses/server/internal/defects"
	"aiglasses/server/internal/devices"
	"aiglasses/server/internal/events"
	"aiglasses/server/internal/monitoring"
	"aiglasses/server/internal/nodes"
	"aiglasses/server/internal/organizations"
	"aiglasses/server/internal/plans"
	"aiglasses/server/internal/points"
	"aiglasses/server/internal/platform/httperr"
	"aiglasses/server/internal/rbac"
	"aiglasses/server/internal/tasks"
	"aiglasses/server/internal/templates"
	"aiglasses/server/internal/users"
	"aiglasses/server/internal/workflows"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	auth          *auth.Service
	attachments   *attachments.Service
	businessCodes *businesscodes.Service
	defects       *defects.Service
	devices       *devices.Service
	nodes         *nodes.Service
	organizations *organizations.Service
	plans         *plans.Service
	points        *points.Service
	tasks         *tasks.Service
	templates     *templates.Service
	users         *users.Service
	workflows     *workflows.Service
	scheduler     *events.Scheduler
	rbac          *rbac.Service
	monitoringHub *monitoring.Hub
}

// NewHandler 创建 HTTP 处理器集合，并注入所有业务服务。
func NewHandler(authSvc *auth.Service, attachmentSvc *attachments.Service, businessCodeSvc *businesscodes.Service, defectSvc *defects.Service, deviceSvc *devices.Service, nodesSvc *nodes.Service, orgSvc *organizations.Service, planSvc *plans.Service, pointsSvc *points.Service, taskSvc *tasks.Service, templateSvc *templates.Service, userSvc *users.Service, workflowSvc *workflows.Service, scheduler *events.Scheduler) *Handler {
	return &Handler{auth: authSvc, attachments: attachmentSvc, businessCodes: businessCodeSvc, defects: defectSvc, devices: deviceSvc, nodes: nodesSvc, organizations: orgSvc, plans: planSvc, points: pointsSvc, tasks: taskSvc, templates: templateSvc, users: userSvc, workflows: workflowSvc, scheduler: scheduler}
}

// Register 注册公开接口、后台接口和眼镜端接口。
func (h *Handler) Register(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	api := r.Group("/api")
	api.POST("/admin/auth/login", h.adminLogin)
	api.POST("/glasses/auth/token", h.glassesLogin)
	admin := api.Group("/admin", auth.Middleware(h.auth, auth.ScopeAdmin))
	// 节点管理
	admin.GET("/nodes", h.listNodes)
	admin.GET("/nodes/unassigned", h.listUnassignedNodes)
	admin.GET("/nodes/:id", h.getNode)
	admin.POST("/nodes", h.createNode)
	admin.POST("/nodes/:id/update", h.updateNode)
	admin.POST("/nodes/:id/delete", h.deleteNode)
	// 巡检点位管理
	admin.GET("/inspection-points", h.listInspectionPoints)
	admin.GET("/inspection-points/all", h.listAllInspectionPoints)
	admin.GET("/inspection-points/:id", h.getInspectionPoint)
	admin.POST("/inspection-points", h.createInspectionPoint)
	admin.POST("/inspection-points/:id/update", h.updateInspectionPoint)
	admin.POST("/inspection-points/:id/enable", h.enableInspectionPoint)
	admin.POST("/inspection-points/:id/disable", h.disableInspectionPoint)
	admin.POST("/inspection-points/:id/delete", h.deleteInspectionPoint)
	// 模板管理
	admin.POST("/templates", h.createTemplate)
	admin.GET("/templates", h.listTemplates)
	admin.GET("/templates/all", h.listAllTemplates)
	admin.GET("/templates/:id", h.getTemplate)
	admin.POST("/templates/:id/update", h.updateTemplate)
	admin.POST("/templates/:id/enable", h.enableTemplate)
	admin.POST("/templates/:id/disable", h.disableTemplate)
	admin.POST("/templates/:id/delete", h.deleteTemplate)
	// 计划管理
	admin.POST("/plans", h.createPlan)
	admin.GET("/plans", h.listPlans)
	admin.GET("/plans/:id", h.getPlan)
	admin.POST("/plans/:id/update", h.updatePlan)
	admin.POST("/plans/:id/enable", h.enablePlan)
	admin.POST("/plans/:id/disable", h.disablePlan)
	admin.POST("/plans/:id/delete", h.deletePlan)
	admin.POST("/plans/:id/generate-now", h.generateNow)
	// 任务管理
	admin.POST("/tasks", h.createTask)
	admin.GET("/tasks", h.adminTasks)
	admin.GET("/tasks/:id", h.taskDetail)
	admin.GET("/tasks/:id/results", h.taskResults)
	admin.POST("/tasks/:id/cancel", h.cancelTask)
	admin.POST("/tasks/:id/complete", h.completeTask)
	admin.POST("/tasks/:id/delete", h.deleteTask)
	admin.GET("/defects", h.listDefects)
	admin.POST("/defects/:id/confirm", h.confirmDefect)
	admin.POST("/defects/:id/close", h.closeDefect)
	admin.GET("/devices", h.listDevices)
	admin.POST("/devices", h.registerDevice)
	admin.POST("/devices/:id/revoke", h.revokeDevice)
	admin.POST("/devices/:id/disable-lost", h.disableLostDevice)
	admin.GET("/business-codes", h.listBusinessCodes)
	admin.POST("/business-codes", h.createBusinessCode)
	admin.POST("/business-codes/generate", h.generateBusinessCode)
	admin.POST("/business-codes/:id/update", h.updateBusinessCode)
	admin.POST("/business-codes/:id/enable", h.enableBusinessCode)
	admin.POST("/business-codes/:id/disable", h.disableBusinessCode)
	admin.POST("/business-codes/:id/delete", h.deleteBusinessCode)
	admin.GET("/organizations", h.listOrganizations)
	admin.GET("/organizations/tree", h.organizationTree)
	admin.POST("/organizations", h.createOrganization)
	admin.POST("/organizations/:id/update", h.updateOrganization)
	admin.POST("/organizations/:id/enable", h.enableOrganization)
	admin.POST("/organizations/:id/disable", h.disableOrganization)
	admin.POST("/organizations/:id/delete", h.deleteOrganization)
	admin.GET("/users", h.listUsers)
	admin.GET("/users/all", h.listAllUsers)
	admin.GET("/users/:id", h.getUser)
	admin.POST("/users", h.createUser)
	admin.POST("/users/:id/update", h.updateUser)
	admin.POST("/users/:id/enable", h.enableUser)
	admin.POST("/users/:id/disable", h.disableUser)
	admin.POST("/users/:id/avatar", h.setUserAvatar)
	admin.GET("/users/:id/avatar", h.getUserAvatar)
	admin.GET("/workflows", h.listWorkflows)
	admin.POST("/workflows", h.createWorkflow)
	admin.GET("/workflows/:id", h.getWorkflow)
	admin.POST("/workflows/:id", h.updateWorkflow)
	admin.POST("/workflows/:id/publish", h.publishWorkflow)
	admin.POST("/workflows/:id/unpublish", h.unpublishWorkflow)
	admin.POST("/workflows/:id/delete", h.deleteWorkflow)
	admin.POST("/workflows/:id/steps", h.addWorkflowStep)
	admin.POST("/workflows/:id/steps/:stepId", h.updateWorkflowStep)
	admin.POST("/workflows/:id/steps/:stepId/delete", h.deleteWorkflowStep)
	admin.POST("/workflows/:id/steps/:stepId/duplicate", h.duplicateWorkflowStep)
	admin.POST("/workflows/:id/steps/reorder", h.reorderWorkflowSteps)

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
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	token, user, err := h.auth.Login(body.Username, body.Password, auth.ScopeAdmin, nil)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	companyName, err := h.auth.OrganizationName(user.OrgCode)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"access_token": token, "user": user, "company_name": companyName})
}

// glassesLogin 处理眼镜端登录并返回携带设备 ID 的 glasses scope token。
func (h *Handler) glassesLogin(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		DeviceID uint64 `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	deviceID := body.DeviceID
	pair, user, err := h.auth.LoginWithRefresh(body.Username, body.Password, deviceID)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken, "token_type": "Bearer", "expires_in": pair.ExpiresIn, "refresh_expires_in": pair.RefreshExpiresIn, "user": user, "device_id": deviceID})
}

// listBusinessCodes 查询业务编码配置列表。
func (h *Handler) listBusinessCodes(c *gin.Context) {
	result, err := h.businessCodes.List(c.Query("keyword"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// createBusinessCode 创建业务编码配置。
func (h *Handler) createBusinessCode(c *gin.Context) {
	var input businesscodes.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.businessCodes.Create(input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// updateBusinessCode 更新业务编码配置。
func (h *Handler) updateBusinessCode(c *gin.Context) {
	var input businesscodes.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.businessCodes.Update(idParam(c, "id"), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// enableBusinessCode 启用业务编码配置。
func (h *Handler) enableBusinessCode(c *gin.Context) {
	if err := h.businessCodes.Enable(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"status": businesscodes.StatusActive})
}

// disableBusinessCode 停用业务编码配置。
func (h *Handler) disableBusinessCode(c *gin.Context) {
	if err := h.businessCodes.Disable(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"status": businesscodes.StatusDisabled})
}

// deleteBusinessCode 删除业务编码配置。
func (h *Handler) deleteBusinessCode(c *gin.Context) {
	if err := h.businessCodes.Delete(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"deleted": true})
}

// generateBusinessCode 生成下一业务编号；该操作会消耗真实 Redis 流水号。
func (h *Handler) generateBusinessCode(c *gin.Context) {
	var body struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	// 从当前用户获取组织编码
	userID := auth.UserID(c)
	user, err := h.users.Get(userID)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	code, err := h.businessCodes.GenerateDaily(c.Request.Context(), body.Code, user.OrgCode)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"code": code})
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

// listTemplates 查询巡检模板列表（分页）。
func (h *Handler) listTemplates(c *gin.Context) {
	result, err := h.templates.List(templates.ListQuery{
		Keyword:  c.Query("keyword"),
		Page:     intQuery(c, "page", 1),
		PageSize: intQuery(c, "page_size", 20),
	})
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// listAllTemplates 查询所有模板（不分页，用于下拉选择）。
func (h *Handler) listAllTemplates(c *gin.Context) {
	result, err := h.templates.ListAll()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// getTemplate 查询模板详情（含节点列表）。
func (h *Handler) getTemplate(c *gin.Context) {
	template, ns, err := h.templates.GetDetail(idParam(c, "id"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"template": template, "nodes": ns})
}

// updateTemplate 更新巡检模板。
func (h *Handler) updateTemplate(c *gin.Context) {
	var input templates.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.templates.Update(idParam(c, "id"), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// enableTemplate 启用模板。
func (h *Handler) enableTemplate(c *gin.Context) {
	if err := h.templates.Enable(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"enabled": true})
}

// disableTemplate 停用模板。
func (h *Handler) disableTemplate(c *gin.Context) {
	if err := h.templates.Disable(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"disabled": true})
}

// deleteTemplate 删除模板。
func (h *Handler) deleteTemplate(c *gin.Context) {
	if err := h.templates.Delete(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"deleted": true})
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

// listPlans 查询巡检任务计划列表（分页）。
func (h *Handler) listPlans(c *gin.Context) {
	result, err := h.plans.List(plans.ListQuery{
		Keyword:  c.Query("keyword"),
		Status:   c.Query("status"),
		Page:     intQuery(c, "page", 1),
		PageSize: intQuery(c, "page_size", 20),
	})
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// getPlan 查询计划详情。
func (h *Handler) getPlan(c *gin.Context) {
	result, err := h.plans.GetDetail(idParam(c, "id"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// updatePlan 更新计划。
func (h *Handler) updatePlan(c *gin.Context) {
	var input plans.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.plans.Update(idParam(c, "id"), input)
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

// disablePlan 停用计划。
func (h *Handler) disablePlan(c *gin.Context) {
	if err := h.plans.Disable(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"disabled": true})
}

// deletePlan 删除计划。
func (h *Handler) deletePlan(c *gin.Context) {
	if err := h.plans.Delete(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"deleted": true})
}

// generateNow 为指定计划手动触发一次任务生成。
func (h *Handler) generateNow(c *gin.Context) {
	generated, err := h.scheduler.GenerateForPlan(idParam(c, "id"), time.Now().UTC())
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"generated": generated})
}

// createTask 后台手动创建巡检任务。
func (h *Handler) createTask(c *gin.Context) {
	var input tasks.AdminCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.tasks.AdminCreate(input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// adminTasks 查询后台任务列表（分页）。
func (h *Handler) adminTasks(c *gin.Context) {
	result, err := h.tasks.AdminList(tasks.AdminListQuery{
		Keyword:    c.Query("keyword"),
		Status:     c.Query("status"),
		TemplateID: func() uint64 { id, _ := strconv.ParseUint(c.Query("template_id"), 10, 64); return id }(),
		Page:       intQuery(c, "page", 1),
		PageSize:   intQuery(c, "page_size", 20),
	})
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// taskResults 查询任务执行结果。
func (h *Handler) taskResults(c *gin.Context) {
	_, ns, results, attachments, defects, err := h.tasks.Detail(idParam(c, "id"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"nodes": ns, "results": results, "attachments": attachments, "defects": defects})
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

// deleteTask 后台删除巡检任务及其关联数据。
func (h *Handler) deleteTask(c *gin.Context) {
	if err := h.tasks.AdminDelete(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"deleted": true})
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
		OrgCode  string `json:"org_code"`
		Status   string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.devices.Register(devices.RegisterInput{
		SerialNo: body.SerialNo,
		Name:     body.Name,
		OrgCode:  body.OrgCode,
		Status:   body.Status,
	})
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

// listOrganizations 查询单位组织列表。
func (h *Handler) listOrganizations(c *gin.Context) {
	result, err := h.organizations.List(c.Query("keyword"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// organizationTree 查询单位组织树。
func (h *Handler) organizationTree(c *gin.Context) {
	result, err := h.organizations.Tree()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// createOrganization 创建单位组织。
func (h *Handler) createOrganization(c *gin.Context) {
	var input organizations.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.organizations.Create(input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// updateOrganization 更新单位组织。
func (h *Handler) updateOrganization(c *gin.Context) {
	var input organizations.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.organizations.Update(idParam(c, "id"), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// enableOrganization 启用单位组织。
func (h *Handler) enableOrganization(c *gin.Context) {
	if err := h.organizations.Enable(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"enabled": true})
}

// disableOrganization 停用单位组织。
func (h *Handler) disableOrganization(c *gin.Context) {
	if err := h.organizations.Disable(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"disabled": true})
}

// deleteOrganization 删除单位组织。
func (h *Handler) deleteOrganization(c *gin.Context) {
	if err := h.organizations.Delete(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"deleted": true})
}

// listUsers 查询后台用户列表。
func (h *Handler) listUsers(c *gin.Context) {
	result, err := h.users.List(users.ListQuery{
		Keyword:  c.Query("keyword"),
		OrgCode:  c.Query("org_code"),
		Status:   c.Query("status"),
		Page:     intQuery(c, "page", 1),
		PageSize: intQuery(c, "page_size", 20),
	})
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// listAllUsers 查询所有启用用户（不分页，用于下拉选择）。
func (h *Handler) listAllUsers(c *gin.Context) {
	result, err := h.users.ListAll()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// getUser 查询后台用户详情。
func (h *Handler) getUser(c *gin.Context) {
	result, err := h.users.Get(idParam(c, "id"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// createUser 创建后台用户。
func (h *Handler) createUser(c *gin.Context) {
	var input users.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.users.Create(input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// updateUser 更新后台用户。
func (h *Handler) updateUser(c *gin.Context) {
	var input users.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.users.Update(idParam(c, "id"), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// enableUser 启用后台用户。
func (h *Handler) enableUser(c *gin.Context) {
	if err := h.users.Enable(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"enabled": true})
}

// disableUser 停用后台用户。
func (h *Handler) disableUser(c *gin.Context) {
	if err := h.users.Disable(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"disabled": true})
}

// setUserAvatar 将用户头像保存到数据库。
func (h *Handler) setUserAvatar(c *gin.Context) {
	file, err := c.FormFile("avatar")
	if err != nil {
		httperr.Respond(c, httperr.New(httperr.ValidationFailed, "avatar is required"))
		return
	}
	opened, err := file.Open()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	defer opened.Close()
	data, err := io.ReadAll(io.LimitReader(opened, users.MaxAvatarBytes+1))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	contentType := users.DetectContentType(data, file.Header.Get("Content-Type"))
	if err := h.users.SetAvatar(idParam(c, "id"), data, contentType); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"uploaded": true})
}

// getUserAvatar 读取数据库中的用户头像。
func (h *Handler) getUserAvatar(c *gin.Context) {
	data, contentType, err := h.users.GetAvatar(idParam(c, "id"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	c.Data(http.StatusOK, contentType, data)
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

// listNodes 查询节点列表（分页）。
func (h *Handler) listNodes(c *gin.Context) {
	query := nodes.ListQuery{
		Keyword:  c.Query("keyword"),
		NodeType: c.Query("node_type"),
		Page:     intQuery(c, "page", 1),
		PageSize: intQuery(c, "page_size", 20),
	}
	if assigned := c.Query("assigned"); assigned != "" {
		b := assigned == "true"
		query.Assigned = &b
	}
	result, err := h.nodes.List(query)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// getNode 查询节点详情。
func (h *Handler) getNode(c *gin.Context) {
	result, err := h.nodes.Get(idParam(c, "id"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// createNode 创建巡检节点。
func (h *Handler) createNode(c *gin.Context) {
	var input nodes.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.nodes.Create(input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// updateNode 更新巡检节点。
func (h *Handler) updateNode(c *gin.Context) {
	var input nodes.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.nodes.Update(idParam(c, "id"), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// deleteNode 删除巡检节点。
func (h *Handler) deleteNode(c *gin.Context) {
	if err := h.nodes.Delete(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"deleted": true})
}

// listUnassignedNodes 查询未分配的节点列表（供模板选择）。
func (h *Handler) listUnassignedNodes(c *gin.Context) {
	result, err := h.nodes.ListUnassigned()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
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

// listWorkflows 查询工作流列表。
func (h *Handler) listWorkflows(c *gin.Context) {
	result, err := h.workflows.List(c.Query("keyword"), c.Query("status"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// createWorkflow 创建工作流。
func (h *Handler) createWorkflow(c *gin.Context) {
	var input workflows.WorkflowInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.workflows.Create(auth.UserID(c), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// getWorkflow 获取工作流详情（包含步骤）。
func (h *Handler) getWorkflow(c *gin.Context) {
	result, err := h.workflows.GetDetail(idParam(c, "id"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// updateWorkflow 更新工作流基本信息。
func (h *Handler) updateWorkflow(c *gin.Context) {
	var input workflows.WorkflowInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.workflows.Update(idParam(c, "id"), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// publishWorkflow 发布工作流。
func (h *Handler) publishWorkflow(c *gin.Context) {
	if err := h.workflows.Publish(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"status": workflows.StatusPublished})
}

// unpublishWorkflow 取消发布工作流。
func (h *Handler) unpublishWorkflow(c *gin.Context) {
	if err := h.workflows.Unpublish(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"status": workflows.StatusDraft})
}

// deleteWorkflow 删除工作流。
func (h *Handler) deleteWorkflow(c *gin.Context) {
	if err := h.workflows.Delete(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"deleted": true})
}

// addWorkflowStep 添加步骤。
func (h *Handler) addWorkflowStep(c *gin.Context) {
	var input workflows.StepInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.workflows.AddStep(idParam(c, "id"), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// updateWorkflowStep 更新步骤。
func (h *Handler) updateWorkflowStep(c *gin.Context) {
	var input workflows.StepInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.workflows.UpdateStep(idParam(c, "stepId"), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// deleteWorkflowStep 删除步骤。
func (h *Handler) deleteWorkflowStep(c *gin.Context) {
	if err := h.workflows.DeleteStep(idParam(c, "stepId")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"deleted": true})
}

// duplicateWorkflowStep 复制步骤。
func (h *Handler) duplicateWorkflowStep(c *gin.Context) {
	result, err := h.workflows.DuplicateStep(idParam(c, "stepId"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// reorderWorkflowSteps 重新排序步骤。
func (h *Handler) reorderWorkflowSteps(c *gin.Context) {
	var body struct {
		StepIDs []uint64 `json:"step_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httperr.Respond(c, err)
		return
	}
	if err := h.workflows.ReorderSteps(idParam(c, "id"), body.StepIDs); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"reordered": true})
}

// listInspectionPoints 查询巡检点位列表（分页）。
func (h *Handler) listInspectionPoints(c *gin.Context) {
	result, err := h.points.List(points.ListQuery{
		Keyword:  c.Query("keyword"),
		Page:     intQuery(c, "page", 1),
		PageSize: intQuery(c, "page_size", 20),
	})
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// listAllInspectionPoints 查询所有启用的巡检点位（精简字段，用于下拉选择）。
func (h *Handler) listAllInspectionPoints(c *gin.Context) {
	result, err := h.points.ListAll()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// getInspectionPoint 查询巡检点位详情。
func (h *Handler) getInspectionPoint(c *gin.Context) {
	result, err := h.points.Get(idParam(c, "id"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// createInspectionPoint 创建巡检点位。
func (h *Handler) createInspectionPoint(c *gin.Context) {
	var input points.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.points.Create(input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// updateInspectionPoint 更新巡检点位。
func (h *Handler) updateInspectionPoint(c *gin.Context) {
	var input points.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.points.Update(idParam(c, "id"), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// enableInspectionPoint 启用巡检点位。
func (h *Handler) enableInspectionPoint(c *gin.Context) {
	if err := h.points.Enable(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"enabled": true})
}

// disableInspectionPoint 停用巡检点位。
func (h *Handler) disableInspectionPoint(c *gin.Context) {
	if err := h.points.Disable(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"disabled": true})
}

// deleteInspectionPoint 删除巡检点位。
func (h *Handler) deleteInspectionPoint(c *gin.Context) {
	if err := h.points.Delete(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"deleted": true})
}
