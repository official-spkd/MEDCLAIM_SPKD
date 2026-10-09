// ============================================================
// MedClaim Go — Model domain, penyimpanan JSON, & seed demo
// (port dari types.ts + prisma/schema.prisma + scripts/seed.ts)
// ============================================================
package main

import (
        "encoding/json"
        "log"
        "fmt"
        "math/rand"
        "os"
        "path/filepath"
        "sort"
        "strconv"
        "strings"
        "sync"
        "time"
)

// ---------- Model ----------

type Hospital struct {
        ID   string `json:"id"`
        Kode string `json:"kode"`
        Nama string `json:"nama"`
        Kota string `json:"kota"`
        Kelas string `json:"kelas"`
}

type User struct {
        ID           string  `json:"id"`
        Email        string  `json:"email"`
        PasswordHash string  `json:"passwordHash"`
        Nama         string  `json:"nama"`
        Role         string  `json:"role"`
        HospitalID   *string `json:"hospitalId"`
        Spesialisasi *string `json:"spesialisasi"`
        DpjpKode     *string `json:"dpjpKode"`
        Aktif        bool    `json:"aktif"`
        CreatedAt    string  `json:"createdAt"`
        LastLoginAt  *string `json:"lastLoginAt"`
}

type Sesi struct {
        ID           string  `json:"id"`
        Email        string  `json:"email"`
        Nama         string  `json:"nama"`
        Role         string  `json:"role"`
        HospitalID   *string `json:"hospitalId"`
        HospitalNama *string `json:"hospitalNama"`
        HospitalKode *string `json:"hospitalKode"`
        Spesialisasi *string `json:"spesialisasi"`
        DpjpKode     *string `json:"dpjpKode"`
}

type UploadBatch struct {
        ID            string  `json:"id"`
        HospitalID    string  `json:"hospitalId"`
        NamaFile      string  `json:"namaFile"`
        Periode       string  `json:"periode"`
        Format        string  `json:"format"`
        DQScore       int     `json:"dqScore"`
        DQGrade       string  `json:"dqGrade"`
        DQReport      *DQReport `json:"dqReport,omitempty"`
        Status        string  `json:"status"` // PROSES | SELESAI | ARSIP | DIHAPUS
        TotalBaris    int     `json:"totalBaris"`
        BarisValid    int     `json:"barisValid"`
        BarisGagal    int     `json:"barisGagal"`
        BarisBaru     int     `json:"barisBaru"`
        BarisBerubah  int     `json:"barisBerubah"`
        BarisUnchanged int    `json:"barisUnchanged"`
        OverrideUsed  bool    `json:"overrideUsed"`
        AlasanOverride *string `json:"alasanOverride,omitempty"`
        UploadedByID  string  `json:"uploadedById"`
        CreatedAt     string  `json:"createdAt"`
}

type DiagnosisRef struct {
        Kode string `json:"kode"`
        Nama string `json:"nama"`
}

type ProsedurRef struct {
        Kode string `json:"kode"`
        Nama string `json:"nama"`
}

type Episode struct {
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
        Diagnosa      []DiagnosisRef   `json:"diagnosa"`
        Prosedur      []ProsedurRef    `json:"prosedur"`
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
        AddOn         []AddOn          `json:"addOn"`
        DQScore       int              `json:"dqScore"`
        RowHash       string           `json:"rowHash"`
        VersionNo     int              `json:"versionNo"`
        Superseded    bool             `json:"superseded"`
        CreatedAt     string           `json:"createdAt"`
}

type LedgerEntry struct {
        ID          string  `json:"id"`
        EpisodeID   string  `json:"episodeId"`
        HospitalID  string  `json:"hospitalId"`
        Sep         string  `json:"sep"`
        NomorMr     string  `json:"nomorMr"`
        NamaPasien  string  `json:"namaPasien"`
        KodeIcbg    string  `json:"kodeIcbg"`
        KelasRugi   string  `json:"kelasRugi"`
        Estimasi    int64   `json:"estimasi"`
        Realisasi   *int64  `json:"realisasi"`
        Status      string  `json:"status"`
        ReferensiRm *string `json:"referensiRm"`
        Catatan     *string `json:"catatan"`
        Strategy    *string `json:"strategy"`
        CreatedAt   string  `json:"createdAt"`
        UpdatedAt   string  `json:"updatedAt"`
}

type LedgerEvent struct {
        ID         string  `json:"id"`
        LedgerID   string  `json:"ledgerId"`
        DariStatus *string `json:"dariStatus"`
        KeStatus   *string `json:"keStatus"`
        AktorNama  string  `json:"aktorNama"`
        Catatan    *string `json:"catatan"`
        Waktu      string  `json:"waktu"`
}

type QueryDok struct {
        ID               string  `json:"id"`
        HospitalID       string  `json:"hospitalId"`
        Sep              string  `json:"sep"`
        NomorMr          string  `json:"nomorMr"`
        Inisial          string  `json:"inisial"`
        TglMasuk         *string `json:"tglMasuk"`
        TglKeluar        *string `json:"tglKeluar"`
        DpjpID           string  `json:"dpjpId"`
        DpjpNama         string  `json:"dpjpNama"`
        DpjpSpesialisasi *string `json:"dpjpSpesialisasi"`
        TemplateNama     *string `json:"templateNama"`
        Pertanyaan       string  `json:"pertanyaan"`
        Konteks          string  `json:"konteks"`
        Status           string  `json:"status"` // DRAFT | TERKIRIM | DIJAWAB | TERLAMBAT
        Deadline         string  `json:"deadline"`
        SentAt           *string `json:"sentAt"`
        AnsweredAt       *string `json:"answeredAt"`
        Jawaban          *string `json:"jawaban"`
        SlaHari          int     `json:"slaHari"`
        CreatedAt        string  `json:"createdAt"`
}

type TemplateDok struct {
        ID       string `json:"id"`
        Nama     string `json:"nama"`
        Isi      string `json:"isi"`
        Kategori string `json:"kategori"`
        Aktif    bool   `json:"aktif"`
}

type Setting struct {
        Key   string `json:"key"`
        Value string `json:"value"`
}

