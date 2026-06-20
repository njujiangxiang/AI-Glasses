// 检查并修复菜单数据
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

	fmt.Println("=== 当前数据库中的菜单数据 ===")
	var menus []database.Permission
	db.Order("id").Find(&menus)

	for _, m := range menus {
		fmt.Printf("ID: %d, PID: %d, 类型: %s, 名称: %s, 编码: %s, 路径: %s\n",
			m.ID, m.Pid, m.Type, m.Name, m.Code, m.Path)
	}

	fmt.Printf("\n总计: %d 条菜单记录\n", len(menus))

	// 检查缺失的一级菜单
	expected := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	found := make(map[uint64]bool)
	for _, m := range menus {
		found[m.ID] = true
	}

	fmt.Println("\n=== 检查缺失的一级菜单 ===")
	for _, id := range expected {
		if !found[id] {
			fmt.Printf("❌ 缺失 ID: %d\n", id)
		}
	}
}
