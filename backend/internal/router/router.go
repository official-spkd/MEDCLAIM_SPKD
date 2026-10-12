package router

import (
	"github.com/gin-gonic/gin"

	"medclaim-backend/internal/handler"
)

func New(db handler.Pinger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	health := handler.NewHealthHandler(db)
	r.GET("/health", health.Live)
	r.GET("/ready", health.Ready)

	// Modul berikutnya (auth, dashboard, claims, ...) didaftarkan di grup ini.
	_ = r.Group("/api/v1")

	return r
}
