// Package queries: query dokumen — pertanyaan klarifikasi dari casemix/coder
// kepada DPJP, lengkap dengan tenggat (SLA) dan jawaban.
package queries

import "time"

type DocumentQuery struct {
	ID           string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	HospitalID   string `gorm:"type:uuid"`
	Sep          string
	NomorMr      string
	Inisial      string
	TglMasuk     *time.Time `gorm:"type:date"`
	TglKeluar    *time.Time `gorm:"type:date"`
	DpjpID       string     `gorm:"type:uuid"`
	DpjpNama     string
	TemplateNama *string
	Pertanyaan   string
	Konteks      string
	Status       string
	Deadline     time.Time
	SentAt       *time.Time
	AnsweredAt   *time.Time
	Jawaban      *string
	SlaHari      int
	CreatedAt    time.Time
}

func (DocumentQuery) TableName() string { return "document_queries" }

type Template struct {
	ID       string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Nama     string `json:"nama"`
	Isi      string `json:"isi"`
	Kategori string `json:"kategori"`
	Aktif    bool   `json:"aktif"`
}

func (Template) TableName() string { return "query_templates" }

// DTO = tipe QueryDok di frontend.
type DTO struct {
	ID               string     `json:"id"`
	HospitalID       string     `json:"hospitalId"`
	Sep              string     `json:"sep"`
	NomorMr          string     `json:"nomorMr"`
	Inisial          string     `json:"inisial"`
	TglMasuk         *string    `json:"tglMasuk"`
	TglKeluar        *string    `json:"tglKeluar"`
	DpjpID           string     `json:"dpjpId"`
	DpjpNama         string     `json:"dpjpNama"`
	DpjpSpesialisasi *string    `json:"dpjpSpesialisasi"`
	TemplateNama     *string    `json:"templateNama"`
	Pertanyaan       string     `json:"pertanyaan"`
	Konteks          string     `json:"konteks"`
	Status           string     `json:"status"`
	Deadline         time.Time  `json:"deadline"`
	SentAt           *time.Time `json:"sentAt"`
	AnsweredAt       *time.Time `json:"answeredAt"`
	Jawaban          *string    `json:"jawaban"`
	SlaHari          int        `json:"slaHari"`
	CreatedAt        time.Time  `json:"createdAt"`
}
