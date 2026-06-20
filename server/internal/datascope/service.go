// Package datascope 实现基于角色的数据范围权限控制，支持全部数据、本组织及下级、仅本组织、仅自己四种范围。
package datascope

import (
	"aiglasses/server/internal/platform/database"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

// NewService 创建数据范围服务，注入数据库访问能力。
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ScopeInfo 用户的数据范围信息
type ScopeInfo struct {
	UserID    uint64   `json:"user_id"`
	OrgCode   string   `json:"org_code"`
	DataScope string   `json:"data_scope"`
	OrgCodes  []string `json:"org_codes"` // 可访问的组织编码列表
}

// GetUserScope 获取用户的数据范围信息
func (s *Service) GetUserScope(userID uint64) (*ScopeInfo, error) {
	// 1. 获取用户信息
	var user database.User
	if err := s.db.Select("org_code, role_id").First(&user, userID).Error; err != nil {
		return nil, err
	}

	// 2. 获取角色的数据范围
	var role database.Role
	if err := s.db.Select("data_scope").First(&role, user.RoleID).Error; err != nil {
		return nil, err
	}

	scope := &ScopeInfo{
		UserID:    userID,
		OrgCode:   user.OrgCode,
		DataScope: role.DataScope,
		OrgCodes:  []string{},
	}

	// 3. 根据数据范围计算可访问的组织列表
	switch role.DataScope {
	case database.DataScopeAll:
		// 全部数据 - 不限制组织
		scope.OrgCodes = nil

	case database.DataScopeOrgAndSub:
		// 本组织及下级 - 需要通过 organizations service 获取
		// 暂时先返回仅本组织，后续由 organizations service 扩展
		if user.OrgCode != "" {
			scope.OrgCodes = []string{user.OrgCode}
		}

	case database.DataScopeOrgOnly:
		// 仅本组织
		if user.OrgCode != "" {
			scope.OrgCodes = []string{user.OrgCode}
		}

	case database.DataScopeSelfOnly:
		// 仅自己 - 由各业务层自己处理 user_id 过滤
		scope.OrgCodes = []string{}
	}

	return scope, nil
}

// GetUserScopeWithOrgCodes 获取用户的数据范围信息（包含完整的子组织列表）
// 需要传入 organizations service 的 GetSubOrgCodes 方法
func (s *Service) GetUserScopeWithOrgCodes(userID uint64, getSubOrgCodes func(string) ([]string, error)) (*ScopeInfo, error) {
	// 1. 获取用户信息
	var user database.User
	if err := s.db.Select("org_code, role_id").First(&user, userID).Error; err != nil {
		return nil, err
	}

	// 2. 获取角色的数据范围
	var role database.Role
	if err := s.db.Select("data_scope").First(&role, user.RoleID).Error; err != nil {
		return nil, err
	}

	scope := &ScopeInfo{
		UserID:    userID,
		OrgCode:   user.OrgCode,
		DataScope: role.DataScope,
		OrgCodes:  []string{},
	}

	// 3. 根据数据范围计算可访问的组织列表
	switch role.DataScope {
	case database.DataScopeAll:
		// 全部数据 - 不限制组织
		scope.OrgCodes = nil

	case database.DataScopeOrgAndSub:
		// 本组织及下级
		if user.OrgCode != "" {
			orgCodes, err := getSubOrgCodes(user.OrgCode)
			if err != nil {
				return nil, err
			}
			scope.OrgCodes = orgCodes
		}

	case database.DataScopeOrgOnly:
		// 仅本组织
		if user.OrgCode != "" {
			scope.OrgCodes = []string{user.OrgCode}
		}

	case database.DataScopeSelfOnly:
		// 仅自己 - 由各业务层自己处理 user_id 过滤
		scope.OrgCodes = []string{}
	}

	return scope, nil
}

// ApplyOrgFilter 应用组织范围过滤到查询
// orgCodeColumn: 组织编码列名（如 "org_code", "users.org_code"）
func (s *Service) ApplyOrgFilter(query *gorm.DB, orgCodeColumn string, scope *ScopeInfo) *gorm.DB {
	if scope.DataScope == database.DataScopeAll {
		return query // 不限制
	}

	if scope.DataScope == database.DataScopeSelfOnly {
		// self_only 模式需要各业务自己处理 user_id 过滤
		// 这里默认不添加组织过滤，由调用者决定如何过滤
		return query
	}

	if len(scope.OrgCodes) > 0 {
		return query.Where(orgCodeColumn+" IN ?", scope.OrgCodes)
	}

	// 无权限 - 默认查询空结果
	return query.Where("1=0")
}

// IsAll 返回是否是全部数据权限
func (s *ScopeInfo) IsAll() bool {
	return s.DataScope == database.DataScopeAll
}

// IsSelfOnly 返回是否是仅自己权限
func (s *ScopeInfo) IsSelfOnly() bool {
	return s.DataScope == database.DataScopeSelfOnly
}

// GetUserID 获取用户ID
func (s *ScopeInfo) GetUserID() uint64 {
	return s.UserID
}

// GetOrgCodes 获取可访问的组织编码列表
func (s *ScopeInfo) GetOrgCodes() []string {
	return s.OrgCodes
}

// HasOrgAccess 检查是否有权访问指定组织
func (s *ScopeInfo) HasOrgAccess(orgCode string) bool {
	if s.DataScope == database.DataScopeAll {
		return true
	}
	if s.DataScope == database.DataScopeSelfOnly {
		return false
	}
	for _, code := range s.OrgCodes {
		if code == orgCode {
			return true
		}
	}
	return false
}
