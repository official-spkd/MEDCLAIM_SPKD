// Package claims: episode klaim (SEP) beserta diagnosis, prosedur, komponen
// biaya, dan add-on. Penulisan episode dilakukan modul uploads.
package claims

import (
	"time"

	"medclaim-backend/internal/types"
)

type Episode struct {
	ID            string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	HospitalID    string `gorm:"type:uuid"`
	UploadBatchID string `gorm:"type:uuid"`
	KodeIcbg      string
	NoSep         string
	NoKartu       string
	NoMr          string
	NamaPasien    string
	TglLahir      types.Date
	JenisKelamin  string
	TglMasuk      types.Date
	TglKeluar     types.Date
	Los           int
	KelasRawat    int
	TipeLayanan   string
	CaraPulang    *string
	DpjpKode      *string
	DpjpNama      *string
	SeverityLevel int
	IdrgKode      *string
	TotalBilling  int64
	TarifInacbg   int64
	Selisih       int64
	// PctRugi tersimpan sebagai persen 0..100 (kolom numeric(8,2)); API
	// mengeluarkannya sebagai pecahan 0..1 seperti yang diharapkan frontend.
	PctRugi    float64
	KelasRugi  string
	DQScore    int `gorm:"column:dq_score"`
	RowHash    string
	VersionNo  int
	Superseded bool
	CreatedAt  time.Time

	Diagnoses  []EpisodeDiagnosis        `gorm:"foreignKey:EpisodeID"`
	Procedures []EpisodeProcedure        `gorm:"foreignKey:EpisodeID"`
	Components []EpisodeBillingComponent `gorm:"foreignKey:EpisodeID"`
	Addons     []EpisodeAddon            `gorm:"foreignKey:EpisodeID"`
	Icbg       *RefInacbg                `gorm:"foreignKey:KodeIcbg;references:Kode"`
}

func (Episode) TableName() string { return "episodes" }

type EpisodeDiagnosis struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	EpisodeID string `gorm:"type:uuid"`
	Urutan    int
	KodeIcd10 string `gorm:"column:kode_icd10"`
	Nama      string
}

func (EpisodeDiagnosis) TableName() string { return "episode_diagnoses" }

type EpisodeProcedure struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	EpisodeID string `gorm:"type:uuid"`
	KodeIcd9  string `gorm:"column:kode_icd9"`
	Nama      string
}

func (EpisodeProcedure) TableName() string { return "episode_procedures" }

type EpisodeBillingComponent struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	EpisodeID string `gorm:"type:uuid"`
	Komponen  string
	Nilai     int64
}

func (EpisodeBillingComponent) TableName() string { return "episode_billing_components" }

type EpisodeAddon struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	EpisodeID string `gorm:"type:uuid"`
	Kode      string
	Jenis     string
	Tarif     int64
	Status    string
}

func (EpisodeAddon) TableName() string { return "episode_addons" }

type RefInacbg struct {
	Kode      string `gorm:"primaryKey"`
	Deskripsi string
	Tarif     int64
}

func (RefInacbg) TableName() string { return "ref_inacbg" }
