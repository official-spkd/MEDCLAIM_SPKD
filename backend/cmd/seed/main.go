package main

import (
	"context"
	"log"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := seed.Run(ctx, pool); err != nil {
		return err
	}
	log.Println("seed selesai")
	return nil
}
