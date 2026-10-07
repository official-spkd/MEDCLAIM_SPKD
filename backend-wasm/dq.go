// ============================================================
// MedClaim Go — Kualitas data: 20 aturan DQ, skor 7 dimensi,
// grade & gerbang (port dari quality.ts, Bagian 11).
// ============================================================
package main

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

var PenaltiTetap = map[string]int{"KRITIS": 100, "PERINGATAN": 25, "INFO": 10}

const AmbangOverride = 60

type DQAturan struct {
	Kode      string `json:"kode"`
	Nama      string `json:"nama"`
	Dimensi   string `json:"dimensi"`
	Severitas string `json:"severitas"`
	Terpicu   bool   `json:"terpicu"`
	Jumlah    int    `json:"jumlah"`
	Pesan     string `json:"pesan"`
}

type DQDimensi struct {
	Kode  string `json:"kode"`
	Label string `json:"label"`
	Skor  int    `json:"skor"`
}

type MatriksModul struct {
	Modul  string `json:"modul"`
	Status string `json:"status"`
}

type DQReport struct {
	Skor            int            `json:"skor"`
	Grade           string         `json:"grade"`
	Aturan          []DQAturan     `json:"aturan"`
	Dimensi         []DQDimensi    `json:"dimensi"`
	MatriksModul    []MatriksModul `json:"matriksModul"`
	PeriodeUsulan   string         `json:"periodeUsulan"`
	Gerbang         string         `json:"gerbang"`
}

type KonteksDQ struct {
	KolomTersedia          []string
	PeriodeDipilih         string
	MedianUploadSebelumnya float64
	SepEksisting           map[string]bool
}

var labelDimensi = [][2]string{
	{"COMPLETENESS", "Kelengkapan"},
	{"VALIDITY", "Keabsahan"},
	{"CONSISTENCY", "Konsistensi"},
	{"TIMELINESS", "Ketepatan Waktu"},
	{"UNIQUENESS", "Keunikan"},
	{"ACCURACY", "Akurasi"},
	{"CONFORMITY", "Kesesuaian"},
}

var kolomWajib = []string{
	"NOSEP", "NOKARTU", "NOMR", "NAMAPASIEN", "TGLLAHIR", "JENISKELAMIN",
	"TGLMASUK", "TGLKELUAR", "LOS", "KELASRAWAT", "TIPELAYANAN", "DIAGNOSA",
	"KODEICBG", "TARIFINACBG",
}

var reIcdFormat = regexp.MustCompile(`^[A-Z]\d{2}(\.\d{1,2})?$`)

func hariIni() string {
	return time.Now().Format(layoutISO)
}

