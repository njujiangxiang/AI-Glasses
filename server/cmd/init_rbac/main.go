// RBAC 权限数据初始化工具
// 使用方法: go run cmd/init_rbac/main.go
package main

import (
	"fmt"
	"log"

	"aiglasses/server/internal/config"
	"aiglasses/server/internal/platform/database"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	fmt.Println("开始初始化 RBAC 权限数据...")

	// 1. 初始化菜单数据
	if err := initMenus(db); err != nil {
		log.Fatalf("初始化菜单失败: %v", err)
	}

	// 2. 初始化角色数据
	if err := initRoles(db); err != nil {
		log.Fatalf("初始化角色失败: %v", err)
	}

	// 3. 为超级管理员分配权限
	if err := assignSuperAdminPermissions(db); err != nil {
		log.Fatalf("分配超级管理员权限失败: %v", err)
	}

	// 4. 更新 admin 用户角色
	if err := updateAdminUserRole(db); err != nil {
		log.Fatalf("更新 admin 用户角色失败: %v", err)
	}

	fmt.Println("✅ RBAC 权限数据初始化完成！")
	fmt.Println("")
	fmt.Println("说明:")
	fmt.Println("1. 已创建 11 个一级菜单 + 子菜单")
	fmt.Println("2. 已创建 3 个默认角色: 超级管理员、普通用户、巡检人员")
	fmt.Println("3. 超级管理员已获得所有菜单权限")
	fmt.Println("4. admin 用户已设置为超级管理员")
	fmt.Println("")
	fmt.Println("下一步:")
	fmt.Println("- 启动服务后使用 admin 账号登录")
	fmt.Println("- 进入 系统管理 → 角色管理 配置其他角色权限")
	fmt.Println("- 进入 系统管理 → 用户管理 为用户分配角色")
}

