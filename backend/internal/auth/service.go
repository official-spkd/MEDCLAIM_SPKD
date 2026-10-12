package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"medclaim-backend/internal/audit"
	"medclaim-backend/internal/response"
	"medclaim-backend/internal/types"
	"medclaim-backend/internal/users"
)

type Session struct {
	ID        string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string `gorm:"type:uuid"`
	CreatedAt time.Time
	ExpiresAt time.Time
}

func (Session) TableName() string { return "sessions" }

type Service struct {
	db    *gorm.DB
	audit *audit.Service
	ttl   time.Duration
}

func NewService(db *gorm.DB, a *audit.Service, ttl time.Duration) *Service {
	return &Service{db: db, audit: a, ttl: ttl}
}

// Hash dummy untuk menyamakan waktu respons saat email tidak ditemukan
// (mencegah enumerasi akun lewat perbedaan waktu).
var hashDummy, _ = bcrypt.GenerateFromPassword([]byte("dummy-medclaim"), bcrypt.DefaultCost)

var errLoginGagal = response.Unauthorized("Email atau kata sandi salah.")

// Login memverifikasi kredensial dan membuat sesi. Mengembalikan token cookie.
func (s *Service) Login(ctx context.Context, email, password string) (*Principal, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var u users.User
	err := s.db.WithContext(ctx).Where("lower(email) = ? AND aktif", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		bcrypt.CompareHashAndPassword(hashDummy, []byte(password))
		return nil, "", errLoginGagal
	}
	if err != nil {
		return nil, "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, "", errLoginGagal
	}

	sess := Session{UserID: u.ID, ExpiresAt: time.Now().Add(s.ttl)}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&sess).Error; err != nil {
			return err
		}
		if err := tx.Model(&users.User{}).Where("id = ?", u.ID).
			Update("last_login_at", time.Now()).Error; err != nil {
			return err
		}
		// Bersih-bersih sesi kedaluwarsa secara oportunistik.
		tx.Where("expires_at < now()").Delete(&Session{})
		s.audit.Catat(tx, u.Email, "LOGIN", "USER", &u.ID, nil)
		return nil
	})
	if err != nil {
		return nil, "", err
	}
	p, err := s.loadPrincipal(ctx, u.ID)
	if err != nil {
		return nil, "", err
	}
	return p, sess.ID, nil
}

// Resolve mengubah token cookie menjadi Principal; nil bila tidak valid/kedaluwarsa.
func (s *Service) Resolve(ctx context.Context, token string) (*Principal, error) {
	if !types.IsUUID(token) {
		return nil, nil
	}
	var userID string
	err := s.db.WithContext(ctx).Raw(`
		SELECT s.user_id::text FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = ? AND s.expires_at > now() AND u.aktif`, token).Scan(&userID).Error
	if err != nil {
		return nil, err
	}
	if userID == "" {
		return nil, nil
	}
	return s.loadPrincipal(ctx, userID)
}

func (s *Service) Logout(ctx context.Context, token string, p *Principal) error {
	if types.IsUUID(token) {
		if err := s.db.WithContext(ctx).Delete(&Session{}, "id = ?", token).Error; err != nil {
			return err
		}
	}
	if p != nil {
		s.audit.Catat(nil, p.Email, "LOGOUT", "USER", &p.ID, nil)
	}
	return nil
}

func (s *Service) loadPrincipal(ctx context.Context, userID string) (*Principal, error) {
	var p Principal
	err := s.db.WithContext(ctx).Raw(`
		SELECT u.id::text AS id, u.email, u.nama, u.role::text AS role,
		       u.hospital_id::text AS hospital_id,
		       h.nama AS hospital_nama, h.kode AS hospital_kode,
		       u.spesialisasi, u.dpjp_kode
		FROM users u LEFT JOIN hospitals h ON h.id = u.hospital_id
		WHERE u.id = ?`, userID).Scan(&p).Error
	if err != nil {
		return nil, err
	}
	if p.ID == "" {
		return nil, nil
	}
	return &p, nil
}

func (s *Service) TTL() time.Duration { return s.ttl }
