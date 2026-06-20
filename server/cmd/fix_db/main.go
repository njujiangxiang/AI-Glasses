// 数据库修复工具 - 处理迁移前的数据问题
// 使用方法: go run cmd/fix_db/main.go
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

	fmt.Println("开始修复数据库...")

	// 1. 修复 roles 表 code 字段为空的问题
	fmt.Print("修复 roles 表 code 字段...")
	if err := fixRolesCode(db); err != nil {
		log.Fatalf("修复 roles 表失败: %v", err)
	}
	fmt.Println(" ✅")

	// 2. 检查并修复 users 表 role_id 字段
	fmt.Print("检查 users 表 role_id 字段...")
	if err := fixUsersRoleID(db); err != nil {
		log.Fatalf("修复 users 表失败: %v", err)
	}
	fmt.Println(" ✅")

	fmt.Println("\n✅ 数据库修复完成！")
	fmt.Println("\n现在可以执行:")
	fmt.Println("  go run cmd/migrate/main.go")
	fmt.Println("  go run cmd/init_rbac/main.go")
}

func fixRolesCode(db *gorm.DB) error {
	// 检查 code 列是否存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'roles' AND column_name = 'code'").Scan(&count)

	if count > 0 {
		// 更新空的 code 字段
		if err := db.Exec("UPDATE roles SET code = CONCAT('role_', id) WHERE code = '' OR code IS NULL").Error; err != nil {
			return err
		}
	}
	return nil
}

func fixUsersRoleID(db *gorm.DB) error {
	// 检查 role_id 列是否存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'role_id'").Scan(&count)

	if count == 0 {
		// 添加 role_id 列
		if err := db.Exec("ALTER TABLE users ADD COLUMN role_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '角色ID'").Error; err != nil {
			return err
		}
	}
	return nil
}
