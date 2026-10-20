package ledger

// Kalibrasi = akurasi estimasi vs realisasi pada entri yang sudah selesai.
type Kalibrasi struct {
	TotalEstimasi     int64   `json:"totalEstimasi"`
	TotalRealisasi    int64   `json:"totalRealisasi"`
	RealizationRate   float64 `json:"realizationRate"`
	HitRate           float64 `json:"hitRate"`
	NSampel           int     `json:"nSampel"`
	KalibrasiTersedia bool    `json:"kalibrasiTersedia"`
	RealisasiNegatif  int     `json:"realisasiNegatif"`
	// EstimasiAktif = total estimasi entri yang belum selesai (pipeline recovery berjalan).
	EstimasiAktif int64 `json:"estimasiAktif"`
	NAktif        int   `json:"nAktif"`
}

// Kalibrasi baru dianggap bermakna setelah minimal 30 sampel.
const minSampelKalibrasi = 30

func HitungKalibrasi(entri []Entry) Kalibrasi {
	var selesai, berubah []Entry
	var estAktif int64
	nAktif := 0
	for _, e := range entri {
		if e.Status != "HASIL" && e.Status != "TIDAK_BERUBAH" {
			estAktif += e.Estimasi
			nAktif++
		}
		if e.Realisasi != nil && (e.Status == "HASIL" || e.Status == "TIDAK_BERUBAH") {
			selesai = append(selesai, e)
			if e.Status == "HASIL" {
				berubah = append(berubah, e)
			}
		}
	}
	var totalEst, totalReal int64
	for _, e := range berubah {
		totalEst += e.Estimasi
		totalReal += *e.Realisasi
	}
	neg := 0
	for _, e := range selesai {
		if *e.Realisasi < 0 {
			neg++
		}
	}
	k := Kalibrasi{
		TotalEstimasi: totalEst, TotalRealisasi: totalReal, NSampel: len(selesai),
		KalibrasiTersedia: len(selesai) >= minSampelKalibrasi, RealisasiNegatif: neg,
		EstimasiAktif: estAktif, NAktif: nAktif,
	}
	if totalEst > 0 {
		k.RealizationRate = float64(totalReal) / float64(totalEst)
	}
	if len(selesai) > 0 {
		k.HitRate = float64(len(berubah)) / float64(len(selesai))
	}
	return k
}
