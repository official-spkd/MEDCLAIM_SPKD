package claims

import (
	"math"
	"sort"

	"medclaim-backend/internal/scoring"
)

type RefDTO struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

type AddonDTO struct {
	Kode   string `json:"kode"`
	Jenis  string `json:"jenis"`
	Tarif  int64  `json:"tarif"`
	Status string `json:"status"`
}

// EpisodeDTO = tipe Episode di frontend (types.ts).
type EpisodeDTO struct {
	ID            string           `json:"id"`
	UploadBatchID string           `json:"uploadBatchId"`
	NoSep         string           `json:"noSep"`
	NoKartu       string           `json:"noKartu"`
	NoMr          string           `json:"noMr"`
	NamaPasien    string           `json:"namaPasien"`
	TglLahir      string           `json:"tglLahir"`
	JenisKelamin  string           `json:"jenisKelamin"`
	TglMasuk      string           `json:"tglMasuk"`
	TglKeluar     string           `json:"tglKeluar"`
	Los           int              `json:"los"`
	KelasRawat    int              `json:"kelasRawat"`
	TipeLayanan   string           `json:"tipeLayanan"`
	CaraPulang    *string          `json:"caraPulang"`
	DpjpKode      *string          `json:"dpjpKode"`
	DpjpNama      *string          `json:"dpjpNama"`
	Diagnosa      []RefDTO         `json:"diagnosa"`
	Prosedur      []RefDTO         `json:"prosedur"`
	KodeIcbg      string           `json:"kodeIcbg"`
	DeskripsiIcbg string           `json:"deskripsiIcbg"`
	SeverityLevel int              `json:"severityLevel"`
	IdrgKode      *string          `json:"idrgKode"`
	Komponen      map[string]int64 `json:"komponen"`
	TotalBilling  int64            `json:"totalBilling"`
	TarifInacbg   int64            `json:"tarifInacbg"`
	Selisih       int64            `json:"selisih"`
	PctRugi       float64          `json:"pctRugi"`
	KelasRugi     string           `json:"kelasRugi"`
	AddOn         []AddonDTO       `json:"addOn"`
	DQScore       int              `json:"dqScore"`
	VersionNo     int              `json:"versionNo"`
	Superseded    bool             `json:"superseded"`
	CreatedAt     string           `json:"createdAt"`

	Readiness *scoring.ReadinessResult `json:"readiness,omitempty"`
	Anomali   []string                 `json:"anomali,omitempty"`
}

// ToDTO memetakan entitas ke DTO. Tanpa hak finansial (DPJP), seluruh angka
// uang disembunyikan dan kelas rugi dikaburkan menjadi RAHASIA.
func ToDTO(e *Episode, lihatFinansial bool) EpisodeDTO {
	d := EpisodeDTO{
		ID: e.ID, UploadBatchID: e.UploadBatchID, NoSep: e.NoSep, NoKartu: e.NoKartu,
		NoMr: e.NoMr, NamaPasien: e.NamaPasien, TglLahir: e.TglLahir.String(),
		JenisKelamin: e.JenisKelamin, TglMasuk: e.TglMasuk.String(), TglKeluar: e.TglKeluar.String(),
		Los: e.Los, KelasRawat: e.KelasRawat, TipeLayanan: e.TipeLayanan,
		CaraPulang: e.CaraPulang, DpjpKode: e.DpjpKode, DpjpNama: e.DpjpNama,
		KodeIcbg: e.KodeIcbg, DeskripsiIcbg: e.KodeIcbg, SeverityLevel: e.SeverityLevel,
		IdrgKode: e.IdrgKode, DQScore: e.DQScore, VersionNo: e.VersionNo,
		Superseded: e.Superseded, CreatedAt: e.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		Diagnosa: []RefDTO{}, Prosedur: []RefDTO{}, AddOn: []AddonDTO{}, Komponen: map[string]int64{},
	}
	if e.Icbg != nil && e.Icbg.Deskripsi != "" {
		d.DeskripsiIcbg = e.Icbg.Deskripsi
	}

	dx := append([]EpisodeDiagnosis(nil), e.Diagnoses...)
	sort.Slice(dx, func(i, j int) bool { return dx[i].Urutan < dx[j].Urutan })
	for _, x := range dx {
		d.Diagnosa = append(d.Diagnosa, RefDTO{x.KodeIcd10, x.Nama})
	}
	for _, x := range e.Procedures {
		d.Prosedur = append(d.Prosedur, RefDTO{x.KodeIcd9, x.Nama})
	}

	if !lihatFinansial {
		d.KelasRugi = "RAHASIA"
		for _, a := range e.Addons {
			d.AddOn = append(d.AddOn, AddonDTO{a.Kode, a.Jenis, 0, a.Status})
		}
		return d
	}
	for _, c := range e.Components {
		d.Komponen[c.Komponen] = c.Nilai
	}
	for _, a := range e.Addons {
		d.AddOn = append(d.AddOn, AddonDTO{a.Kode, a.Jenis, a.Tarif, a.Status})
	}
	d.TotalBilling, d.TarifInacbg, d.Selisih = e.TotalBilling, e.TarifInacbg, e.Selisih
	d.PctRugi = math.Round(e.PctRugi*100) / 10000 // persen (2 desimal) → pecahan
	d.KelasRugi = e.KelasRugi
	return d
}

// ToScoring menyiapkan masukan untuk paket scoring (angka asli, bukan DTO).
func ToScoring(e *Episode) *scoring.Episode {
	s := &scoring.Episode{
		ID: e.ID, NoSep: e.NoSep, NoMr: e.NoMr,
		TglMasuk: e.TglMasuk.String(), TglKeluar: e.TglKeluar.String(),
		Los: e.Los, SeverityLevel: e.SeverityLevel, KodeIcbg: e.KodeIcbg,
		Komponen: map[string]int64{}, TotalBilling: e.TotalBilling,
		TarifInacbg: e.TarifInacbg, Selisih: e.Selisih, KelasRugi: e.KelasRugi,
	}
	dx := append([]EpisodeDiagnosis(nil), e.Diagnoses...)
	sort.Slice(dx, func(i, j int) bool { return dx[i].Urutan < dx[j].Urutan })
	for _, x := range dx {
		s.Diagnosa = append(s.Diagnosa, x.KodeIcd10)
	}
	for _, c := range e.Components {
		s.Komponen[c.Komponen] = c.Nilai
	}
	return s
}
