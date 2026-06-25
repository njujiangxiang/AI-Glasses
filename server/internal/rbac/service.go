// Package rbac 提供服务端权限判断。菜单可见性只是前端体验，敏感接口必须在这里校验。
package rbac

import (
	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"gorm.io/gorm"
)

const MonitorViewPerm = "system:monitor:view"

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// CanViewMonitor 校验当前用户是否拥有实时监控查看权限。
//
// 权限完全由 system:monitor:view 控制（见 handlers.go 中 /monitoring/logs/recent
// 的注释："权限由 system:monitor:view 单独控制"）。这里不再叠加角色码白名单：
// 角色码在历史数据中可能漂移（例如旧版 seed 写入的 role_1、init_rbac 写入的
// super_admin、新版 seed 写入的 admin 不一致），白名单会让"已分配权限"的用户被误拒。
func (s *Service) CanViewMonitor(userID uint64) error {
	if s == nil || s.db == nil {
		return httperr.New(httperr.AuthForbidden, "无权查看实时监控")
	}
	var user database.User
	if err := s.db.Where("id = ? AND status = ?", userID, "active").First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return httperr.New(httperr.AuthForbidden, "无权查看实时监控")
		}
		return err
	}
	if user.Role == 0 {
		return httperr.New(httperr.AuthForbidden, "无权查看实时监控")
	}
	var role database.Role
	if err := s.db.Where("id = ? AND status = ?", user.Role, "active").First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return httperr.New(httperr.AuthForbidden, "无权查看实时监控")
		}
		return err
	}

	var count int64
	err := s.db.Table("role_permissions").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", role.ID).
		Where("permissions.status = ?", "active").
		Where("permissions.perms = ? OR permissions.code = ?", MonitorViewPerm, MonitorViewPerm).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return httperr.New(httperr.AuthForbidden, "无权查看实时监控")
	}
	return nil
}
