package uploads

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"medclaim-backend/internal/scoring"
)

// Kualitas data (Data Quality): 20 aturan, skor 7 dimensi, grade, dan gerbang
// upload (LOLOS / BLOKIR / OVERRIDE).

var penaltiTetap = map[string]int{"KRITIS": 100, "PERINGATAN": 25, "INFO": 10}

// Skor di bawah ambang ini (atau ada aturan KRITIS terpicu) memblokir commit.
const ambangOverride = 60

type DQAturan struct {
	Kode      string `json:"kode"`
	Nama      string `json:"nama"`
	Dimensi   string `json:"dimensi"`
	Severitas string `json:"severitas"`
	Terpicu   bool   `json:"terpicu"`
	Jumlah    int    `json:"jumlah"`
	Pesan     string `json:"pesan"`
	// Baris = rincian baris yang melanggar aturan (dibatasi batasBarisDQ per aturan; Terpotong = ada yang tidak ikut).
	Baris     []DQBaris `json:"baris,omitempty"`
	Terpotong bool      `json:"terpotong,omitempty"`
}

// DQBaris menunjuk satu baris file yang perlu diperbaiki.
type DQBaris struct {
	Nomor      int    `json:"nomor"`
	NoSep      string `json:"noSep"`
	NoMr       string `json:"noMr"`
	Nama       string `json:"nama"`
	Keterangan string `json:"keterangan"`
}

// Batas rincian baris per aturan agar laporan tetap ringan; jumlah total tetap akurat.
const batasBarisDQ = 500

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
	Skor          int            `json:"skor"`
	Grade         string         `json:"grade"`
	Aturan        []DQAturan     `json:"aturan"`
	Dimensi       []DQDimensi    `json:"dimensi"`
	MatriksModul  []MatriksModul `json:"matriksModul"`
	PeriodeUsulan string         `json:"periodeUsulan"`
	Gerbang       string         `json:"gerbang"`
}

// KonteksDQ = informasi di luar isi file yang dibutuhkan beberapa aturan.
type KonteksDQ struct {
	KolomTersedia          []string
	PeriodeDipilih         string
	MedianUploadSebelumnya float64
	SepEksisting           map[string]bool
}

var labelDimensi = [][2]string{
	{"COMPLETENESS", "Kelengkapan"}, {"VALIDITY", "Keabsahan"}, {"CONSISTENCY", "Konsistensi"},
	{"TIMELINESS", "Ketepatan Waktu"}, {"UNIQUENESS", "Keunikan"}, {"ACCURACY", "Akurasi"},
	{"CONFORMITY", "Kesesuaian"},
}

