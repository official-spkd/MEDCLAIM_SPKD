package main

import (
	"log"

	"medclaim-backend/internal/config"
	"medclaim-backend/internal/database"
	"medclaim-backend/internal/seed"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Env == "production" {
		log.Println("seed dilewati: APP_ENV=production")
		return nil
	}
	db, err := database.ConnectGORM(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	if sqlDB, err := db.DB(); err == nil {
		defer sqlDB.Close()
	}
	if err := seed.Run(db, cfg.SeedPassword); err != nil {
		return err
	}
	log.Println("seed selesai")
	return nil
}