// evaluasiAturanDQ — 20 aturan atas baris valid + konteks upload.
func evaluasiAturanDQ(baris []BarisValid, gagal int, ctx KonteksDQ) []DQAturan {
	setKolom := map[string]bool{}
	for _, k := range ctx.KolomTersedia {
		setKolom[strings.ToUpper(k)] = true
	}
	ada := func(k string) bool { return setKolom[k] }
	var hilang []string
	for _, k := range kolomWajib {
		if !ada(k) {
			hilang = append(hilang, k)
		}
	}

	off := 0
	terbalik := 0
	tarifNol := 0
	icdInvalid := 0
	kelasInvalid := 0
	jkInvalid := 0
	losTidakCocok := 0
	rjtpLos := 0
	dpjpKosong := 0
	usiaAnomali := 0
	billingNolRanap := 0
	masaDepan := 0
	sepHitung := map[string]int{}
	drugViolation := 0

	for i := range baris {
		b := &baris[i]
		if tanggalValid(b.TglMasuk) && b.TglMasuk[:7] != ctx.PeriodeDipilih {
			off++
		}
		if tanggalValid(b.TglMasuk) && tanggalValid(b.TglKeluar) && b.TglKeluar < b.TglMasuk {
			terbalik++
		}
		if b.TarifInacbg <= 0 {
			tarifNol++
		}
		for _, d := range b.Diagnosa {
			if !reIcdFormat.MatchString(strings.ToUpper(d)) {
				icdInvalid++
				break
			}
		}
		if b.KelasRawat != 1 && b.KelasRawat != 2 && b.KelasRawat != 3 {
			kelasInvalid++
		}
		if b.JenisKelamin != "L" && b.JenisKelamin != "P" {
			jkInvalid++
		}
		if tanggalValid(b.TglMasuk) && tanggalValid(b.TglKeluar) {
			hari := selisihHari(b.TglMasuk, b.TglKeluar)
			if hari < 0 || abs(hari-b.Los) > 1 {
				losTidakCocok++
			}
		}
		if b.TipeLayanan == "RJTP" && b.Los > 1 {
			rjtpLos++
		}
		if b.DpjpKode == "" {
			dpjpKosong++
		}
		if tanggalValid(b.TglLahir) && tanggalValid(b.TglMasuk) {
			umur := selisihHari(b.TglLahir, b.TglMasuk)
			if umur < 0 || float64(umur) > 120*365.25 {
				usiaAnomali++
			}
		}
		if b.TipeLayanan == "RANAP" && b.TotalBilling == 0 {
			billingNolRanap++
		}
		if tanggalValid(b.TglKeluar) && b.TglKeluar > hariIni() {
			masaDepan++
		}
		sepHitung[b.NoSep]++
		kemo := b.Komponen["OBAT_KEMO"] > 0
		kronis := b.Komponen["OBAT_KRONIS"]+b.Komponen["DRUG_KRONIS_FEE"] > 0
		neoplasma, kronisValid := false, false
		for _, d := range b.Diagnosa {
			up := strings.ToUpper(d)
			if regexp.MustCompile(`^C\d{2}`).MatchString(up) {
				neoplasma = true
			}
			if reKronis.MatchString(up) {
				kronisValid = true
			}
		}
		if (kemo && !neoplasma) || (kronis && !kronisValid) {
			drugViolation++
		}
	}
	sepGanda := 0
	for _, n := range sepHitung {
		if n > 1 {
			sepGanda += n
		}
	}
	sepEksisting := 0
	for i := range baris {
		if ctx.SepEksisting[baris[i].NoSep] {
			sepEksisting++
		}
	}

	pctOff := 0.0
	if len(baris) > 0 {
		pctOff = float64(off) / float64(len(baris))
	}
	dpjpPct := 1.0
	if len(baris) > 0 {
		dpjpPct = float64(dpjpKosong) / float64(len(baris))
	}
	deviasiMedian := -1.0
	if ctx.MedianUploadSebelumnya > 0 {
		deviasiMedian = float64(len(baris)) / ctx.MedianUploadSebelumnya
	}

	aturan := []DQAturan{
		{Kode: "DQ-COL-01", Nama: "Kolom wajib hilang", Dimensi: "COMPLETENESS", Severitas: "KRITIS", Jumlah: len(hilang), Pesan: pesanHilang(hilang)},
		{Kode: "DQ-COL-02", Nama: "Kekayaan kolom (format Basic)", Dimensi: "COMPLETENESS", Severitas: "INFO", Jumlah: boolInt(!ada("VISITATION_FEE")), Pesan: pesanAda(ada("VISITATION_FEE"), "Format Detail — komponen billing tersedia", "Ekspor Basic tanpa rincian 18 komponen billing — modul finansial terbatas")},
		{Kode: "DQ-COL-03", Nama: "Kolom penting kosong (DPJP/cara pulang)", Dimensi: "COMPLETENESS", Severitas: "PERINGATAN", Jumlah: pctInt(dpjpPct, len(baris)), Pesan: pesanPersen(dpjpPct, len(baris), "% baris tanpa kode DPJP", "Kolom DPJP/cara pulang terisi memadai")},
		{Kode: "DQ-VAL-01", Nama: "Format tanggal tidak valid", Dimensi: "VALIDITY", Severitas: "KRITIS", Jumlah: hitungTanggalTakValid(baris), Pesan: "Tanggal harus format yyyy-mm-dd"},
		{Kode: "DQ-VAL-02", Nama: "Tanggal terbalik", Dimensi: "VALIDITY", Severitas: "KRITIS", Jumlah: terbalik, Pesan: "Tanggal keluar lebih awal dari tanggal masuk"},
		{Kode: "DQ-VAL-03", Nama: "Tarif INA-CBG ≤ 0", Dimensi: "VALIDITY", Severitas: "PERINGATAN", Jumlah: tarifNol, Pesan: "Tarif INA-CBG harus bernilai positif"},
		{Kode: "DQ-VAL-04", Nama: "Format kode ICD tidak valid", Dimensi: "VALIDITY", Severitas: "PERINGATAN", Jumlah: icdInvalid, Pesan: "Kode ICD-10 harus berpola A##(.#) — cek pemisah titik/koma"},
		{Kode: "DQ-VAL-05", Nama: "Kelas rawat di luar 1/2/3", Dimensi: "VALIDITY", Severitas: "KRITIS", Jumlah: kelasInvalid, Pesan: "Kelas rawat hanya boleh 1, 2, atau 3"},
		{Kode: "DQ-VAL-06", Nama: "Jenis kelamin bukan L/P", Dimensi: "VALIDITY", Severitas: "PERINGATAN", Jumlah: jkInvalid, Pesan: "Jenis kelamin harus L atau P"},
		{Kode: "DQ-CON-01", Nama: "LOS tidak cocok dengan tanggal", Dimensi: "CONSISTENCY", Severitas: "PERINGATAN", Jumlah: losTidakCocok, Pesan: "LOS tidak sama dengan selisih tanggal masuk–keluar"},
		{Kode: "DQ-CON-02", Nama: "RJTP dengan LOS > 1 hari", Dimensi: "CONSISTENCY", Severitas: "PERINGATAN", Jumlah: rjtpLos, Pesan: "Rawat jalan/top-up dengan LOS > 1 hari janggal"},
		{Kode: "DQ-CON-03", Nama: "Drug alignment tidak konsisten", Dimensi: "CONSISTENCY", Severitas: "PERINGATAN", Jumlah: drugViolation, Pesan: "OBAT_KEMO tanpa diagnosis C## / obat kronis tanpa diagnosis kronis"},
		{Kode: "DQ-CUP-01", Nama: "Tanggal di luar periode terpilih", Dimensi: "TIMELINESS", Severitas: "PERINGATAN", Jumlah: cupJumlah(pctOff, off), Pesan: cupPesan(pctOff, ctx.PeriodeDipilih)},
		{Kode: "DQ-CUP-02", Nama: "Volume menyimpang dari median upload sebelumnya", Dimensi: "TIMELINESS", Severitas: "PERINGATAN", Jumlah: cup02Jumlah(deviasiMedian, len(baris)), Pesan: cup02Pesan(deviasiMedian, len(baris))},
		{Kode: "DQ-TIM-01", Nama: "Tanggal keluar di masa depan", Dimensi: "TIMELINESS", Severitas: "PERINGATAN", Jumlah: masaDepan, Pesan: "Terdapat tanggal keluar lebih besar dari hari ini — data belum final"},
		{Kode: "DQ-UNQ-01", Nama: "SEP ganda dalam file", Dimensi: "UNIQUENESS", Severitas: "KRITIS", Jumlah: sepGanda, Pesan: "Nomor SEP harus unik dalam satu file"},
		{Kode: "DQ-UNQ-02", Nama: "SEP sudah ada pada upload sebelumnya", Dimensi: "UNIQUENESS", Severitas: "INFO", Jumlah: sepEksisting, Pesan: "SEP terdeteksi pada upload sebelumnya — akan diproses sebagai versi"},
		{Kode: "DQ-ACC-01", Nama: "Usia tidak wajar", Dimensi: "ACCURACY", Severitas: "PERINGATAN", Jumlah: usiaAnomali, Pesan: "Usia dihitung dari tanggal lahir bernilai < 0 atau > 120 tahun"},
		{Kode: "DQ-ACC-02", Nama: "Billing RANAP tanpa rincian", Dimensi: "ACCURACY", Severitas: "PERINGATAN", Jumlah: billingNolRanap, Pesan: "Rawat inap dengan total billing 0 — rincian komponen tidak terisi"},
		{Kode: "DQ-CNF-01", Nama: "Struktur header tidak sesuai templat V-Claim", Dimensi: "CONFORMITY", Severitas: "PERINGATAN", Jumlah: cnf01(ada), Pesan: pesanCnf01(ada)},
	}
	for i := range aturan {
		aturan[i].Terpicu = aturan[i].Jumlah > 0
	}
	return aturan
}

