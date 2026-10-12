package uploads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"medclaim-backend/internal/audit"
	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/claims"
	"medclaim-backend/internal/hospitals"
	"medclaim-backend/internal/ledger"
	"medclaim-backend/internal/response"
	"medclaim-backend/internal/scoring"
	"medclaim-backend/internal/types"
)

type Service struct {
	db        *gorm.DB
	audit     *audit.Service
	hospitals *hospitals.Service
}

func NewService(db *gorm.DB, a *audit.Service, h *hospitals.Service) *Service {
	return &Service{db: db, audit: a, hospitals: h}
}

var rePeriode = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

// cek memvalidasi permintaan dan hak pemanggil atas RS yang dituju.
func (s *Service) cek(ctx context.Context, p *auth.Principal, req *Permintaan) (*hospitals.Hospital, error) {
	if !types.IsUUID(req.HospitalID) {
		return nil, response.BadRequest("Rumah sakit tidak ditemukan.")
	}
	// Non-superadmin hanya boleh mengunggah untuk RS miliknya.
	if !p.IsSuperadmin() && (p.HospitalID == nil || *p.HospitalID != req.HospitalID) {
		return nil, response.Forbidden("Anda hanya dapat mengunggah data untuk rumah sakit Anda.")
	}
	h, err := s.hospitals.Get(ctx, req.HospitalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.BadRequest("Rumah sakit tidak ditemukan.")
		}
		return nil, err
	}
	if !rePeriode.MatchString(req.Periode) {
		return nil, response.Unprocessable("Periode harus berformat yyyy-mm.")
	}
	if strings.TrimSpace(req.Isi) == "" {
		return nil, response.Unprocessable("Isi file masih kosong.")
	}
	req.NamaFile = strings.TrimSpace(req.NamaFile)
	if req.NamaFile == "" {
		req.NamaFile = "tanpa-nama.txt"
	}
	if len(req.NamaFile) > 255 {
		req.NamaFile = req.NamaFile[:255]
	}
	return h, nil
}

func formatDari(kolom []string) string {
	for _, k := range kolom {
		if k == "VISITATION_FEE" {
			return "DETAIL"
		}
	}
	return "BASIC"
}

func (s *Service) konteksDQ(ctx context.Context, hospitalID, periode string, kolom []string) (KonteksDQ, error) {
	var total []int
	if err := s.db.WithContext(ctx).Model(&Batch{}).
		Where("hospital_id = ? AND status = 'SELESAI'", hospitalID).
		Pluck("total_baris", &total).Error; err != nil {
		return KonteksDQ{}, err
	}
	vals := make([]float64, len(total))
	for i, t := range total {
		vals[i] = float64(t)
	}
	var seps []string
	if err := s.db.WithContext(ctx).Table("episodes").
		Where("hospital_id = ? AND superseded = false", hospitalID).
		Pluck("no_sep", &seps).Error; err != nil {
		return KonteksDQ{}, err
	}
	set := make(map[string]bool, len(seps))
	for _, x := range seps {
		set[x] = true
	}
	return KonteksDQ{
		KolomTersedia: kolom, PeriodeDipilih: periode,
		MedianUploadSebelumnya: scoring.Median(vals), SepEksisting: set,
	}, nil
}

// Preview memvalidasi file tanpa menyimpan apa pun.
func (s *Service) Preview(ctx context.Context, p *auth.Principal, req Permintaan) (map[string]any, error) {
	if _, err := s.cek(ctx, p, &req); err != nil {
		return nil, err
	}
	parse := parseTSV(req.Isi)
	v := validasiBaris(parse)
	kctx, err := s.konteksDQ(ctx, req.HospitalID, req.Periode, v.KolomTersedia)
	if err != nil {
		return nil, err
	}
	report := buatLaporanDQ(v.Valid, kctx)

	gagal := v.Gagal
	if len(gagal) > 50 {
		gagal = gagal[:50]
	}
	hitung := map[string]int{}
	for _, b := range v.Valid {
		hitung[b.NoSep]++
	}
	dup := []string{}
	for sep, n := range hitung {
		if n > 1 {
			dup = append(dup, fmt.Sprintf("%s (%d×)", sep, n))
		}
	}
	sort.Strings(dup)
	header := parse.Header
	if header == nil {
		header = []string{}
	}
	return map[string]any{
		"header": header, "delimiter": parse.Delimiter, "format": formatDari(v.KolomTersedia),
		"totalBaris": len(parse.Baris), "barisValid": len(v.Valid), "barisGagal": len(v.Gagal),
		"gagalDetail": gagal, "dq": report, "duplikatSep": dup, "kolomTersedia": v.KolomTersedia,
	}, nil
}

