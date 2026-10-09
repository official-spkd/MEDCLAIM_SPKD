// ============================================================
// MedClaim Go — Logika bisnis inti (port dari analytics.ts):
// rumus finansial, klasifikasi rugi S1/S2/Structural,
// readiness, deteksi anomali, kalibrasi ledger.
// ============================================================
package main

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var KOMPONEN_BILLING = []string{
	"VISITATION_FEE", "ACCOMODATION_FEE", "ICU_FEE", "ICU_SUPPORT_FEE",
	"NURSE_FEE", "TREATMENT_FEE", "DRUG_FEE", "DRUG_KRONIS_FEE",
	"DRUG_KEMO_FEE", "OBAT_KRONIS", "OBAT_KEMO", "ALKES_BHP",
	"LAB", "RAD", "REHAB_MEDIK", "KAMAR_AKOMODASI", "OPERASIONAL", "DARAH",
}

// hitungSelisih = tarif − billing (positif = surplus, negatif = rugi).
func hitungSelisih(tarifInacbg, totalBilling int64) int64 {
	return tarifInacbg - totalBilling
}

func hitungPctRugi(tarifInacbg, totalBilling int64) float64 {
	if tarifInacbg <= 0 {
		return 0
	}
	sel := hitungSelisih(tarifInacbg, totalBilling)
	if sel >= 0 {
		return 0
	}
	return math.Min(math.Abs(float64(sel))/float64(tarifInacbg), 1)
}

func klasifikasiRugi(tarifInacbg, totalBilling int64) string {
	pct := hitungPctRugi(tarifInacbg, totalBilling)
	if pct == 0 {
		return "SURPLUS"
	}
	if pct < 0.4 {
		return "S1_CODING"
	}
	if pct < 0.7 {
		return "S2_MIXED"
	}
	return "STRUCTURAL"
}

// basisGrupCbg: buang digit severity terakhir (Q-5-14-0 → Q-5-14).
func basisGrupCbg(kodeIcbg string) string {
	parts := strings.Split(kodeIcbg, "-")
	if len(parts) == 4 {
		return strings.Join(parts[:3], "-")
	}
	return kodeIcbg
}

func median(nilai []float64) float64 {
	if len(nilai) == 0 {
		return 0
	}
	s := append([]float64(nil), nilai...)
	sort.Float64s(s)
	mid := len(s) / 2
	if len(s)%2 == 0 {
		return (s[mid-1] + s[mid]) / 2
	}
	return s[mid]
}

// ---------- Readiness ----------

var BobotReadinessDefault = map[string]int{
	"severity": 30, "drug": 25, "singleDiag": 20, "icd": 15, "los": 10,
}

var reICD10 = regexp.MustCompile(`^[A-Z]\d{2}(\.\d{1,2})?$`)
var reNeoplasma = regexp.MustCompile(`^C\d{2}`)
var reKronis = regexp.MustCompile(`^E1\d|^I1\d|^I2\d|^N18|^J4[4-7]`)

func icd10Valid(kode string) bool {
	return reICD10.MatchString(strings.ToUpper(kode))
}

func drugAlignment(diagnosa []string, komponen map[string]int64) (bool, []string) {
	flag := []string{}
	kemo := komponen["OBAT_KEMO"] > 0
	kronis := komponen["OBAT_KRONIS"]+komponen["DRUG_KRONIS_FEE"] > 0
	adaNeoplasma, adaKronis := false, false
	for _, d := range diagnosa {
		k := strings.ToUpper(d)
		if reNeoplasma.MatchString(k) {
			adaNeoplasma = true
		}
		if reKronis.MatchString(k) {
			adaKronis = true
		}
	}
	if kemo && !adaNeoplasma {
		flag = append(flag, "OBAT_KEMO > 0 tanpa diagnosis neoplasma (C##)")
	}
	if kronis && !adaKronis {
		flag = append(flag, "Obat kronis > 0 tanpa diagnosis E1x/I1x–I2x/N18/J44–47")
	}
	return len(flag) == 0, flag
}

type ReadinessResult struct {
	Total    int             `json:"total"`
	Komponen map[string]int  `json:"komponen"`
	Flag     []string        `json:"flag"`
}

