// Package router merangkai seluruh modul menjadi satu mesin HTTP Gin.
package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"medclaim-backend/internal/admin"
	"medclaim-backend/internal/analytics"
	"medclaim-backend/internal/audit"
	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/bootstrap"
	"medclaim-backend/internal/claims"
	"medclaim-backend/internal/config"
	"medclaim-backend/internal/health"
	"medclaim-backend/internal/hospitals"
	"medclaim-backend/internal/ledger"
	"medclaim-backend/internal/middleware"
	"medclaim-backend/internal/queries"
	"medclaim-backend/internal/settings"
	"medclaim-backend/internal/uploads"
)

func New(cfg *config.Config, db *gorm.DB) (*gin.Engine, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("sql.DB: %w", err)
	}

	r := gin.New()
	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	r.Use(gin.Logger(), gin.Recovery(), middleware.SecurityHeaders(), middleware.CORS(cfg.CORSOrigins))

	// ---- Layanan (dependency injection manual) ----
	auditSvc := audit.NewService(db)
	hospitalSvc := hospitals.NewService(db)
	settingsSvc := settings.NewService(db)
	authSvc := auth.NewService(db, auditSvc, cfg.SessionTTL)
	claimsSvc := claims.NewService(db)
	ledgerSvc := ledger.NewService(db, auditSvc)
	queriesSvc := queries.NewService(db, auditSvc, settingsSvc)
	uploadsSvc := uploads.NewService(db, auditSvc, hospitalSvc)
	analyticsSvc := analytics.NewService(db, claimsSvc)
	adminSvc := admin.NewService(db, auditSvc, hospitalSvc)

	// ---- Handler ----
	healthH := health.NewHandler(sqlDB)
	authH := auth.NewHandler(authSvc, auth.CookieOpts{Secure: cfg.CookieSecure, SameSite: cfg.CookieSameSite})
	bootstrapH := bootstrap.NewHandler(db, hospitalSvc, uploadsSvc, queriesSvc, settingsSvc)
	claimsH := claims.NewHandler(claimsSvc)
	analyticsH := analytics.NewHandler(analyticsSvc)
	uploadsH := uploads.NewHandler(uploadsSvc)
	ledgerH := ledger.NewHandler(ledgerSvc)
	queriesH := queries.NewHandler(queriesSvc)
	adminH := admin.NewHandler(adminSvc, auditSvc)

	r.GET("/health", healthH.Live)
	r.GET("/ready", healthH.Ready)

	api := r.Group("/api")
	api.POST("/auth/login", authH.Login)
	api.POST("/auth/logout", middleware.OptionalAuth(authSvc), authH.Logout)

	// ---- Semua route di bawah ini butuh sesi ----
	a := api.Group("", middleware.Auth(authSvc))
	a.GET("/auth/me", authH.Me)
	a.GET("/bootstrap", bootstrapH.Get)

	a.GET("/dashboard", analyticsH.Dashboard)
	a.GET("/anomali", analyticsH.Anomali)
	a.GET("/financial", analyticsH.Financial)

	a.GET("/episodes", claimsH.List)
	a.GET("/episodes/:id", claimsH.Detail)

	a.GET("/uploads", uploadsH.List)
	up := a.Group("", middleware.RequireAccess(auth.GrupUpload, "Hanya Super Admin dan Casemix yang dapat mengunggah data."))
	up.POST("/uploads/preview", uploadsH.Preview)
	up.POST("/uploads/commit", uploadsH.Commit)
	up.POST("/uploads/:id/arsip", uploadsH.Arsip)
	up.GET("/samples", uploadsH.Samples)
	up.GET("/samples/:nama", uploadsH.SampleIsi)
	a.DELETE("/uploads/:id", middleware.RequireRole("Hanya Super Admin yang dapat menghapus upload.", "SUPERADMIN"), uploadsH.Hapus)

	lg := a.Group("/ledger", middleware.RequireAccess(auth.GrupLedger, "Peran Anda tidak memiliki akses ke Recovery Ledger."))
	lg.GET("", ledgerH.List)
	lg.GET("/:id", ledgerH.Detail)
	lgW := lg.Group("", middleware.RequireMutate(auth.GrupLedger, "Peran Anda hanya dapat melihat ledger (baca-saja)."))
	lgW.POST("", ledgerH.Buat)
	lgW.POST("/:id/transisi", ledgerH.Transisi)

	a.GET("/queries", middleware.RequireRole("Peran Anda tidak memiliki akses modul query.", "SUPERADMIN", "CASEMIX", "CODER", "DPJP"), queriesH.List)
	a.POST("/queries", middleware.RequireMutate(auth.GrupQuery, "Peran Anda tidak dapat membuat query."), queriesH.Buat)
	a.POST("/queries/:id/answer", queriesH.Jawab) // otorisasi rinci (DPJP pemilik / casemix) ada di service
	a.GET("/templates", queriesH.Templates)

	ad := a.Group("/admin", middleware.RequireAccess(auth.GrupAdmin, "Hanya Super Admin yang dapat mengakses administrasi."))
	ad.GET("/users", adminH.Users)
	ad.POST("/users", adminH.BuatUser)
	ad.POST("/users/:id/toggle", adminH.Toggle)
	ad.GET("/audit", adminH.Audit)

	return r, nil
}
