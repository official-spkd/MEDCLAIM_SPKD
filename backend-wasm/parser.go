// ============================================================
// MedClaim Go — Parser ekspor V-Claim (TSV) berbasis NAMA
// KOLOM + alias (port setia dari src/lib/medclaim/parser.ts)
// ============================================================
package main

import (
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var aliasKolom = map[string]string{
	"nosep": "NOSEP", "no.sep": "NOSEP",
	"nokartu": "NOKARTU", "no.kartu": "NOKARTU",
	"nomr": "NOMR", "no.mr": "NOMR",
	"namapasien": "NAMAPASIEN", "nama": "NAMAPASIEN",
	"tgllahir": "TGLLAHIR", "tanggal lahir": "TGLLAHIR",
	"jeniskelamin": "JENISKELAMIN", "gender": "JENISKELAMIN",
	"tglmasuk": "TGLMASUK", "tgl. masuk": "TGLMASUK", "tanggalmasuk": "TGLMASUK",
	"tglkeluar": "TGLKELUAR", "tgl. keluar": "TGLKELUAR", "tanggalkeluar": "TGLKELUAR",
	"los": "LOS", "length of stay": "LOS",
	"kelasrawat": "KELASRAWAT", "kelas": "KELASRAWAT",
	"tipelayanan": "TIPELAYANAN",
	"carapulang": "CARAPULANG",
	"dpjpkode": "DPJPKODE", "dpjp": "DPJPKODE",
	"dpjpnama": "DPJPNAMA",
	"diagnosa": "DIAGNOSA", "diaglist": "DIAGNOSA", "icd10": "DIAGNOSA",
	"prosedur": "PROSEDUR", "icd9": "PROSEDUR", "icd9cm": "PROSEDUR",
	"kodeicbg": "KODEICBG", "kodeinacbg": "KODEICBG", "inacbg": "KODEICBG",
	"deskripsiicbg": "DESKRIPSICBG", "deskripsiinacbg": "DESKRIPSICBG",
	"severitylevel": "SEVERITYLEVEL", "severity": "SEVERITYLEVEL", "lvl": "SEVERITYLEVEL",
	"idrgkode": "IDRGKODE", "idrg": "IDRGKODE", "idrgcode": "IDRGKODE",
	"kodeaddon": "KODE_ADDON", "addon": "KODE_ADDON",
	"tarifaddon": "TARIF_ADDON",
	"tarifinacbg": "TARIFINACBG", "tarif": "TARIFINACBG",
}

var aliasKomponen = map[string]string{
	"visitation_fee": "VISITATION_FEE",
	"accommodation_fee": "ACCOMODATION_FEE", "accomodation_fee": "ACCOMODATION_FEE",
	"icu_fee": "ICU_FEE",
	"icu_investigation": "ICU_SUPPORT_FEE", "icu_support_fee": "ICU_SUPPORT_FEE",
	"nurse_fee":        "NURSE_FEE",
	"treatment_fee":    "TREATMENT_FEE",
	"drug_fee":         "DRUG_FEE",
	"drug_kronis_fee":  "DRUG_KRONIS_FEE",
	"drug_kemo_fee":    "DRUG_KEMO_FEE",
	"obat_kronis":      "OBAT_KRONIS",
	"obat_kemo":        "OBAT_KEMO",
	"alkes_bhp":        "ALKES_BHP",
	"lab":              "LAB",
	"rad":              "RAD",
	"rehab_medik":      "REHAB_MEDIK",
	"rehab_medik_2":    "REHAB_MEDIK",
	"kamar_akomodasi":  "KAMAR_AKOMODASI",
	"operasional":      "OPERASIONAL",
	"darah":            "DARAH",
}

func normKolom(k string) string {
	key := strings.ToLower(strings.TrimSpace(k))
	key = strings.Join(strings.Fields(key), " ")
	if v, ok := aliasKolom[key]; ok {
		return v
	}
	if v, ok := aliasKomponen[key]; ok {
		return v
	}
	return strings.ToUpper(key)
}

func nilaiBersih(v string) string {
	s := strings.TrimSpace(v)
	up := strings.ToUpper(s)
	if s == "" || up == "NONE" || s == "-" || up == "NULL" {
		return ""
	}
	return s
}

// angkaBersih — tangani format id (1.234.567,89) dan en (1,234,567.89).
func angkaBersih(v string) float64 {
	raw := nilaiBersih(v)
	reBukanAngka := regexp.MustCompile(`[^0-9.,-]`)
	raw = reBukanAngka.ReplaceAllString(raw, "")
	var num float64
	var err error
	if strings.Contains(raw, ",") && strings.Contains(raw, ".") {
		if strings.LastIndex(raw, ",") > strings.LastIndex(raw, ".") {
			r := strings.ReplaceAll(raw, ".", "")
			r = strings.Replace(r, ",", ".", 1)
			num, err = strconv.ParseFloat(r, 64)
		} else {
			r := strings.ReplaceAll(raw, ",", "")
			num, err = strconv.ParseFloat(r, 64)
		}
	} else if strings.Contains(raw, ",") {
		reDesimal := regexp.MustCompile(`,\d{1,2}$`)
		if reDesimal.MatchString(raw) {
			r := strings.Replace(raw, ",", ".", 1)
			num, err = strconv.ParseFloat(r, 64)
		} else {
			r := strings.ReplaceAll(raw, ",", "")
			num, err = strconv.ParseFloat(r, 64)
		}
	} else if reRibuanTitik.MatchString(raw) {
		// pola ribuan id: 1.234.567 (tanpa desimal) — buang titik
		r := strings.ReplaceAll(raw, ".", "")
		num, err = strconv.ParseFloat(r, 64)
	} else {
		num, err = strconv.ParseFloat(raw, 64)
	}
	if err != nil || math.IsNaN(num) || math.IsInf(num, 0) {
		return 0
	}
	return num
}

var reIcdSplit = regexp.MustCompile(`[;,|]`)
var reRibuanTitik = regexp.MustCompile(`^\d{1,3}(\.\d{3})+$`)

func daftarKode(v string) []string {
	if nilaiBersih(v) == "" {
		return nil
	}
	parts := reIcdSplit.Split(v, -1)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.ToUpper(strings.TrimSpace(p))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// ---------- Tanggal ----------

const layoutISO = "2006-01-02"

var reTanggal = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func tanggalValid(s string) bool {
	if !reTanggal.MatchString(s) {
		return false
	}
	_, err := time.Parse(layoutISO, s)
	return err == nil
}

func selisihHari(masuk, keluar string) int {
	m, err1 := time.Parse(layoutISO, masuk)
	k, err2 := time.Parse(layoutISO, keluar)
	if err1 != nil || err2 != nil {
		return -1
	}
	return int(math.Round(k.Sub(m).Hours() / 24))
}

// ---------- Parse ----------

type BarisMentah struct {
	Kolom map[string]string
	Nomor int
}

type HasilParse struct {
	Header    []string
	Baris     []BarisMentah
	Delimiter string
}

func hitungKarakter(s, c string) int {
	return strings.Count(s, c)
}

func parseTSV(teks string) HasilParse {
	barisan := strings.SplitN(teks, "\n", 2)
	pertama := barisan[0]
	tab := hitungKarakter(pertama, "\t")
	semi := hitungKarakter(pertama, ";")
	koma := hitungKarakter(pertama, ",")
	delimiter := "\t"
	if semi > tab && semi >= koma {
		delimiter = ";"
	} else if koma > tab && koma > semi {
		delimiter = ","
	}

	lines := []string{}
	for _, l := range strings.Split(teks, "\n") {
		l = strings.TrimRight(l, "\r")
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return HasilParse{Delimiter: delimiter}
	}
	header := strings.Split(lines[0], delimiter)
	for i := range header {
		header[i] = strings.TrimSpace(header[i])
	}
	baris := []BarisMentah{}
	for i := 1; i < len(lines); i++ {
		sel := strings.Split(lines[i], delimiter)
		kolom := map[string]string{}
		for idx, h := range header {
			v := ""
			if idx < len(sel) {
				v = sel[idx]
			}
			kolom[normKolom(h)] = v
		}
		baris = append(baris, BarisMentah{Kolom: kolom, Nomor: i + 1})
	}
	return HasilParse{Header: header, Baris: baris, Delimiter: delimiter}
}

// ---------- Validasi keras baris ----------

type AddOn struct {
	Kode   string  `json:"kode"`
	Jenis  string  `json:"jenis"`
	Tarif  int64   `json:"tarif"`
	Status string  `json:"status"`
}

type BarisValid struct {
	NoSep         string            `json:"noSep"`
	NoKartu       string            `json:"noKartu"`
	NoMr          string            `json:"noMr"`
	NamaPasien    string            `json:"namaPasien"`
	TglLahir      string            `json:"tglLahir"`
	JenisKelamin  string            `json:"jenisKelamin"`
	TglMasuk      string            `json:"tglMasuk"`
	TglKeluar     string            `json:"tglKeluar"`
	Los           int               `json:"los"`
	KelasRawat    int               `json:"kelasRawat"`
	TipeLayanan   string            `json:"tipeLayanan"`
	CaraPulang    string            `json:"caraPulang,omitempty"`
	DpjpKode      string            `json:"dpjpKode,omitempty"`
	Diagnosa      []string          `json:"diagnosa"`
	Prosedur      []string          `json:"prosedur"`
	KodeIcbg      string            `json:"kodeIcbg"`
	DeskripsiIcbg string            `json:"deskripsiIcbg"`
	SeverityLevel int               `json:"severityLevel"`
	IdrgKode      string            `json:"idrgKode,omitempty"`
	Komponen      map[string]int64  `json:"komponen"`
	TotalBilling  int64             `json:"totalBilling"`
	TarifInacbg   int64             `json:"tarifInacbg"`
	AddOn         []AddOn           `json:"addOn,omitempty"`
}

type BarisGagal struct {
	Nomor  int    `json:"nomorBaris"`
	Alasan string `json:"alasan"`
}

type HasilValidasi struct {
	Valid         []BarisValid `json:"valid"`
	Gagal         []BarisGagal `json:"gagal"`
	KolomTersedia []string     `json:"kolomTersedia"`
}

func kolomAtau(kolom map[string]string, kunci, bawaan string) string {
	if v, ok := kolom[kunci]; ok {
		return v
	}
	return bawaan
}

func validasiBaris(p HasilParse) HasilValidasi {
	valid := []BarisValid{}
	gagal := []BarisGagal{}
	for _, b := range p.Baris {
		noSep := nilaiBersih(b.Kolom["NOSEP"])
		if noSep == "" {
			gagal = append(gagal, BarisGagal{b.Nomor, "NOSEP kosong — baris wajib dibuang"})
			continue
		}
		tglMasuk := nilaiBersih(b.Kolom["TGLMASUK"])
		tglKeluar := nilaiBersih(b.Kolom["TGLKELUAR"])
		if tglMasuk == "" || tglKeluar == "" {
			gagal = append(gagal, BarisGagal{b.Nomor, "SEP " + noSep + ": tanggal masuk/keluar wajib diisi"})
			continue
		}
		kodeIcbg := strings.ToUpper(nilaiBersih(b.Kolom["KODEICBG"]))
		if kodeIcbg == "" {
			gagal = append(gagal, BarisGagal{b.Nomor, "SEP " + noSep + ": kode INA-CBG wajib diisi"})
			continue
		}
		tarifInacbg := int64(angkaBersih(b.Kolom["TARIFINACBG"]))
		diagnosa := daftarKode(b.Kolom["DIAGNOSA"])
		if len(diagnosa) == 0 {
			gagal = append(gagal, BarisGagal{b.Nomor, "SEP " + noSep + ": minimal 1 kode diagnosis"})
			continue
		}

		komponen := map[string]int64{}
		var total int64
		for _, k := range KOMPONEN_BILLING {
			v := int64(angkaBersih(kolomAtau(b.Kolom, k, "0")))
			komponen[k] = v
			total += v
		}

		tipeRaw := strings.ToUpper(nilaiBersih(kolomAtau(b.Kolom, "TIPELAYANAN", "RANAP")))
		if tipeRaw == "" {
			tipeRaw = "RANAP"
		}
		tipeLayanan := "RANAP"
		if strings.Contains(tipeRaw, "RJ") || strings.Contains(tipeRaw, "JALAN") {
			tipeLayanan = "RJTP"
		}

		addOn := []AddOn{}
		if kodeAddon := nilaiBersih(b.Kolom["KODE_ADDON"]); kodeAddon != "" {
			jenis := "SP"
			reJenis := regexp.MustCompile(`^(SP|SR|SI|SD)`)
			if m := reJenis.FindStringSubmatch(strings.ToUpper(kodeAddon)); m != nil {
				jenis = m[1]
			}
			tarif := int64(angkaBersih(kolomAtau(b.Kolom, "TARIF_ADDON", "0")))
			status := "TERLEWAT"
			if tarif > 0 {
				status = "DIKLAIM"
			}
			addOn = append(addOn, AddOn{Kode: kodeAddon, Jenis: jenis, Tarif: tarif, Status: status})
		}

		losVal := angkaBersih(kolomAtau(b.Kolom, "LOS", "0"))
		if losVal == 0 {
			losVal = float64(selisihHari(tglMasuk, tglKeluar))
		}
		los := int(math.Max(0, math.Round(losVal)))
		if los == 0 {
			los = 1
		}
		kelasVal := angkaBersih(kolomAtau(b.Kolom, "KELASRAWAT", "1"))
		if kelasVal == 0 {
			kelasVal = 1
		}
		kelasRawat := int(math.Min(3, math.Max(1, math.Round(kelasVal))))
		sevVal := angkaBersih(kolomAtau(b.Kolom, "SEVERITYLEVEL", "1"))
		severity := int(math.Max(0, math.Min(3, math.Round(sevVal))))

		namaPasien := nilaiBersih(b.Kolom["NAMAPASIEN"])
		if namaPasien == "" {
			namaPasien = "(tanpa nama)"
		}
		jk := strings.ToUpper(nilaiBersih(kolomAtau(b.Kolom, "JENISKELAMIN", "L")))
		if jk == "" {
			jk = "L"
		}
		if strings.HasPrefix(jk, "P") {
			jk = "P"
		} else {
			jk = "L"
		}
		tglLahir := nilaiBersih(b.Kolom["TGLLAHIR"])
		if tglLahir == "" {
			tglLahir = "1900-01-01"
		}
		deskripsi := nilaiBersih(b.Kolom["DESKRIPSICBG"])
		if deskripsi == "" {
			deskripsi = kodeIcbg
		}
		caraPulang := nilaiBersih(b.Kolom["CARAPULANG"])
		dpjp := nilaiBersih(b.Kolom["DPJPKODE"])
		idrg := nilaiBersih(b.Kolom["IDRGKODE"])
		noKartu := nilaiBersih(b.Kolom["NOKARTU"])
		if noKartu == "" {
			noKartu = "-"
		}
		noMr := nilaiBersih(b.Kolom["NOMR"])
		if noMr == "" {
			noMr = "-"
		}

		valid = append(valid, BarisValid{
			NoSep: noSep, NoKartu: noKartu, NoMr: noMr, NamaPasien: namaPasien,
			TglLahir: tglLahir, JenisKelamin: jk, TglMasuk: tglMasuk, TglKeluar: tglKeluar,
			Los: los, KelasRawat: kelasRawat, TipeLayanan: tipeLayanan,
			CaraPulang: caraPulang, DpjpKode: dpjp,
			Diagnosa: diagnosa, Prosedur: daftarKode(b.Kolom["PROSEDUR"]),
			KodeIcbg: kodeIcbg, DeskripsiIcbg: deskripsi, SeverityLevel: severity,
			IdrgKode: idrg, Komponen: komponen, TotalBilling: total, TarifInacbg: tarifInacbg,
			AddOn: addOn,
		})
	}

	set := map[string]bool{}
	kolomTersedia := []string{}
	for _, b := range p.Baris {
		for k := range b.Kolom {
			if k != "" && !set[k] {
				set[k] = true
				kolomTersedia = append(kolomTersedia, k)
			}
		}
	}
	sort.Strings(kolomTersedia)
	return HasilValidasi{Valid: valid, Gagal: gagal, KolomTersedia: kolomTersedia}
}

// ---------- Row hash (versioning) ----------

func barisKanonik(b BarisValid) string {
	diag := append([]string(nil), b.Diagnosa...)
	sort.Strings(diag)
	pros := append([]string(nil), b.Prosedur...)
	sort.Strings(pros)
	addon := make([]string, 0, len(b.AddOn))
	sorted := append([]AddOn(nil), b.AddOn...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Kode < sorted[j].Kode })
	for _, a := range sorted {
		addon = append(addon, a.Kode+":"+a.Status+":"+strconv.FormatInt(a.Tarif, 10))
	}
	komp := make([]int64, 0, len(KOMPONEN_BILLING))
	for _, k := range KOMPONEN_BILLING {
		komp = append(komp, b.Komponen[k])
	}
	parts := []string{
		`{"sep":"` + b.NoSep + `"`,
		`"kartu":"` + b.NoKartu + `"`,
		`"mr":"` + b.NoMr + `"`,
		`"nama":"` + b.NamaPasien + `"`,
		`"lahir":"` + b.TglLahir + `"`,
		`"jk":"` + b.JenisKelamin + `"`,
		`"masuk":"` + b.TglMasuk + `"`,
		`"keluar":"` + b.TglKeluar + `"`,
		`"los":` + strconv.Itoa(b.Los),
		`"kelas":` + strconv.Itoa(b.KelasRawat),
		`"tipe":"` + b.TipeLayanan + `"`,
		`"pulang":"` + b.CaraPulang + `"`,
		`"dpjp":"` + b.DpjpKode + `"`,
		`"diag":[` + strings.Join(quoteSemua(diag), ",") + `]`,
		`"pros":[` + strings.Join(quoteSemua(pros), ",") + `]`,
		`"icbg":"` + b.KodeIcbg + `"`,
		`"desk":"` + b.DeskripsiIcbg + `"`,
		`"sev":` + strconv.Itoa(b.SeverityLevel),
		`"idrg":"` + b.IdrgKode + `"`,
		`"tarif":` + strconv.FormatInt(b.TarifInacbg, 10),
		`"komp":[` + int64Join(komp) + `]`,
		`"addon":[` + strings.Join(quoteSemua(addon), ",") + `]}`,
	}
	return strings.Join(parts, ",")
}

func quoteSemua(arr []string) []string {
	out := make([]string, len(arr))
	for i, s := range arr {
		out[i] = `"` + s + `"`
	}
	return out
}

func int64Join(arr []int64) string {
	out := make([]string, len(arr))
	for i, v := range arr {
		out[i] = strconv.FormatInt(v, 10)
	}
	return strings.Join(out, ",")
}

// rowHash — FNV-1a 32-bit ganda atas bentuk kanonik + salt (hospital).
func rowHash(b BarisValid, salt string) string {
	teks := salt + "|" + barisKanonik(b)
	h1 := fnv1a(teks, false)
	h2 := fnv1a(teks, true)
	return h1 + h2
}

func fnv1a(teks string, terbalik bool) string {
	const (
		offset uint32 = 0x811c9dc5
		prime  uint32 = 0x01000193
	)
	h := offset
	if terbalik {
		h = offset ^ uint32(len(teks))
	}
	n := len(teks)
	for i := 0; i < n; i++ {
		var c byte
		if terbalik {
			c = teks[n-1-i]
		} else {
			c = teks[i]
		}
		h ^= uint32(c)
		h *= prime
	}
	return pad8(strconv.FormatUint(uint64(h), 16))
}

func pad8(s string) string {
	for len(s) < 8 {
		s = "0" + s
	}
	return s
}
