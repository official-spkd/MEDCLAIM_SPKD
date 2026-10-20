package uploads

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"medclaim-backend/internal/scoring"
)

// Parser ekspor V-Claim (TSV/CSV) berbasis NAMA KOLOM + alias, sehingga urutan
// kolom bebas dan variasi penamaan antar-SIMRS dapat dipetakan.

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
	"carapulang":  "CARAPULANG",
	"dpjpkode":    "DPJPKODE", "dpjp": "DPJPKODE",
	"dpjpnama": "DPJPNAMA",
	"diagnosa": "DIAGNOSA", "diaglist": "DIAGNOSA", "icd10": "DIAGNOSA",
	"prosedur": "PROSEDUR", "icd9": "PROSEDUR", "icd9cm": "PROSEDUR",
	"kodeicbg": "KODEICBG", "kodeinacbg": "KODEICBG", "inacbg": "KODEICBG",
	"deskripsiicbg": "DESKRIPSICBG", "deskripsiinacbg": "DESKRIPSICBG",
	"severitylevel": "SEVERITYLEVEL", "severity": "SEVERITYLEVEL", "lvl": "SEVERITYLEVEL",
	"idrgkode": "IDRGKODE", "idrg": "IDRGKODE", "idrgcode": "IDRGKODE",
	"kodeaddon": "KODE_ADDON", "addon": "KODE_ADDON",
	"tarifaddon":  "TARIF_ADDON",
	"tarifinacbg": "TARIFINACBG", "tarif": "TARIFINACBG",
}

var aliasKomponen = map[string]string{
	"visitation_fee":    "VISITATION_FEE",
	"accommodation_fee": "ACCOMODATION_FEE", "accomodation_fee": "ACCOMODATION_FEE",
	"icu_fee":           "ICU_FEE",
	"icu_investigation": "ICU_SUPPORT_FEE", "icu_support_fee": "ICU_SUPPORT_FEE",
	"nurse_fee":       "NURSE_FEE",
	"treatment_fee":   "TREATMENT_FEE",
	"drug_fee":        "DRUG_FEE",
	"drug_kronis_fee": "DRUG_KRONIS_FEE",
	"drug_kemo_fee":   "DRUG_KEMO_FEE",
	"obat_kronis":     "OBAT_KRONIS",
	"obat_kemo":       "OBAT_KEMO",
	"alkes_bhp":       "ALKES_BHP",
	"lab":             "LAB",
	"rad":             "RAD",
	"rehab_medik":     "REHAB_MEDIK",
	"rehab_medik_2":   "REHAB_MEDIK",
	"kamar_akomodasi": "KAMAR_AKOMODASI",
	"operasional":     "OPERASIONAL",
	"darah":           "DARAH",
}

var (
	reBukanAngka = regexp.MustCompile(`[^0-9.,-]`)
	reDesimalId  = regexp.MustCompile(`,\d{1,2}$`)
	reRibuanId   = regexp.MustCompile(`^\d{1,3}(\.\d{3})+$`)
	reIcdSplit   = regexp.MustCompile(`[;,|]`)
	reJenisAddon = regexp.MustCompile(`^(SP|SR|SI|SD)`)
)

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

// angkaBersih menangani format Indonesia (1.234.567,89) dan Inggris (1,234,567.89).
func angkaBersih(v string) float64 {
	raw := reBukanAngka.ReplaceAllString(nilaiBersih(v), "")
	var num float64
	var err error
	switch {
	case strings.Contains(raw, ",") && strings.Contains(raw, "."):
		if strings.LastIndex(raw, ",") > strings.LastIndex(raw, ".") {
			num, err = strconv.ParseFloat(strings.Replace(strings.ReplaceAll(raw, ".", ""), ",", ".", 1), 64)
		} else {
			num, err = strconv.ParseFloat(strings.ReplaceAll(raw, ",", ""), 64)
		}
	case strings.Contains(raw, ","):
		if reDesimalId.MatchString(raw) {
			num, err = strconv.ParseFloat(strings.Replace(raw, ",", ".", 1), 64)
		} else {
			num, err = strconv.ParseFloat(strings.ReplaceAll(raw, ",", ""), 64)
		}
	case reRibuanId.MatchString(raw):
		num, err = strconv.ParseFloat(strings.ReplaceAll(raw, ".", ""), 64)
	default:
		num, err = strconv.ParseFloat(raw, 64)
	}
	if err != nil || math.IsNaN(num) || math.IsInf(num, 0) {
		return 0
	}
	return num
}