type AuditLog struct {
        ID       string  `json:"id"`
        UserEmail string `json:"userEmail"`
        Aksi     string  `json:"aksi"`
        Entitas  string  `json:"entitas"`
        EntitasID *string `json:"entitasId"`
        Meta     *string `json:"meta"`
        Waktu    string  `json:"waktu"`
}

// ---------- Penyimpanan ----------

type Store struct {
        mu          sync.RWMutex
        file        string
        Hospitals   []*Hospital      `json:"hospitals"`
        Users       []*User          `json:"users"`
        Uploads     []*UploadBatch   `json:"uploads"`
        Episodes    []*Episode       `json:"episodes"`
        Ledger      []*LedgerEntry   `json:"ledger"`
        LedgerEvents []*LedgerEvent  `json:"ledgerEvents"`
        Queries     []*QueryDok      `json:"queries"`
        Templates   []*TemplateDok   `json:"templates"`
        Settings    []*Setting       `json:"settings"`
        Audit       []*AuditLog      `json:"audit"`
}

var store = &Store{}

func (s *Store) simpan() {
        s.mu.RLock()
        data, err := json.Marshal(s)
        s.mu.RUnlock()
        if err != nil {
                return
        }
        tmp := s.file + ".tmp"
        if err := os.WriteFile(tmp, data, 0644); err != nil {
                return
        }
        os.Rename(tmp, s.file)
}

func (s *Store) muat() bool {
        data, err := bacaFile(s.file)
        if err != nil {
                return false
        }
        if err := json.Unmarshal(data, s); err != nil {
                return false
        }
        return true
}

func (s *Store) cariUser(id string) *User {
        for _, u := range s.Users {
                if u.ID == id {
                        return u
                }
        }
        return nil
}

func (s *Store) cariHospital(id string) *Hospital {
        for _, h := range s.Hospitals {
                if h.ID == id {
                        return h
                }
        }
        return nil
}

func (s *Store) catatAudit(email, aksi, entitas string, entitasID, meta *string) {
        s.Audit = append(s.Audit, &AuditLog{
                ID: idBaru("aud"), UserEmail: email, Aksi: aksi, Entitas: entitas,
                EntitasID: entitasID, Meta: meta, Waktu: sekarangISO(),
        })
}

