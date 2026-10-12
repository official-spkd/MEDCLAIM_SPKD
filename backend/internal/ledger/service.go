package ledger

import (
	"context"
	"errors"
	"math"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"medclaim-backend/internal/audit"
	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/response"
	"medclaim-backend/internal/scoring"
	"medclaim-backend/internal/types"
)

type Service struct {
	db    *gorm.DB
	audit *audit.Service
}

func NewService(db *gorm.DB, a *audit.Service) *Service { return &Service{db: db, audit: a} }

const kolomDTO = `l.id::text AS id, l.episode_id::text AS episode_id, l.hospital_id::text AS hospital_id,
	l.sep, e.no_mr AS nomor_mr, e.nama_pasien, e.kode_icbg,
	l.kelas_rugi::text AS kelas_rugi, l.estimasi, l.realisasi, l.status::text AS status,
	l.referensi_rm, l.catatan, l.strategy, l.created_at, l.updated_at`

func (s *Service) base(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx).Table("ledger_entries l").
		Joins("JOIN episodes e ON e.id = l.episode_id").Select(kolomDTO)
}

type ListResult struct {
	Items     []EntryDTO `json:"items"`
	Kalibrasi Kalibrasi  `json:"kalibrasi"`
}

func (s *Service) List(ctx context.Context, p *auth.Principal, rs, status, cari string) (*ListResult, error) {
	hid, err := p.ScopeHospital(rs)
	if err != nil {
		return nil, err
	}
	q := s.base(ctx)
	if hid != "" {
		q = q.Where("l.hospital_id = ?", hid)
	}
	if status != "" && status != "SEMUA" {
		if _, ok := sah[status]; !ok {
			return nil, response.BadRequest("Parameter status tidak valid.")
		}
		q = q.Where("l.status = ?", status)
	}
	if k := strings.ToLower(strings.TrimSpace(cari)); k != "" {
		like := "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(k) + "%"
		q = q.Where("(lower(l.sep) LIKE ? OR lower(e.nama_pasien) LIKE ? OR lower(e.no_mr) LIKE ?)", like, like, like)
	}
	items := []EntryDTO{}
	if err := q.Order("l.updated_at DESC").Scan(&items).Error; err != nil {
		return nil, err
	}

	// Kalibrasi dihitung atas seluruh ledger dalam cakupan tenant (bukan hasil filter status).
	var kal []Entry
	kq := s.db.WithContext(ctx).Select("estimasi, realisasi, status")
	if hid != "" {
		kq = kq.Where("hospital_id = ?", hid)
	}
	if err := kq.Find(&kal).Error; err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Kalibrasi: HitungKalibrasi(kal)}, nil
}

func (s *Service) Detail(ctx context.Context, p *auth.Principal, id string) (*EntryDTO, error) {
	if !types.IsUUID(id) {
		return nil, response.NotFound("Entri ledger tidak ditemukan.")
	}
	return s.detail(s.base(ctx), p, id)
}

func (s *Service) detail(q *gorm.DB, p *auth.Principal, id string) (*EntryDTO, error) {
	q = q.Where("l.id = ?", id)
	if !p.IsSuperadmin() {
		if p.HospitalID == nil {
			return nil, response.Forbidden("Akun belum terikat ke rumah sakit.")
		}
		q = q.Where("l.hospital_id = ?", *p.HospitalID)
	}
	var out []EntryDTO
	if err := q.Scan(&out).Error; err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, response.NotFound("Entri ledger tidak ditemukan.")
	}
	dto := out[0]
	var evs []Event
	if err := q.Session(&gorm.Session{NewDB: true}).Model(&Event{}).
		Where("ledger_id = ?", id).Order("waktu").Find(&evs).Error; err != nil {
		return nil, err
	}
	dto.Events = make([]EventDTO, 0, len(evs))
	for _, e := range evs {
		dto.Events = append(dto.Events, EventDTO{e.ID, e.DariStatus, e.KeStatus, e.AktorNama, e.Catatan, e.Waktu})
	}
	return &dto, nil
}

