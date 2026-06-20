// 检查用户菜单权限
package main

import (
	"fmt"
	"log"

	"aiglasses/server/internal/config"
	"aiglasses/server/internal/menus"
	"aiglasses/server/internal/platform/database"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	fmt.Println("=== 检查 admin 用户信息 ===")
	var user database.User
	db.Where("username = ?", "admin").First(&user)
	fmt.Printf("用户ID: %d, 用户名: %s, 角色ID: %d\n", user.ID, user.Username, user.RoleID)

	if user.RoleID > 0 {
		fmt.Println("\n=== 检查角色关联的菜单权限 ===")
		var role database.Role
		db.First(&role, user.RoleID)
		fmt.Printf("角色ID: %d, 名称: %s, 编码: %s\n", role.ID, role.Name, role.Code)

		var rps []database.RolePermission
		db.Where("role_id = ?", user.RoleID).Find(&rps)
		fmt.Printf("\n角色关联的权限数量: %d\n", len(rps))

		menuSvc := menus.NewService(db)
		userMenus, err := menuSvc.GetUserMenus(user.ID)
		if err != nil {
			log.Fatalf("获取用户菜单失败: %v", err)
		}

		fmt.Printf("\n用户可访问的菜单树 (%d个顶级菜单):\n", len(userMenus))
		printMenuTree(userMenus, 0)

		fmt.Println("\n=== 完整的菜单列表 ===")
		allMenus, _ := menuSvc.ListAll()
		for _, m := range allMenus {
			fmt.Printf("ID: %d, PID: %d, 类型: %s, 名称: %s, 路径: %s\n",
				m.ID, m.Pid, m.Type, m.Name, m.Path)
		}
	} else {
		fmt.Println("⚠️  admin 用户没有关联角色！")
		fmt.Println("正在为 admin 用户分配超级管理员角色...")

		db.Model(&user).Update("role_id", 1)
		fmt.Println("✅ 已为 admin 用户分配角色 ID=1 (超级管理员)")

		// 重新检查
		db.First(&user, user.ID)
		fmt.Printf("更新后 - 用户ID: %d, 角色ID: %d\n", user.ID, user.RoleID)
	}
}

func printMenuTree(menus []menus.MenuDTO, level int) {
	for _, m := range menus {
		prefix := ""
		for i := 0; i < level; i++ {
			prefix += "  "
		}
		fmt.Printf("%s- %s (路径: %s)\n", prefix, m.Name, m.Path)
		if len(m.Children) > 0 {
			printMenuTree(m.Children, level+1)
		}
	}
}
