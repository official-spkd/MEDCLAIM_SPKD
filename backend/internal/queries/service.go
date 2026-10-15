package queries

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"medclaim-backend/internal/audit"
	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/response"
	"medclaim-backend/internal/settings"
	"medclaim-backend/internal/types"
)

type Service struct {
	db       *gorm.DB
	audit    *audit.Service
	settings *settings.Service
}

func NewService(db *gorm.DB, a *audit.Service, st *settings.Service) *Service {
	return &Service{db: db, audit: a, settings: st}
}

const kolomDTO = `q.id::text AS id, q.hospital_id::text AS hospital_id, q.sep, q.nomor_mr, q.inisial,
	to_char(q.tgl_masuk, 'YYYY-MM-DD') AS tgl_masuk, to_char(q.tgl_keluar, 'YYYY-MM-DD') AS tgl_keluar,
	q.dpjp_id::text AS dpjp_id, q.dpjp_nama, u.spesialisasi AS dpjp_spesialisasi,
	q.template_nama, q.pertanyaan, q.konteks, q.status::text AS status, q.deadline,
	q.sent_at, q.answered_at, q.jawaban, q.sla_hari, q.created_at`

func (s *Service) base(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx).Table("document_queries q").
		Joins("LEFT JOIN users u ON u.id = q.dpjp_id").Select(kolomDTO)
}

// status efektif: TERKIRIM yang lewat tenggat tampil sebagai TERLAMBAT.
func tampilkan(d *DTO) {
	if d.Status == "TERKIRIM" && d.Deadline.Before(time.Now()) {
		d.Status = "TERLAMBAT"
	}
}

func (s *Service) Templates(ctx context.Context) ([]Template, error) {
	out := []Template{}
	err := s.db.WithContext(ctx).Order("nama").Find(&out).Error
	return out, err
}

func (s *Service) List(ctx context.Context, p *auth.Principal, status string) ([]DTO, error) {
	q := s.base(ctx)
	hid, err := p.ScopeHospital("")
	if err != nil {
		return nil, err
	}
	if hid != "" {
		q = q.Where("q.hospital_id = ?", hid)
	}
	if p.Role == "DPJP" {
		q = q.Where("q.dpjp_id = ?", p.ID)
	}
	items := []DTO{}
	if err := q.Order("q.created_at DESC").Scan(&items).Error; err != nil {
		return nil, err
	}
	out := make([]DTO, 0, len(items))
	for i := range items {
		tampilkan(&items[i])
		if status != "" && status != "SEMUA" && items[i].Status != status {
			continue
		}
		out = append(out, items[i])
	}
	return out, nil
}

func (s *Service) get(ctx context.Context, id string) (*DTO, error) {
	var out []DTO
	if err := s.base(ctx).Where("q.id = ?", id).Scan(&out).Error; err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, response.NotFound("Query tidak ditemukan.")
	}
	tampilkan(&out[0])
	return &out[0], nil
}

type BodyBuat struct {
	Sep        string `json:"sep"`
	DpjpID     string `json:"dpjpId"`
	TemplateID string `json:"templateId"`
	Pertanyaan string `json:"pertanyaan"`
	Konteks    string `json:"konteks"`
}

func inisial(nama string) string {
	r := func(s string) string { return strings.ToUpper(string([]rune(s)[:1])) }
	parts := strings.Fields(nama)
	switch len(parts) {
	case 0:
		return "XX"
	case 1:
		rs := []rune(strings.ToUpper(parts[0]))
		if len(rs) >= 2 {
			return string(rs[:2])
		}
		return string(rs)
	}
	return r(parts[0]) + r(parts[1])
}

