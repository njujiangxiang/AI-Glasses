package roles

import (
	"strconv"
	"strings"
	"time"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"gorm.io/gorm"
)

const (
	StatusActive   = "active"
	StatusDisabled = "disabled"

	MenuTypeDir    = "M"
	MenuTypeMenu   = "C"
	MenuTypeAction = "A"
)

type ListQuery struct {
	Keyword  string
	Page     int
	PageSize int
}

type CreateInput struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	DataScope   string `json:"data_scope"`
	Sort        int    `json:"sort"`
	Status      string `json:"status"`
	MenuIDs     string `json:"menu_ids"`
}

type UpdateInput struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	DataScope   string `json:"data_scope"`
	Sort        int    `json:"sort"`
	Status      string `json:"status"`
	MenuIDs     string `json:"menu_ids"`
}

type RoleDTO struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	DataScope   string    `json:"data_scope"`
	Sort        int       `json:"sort"`
	Status      string    `json:"status"`
	MemberCount int64     `json:"member_count"`
	Menus       []uint64  `json:"menus"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListResult struct {
	Items []RoleDTO `json:"items"`
	Total int64     `json:"total"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// List 查询角色列表
func (s *Service) List(query ListQuery) (ListResult, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	db := s.db.Model(&database.Role{})
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", like, like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return ListResult{}, err
	}
	var models []database.Role
	if err := db.Order("sort asc, id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&models).Error; err != nil {
		return ListResult{}, err
	}
	items := make([]RoleDTO, 0, len(models))
	for _, role := range models {
		items = append(items, s.toDTO(role))
	}
	return ListResult{Items: items, Total: total}, nil
}

