// Package hospitals memegang data rumah sakit (tenant).
package hospitals

import (
	"context"

	"gorm.io/gorm"
)

type Hospital struct {
	ID    string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Kode  string `json:"kode"`
	Nama  string `json:"nama"`
	Kota  string `json:"kota"`
	Kelas string `json:"kelas"`
}

func (Hospital) TableName() string { return "hospitals" }

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// List mengembalikan semua RS (hospitalID nil) atau hanya satu RS.
func (s *Service) List(ctx context.Context, hospitalID *string) ([]Hospital, error) {
	out := []Hospital{}
	q := s.db.WithContext(ctx).Order("kode")
	if hospitalID != nil {
		q = q.Where("id = ?", *hospitalID)
	}
	if err := q.Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) Exists(ctx context.Context, id string) (bool, error) {
	var n int64
	err := s.db.WithContext(ctx).Model(&Hospital{}).Where("id = ?", id).Count(&n).Error
	return n > 0, err
}

func (s *Service) Get(ctx context.Context, id string) (*Hospital, error) {
	var h Hospital
	if err := s.db.WithContext(ctx).First(&h, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &h, nil
}
