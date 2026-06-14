package organizations

import (
	"regexp"
	"strings"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"gorm.io/gorm"
)

const (
	StatusActive   = "active"
	StatusDisabled = "disabled"
)

var codePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

type CreateInput struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	ParentCode string `json:"parent_code"`
	Status     string `json:"status"`
}

type UpdateInput struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	ParentCode string `json:"parent_code"`
	Status     string `json:"status"`
}

type TreeNode struct {
	database.Organization
	Children []TreeNode `json:"children"`
}

type Service struct {
	db *gorm.DB
}

// NewService 创建组织管理服务，注入数据库访问能力。
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// Create 创建单位组织，并校验编码、名称和父级关系。
func (s *Service) Create(input CreateInput) (database.Organization, error) {
	org := database.Organization{Code: strings.TrimSpace(input.Code), Name: strings.TrimSpace(input.Name), ParentCode: strings.TrimSpace(input.ParentCode), Status: normalizeStatus(input.Status)}
	if err := s.validateOrg(org, 0); err != nil {
		return database.Organization{}, err
	}
	return org, s.db.Create(&org).Error
}

// Update 更新单位组织基础信息和父级编码。
func (s *Service) Update(id uint64, input UpdateInput) (database.Organization, error) {
	var org database.Organization
	if err := s.db.First(&org, id).Error; err != nil {
		return database.Organization{}, notFound(err, "organization not found")
	}
	org.Code = strings.TrimSpace(input.Code)
	org.Name = strings.TrimSpace(input.Name)
	org.ParentCode = strings.TrimSpace(input.ParentCode)
	org.Status = normalizeStatus(input.Status)
	if err := s.validateOrg(org, id); err != nil {
		return database.Organization{}, err
	}
	return org, s.db.Save(&org).Error
}

// List 查询单位组织列表，支持按编码或名称模糊过滤。
func (s *Service) List(keyword string) ([]database.Organization, error) {
	var orgs []database.Organization
	query := s.db.Order("parent_code asc, code asc")
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	return orgs, query.Find(&orgs).Error
}

// Tree 查询单位组织树，供后台选择上级单位和所属单位使用。
func (s *Service) Tree() ([]TreeNode, error) {
	orgs, err := s.List("")
	if err != nil {
		return nil, err
	}
	nodes := make(map[string]*TreeNode, len(orgs))
	for _, org := range orgs {
		item := TreeNode{Organization: org}
		nodes[org.Code] = &item
	}
	roots := make([]TreeNode, 0)
	for _, org := range orgs {
		node := nodes[org.Code]
		if org.ParentCode != "" {
			if parent := nodes[org.ParentCode]; parent != nil {
				parent.Children = append(parent.Children, *node)
				continue
			}
		}
		roots = append(roots, *node)
	}
	return roots, nil
}

// Enable 启用单位组织。
func (s *Service) Enable(id uint64) error { return s.setStatus(id, StatusActive) }

// Disable 停用单位组织。
func (s *Service) Disable(id uint64) error { return s.setStatus(id, StatusDisabled) }

// Delete 删除无下级组织且无用户引用的单位组织。
func (s *Service) Delete(id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var org database.Organization
		if err := tx.First(&org, id).Error; err != nil {
			return notFound(err, "organization not found")
		}
		var children int64
		if err := tx.Model(&database.Organization{}).Where("parent_code = ?", org.Code).Count(&children).Error; err != nil {
			return err
		}
		if children > 0 {
			return httperr.New(httperr.ValidationFailed, "organization has children")
		}
		var users int64
		if err := tx.Model(&database.User{}).Where("org_code = ?", org.Code).Count(&users).Error; err != nil {
			return err
		}
		if users > 0 {
			return httperr.New(httperr.ValidationFailed, "organization has users")
		}
		return tx.Delete(&org).Error
	})
}

func (s *Service) setStatus(id uint64, status string) error {
	result := s.db.Model(&database.Organization{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "organization not found")
	}
	return nil
}

func (s *Service) validateOrg(org database.Organization, currentID uint64) error {
	if !codePattern.MatchString(org.Code) {
		return httperr.New(httperr.ValidationFailed, "invalid organization code")
	}
	if org.Name == "" {
		return httperr.New(httperr.ValidationFailed, "organization name is required")
	}
	if org.Status != StatusActive && org.Status != StatusDisabled {
		return httperr.New(httperr.ValidationFailed, "invalid organization status")
	}
	var same database.Organization
	if err := s.db.Where("code = ?", org.Code).First(&same).Error; err == nil && same.ID != currentID {
		return httperr.New(httperr.ValidationFailed, "organization code already exists")
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if org.ParentCode == "" {
		return nil
	}
	if org.ParentCode == org.Code {
		return httperr.New(httperr.ValidationFailed, "parent organization cannot be itself")
	}
	return s.ensureParent(org.Code, org.ParentCode)
}

func (s *Service) ensureParent(code, parentCode string) error {
	seen := map[string]bool{code: true}
	for depth := 0; parentCode != "" && depth < 64; depth++ {
		if seen[parentCode] {
			return httperr.New(httperr.ValidationFailed, "organization parent cycle detected")
		}
		seen[parentCode] = true
		var parent database.Organization
		if err := s.db.Where("code = ?", parentCode).First(&parent).Error; err != nil {
			return notFound(err, "parent organization not found")
		}
		parentCode = parent.ParentCode
	}
	if parentCode != "" {
		return httperr.New(httperr.ValidationFailed, "organization tree is too deep")
	}
	return nil
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
