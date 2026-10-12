// Package ledger: Recovery Ledger — kandidat koreksi klaim, mesin status,
// riwayat perubahan, dan kalibrasi estimasi vs realisasi.
package ledger

import "time"

type Entry struct {
	ID          string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EpisodeID   string `gorm:"type:uuid"`
	HospitalID  string `gorm:"type:uuid"`
	Sep         string
	KelasRugi   string
	Estimasi    int64
	Realisasi   *int64
	Status      string
	Strategy    *string
	ReferensiRm *string
	Catatan     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Entry) TableName() string { return "ledger_entries" }

type Event struct {
	ID         string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	LedgerID   string `gorm:"type:uuid"`
	DariStatus *string
	KeStatus   *string
	AktorNama  string
	Catatan    *string
	Waktu      time.Time `gorm:"autoCreateTime"`
}

func (Event) TableName() string { return "ledger_events" }

// EntryDTO = tipe Ledger di frontend; nomorMr/namaPasien/kodeIcbg berasal dari episode.
type EntryDTO struct {
	ID          string     `json:"id"`
	EpisodeID   string     `json:"episodeId"`
	HospitalID  string     `json:"hospitalId"`
	Sep         string     `json:"sep"`
	NomorMr     string     `json:"nomorMr"`
	NamaPasien  string     `json:"namaPasien"`
	KodeIcbg    string     `json:"kodeIcbg"`
	KelasRugi   string     `json:"kelasRugi"`
	Estimasi    int64      `json:"estimasi"`
	Realisasi   *int64     `json:"realisasi"`
	Status      string     `json:"status"`
	ReferensiRm *string    `json:"referensiRm"`
	Catatan     *string    `json:"catatan"`
	Strategy    *string    `json:"strategy"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Events      []EventDTO `gorm:"-" json:"events"` // null di daftar, array (mungkin kosong) di detail
}

type EventDTO struct {
	ID         string    `json:"id"`
	DariStatus *string   `json:"dariStatus"`
	KeStatus   *string   `json:"keStatus"`
	AktorNama  string    `json:"aktorNama"`
	Catatan    *string   `json:"catatan"`
	Waktu      time.Time `json:"waktu"`
}
