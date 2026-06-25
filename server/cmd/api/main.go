// Package main 启动智能眼镜巡检 API 服务。它负责加载运行配置、连接本地基础设施，
// 并把各业务服务装配到 Gin HTTP 路由中，供后台 Web 和 Android 眼镜端共同调用。
package main

import (
	"log"
	_ "time/tzdata"

	"aiglasses/server/internal/attachments"
	"aiglasses/server/internal/auth"
	"aiglasses/server/internal/businesscodes"
	"aiglasses/server/internal/config"
	"aiglasses/server/internal/defects"
	"aiglasses/server/internal/devices"
	"aiglasses/server/internal/events"
	"aiglasses/server/internal/glass_api"
	"aiglasses/server/internal/httpapi"
	"aiglasses/server/internal/nodes"
	"aiglasses/server/internal/organizations"
	"aiglasses/server/internal/plans"
	"aiglasses/server/internal/points"
	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/tasks"
	"aiglasses/server/internal/templates"
	"aiglasses/server/internal/users"
	"aiglasses/server/internal/workflows"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// main 是 API 服务入口，负责按配置完成数据库、缓存、对象存储和 HTTP 路由初始化。
func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}
	var redisClient *redis.Client
	if cfg.RedisAddr != "" {
		redisClient = redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword})
	}
	businessCodeSvc, err := businesscodes.NewService(db, redisClient)
	if err != nil {
		log.Fatal(err)
	}
	authSvc := auth.NewServiceWithRefresh(db, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	defectSvc := defects.NewService(db)
	taskSvc := tasks.NewService(db, redisClient)
	attachmentSvc, err := attachments.NewService(db, cfg)
	if err != nil {
		log.Fatal(err)
	}
	scheduler := events.NewScheduler(db, cfg)
	handler := httpapi.NewHandler(
		authSvc,
		attachmentSvc,
		businessCodeSvc,
		defectSvc,
		devices.NewService(db),
		nodes.NewService(db),
		organizations.NewService(db),
		plans.NewService(db),
		points.NewService(db),
		taskSvc,
		templates.NewService(db),
		users.NewService(db),
		workflows.NewService(db),
		scheduler,
	)
	r := gin.Default()
	handler.Register(r)
	glass_api.NewHandlerWithRuntime(authSvc, taskSvc, defectSvc, db, cfg).Register(r)
	log.Fatal(r.Run(cfg.HTTPAddr))
}
