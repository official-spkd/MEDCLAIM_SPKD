// Package scoring berisi logika bisnis murni (tanpa database): rumus selisih
// tarif-billing, klasifikasi kelas rugi, readiness, deteksi anomali, dan
// estimasi upgrade tarif. Modul lain (claims, uploads, analytics) memanggilnya.
package scoring

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// KomponenBilling = 18 komponen biaya pada ekspor V-Claim Detail (urutan tetap).
var KomponenBilling = []string{
	"VISITATION_FEE", "ACCOMODATION_FEE", "ICU_FEE", "ICU_SUPPORT_FEE",
	"NURSE_FEE", "TREATMENT_FEE", "DRUG_FEE", "DRUG_KRONIS_FEE",
	"DRUG_KEMO_FEE", "OBAT_KRONIS", "OBAT_KEMO", "ALKES_BHP",
	"LAB", "RAD", "REHAB_MEDIK", "KAMAR_AKOMODASI", "OPERASIONAL", "DARAH",
}

// Episode adalah masukan minimal untuk scoring; dibuat oleh modul claims.
type Episode struct {
	ID            string
	NoSep         string
	NoMr          string
	TglMasuk      string // YYYY-MM-DD
	TglKeluar     string // YYYY-MM-DD
	Los           int
	SeverityLevel int
	KodeIcbg      string
	Diagnosa      []string
	Komponen      map[string]int64
	TotalBilling  int64
	TarifInacbg   int64
	Selisih       int64
	KelasRugi     string
}

// ---------- Tanggal ----------

const LayoutISO = "2006-01-02"

var reTanggal = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func TanggalValid(s string) bool {
	if !reTanggal.MatchString(s) {
		return false
	}
	_, err := time.Parse(LayoutISO, s)
	return err == nil
}

// SelisihHari = keluar − masuk dalam hari; -1 bila salah satu tanggal tak valid.
func SelisihHari(masuk, keluar string) int {
	m, err1 := time.Parse(LayoutISO, masuk)
	k, err2 := time.Parse(LayoutISO, keluar)
	if err1 != nil || err2 != nil {
		return -1
	}
	return int(math.Round(k.Sub(m).Hours() / 24))
}

func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// ---------- Rumus finansial ----------

// HitungSelisih = tarif − billing (positif = surplus, negatif = rugi).
func HitungSelisih(tarifInacbg, totalBilling int64) int64 { return tarifInacbg - totalBilling }

// HitungPctRugi mengembalikan pecahan 0..1 (0 bila surplus/tarif tak valid).
func HitungPctRugi(tarifInacbg, totalBilling int64) float64 {
	if tarifInacbg <= 0 {
		return 0
	}
	sel := HitungSelisih(tarifInacbg, totalBilling)
	if sel >= 0 {
		return 0
	}
	return math.Min(math.Abs(float64(sel))/float64(tarifInacbg), 1)
}

// KlasifikasiRugi: SURPLUS | S1_CODING (<40%) | S2_MIXED (<70%) | STRUCTURAL.
func KlasifikasiRugi(tarifInacbg, totalBilling int64) string {
	pct := HitungPctRugi(tarifInacbg, totalBilling)
	switch {
	case pct == 0:
		return "SURPLUS"
	case pct < 0.4:
		return "S1_CODING"
	case pct < 0.7:
		return "S2_MIXED"
	default:
		return "STRUCTURAL"
	}
}

// StrategiUntuk = saran tindakan recovery per kelas rugi.
func StrategiUntuk(kelasRugi string) string {
	switch kelasRugi {
	case "S1_CODING":
		return "S1 — Klarifikasi CC/MCC dengan DPJP & kaji ulang koding"
	case "S2_MIXED":
		return "S2 — Klarifikasi diagnosis sekunder + verifikasi rincian billing"
	default:
		return "S3 — Verifikasi struktur tarif & komponen billing"
	}
}

// BasisGrupCbg membuang digit severity terakhir (Q-5-14-0 → Q-5-14).
func BasisGrupCbg(kodeIcbg string) string {
	parts := strings.Split(kodeIcbg, "-")
	if len(parts) == 4 {
		return strings.Join(parts[:3], "-")
	}
	return kodeIcbg
}

