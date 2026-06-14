// Package main 提供本地开发使用的一次性数据库初始化命令。它复用 API 服务相同的配置、
// 迁移逻辑和种子数据，用于在启动后台 Web 和眼镜端之前准备全新的 MySQL 数据库。
package main

import (
	"log"

	"aiglasses/server/internal/config"
	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/seed"
)

// main 是数据库初始化入口，负责执行迁移和写入本地开发种子数据。
func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}
	if err := seed.Run(db); err != nil {
		log.Fatal(err)
	}
	log.Println("database initialized")
	log.Println("admin login username: admin")
	log.Println("glasses login username: inspector, device_id: query devices table for GLASS-DEMO-001")
}
