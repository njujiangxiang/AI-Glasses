package menus

import (
	"strings"
	"time"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"gorm.io/gorm"
)

const (
	StatusActive   = "active"
	StatusDisabled = "disabled"

	TypeDir    = "M"
	TypeMenu   = "C"
	TypeAction = "A"
)

type CreateInput struct {
	Pid       uint64 `json:"pid"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Icon      string `json:"icon"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Sort      int    `json:"sort"`
	Perms     string `json:"perms"`
	Visible   bool   `json:"visible"`
	Status    string `json:"status"`
	IsCache   bool   `json:"is_cache"`
}

type UpdateInput struct {
	Pid       uint64 `json:"pid"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Icon      string `json:"icon"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Sort      int    `json:"sort"`
	Perms     string `json:"perms"`
	Visible   bool   `json:"visible"`
	Status    string `json:"status"`
	IsCache   bool   `json:"is_cache"`
}

type MenuDTO struct {
	ID        uint64    `json:"id"`
	Pid       uint64    `json:"pid"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Icon      string    `json:"icon"`
	Path      string    `json:"path"`
	Component string    `json:"component"`
	Sort      int       `json:"sort"`
	Perms     string    `json:"perms"`
	Visible   bool      `json:"visible"`
	Status    string    `json:"status"`
	IsCache   bool      `json:"is_cache"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Children  []MenuDTO `json:"children,omitempty"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// ListTree 获取菜单树
func (s *Service) ListTree() ([]MenuDTO, error) {
	var models []database.Permission
	if err := s.db.Order("sort asc, id asc").Find(&models).Error; err != nil {
		return nil, err
	}
	menus := make([]MenuDTO, 0, len(models))
	for _, m := range models {
		menus = append(menus, toDTO(m))
	}
	return buildTree(menus, 0), nil
}

// ListAll 获取所有菜单列表
func (s *Service) ListAll() ([]MenuDTO, error) {
	var models []database.Permission
	if err := s.db.Order("sort asc, id asc").Find(&models).Error; err != nil {
		return nil, err
	}
	menus := make([]MenuDTO, 0, len(models))
	for _, m := range models {
		menus = append(menus, toDTO(m))
	}
	return menus, nil
}

// Get 获取菜单详情
func (s *Service) Get(id uint64) (MenuDTO, error) {
	var menu database.Permission
	if err := s.db.First(&menu, id).Error; err != nil {
		return MenuDTO{}, notFound(err, "menu not found")
	}
	return toDTO(menu), nil
}

// Create 创建菜单
func (s *Service) Create(input CreateInput) (MenuDTO, error) {
	menu := database.Permission{
		Pid:       input.Pid,
		Type:      strings.TrimSpace(input.Type),
		Name:      strings.TrimSpace(input.Name),
		Code:      strings.TrimSpace(input.Code),
		Icon:      strings.TrimSpace(input.Icon),
		Path:      strings.TrimSpace(input.Path),
		Component: strings.TrimSpace(input.Component),
		Sort:      input.Sort,
		Perms:     strings.TrimSpace(input.Perms),
		Visible:   input.Visible,
		Status:    normalizeStatus(input.Status),
		IsCache:   input.IsCache,
	}
	if menu.Type == "" {
		menu.Type = TypeMenu
	}
	if err := s.validateMenu(menu, 0); err != nil {
		return MenuDTO{}, err
	}
	if err := s.db.Create(&menu).Error; err != nil {
		return MenuDTO{}, err
	}
	return toDTO(menu), nil
}

// Update 更新菜单
func (s *Service) Update(id uint64, input UpdateInput) (MenuDTO, error) {
	var menu database.Permission
	if err := s.db.First(&menu, id).Error; err != nil {
		return MenuDTO{}, notFound(err, "menu not found")
	}
	menu.Pid = input.Pid
	menu.Type = strings.TrimSpace(input.Type)
	menu.Name = strings.TrimSpace(input.Name)
	menu.Code = strings.TrimSpace(input.Code)
	menu.Icon = strings.TrimSpace(input.Icon)
	menu.Path = strings.TrimSpace(input.Path)
	menu.Component = strings.TrimSpace(input.Component)
	menu.Sort = input.Sort
	menu.Perms = strings.TrimSpace(input.Perms)
	menu.Visible = input.Visible
	menu.IsCache = input.IsCache
	if input.Status != "" {
		menu.Status = input.Status
	}
	if err := s.validateMenu(menu, id); err != nil {
		return MenuDTO{}, err
	}
	if err := s.db.Save(&menu).Error; err != nil {
		return MenuDTO{}, err
	}
	return toDTO(menu), nil
}

// Delete 删除菜单
func (s *Service) Delete(id uint64) error {
	var count int64
	s.db.Model(&database.Permission{}).Where("pid = ?", id).Count(&count)
	if count > 0 {
		return httperr.New(httperr.ValidationFailed, "该菜单下有子菜单，请先删除子菜单")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("permission_id = ?", id).Delete(&database.RolePermission{}).Error; err != nil {
			return err
		}
		return tx.Delete(&database.Permission{}, id).Error
	})
}

// GetUserMenus 获取用户的菜单权限树
func (s *Service) GetUserMenus(userID uint64) ([]MenuDTO, error) {
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, notFound(err, "user not found")
	}
	if user.RoleID == 0 {
		return []MenuDTO{}, nil
	}
	var rp []database.RolePermission
	if err := s.db.Where("role_id = ?", user.RoleID).Find(&rp).Error; err != nil {
		return nil, err
	}
	permIDs := make([]uint64, 0, len(rp))
	for _, item := range rp {
		permIDs = append(permIDs, item.PermissionID)
	}
	if len(permIDs) == 0 {
		return []MenuDTO{}, nil
	}
	// 获取所有权限菜单，包括父级菜单以构建完整的树结构
	var allMenus []database.Permission
	if err := s.db.Order("sort asc, id asc").Find(&allMenus).Error; err != nil {
		return nil, err
	}
	// 构建用户有权限的菜单集合
	permSet := make(map[uint64]bool)
	for _, id := range permIDs {
		permSet[id] = true
	}
	// 递归收集所有父级菜单ID，确保树结构完整
	allMenuMap := make(map[uint64]database.Permission)
	for _, m := range allMenus {
		allMenuMap[m.ID] = m
	}
	// 收集所有需要包含的菜单（有权限的菜单及其所有父级）
	includeIDs := make(map[uint64]bool)
	for _, id := range permIDs {
		collectAncestors(id, allMenuMap, includeIDs)
	}
	// 过滤出需要显示的菜单
	var models []database.Permission
	for _, m := range allMenus {
		if includeIDs[m.ID] && m.Visible && m.Status == StatusActive {
			models = append(models, m)
		}
	}
	menus := make([]MenuDTO, 0, len(models))
	for _, m := range models {
		menus = append(menus, toDTO(m))
	}
	return buildTree(menus, 0), nil
}

// collectAncestors 递归收集所有父级菜单ID
func collectAncestors(id uint64, menuMap map[uint64]database.Permission, includeIDs map[uint64]bool) {
	if id == 0 || includeIDs[id] {
		return
	}
	includeIDs[id] = true
	if menu, ok := menuMap[id]; ok && menu.Pid > 0 {
		collectAncestors(menu.Pid, menuMap, includeIDs)
	}
}

func (s *Service) validateMenu(menu database.Permission, currentID uint64) error {
	if menu.Name == "" {
		return httperr.New(httperr.ValidationFailed, "菜单名称不能为空")
	}
	if menu.Type != TypeDir && menu.Type != TypeMenu && menu.Type != TypeAction {
		return httperr.New(httperr.ValidationFailed, "无效的菜单类型")
	}
	if menu.Status != StatusActive && menu.Status != StatusDisabled {
		return httperr.New(httperr.ValidationFailed, "无效的菜单状态")
	}
	if menu.Pid > 0 {
		var parent database.Permission
		if err := s.db.First(&parent, menu.Pid).Error; err != nil {
			return notFound(err, "父菜单不存在")
		}
	}
	var same database.Permission
	if menu.Code != "" {
		if err := s.db.Select("id").Where("code = ?", menu.Code).First(&same).Error; err == nil && same.ID != currentID {
			return httperr.New(httperr.ValidationFailed, "菜单编码已存在")
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
	}
	return nil
}

func buildTree(menus []MenuDTO, pid uint64) []MenuDTO {
	tree := make([]MenuDTO, 0)
	for _, m := range menus {
		if m.Pid == pid {
			m.Children = buildTree(menus, m.ID)
			tree = append(tree, m)
		}
	}
	return tree
}

func toDTO(menu database.Permission) MenuDTO {
	return MenuDTO{
		ID:        menu.ID,
		Pid:       menu.Pid,
		Type:      menu.Type,
		Name:      menu.Name,
		Code:      menu.Code,
		Icon:      menu.Icon,
		Path:      menu.Path,
		Component: menu.Component,
		Sort:      menu.Sort,
		Perms:     menu.Perms,
		Visible:   menu.Visible,
		Status:    menu.Status,
		IsCache:   menu.IsCache,
		CreatedAt: menu.CreatedAt,
		UpdatedAt: menu.UpdatedAt,
	}
}

func normalizeStatus(status string) string {
	if status == "" {
		return StatusActive
	}
	return status
}

func notFound(err error, message string) error {
	if err == gorm.ErrRecordNotFound {
		return httperr.New(httperr.ResourceNotFound, message)
	}
	return err
}
