// Package analytics: dashboard, anomali, dan ringkasan finansial. Semua angka
// dihitung dari episode aktif dalam cakupan hak akses pemanggil.
package analytics

import (
	"context"
	"sort"
	"time"

	"gorm.io/gorm"

	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/claims"
	"medclaim-backend/internal/response"
	"medclaim-backend/internal/scoring"
)

type Service struct {
	db     *gorm.DB
	claims *claims.Service
}

func NewService(db *gorm.DB, c *claims.Service) *Service { return &Service{db: db, claims: c} }

type KPI struct {
	TotalEpisode    int     `json:"totalEpisode"`
	TotalTarif      int64   `json:"totalTarif"`
	TotalBilling    int64   `json:"totalBilling"`
	TotalSelisih    int64   `json:"totalSelisih"`
	PctRugiRata     float64 `json:"pctRugiRata"`
	PotensiRecovery int64   `json:"potensiRecovery"`
	EstimasiRugi    int64   `json:"estimasiRugi"`
	ReadinessRata   int     `json:"readinessRata"`
	DQRata          int     `json:"dqRata"`
	LosRata         float64 `json:"losRata"`
	JumlahUpload    int     `json:"jumlahUpload"`
	LedgerAktif     int     `json:"ledgerAktif"`
	QueryTerbuka    int     `json:"queryTerbuka"`
	QueryTerlambat  int     `json:"queryTerlambat"`
	AnomaliJumlah   int     `json:"anomaliJumlah"`
}

type cbgRow struct {
	Kode string `json:"kode"`
	Desk string `json:"desk"`
	N    int    `json:"jumlah"`
	Rugi int    `json:"rugi"`
}

func (s *Service) Dashboard(ctx context.Context, p *auth.Principal, f claims.Filter) (map[string]any, error) {
	eps, err := s.claims.LoadActive(ctx, p, f)
	if err != nil {
		return nil, err
	}
	rs, err := p.ScopeHospital(f.Rs)
	if err != nil {
		return nil, err
	}

	sc := make([]*scoring.Episode, len(eps))
	desk := map[string]string{}
	for i := range eps {
		sc[i] = claims.ToScoring(&eps[i])
		d := eps[i].KodeIcbg
		if eps[i].Icbg != nil && eps[i].Icbg.Deskripsi != "" {
			d = eps[i].Icbg.Deskripsi
		}
		desk[eps[i].KodeIcbg] = d
	}

	var totalTarif, totalBilling, estimasiRugi int64
	nRugi, sumLos, readySum := 0, 0, 0
	sumPct := 0.0
	dist := map[string]int{"SURPLUS": 0, "S1_CODING": 0, "S2_MIXED": 0, "STRUCTURAL": 0}
	kelas := map[int]int{1: 0, 2: 0, 3: 0}
	cbg := map[string]*cbgRow{}
	readyKomp := map[string]int{}
	var rugiEps []*scoring.Episode

	for _, e := range sc {
		totalTarif += e.TarifInacbg
		totalBilling += e.TotalBilling
		if e.KelasRugi != "SURPLUS" {
			nRugi++
			sumPct += scoring.HitungPctRugi(e.TarifInacbg, e.TotalBilling)
			rugiEps = append(rugiEps, e)
		}
		dist[e.KelasRugi]++
		sumLos += e.Los
		if e.Selisih < 0 {
			estimasiRugi += -e.Selisih
		}
		c := cbg[e.KodeIcbg]
		if c == nil {
			c = &cbgRow{Kode: e.KodeIcbg, Desk: desk[e.KodeIcbg]}
			cbg[e.KodeIcbg] = c
		}
		c.N++
		if e.KelasRugi != "SURPLUS" && e.Selisih < 0 {
			c.Rugi++
		}
		rd := scoring.Readiness(e)
		readySum += rd.Total
		for k, v := range rd.Komponen {
			readyKomp[k] += v
		}
	}
	for i := range eps {
		if k := eps[i].KelasRawat; k >= 1 && k <= 3 {
			kelas[k]++
		}
	}

	est := scoring.EstimasiUpgradeTarif(sc)
	var potensi int64
	for _, e := range rugiEps {
		potensi += est[e.ID]
	}

	// Tren per bulan masuk.
	type bulan struct {
		n    int
		rugi int64
	}
	perBulan := map[string]*bulan{}
	for _, e := range sc {
		ym := e.TglMasuk[:7]
		if perBulan[ym] == nil {
			perBulan[ym] = &bulan{}
		}
		perBulan[ym].n++
		if e.Selisih < 0 {
			perBulan[ym].rugi += -e.Selisih
		}
	}
	yms := make([]string, 0, len(perBulan))
	for ym := range perBulan {
		yms = append(yms, ym)
	}
	sort.Strings(yms)
	tren := make([]map[string]any, 0, len(yms))
	for _, ym := range yms {
		tren = append(tren, map[string]any{"periode": ym, "episode": perBulan[ym].n, "rugi": perBulan[ym].rugi})
	}

	rows := make([]cbgRow, 0, len(cbg))
	for _, c := range cbg {
		rows = append(rows, *c)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].N != rows[j].N {
			return rows[i].N > rows[j].N
		}
		return rows[i].Kode < rows[j].Kode
	})
	if len(rows) > 8 {
		rows = rows[:8]
	}

	n := len(sc)
	kpi := KPI{
		TotalEpisode: n, TotalTarif: totalTarif, TotalBilling: totalBilling,
		TotalSelisih: totalTarif - totalBilling, PotensiRecovery: potensi,
		EstimasiRugi: estimasiRugi, AnomaliJumlah: len(scoring.DeteksiAnomali(sc)),
	}
	readyRata := 0
	kompRata := map[string]int{}
	if n > 0 {
		readyRata = readySum / n
		kpi.LosRata = float64(sumLos) / float64(n)
		for k, v := range readyKomp {
			kompRata[k] = v / n
		}
	}
	if nRugi > 0 {
		kpi.PctRugiRata = sumPct / float64(nRugi)
	}
	kpi.ReadinessRata = readyRata

	if err := s.isiKPIOperasional(ctx, p, rs, &kpi); err != nil {
		return nil, err
	}

	out := map[string]any{
		"kpi":       kpi,
		"readiness": map[string]any{"total": readyRata, "komponen": kompRata},
	}
	if !auth.BolehLihatFinansial(p.Role) {
		return out, nil // DPJP: hanya KPI & readiness
	}
	out["distribusiRugi"] = dist
	out["tren"] = tren
	out["topCbg"] = rows
	out["kelasRawat"] = []map[string]any{
		{"kelas": "1", "jumlah": kelas[1]},
		{"kelas": "2", "jumlah": kelas[2]},
		{"kelas": "3", "jumlah": kelas[3]},
	}
	return out, nil
}