var kolomWajib = []string{
	"NOSEP", "NOKARTU", "NOMR", "NAMAPASIEN", "TGLLAHIR", "JENISKELAMIN",
	"TGLMASUK", "TGLKELUAR", "LOS", "KELASRAWAT", "TIPELAYANAN", "DIAGNOSA",
	"KODEICBG", "TARIFINACBG",
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func pct(p float64) string { return strconv.Itoa(int(math.Round(p * 100))) }

func evaluasiAturanDQ(baris []BarisValid, ctx KonteksDQ) []DQAturan {
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

	rincian := map[string][]DQBaris{}
	tandai := func(kode string, b *BarisValid, ket string) {
		if len(rincian[kode]) < batasBarisDQ {
			rincian[kode] = append(rincian[kode], DQBaris{Nomor: b.Nomor, NoSep: b.NoSep, NoMr: b.NoMr, Nama: b.NamaPasien, Keterangan: ket})
		}
	}

	var off, terbalik, tarifNol, icdInvalid, kelasInvalid, jkInvalid int
	var losTidakCocok, rjtpLos, dpjpKosong, usiaAnomali, billingNolRanap, masaDepan, drugViolation, tglTakValid int
	sepHitung := map[string]int{}
	hariIni := time.Now().Format(scoring.LayoutISO)

	for i := range baris {
		b := &baris[i]
		okMasuk, okKeluar := scoring.TanggalValid(b.TglMasuk), scoring.TanggalValid(b.TglKeluar)
		if !okMasuk || !okKeluar {
			tglTakValid++
			tandai("DQ-VAL-01", b, "Masuk: "+b.TglMasuk+" · Keluar: "+b.TglKeluar)
		}
		if okMasuk && b.TglMasuk[:7] != ctx.PeriodeDipilih {
			off++
			tandai("DQ-CUP-01", b, "Tanggal masuk "+b.TglMasuk+" di luar periode "+ctx.PeriodeDipilih)
		}
		if okMasuk && okKeluar && b.TglKeluar < b.TglMasuk {
			terbalik++
			tandai("DQ-VAL-02", b, "Masuk "+b.TglMasuk+" · Keluar "+b.TglKeluar)
		}
		if b.TarifInacbg <= 0 {
			tarifNol++
			tandai("DQ-VAL-03", b, "Tarif INA-CBG: "+strconv.FormatInt(b.TarifInacbg, 10))
		}
		for _, d := range b.Diagnosa {
			if !scoring.ICD10Valid(d) {
				icdInvalid++
				tandai("DQ-VAL-04", b, "Kode ICD tidak valid: "+d)
				break
			}
		}
		if b.KelasRawat < 1 || b.KelasRawat > 3 {
			kelasInvalid++
			tandai("DQ-VAL-05", b, "Kelas rawat: "+strconv.Itoa(b.KelasRawat))
		}
		if b.JenisKelamin != "L" && b.JenisKelamin != "P" {
			jkInvalid++
			tandai("DQ-VAL-06", b, "Jenis kelamin: "+b.JenisKelamin)
		}
		if okMasuk && okKeluar {
			hari := scoring.SelisihHari(b.TglMasuk, b.TglKeluar)
			if hari < 0 || scoring.Abs(hari-b.Los) > 1 {
				losTidakCocok++
				tandai("DQ-CON-01", b, "LOS "+strconv.Itoa(b.Los)+" hari, selisih tanggal "+strconv.Itoa(hari)+" hari")
			}
		}
		if b.TipeLayanan == "RJTP" && b.Los > 1 {
			rjtpLos++
			tandai("DQ-CON-02", b, "RJTP dengan LOS "+strconv.Itoa(b.Los)+" hari")
		}
		if b.DpjpKode == "" {
			dpjpKosong++
			tandai("DQ-COL-03", b, "Kode DPJP kosong")
		}
		if scoring.TanggalValid(b.TglLahir) && okMasuk {
			umur := scoring.SelisihHari(b.TglLahir, b.TglMasuk)
			if umur < 0 || float64(umur) > 120*365.25 {
				usiaAnomali++
				tandai("DQ-ACC-01", b, "Tgl lahir "+b.TglLahir+" · masuk "+b.TglMasuk)
			}
		}
		if b.TipeLayanan == "RANAP" && b.TotalBilling == 0 {
			billingNolRanap++
			tandai("DQ-ACC-02", b, "Rawat inap dengan total billing 0")
		}
		if okKeluar && b.TglKeluar > hariIni {
			masaDepan++
			tandai("DQ-TIM-01", b, "Tanggal keluar "+b.TglKeluar)
		}
		sepHitung[b.NoSep]++
		if ok, flag := scoring.DrugAlignment(b.Diagnosa, b.Komponen); !ok {
			drugViolation++
			tandai("DQ-CON-03", b, strings.Join(flag, "; "))
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
		if n := sepHitung[baris[i].NoSep]; n > 1 {
			tandai("DQ-UNQ-01", &baris[i], "SEP muncul "+strconv.Itoa(n)+"× dalam file")
		}
		if ctx.SepEksisting[baris[i].NoSep] {
			sepEksisting++
			tandai("DQ-UNQ-02", &baris[i], "SEP sudah ada pada upload sebelumnya")
		}
	}

	n := len(baris)
	pctOff, dpjpPct := 0.0, 1.0
	if n > 0 {
		pctOff = float64(off) / float64(n)
		dpjpPct = float64(dpjpKosong) / float64(n)
	}
	deviasi := -1.0
	if ctx.MedianUploadSebelumnya > 0 {
		deviasi = float64(n) / ctx.MedianUploadSebelumnya
	}

	// Pesan & jumlah bergantung pada ambang masing-masing aturan.
	pesanHilang := "Seluruh kolom wajib tersedia"
	if len(hilang) > 0 {
		pesanHilang = "Kolom wajib tidak ditemukan: " + strings.Join(hilang, ", ")
	}
	col02 := "Ekspor Basic tanpa rincian 18 komponen billing — modul finansial terbatas"
	if ada("VISITATION_FEE") {
		col02 = "Format Detail — komponen billing tersedia"
	}
	col03Jml, col03Pesan := 0, "Kolom DPJP/cara pulang terisi memadai"
	if dpjpPct > 0.1 {
		col03Jml = int(math.Round(dpjpPct * float64(n)))
		col03Pesan = pct(dpjpPct) + "% baris tanpa kode DPJP"
	}
	cup01Jml, cup01Pesan := 0, "Tanggal masuk sesuai periode"
	if pctOff > 0.05 {
		cup01Jml = off
		cup01Pesan = pct(pctOff) + "% tanggal masuk di luar " + ctx.PeriodeDipilih + " (ambang 5%)"
	}
	cup02Jml, cup02Pesan := 0, "Volume dalam batas wajar"
	if deviasi >= 0 && (deviasi < 0.5 || deviasi > 2) {
		cup02Jml = n
		cup02Pesan = "Baris " + strconv.Itoa(n) + " vs median " + strconv.Itoa(int(math.Round(ctx.MedianUploadSebelumnya))) + " (<50% / >200%)"
	}
	cnfJml, cnfPesan := 0, "Header sesuai templat"
	if !ada("NOSEP") || !ada("DIAGNOSA") {
		cnfJml = 1
		if !ada("NOSEP") {
			cnfPesan = "Header tidak memuat kolom kunci V-Claim (NOSEP/DIAGNOSA)"
		}
	}

	aturan := []DQAturan{
		{Kode: "DQ-COL-01", Nama: "Kolom wajib hilang", Dimensi: "COMPLETENESS", Severitas: "KRITIS", Jumlah: len(hilang), Pesan: pesanHilang},
		{Kode: "DQ-COL-02", Nama: "Kekayaan kolom (format Basic)", Dimensi: "COMPLETENESS", Severitas: "INFO", Jumlah: boolInt(!ada("VISITATION_FEE")), Pesan: col02},
		{Kode: "DQ-COL-03", Nama: "Kolom penting kosong (DPJP/cara pulang)", Dimensi: "COMPLETENESS", Severitas: "PERINGATAN", Jumlah: col03Jml, Pesan: col03Pesan},
		{Kode: "DQ-VAL-01", Nama: "Format tanggal tidak valid", Dimensi: "VALIDITY", Severitas: "KRITIS", Jumlah: tglTakValid, Pesan: "Tanggal harus format yyyy-mm-dd"},
		{Kode: "DQ-VAL-02", Nama: "Tanggal terbalik", Dimensi: "VALIDITY", Severitas: "KRITIS", Jumlah: terbalik, Pesan: "Tanggal keluar lebih awal dari tanggal masuk"},
		{Kode: "DQ-VAL-03", Nama: "Tarif INA-CBG ≤ 0", Dimensi: "VALIDITY", Severitas: "PERINGATAN", Jumlah: tarifNol, Pesan: "Tarif INA-CBG harus bernilai positif"},
		{Kode: "DQ-VAL-04", Nama: "Format kode ICD tidak valid", Dimensi: "VALIDITY", Severitas: "PERINGATAN", Jumlah: icdInvalid, Pesan: "Kode ICD-10 harus berpola A##(.#) — cek pemisah titik/koma"},
		{Kode: "DQ-VAL-05", Nama: "Kelas rawat di luar 1/2/3", Dimensi: "VALIDITY", Severitas: "KRITIS", Jumlah: kelasInvalid, Pesan: "Kelas rawat hanya boleh 1, 2, atau 3"},
		{Kode: "DQ-VAL-06", Nama: "Jenis kelamin bukan L/P", Dimensi: "VALIDITY", Severitas: "PERINGATAN", Jumlah: jkInvalid, Pesan: "Jenis kelamin harus L atau P"},
		{Kode: "DQ-CON-01", Nama: "LOS tidak cocok dengan tanggal", Dimensi: "CONSISTENCY", Severitas: "PERINGATAN", Jumlah: losTidakCocok, Pesan: "LOS tidak sama dengan selisih tanggal masuk–keluar"},
		{Kode: "DQ-CON-02", Nama: "RJTP dengan LOS > 1 hari", Dimensi: "CONSISTENCY", Severitas: "PERINGATAN", Jumlah: rjtpLos, Pesan: "Rawat jalan/top-up dengan LOS > 1 hari janggal"},
		{Kode: "DQ-CON-03", Nama: "Drug alignment tidak konsisten", Dimensi: "CONSISTENCY", Severitas: "PERINGATAN", Jumlah: drugViolation, Pesan: "OBAT_KEMO tanpa diagnosis C## / obat kronis tanpa diagnosis kronis"},
		{Kode: "DQ-CUP-01", Nama: "Tanggal di luar periode terpilih", Dimensi: "TIMELINESS", Severitas: "PERINGATAN", Jumlah: cup01Jml, Pesan: cup01Pesan},
		{Kode: "DQ-CUP-02", Nama: "Volume menyimpang dari median upload sebelumnya", Dimensi: "TIMELINESS", Severitas: "PERINGATAN", Jumlah: cup02Jml, Pesan: cup02Pesan},
		{Kode: "DQ-TIM-01", Nama: "Tanggal keluar di masa depan", Dimensi: "TIMELINESS", Severitas: "PERINGATAN", Jumlah: masaDepan, Pesan: "Terdapat tanggal keluar lebih besar dari hari ini — data belum final"},
		{Kode: "DQ-UNQ-01", Nama: "SEP ganda dalam file", Dimensi: "UNIQUENESS", Severitas: "KRITIS", Jumlah: sepGanda, Pesan: "Nomor SEP harus unik dalam satu file"},
		{Kode: "DQ-UNQ-02", Nama: "SEP sudah ada pada upload sebelumnya", Dimensi: "UNIQUENESS", Severitas: "INFO", Jumlah: sepEksisting, Pesan: "SEP terdeteksi pada upload sebelumnya — akan diproses sebagai versi"},
		{Kode: "DQ-ACC-01", Nama: "Usia tidak wajar", Dimensi: "ACCURACY", Severitas: "PERINGATAN", Jumlah: usiaAnomali, Pesan: "Usia dihitung dari tanggal lahir bernilai < 0 atau > 120 tahun"},
		{Kode: "DQ-ACC-02", Nama: "Billing RANAP tanpa rincian", Dimensi: "ACCURACY", Severitas: "PERINGATAN", Jumlah: billingNolRanap, Pesan: "Rawat inap dengan total billing 0 — rincian komponen tidak terisi"},
		{Kode: "DQ-CNF-01", Nama: "Struktur header tidak sesuai templat V-Claim", Dimensi: "CONFORMITY", Severitas: "PERINGATAN", Jumlah: cnfJml, Pesan: cnfPesan},
	}
	for i := range aturan {
		aturan[i].Terpicu = aturan[i].Jumlah > 0
		if aturan[i].Terpicu {
			aturan[i].Baris = rincian[aturan[i].Kode]
			aturan[i].Terpotong = aturan[i].Jumlah > len(aturan[i].Baris) && len(aturan[i].Baris) > 0
		}
	}
	return aturan
}

// buatLaporanDQ menghitung skor, grade, skor per dimensi, matriks ketersediaan
// modul, periode usulan, dan status gerbang.
func buatLaporanDQ(baris []BarisValid, ctx KonteksDQ) DQReport {
	aturan := evaluasiAturanDQ(baris, ctx)
	penalti, kritis := 0, false
	for _, a := range aturan {
		if a.Terpicu {
			penalti += penaltiTetap[a.Severitas]
			kritis = kritis || a.Severitas == "KRITIS"
		}
	}
	skor := max(0, min(100, 100-penalti))
	grade := "D"
	switch {
	case skor >= 90:
		grade = "A"
	case skor >= 80:
		grade = "B"
	case skor >= 60:
		grade = "C"
	}

	dimensi := make([]DQDimensi, 0, len(labelDimensi))
	for _, d := range labelDimensi {
		p := 0
		for _, a := range aturan {
			if a.Dimensi == d[0] && a.Terpicu {
				p += penaltiTetap[a.Severitas]
			}
		}
		dimensi = append(dimensi, DQDimensi{Kode: d[0], Label: d[1], Skor: max(0, 100-p)})
	}

	setKolom := map[string]bool{}
	for _, k := range ctx.KolomTersedia {
		setKolom[strings.ToUpper(k)] = true
	}
	status := func(penuh bool, lain string) string {
		if penuh {
			return "PENUH"
		}
		return lain
	}
	detail := setKolom["VISITATION_FEE"]
	matriks := []MatriksModul{
		{"Dashboard", "PENUH"},
		{"Claim Explorer", "PENUH"},
		{"Patient Tracer", "PENUH"},
		{"Financial Intelligence", status(detail, "TERBATAS")},
		{"Diagnostic Costs", status(setKolom["LAB"] && setKolom["RAD"], "TERBATAS")},
		{"Add-on Tarif", status(setKolom["KODE_ADDON"], "TIDAK_TERSEDIA")},
		{"DPJP Performance", status(setKolom["DPJPKODE"], "TERBATAS")},
		{"Claim Quality", "PENUH"},
		{"Anomaly Detection", "PENUH"},
	}

	gerbang := "LOLOS"
	if skor < ambangOverride || kritis {
		gerbang = "BLOKIR"
	}

	// Periode usulan = bulan terbanyak pada tanggal masuk.
	hitung := map[string]int{}
	for _, b := range baris {
		if scoring.TanggalValid(b.TglMasuk) {
			hitung[b.TglMasuk[:7]]++
		}
	}
	type kv struct {
		k string
		v int
	}
	var daftar []kv
	for k, v := range hitung {
		daftar = append(daftar, kv{k, v})
	}
	sort.Slice(daftar, func(i, j int) bool {
		if daftar[i].v != daftar[j].v {
			return daftar[i].v > daftar[j].v
		}
		return daftar[i].k < daftar[j].k
	})
	usulan := ctx.PeriodeDipilih
	if len(daftar) > 0 {
		usulan = daftar[0].k
	}

	return DQReport{
		Skor: skor, Grade: grade, Aturan: aturan, Dimensi: dimensi,
		MatriksModul: matriks, PeriodeUsulan: usulan, Gerbang: gerbang,
	}
}