// barisSiap = baris valid yang benar-benar bisa disimpan, dideduplikasi per SEP
// (jika SEP muncul berkali-kali dalam file, baris TERAKHIR yang dipakai).
func barisSiap(valid []BarisValid) (siap []BarisValid, tak []BarisGagal) {
	idx := map[string]int{}
	for _, b := range valid {
		if !scoring.TanggalValid(b.TglMasuk) || !scoring.TanggalValid(b.TglKeluar) {
			tak = append(tak, BarisGagal{0, "SEP " + b.NoSep + ": tanggal masuk/keluar tidak valid — tidak disimpan"})
			continue
		}
		if !scoring.TanggalValid(b.TglLahir) {
			b.TglLahir = "1900-01-01"
		}
		if i, ada := idx[b.NoSep]; ada {
			siap[i] = b
			continue
		}
		idx[b.NoSep] = len(siap)
		siap = append(siap, b)
	}
	return siap, tak
}

type episodeLama struct {
	ID         string
	NoSep      string
	RowHash    string
	VersionNo  int
	Hospital   string
	Superseded bool
}

func (s *Service) Commit(ctx context.Context, p *auth.Principal, req Permintaan) (*HasilIngest, error) {
	h, err := s.cek(ctx, p, &req)
	if err != nil {
		return nil, err
	}
	if req.Override && !auth.BolehOverrideDQ(p.Role) {
		return nil, response.Forbidden("Override gerbang DQ hanya oleh Casemix/Super Admin.")
	}
	if req.Override && len([]rune(strings.TrimSpace(req.AlasanOverride))) < 10 {
		return nil, response.Unprocessable("Alasan override wajib diisi minimal 10 karakter.")
	}

	parse := parseTSV(req.Isi)
	v := validasiBaris(parse)
	kctx, err := s.konteksDQ(ctx, req.HospitalID, req.Periode, v.KolomTersedia)
	if err != nil {
		return nil, err
	}
	report := buatLaporanDQ(v.Valid, kctx)
	if report.Gerbang == "BLOKIR" && !req.Override {
		return nil, response.WithData(http.StatusConflict,
			fmt.Sprintf("Gerbang DQ: BLOKIR (skor %d, grade %s). Perbaiki data atau gunakan override oleh Casemix/Super Admin.", report.Skor, report.Grade),
			report)
	}
	if req.Override && report.Gerbang == "BLOKIR" {
		report.Gerbang = "OVERRIDE"
	}
	siap, tidakDisimpan := barisSiap(v.Valid)
	laporanJSON, err := json.Marshal(report)
	if err != nil {
		return nil, err
	}

	var hasil HasilIngest
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Satu upload per RS pada satu waktu: menjaga versioning tetap konsisten.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", req.HospitalID).Error; err != nil {
			return err
		}

		aktif, err := muatAktif(tx, req.HospitalID, siap)
		if err != nil {
			return err
		}
		namaDPJP, err := petaDPJP(tx, req.HospitalID)
		if err != nil {
			return err
		}

		type item struct {
			b       BarisValid
			hash    string
			version int
		}
		var tulis []item
		var supersede []string
		var diubah []string // ID versi lama yang berubah (untuk diff)
		var barisBaru, barisBerubah, tak int
		for _, b := range siap {
			hash := rowHash(b, h.Kode)
			lama, ada := aktif[b.NoSep]
			switch {
			case ada && lama.RowHash == hash:
				tak++
			case ada:
				supersede = append(supersede, lama.ID)
				diubah = append(diubah, lama.ID)
				tulis = append(tulis, item{b, hash, lama.VersionNo + 1})
				barisBerubah++
			default:
				tulis = append(tulis, item{b, hash, 1})
				barisBaru++
			}
		}

		diff, err := hitungDiff(tx, diubah, aktif, siap)
		if err != nil {
			return err
		}

		batch := Batch{
			HospitalID: req.HospitalID, UploadedByID: p.ID, NamaFile: req.NamaFile,
			Periode: req.Periode, Format: formatDari(v.KolomTersedia),
			DQScore: report.Skor, DQGrade: report.Grade, DQReport: laporanJSON,
			Status: "SELESAI", TotalBaris: len(parse.Baris), BarisValid: len(v.Valid),
			BarisGagal: len(v.Gagal) + len(tidakDisimpan), BarisBaru: barisBaru,
			BarisBerubah: barisBerubah, BarisUnchanged: tak, OverrideUsed: req.Override,
		}
		if a := strings.TrimSpace(req.AlasanOverride); req.Override && a != "" {
			batch.AlasanOverride = &a
		}
		if err := tx.Create(&batch).Error; err != nil {
			return err
		}

		// Versi lama harus ditandai superseded SEBELUM versi baru dimasukkan
		// (indeks unik parsial: satu versi aktif per SEP).
		if len(supersede) > 0 {
			if err := tx.Model(&claims.Episode{}).Where("id IN ?", supersede).
				Update("superseded", true).Error; err != nil {
				return err
			}
		}
		if err := upsertRefInacbg(tx, siap); err != nil {
			return err
		}

		eps := make([]claims.Episode, 0, len(tulis))
		for _, it := range tulis {
			eps = append(eps, bangunEpisode(it.b, it.hash, it.version, req.HospitalID, batch.ID, report.Skor, namaDPJP))
		}
		if len(eps) > 0 {
			if err := tx.Omit("Icbg").CreateInBatches(&eps, 100).Error; err != nil {
				return err
			}
		}

		ledgerBaru, err := buatKandidat(tx, req.HospitalID, eps)
		if err != nil {
			return err
		}

		meta := fmt.Sprintf("baru=%d berubah=%d unchanged=%d", barisBaru, barisBerubah, tak)
		s.audit.Catat(tx, p.Email, "UPLOAD_COMMIT", "UPLOAD", &batch.ID, &meta)

		hasil = HasilIngest{
			Upload: &batch, BarisBaru: barisBaru, BarisBerubah: barisBerubah,
			BarisUnchanged: tak, LedgerBaru: ledgerBaru, Diff: diff,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &hasil, nil
}

