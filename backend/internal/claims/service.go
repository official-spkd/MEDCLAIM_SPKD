package claims

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/response"
	"medclaim-backend/internal/scoring"
	"medclaim-backend/internal/types"
)

// Filter = query string yang dipakai semua halaman analitik.
type Filter struct {
	Rs, Upload, Tipe, Dari, Sampai, KelasRugi, Kelas, Sev, Q string
}

const batasMaks = 20000

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

func (f Filter) validasi() error {
	if f.Upload != "" && !types.IsUUID(f.Upload) {
		return response.BadRequest("Parameter upload tidak valid.")
	}
	if f.Tipe != "" && f.Tipe != "SEMUA" && f.Tipe != "RANAP" && f.Tipe != "RJTP" {
		return response.BadRequest("Parameter tipe tidak valid.")
	}
	for _, t := range []string{f.Dari, f.Sampai} {
		if t != "" && !scoring.TanggalValid(t) {
			return response.BadRequest("Format tanggal harus yyyy-mm-dd.")
		}
	}
	for _, n := range []string{f.Kelas, f.Sev} {
		if n != "" {
			if _, err := strconv.Atoi(n); err != nil {
				return response.BadRequest("Parameter numerik tidak valid.")
			}
		}
	}
	switch f.KelasRugi {
	case "", "SEMUA", "SURPLUS", "S1_CODING", "S2_MIXED", "STRUCTURAL", "RAHASIA":
	default:
		return response.BadRequest("Parameter kelasRugi tidak valid.")
	}
	return nil
}

// Scope membangun query episode aktif yang sudah dibatasi hak akses:
// tenant (RS) untuk non-superadmin, dan hanya pasien miliknya untuk DPJP.
func (s *Service) Scope(ctx context.Context, p *auth.Principal, f Filter) (*gorm.DB, error) {
	if err := f.validasi(); err != nil {
		return nil, err
	}
	rs, err := p.ScopeHospital(f.Rs)
	if err != nil {
		return nil, err
	}
	q := s.db.WithContext(ctx).Model(&Episode{}).Where("episodes.superseded = false")
	if rs != "" {
		q = q.Where("episodes.hospital_id = ?", rs)
	}
	if p.Role == "DPJP" {
		if p.DpjpKode == nil {
			return q.Where("1 = 0"), nil
		}
		q = q.Where("episodes.dpjp_kode = ?", *p.DpjpKode)
	}
	if f.Upload != "" {
		q = q.Where("episodes.upload_batch_id = ?", f.Upload)
	}
	if f.Tipe != "" && f.Tipe != "SEMUA" {
		q = q.Where("episodes.tipe_layanan = ?", f.Tipe)
	}
	if f.Dari != "" {
		q = q.Where("episodes.tgl_masuk >= ?", f.Dari)
	}
	if f.Sampai != "" {
		q = q.Where("episodes.tgl_masuk <= ?", f.Sampai)
	}
	if f.KelasRugi != "" && f.KelasRugi != "SEMUA" {
		q = q.Where("episodes.kelas_rugi = ?", f.KelasRugi)
	}
	if f.Kelas != "" {
		n, _ := strconv.Atoi(f.Kelas)
		q = q.Where("episodes.kelas_rawat = ?", n)
	}
	if f.Sev != "" {
		n, _ := strconv.Atoi(f.Sev)
		q = q.Where("episodes.severity_level = ?", n)
	}
	if k := strings.TrimSpace(f.Q); k != "" {
		like := "%" + escapeLike(strings.ToLower(k)) + "%"
		q = q.Where(`(lower(episodes.no_sep) LIKE ? OR lower(episodes.no_mr) LIKE ?
			OR lower(episodes.nama_pasien) LIKE ? OR lower(episodes.kode_icbg) LIKE ?
			OR EXISTS (SELECT 1 FROM episode_diagnoses d
			           WHERE d.episode_id = episodes.id AND lower(d.kode_icd10) LIKE ?))`,
			like, like, like, like, like)
	}
	return q, nil
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func preloadAll(q *gorm.DB) *gorm.DB {
	return q.Preload("Diagnoses").Preload("Procedures").Preload("Components").
		Preload("Addons").Preload("Icbg")
}

// LoadActive memuat episode aktif beserta relasinya (untuk analitik).
func (s *Service) LoadActive(ctx context.Context, p *auth.Principal, f Filter) ([]Episode, error) {
	q, err := s.Scope(ctx, p, f)
	if err != nil {
		return nil, err
	}
	var eps []Episode
	if err := preloadAll(q).Order("episodes.no_sep").Limit(batasMaks).Find(&eps).Error; err != nil {
		return nil, err
	}
	return eps, nil
}

type ListResult struct {
	Total int          `json:"total"`
	Items []EpisodeDTO `json:"items"`
}

func (s *Service) List(ctx context.Context, p *auth.Principal, f Filter) (*ListResult, error) {
	eps, err := s.LoadActive(ctx, p, f)
	if err != nil {
		return nil, err
	}
	lihat := auth.BolehLihatFinansial(p.Role)
	items := make([]EpisodeDTO, 0, len(eps))
	for i := range eps {
		items = append(items, ToDTO(&eps[i], lihat))
	}
	return &ListResult{Total: len(items), Items: items}, nil
}

type DetailResult struct {
	Episode EpisodeDTO   `json:"episode"`
	Versi   []EpisodeDTO `json:"versi"`
}

func (s *Service) Detail(ctx context.Context, p *auth.Principal, id string) (*DetailResult, error) {
	if !types.IsUUID(id) {
		return nil, response.NotFound("Episode tidak ditemukan.")
	}
	var ep Episode
	q := preloadAll(s.db.WithContext(ctx)).Where("episodes.id = ?", id)
	if !p.IsSuperadmin() {
		if p.HospitalID == nil {
			return nil, response.Forbidden("Akun belum terikat ke rumah sakit.")
		}
		q = q.Where("episodes.hospital_id = ?", *p.HospitalID)
	}
	if err := q.First(&ep).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFound("Episode tidak ditemukan.")
		}
		return nil, err
	}
	if p.Role == "DPJP" && (ep.DpjpKode == nil || p.DpjpKode == nil || *ep.DpjpKode != *p.DpjpKode) {
		return nil, response.Forbidden("Episode ini bukan pasien DPJP Anda.")
	}

	lihat := auth.BolehLihatFinansial(p.Role)
	sc := ToScoring(&ep)
	anom := []string{}
	for _, a := range scoring.DeteksiAnomali([]*scoring.Episode{sc}) {
		anom = append(anom, a.Detail)
	}
	ready := scoring.Readiness(sc)
	dto := ToDTO(&ep, lihat)
	dto.Readiness = &ready
	dto.Anomali = anom

	var versi []Episode
	if err := preloadAll(s.db.WithContext(ctx)).
		Where("hospital_id = ? AND no_sep = ?", ep.HospitalID, ep.NoSep).
		Find(&versi).Error; err != nil {
		return nil, err
	}
	sort.Slice(versi, func(i, j int) bool { return versi[i].VersionNo < versi[j].VersionNo })
	out := make([]EpisodeDTO, 0, len(versi))
	for i := range versi {
		out = append(out, ToDTO(&versi[i], lihat))
	}
	return &DetailResult{Episode: dto, Versi: out}, nil
}