func initMenus(db *gorm.DB) error {
	fmt.Println("正在初始化菜单数据...")

	menus := []database.Permission{
		// 一级菜单
		{ID: 1, Pid: 0, Type: "M", Name: "工作台", Code: "workbench", Icon: "DataAnalysis", Path: "/workbench", Sort: 1, Visible: true, Status: "active"},
		{ID: 2, Pid: 0, Type: "M", Name: "任务模板", Code: "templates", Icon: "Document", Path: "/templates", Sort: 2, Visible: true, Status: "active"},
		{ID: 3, Pid: 0, Type: "M", Name: "工作流管理", Code: "workflows", Icon: "Operation", Path: "/workflows", Sort: 3, Visible: true, Status: "active"},
		{ID: 4, Pid: 0, Type: "M", Name: "任务计划", Code: "plans", Icon: "Calendar", Path: "/plans", Sort: 4, Visible: true, Status: "active"},
		{ID: 5, Pid: 0, Type: "M", Name: "任务管理", Code: "tasks", Icon: "Tickets", Path: "/tasks", Sort: 5, Visible: true, Status: "active"},
		{ID: 6, Pid: 0, Type: "M", Name: "作业任务单", Code: "tasksheets", Icon: "Document", Path: "/tasksheets", Sort: 6, Visible: true, Status: "active"},
		{ID: 7, Pid: 0, Type: "M", Name: "缺陷管理", Code: "defects", Icon: "Bell", Path: "/defects", Sort: 7, Visible: true, Status: "active"},
		{ID: 8, Pid: 0, Type: "M", Name: "点位管理", Code: "inspection_points", Icon: "MapLocation", Path: "/inspection-points", Sort: 8, Visible: true, Status: "active"},
		{ID: 9, Pid: 0, Type: "M", Name: "台账和主数据管理", Code: "master_data", Icon: "Collection", Sort: 9, Visible: true, Status: "active"},
		{ID: 10, Pid: 0, Type: "M", Name: "巡检报告", Code: "reports", Icon: "DataBoard", Path: "/reports", Sort: 10, Visible: true, Status: "active"},
		{ID: 99, Pid: 0, Type: "M", Name: "系统管理", Code: "system", Icon: "Setting", Sort: 99, Visible: true, Status: "active"},

		// 二级菜单 - 台账和主数据管理
		{ID: 81, Pid: 9, Type: "C", Name: "设备管理", Code: "devices", Icon: "Monitor", Path: "/devices", Sort: 1, Visible: true, Status: "active"},

		// 二级菜单 - 系统管理
		{ID: 91, Pid: 99, Type: "C", Name: "组织管理", Code: "organizations", Icon: "OfficeBuilding", Path: "/organizations", Sort: 1, Visible: true, Status: "active"},
		{ID: 92, Pid: 99, Type: "C", Name: "用户管理", Code: "users", Icon: "User", Path: "/users", Sort: 2, Visible: true, Status: "active"},
		{ID: 93, Pid: 99, Type: "C", Name: "角色管理", Code: "roles", Icon: "Lock", Path: "/roles", Sort: 3, Visible: true, Status: "active"},
		{ID: 94, Pid: 99, Type: "C", Name: "菜单权限", Code: "menus", Icon: "Setting", Path: "/menus", Sort: 4, Visible: true, Status: "active"},
		{ID: 95, Pid: 99, Type: "C", Name: "业务编码配置", Code: "business_codes", Icon: "Key", Path: "/business-codes", Sort: 5, Visible: true, Status: "active"},
		{ID: 96, Pid: 99, Type: "C", Name: "实时监控", Code: "monitoring_logs", Icon: "Monitor", Path: "/monitoring/logs", Sort: 6, Visible: true, Status: "active"},

		// 按钮级权限 - 实时监控
		{ID: 961, Pid: 96, Type: "A", Name: "实时监控查看", Code: "monitor:view", Perms: "system:monitor:view", Sort: 1, Visible: true, Status: "active"},

		// 按钮级权限 - 用户管理
		{ID: 921, Pid: 92, Type: "A", Name: "用户查询", Code: "user:list", Perms: "system:user:list", Sort: 1, Visible: true, Status: "active"},
		{ID: 922, Pid: 92, Type: "A", Name: "用户新增", Code: "user:add", Perms: "system:user:add", Sort: 2, Visible: true, Status: "active"},
		{ID: 923, Pid: 92, Type: "A", Name: "用户编辑", Code: "user:edit", Perms: "system:user:edit", Sort: 3, Visible: true, Status: "active"},
		{ID: 924, Pid: 92, Type: "A", Name: "用户删除", Code: "user:delete", Perms: "system:user:delete", Sort: 4, Visible: true, Status: "active"},
		{ID: 925, Pid: 92, Type: "A", Name: "用户启用", Code: "user:enable", Perms: "system:user:enable", Sort: 5, Visible: true, Status: "active"},
		{ID: 926, Pid: 92, Type: "A", Name: "用户停用", Code: "user:disable", Perms: "system:user:disable", Sort: 6, Visible: true, Status: "active"},

		// 按钮级权限 - 角色管理
		{ID: 931, Pid: 93, Type: "A", Name: "角色查询", Code: "role:list", Perms: "system:role:list", Sort: 1, Visible: true, Status: "active"},
		{ID: 932, Pid: 93, Type: "A", Name: "角色新增", Code: "role:add", Perms: "system:role:add", Sort: 2, Visible: true, Status: "active"},
		{ID: 933, Pid: 93, Type: "A", Name: "角色编辑", Code: "role:edit", Perms: "system:role:edit", Sort: 3, Visible: true, Status: "active"},
		{ID: 934, Pid: 93, Type: "A", Name: "角色删除", Code: "role:delete", Perms: "system:role:delete", Sort: 4, Visible: true, Status: "active"},
		{ID: 935, Pid: 93, Type: "A", Name: "分配权限", Code: "role:assign", Perms: "system:role:assign", Sort: 5, Visible: true, Status: "active"},

		// 按钮级权限 - 菜单管理
		{ID: 941, Pid: 94, Type: "A", Name: "菜单查询", Code: "menu:list", Perms: "system:menu:list", Sort: 1, Visible: true, Status: "active"},
		{ID: 942, Pid: 94, Type: "A", Name: "菜单新增", Code: "menu:add", Perms: "system:menu:add", Sort: 2, Visible: true, Status: "active"},
		{ID: 943, Pid: 94, Type: "A", Name: "菜单编辑", Code: "menu:edit", Perms: "system:menu:edit", Sort: 3, Visible: true, Status: "active"},
		{ID: 944, Pid: 94, Type: "A", Name: "菜单删除", Code: "menu:delete", Perms: "system:menu:delete", Sort: 4, Visible: true, Status: "active"},

		// 按钮级权限 - 组织管理
		{ID: 911, Pid: 91, Type: "A", Name: "组织查询", Code: "org:list", Perms: "system:org:list", Sort: 1, Visible: true, Status: "active"},
		{ID: 912, Pid: 91, Type: "A", Name: "组织新增", Code: "org:add", Perms: "system:org:add", Sort: 2, Visible: true, Status: "active"},
		{ID: 913, Pid: 91, Type: "A", Name: "组织编辑", Code: "org:edit", Perms: "system:org:edit", Sort: 3, Visible: true, Status: "active"},
		{ID: 914, Pid: 91, Type: "A", Name: "组织删除", Code: "org:delete", Perms: "system:org:delete", Sort: 4, Visible: true, Status: "active"},
	}

	for _, menu := range menus {
		result := db.FirstOrCreate(&database.Permission{}, database.Permission{ID: menu.ID})
		if result.Error != nil {
			return fmt.Errorf("创建菜单 %d-%s 失败: %w", menu.ID, menu.Name, result.Error)
		}
		if result.RowsAffected > 0 {
			// 新创建的，更新完整信息
			db.Model(&menu).Where("id = ?", menu.ID).Updates(menu)
			fmt.Printf("  [创建] 菜单: %s\n", menu.Name)
		} else {
			fmt.Printf("  [跳过] 菜单已存在: %s\n", menu.Name)
		}
	}

	return nil
}

