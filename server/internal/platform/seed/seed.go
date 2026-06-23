// Package seed 写入可重复执行的本地开发初始化数据。它创建默认角色、用户、演示班组和
// 已激活的演示眼镜设备，使数据库初始化后可立即验证后台 UI 与眼镜端流程。
package seed

import (
	"time"

	"aiglasses/server/internal/platform/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Run 在事务中写入默认角色、用户、班组和演示设备，重复执行不会产生重复数据。
func Run(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		roles := []database.Role{{Name: "系统管理员"}, {Name: "任务管理员"}, {Name: "班组长"}, {Name: "巡检员"}}
		for _, role := range roles {
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).Create(&role).Error; err != nil {
				return err
			}
		}

		permissions := []database.Permission{
			{Code: "admin:*"}, {Code: "admin:templates"}, {Code: "admin:plans"}, {Code: "admin:tasks"},
			{Code: "admin:defects"}, {Code: "admin:devices"}, {Code: "glasses:tasks"},
		}
		for _, permission := range permissions {
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).Create(&permission).Error; err != nil {
				return err
			}
		}

		rootOrg := database.Organization{Code: "ROOT", Name: "默认单位", ParentCode: "", Status: "active"}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "status"})}).Create(&rootOrg).Error; err != nil {
			return err
		}

		defaultHash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		hashStr := string(defaultHash)
		users := []database.User{
			{Username: "admin", PasswordHash: hashStr, DisplayName: "系统管理员", Name: "系统管理员", Gender: "unknown", OrgCode: "ROOT", Status: "active"},
			{Username: "manager", PasswordHash: hashStr, DisplayName: "任务管理员", Name: "任务管理员", Gender: "unknown", OrgCode: "ROOT", Status: "active"},
			{Username: "leader", PasswordHash: hashStr, DisplayName: "巡检班组长", Name: "巡检班组长", Gender: "unknown", OrgCode: "ROOT", Status: "active"},
			{Username: "inspector", PasswordHash: hashStr, DisplayName: "巡检员", Name: "巡检员", Gender: "unknown", OrgCode: "ROOT", Status: "active"},
		}
		for _, user := range users {
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "username"}}, DoUpdates: clause.AssignmentColumns([]string{"password_hash", "display_name", "name", "gender", "org_code", "status"})}).Create(&user).Error; err != nil {
				return err
			}
		}

		var adminRole, managerRole, leaderRole, inspectorRole database.Role
		if err := tx.Where("name = ?", "系统管理员").First(&adminRole).Error; err != nil {
			return err
		}
		if err := tx.Where("name = ?", "任务管理员").First(&managerRole).Error; err != nil {
			return err
		}
		if err := tx.Where("name = ?", "班组长").First(&leaderRole).Error; err != nil {
			return err
		}
		if err := tx.Where("name = ?", "巡检员").First(&inspectorRole).Error; err != nil {
			return err
		}

		roleByUser := map[string]uint64{"admin": adminRole.ID, "manager": managerRole.ID, "leader": leaderRole.ID, "inspector": inspectorRole.ID}
		for username, roleID := range roleByUser {
			var user database.User
			if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
				return err
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&database.UserRole{UserID: user.ID, RoleID: roleID}).Error; err != nil {
				return err
			}
		}

		var team database.Team
		if err := tx.Where(database.Team{Name: "A 区巡检班组"}).FirstOrCreate(&team).Error; err != nil {
			return err
		}
		for _, username := range []string{"leader", "inspector"} {
			var user database.User
			if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
				return err
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&database.TeamMember{TeamID: team.ID, UserID: user.ID}).Error; err != nil {
				return err
			}
		}

		var inspector database.User
		if err := tx.Where("username = ?", "inspector").First(&inspector).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		device := database.Device{SerialNo: "GLASS-DEMO-001", Name: "演示智能眼镜", Status: "active", BoundUserID: &inspector.ID, BoundAt: &now}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "serial_no"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "status", "bound_user_id", "bound_at"})}).Create(&device).Error; err != nil {
			return err
		}

		// 初始化任务类型字典
		taskTypes := []database.TaskTypeDict{
			{TypeCode: "text", TypeName: "文本项", TypeDesc: "纯文本输入", SupportAlgorithm: "0", SupportQuery: "0", SupportMandatory: "1"},
			{TypeCode: "read", TypeName: "读数项", TypeDesc: "读取设备表计数据", SupportAlgorithm: "0", SupportQuery: "1", SupportMandatory: "1"},
			{TypeCode: "check", TypeName: "确认项", TypeDesc: "多选确认类节点", SupportAlgorithm: "0", SupportQuery: "0", SupportMandatory: "1"},
			{TypeCode: "photo", TypeName: "拍照项", TypeDesc: "照片采集", SupportAlgorithm: "1", SupportQuery: "0", SupportMandatory: "1"},
			{TypeCode: "video", TypeName: "录像项", TypeDesc: "视频采集", SupportAlgorithm: "1", SupportQuery: "0", SupportMandatory: "1"},
			{TypeCode: "audio", TypeName: "语音项", TypeDesc: "音频采集", SupportAlgorithm: "1", SupportQuery: "0", SupportMandatory: "1"},
		}
		for _, tt := range taskTypes {
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "type_code"}}, DoUpdates: clause.AssignmentColumns([]string{"type_name", "type_desc", "support_algorithm", "support_query", "support_mandatory"})}).Create(&tt).Error; err != nil {
				return err
			}
		}

		// 初始化业务编码配置
		businessCodes := []database.BusinessCode{
			{Name: "巡检任务", Code: "TK", DateFormat: "yyyyMMdd", SeqPadding: 4, Separator: "-", UseSeparator: true, Status: "active"},
			{Name: "缺陷记录", Code: "DEF", DateFormat: "yyyyMMdd", SeqPadding: 4, Separator: "-", UseSeparator: true, Status: "active"},
		}
		for _, bc := range businessCodes {
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "date_format", "seq_padding", "separator", "use_separator", "status"})}).Create(&bc).Error; err != nil {
				return err
			}
		}

		// 初始化巡检点位
		inspectionPoints := []database.InspectionPoint{
			{Name: "A区一号柜", EquipmentName: "1号变压器", Area: "A区", Substation: "城东变电站", Location: "A区高压室一层", Description: "A区主变压器巡检点"},
			{Name: "A区二号柜", EquipmentName: "2号变压器", Area: "A区", Substation: "城东变电站", Location: "A区高压室二层", Description: "A区副变压器巡检点"},
			{Name: "B区一层", EquipmentName: "B区配电柜", Area: "B区", Substation: "城东变电站", Location: "B区配电室一层", Description: "B区配电设备巡检点"},
			{Name: "B区二层", EquipmentName: "B区开关柜", Area: "B区", Substation: "城东变电站", Location: "B区配电室二层", Description: "B区开关设备巡检点"},
			{Name: "C区线路", EquipmentName: "C区电缆", Area: "C区", Substation: "城西变电站", Location: "C区电缆廊道", Description: "C区电缆巡检点"},
		}
		for _, ip := range inspectionPoints {
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoUpdates: clause.AssignmentColumns([]string{"equipment_name", "location", "area", "substation", "description"})}).Create(&ip).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
