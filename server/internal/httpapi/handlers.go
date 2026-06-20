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
	"aiglasses/server/internal/datascope"
	"aiglasses/server/internal/defects"
	"aiglasses/server/internal/devices"
	"aiglasses/server/internal/events"
	"aiglasses/server/internal/menus"
	"aiglasses/server/internal/monitoring"
	"aiglasses/server/internal/organizations"
	"aiglasses/server/internal/plans"
	"aiglasses/server/internal/platform/httperr"
	"aiglasses/server/internal/rbac"
	"aiglasses/server/internal/roles"
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
	datascope     *datascope.Service
	defects       *defects.Service
	devices       *devices.Service
	menus         *menus.Service
	monitoringHub *monitoring.Hub
	organizations *organizations.Service
	plans         *plans.Service
	rbac          *rbac.Service
	roles         *roles.Service
	tasks         *tasks.Service
	templates     *templates.Service
	users         *users.Service
	workflows     *workflows.Service
	scheduler     *events.Scheduler
}

// NewHandler 创建 HTTP 处理器集合，并注入所有业务服务。
func NewHandler(authSvc *auth.Service, attachmentSvc *attachments.Service, businessCodeSvc *businesscodes.Service, datascopeSvc *datascope.Service, defectSvc *defects.Service, deviceSvc *devices.Service, menuSvc *menus.Service, orgSvc *organizations.Service, planSvc *plans.Service, roleSvc *roles.Service, taskSvc *tasks.Service, templateSvc *templates.Service, userSvc *users.Service, workflowSvc *workflows.Service, scheduler *events.Scheduler, rbacSvc *rbac.Service, monitoringHub *monitoring.Hub) *Handler {
	return &Handler{auth: authSvc, attachments: attachmentSvc, businessCodes: businessCodeSvc, datascope: datascopeSvc, defects: defectSvc, devices: deviceSvc, menus: menuSvc, monitoringHub: monitoringHub, organizations: orgSvc, plans: planSvc, rbac: rbacSvc, roles: roleSvc, tasks: taskSvc, templates: templateSvc, users: userSvc, workflows: workflowSvc, scheduler: scheduler}
}

// DataScopeMiddleware 数据范围过滤中间件，将当前用户的数据范围信息注入上下文
func DataScopeMiddleware(datascopeSvc *datascope.Service, orgSvc *organizations.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := auth.UserID(c)
		// 使用带完整组织树查询的方法
		scope, err := datascopeSvc.GetUserScopeWithOrgCodes(userID, orgSvc.GetSubOrgCodes)
		if err != nil {
			httperr.Respond(c, err)
			c.Abort()
			return
		}
		c.Set("datascope", scope)
		c.Next()
	}
}

// DataScope 从 Gin 上下文中读取数据范围信息
func DataScope(c *gin.Context) *datascope.ScopeInfo {
	value, exists := c.Get("datascope")
	if !exists {
		return nil
	}
	return value.(*datascope.ScopeInfo)
}

// Register 注册公开接口、后台接口和眼镜端接口。
func (h *Handler) Register(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	api := r.Group("/api")
	api.POST("/admin/auth/login", h.adminLogin)
	api.POST("/glasses/auth/token", h.glassesLogin)
	adminAuth := api.Group("/admin", auth.Middleware(h.auth, auth.ScopeAdmin))
	adminAuth.GET("/users/me", h.currentUser)
	adminAuth.POST("/users/me/update", h.updateCurrentUser)
	// 实时日志属于系统运维能力，不走业务数据范围过滤；权限由 system:monitor:view 单独控制。
	adminAuth.GET("/monitoring/logs/recent", h.monitorViewRequired(), h.recentMonitorLogs)
	admin := adminAuth.Group("", DataScopeMiddleware(h.datascope, h.organizations))
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
	admin.GET("/users/:id", h.getUser)
	admin.POST("/users", h.createUser)
	admin.POST("/users/:id/update", h.updateUser)
	admin.POST("/users/:id/enable", h.enableUser)
	admin.POST("/users/:id/disable", h.disableUser)
	admin.POST("/users/:id/avatar", h.setUserAvatar)
	admin.GET("/users/:id/avatar", h.getUserAvatar)
	admin.GET("/roles", h.listRoles)
	admin.GET("/roles/all", h.listAllRoles)
	admin.GET("/roles/:id", h.getRole)
	admin.POST("/roles", h.createRole)
	admin.POST("/roles/:id/update", h.updateRole)
	admin.POST("/roles/:id/menus", h.updateRoleMenus)
	admin.POST("/roles/:id/delete", h.deleteRole)
	admin.GET("/menus", h.listMenus)
	admin.GET("/menus/tree", h.menuTree)
	admin.GET("/menus/mine", h.myMenus)
	admin.GET("/menus/:id", h.getMenu)
	admin.POST("/menus", h.createMenu)
	admin.POST("/menus/:id/update", h.updateMenu)
	admin.POST("/menus/:id/delete", h.deleteMenu)
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
	token, user, err := h.auth.Login(body.Username, body.Password, auth.ScopeGlasses, &deviceID)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"access_token": token, "user": user, "device_id": deviceID})
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
	code, err := h.businessCodes.GenerateDaily(c.Request.Context(), body.Code)
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

