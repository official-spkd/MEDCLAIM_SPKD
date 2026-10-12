// Package bootstrap menyediakan satu endpoint pemuat awal untuk frontend:
// sesi, daftar RS, upload, template, pengaturan, dan daftar dokter.
package bootstrap

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/hospitals"
	"medclaim-backend/internal/queries"
	"medclaim-backend/internal/response"
	"medclaim-backend/internal/settings"
	"medclaim-backend/internal/uploads"
	"medclaim-backend/internal/users"
)

type Handler struct {
	db        *gorm.DB
	hospitals *hospitals.Service
	uploads   *uploads.Service
	queries   *queries.Service
	settings  *settings.Service
}

func NewHandler(db *gorm.DB, h *hospitals.Service, u *uploads.Service, q *queries.Service, s *settings.Service) *Handler {
	return &Handler{db: db, hospitals: h, uploads: u, queries: q, settings: s}
}

// Get: GET /api/bootstrap
func (h *Handler) Get(c *gin.Context) {
	p := auth.Must(c)
	ctx := c.Request.Context()

	// Non-superadmin hanya melihat RS-nya sendiri (isolasi tenant).
	var hid *string
	if !p.IsSuperadmin() {
		hid = p.HospitalID
	}
	hs, err := h.hospitals.List(ctx, hid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	ups, err := h.uploads.List(ctx, p)
	if err != nil {
		response.Fail(c, err)
		return
	}
	tpl, err := h.queries.Templates(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	st, err := h.settings.All(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}

	dokter := []users.User{}
	dq := h.db.WithContext(ctx).Where("role = 'DPJP' AND aktif").Order("nama")
	if hid != nil {
		dq = dq.Where("hospital_id = ?", *hid)
	}
	if err := dq.Find(&dokter).Error; err != nil {
		response.Fail(c, err)
		return
	}

	resp := gin.H{
		"me": p, "hospitals": hs, "uploads": ups, "templates": tpl,
		"settings": st, "dokter": dokter,
	}
	if p.IsSuperadmin() {
		all := []users.User{}
		if err := h.db.WithContext(ctx).Order("created_at").Find(&all).Error; err != nil {
			response.Fail(c, err)
			return
		}
		resp["users"] = all
	}
	c.JSON(http.StatusOK, resp)
}