func daftarKode(v string) []string {
	if nilaiBersih(v) == "" {
		return nil
	}
	out := []string{}
	for _, p := range reIcdSplit.Split(v, -1) {
		if s := strings.ToUpper(strings.TrimSpace(p)); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// ---------- Parse ----------

type barisMentah struct {
	Kolom map[string]string
	Nomor int
}

type hasilParse struct {
	Header    []string
	Baris     []barisMentah
	Delimiter string
}

func parseTSV(teks string) hasilParse {
	teks = strings.TrimPrefix(teks, "\xef\xbb\xbf") // BOM UTF-8 dari Excel
	pertama := strings.SplitN(teks, "\n", 2)[0]
	tab, semi, koma := strings.Count(pertama, "\t"), strings.Count(pertama, ";"), strings.Count(pertama, ",")
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
		return hasilParse{Delimiter: delimiter}
	}
	header := strings.Split(lines[0], delimiter)
	for i := range header {
		header[i] = strings.TrimSpace(header[i])
	}
	baris := make([]barisMentah, 0, len(lines)-1)
	for i := 1; i < len(lines); i++ {
		sel := strings.Split(lines[i], delimiter)
		kolom := make(map[string]string, len(header))
		for idx, h := range header {
			v := ""
			if idx < len(sel) {
				v = sel[idx]
			}
			kolom[normKolom(h)] = v
		}
		baris = append(baris, barisMentah{Kolom: kolom, Nomor: i + 1})
	}
	return hasilParse{Header: header, Baris: baris, Delimiter: delimiter}
}

// ---------- Validasi keras per baris ----------

type AddOn struct {
	Kode   string `json:"kode"`
	Jenis  string `json:"jenis"`
	Tarif  int64  `json:"tarif"`
	Status string `json:"status"`
}

type BarisValid struct {
	Nomor         int // nomor baris pada file (untuk rincian laporan DQ)
	NoSep         string
	NoKartu       string
	NoMr          string
	NamaPasien    string
	TglLahir      string
	JenisKelamin  string
	TglMasuk      string
	TglKeluar     string
	Los           int
	KelasRawat    int
	TipeLayanan   string
	CaraPulang    string
	DpjpKode      string
	DpjpNama      string
	Diagnosa      []string
	Prosedur      []string
	KodeIcbg      string
	DeskripsiIcbg string
	SeverityLevel int
	IdrgKode      string
	Komponen      map[string]int64
	TotalBilling  int64
	TarifInacbg   int64
	AddOn         []AddOn
}

type BarisGagal struct {
	Nomor  int    `json:"nomorBaris"`
	Alasan string `json:"alasan"`
}

type hasilValidasi struct {
	Valid         []BarisValid
	Gagal         []BarisGagal
	KolomTersedia []string
}

func kolomAtau(kolom map[string]string, kunci, bawaan string) string {
	if v, ok := kolom[kunci]; ok {
		return v
	}
	return bawaan
}

func validasiBaris(p hasilParse) hasilValidasi {
	valid := []BarisValid{}
	gagal := []BarisGagal{}
	for _, b := range p.Baris {
		noSep := nilaiBersih(b.Kolom["NOSEP"])
		if noSep == "" {
			gagal = append(gagal, BarisGagal{b.Nomor, "NOSEP kosong — baris wajib dibuang"})
			continue
		}
		tglMasuk, tglKeluar := nilaiBersih(b.Kolom["TGLMASUK"]), nilaiBersih(b.Kolom["TGLKELUAR"])
		if tglMasuk == "" || tglKeluar == "" {
			gagal = append(gagal, BarisGagal{b.Nomor, "SEP " + noSep + ": tanggal masuk/keluar wajib diisi"})
			continue
		}
		kodeIcbg := strings.ToUpper(nilaiBersih(b.Kolom["KODEICBG"]))
		if kodeIcbg == "" {
			gagal = append(gagal, BarisGagal{b.Nomor, "SEP " + noSep + ": kode INA-CBG wajib diisi"})
			continue
		}
		diagnosa := daftarKode(b.Kolom["DIAGNOSA"])
		if len(diagnosa) == 0 {
			gagal = append(gagal, BarisGagal{b.Nomor, "SEP " + noSep + ": minimal 1 kode diagnosis"})
			continue
		}

		komponen := map[string]int64{}
		var total int64
		for _, k := range scoring.KomponenBilling {
			v := int64(angkaBersih(kolomAtau(b.Kolom, k, "0")))
			komponen[k] = v
			total += v
		}

		tipeRaw := strings.ToUpper(nilaiBersih(kolomAtau(b.Kolom, "TIPELAYANAN", "RANAP")))
		tipe := "RANAP"
		if strings.Contains(tipeRaw, "RJ") || strings.Contains(tipeRaw, "JALAN") {
			tipe = "RJTP"
		}

		addOn := []AddOn{}
		if kode := nilaiBersih(b.Kolom["KODE_ADDON"]); kode != "" {
			jenis := "SP"
			if m := reJenisAddon.FindStringSubmatch(strings.ToUpper(kode)); m != nil {
				jenis = m[1]
			}
			tarif := int64(angkaBersih(kolomAtau(b.Kolom, "TARIF_ADDON", "0")))
			status := "TERLEWAT"
			if tarif > 0 {
				status = "DIKLAIM"
			}
			addOn = append(addOn, AddOn{Kode: kode, Jenis: jenis, Tarif: tarif, Status: status})
		}

		losVal := angkaBersih(kolomAtau(b.Kolom, "LOS", "0"))
		if losVal == 0 {
			losVal = float64(scoring.SelisihHari(tglMasuk, tglKeluar))
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
		severity := int(math.Max(0, math.Min(3, math.Round(angkaBersih(kolomAtau(b.Kolom, "SEVERITYLEVEL", "1"))))))

		namaPasien := nilaiBersih(b.Kolom["NAMAPASIEN"])
		if namaPasien == "" {
			namaPasien = "(tanpa nama)"
		}
		jk := "L"
		if strings.HasPrefix(strings.ToUpper(nilaiBersih(kolomAtau(b.Kolom, "JENISKELAMIN", "L"))), "P") {
			jk = "P"
		}
		tglLahir := nilaiBersih(b.Kolom["TGLLAHIR"])
		if tglLahir == "" {
			tglLahir = "1900-01-01"
		}
		deskripsi := nilaiBersih(b.Kolom["DESKRIPSICBG"])
		if deskripsi == "" {
			deskripsi = kodeIcbg
		}
		noKartu := nilaiBersih(b.Kolom["NOKARTU"])
		if noKartu == "" {
			noKartu = "-"
		}
		noMr := nilaiBersih(b.Kolom["NOMR"])
		if noMr == "" {
			noMr = "-"
		}

		valid = append(valid, BarisValid{
			Nomor: b.Nomor, NoSep: noSep, NoKartu: noKartu, NoMr: noMr, NamaPasien: namaPasien,
			TglLahir: tglLahir, JenisKelamin: jk, TglMasuk: tglMasuk, TglKeluar: tglKeluar,
			Los: los, KelasRawat: kelasRawat, TipeLayanan: tipe,
			CaraPulang: nilaiBersih(b.Kolom["CARAPULANG"]), DpjpKode: nilaiBersih(b.Kolom["DPJPKODE"]),
			DpjpNama: nilaiBersih(b.Kolom["DPJPNAMA"]),
			Diagnosa: diagnosa, Prosedur: daftarKode(b.Kolom["PROSEDUR"]),
			KodeIcbg: kodeIcbg, DeskripsiIcbg: deskripsi, SeverityLevel: severity,
			IdrgKode: nilaiBersih(b.Kolom["IDRGKODE"]), Komponen: komponen, TotalBilling: total,
			TarifInacbg: int64(angkaBersih(b.Kolom["TARIFINACBG"])), AddOn: addOn,
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
	return hasilValidasi{Valid: valid, Gagal: gagal, KolomTersedia: kolomTersedia}
}

// ---------- Row hash (deteksi perubahan antar upload) ----------

func (b BarisValid) kanonik() string {
	diag := append([]string(nil), b.Diagnosa...)
	pros := append([]string(nil), b.Prosedur...)
	sort.Strings(diag)
	sort.Strings(pros)
	addon := make([]string, 0, len(b.AddOn))
	sorted := append([]AddOn(nil), b.AddOn...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Kode < sorted[j].Kode })
	for _, a := range sorted {
		addon = append(addon, a.Kode+":"+a.Status+":"+strconv.FormatInt(a.Tarif, 10))
	}
	komp := make([]string, 0, len(scoring.KomponenBilling))
	for _, k := range scoring.KomponenBilling {
		komp = append(komp, strconv.FormatInt(b.Komponen[k], 10))
	}
	return strings.Join([]string{
		b.NoSep, b.NoKartu, b.NoMr, b.NamaPasien, b.TglLahir, b.JenisKelamin,
		b.TglMasuk, b.TglKeluar, strconv.Itoa(b.Los), strconv.Itoa(b.KelasRawat),
		b.TipeLayanan, b.CaraPulang, b.DpjpKode, strings.Join(diag, ","),
		strings.Join(pros, ","), b.KodeIcbg, b.DeskripsiIcbg, strconv.Itoa(b.SeverityLevel),
		b.IdrgKode, strconv.FormatInt(b.TarifInacbg, 10), strings.Join(komp, ","),
		strings.Join(addon, ","),
	}, "\x1f")
}

// rowHash = SHA-256 atas bentuk kanonik + salt tenant.
func rowHash(b BarisValid, salt string) string {
	sum := sha256.Sum256([]byte(salt + "\x1e" + b.kanonik()))
	return hex.EncodeToString(sum[:])
}