func (s *Service) Buat(ctx context.Context, p *auth.Principal, b BodyBuat) (*DTO, error) {
	sep := strings.TrimSpace(b.Sep)
	if sep == "" {
		return nil, response.Unprocessable("SEP wajib diisi.")
	}
	if !types.IsUUID(b.DpjpID) {
		return nil, response.NotFound("DPJP tidak ditemukan.")
	}

	// Episode aktif untuk SEP ini, dalam cakupan tenant pemanggil.
	var eps []struct {
		HospitalID, NoMr, NamaPasien string
		DpjpKode                     *string
		TglMasuk, TglKeluar          string
	}
	eq := s.db.WithContext(ctx).Table("episodes").
		Select(`hospital_id::text AS hospital_id, no_mr, nama_pasien, dpjp_kode,
			to_char(tgl_masuk,'YYYY-MM-DD') AS tgl_masuk, to_char(tgl_keluar,'YYYY-MM-DD') AS tgl_keluar`).
		Where("no_sep = ? AND superseded = false", sep)
	if !p.IsSuperadmin() {
		if p.HospitalID == nil {
			return nil, response.Forbidden("Akun belum terikat ke rumah sakit.")
		}
		eq = eq.Where("hospital_id = ?", *p.HospitalID)
	}
	if err := eq.Scan(&eps).Error; err != nil {
		return nil, err
	}
	if len(eps) == 0 {
		return nil, response.NotFound("SEP tidak ditemukan pada data aktif.")
	}
	if len(eps) > 1 {
		return nil, response.Unprocessable("SEP ditemukan di lebih dari satu rumah sakit; gunakan akun RS terkait.")
	}
	ep := eps[0]

	var dpjp struct {
		ID, Nama string
		DpjpKode *string
		Hospital *string
	}
	if err := s.db.WithContext(ctx).Table("users").
		Select("id::text AS id, nama, dpjp_kode, hospital_id::text AS hospital").
		Where("id = ? AND role = 'DPJP' AND aktif", b.DpjpID).Scan(&dpjp).Error; err != nil {
		return nil, err
	}
	if dpjp.ID == "" || dpjp.Hospital == nil || *dpjp.Hospital != ep.HospitalID {
		return nil, response.NotFound("DPJP tidak ditemukan.")
	}
	if ep.DpjpKode == nil || dpjp.DpjpKode == nil || *ep.DpjpKode != *dpjp.DpjpKode {
		return nil, response.Unprocessable("DPJP yang dipilih bukan DPJP pada episode ini.")
	}

	pertanyaan := strings.TrimSpace(b.Pertanyaan)
	var tplNama *string
	if b.TemplateID != "" {
		if !types.IsUUID(b.TemplateID) {
			return nil, response.Unprocessable("Template tidak valid.")
		}
		var t Template
		if err := s.db.WithContext(ctx).First(&t, "id = ? AND aktif", b.TemplateID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, response.Unprocessable("Template tidak ditemukan.")
			}
			return nil, err
		}
		tplNama = &t.Nama
		if pertanyaan == "" {
			pertanyaan = t.Isi
		}
	}
	if pertanyaan == "" {
		return nil, response.Unprocessable("Pertanyaan wajib diisi.")
	}

	sla := s.settings.IntOr(ctx, "sla.query_hari", 3)
	now := time.Now()
	masuk, _ := time.Parse("2006-01-02", ep.TglMasuk)
	keluar, _ := time.Parse("2006-01-02", ep.TglKeluar)
	q := DocumentQuery{
		HospitalID: ep.HospitalID, Sep: sep, NomorMr: ep.NoMr, Inisial: inisial(ep.NamaPasien),
		TglMasuk: &masuk, TglKeluar: &keluar, DpjpID: dpjp.ID, DpjpNama: dpjp.Nama,
		TemplateNama: tplNama, Pertanyaan: pertanyaan, Konteks: b.Konteks,
		Status: "TERKIRIM", Deadline: now.AddDate(0, 0, sla), SentAt: &now, SlaHari: sla,
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Create(&q).Error; err != nil {
			return err
		}
		s.audit.Catat(tx, p.Email, "QUERY_CREATE", "QUERY", &q.ID, &sep)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.get(ctx, q.ID)
}

func (s *Service) Jawab(ctx context.Context, p *auth.Principal, id, jawaban string) (*DTO, error) {
	if !types.IsUUID(id) {
		return nil, response.NotFound("Query tidak ditemukan.")
	}
	jawaban = strings.TrimSpace(jawaban)

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var q DocumentQuery
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&q, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("Query tidak ditemukan.")
			}
			return err
		}
		// Isolasi tenant: query RS lain dianggap tidak ada.
		if !p.IsSuperadmin() && (p.HospitalID == nil || q.HospitalID != *p.HospitalID) {
			return response.NotFound("Query tidak ditemukan.")
		}
		if q.Status == "DIJAWAB" {
			return response.Conflict("Query sudah dijawab — jawaban terkunci.")
		}
		// Hanya DPJP tujuan yang boleh menjawab; Casemix/Coder/Superadmin hanya mengajukan.
		if p.Role != "DPJP" {
			return response.Forbidden("Hanya DPJP yang dituju yang dapat menjawab query.")
		}
		if q.DpjpID != p.ID {
			return response.Forbidden("Query ini bukan untuk DPJP Anda.")
		}
		if len([]rune(jawaban)) < 5 {
			return response.Unprocessable("Jawaban minimal 5 karakter.")
		}
		if err := tx.Model(&DocumentQuery{}).Where("id = ?", id).
			Updates(map[string]any{"status": "DIJAWAB", "answered_at": time.Now(), "jawaban": jawaban}).Error; err != nil {
			return err
		}
		s.audit.Catat(tx, p.Email, "QUERY_ANSWER", "QUERY", &id, nil)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.get(ctx, id)
}
