// Package admin: manajemen pengguna dan jejak audit (khusus SUPERADMIN).
package admin

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"medclaim-backend/internal/audit"
	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/hospitals"
	"medclaim-backend/internal/response"
	"medclaim-backend/internal/types"
	"medclaim-backend/internal/users"
)

type Service struct {
	db        *gorm.DB
	audit     *audit.Service
	hospitals *hospitals.Service
}

func NewService(db *gorm.DB, a *audit.Service, h *hospitals.Service) *Service {
	return &Service{db: db, audit: a, hospitals: h}
}

func (s *Service) Users(ctx context.Context) ([]users.User, error) {
	out := []users.User{}
	err := s.db.WithContext(ctx).Order("created_at").Find(&out).Error
	return out, err
}

type BodyUser struct {
	Email        string  `json:"email"`
	Nama         string  `json:"nama"`
	Role         string  `json:"role"`
	HospitalID   *string `json:"hospitalId"`
	Spesialisasi *string `json:"spesialisasi"`
	DpjpKode     *string `json:"dpjpKode"`
	Password     string  `json:"password"`
}

const hurufSandi = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789"

func sandiAcak(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = hurufSandi[int(b[i])%len(hurufSandi)]
	}
	return string(b), nil
}

type UserDibuat struct {
	users.User
	// PasswordSementara hanya muncul sekali, pada respons pembuatan akun, bila
	// admin tidak mengisi password sendiri.
	PasswordSementara string `json:"passwordSementara,omitempty"`
}

func (s *Service) BuatUser(ctx context.Context, p *auth.Principal, b BodyUser) (*UserDibuat, error) {
	email := strings.ToLower(strings.TrimSpace(b.Email))
	nama := strings.TrimSpace(b.Nama)
	if email == "" || !strings.Contains(email, "@") || nama == "" || !users.PeranValid(b.Role) {
		return nil, response.Unprocessable("Email, nama, dan peran wajib diisi dengan benar.")
	}
	// Aturan tenant: SUPERADMIN lintas RS; peran lain wajib terikat satu RS yang ada.
	var hid *string
	if b.Role == "SUPERADMIN" {
		if b.HospitalID != nil && *b.HospitalID != "" {
			return nil, response.Unprocessable("Super Admin tidak boleh terikat ke rumah sakit.")
		}
	} else {
		if b.HospitalID == nil || !types.IsUUID(*b.HospitalID) {
			return nil, response.Unprocessable("Rumah sakit wajib dipilih untuk peran ini.")
		}
		ok, err := s.hospitals.Exists(ctx, *b.HospitalID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, response.Unprocessable("Rumah sakit tidak ditemukan.")
		}
		hid = b.HospitalID
	}
	if b.Role == "DPJP" && (b.DpjpKode == nil || strings.TrimSpace(*b.DpjpKode) == "") {
		return nil, response.Unprocessable("Kode DPJP wajib diisi untuk peran DPJP.")
	}

	sandi := b.Password
	sementara := ""
	if sandi == "" {
		var err error
		if sandi, err = sandiAcak(14); err != nil {
			return nil, err
		}
		sementara = sandi
	} else if len(sandi) < 8 {
		return nil, response.Unprocessable("Password minimal 8 karakter.")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(sandi), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := users.User{
		Email: email, PasswordHash: string(hash), Nama: nama, Role: b.Role, HospitalID: hid,
		Spesialisasi: blankKeNil(b.Spesialisasi), DpjpKode: blankKeNil(b.DpjpKode), Aktif: true,
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var n int64
		if err := tx.Model(&users.User{}).Where("lower(email) = ?", email).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return response.Conflict("Email sudah terdaftar.")
		}
		if err := tx.Create(&u).Error; err != nil {
			return err
		}
		s.audit.Catat(tx, p.Email, "USER_CREATE", "USER", &u.ID, &u.Email)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &UserDibuat{User: u, PasswordSementara: sementara}, nil
}

func blankKeNil(s *string) *string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	v := strings.TrimSpace(*s)
	return &v
}

func (s *Service) Toggle(ctx context.Context, p *auth.Principal, id string) (*users.User, error) {
	if !types.IsUUID(id) {
		return nil, response.NotFound("Pengguna tidak ditemukan.")
	}
	if id == p.ID {
		return nil, response.Unprocessable("Tidak dapat menonaktifkan akun sendiri.")
	}
	var u users.User
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&u, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("Pengguna tidak ditemukan.")
			}
			return err
		}
		u.Aktif = !u.Aktif
		if err := tx.Model(&users.User{}).Where("id = ?", id).Update("aktif", u.Aktif).Error; err != nil {
			return err
		}
		if !u.Aktif { // akun nonaktif langsung kehilangan semua sesinya
			if err := tx.Exec("DELETE FROM sessions WHERE user_id = ?", id).Error; err != nil {
				return err
			}
		}
		ket := "nonaktif"
		if u.Aktif {
			ket = "aktif"
		}
		s.audit.Catat(tx, p.Email, "USER_TOGGLE", "USER", &id, &ket)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

type Handler struct {
	svc   *Service
	audit *audit.Service
}

func NewHandler(svc *Service, a *audit.Service) *Handler { return &Handler{svc: svc, audit: a} }

// Users: GET /api/admin/users
func (h *Handler) Users(c *gin.Context) {
	items, err := h.svc.Users(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// BuatUser: POST /api/admin/users
func (h *Handler) BuatUser(c *gin.Context) {
	var body BodyUser
	if err := response.Bind(c, &body); err != nil {
		response.Fail(c, err)
		return
	}
	out, err := h.svc.BuatUser(c.Request.Context(), auth.Must(c), body)
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Toggle: POST /api/admin/users/:id/toggle
func (h *Handler) Toggle(c *gin.Context) {
	out, err := h.svc.Toggle(c.Request.Context(), auth.Must(c), c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Audit: GET /api/admin/audit
func (h *Handler) Audit(c *gin.Context) {
	items, err := h.audit.Terbaru(c.Request.Context(), 200)
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
