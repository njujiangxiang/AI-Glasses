package main

import (
	"log"
	_ "time/tzdata"

	"aiglasses/server/internal/config"
	"aiglasses/server/internal/platform/database"
	"aiglasses/server/internal/platform/seed"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	if err := seed.Run(db); err != nil {
		log.Fatal(err)
	}
	log.Println("seed data written successfully")
	log.Println("admin login username: admin")
	log.Println("admin login password: admin")
}