// adminTasks 查询后台任务列表（带数据范围过滤）。
func (h *Handler) adminTasks(c *gin.Context) {
	scope := DataScope(c)
	result, err := h.tasks.AdminListWithScope(c.Query("status"), intQuery(c, "limit", 50), scope)
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

// listUsers 查询后台用户列表（带数据范围过滤）。
func (h *Handler) listUsers(c *gin.Context) {
	scope := DataScope(c)
	result, err := h.users.ListWithScope(users.ListQuery{
		Keyword:  c.Query("keyword"),
		OrgCode:  c.Query("org_code"),
		Status:   c.Query("status"),
		Page:     intQuery(c, "page", 1),
		PageSize: intQuery(c, "page_size", 20),
	}, scope)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// currentUser 查询当前登录用户详情，供个人中心和顶部用户信息使用。
func (h *Handler) currentUser(c *gin.Context) {
	result, err := h.currentUserPayload(auth.UserID(c))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// updateCurrentUser 更新当前登录用户的个人资料，不修改组织、角色、状态等管理员字段。
func (h *Handler) updateCurrentUser(c *gin.Context) {
	var input users.ProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	if _, err := h.users.UpdateProfile(auth.UserID(c), input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.currentUserPayload(auth.UserID(c))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

func (h *Handler) currentUserPayload(userID uint64) (gin.H, error) {
	user, err := h.users.Get(userID)
	if err != nil {
		return nil, err
	}
	orgName, err := h.auth.OrganizationName(user.OrgCode)
	if err != nil {
		return nil, err
	}
	return gin.H{"user": user, "org_name": orgName, "company_name": orgName}, nil
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

// listRoles 查询角色列表。
func (h *Handler) listRoles(c *gin.Context) {
	result, err := h.roles.List(roles.ListQuery{
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

// listAllRoles 查询所有启用的角色。
func (h *Handler) listAllRoles(c *gin.Context) {
	result, err := h.roles.ListAll()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// getRole 查询角色详情。
func (h *Handler) getRole(c *gin.Context) {
	result, err := h.roles.Get(idParam(c, "id"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// createRole 创建角色。
func (h *Handler) createRole(c *gin.Context) {
	var input roles.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.roles.Create(input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// updateRole 更新角色。
func (h *Handler) updateRole(c *gin.Context) {
	var input roles.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.roles.Update(idParam(c, "id"), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// deleteRole 删除角色。
func (h *Handler) deleteRole(c *gin.Context) {
	if err := h.roles.Delete(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"deleted": true})
}

// updateRoleMenus 更新角色菜单权限。
func (h *Handler) updateRoleMenus(c *gin.Context) {
	var input struct {
		MenuIDs string `json:"menu_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	if err := h.roles.UpdateMenus(idParam(c, "id"), input.MenuIDs); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"success": true})
}

// listMenus 查询菜单列表。
func (h *Handler) listMenus(c *gin.Context) {
	result, err := h.menus.ListAll()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// menuTree 查询菜单树。
func (h *Handler) menuTree(c *gin.Context) {
	result, err := h.menus.ListTree()
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// getMenu 查询菜单详情。
func (h *Handler) getMenu(c *gin.Context) {
	result, err := h.menus.Get(idParam(c, "id"))
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// createMenu 创建菜单。
func (h *Handler) createMenu(c *gin.Context) {
	var input menus.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.menus.Create(input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.Created(c, result)
}

// updateMenu 更新菜单。
func (h *Handler) updateMenu(c *gin.Context) {
	var input menus.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httperr.Respond(c, err)
		return
	}
	result, err := h.menus.Update(idParam(c, "id"), input)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}

// deleteMenu 删除菜单。
func (h *Handler) deleteMenu(c *gin.Context) {
	if err := h.menus.Delete(idParam(c, "id")); err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, gin.H{"deleted": true})
}

// myMenus 获取当前用户的菜单权限树。
func (h *Handler) myMenus(c *gin.Context) {
	userID := auth.UserID(c)
	result, err := h.menus.GetUserMenus(userID)
	if err != nil {
		httperr.Respond(c, err)
		return
	}
	httperr.OK(c, result)
}