// ListAll 查询所有启用的角色
func (s *Service) ListAll() ([]RoleDTO, error) {
	var models []database.Role
	if err := s.db.Where("status = ?", StatusActive).Order("sort asc, id desc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]RoleDTO, 0, len(models))
	for _, role := range models {
		items = append(items, s.toDTO(role))
	}
	return items, nil
}

// Get 查询角色详情
func (s *Service) Get(id uint64) (RoleDTO, error) {
	var role database.Role
	if err := s.db.First(&role, id).Error; err != nil {
		return RoleDTO{}, notFound(err, "role not found")
	}
	return s.toDTO(role), nil
}

// Create 创建角色
func (s *Service) Create(input CreateInput) (RoleDTO, error) {
	role := database.Role{
		Name:        strings.TrimSpace(input.Name),
		Code:        strings.TrimSpace(input.Code),
		Description: strings.TrimSpace(input.Description),
		DataScope:   normalizeDataScope(input.DataScope),
		Sort:        input.Sort,
		Status:      normalizeStatus(input.Status),
	}
	if err := s.validateRole(role, 0); err != nil {
		return RoleDTO{}, err
	}
	tx := s.db.Begin()
	if err := tx.Create(&role).Error; err != nil {
		tx.Rollback()
		return RoleDTO{}, err
	}
	if err := s.updateRoleMenus(tx, role.ID, input.MenuIDs); err != nil {
		tx.Rollback()
		return RoleDTO{}, err
	}
	if err := tx.Commit().Error; err != nil {
		return RoleDTO{}, err
	}
	return s.toDTO(role), nil
}

// Update 更新角色
func (s *Service) Update(id uint64, input UpdateInput) (RoleDTO, error) {
	var role database.Role
	if err := s.db.First(&role, id).Error; err != nil {
		return RoleDTO{}, notFound(err, "role not found")
	}
	role.Name = strings.TrimSpace(input.Name)
	role.Code = strings.TrimSpace(input.Code)
	role.Description = strings.TrimSpace(input.Description)
	if input.DataScope != "" {
		role.DataScope = input.DataScope
	}
	role.Sort = input.Sort
	if input.Status != "" {
		role.Status = input.Status
	}
	if err := s.validateRole(role, id); err != nil {
		return RoleDTO{}, err
	}
	tx := s.db.Begin()
	if err := tx.Save(&role).Error; err != nil {
		tx.Rollback()
		return RoleDTO{}, err
	}
	if input.MenuIDs != "" {
		if err := s.updateRoleMenus(tx, id, input.MenuIDs); err != nil {
			tx.Rollback()
			return RoleDTO{}, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return RoleDTO{}, err
	}
	return s.toDTO(role), nil
}

// Delete 删除角色
func (s *Service) Delete(id uint64) error {
	var count int64
	s.db.Model(&database.User{}).Where("role_id = ?", id).Count(&count)
	if count > 0 {
		return httperr.New(httperr.ValidationFailed, "该角色下还有用户，无法删除")
	}
	tx := s.db.Begin()
	if err := tx.Where("role_id = ?", id).Delete(&database.RolePermission{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&database.Role{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// UpdateMenus 更新角色菜单权限
func (s *Service) UpdateMenus(id uint64, menuIDs string) error {
	var role database.Role
	if err := s.db.First(&role, id).Error; err != nil {
		return notFound(err, "role not found")
	}
	return s.updateRoleMenus(s.db, id, menuIDs)
}

// updateRoleMenus 更新角色权限关联
func (s *Service) updateRoleMenus(tx *gorm.DB, roleID uint64, menuIDs string) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&database.RolePermission{}).Error; err != nil {
		return err
	}
	if menuIDs == "" {
		return nil
	}
	ids := strings.Split(menuIDs, ",")
	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		menuID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || menuID == 0 {
			continue
		}
		if err := tx.Create(&database.RolePermission{RoleID: roleID, PermissionID: menuID}).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetRoleMenuIDs 获取角色的菜单权限ID列表
func (s *Service) GetRoleMenuIDs(roleID uint64) ([]uint64, error) {
	var rp []database.RolePermission
	if err := s.db.Where("role_id = ?", roleID).Find(&rp).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(rp))
	for _, item := range rp {
		ids = append(ids, item.PermissionID)
	}
	return ids, nil
}

func (s *Service) validateRole(role database.Role, currentID uint64) error {
	if role.Name == "" {
		return httperr.New(httperr.ValidationFailed, "角色名称不能为空")
	}
	if role.Status != StatusActive && role.Status != StatusDisabled {
		return httperr.New(httperr.ValidationFailed, "无效的角色状态")
	}
	var same database.Role
	if err := s.db.Select("id").Where("name = ?", role.Name).First(&same).Error; err == nil && same.ID != currentID {
		return httperr.New(httperr.ValidationFailed, "角色名称已存在")
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if role.Code != "" {
		if err := s.db.Select("id").Where("code = ?", role.Code).First(&same).Error; err == nil && same.ID != currentID {
			return httperr.New(httperr.ValidationFailed, "角色编码已存在")
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
	}
	return nil
}

func (s *Service) toDTO(role database.Role) RoleDTO {
	var memberCount int64
	s.db.Model(&database.User{}).Where("role_id = ?", role.ID).Count(&memberCount)
	menus, _ := s.GetRoleMenuIDs(role.ID)
	return RoleDTO{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		DataScope:   role.DataScope,
		Sort:        role.Sort,
		Status:      role.Status,
		MemberCount: memberCount,
		Menus:       menus,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func normalizeStatus(status string) string {
	if status == "" {
		return StatusActive
	}
	return status
}

func normalizeDataScope(dataScope string) string {
	if dataScope == "" {
		return database.DataScopeOrgOnly
	}
	// 验证是否是合法值
	switch dataScope {
	case database.DataScopeAll, database.DataScopeOrgAndSub, database.DataScopeOrgOnly, database.DataScopeSelfOnly:
		return dataScope
	default:
		return database.DataScopeOrgOnly
	}
}

func notFound(err error, message string) error {
	if err == gorm.ErrRecordNotFound {
		return httperr.New(httperr.ResourceNotFound, message)
	}
	return err
}
