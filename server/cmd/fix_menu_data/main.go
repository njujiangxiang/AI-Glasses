// 修复菜单数据
package main

import (
	"fmt"
	"log"

	"aiglasses/server/internal/config"
	"aiglasses/server/internal/platform/database"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	fmt.Println("正在修复菜单数据...")

	// 定义完整的菜单数据
	menus := []database.Permission{
		// 一级菜单
		{ID: 1, Pid: 0, Type: "M", Name: "工作台", Code: "workbench", Icon: "DataAnalysis", Path: "/workbench", Sort: 1, Visible: true, Status: "active"},
		{ID: 2, Pid: 0, Type: "M", Name: "巡检模板", Code: "templates", Icon: "Document", Path: "/templates", Sort: 2, Visible: true, Status: "active"},
		{ID: 3, Pid: 0, Type: "M", Name: "工作流管理", Code: "workflows", Icon: "Operation", Path: "/workflows", Sort: 3, Visible: true, Status: "active"},
		{ID: 4, Pid: 0, Type: "M", Name: "任务计划", Code: "plans", Icon: "Calendar", Path: "/plans", Sort: 4, Visible: true, Status: "active"},
		{ID: 5, Pid: 0, Type: "M", Name: "任务管理", Code: "tasks", Icon: "Tickets", Path: "/tasks", Sort: 5, Visible: true, Status: "active"},
		{ID: 6, Pid: 0, Type: "M", Name: "作业任务单", Code: "tasksheets", Icon: "Document", Path: "/tasksheets", Sort: 6, Visible: true, Status: "active"},
		{ID: 7, Pid: 0, Type: "M", Name: "缺陷管理", Code: "defects", Icon: "Bell", Path: "/defects", Sort: 7, Visible: true, Status: "active"},
		{ID: 8, Pid: 0, Type: "M", Name: "台账和主数据管理", Code: "master_data", Icon: "Collection", Path: "", Sort: 8, Visible: true, Status: "active"},
		{ID: 9, Pid: 0, Type: "M", Name: "系统管理", Code: "system", Icon: "Setting", Path: "", Sort: 99, Visible: true, Status: "active"},

		// 二级菜单
		{ID: 81, Pid: 8, Type: "C", Name: "设备管理", Code: "devices", Icon: "Monitor", Path: "/devices", Sort: 1, Visible: true, Status: "active"},

		{ID: 91, Pid: 9, Type: "C", Name: "组织管理", Code: "organizations", Icon: "OfficeBuilding", Path: "/organizations", Sort: 1, Visible: true, Status: "active"},
		{ID: 92, Pid: 9, Type: "C", Name: "用户管理", Code: "users", Icon: "User", Path: "/users", Sort: 2, Visible: true, Status: "active"},
		{ID: 93, Pid: 9, Type: "C", Name: "角色管理", Code: "roles", Icon: "Lock", Path: "/roles", Sort: 3, Visible: true, Status: "active"},
		{ID: 94, Pid: 9, Type: "C", Name: "菜单权限", Code: "menus", Icon: "Setting", Path: "/menus", Sort: 4, Visible: true, Status: "active"},
		{ID: 95, Pid: 9, Type: "C", Name: "业务编码配置", Code: "business_codes", Icon: "Key", Path: "/business-codes", Sort: 5, Visible: true, Status: "active"},

		// 按钮级权限
		{ID: 911, Pid: 91, Type: "A", Name: "组织查询", Code: "org:list", Perms: "system:org:list", Sort: 1, Visible: true, Status: "active"},
		{ID: 912, Pid: 91, Type: "A", Name: "组织新增", Code: "org:add", Perms: "system:org:add", Sort: 2, Visible: true, Status: "active"},
		{ID: 913, Pid: 91, Type: "A", Name: "组织编辑", Code: "org:edit", Perms: "system:org:edit", Sort: 3, Visible: true, Status: "active"},
		{ID: 914, Pid: 91, Type: "A", Name: "组织删除", Code: "org:delete", Perms: "system:org:delete", Sort: 4, Visible: true, Status: "active"},

		{ID: 921, Pid: 92, Type: "A", Name: "用户查询", Code: "user:list", Perms: "system:user:list", Sort: 1, Visible: true, Status: "active"},
		{ID: 922, Pid: 92, Type: "A", Name: "用户新增", Code: "user:add", Perms: "system:user:add", Sort: 2, Visible: true, Status: "active"},
		{ID: 923, Pid: 92, Type: "A", Name: "用户编辑", Code: "user:edit", Perms: "system:user:edit", Sort: 3, Visible: true, Status: "active"},
		{ID: 924, Pid: 92, Type: "A", Name: "用户删除", Code: "user:delete", Perms: "system:user:delete", Sort: 4, Visible: true, Status: "active"},
		{ID: 925, Pid: 92, Type: "A", Name: "用户启用", Code: "user:enable", Perms: "system:user:enable", Sort: 5, Visible: true, Status: "active"},
		{ID: 926, Pid: 92, Type: "A", Name: "用户停用", Code: "user:disable", Perms: "system:user:disable", Sort: 6, Visible: true, Status: "active"},

		{ID: 931, Pid: 93, Type: "A", Name: "角色查询", Code: "role:list", Perms: "system:role:list", Sort: 1, Visible: true, Status: "active"},
		{ID: 932, Pid: 93, Type: "A", Name: "角色新增", Code: "role:add", Perms: "system:role:add", Sort: 2, Visible: true, Status: "active"},
		{ID: 933, Pid: 93, Type: "A", Name: "角色编辑", Code: "role:edit", Perms: "system:role:edit", Sort: 3, Visible: true, Status: "active"},
		{ID: 934, Pid: 93, Type: "A", Name: "角色删除", Code: "role:delete", Perms: "system:role:delete", Sort: 4, Visible: true, Status: "active"},
		{ID: 935, Pid: 93, Type: "A", Name: "分配权限", Code: "role:assign", Perms: "system:role:assign", Sort: 5, Visible: true, Status: "active"},

		{ID: 941, Pid: 94, Type: "A", Name: "菜单查询", Code: "menu:list", Perms: "system:menu:list", Sort: 1, Visible: true, Status: "active"},
		{ID: 942, Pid: 94, Type: "A", Name: "菜单新增", Code: "menu:add", Perms: "system:menu:add", Sort: 2, Visible: true, Status: "active"},
		{ID: 943, Pid: 94, Type: "A", Name: "菜单编辑", Code: "menu:edit", Perms: "system:menu:edit", Sort: 3, Visible: true, Status: "active"},
		{ID: 944, Pid: 94, Type: "A", Name: "菜单删除", Code: "menu:delete", Perms: "system:menu:delete", Sort: 4, Visible: true, Status: "active"},
	}

	for _, menu := range menus {
		result := db.Save(&menu)
		if result.Error != nil {
			log.Printf("更新菜单失败 %d-%s: %v", menu.ID, menu.Name, result.Error)
		} else {
			fmt.Printf("✅ 更新菜单: %s\n", menu.Name)
		}
	}

	fmt.Println("\n正在为超级管理员重新分配所有菜单权限...")

	// 先删除旧的关联
	db.Where("role_id = ?", 1).Delete(&database.RolePermission{})

	// 重新分配所有目录和菜单权限
	for _, menu := range menus {
		if menu.Type == "M" || menu.Type == "C" {
			rp := database.RolePermission{RoleID: 1, PermissionID: menu.ID}
			db.Create(&rp)
		}
	}

	fmt.Printf("✅ 已为超级管理员分配 %d 个菜单权限\n", len(menus))
	fmt.Println("\n✅ 菜单数据修复完成！")
}
