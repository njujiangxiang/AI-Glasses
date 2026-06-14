package devices

import (
	"time"

	"aiglasses/server/internal/platform/database"
	"gorm.io/gorm"
)

const (
	StatusPending      = "pending"
	StatusActive       = "active"
	StatusRevoked      = "revoked"
	StatusLostDisabled = "lost_disabled"
	SessionActive      = "active"
	SessionRevoked     = "revoked"
)

type Service struct {
	db *gorm.DB
}

// NewService 创建设备服务，注入数据库访问能力。
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// Register 登记一台待绑定或待启用的智能眼镜设备。
func (s *Service) Register(serialNo, name string) (database.Device, error) {
	device := database.Device{SerialNo: serialNo, Name: name, Status: StatusPending}
	return device, s.db.Create(&device).Error
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

// List 查询全部设备，供后台设备管理页面展示。
func (s *Service) List() ([]database.Device, error) {
	var devices []database.Device
	return devices, s.db.Order("id desc").Find(&devices).Error
}