func readinessScore(ep *Episode) ReadinessResult {
	flag := []string{}
	sevOk := ep.SeverityLevel >= 1 && ep.SeverityLevel <= 3 && len(strings.Split(ep.KodeIcbg, "-")) >= 3
	if !sevOk {
		flag = append(flag, "Severity level tidak valid / tidak tergrouping")
	}
	dx := make([]string, len(ep.Diagnosa))
	for i, d := range ep.Diagnosa {
		dx[i] = d.Kode
	}
	drugOk, drugFlag := drugAlignment(dx, ep.Komponen)
	flag = append(flag, drugFlag...)
	singleOk := len(ep.Diagnosa) > 1
	if !singleOk {
		flag = append(flag, "Single diagnosis — peluang CC/MCC terlewat")
	}
	semuaValid := true
	for _, d := range ep.Diagnosa {
		if !icd10Valid(d.Kode) {
			semuaValid = false
			break
		}
	}
	if !semuaValid {
		flag = append(flag, "Terdapat kode ICD-10 tidak valid")
	}
	hari := selisihHari(ep.TglMasuk, ep.TglKeluar)
	losOk := hari >= 0 && abs(hari-ep.Los) <= 1 && ep.Los > 0
	if !losOk {
		flag = append(flag, "LOS tidak konsisten dengan tanggal masuk/keluar")
	}
	b := BobotReadinessDefault
	komp := map[string]int{}
	if sevOk {
		komp["severity"] = b["severity"]
	}
	if drugOk {
		komp["drug"] = b["drug"]
	}
	if singleOk {
		komp["singleDiag"] = b["singleDiag"]
	}
	if semuaValid {
		komp["icd"] = b["icd"]
	}
	if losOk {
		komp["los"] = b["los"]
	}
	total := 0
	for _, v := range komp {
		total += v
	}
	return ReadinessResult{Total: total, Komponen: komp, Flag: flag}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// ---------- Deteksi anomali ----------

type AnomaliHasil struct {
	EpisodeID string `json:"episodeId"`
	Jenis     string `json:"jenis"`
	Detail    string `json:"detail"`
}

func diagnosaKodeList(ep *Episode) []string {
	out := make([]string, len(ep.Diagnosa))
	for i, d := range ep.Diagnosa {
		out[i] = strings.ToUpper(d.Kode)
	}
	return out
}

func diagnosaTerkait(a, b []string) bool {
	setB := map[string]bool{}
	for _, x := range b {
		setB[x] = true
	}
	kategoriB := map[string]bool{}
	for _, x := range b {
		if len(x) >= 3 {
			kategoriB[x[:3]] = true
		}
	}
	for _, x := range a {
		if setB[x] {
			return true
		}
		if len(x) >= 3 && kategoriB[x[:3]] {
			return true
		}
	}
	return false
}

func deteksiAnomali(episodes []*Episode) []AnomaliHasil {
	hasil := []AnomaliHasil{}
	aktif := []*Episode{}
	for _, e := range episodes {
		if !e.Superseded {
			aktif = append(aktif, e)
		}
	}

	// LOS outlier per grup basis CBG
	perGrup := map[string][]float64{}
	for _, ep := range aktif {
		basis := basisGrupCbg(ep.KodeIcbg)
		perGrup[basis] = append(perGrup[basis], float64(ep.Los))
	}
	for _, ep := range aktif {
		basis := basisGrupCbg(ep.KodeIcbg)
		med := median(perGrup[basis])
		if med > 0 && float64(ep.Los) > 2.5*med {
			hasil = append(hasil, AnomaliHasil{
				EpisodeID: ep.ID, Jenis: "LOS_OUTLIER",
				Detail: fmt.Sprintf("LOS %d hari > 2,5× median grup %s (%.1f hari)", ep.Los, basis, med),
			})
		}
		if ep.TotalBilling == 0 {
			hasil = append(hasil, AnomaliHasil{
				EpisodeID: ep.ID, Jenis: "SELISIH_EKSTREM",
				Detail: "Total billing = 0 — komponen billing tidak terisi",
			})
		} else {
			dominan := int64(0)
			for _, v := range ep.Komponen {
				if v > dominan {
					dominan = v
				}
			}
			if float64(dominan)/float64(ep.TotalBilling) > 0.8 {
				hasil = append(hasil, AnomaliHasil{
					EpisodeID: ep.ID, Jenis: "SELISIH_EKSTREM",
					Detail: "Satu komponen billing > 80% dari total — perlu verifikasi rincian",
				})
			}
		}
	}

	// Duplikat SEP
	bySep := map[string][]*Episode{}
	var urutanSep []string
	for _, ep := range aktif {
		if _, ok := bySep[ep.NoSep]; !ok {
			urutanSep = append(urutanSep, ep.NoSep)
		}
		bySep[ep.NoSep] = append(bySep[ep.NoSep], ep)
	}
	for _, sep := range urutanSep {
		list := bySep[sep]
		if len(list) > 1 {
			hasil = append(hasil, AnomaliHasil{
				EpisodeID: list[0].ID, Jenis: "DUPLIKAT",
				Detail: fmt.Sprintf("SEP %s muncul %d× pada data aktif", sep, len(list)),
			})
		}
	}

	// Readmisi ≤ 30 hari
	byMr := map[string][]*Episode{}
	var urutanMr []string
	for _, ep := range aktif {
		if _, ok := byMr[ep.NoMr]; !ok {
			urutanMr = append(urutanMr, ep.NoMr)
		}
		byMr[ep.NoMr] = append(byMr[ep.NoMr], ep)
	}
	for _, mr := range urutanMr {
		list := append([]*Episode(nil), byMr[mr]...)
		sort.Slice(list, func(i, j int) bool { return list[i].TglMasuk < list[j].TglMasuk })
		for i := 1; i < len(list); i++ {
			prev, cur := list[i-1], list[i]
			hari := selisihHari(prev.TglKeluar, cur.TglMasuk)
			if hari >= 0 && hari <= 30 && diagnosaTerkait(diagnosaKodeList(prev), diagnosaKodeList(cur)) {
				hasil = append(hasil, AnomaliHasil{
					EpisodeID: cur.ID, Jenis: "READMISI",
					Detail: fmt.Sprintf("Masuk ulang %d hari setelah pulang (diagnosis sama/terkait, MR %s)", hari, cur.NoMr),
				})
			}
		}
	}
	return hasil
}

// ---------- Estimasi upgrade tarif ----------

func estimasiUpgradeTarif(episodes []*Episode) map[string]int64 {
	grup := map[string]map[int][2]float64{} // basis → severity → {total, n}
	for _, ep := range episodes {
		if ep.SeverityLevel < 1 {
			continue
		}
		basis := basisGrupCbg(ep.KodeIcbg)
		if grup[basis] == nil {
			grup[basis] = map[int][2]float64{}
		}
		cur := grup[basis][ep.SeverityLevel]
		cur[0] += float64(ep.TarifInacbg)
		cur[1]++
		grup[basis][ep.SeverityLevel] = cur
	}
	estimasi := map[string]int64{}
	for _, ep := range episodes {
		if ep.KelasRugi == "SURPLUS" || ep.SeverityLevel < 1 {
			estimasi[ep.ID] = 0
			continue
		}
		basis := basisGrupCbg(ep.KodeIcbg)
		next, ok := grup[basis][ep.SeverityLevel+1]
		if !ok || next[1] == 0 {
			estimasi[ep.ID] = 0
			continue
		}
		tarifRataNext := next[0] / next[1]
		delta := tarifRataNext - float64(ep.TarifInacbg)
		rugiSaatIni := math.Abs(float64(ep.Selisih))
		estimasi[ep.ID] = int64(math.Max(0, math.Min(delta, rugiSaatIni)))
	}
	return estimasi
}

// ---------- Kalibrasi ledger ----------

type KalibrasiLedger struct {
	TotalEstimasi     int64   `json:"totalEstimasi"`
	TotalRealisasi    int64   `json:"totalRealisasi"`
	RealizationRate   float64 `json:"realizationRate"`
	HitRate           float64 `json:"hitRate"`
	NSampel           int     `json:"nSampel"`
	KalibrasiTersedia bool    `json:"kalibrasiTersedia"`
	RealisasiNegatif  int     `json:"realisasiNegatif"`
}

func kalibrasiLedger(entri []*LedgerEntry) KalibrasiLedger {
	denganRealisasi := []*LedgerEntry{}
	for _, e := range entri {
		if e.Realisasi != nil && (e.Status == "HASIL" || e.Status == "TIDAK_BERUBAH") {
			denganRealisasi = append(denganRealisasi, e)
		}
	}
	berubah := []*LedgerEntry{}
	for _, e := range denganRealisasi {
		if e.Status == "HASIL" {
			berubah = append(berubah, e)
		}
	}
	var totalEst, totalReal int64
	for _, e := range berubah {
		totalEst += e.Estimasi
		totalReal += *e.Realisasi
	}
	penyebut := len(denganRealisasi)
	realRate := 0.0
	if totalEst > 0 {
		realRate = float64(totalReal) / float64(totalEst)
	}
	hitRate := 0.0
	if penyebut > 0 {
		hitRate = float64(len(berubah)) / float64(penyebut)
	}
	neg := 0
	for _, e := range denganRealisasi {
		if e.Realisasi != nil && *e.Realisasi < 0 {
			neg++
		}
	}
	return KalibrasiLedger{
		TotalEstimasi: totalEst, TotalRealisasi: totalReal,
		RealizationRate: realRate, HitRate: hitRate, NSampel: penyebut,
		KalibrasiTersedia: penyebut >= 30, RealisasiNegatif: neg,
	}
}

// ---------- Format ----------

func fmtInt64(n int64) string {
	s := strconv.FormatInt(n, 10)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	out := strings.Join(parts, ".")
	if neg {
		out = "-" + out
	}
	return out
}