// muatAktif memuat versi aktif untuk SEP-SEP dalam file (per potongan 1000).
func muatAktif(tx *gorm.DB, hospitalID string, baris []BarisValid) (map[string]episodeLama, error) {
	out := map[string]episodeLama{}
	seps := make([]string, len(baris))
	for i, b := range baris {
		seps[i] = b.NoSep
	}
	for i := 0; i < len(seps); i += 1000 {
		j := min(i+1000, len(seps))
		var rows []episodeLama
		if err := tx.Table("episodes").
			Select("id::text AS id, no_sep, row_hash, version_no").
			Where("hospital_id = ? AND superseded = false AND no_sep IN ?", hospitalID, seps[i:j]).
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			out[r.NoSep] = r
		}
	}
	return out, nil
}

// petaDPJP: kode DPJP → nama, dipakai bila file tidak memuat kolom DPJPNAMA.
func petaDPJP(tx *gorm.DB, hospitalID string) (map[string]string, error) {
	var rows []struct{ DpjpKode, Nama string }
	if err := tx.Table("users").Select("dpjp_kode, nama").
		Where("hospital_id = ? AND role = 'DPJP' AND dpjp_kode IS NOT NULL", hospitalID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[string]string, len(rows))
	for _, r := range rows {
		m[r.DpjpKode] = r.Nama
	}
	return m, nil
}

