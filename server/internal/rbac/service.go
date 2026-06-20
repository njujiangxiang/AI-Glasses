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

func (s *Service) CanViewMonitor(userID uint64) error {
	return s.HasRoleAndPermission(userID, []string{"admin", "super_admin"}, MonitorViewPerm)
}

func (s *Service) HasRoleAndPermission(userID uint64, roleCodes []string, perm string) error {
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
	if user.RoleID == 0 {
		return httperr.New(httperr.AuthForbidden, "无权查看实时监控")
	}
	var role database.Role
	if err := s.db.Where("id = ? AND status = ?", user.RoleID, "active").First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return httperr.New(httperr.AuthForbidden, "无权查看实时监控")
		}
		return err
	}
	if !contains(roleCodes, role.Code) {
		return httperr.New(httperr.AuthForbidden, "无权查看实时监控")
	}

	var count int64
	err := s.db.Table("role_permissions").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", role.ID).
		Where("permissions.status = ?", "active").
		Where("permissions.perms = ? OR permissions.code = ?", perm, perm).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return httperr.New(httperr.AuthForbidden, "无权查看实时监控")
	}
	return nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
