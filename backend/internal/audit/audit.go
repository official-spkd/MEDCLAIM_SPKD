// Package audit mencatat jejak aksi pengguna.
package audit

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"
)

type Log struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserEmail string    `json:"userEmail"`
	Aksi      string    `json:"aksi"`
	Entitas   string    `json:"entitas"`
	EntitasID *string   `json:"entitasId"`
	Meta      *string   `json:"meta"`
	Waktu     time.Time `gorm:"autoCreateTime" json:"waktu"`
}

func (Log) TableName() string { return "audit_logs" }

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// Catat menulis satu baris audit memakai tx (boleh transaksi pemanggil agar
// atomik dengan perubahannya; nil = koneksi biasa). Kegagalan audit dicatat ke
// log dan tidak menggagalkan aksi utama.
func (s *Service) Catat(tx *gorm.DB, email, aksi, entitas string, entitasID, meta *string) {
	if tx == nil {
		tx = s.db
	}
	l := Log{UserEmail: email, Aksi: aksi, Entitas: entitas, EntitasID: entitasID, Meta: meta}
	if err := tx.Create(&l).Error; err != nil {
		log.Printf("audit gagal (%s %s): %v", aksi, entitas, err)
	}
}

// Terbaru mengambil n log terbaru.
func (s *Service) Terbaru(ctx context.Context, n int) ([]Log, error) {
	out := []Log{}
	err := s.db.WithContext(ctx).Order("waktu DESC").Limit(n).Find(&out).Error
	return out, err
}

func Str(s string) *string { return &s }