// Buat membuat kandidat manual dari satu episode.
func (s *Service) Buat(ctx context.Context, p *auth.Principal, episodeID string) (*EntryDTO, error) {
	if !types.IsUUID(episodeID) {
		return nil, response.NotFound("Episode tidak ditemukan.")
	}
	var ep struct {
		ID, HospitalID, NoSep, KelasRugi string
		Selisih                          int64
		Superseded                       bool
	}
	q := s.db.WithContext(ctx).Table("episodes").
		Select("id::text AS id, hospital_id::text AS hospital_id, no_sep, kelas_rugi::text AS kelas_rugi, selisih, superseded").
		Where("id = ?", episodeID)
	if !p.IsSuperadmin() {
		if p.HospitalID == nil {
			return nil, response.Forbidden("Akun belum terikat ke rumah sakit.")
		}
		q = q.Where("hospital_id = ?", *p.HospitalID)
	}
	if err := q.Scan(&ep).Error; err != nil {
		return nil, err
	}
	if ep.ID == "" {
		return nil, response.NotFound("Episode tidak ditemukan.")
	}
	if ep.KelasRugi == "SURPLUS" {
		return nil, response.Unprocessable("Episode surplus tidak menjadi kandidat recovery.")
	}

	strategi := scoring.StrategiUntuk(ep.KelasRugi)
	e := Entry{
		EpisodeID: ep.ID, HospitalID: ep.HospitalID, Sep: ep.NoSep, KelasRugi: ep.KelasRugi,
		Estimasi: int64(math.Abs(float64(ep.Selisih)) / 2), Status: "KANDIDAT", Strategy: &strategi,
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&e)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return response.Conflict("SEP sudah terdaftar pada ledger.")
		}
		s.audit.Catat(tx, p.Email, "LEDGER_CREATE", "LEDGER", &e.ID, nil)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Detail(ctx, p, e.ID)
}

type BodyTransisi struct {
	Ke          string `json:"ke"`
	Alasan      string `json:"alasan"`
	ReferensiRm string `json:"referensiRm"`
	Realisasi   *int64 `json:"realisasi"`
}

func (s *Service) Transisi(ctx context.Context, p *auth.Principal, id string, b BodyTransisi) (*EntryDTO, error) {
	if !types.IsUUID(id) {
		return nil, response.NotFound("Entri ledger tidak ditemukan.")
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var e Entry
		q := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id)
		if !p.IsSuperadmin() {
			if p.HospitalID == nil {
				return response.Forbidden("Akun belum terikat ke rumah sakit.")
			}
			q = q.Where("hospital_id = ?", *p.HospitalID)
		}
		if err := q.First(&e).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("Entri ledger tidak ditemukan.")
			}
			return err
		}
		hasil := CekTransisi(PermintaanTransisi{
			Dari: e.Status, Ke: b.Ke, Peran: p.Role,
			ReferensiRm: b.ReferensiRm, Alasan: b.Alasan, Realisasi: b.Realisasi,
		})
		if !hasil.OK {
			return response.WithData(422, hasil.Alasan, hasil)
		}

		dari, ke := e.Status, b.Ke
		upd := map[string]any{"status": ke, "updated_at": gorm.Expr("now()")}
		if b.Realisasi != nil {
			upd["realisasi"] = *b.Realisasi
		}
		if r := strings.TrimSpace(b.ReferensiRm); r != "" {
			upd["referensi_rm"] = r
		}
		if err := tx.Model(&Entry{}).Where("id = ?", id).Updates(upd).Error; err != nil {
			return err
		}
		var catatan *string
		if a := strings.TrimSpace(b.Alasan); a != "" {
			catatan = &a
		}
		ev := Event{LedgerID: id, DariStatus: &dari, KeStatus: &ke, AktorNama: p.Nama, Catatan: catatan}
		if err := tx.Create(&ev).Error; err != nil {
			return err
		}
		s.audit.Catat(tx, p.Email, "LEDGER_TRANSISI", "LEDGER", &id, audit.Str(dari+" → "+ke))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Detail(ctx, p, id)
}

// Kandidat = masukan pembuatan entri otomatis saat upload.
type Kandidat struct {
	EpisodeID, HospitalID, Sep, KelasRugi string
	Estimasi                              int64
}

// BuatKandidatTx menyimpan kandidat memakai transaksi pemanggil. SEP yang sudah
// punya entri dilewati (indeks unik hospital+sep). Mengembalikan jumlah baru.
func BuatKandidatTx(tx *gorm.DB, ks []Kandidat) (int, error) {
	if len(ks) == 0 {
		return 0, nil
	}
	rows := make([]Entry, len(ks))
	for i, k := range ks {
		st := scoring.StrategiUntuk(k.KelasRugi)
		rows[i] = Entry{
			EpisodeID: k.EpisodeID, HospitalID: k.HospitalID, Sep: k.Sep, KelasRugi: k.KelasRugi,
			Estimasi: k.Estimasi, Status: "KANDIDAT", Strategy: &st,
		}
	}
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(rows, 200)
	return int(res.RowsAffected), res.Error
}