func initRoles(db *gorm.DB) error {
	fmt.Println("正在初始化角色数据...")

	roles := []database.Role{
		{ID: 1, Name: "超级管理员", Code: "super_admin", Description: "系统超级管理员，拥有所有权限", Sort: 1, Status: "active"},
		{ID: 2, Name: "普通用户", Code: "user", Description: "普通业务用户", Sort: 2, Status: "active"},
		{ID: 3, Name: "巡检人员", Code: "inspector", Description: "负责执行巡检任务", Sort: 3, Status: "active"},
	}

	for _, role := range roles {
		result := db.FirstOrCreate(&database.Role{}, database.Role{ID: role.ID})
		if result.Error != nil {
			return fmt.Errorf("创建角色 %d-%s 失败: %w", role.ID, role.Name, result.Error)
		}
		if result.RowsAffected > 0 {
			db.Model(&role).Where("id = ?", role.ID).Updates(role)
			fmt.Printf("  [创建] 角色: %s\n", role.Name)
		} else {
			fmt.Printf("  [跳过] 角色已存在: %s\n", role.Name)
		}
	}

	return nil
}

func assignSuperAdminPermissions(db *gorm.DB) error {
	fmt.Println("正在为超级管理员分配权限...")

	// 先删除旧的关联
	db.Where("role_id = ?", 1).Delete(&database.RolePermission{})

	// 获取所有菜单和目录权限，并额外授予实时监控接口所需的查看权限。
	var permissions []database.Permission
	if err := db.Where("type IN ? OR perms = ?", []string{"M", "C"}, "system:monitor:view").Find(&permissions).Error; err != nil {
		return err
	}

	// 批量分配权限
	for _, perm := range permissions {
		rp := database.RolePermission{RoleID: 1, PermissionID: perm.ID}
		db.FirstOrCreate(&rp, rp)
	}

	fmt.Printf("  已为超级管理员分配 %d 个菜单权限\n", len(permissions))
	return nil
}

func updateAdminUserRole(db *gorm.DB) error {
	fmt.Println("正在更新 admin 用户角色...")

	result := db.Model(&database.User{}).Where("username = ?", "admin").Update("role_id", 1)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		fmt.Println("  [更新] admin 用户已设置为超级管理员")
	} else {
		fmt.Println("  [跳过] 未找到 admin 用户或已设置角色")
	}

	return nil
}