func sekarangISO() string {
        return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

func buatSesi(u *User) *Sesi {
        s := &Sesi{
                ID: u.ID, Email: u.Email, Nama: u.Nama, Role: u.Role,
                Spesialisasi: u.Spesialisasi, DpjpKode: u.DpjpKode,
        }
        if u.HospitalID != nil {
                s.HospitalID = u.HospitalID
                if h := store.cariHospital(*u.HospitalID); h != nil {
                        s.HospitalNama = &h.Nama
                        s.HospitalKode = &h.Kode
                }
        }
        return s
}

// ---------- Seed ----------

var katalogCBG = []struct {
        Kode, Desk string
        Tarif      int64
}{
        {"Q-5-14", "Septicemia", 6_200_000},
        {"I-4-08", "Gastritis & duodenitis", 2_450_000},
        {"K-4-06", "Appendektomi", 5_100_000},
        {"J-4-07", "Pneumonia", 4_050_000},
        {"N-4-09", "Hemodialisa", 3_150_000},
        {"O-2-02", "Sectio caesarea", 4_600_000},
        {"E-4-02", "Diabetes mellitus", 3_300_000},
        {"C-4-30", "Neoplasma maligna", 8_400_000},
        {"S-4-05", "Fiksasi fraktur", 4_950_000},
        {"A-4-06", "Gastroenteritis & dehidrasi", 2_050_000},
        {"I-4-09", "Gagal jantung", 4_700_000},
        {"RJ-2-09", "Kunjungan lanjutan RJTP", 480_000},
}

var sevMult = map[int]float64{1: 1.0, 2: 1.38, 3: 1.85}

var dxUmum = []DiagnosisRef{
        {"K29.7", "Gastritis"}, {"K35.8", "Apendisitis akut"}, {"J18.9", "Pneumonia"},
        {"E11.9", "DM tipe 2"}, {"N18.5", "CKD stage 5"}, {"C34.90", "Kanker paru"},
        {"I50.9", "Gagal jantung"}, {"A41.9", "Sepsis"}, {"A09", "Gastroenteritis"},
        {"S72.00", "Fraktur femur"}, {"O82", "Persalinan SC"}, {"D64.9", "Anemia"},
        {"I10", "Hipertensi"}, {"J44.1", "COPD eksaserbasi"}, {"K80.20", "Kolelitiase"},
}

var dxCC = []DiagnosisRef{
        {"I50.0", "Gagal jantung kongestif"}, {"N17.9", "Gagal ginjal akut"},
        {"J96.01", "Gagal napas"}, {"E87.6", "Hipokalemia"}, {"D69.6", "Trombositopenia"},
        {"A41.51", "Sepsis gram negatif"}, {"I21.09", "Infark miokard"},
}

var katalogProsedur = []ProsedurRef{
        {"47.09", "Apendektomi"}, {"48.29", "Kolostomi"}, {"74.1", "Sectio caesarea inferiornya"},
        {"39.95", "Hemodialisa"}, {"79.35", "Fiksasi internal fraktur"}, {"96.04", "Intubasi"},
        {"88.72", "Ekokardiografi"}, {"51.23", "Kolesistektomi"},
}

var namaDepan = []string{"Andi", "Budi", "Citra", "Dewi", "Eko", "Fitri", "Gunawan", "Hesti", "Irfan", "Joko", "Kartika", "Lestari", "Mulyadi", "Nur", "Oscar", "Putri", "Rahmat", "Sinta", "Taufik", "Umar", "Vina", "Wahyu", "Yanti", "Zainal", "Bagas", "Candra", "Dian", "Erlangga"}
var namaBelakang = []string{"Santoso", "Wijaya", "Kusuma", "Halim", "Nugroho", "Saputra", "Permata", "Anggraini", "Pratama", "Utami", "Siregar", "Maulana", "Hidayat", "Puspita", "Firmansyah"}

var katalogDPJP = []struct {
        Nama, Kode, Spes string
}{
        {"dr. Budi Prakoso, Sp.PD", "DPJP-0101", "Penyakit Dalam"},
        {"dr. Sinta Maharani, Sp.B", "DPJP-0102", "Bedah"},
        {"dr. Rahmat Hidayat, Sp.An", "DPJP-0103", "Anestesi"},
        {"dr. Yanti Lestari, Sp.OG", "DPJP-0104", "Obstetri Ginekologi"},
        {"dr. Eko Prasetyo, Sp.P", "DPJP-0105", "Paru"},
        {"dr. Fitri Utami, Sp.A", "DPJP-0106", "Anak"},
}

var komponenPool = []string{"VISITATION_FEE", "ACCOMODATION_FEE", "NURSE_FEE", "TREATMENT_FEE", "DRUG_FEE", "ALKES_BHP", "LAB", "RAD", "KAMAR_AKOMODASI", "OPERASIONAL", "OBAT_KRONIS", "OBAT_KEMO", "DRUG_KRONIS_FEE", "DRUG_KEMO_FEE", "ICU_FEE", "ICU_SUPPORT_FEE", "REHAB_MEDIK", "DARAH"}

func acakPilih[T any](rnd *rand.Rand, arr []T) T {
        return arr[rnd.Intn(len(arr))]
}

func tanggalAcak(rnd *rand.Rand, ym string) string {
        var y, m int
        fmt.Sscanf(ym, "%d-%d", &y, &m)
        hari := 1 + rnd.Intn(28)
        return fmt.Sprintf("%04d-%02d-%02d", y, m, hari)
}

func tambahHari(iso string, hari int) string {
        t, err := time.Parse(layoutISO, iso)
        if err != nil {
                return iso
        }
        return t.AddDate(0, 0, hari).Format(layoutISO)
}

func komponenBillingAcak(rnd *rand.Rand, total int64, kemo, kronis bool, lab, rad int64) map[string]int64 {
        k := map[string]int64{}
        for _, key := range komponenPool {
                k[key] = 0
        }
        k["VISITATION_FEE"] = 200_000 + int64(rnd.Intn(120))*1000
        k["ACCOMODATION_FEE"] = int64(float64(total) * (0.22 + rnd.Float64()*0.12))
        k["NURSE_FEE"] = int64(float64(total) * (0.10 + rnd.Float64()*0.08))
        k["TREATMENT_FEE"] = int64(float64(total) * (0.12 + rnd.Float64()*0.10))
        k["DRUG_FEE"] = int64(float64(total) * (0.12 + rnd.Float64()*0.12))
        k["ALKES_BHP"] = int64(float64(total) * (0.10 + rnd.Float64()*0.10))
        if lab > 0 {
                k["LAB"] = lab
        } else {
                k["LAB"] = int64(float64(total) * (0.05 + rnd.Float64()*0.07))
        }
        if rad > 0 {
                k["RAD"] = rad
        } else {
                k["RAD"] = int64(float64(total) * (0.03 + rnd.Float64()*0.05))
        }
        k["KAMAR_AKOMODASI"] = int64(float64(total) * 0.05)
        if rnd.Float64() < 0.25 {
                k["OPERASIONAL"] = int64(float64(total) * 0.18)
        }
        if kemo {
                k["OBAT_KEMO"] = int64(float64(total) * 0.14)
                k["DRUG_KEMO_FEE"] = int64(float64(total) * 0.06)
        }
        if kronis {
                k["OBAT_KRONIS"] = int64(float64(total) * 0.08)
                k["DRUG_KRONIS_FEE"] = int64(float64(total) * 0.05)
        }
        return k
}

// generator episode sintetis untuk RS002/RS003 → TSV format DETAIL/BASIC.
func buatTSVSintetis(rnd *rand.Rand, nBaris int, periode string, format string, sepAwal, mrAwal int, salt string) (string, int, int) {
        var header []string
        if format == "BASIC" {
                header = []string{"NOSEP", "NOKARTU", "NOMR", "NAMAPASIEN", "TGLLAHIR", "JENISKELAMIN", "TGLMASUK", "TGLKELUAR", "LOS", "KELASRAWAT", "TIPELAYANAN", "DIAGNOSA", "PROSEDUR", "KODEICBG", "DESKRIPSICBG", "SEVERITYLEVEL", "TARIFINACBG"}
        } else {
                header = []string{"NOSEP", "NOKARTU", "NOMR", "NAMAPASIEN", "TGLLAHIR", "JENISKELAMIN", "TGLMASUK", "TGLKELUAR", "LOS", "KELASRAWAT", "TIPELAYANAN", "CARAPULANG", "DPJPKODE", "DPJPNAMA", "DIAGNOSA", "PROSEDUR", "KODEICBG", "DESKRIPSICBG", "SEVERITYLEVEL", "IDRGKODE", "TARIFINACBG", "KODE_ADDON", "TARIF_ADDON"}
                for _, k := range komponenPool {
                        header = append(header, k)
                }
        }
        var sb strings.Builder
        sb.WriteString(strings.Join(header, "\t"))
        sb.WriteString("\n")
        sep, mr := sepAwal, mrAwal
        for i := 0; i < nBaris; i++ {
                sep++
                mr++
                noSep := fmt.Sprintf("0301%011d", sep)
                noMr := fmt.Sprintf("RM-%d", mr)
                noKartu := fmt.Sprintf("000121%08d", sep%100_000_000)
                nama := acakPilih(rnd, namaDepan) + " " + acakPilih(rnd, namaBelakang)
                tahunLahir := 1945 + rnd.Intn(60)
                tglLahir := fmt.Sprintf("%d-%02d-%02d", tahunLahir, 1+rnd.Intn(12), 1+rnd.Intn(28))
                tipe := "RANAP"
                if rnd.Float64() < 0.22 {
                        tipe = "RJTP"
                }
                cbg := acakPilih(rnd, katalogCBG[:len(katalogCBG)-1])
                if tipe == "RJTP" && rnd.Float64() < 0.7 {
                        cbg = katalogCBG[len(katalogCBG)-1]
                }
                sev := []int{1, 1, 1, 2, 2, 3}[rnd.Intn(6)]
                if strings.HasPrefix(cbg.Kode, "RJ") {
                        sev = 1
                }
                tarif := int64(float64(cbg.Tarif) * sevMult[sev] * (0.96 + rnd.Float64()*0.08))
                kelas := []int{1, 2, 2, 3, 3, 3}[rnd.Intn(6)]
                masuk := tanggalAcak(rnd, periode)
                los := 1 + rnd.Intn(7)
                if tipe == "RJTP" {
                        los = 1
                }
                keluar := tambahHari(masuk, los)

                undi := rnd.Float64()
                var dx []DiagnosisRef
                kemo, kronis := false, rnd.Float64() < 0.3
                billing := tarif
                if undi < 0.52 {
                        billing = int64(float64(tarif) * (0.62 + rnd.Float64()*0.3))
                        dx = []DiagnosisRef{acakPilih(rnd, dxUmum)}
                        if rnd.Float64() < 0.6 {
                                d2 := acakPilih(rnd, dxUmum)
                                if d2.Kode != dx[0].Kode {
                                        dx = append(dx, d2)
                                }
                        }
                } else if undi < 0.78 {
                        billing = int64(float64(tarif) * (1.04 + rnd.Float64()*0.18))
                        dx = []DiagnosisRef{acakPilih(rnd, dxUmum)}
                } else if undi < 0.92 {
                        billing = int64(float64(tarif) * (1.42 + rnd.Float64()*0.3))
                        dx = []DiagnosisRef{acakPilih(rnd, dxUmum)}
                        for j := 0; j < 2+rnd.Intn(2); j++ {
                                c := acakPilih(rnd, dxCC)
                                dx = append(dx, c)
                        }
                } else {
                        billing = int64(float64(tarif) * (2.6 + rnd.Float64()*0.8))
                        dx = []DiagnosisRef{acakPilih(rnd, dxUmum)}
                }
                if kronis {
                        dx = append(dx, DiagnosisRef{"E11.9", "DM tipe 2"})
                }
                var diagStr []string
                for _, d := range dx {
                        diagStr = append(diagStr, d.Kode)
                }
                pro := []ProsedurRef{}
                if rnd.Float64() < 0.35 {
                        pro = append(pro, acakPilih(rnd, katalogProsedur))
                }
                var proStr []string
                for _, p := range pro {
                        proStr = append(proStr, p.Kode)
                }
                dpjp := acakPilih(rnd, katalogDPJP)
                kemoDx := false
                for _, d := range diagStr {
                        if strings.HasPrefix(d, "C") {
                                kemoDx = true
                        }
                }
                if kemoDx && rnd.Float64() < 0.5 {
                        kemo = true
                }
                komponen := map[string]int64{}
                if format == "DETAIL" {
                        komponen = komponenBillingAcak(rnd, billing, kemo, kronis, 0, 0)
                }
                caraPulang := []string{"Atas Persetujuan Dokter", "Membaik", "Rujuk Lanjut"}[rnd.Intn(3)]

                var vals []string
                if format == "BASIC" {
                        vals = []string{noSep, noKartu, noMr, nama, tglLahir, "P", masuk, keluar, itoa(los), itoa(kelas), tipe,
                                strings.Join(diagStr, ","), strings.Join(proStr, ","), cbg.Kode, cbg.Desk, itoa(sev), strconv.FormatInt(tarif, 10)}
                } else {
                        addonKode, addonTarif := "", int64(0)
                        if rnd.Float64() < 0.08 {
                                addonKode = "SP-1" + itoa(10+rnd.Intn(89))
                                addonTarif = 150_000 + int64(rnd.Intn(20))*10_000
                        }
                        vals = []string{noSep, noKartu, noMr, nama, tglLahir, "P", masuk, keluar, itoa(los), itoa(kelas), tipe, caraPulang, dpjp.Kode, dpjp.Nama,
                                strings.Join(diagStr, ","), strings.Join(proStr, ","), cbg.Kode, cbg.Desk, itoa(sev), cbg.Kode, strconv.FormatInt(tarif, 10), addonKode, strconv.FormatInt(addonTarif, 10)}
                        for _, k := range komponenPool {
                                vals = append(vals, strconv.FormatInt(komponen[k], 10))
                        }
                }
                sb.WriteString(strings.Join(vals, "\t"))
                sb.WriteString("\n")
        }
        return sb.String(), sep, mr
}

// ---------- Ingest upload (dipakai seed & API commit) ----------

type HasilIngest struct {
        Upload    *UploadBatch `json:"upload"`
        BarisBaru int          `json:"barisBaru"`
        BarisBerubah int       `json:"barisBerubah"`
        BarisUnchanged int     `json:"barisUnchanged"`
        LedgerBaru int         `json:"ledgerBaru"`
        Diff      map[string][]FieldDiff `json:"diff,omitempty"`
}

type FieldDiff struct {
        Field string `json:"field"`
        Label string `json:"label"`
        Lama  string `json:"lama"`
        Baru  string `json:"baru"`
}

func medianUploadSebelumnya(s *Store, hospitalID, excludeID string) float64 {
        vals := []float64{}
        for _, up := range s.Uploads {
                if up.HospitalID == hospitalID && up.ID != excludeID && up.Status == "SELESAI" {
                        vals = append(vals, float64(up.TotalBaris))
                }
        }
        return median(vals)
}

func sepEksistingSet(s *Store, hospitalID string) map[string]bool {
        set := map[string]bool{}
        for _, ep := range s.Episodes {
                if ep.Superseded {
                        continue
                }
                if up := s.cariUpload(ep.UploadBatchID); up != nil && up.HospitalID == hospitalID {
                        set[ep.NoSep] = true
                }
        }
        return set
}

func (s *Store) cariUpload(id string) *UploadBatch {
        for _, u := range s.Uploads {
                if u.ID == id {
                        return u
                }
        }
        return nil
}

func (s *Store) cariEpisode(id string) *Episode {
        for _, e := range s.Episodes {
                if e.ID == id {
                        return e
                }
        }
        return nil
}

var namaDx = func() map[string]string {
        m := map[string]string{}
        for _, d := range dxUmum {
                m[d.Kode] = d.Nama
        }
        for _, d := range dxCC {
                m[d.Kode] = d.Nama
        }
        m["J44.1"] = "COPD eksaserbasi"
        return m
}()

func namaDiagnosa(kode string) string {
        if n, ok := namaDx[kode]; ok {
                return n
        }
        return "Diagnosis " + kode
}

var namaPros = func() map[string]string {
        m := map[string]string{}
        for _, p := range katalogProsedur {
                m[p.Kode] = p.Nama
        }
        return m
}()

func namaProsedur(kode string) string {
        if n, ok := namaPros[kode]; ok {
                return n
        }
        return "Prosedur " + kode
}

// ingestUpload — validasi, DQ report, versioning & auto-match ledger.
func (s *Store) ingestUpload(hospitalID, periode, namaFile, isi, userID, alasanOverride string, override bool) (*HasilIngest, *RespError) {
        h := s.cariHospital(hospitalID)
        if h == nil {
                return nil, &RespError{400, "Rumah sakit tidak ditemukan.", nil}
        }
        salt := h.Kode
        parse := parseTSV(isi)
        validasi := validasiBaris(parse)

        format := "BASIC"
        for _, k := range validasi.KolomTersedia {
                if k == "VISITATION_FEE" {
                        format = "DETAIL"
                        break
                }
        }

        ctx := KonteksDQ{
                KolomTersedia:          validasi.KolomTersedia,
                PeriodeDipilih:         periode,
                MedianUploadSebelumnya: medianUploadSebelumnya(s, hospitalID, ""),
                SepEksisting:           sepEksistingSet(s, hospitalID),
        }
        report := buatLaporanDQ(validasi.Valid, len(validasi.Gagal), ctx)

        // Gerbang DQ
        if report.Gerbang == "BLOKIR" && !override {
                return nil, &RespError{
                        Kode:  409,
                        Pesan: fmt.Sprintf("Gerbang DQ: BLOKIR (skor %d, grade %s). Perbaiki data atau gunakan override oleh Casemix/Super Admin.", report.Skor, report.Gerbang),
                        Data:  report,
                }
        }
        if override && report.Gerbang == "BLOKIR" {
                report.Gerbang = "OVERRIDE"
        }

        // Versioning
        totalBaris := len(parse.Baris)
        barisValid := len(validasi.Valid)
        up := &UploadBatch{
                ID: idBaru("up"), HospitalID: hospitalID, NamaFile: namaFile, Periode: periode,
                Format: format, DQScore: report.Skor, DQGrade: report.Grade, DQReport: &report,
                Status: "SELESAI", TotalBaris: totalBaris, BarisValid: barisValid,
                BarisGagal: len(validasi.Gagal), OverrideUsed: override,
                UploadedByID: userID, CreatedAt: sekarangISO(),
        }
        if override && alasanOverride != "" {
                up.AlasanOverride = &alasanOverride
        }

        // indeks episode aktif per SEP untuk hospital ini
        aktifPerSEP := map[string]*Episode{}
        for _, ep := range s.Episodes {
                if ep.Superseded {
                        continue
                }
                if u2 := s.cariUpload(ep.UploadBatchID); u2 != nil && u2.HospitalID == hospitalID {
                        aktifPerSEP[ep.NoSep] = ep
                }
        }

        hasil := &HasilIngest{Upload: up}
        hasil.Diff = map[string][]FieldDiff{}
        ledgerAda := map[string]bool{}
        for _, l := range s.Ledger {
                ledgerAda[l.Sep] = true
        }
        estimasiMap := estimasiUpgradeTarif(s.Episodes)
        now := sekarangISO()

        for i := range validasi.Valid {
                b := &validasi.Valid[i]
                hash := rowHash(*b, salt)
                lama, ada := aktifPerSEP[b.NoSep]
                if ada && lama.RowHash == hash {
                        hasil.BarisUnchanged++
                        continue
                }
                if ada {
                        // versi baru
                        if diff := diffBaris(lama, b); len(diff) > 0 {
                                hasil.Diff[b.NoSep] = diff
                        }
                        lama.Superseded = true
                }
                var dpjpNama *string
                if b.DpjpKode != "" {
                        for _, d := range katalogDPJP {
                                if d.Kode == b.DpjpKode {
                                        n := d.Nama
                                        dpjpNama = &n
                                        break
                                }
                        }
                }
                var caraPulang, idrg *string
                if b.CaraPulang != "" {
                        caraPulang = &b.CaraPulang
                }
                if b.IdrgKode != "" {
                        idrg = &b.IdrgKode
                }
                versionNo := 1
                if ada {
                        versionNo = lama.VersionNo + 1
                }
                selisih := hitungSelisih(b.TarifInacbg, b.TotalBilling)
                ep := &Episode{
                        ID: b.NoSep + "#v" + itoa(versionNo), UploadBatchID: up.ID,
                        NoSep: b.NoSep, NoKartu: b.NoKartu, NoMr: b.NoMr, NamaPasien: b.NamaPasien,
                        TglLahir: b.TglLahir, JenisKelamin: b.JenisKelamin,
                        TglMasuk: b.TglMasuk, TglKeluar: b.TglKeluar, Los: b.Los,
                        KelasRawat: b.KelasRawat, TipeLayanan: b.TipeLayanan,
                        CaraPulang: caraPulang, DpjpKode: func() *string { if b.DpjpKode == "" { return nil }; return &b.DpjpKode }(),
                        DpjpNama: dpjpNama,
                        Diagnosa: refsDiagnosa(b.Diagnosa), Prosedur: refsProsedur(b.Prosedur),
                        KodeIcbg: b.KodeIcbg, DeskripsiIcbg: b.DeskripsiIcbg, SeverityLevel: b.SeverityLevel,
                        IdrgKode: idrg, Komponen: b.Komponen, TotalBilling: b.TotalBilling,
                        TarifInacbg: b.TarifInacbg, Selisih: selisih,
                        PctRugi: hitungPctRugi(b.TarifInacbg, b.TotalBilling),
                        KelasRugi: klasifikasiRugi(b.TarifInacbg, b.TotalBilling),
                        AddOn: b.AddOn, DQScore: report.Skor, RowHash: hash,
                        VersionNo: versionNo, CreatedAt: now,
                }
                s.Episodes = append(s.Episodes, ep)
                if ada {
                        hasil.BarisBerubah++
                } else {
                        hasil.BarisBaru++
                }
                // Auto-match ledger: episode rugi (bukan surplus) tanpa entri
                if !ledgerAda[b.NoSep] && ep.KelasRugi != "SURPLUS" {
                        est := estimasiMap[ep.ID]
                        if est == 0 {
                                est = ep.Selisih / -2
                        }
                        strategi := strategiUntuk(ep)
                        s.Ledger = append(s.Ledger, &LedgerEntry{
                                ID: idBaru("lg"), EpisodeID: ep.ID, HospitalID: hospitalID,
                                Sep: ep.NoSep, NomorMr: ep.NoMr, NamaPasien: ep.NamaPasien,
                                KodeIcbg: ep.KodeIcbg, KelasRugi: ep.KelasRugi, Estimasi: est,
                                Status: "KANDIDAT", Strategy: &strategi,
                                CreatedAt: now, UpdatedAt: now,
                        })
                        ledgerAda[b.NoSep] = true
                        hasil.LedgerBaru++
                }
        }

        up.BarisBaru = hasil.BarisBaru
        up.BarisBerubah = hasil.BarisBerubah
        up.BarisUnchanged = hasil.BarisUnchanged
        s.Uploads = append(s.Uploads, up)
        return hasil, nil
}

func refsDiagnosa(kode []string) []DiagnosisRef {
        out := make([]DiagnosisRef, len(kode))
        for i, k := range kode {
                out[i] = DiagnosisRef{Kode: k, Nama: namaDiagnosa(k)}
        }
        return out
}

func refsProsedur(kode []string) []ProsedurRef {
        out := make([]ProsedurRef, len(kode))
        for i, k := range kode {
                out[i] = ProsedurRef{Kode: k, Nama: namaProsedur(k)}
        }
        return out
}

func strategiUntuk(ep *Episode) string {
        switch ep.KelasRugi {
        case "S1_CODING":
                return "S1 — Klarifikasi CC/MCC dengan DPJP & kaji ulang koding"
        case "S2_MIXED":
                return "S2 — Klarifikasi diagnosis sekunder + verifikasi rincian billing"
        default:
                return "S3 — Verifikasi struktur tarif & komponen billing"
        }
}

// diffBaris — bandingkan versi lama & baru.
func diffBaris(lama *Episode, baru *BarisValid) []FieldDiff {
        diff := []FieldDiff{}
        sama := func(a, b []string) bool {
                if len(a) != len(b) {
                        return false
                }
                x := append([]string(nil), a...)
                y := append([]string(nil), b...)
                sort.Strings(x)
                sort.Strings(y)
                for i := range x {
                        if x[i] != y[i] {
                                return false
                        }
                }
                return true
        }
        if !sama(lama.diagnosaKode(), baru.Diagnosa) {
                diff = append(diff, FieldDiff{"diagnosa", "Diagnosis (ICD-10)", strings.Join(lama.diagnosaKode(), ", "), strings.Join(baru.Diagnosa, ", ")})
        }
        if !sama(lama.prosedurKode(), baru.Prosedur) {
                diff = append(diff, FieldDiff{"prosedur", "Prosedur (ICD-9-CM)", strings.Join(lama.prosedurKode(), ", "), strings.Join(baru.Prosedur, ", ")})
        }
        if lama.KodeIcbg != baru.KodeIcbg {
                diff = append(diff, FieldDiff{"kodeIcbg", "Kode INA-CBG", lama.KodeIcbg, baru.KodeIcbg})
        }
        if lama.SeverityLevel != baru.SeverityLevel {
                diff = append(diff, FieldDiff{"severityLevel", "Severity Level", itoa(lama.SeverityLevel), itoa(baru.SeverityLevel)})
        }
        if lama.TarifInacbg != baru.TarifInacbg {
                diff = append(diff, FieldDiff{"tarifInacbg", "Tarif INA-CBG", fmtInt64(lama.TarifInacbg), fmtInt64(baru.TarifInacbg)})
        }
        if lama.Los != baru.Los {
                diff = append(diff, FieldDiff{"los", "LOS (hari)", itoa(lama.Los), itoa(baru.Los)})
        }
        if lama.KelasRawat != baru.KelasRawat {
                diff = append(diff, FieldDiff{"kelasRawat", "Kelas Rawat", itoa(lama.KelasRawat), itoa(baru.KelasRawat)})
        }
        return diff
}

func (ep *Episode) diagnosaKode() []string {
        out := make([]string, len(ep.Diagnosa))
        for i, d := range ep.Diagnosa {
                out[i] = d.Kode
        }
        return out
}

func (ep *Episode) prosedurKode() []string {
        out := make([]string, len(ep.Prosedur))
        for i, p := range ep.Prosedur {
                out[i] = p.Kode
        }
        return out
}

// ---------- Seed utama ----------

const direktoriSample = "/home/z/my-project/samples"

func seed(s *Store) {
        rnd := rand.New(rand.NewSource(20260214))
        now := sekarangISO()

        // Hospitals
        rs1 := &Hospital{ID: "rs1", Kode: "RS001", Nama: "RSUD Kota Anggrek", Kota: "Bandung", Kelas: "B"}
        rs2 := &Hospital{ID: "rs2", Kode: "RS002", Nama: "RS Harapan Medika", Kota: "Bekasi", Kelas: "B2"}
        rs3 := &Hospital{ID: "rs3", Kode: "RS003", Nama: "RS Pelayanan Sentosa", Kota: "Semarang", Kelas: "C"}
        s.Hospitals = []*Hospital{rs1, rs2, rs3}

        // Users
        log.Println("seed: hospitals ok, hashing sandi"); sandi := hashSandi("MedClaim#demo2026"); log.Println("seed: sandi selesai")
        rs1id := rs1.ID
        s.Users = []*User{
                {ID: "u-admin", Email: "admin@demo.medclaim.id", PasswordHash: sandi, Nama: "Admin SPKD", Role: "SUPERADMIN", Aktif: true, CreatedAt: now},
                {ID: "u-casemix", Email: "casemix@demo.medclaim.id", PasswordHash: sandi, Nama: "Rina Casemix", Role: "CASEMIX", HospitalID: &rs1id, Aktif: true, CreatedAt: now},
                {ID: "u-coder", Email: "coder@demo.medclaim.id", PasswordHash: sandi, Nama: "Citra Coder", Role: "CODER", HospitalID: &rs1id, Aktif: true, CreatedAt: now},
                {ID: "u-direktur", Email: "direktur@demo.medclaim.id", PasswordHash: sandi, Nama: "Drs. Hendra Direktur", Role: "DIREKTUR", HospitalID: &rs1id, Aktif: true, CreatedAt: now},
                {ID: "u-dpjp", Email: "dpjp@demo.medclaim.id", PasswordHash: sandi, Nama: katalogDPJP[0].Nama, Role: "DPJP", HospitalID: &rs1id, Spesialisasi: strPtr(katalogDPJP[0].Spes), DpjpKode: strPtr(katalogDPJP[0].Kode), Aktif: true, CreatedAt: now},
        }
        for i := 1; i < len(katalogDPJP); i++ {
                d := katalogDPJP[i]
                s.Users = append(s.Users, &User{
                        ID: "u-dpjp" + itoa(i+1), Email: "dpjp" + d.Kode[len(d.Kode)-2:] + "@demo.medclaim.id",
                        PasswordHash: sandi, Nama: d.Nama, Role: "DPJP", HospitalID: &rs1id,
                        Spesialisasi: strPtr(d.Spes), DpjpKode: strPtr(d.Kode), Aktif: true, CreatedAt: now,
                })
        }

        // Templates & settings
        s.Templates = []*TemplateDok{
                {ID: "tpl-1", Nama: "Klarifikasi Diagnosis Utama", Isi: "Bagaimana kondisi klinis utama pasien pada episode ini dan temuan penunjang apa yang mendukung diagnosis utamanya?", Kategori: "KODING", Aktif: true},
                {ID: "tpl-2", Nama: "Kelengkapan Resumé", Isi: "Dokumen apa yang perlu dilengkapi pada resumé medis episode ini agar kondisi komorbid pasien terdokumentasi utuh?", Kategori: "DOKUMEN", Aktif: true},
                {ID: "tpl-3", Nama: "Catatan Operasi", Isi: "Apakah terdapat catatan tindakan/operasi yang belum tercatat pada episode ini dan kapan tindakan tersebut dilakukan?", Kategori: "KODING", Aktif: true},
        }
        s.Settings = []*Setting{
                {Key: "readiness.weights", Value: `{"severity":30,"drug":25,"singleDiag":20,"icd":15,"los":10}`},
                {Key: "dq.penalti", Value: `{"kritis":100,"peringatan":25,"info":10}`},
                {Key: "dq.max_invalid_row_pct", Value: "5"},
                {Key: "sla.query_hari", Value: "3"},
        }

        casemixID := "u-casemix"

        // RS001 — ingest file sample asli (basic → detail → detail-v2)
        seedSample := func(namaFile, periode string) {
                isi, err := bacaFile(filepath.Join(direktoriSample, namaFile))
                if err != nil {
                        return
                }
                s.ingestUpload(rs1.ID, periode, namaFile, string(isi), casemixID, "", false)
        }
        log.Println("seed: mulai basic")
        isiBasic, errBasic := bacaFile(filepath.Join(direktoriSample, "vclaim-basic.txt"))
        if errBasic != nil {
                log.Println("seed: basic gagal dibaca:", errBasic)
        } else if _, errResp := s.ingestUpload(rs1.ID, "2025-10", "vclaim-basic.txt", string(isiBasic), casemixID, "Format Basic memang tanpa komponen billing dan DPJP — lanjut sebagai versi awal.", true); errResp != nil {
                log.Println("seed: basic GAGAL:", errResp.Pesan)
        }
        log.Println("seed: basic selesai")
        log.Println("seed: mulai detail"); seedSample("vclaim-detail.txt", "2025-11"); log.Println("seed: detail selesai")
        log.Println("seed: mulai detail-v2"); seedSample("vclaim-detail-v2.txt", "2025-11"); log.Println("seed: detail-v2 selesai")

        // RS002 & RS003 — generator sintetis
        sep, mr := 500_000, 50_000
        konfig := []struct {
                rs       *Hospital
                nUpload  int
                perUp    []int
                periode  []string
                format   []string
        }{
                {rs2, 3, []int{220, 230, 240}, []string{"2025-08", "2025-09", "2025-10"}, []string{"DETAIL", "DETAIL", "DETAIL"}},
                {rs3, 2, []int{150, 160}, []string{"2025-09", "2025-10"}, []string{"BASIC", "DETAIL"}},
        }
        for _, k := range konfig {
                for u := 0; u < k.nUpload; u++ {
                        log.Println("seed: upload", k.rs.ID, u)
                        log.Println("seed: generate tsv", k.perUp[u]); isi, sepBaru, mrBaru := buatTSVSintetis(rnd, k.perUp[u], k.periode[u], k.format[u], sep, mr, k.rs.Kode); log.Println("seed: generate tsv selesai")
                        sep, mr = sepBaru, mrBaru
                        nama := fmt.Sprintf("VCLAIM-%s-%s-%s.txt", k.rs.Kode, k.periode[u], k.format[u])
                        s.ingestUpload(k.rs.ID, k.periode[u], nama, isi, casemixID, "", false)
                }
        }

        // Seed ledger: dorong sebagian kandidat melalui state machine
        log.Println("seed: ledger mulai, entri:", len(s.Ledger))
        s.seedLedgerProgres(rnd)
        log.Println("seed: ledger selesai, event:", len(s.LedgerEvents))

        // Seed query
        s.seedQueries(rnd)

        // Audit awal
        aksi := "SEED"
        s.catatAudit("system@medclaim.id", aksi, "DATABASE", nil, strPtr("Seed data demo MedClaim Go"))
}

func (s *Store) seedLedgerProgres(rnd *rand.Rand) {
        kandidat := []*LedgerEntry{}
        for _, l := range s.Ledger {
                if l.Status == "KANDIDAT" {
                        kandidat = append(kandidat, l)
                }
        }
        sort.Slice(kandidat, func(i, j int) bool { return kandidat[i].Sep < kandidat[j].Sep })
        now := sekarangISO()
        aktorCase := "Rina Casemix"
        aktorCoder := "Citra Coder"
        maxN := len(kandidat)
        if maxN > 60 {
                maxN = 60
        }
        for i := 0; i < maxN; i++ {
                l := kandidat[i]
                gelombang := i % 6
                switch gelombang {
                case 1: // DIPERIKSA
                        s.pindahLedger(l, "DIPERIKSA", nil, aktorCoder, nil, now)
                case 2: // DIKOREKSI
                        s.pindahLedger(l, "DIPERIKSA", nil, aktorCoder, nil, now)
                        s.pindahLedger(l, "DIKOREKSI", nil, aktorCoder, strPtr("RM/"+l.NomorMr+"/resume-medIS"), now)
                case 3: // DIAJUKAN
                        s.pindahLedger(l, "DIPERIKSA", nil, aktorCoder, nil, now)
                        s.pindahLedger(l, "DIKOREKSI", nil, aktorCoder, strPtr("RM/"+l.NomorMr+"/resume-medIS"), now)
                        s.pindahLedger(l, "DIAJUKAN", nil, aktorCase, nil, now)
                case 4: // HASIL (realisasi)
                        s.pindahLedger(l, "DIPERIKSA", nil, aktorCoder, nil, now)
                        s.pindahLedger(l, "DIKOREKSI", nil, aktorCoder, strPtr("RM/"+l.NomorMr+"/resume-medIS"), now)
                        s.pindahLedger(l, "DIAJUKAN", nil, aktorCase, nil, now)
                        real := l.Estimasi * int64(70+rnd.Intn(45)) / 100
                        s.pindahLedger(l, "HASIL", &real, aktorCase, nil, now)
                case 5: // TIDAK BERUBAH
                        s.pindahLedger(l, "DIPERIKSA", nil, aktorCoder, nil, now)
                        s.pindahLedger(l, "TIDAK_BERUBAH", nil, aktorCase, strPtr("Dokumentasi lengkap, koding sudah sesuai resume medis."), now)
                }
        }
}

func (s *Store) pindahLedger(l *LedgerEntry, ke string, realisasi *int64, aktor string, catatan *string, waktu string) {
        dari := l.Status
        l.Status = ke
        l.UpdatedAt = waktu
        if realisasi != nil {
                l.Realisasi = realisasi
        }
        var dariPtr *string
        if dari != "" {
                dariPtr = &dari
        }
        kePtr := ke
        s.LedgerEvents = append(s.LedgerEvents, &LedgerEvent{
                ID: idBaru("ev"), LedgerID: l.ID, DariStatus: dariPtr, KeStatus: &kePtr,
                AktorNama: aktor, Catatan: catatan, Waktu: waktu,
        })
}

func (s *Store) seedQueries(rnd *rand.Rand) {
        now := time.Now().UTC()
        dpjpUsers := []*User{}
        for _, u := range s.Users {
                if u.Role == "DPJP" {
                        dpjpUsers = append(dpjpUsers, u)
                }
        }
        // pilih episode rugi sebagai konteks
        target := []*Episode{}
        for _, ep := range s.Episodes {
                if !ep.Superseded && ep.KelasRugi != "SURPLUS" && ep.DpjpKode != nil {
                        target = append(target, ep)
                        if len(target) >= 40 {
                                break
                        }
                }
        }
        for i, ep := range target {
                if i >= 8 {
                        break
                }
                dpjp := dpjpUsers[0]
                for _, u := range dpjpUsers {
                        if u.DpjpKode != nil && *u.DpjpKode == *ep.DpjpKode {
                                dpjp = u
                                break
                        }
                }
                tpl := s.Templates[i%len(s.Templates)]
                status := []string{"TERKIRIM", "TERKIRIM", "DIJAWAB", "TERKIRIM", "DIJAWAB", "TERKIRIM", "DIJAWAB", "TERKIRIM"}[i]
                umurHari := 1 + rnd.Intn(6)
                sent := now.AddDate(0, 0, -umurHari)
                deadline := sent.AddDate(0, 0, 3)
                q := &QueryDok{
                        ID: idBaru("qr"), HospitalID: "rs1", Sep: ep.NoSep, NomorMr: ep.NoMr,
                        Inisial: inisial(ep.NamaPasien), TglMasuk: &ep.TglMasuk, TglKeluar: &ep.TglKeluar,
                        DpjpID: dpjp.ID, DpjpNama: dpjp.Nama, DpjpSpesialisasi: dpjp.Spesialisasi,
                        TemplateNama: &tpl.Nama, Pertanyaan: tpl.Isi,
                        Konteks: "Episode " + ep.NoSep + " — " + ep.KodeIcbg + " (" + ep.DeskripsiIcbg + "), kelas rugi " + ep.KelasRugi + ".",
                        Status: status, Deadline: deadline.UTC().Format("2006-01-02T15:04:05.000Z"),
                        SentAt: strPtr(sent.UTC().Format("2006-01-02T15:04:05.000Z")),
                        SlaHari: 3, CreatedAt: sent.UTC().Format("2006-01-02T15:04:05.000Z"),
                }
                if status == "DIJAWAB" {
                        jawab := sent.AddDate(0, 0, 1).UTC().Format("2006-01-02T15:04:05.000Z")
                        q.Status = "DIJAWAB"
                        q.AnsweredAt = &jawab
                        q.Jawaban = strPtr("Sudah diklarifikasi: diagnosis sekunder terdokumentasi pada resumé medis dan penunjang. Silakan sesuaikan koding sesuai temuan klinis tersebut.")
                }
                s.Queries = append(s.Queries, q)
        }
}

func inisial(nama string) string {
        parts := strings.Fields(nama)
        if len(parts) == 0 {
                return "XX"
        }
        if len(parts) == 1 {
                return parts[0][:2]
        }
        return parts[0][:1] + parts[1][:1]
}

func strPtr(s string) *string { return &s }
