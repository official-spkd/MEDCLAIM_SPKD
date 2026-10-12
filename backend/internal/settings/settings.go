// Package settings membaca pengaturan sistem (key/value).
package settings

import (
	"context"
	"strconv"

	"gorm.io/gorm"
)

type Setting struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `json:"value"`
}

func (Setting) TableName() string { return "settings" }

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) All(ctx context.Context) ([]Setting, error) {
	out := []Setting{}
	err := s.db.WithContext(ctx).Order("key").Find(&out).Error
	return out, err
}

// IntOr membaca setting bertipe angka; memakai bawaan bila tidak ada/rusak.
func (s *Service) IntOr(ctx context.Context, key string, def int) int {
	var st Setting
	if err := s.db.WithContext(ctx).First(&st, "key = ?", key).Error; err != nil {
		return def
	}
	if n, err := strconv.Atoi(st.Value); err == nil {
		return n
	}
	return def
}
