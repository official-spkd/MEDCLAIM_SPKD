// Package uploads: unggah ekspor V-Claim, validasi + skor kualitas data (DQ),
// versioning episode, dan pembuatan kandidat ledger otomatis.
package uploads

import (
	"time"

	"medclaim-backend/internal/types"
)

// Batch memetakan upload_batches; tag json menjadikannya tipe Upload di frontend.
type Batch struct {
	ID             string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HospitalID     string      `gorm:"type:uuid" json:"hospitalId"`
	UploadedByID   string      `gorm:"type:uuid" json:"uploadedById"`
	NamaFile       string      `json:"namaFile"`
	Periode        string      `json:"periode"`
	Format         string      `json:"format"`
	DQScore        int         `gorm:"column:dq_score" json:"dqScore"`
	DQGrade        string      `gorm:"column:dq_grade" json:"dqGrade"`
	DQReport       types.JSONB `gorm:"column:dq_report" json:"dqReport,omitempty"`
	Status         string      `json:"status"`
	TotalBaris     int         `json:"totalBaris"`
	BarisValid     int         `json:"barisValid"`
	BarisGagal     int         `json:"barisGagal"`
	BarisBaru      int         `json:"barisBaru"`
	BarisBerubah   int         `json:"barisBerubah"`
	BarisUnchanged int         `json:"barisUnchanged"`
	OverrideUsed   bool        `json:"overrideUsed"`
	AlasanOverride *string     `json:"alasanOverride,omitempty"`
	CreatedAt      time.Time   `json:"createdAt"`
}

func (Batch) TableName() string { return "upload_batches" }

type FieldDiff struct {
	Field string `json:"field"`
	Label string `json:"label"`
	Lama  string `json:"lama"`
	Baru  string `json:"baru"`
}

type HasilIngest struct {
	Upload         *Batch                 `json:"upload"`
	BarisBaru      int                    `json:"barisBaru"`
	BarisBerubah   int                    `json:"barisBerubah"`
	BarisUnchanged int                    `json:"barisUnchanged"`
	LedgerBaru     int                    `json:"ledgerBaru"`
	Diff           map[string][]FieldDiff `json:"diff"`
}

// Permintaan = body preview & commit.
type Permintaan struct {
	HospitalID     string `json:"hospitalId"`
	Periode        string `json:"periode"`
	NamaFile       string `json:"namaFile"`
	Isi            string `json:"isi"`
	Override       bool   `json:"override"`
	AlasanOverride string `json:"alasanOverride"`
}