func pesanHilang(hilang []string) string {
	if len(hilang) == 0 {
		return "Seluruh kolom wajib tersedia"
	}
	return "Kolom wajib tidak ditemukan: " + strings.Join(hilang, ", ")
}

func pesanAda(ada bool, ok, tidak string) string {
	if ada {
		return ok
	}
	return tidak
}

func pesanPersen(pct float64, n int, ok, tidak string) string {
	if pct > 0.1 {
		return itoa(int(math.Round(pct*100))) + ok
	}
	return tidak
}

func cupJumlah(pct float64, off int) int {
	if pct > 0.05 {
		return off
	}
	return 0
}

func cupPesan(pct float64, periode string) string {
	if pct > 0.05 {
		return itoa(int(math.Round(pct*100))) + "% tanggal masuk di luar " + periode + " (ambang 5%)"
	}
	return "Tanggal masuk sesuai periode"
}

func cup02Jumlah(dev float64, n int) int {
	if dev >= 0 && (dev < 0.5 || dev > 2) {
		return n
	}
	return 0
}

func cup02Pesan(dev float64, n int) string {
	if dev >= 0 && (dev < 0.5 || dev > 2) {
		return "Baris " + itoa(n) + " vs median " + itoa(int(dev*float64(n))) + " (<50% / >200%)"
	}
	return "Volume dalam batas wajar"
}

