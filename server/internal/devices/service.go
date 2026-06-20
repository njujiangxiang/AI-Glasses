package devices

import (
	"errors"
	"time"

	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/httperr"
	"gorm.io/gorm"
)

const (
	StatusPending      = "pending"
	StatusActive       = "active"
	StatusRevoked      = "revoked"
	StatusLostDisabled = "lost_disabled"
	StatusDisabled     = "disabled"
	SessionActive      = "active"
	SessionRevoked     = "revoked"
)

type Service struct {
	db *gorm.DB
}

// NewService 创建设备服务，注入数据库访问能力。
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// RegisterInput 登记设备的输入参数。
type RegisterInput struct {
	SerialNo   string `json:"serial_no"`
	Name       string `json:"name"`
	OrgCode    string `json:"org_code"`
	Status     string `json:"status"`
	BoundUserID *uint64 `json:"bound_user_id"`
}

// Register 登记一台待绑定或待启用的智能眼镜设备。
func (s *Service) Register(input RegisterInput) (database.Device, error) {
	input.Name = trim(input.Name)
	input.OrgCode = trim(input.OrgCode)
	status := normalizeStatus(input.Status)
	device := database.Device{
		SerialNo:    input.SerialNo,
		Name:        input.Name,
		OrgCode:     input.OrgCode,
		Status:      status,
		BoundUserID: input.BoundUserID,
	}
	if err := s.db.Create(&device).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return database.Device{}, httperr.New(httperr.ValidationFailed, "序列号已存在")
		}
		return database.Device{}, err
	}
	return device, nil
}

// UpdateInput 更新设备的输入参数。
type UpdateInput struct {
	Name       string `json:"name"`
	OrgCode    string `json:"org_code"`
	Status     string `json:"status"`
	BoundUserID *uint64 `json:"bound_user_id"`
}

// Update 更新设备信息（名称、组织机构、状态、绑定用户）。
func (s *Service) Update(id uint64, input UpdateInput) (database.Device, error) {
	var device database.Device
	if err := s.db.First(&device, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return database.Device{}, httperr.New(httperr.ResourceNotFound, "设备不存在")
		}
		return database.Device{}, err
	}
	input.Name = trim(input.Name)
	input.OrgCode = trim(input.OrgCode)
	device.Name = input.Name
	device.OrgCode = input.OrgCode
	if input.Status != "" {
		device.Status = normalizeStatus(input.Status)
	}
	if input.BoundUserID != nil {
		device.BoundUserID = input.BoundUserID
	}
	if err := s.db.Save(&device).Error; err != nil {
		return database.Device{}, err
	}
	return device, nil
}

// Delete 删除设备。
func (s *Service) Delete(id uint64) error {
	var device database.Device
	if err := s.db.First(&device, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return httperr.New(httperr.ResourceNotFound, "设备不存在")
		}
		return err
	}
	// 如果设备已绑定用户，先解绑
	if device.BoundUserID != nil {
		if err := s.db.Model(&database.Device{}).Where("id = ?", id).Update("bound_user_id", nil).Error; err != nil {
			return err
		}
	}
	result := s.db.Delete(&database.Device{}, id)
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "设备不存在")
	}
	return nil
}

// Enable 启用设备。
func (s *Service) Enable(id uint64) error {
	result := s.db.Model(&database.Device{}).Where("id = ?", id).Update("status", StatusActive)
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "设备不存在")
	}
	return nil
}

// Disable 停用设备。
func (s *Service) Disable(id uint64) error {
	result := s.db.Model(&database.Device{}).Where("id = ?", id).Update("status", StatusDisabled)
	if result.RowsAffected == 0 {
		return httperr.New(httperr.ResourceNotFound, "设备不存在")
	}
	return nil
}

// Bind 将设备绑定到指定巡检用户并标记为 active。
func (s *Service) Bind(deviceID, userID uint64) error {
	now := time.Now().UTC()
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&database.Device{}).Where("id = ? AND status IN ?", deviceID, []string{StatusPending, StatusActive}).Updates(map[string]any{
			"status":        StatusActive,
			"bound_user_id": userID,
			"bound_at":      now,
		}).Error; err != nil {
			return err
		}
		return tx.Create(&database.DeviceAuditLog{DeviceID: deviceID, ActorID: userID, Action: "bind"}).Error
	})
}

// Revoke 撤销设备访问权限，并写入设备审计日志。
func (s *Service) Revoke(actorID, deviceID uint64, reason string) error {
	return s.transition(actorID, deviceID, StatusRevoked, "revoke", reason)
}

// DisableLost 将设备标记为丢失禁用，并写入设备审计日志。
func (s *Service) DisableLost(actorID, deviceID uint64, reason string) error {
	return s.transition(actorID, deviceID, StatusLostDisabled, "disable_lost", reason)
}

// transition 统一处理设备状态迁移和审计日志写入。
func (s *Service) transition(actorID, deviceID uint64, status, action, reason string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&database.Device{}).Where("id = ?", deviceID).Update("status", status).Error; err != nil {
			return err
		}
		if err := tx.Model(&database.DeviceSession{}).Where("device_id = ? AND status = ?", deviceID, SessionActive).Update("status", SessionRevoked).Error; err != nil {
			return err
		}
		return tx.Create(&database.DeviceAuditLog{DeviceID: deviceID, ActorID: actorID, Action: action, Reason: reason}).Error
	})
}

// DeviceListResponse 设备列表响应，包含绑定用户姓名。
type DeviceListResponse struct {
	ID            uint64  `json:"id"`
	SerialNo      string  `json:"serial_no"`
	Name          string  `json:"name"`
	OrgCode       string  `json:"org_code"`
	Status        string  `json:"status"`
	BoundUserID   *uint64 `json:"bound_user_id"`
	BoundUserName string  `json:"bound_user_name"`
	CreatedAt     time.Time `json:"created_at"`
}

// List 查询全部设备，供后台设备管理页面展示，包含绑定用户姓名。
func (s *Service) List() ([]DeviceListResponse, error) {
	var devices []DeviceListResponse
	// 左连接用户表，获取绑定用户的姓名
	err := s.db.Model(&database.Device{}).
		Select("devices.*, users.name as bound_user_name").
		Joins("LEFT JOIN users ON devices.bound_user_id = users.id").
		Order("devices.id desc").
		Scan(&devices).Error
	return devices, err
}

// Get 查询单个设备详情。
func (s *Service) Get(id uint64) (database.Device, error) {
	var device database.Device
	if err := s.db.First(&device, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return database.Device{}, httperr.New(httperr.ResourceNotFound, "设备不存在")
		}
		return database.Device{}, err
	}
	return device, nil
}

func trim(s string) string {
	return s
}

func normalizeStatus(status string) string {
	if status == "" {
		return StatusPending
	}
	return status
}
