package httpapi

import (
	"strconv"

	"aiglasses/server/internal/auth"
	"aiglasses/server/internal/monitoring"
	"aiglasses/server/internal/platform/httperr"
	"github.com/gin-gonic/gin"
)

func (h *Handler) monitorViewRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.rbac == nil {
			httperr.Respond(c, httperr.New(httperr.AuthForbidden, "无权查看实时监控"))
			c.Abort()
			return
		}
		if err := h.rbac.CanViewMonitor(auth.UserID(c)); err != nil {
			httperr.Respond(c, err)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *Handler) recentMonitorLogs(c *gin.Context) {
	if h.monitoringHub == nil {
		h.monitoringHub = monitoring.NewHub()
	}
	limit := monitoring.DefaultLimit
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}
	afterID := uint64(0)
	if raw := c.Query("after_id"); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			httperr.Respond(c, httperr.New(httperr.ValidationFailed, "after_id 必须是非负整数"))
			return
		}
		afterID = parsed
	}
	httperr.OK(c, h.monitoringHub.Recent(limit, afterID))
}