// isiKPIOperasional: jumlah upload, rata-rata DQ, ledger aktif, dan query.
func (s *Service) isiKPIOperasional(ctx context.Context, p *auth.Principal, rs string, k *KPI) error {
	db := s.db.WithContext(ctx)

	var up struct {
		N    int
		Rata float64
	}
	q := db.Table("upload_batches").Select("count(*) AS n, coalesce(avg(dq_score), 0) AS rata").
		Where("status = 'SELESAI'")
	if rs != "" {
		q = q.Where("hospital_id = ?", rs)
	}
	if err := q.Scan(&up).Error; err != nil {
		return err
	}
	k.JumlahUpload, k.DQRata = up.N, int(up.Rata)

	lq := db.Table("ledger_entries").Where("status NOT IN ('HASIL', 'TIDAK_BERUBAH')")
	if rs != "" {
		lq = lq.Where("hospital_id = ?", rs)
	}
	var ledgerAktif int64
	if err := lq.Count(&ledgerAktif).Error; err != nil {
		return err
	}
	k.LedgerAktif = int(ledgerAktif)

	qq := db.Table("document_queries").Where("status = 'TERKIRIM'")
	if rs != "" {
		qq = qq.Where("hospital_id = ?", rs)
	}
	if p.Role == "DPJP" {
		qq = qq.Where("dpjp_id = ?", p.ID)
	}
	var deadline []time.Time
	if err := qq.Pluck("deadline", &deadline).Error; err != nil {
		return err
	}
	now := time.Now()
	for _, d := range deadline {
		if d.Before(now) {
			k.QueryTerlambat++
		} else {
			k.QueryTerbuka++
		}
	}
	return nil
}

type AnomaliItem struct {
	EpisodeID string `json:"episodeId"`
	Jenis     string `json:"jenis"`
	Detail    string `json:"detail"`
	Sep       string `json:"sep"`
	Mr        string `json:"mr"`
	Nama      string `json:"nama"`
	Icbg      string `json:"icbg"`
	Los       int    `json:"los"`
	TglMasuk  string `json:"tglMasuk"`
}

func (s *Service) Anomali(ctx context.Context, p *auth.Principal, f claims.Filter) (map[string]any, error) {
	eps, err := s.claims.LoadActive(ctx, p, f)
	if err != nil {
		return nil, err
	}
	sc := make([]*scoring.Episode, len(eps))
	byID := map[string]*claims.Episode{}
	for i := range eps {
		sc[i] = claims.ToScoring(&eps[i])
		byID[eps[i].ID] = &eps[i]
	}
	items := []AnomaliItem{}
	for _, a := range scoring.DeteksiAnomali(sc) {
		e := byID[a.EpisodeID]
		if e == nil {
			continue
		}
		items = append(items, AnomaliItem{
			EpisodeID: a.EpisodeID, Jenis: a.Jenis, Detail: a.Detail, Sep: e.NoSep, Mr: e.NoMr,
			Nama: e.NamaPasien, Icbg: e.KodeIcbg, Los: e.Los, TglMasuk: e.TglMasuk.String(),
		})
	}
	return map[string]any{"total": len(items), "items": items}, nil
}

func (s *Service) Financial(ctx context.Context, p *auth.Principal, f claims.Filter) (map[string]any, error) {
	if !auth.BolehLihatFinansial(p.Role) {
		return nil, response.Forbidden("Peran DPJP tidak memiliki akses ke modul finansial.")
	}
	eps, err := s.claims.LoadActive(ctx, p, f)
	if err != nil {
		return nil, err
	}
	tot := map[string]int64{}
	for i := range eps {
		for _, c := range eps[i].Components {
			tot[c.Komponen] += c.Nilai
		}
	}
	var total int64
	for _, k := range scoring.KomponenBilling {
		total += tot[k]
	}
	type row struct {
		Komponen string  `json:"komponen"`
		Nilai    int64   `json:"nilai"`
		Pct      float64 `json:"pct"`
	}
	rows := make([]row, 0, len(scoring.KomponenBilling))
	for _, k := range scoring.KomponenBilling {
		r := row{Komponen: k, Nilai: tot[k]}
		if total > 0 {
			r.Pct = float64(tot[k]) / float64(total)
		}
		rows = append(rows, r)
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Nilai > rows[j].Nilai })
	return map[string]any{"total": total, "komponen": rows}, nil
}
