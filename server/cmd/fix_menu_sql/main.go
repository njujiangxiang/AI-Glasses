// 使用原始 SQL 修复菜单数据
package main

import (
	"fmt"
	"log"

	"aiglasses/server/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := gorm.Open(mysql.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	fmt.Println("正在使用原始 SQL 修复菜单数据...")

	menus := []struct {
		ID        uint64
		Pid       uint64
		Type      string
		Name      string
		Code      string
		Icon      string
		Path      string
		Sort      int
		Perms     string
		Visible   bool
		IsCache   bool
		Status    string
	}{
		{1, 0, "M", "工作台", "workbench", "DataAnalysis", "/workbench", 1, "", true, false, "active"},
		{2, 0, "M", "巡检模板", "templates", "Document", "/templates", 2, "", true, false, "active"},
		{3, 0, "M", "工作流管理", "workflows", "Operation", "/workflows", 3, "", true, false, "active"},
		{4, 0, "M", "任务计划", "plans", "Calendar", "/plans", 4, "", true, false, "active"},
		{5, 0, "M", "任务管理", "tasks", "Tickets", "/tasks", 5, "", true, false, "active"},
		{6, 0, "M", "作业任务单", "tasksheets", "Document", "/tasksheets", 6, "", true, false, "active"},
		{7, 0, "M", "缺陷管理", "defects", "Bell", "/defects", 7, "", true, false, "active"},
		{8, 0, "M", "台账和主数据管理", "master_data", "Collection", "", 8, "", true, false, "active"},
		{9, 0, "M", "系统管理", "system", "Setting", "", 99, "", true, false, "active"},

		{81, 8, "C", "设备管理", "devices", "Monitor", "/devices", 1, "", true, false, "active"},

		{91, 9, "C", "组织管理", "organizations", "OfficeBuilding", "/organizations", 1, "", true, false, "active"},
		{92, 9, "C", "用户管理", "users", "User", "/users", 2, "", true, false, "active"},
		{93, 9, "C", "角色管理", "roles", "Lock", "/roles", 3, "", true, false, "active"},
		{94, 9, "C", "菜单权限", "menus", "Setting", "/menus", 4, "", true, false, "active"},
		{95, 9, "C", "业务编码配置", "business_codes", "Key", "/business-codes", 5, "", true, false, "active"},

		{911, 91, "A", "组织查询", "org:list", "", "", 1, "system:org:list", true, false, "active"},
		{912, 91, "A", "组织新增", "org:add", "", "", 2, "system:org:add", true, false, "active"},
		{913, 91, "A", "组织编辑", "org:edit", "", "", 3, "system:org:edit", true, false, "active"},
		{914, 91, "A", "组织删除", "org:delete", "", "", 4, "system:org:delete", true, false, "active"},

		{921, 92, "A", "用户查询", "user:list", "", "", 1, "system:user:list", true, false, "active"},
		{922, 92, "A", "用户新增", "user:add", "", "", 2, "system:user:add", true, false, "active"},
		{923, 92, "A", "用户编辑", "user:edit", "", "", 3, "system:user:edit", true, false, "active"},
		{924, 92, "A", "用户删除", "user:delete", "", "", 4, "system:user:delete", true, false, "active"},
		{925, 92, "A", "用户启用", "user:enable", "", "", 5, "system:user:enable", true, false, "active"},
		{926, 92, "A", "用户停用", "user:disable", "", "", 6, "system:user:disable", true, false, "active"},

		{931, 93, "A", "角色查询", "role:list", "", "", 1, "system:role:list", true, false, "active"},
		{932, 93, "A", "角色新增", "role:add", "", "", 2, "system:role:add", true, false, "active"},
		{933, 93, "A", "角色编辑", "role:edit", "", "", 3, "system:role:edit", true, false, "active"},
		{934, 93, "A", "角色删除", "role:delete", "", "", 4, "system:role:delete", true, false, "active"},
		{935, 93, "A", "分配权限", "role:assign", "", "", 5, "system:role:assign", true, false, "active"},

		{941, 94, "A", "菜单查询", "menu:list", "", "", 1, "system:menu:list", true, false, "active"},
		{942, 94, "A", "菜单新增", "menu:add", "", "", 2, "system:menu:add", true, false, "active"},
		{943, 94, "A", "菜单编辑", "menu:edit", "", "", 3, "system:menu:edit", true, false, "active"},
		{944, 94, "A", "菜单删除", "menu:delete", "", "", 4, "system:menu:delete", true, false, "active"},
	}

	for _, menu := range menus {
		err := db.Exec(`
			UPDATE permissions
			SET pid = ?, type = ?, name = ?, code = ?, icon = ?, path = ?,
			    sort = ?, perms = ?, visible = ?, is_cache = ?, status = ?
			WHERE id = ?
		`, menu.Pid, menu.Type, menu.Name, menu.Code, menu.Icon, menu.Path,
			menu.Sort, menu.Perms, menu.Visible, menu.IsCache, menu.Status, menu.ID).Error

		if err != nil {
			log.Printf("更新菜单失败 %d-%s: %v", menu.ID, menu.Name, err)
		} else {
			fmt.Printf("✅ 更新菜单: %s\n", menu.Name)
		}
	}

	fmt.Println("\n正在为超级管理员重新分配所有菜单权限...")
	db.Exec("DELETE FROM role_permissions WHERE role_id = 1")
	for _, menu := range menus {
		if menu.Type == "M" || menu.Type == "C" {
			db.Exec("INSERT IGNORE INTO role_permissions (role_id, permission_id) VALUES (1, ?)", menu.ID)
		}
	}

	fmt.Printf("✅ 已为超级管理员分配 %d 个菜单权限\n", len(menus))
	fmt.Println("\n✅ 菜单数据修复完成！")
}
