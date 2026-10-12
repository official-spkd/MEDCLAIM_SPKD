package main

import (
	"context"
	"flag"
	"log"
	"time"

	"medclaim-backend/internal/config"
	"medclaim-backend/internal/database"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	fresh := flag.Bool("fresh", false, "HAPUS seluruh isi schema public lalu migrasi ulang dari awal (hanya development)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if *fresh {
		if cfg.Env == "production" {
			log.Fatal("-fresh ditolak: APP_ENV=production")
		}
		log.Println("-fresh: menghapus schema public dan membuatnya ulang")
		if err := database.Reset(ctx, pool); err != nil {
			return err
		}
	}

	if err := database.Migrate(ctx, pool); err != nil {
		return err
	}
	log.Println("migrasi selesai")
	return nil
}