func Median(nilai []float64) float64 {
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

var bobotReadiness = map[string]int{"severity": 30, "drug": 25, "singleDiag": 20, "icd": 15, "los": 10}

var (
	reICD10     = regexp.MustCompile(`^[A-Z]\d{2}(\.\d{1,2})?$`)
	reNeoplasma = regexp.MustCompile(`^C\d{2}`)
	reKronis    = regexp.MustCompile(`^E1\d|^I1\d|^I2\d|^N18|^J4[4-7]`)
)

func ICD10Valid(kode string) bool  { return reICD10.MatchString(strings.ToUpper(kode)) }
func IsNeoplasma(kode string) bool { return reNeoplasma.MatchString(strings.ToUpper(kode)) }
func IsKronis(kode string) bool    { return reKronis.MatchString(strings.ToUpper(kode)) }

// DrugAlignment: obat kemo harus ditemani diagnosis neoplasma, obat kronis
// diagnosis kronis. Mengembalikan (ok, daftar flag).
func DrugAlignment(diagnosa []string, komponen map[string]int64) (bool, []string) {
	flag := []string{}
	kemo := komponen["OBAT_KEMO"] > 0
	kronis := komponen["OBAT_KRONIS"]+komponen["DRUG_KRONIS_FEE"] > 0
	adaNeoplasma, adaKronis := false, false
	for _, d := range diagnosa {
		if IsNeoplasma(d) {
			adaNeoplasma = true
		}
		if IsKronis(d) {
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
	Total    int            `json:"total"`
	Komponen map[string]int `json:"komponen"`
	Flag     []string       `json:"flag"`
}

func Readiness(ep *Episode) ReadinessResult {
	flag := []string{}
	sevOk := ep.SeverityLevel >= 1 && ep.SeverityLevel <= 3 && len(strings.Split(ep.KodeIcbg, "-")) >= 3
	if !sevOk {
		flag = append(flag, "Severity level tidak valid / tidak tergrouping")
	}
	drugOk, drugFlag := DrugAlignment(ep.Diagnosa, ep.Komponen)
	flag = append(flag, drugFlag...)
	singleOk := len(ep.Diagnosa) > 1
	if !singleOk {
		flag = append(flag, "Single diagnosis — peluang CC/MCC terlewat")
	}
	semuaValid := true
	for _, d := range ep.Diagnosa {
		if !ICD10Valid(d) {
			semuaValid = false
			break
		}
	}
	if !semuaValid {
		flag = append(flag, "Terdapat kode ICD-10 tidak valid")
	}
	hari := SelisihHari(ep.TglMasuk, ep.TglKeluar)
	losOk := hari >= 0 && Abs(hari-ep.Los) <= 1 && ep.Los > 0
	if !losOk {
		flag = append(flag, "LOS tidak konsisten dengan tanggal masuk/keluar")
	}

	komp := map[string]int{}
	total := 0
	add := func(ok bool, k string) {
		if ok {
			komp[k] = bobotReadiness[k]
			total += bobotReadiness[k]
		}
	}
	add(sevOk, "severity")
	add(drugOk, "drug")
	add(singleOk, "singleDiag")
	add(semuaValid, "icd")
	add(losOk, "los")
	return ReadinessResult{Total: total, Komponen: komp, Flag: flag}
}

// ---------- Anomali ----------

type Anomali struct {
	EpisodeID string `json:"episodeId"`
	Jenis     string `json:"jenis"`
	Detail    string `json:"detail"`
}

func diagnosaTerkait(a, b []string) bool {
	setB := map[string]bool{}
	kategoriB := map[string]bool{}
	for _, x := range b {
		setB[x] = true
		if len(x) >= 3 {
			kategoriB[x[:3]] = true
		}
	}
	for _, x := range a {
		if setB[x] || (len(x) >= 3 && kategoriB[x[:3]]) {
			return true
		}
	}
	return false
}

func upperAll(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToUpper(s)
	}
	return out
}

// DeteksiAnomali: LOS outlier, billing nol/dominan, SEP ganda, readmisi ≤ 30 hari.
// Masukan harus episode aktif (bukan superseded).
func DeteksiAnomali(aktif []*Episode) []Anomali {
	hasil := []Anomali{}

	perGrup := map[string][]float64{}
	for _, ep := range aktif {
		b := BasisGrupCbg(ep.KodeIcbg)
		perGrup[b] = append(perGrup[b], float64(ep.Los))
	}
	for _, ep := range aktif {
		b := BasisGrupCbg(ep.KodeIcbg)
		med := Median(perGrup[b])
		if med > 0 && float64(ep.Los) > 2.5*med {
			hasil = append(hasil, Anomali{ep.ID, "LOS_OUTLIER",
				"LOS " + itoa(ep.Los) + " hari > 2,5× median grup " + b + " (" + ftoa1(med) + " hari)"})
		}
		if ep.TotalBilling == 0 {
			hasil = append(hasil, Anomali{ep.ID, "SELISIH_EKSTREM", "Total billing = 0 — komponen billing tidak terisi"})
			continue
		}
		var dominan int64
		for _, v := range ep.Komponen {
			if v > dominan {
				dominan = v
			}
		}
		if float64(dominan)/float64(ep.TotalBilling) > 0.8 {
			hasil = append(hasil, Anomali{ep.ID, "SELISIH_EKSTREM", "Satu komponen billing > 80% dari total — perlu verifikasi rincian"})
		}
	}

	// SEP ganda pada data aktif (seharusnya mustahil lewat indeks unik; tetap dicek).
	bySep := map[string][]*Episode{}
	var urutSep []string
	for _, ep := range aktif {
		if _, ok := bySep[ep.NoSep]; !ok {
			urutSep = append(urutSep, ep.NoSep)
		}
		bySep[ep.NoSep] = append(bySep[ep.NoSep], ep)
	}
	for _, sep := range urutSep {
		if l := bySep[sep]; len(l) > 1 {
			hasil = append(hasil, Anomali{l[0].ID, "DUPLIKAT", "SEP " + sep + " muncul " + itoa(len(l)) + "× pada data aktif"})
		}
	}

	// Readmisi ≤ 30 hari dengan diagnosis sama/terkait.
	byMr := map[string][]*Episode{}
	var urutMr []string
	for _, ep := range aktif {
		if _, ok := byMr[ep.NoMr]; !ok {
			urutMr = append(urutMr, ep.NoMr)
		}
		byMr[ep.NoMr] = append(byMr[ep.NoMr], ep)
	}
	for _, mr := range urutMr {
		l := append([]*Episode(nil), byMr[mr]...)
		sort.Slice(l, func(i, j int) bool { return l[i].TglMasuk < l[j].TglMasuk })
		for i := 1; i < len(l); i++ {
			prev, cur := l[i-1], l[i]
			hari := SelisihHari(prev.TglKeluar, cur.TglMasuk)
			if hari >= 0 && hari <= 30 && diagnosaTerkait(upperAll(prev.Diagnosa), upperAll(cur.Diagnosa)) {
				hasil = append(hasil, Anomali{cur.ID, "READMISI",
					"Masuk ulang " + itoa(hari) + " hari setelah pulang (diagnosis sama/terkait, MR " + cur.NoMr + ")"})
			}
		}
	}
	return hasil
}

// ---------- Estimasi upgrade tarif ----------

// EstimasiUpgradeTarif memperkirakan tambahan tarif bila severity naik satu
// level, dibatasi oleh besar rugi saat ini. Kunci hasil: Episode.ID.
func EstimasiUpgradeTarif(episodes []*Episode) map[string]int64 {
	grup := map[string]map[int][2]float64{} // basis → severity → {total tarif, n}
	for _, ep := range episodes {
		if ep.SeverityLevel < 1 {
			continue
		}
		b := BasisGrupCbg(ep.KodeIcbg)
		if grup[b] == nil {
			grup[b] = map[int][2]float64{}
		}
		cur := grup[b][ep.SeverityLevel]
		cur[0] += float64(ep.TarifInacbg)
		cur[1]++
		grup[b][ep.SeverityLevel] = cur
	}
	est := map[string]int64{}
	for _, ep := range episodes {
		if ep.KelasRugi == "SURPLUS" || ep.SeverityLevel < 1 {
			est[ep.ID] = 0
			continue
		}
		next, ok := grup[BasisGrupCbg(ep.KodeIcbg)][ep.SeverityLevel+1]
		if !ok || next[1] == 0 {
			est[ep.ID] = 0
			continue
		}
		delta := next[0]/next[1] - float64(ep.TarifInacbg)
		rugi := math.Abs(float64(ep.Selisih))
		est[ep.ID] = int64(math.Max(0, math.Min(delta, rugi)))
	}
	return est
}

// ---------- Format kecil ----------

func itoa(n int) string { return strconv.Itoa(n) }

// ftoa1: 2.5 -> "2,5" (koma desimal gaya Indonesia).
func ftoa1(f float64) string {
	return strings.Replace(fmt.Sprintf("%.1f", f), ".", ",", 1)
}

// FormatRibuan: 1234567 -> "1.234.567".
func FormatRibuan(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	d := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i, c := range d {
		if i > 0 && (len(d)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}
