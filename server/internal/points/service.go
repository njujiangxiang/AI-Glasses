// Package points 实现巡检点位管理。点位是巡检计划的关联对象，用于标识具体的巡检位置。
package points

import (
	"strings"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"gorm.io/gorm"
)

// Option 用于下拉选择列表的精简结构。
type Option struct {
	ID            uint64 `json:"id"`
	Name          string `json:"name"`
	EquipmentName string `json:"equipment_name"`
	Location      string `json:"location"`
	Area          string `json:"area"`
	Substation    string `json:"substation"`
}

type CreateInput struct {
	Name          string `json:"name"`
	EquipmentName string `json:"equipment_name"`
	Location      string `json:"location"`
	Area          string `json:"area"`
	Substation    string `json:"substation"`
	Description   string `json:"description"`
}

type UpdateInput struct {
	Name          string `json:"name"`
	EquipmentName string `json:"equipment_name"`
	Location      string `json:"location"`
	Area          string `json:"area"`
	Substation    string `json:"substation"`
	Description   string `json:"description"`
}

type ListQuery struct {
	Keyword  string
	Page     int
	PageSize int
}

type ListResult struct {
	Items []database.InspectionPoint `json:"items"`
	Total int64                      `json:"total"`
}

type Service struct {
	db *gorm.DB
}

// NewService 创建巡检点位服务，注入数据库访问能力。
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// List 查询巡检点位列表，支持分页和关键词搜索。
func (s *Service) List(query ListQuery) (ListResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	db := s.db.Model(&database.InspectionPoint{})
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + escapeLike(keyword) + "%"
		db = db.Where("name LIKE ? OR equipment_name LIKE ? OR area LIKE ? OR substation LIKE ?", like, like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return ListResult{}, err
	}

	var items []database.InspectionPoint
	if err := db.Order("id desc").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error; err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, Total: total}, nil
}

// ListAll 查询所有巡检点位（精简字段，用于下拉选择）。
func (s *Service) ListAll() ([]Option, error) {
	var options []Option
	err := s.db.Model(&database.InspectionPoint{}).
		Where("enabled = ?", true).
		Select("id, name, equipment_name, location, area, substation").
		Order("name asc").
		Find(&options).Error
	return options, err
}

// Get 查询单个巡检点位详情。
func (s *Service) Get(id uint64) (database.InspectionPoint, error) {
	var point database.InspectionPoint
	if err := s.db.First(&point, id).Error; err != nil {
		return point, notFound(err, "inspection point not found")
	}
	return point, nil
}

// Create 创建巡检点位。
func (s *Service) Create(input CreateInput) (database.InspectionPoint, error) {
	if input.Name == "" {
		return database.InspectionPoint{}, httperr.New(httperr.TaskStateConflict, "point name is required")
	}

	point := database.InspectionPoint{
		Name:          input.Name,
		EquipmentName: input.EquipmentName,
		Location:      input.Location,
		Area:          input.Area,
		Substation:    input.Substation,
		Description:   input.Description,
		Enabled:       true,
	}
	return point, s.db.Create(&point).Error
}

// Update 更新巡检点位。
func (s *Service) Update(id uint64, input UpdateInput) (database.InspectionPoint, error) {
	var point database.InspectionPoint
	if err := s.db.First(&point, id).Error; err != nil {
		return point, notFound(err, "inspection point not found")
	}
	if input.Name == "" {
		return database.InspectionPoint{}, httperr.New(httperr.TaskStateConflict, "point name is required")
	}

	point.Name = input.Name
	point.EquipmentName = input.EquipmentName
	point.Location = input.Location
	point.Area = input.Area
	point.Substation = input.Substation
	point.Description = input.Description
	return point, s.db.Save(&point).Error
}

// Enable 启用巡检点位。
func (s *Service) Enable(id uint64) error {
	result := s.db.Model(&database.InspectionPoint{}).Where("id = ?", id).Update("enabled", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "inspection point not found")
	}
	return nil
}

// Disable 停用巡检点位。
func (s *Service) Disable(id uint64) error {
	result := s.db.Model(&database.InspectionPoint{}).Where("id = ?", id).Update("enabled", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "inspection point not found")
	}
	return nil
}

// Delete 删除巡检点位。
func (s *Service) Delete(id uint64) error {
	result := s.db.Delete(&database.InspectionPoint{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "inspection point not found")
	}
	return nil
}

func notFound(err error, message string) error {
	if err == gorm.ErrRecordNotFound {
		return httperr.New(httperr.ResourceNotFound, message)
	}
	return err
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
