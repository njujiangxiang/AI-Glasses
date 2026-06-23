
package database

import "gorm.io/gorm"

// AutoMigrate 执行 GORM 自动迁移，确保核心业务表结构存在。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{}, &Organization{}, &BusinessCode{}, &Role{}, &Permission{}, &UserRole{}, &RolePermission{},
		&Team{}, &TeamMember{},
		&Device{}, &DeviceSession{}, &DeviceAuditLog{},
		&TaskTypeDict{}, &AlgorithmConfig{}, &RealtimeQueryConfig{},
		&InspectionTemplate{}, &InspectionTemplateNode{}, &TemplateNodeConfig{},
		&Workflow{}, &WorkflowStep{},
		&InspectionPoint{},
		&TaskPlan{}, &InspectionTask{}, &InspectionTaskNode{}, &TaskNodeResult{},
		&Attachment{}, &Defect{}, &AuditLog{}, &OutboxEvent{},
	)
}