func upsertRefInacbg(tx *gorm.DB, baris []BarisValid) error {
	seen := map[string]claims.RefInacbg{}
	for _, b := range baris {
		if _, ada := seen[b.KodeIcbg]; !ada {
			seen[b.KodeIcbg] = claims.RefInacbg{Kode: b.KodeIcbg, Deskripsi: b.DeskripsiIcbg, Tarif: b.TarifInacbg}
		}
	}
	for _, r := range seen {
		// Deskripsi diperbarui hanya bila file memuat deskripsi sungguhan (bukan sekadar kode).
		err := tx.Exec(`INSERT INTO ref_inacbg (kode, deskripsi, tarif) VALUES (?, ?, ?)
			ON CONFLICT (kode) DO UPDATE SET deskripsi = CASE
				WHEN EXCLUDED.deskripsi <> EXCLUDED.kode THEN EXCLUDED.deskripsi
				ELSE ref_inacbg.deskripsi END`, r.Kode, r.Deskripsi, r.Tarif).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func bangunEpisode(b BarisValid, hash string, version int, hospitalID, batchID string, dq int, namaDPJP map[string]string) claims.Episode {
	tgl := func(s string) types.Date { d, _ := types.ParseDate(s); return d }
	strPtr := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	dpjpNama := b.DpjpNama
	if dpjpNama == "" && b.DpjpKode != "" {
		dpjpNama = namaDPJP[b.DpjpKode]
	}
	selisih := scoring.HitungSelisih(b.TarifInacbg, b.TotalBilling)
	e := claims.Episode{
		HospitalID: hospitalID, UploadBatchID: batchID, KodeIcbg: b.KodeIcbg,
		NoSep: b.NoSep, NoKartu: b.NoKartu, NoMr: b.NoMr, NamaPasien: b.NamaPasien,
		TglLahir: tgl(b.TglLahir), JenisKelamin: b.JenisKelamin,
		TglMasuk: tgl(b.TglMasuk), TglKeluar: tgl(b.TglKeluar),
		Los: b.Los, KelasRawat: b.KelasRawat, TipeLayanan: b.TipeLayanan,
		CaraPulang: strPtr(b.CaraPulang), DpjpKode: strPtr(b.DpjpKode), DpjpNama: strPtr(dpjpNama),
		SeverityLevel: b.SeverityLevel, IdrgKode: strPtr(b.IdrgKode),
		TotalBilling: b.TotalBilling, TarifInacbg: b.TarifInacbg, Selisih: selisih,
		PctRugi:   math.Round(scoring.HitungPctRugi(b.TarifInacbg, b.TotalBilling)*10000) / 100,
		KelasRugi: scoring.KlasifikasiRugi(b.TarifInacbg, b.TotalBilling),
		DQScore:   dq, RowHash: hash, VersionNo: version,
	}
	for i, d := range b.Diagnosa {
		e.Diagnoses = append(e.Diagnoses, claims.EpisodeDiagnosis{Urutan: i + 1, KodeIcd10: d, Nama: claims.NamaDiagnosa(d)})
	}
	for _, pr := range b.Prosedur {
		e.Procedures = append(e.Procedures, claims.EpisodeProcedure{KodeIcd9: pr, Nama: claims.NamaProsedur(pr)})
	}
	for _, k := range scoring.KomponenBilling {
		e.Components = append(e.Components, claims.EpisodeBillingComponent{Komponen: k, Nilai: b.Komponen[k]})
	}
	for _, a := range b.AddOn {
		e.Addons = append(e.Addons, claims.EpisodeAddon{Kode: a.Kode, Jenis: a.Jenis, Tarif: a.Tarif, Status: a.Status})
	}
	return e
}

// buatKandidat membuat entri ledger untuk episode rugi yang SEP-nya belum
// punya entri (indeks unik hospital+sep menjamin tidak ada duplikat).
func buatKandidat(tx *gorm.DB, hospitalID string, baru []claims.Episode) (int, error) {
	if len(baru) == 0 {
		return 0, nil
	}
	// Estimasi upgrade tarif dihitung atas seluruh episode aktif RS ini.
	var semua []claims.Episode
	if err := tx.Select("id", "kode_icbg", "severity_level", "tarif_inacbg", "kelas_rugi", "selisih").
		Where("hospital_id = ? AND superseded = false", hospitalID).Find(&semua).Error; err != nil {
		return 0, err
	}
	sc := make([]*scoring.Episode, len(semua))
	for i := range semua {
		e := &semua[i]
		sc[i] = &scoring.Episode{ID: e.ID, KodeIcbg: e.KodeIcbg, SeverityLevel: e.SeverityLevel,
			TarifInacbg: e.TarifInacbg, KelasRugi: e.KelasRugi, Selisih: e.Selisih}
	}
	est := scoring.EstimasiUpgradeTarif(sc)

	var ks []ledger.Kandidat
	for _, e := range baru {
		if e.KelasRugi == "SURPLUS" {
			continue
		}
		nilai := est[e.ID]
		if nilai == 0 {
			nilai = int64(math.Abs(float64(e.Selisih)) / 2)
		}
		ks = append(ks, ledger.Kandidat{EpisodeID: e.ID, HospitalID: hospitalID, Sep: e.NoSep,
			KelasRugi: e.KelasRugi, Estimasi: nilai})
	}
	return ledger.BuatKandidatTx(tx, ks)
}

// hitungDiff membandingkan versi lama (yang berubah) dengan baris baru.
func hitungDiff(tx *gorm.DB, idLama []string, aktif map[string]episodeLama, siap []BarisValid) (map[string][]FieldDiff, error) {
	out := map[string][]FieldDiff{}
	if len(idLama) == 0 {
		return out, nil
	}
	var lama []claims.Episode
	if err := tx.Preload("Diagnoses").Preload("Procedures").Where("id IN ?", idLama).Find(&lama).Error; err != nil {
		return nil, err
	}
	baruPerSep := map[string]BarisValid{}
	for _, b := range siap {
		baruPerSep[b.NoSep] = b
	}
	for i := range lama {
		l := &lama[i]
		if d := diffBaris(l, baruPerSep[l.NoSep]); len(d) > 0 {
			out[l.NoSep] = d
		}
	}
	return out, nil
}

func samaSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x, y := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func diffBaris(lama *claims.Episode, baru BarisValid) []FieldDiff {
	var dxLama, prLama []string
	for _, d := range lama.Diagnoses {
		dxLama = append(dxLama, d.KodeIcd10)
	}
	for _, p := range lama.Procedures {
		prLama = append(prLama, p.KodeIcd9)
	}
	var d []FieldDiff
	if !samaSet(dxLama, baru.Diagnosa) {
		d = append(d, FieldDiff{"diagnosa", "Diagnosis (ICD-10)", strings.Join(dxLama, ", "), strings.Join(baru.Diagnosa, ", ")})
	}
	if !samaSet(prLama, baru.Prosedur) {
		d = append(d, FieldDiff{"prosedur", "Prosedur (ICD-9-CM)", strings.Join(prLama, ", "), strings.Join(baru.Prosedur, ", ")})
	}
	if lama.KodeIcbg != baru.KodeIcbg {
		d = append(d, FieldDiff{"kodeIcbg", "Kode INA-CBG", lama.KodeIcbg, baru.KodeIcbg})
	}
	if lama.SeverityLevel != baru.SeverityLevel {
		d = append(d, FieldDiff{"severityLevel", "Severity Level", strconv.Itoa(lama.SeverityLevel), strconv.Itoa(baru.SeverityLevel)})
	}
	if lama.TarifInacbg != baru.TarifInacbg {
		d = append(d, FieldDiff{"tarifInacbg", "Tarif INA-CBG", scoring.FormatRibuan(lama.TarifInacbg), scoring.FormatRibuan(baru.TarifInacbg)})
	}
	if lama.Los != baru.Los {
		d = append(d, FieldDiff{"los", "LOS (hari)", strconv.Itoa(lama.Los), strconv.Itoa(baru.Los)})
	}
	if lama.KelasRawat != baru.KelasRawat {
		d = append(d, FieldDiff{"kelasRawat", "Kelas Rawat", strconv.Itoa(lama.KelasRawat), strconv.Itoa(baru.KelasRawat)})
	}
	return d
}

// ---------- Daftar & siklus hidup upload ----------

func (s *Service) List(ctx context.Context, p *auth.Principal) ([]Batch, error) {
	hid, err := p.ScopeHospital("")
	if err != nil {
		return nil, err
	}
	out := []Batch{}
	q := s.db.WithContext(ctx).Where("status <> 'DIHAPUS'").Order("created_at DESC")
	if hid != "" {
		q = q.Where("hospital_id = ?", hid)
	}
	return out, q.Find(&out).Error
}

func (s *Service) ubahStatus(ctx context.Context, p *auth.Principal, id, status, aksi string) (*Batch, error) {
	if !types.IsUUID(id) {
		return nil, response.NotFound("Upload tidak ditemukan.")
	}
	var b Batch
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status <> 'DIHAPUS'", id)
		if !p.IsSuperadmin() {
			if p.HospitalID == nil {
				return response.Forbidden("Akun belum terikat ke rumah sakit.")
			}
			q = q.Where("hospital_id = ?", *p.HospitalID)
		}
		if err := q.First(&b).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("Upload tidak ditemukan.")
			}
			return err
		}
		if err := tx.Model(&Batch{}).Where("id = ?", id).Update("status", status).Error; err != nil {
			return err
		}
		b.Status = status
		s.audit.Catat(tx, p.Email, aksi, "UPLOAD", &id, nil)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Service) Arsip(ctx context.Context, p *auth.Principal, id string) (*Batch, error) {
	return s.ubahStatus(ctx, p, id, "ARSIP", "UPLOAD_ARSIP")
}

// Hapus = soft delete (status DIHAPUS); data episode tetap untuk jejak audit.
func (s *Service) Hapus(ctx context.Context, p *auth.Principal, id string) error {
	_, err := s.ubahStatus(ctx, p, id, "DIHAPUS", "UPLOAD_DELETE")
	return err
}
