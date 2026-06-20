// Package main 数据库迁移脚本，用于添加 data_scope 字段到 roles 表
package main

import (
	"log"

	"aiglasses/server/internal/config"
	"aiglasses/server/internal/platform/database"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 1. 添加 data_scope 列到 roles 表
	log.Println("正在添加 data_scope 列到 roles 表...")
	err = db.Exec("ALTER TABLE roles ADD COLUMN data_scope VARCHAR(32) NOT NULL DEFAULT 'org_only'").Error
	if err != nil {
		log.Println("添加列失败 (可能已存在):", err)
	} else {
		log.Println("✅ data_scope 列添加成功")
	}

	// 2. 创建索引
	log.Println("正在创建 data_scope 索引...")
	err = db.Exec("CREATE INDEX idx_roles_data_scope ON roles(data_scope)").Error
	if err != nil {
		log.Println("创建索引失败 (可能已存在):", err)
	} else {
		log.Println("✅ idx_roles_data_scope 索引创建成功")
	}

	// 3. 创建 users.org_code 索引（如果不存在）
	log.Println("正在创建 users.org_code 索引...")
	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_users_org_code ON users(org_code)").Error
	if err != nil {
		log.Println("创建索引失败:", err)
	} else {
		log.Println("✅ idx_users_org_code 索引创建成功")
	}

	// 4. 创建 organizations.parent_code 索引（如果不存在）
	log.Println("正在创建 organizations.parent_code 索引...")
	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_organizations_parent_code ON organizations(parent_code)").Error
	if err != nil {
		log.Println("创建索引失败:", err)
	} else {
		log.Println("✅ idx_organizations_parent_code 索引创建成功")
	}

	// 5. 更新现有角色的数据范围
	log.Println("正在更新现有角色数据...")
	result := db.Model(&database.Role{}).Where("data_scope = '' OR data_scope IS NULL").Update("data_scope", "org_only")
	if result.Error != nil {
		log.Println("更新数据失败:", result.Error)
	} else {
		log.Printf("✅ 已更新 %d 个角色的默认数据范围", result.RowsAffected)
	}

	// 6. 查询并显示当前角色的数据范围
	var roles []struct {
		ID        uint64
		Name      string
		DataScope string
	}
	db.Model(&database.Role{}).Select("id, name, data_scope").Find(&roles)
	log.Println("\n当前角色数据范围配置:")
	for _, r := range roles {
		log.Printf("  - %s (ID: %d): %s", r.Name, r.ID, r.DataScope)
	}

	log.Println("\n✅ 数据迁移完成!")
}
