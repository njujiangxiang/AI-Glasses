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
		// 初始化任务类型字典、算法配置、实时查询配置
		if err := seedARSupportData(tx); err != nil {
			return err
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

func seedARSupportData(tx *gorm.DB) error {
	taskTypes := []database.TaskTypeDict{
		{TypeCode: "check", TypeName: "检查确认", TypeDesc: "确认设备状态或选项", SupportMandatory: true},
		{TypeCode: "read", TypeName: "读数记录", TypeDesc: "读取仪表或设备数值", SupportAlgorithm: true, SupportQuery: true, SupportMandatory: true},
		{TypeCode: "photo", TypeName: "拍照留证", TypeDesc: "拍摄现场照片", SupportAlgorithm: true, SupportMandatory: true},
		{TypeCode: "text", TypeName: "文本记录", TypeDesc: "填写文本备注"},
		{TypeCode: "defect_report", TypeName: "缺陷上报", TypeDesc: "眼镜端主动上报缺陷"},
		{TypeCode: "realtime_query", TypeName: "实时查询", TypeDesc: "调用实时数据查询配置", SupportQuery: true},
	}
	for _, item := range taskTypes {
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "type_code"}}, DoUpdates: clause.AssignmentColumns([]string{"type_name", "type_desc", "support_algorithm", "support_query", "support_mandatory"})}).Create(&item).Error; err != nil {
			return err
		}
	}
	algorithms := []database.AlgorithmConfig{
		{Name: "表计读数识别Mock", ServiceURL: "mock://meter-reading", InputParams: `{"attachment_ids":"array"}`, OutputParams: `{"value":"12.34","confidence":0.98}`, IsEnable: true},
		{Name: "图像异常检测Mock", ServiceURL: "mock://image-anomaly", InputParams: `{"attachment_ids":"array"}`, OutputParams: `{"is_abnormal":"0","confidence":0.98}`, IsEnable: true},
	}
	for _, item := range algorithms {
		if err := tx.Where("name = ?", item.Name).Assign(item).FirstOrCreate(&item).Error; err != nil {
			return err
		}
	}
	queries := []database.RealtimeQueryConfig{
		{Name: "设备状态查询Mock", APIURL: "mock://equipment-status", RequestParams: `{"task_id":"string","node_id":"string"}`, ResponseParams: `{"status":"normal","temperature":"36.5℃"}`, IsEnable: true},
		{Name: "环境状态查询Mock", APIURL: "mock://environment", RequestParams: `{"area":"string"}`, ResponseParams: `{"humidity":"45%","wind":"normal"}`, IsEnable: true},
	}
	for _, item := range queries {
		if err := tx.Where("name = ?", item.Name).Assign(item).FirstOrCreate(&item).Error; err != nil {
			return err
		}
	}
	return nil
}
