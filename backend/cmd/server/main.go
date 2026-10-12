package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"medclaim-backend/internal/config"
	"medclaim-backend/internal/database"
	"medclaim-backend/internal/router"
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
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.ConnectGORM(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	if sqlDB, err := db.DB(); err == nil {
		defer sqlDB.Close()
	}

	engine, err := router.New(cfg, db)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second, // upload file besar
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("MedClaim berjalan di :%s (env=%s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case <-stop:
		log.Println("menghentikan server...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