func cnf01(ada func(string) bool) int {
	if !ada("NOSEP") || !ada("DIAGNOSA") {
		return 1
	}
	return 0
}

func pesanCnf01(ada func(string) bool) string {
	if !ada("NOSEP") {
		return "Header tidak memuat kolom kunci V-Claim (NOSEP/DIAGNOSA)"
	}
	return "Header sesuai templat"
}

func hitungTanggalTakValid(baris []BarisValid) int {
	n := 0
	for _, b := range baris {
		if !tanggalValid(b.TglMasuk) || !tanggalValid(b.TglKeluar) {
			n++
		}
	}
	return n
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func pctInt(pct float64, total int) int {
	if pct > 0.1 {
		return int(math.Round(pct * float64(total)))
	}
	return 0
}

func itoa(n int) string {
	return fmtInt64(int64(n))
}

// buatLaporanDQ — skor, grade, dimensi, matriks modul, gerbang & periode usulan.
func buatLaporanDQ(baris []BarisValid, gagal int, ctx KonteksDQ) DQReport {
	aturan := evaluasiAturanDQ(baris, gagal, ctx)
	penalti := 0
	kritis := false
	for _, a := range aturan {
		if a.Terpicu {
			penalti += PenaltiTetap[a.Severitas]
			if a.Severitas == "KRITIS" {
				kritis = true
			}
		}
	}
	skor := 100 - penalti
	if skor < 0 {
		skor = 0
	}
	if skor > 100 {
		skor = 100
	}
	grade := "D"
	if skor >= 90 {
		grade = "A"
	} else if skor >= 80 {
		grade = "B"
	} else if skor >= 60 {
		grade = "C"
	}

	dimensi := []DQDimensi{}
	for _, d := range labelDimensi {
		penaltiDim := 0
		for _, a := range aturan {
			if a.Dimensi == d[0] && a.Terpicu {
				penaltiDim += PenaltiTetap[a.Severitas]
			}
		}
		s := 100 - penaltiDim
		if s < 0 {
			s = 0
		}
		dimensi = append(dimensi, DQDimensi{Kode: d[0], Label: d[1], Skor: s})
	}

	setKolom := map[string]bool{}
	for _, k := range ctx.KolomTersedia {
		setKolom[strings.ToUpper(k)] = true
	}
	detail := setKolom["VISITATION_FEE"]
	statusModul := func(penuh bool, terbatas string) string {
		if penuh {
			return "PENUH"
		}
		return terbatas
	}
	matriks := []MatriksModul{
		{Modul: "Dashboard", Status: "PENUH"},
		{Modul: "Claim Explorer", Status: "PENUH"},
		{Modul: "Patient Tracer", Status: "PENUH"},
		{Modul: "Financial Intelligence", Status: statusModul(detail, "TERBATAS")},
		{Modul: "Diagnostic Costs", Status: statusModul(setKolom["LAB"] && setKolom["RAD"], "TERBATAS")},
		{Modul: "Add-on Tarif", Status: func() string {
			if setKolom["KODE_ADDON"] {
				return "PENUH"
			}
			return "TIDAK_TERSEDIA"
		}()},
		{Modul: "DPJP Performance", Status: statusModul(setKolom["DPJPKODE"], "TERBATAS")},
		{Modul: "Claim Quality", Status: "PENUH"},
		{Modul: "Anomaly Detection", Status: "PENUH"},
	}

	gerbang := "LOLOS"
	if skor < AmbangOverride || kritis {
		gerbang = "BLOKIR"
	}

	// Periode usulan: modus bulan tanggal masuk
	hitungBulan := map[string]int{}
	for _, b := range baris {
		if tanggalValid(b.TglMasuk) {
			hitungBulan[b.TglMasuk[:7]]++
		}
	}
	type kv struct {
		k string
		v int
	}
	var daftar []kv
	for k, v := range hitungBulan {
		daftar = append(daftar, kv{k, v})
	}
	sort.Slice(daftar, func(i, j int) bool {
		if daftar[i].v != daftar[j].v {
			return daftar[i].v > daftar[j].v
		}
		return daftar[i].k < daftar[j].k
	})
	periodeUsulan := ctx.PeriodeDipilih
	if len(daftar) > 0 {
		periodeUsulan = daftar[0].k
	}

	return DQReport{
		Skor: skor, Grade: grade, Aturan: aturan, Dimensi: dimensi,
		MatriksModul: matriks, PeriodeUsulan: periodeUsulan, Gerbang: gerbang,
	}
}
