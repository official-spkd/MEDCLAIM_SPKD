// ============================================================
// MedClaim Go — API handlers (port dari src/app/api/medclaim/*)
// ============================================================
package main

import (
        "encoding/json"
        "fmt"
        "math"
        "net/http"
        "path/filepath"
        "sort"
        "strconv"
        "strings"
        "time"
)

// ---------- Util respons ----------

type RespError struct {
        Kode  int         `json:"-"`
        Pesan string      `json:"pesan"`
        Data  interface{} `json:"data,omitempty"`
}

func tulisJSON(w http.ResponseWriter, kode int, v interface{}) {
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        w.WriteHeader(kode)
        json.NewEncoder(w).Encode(v)
}

func tulisGagal(w http.ResponseWriter, e *RespError) {
        tulisJSON(w, e.Kode, e)
}

func bacaBody(r *http.Request, target interface{}) *RespError {
        defer r.Body.Close()
        if err := json.NewDecoder(r.Body).Decode(target); err != nil {
                return &RespError{400, "Body request tidak valid: " + err.Error(), nil}
        }
        return nil
}

// ---------- DTO episode (penyembunyian finansial untuk DPJP) ----------

type EpisodePub struct {
        *Episode
        Readiness *ReadinessResult `json:"readiness,omitempty"`
        Anomali   []string         `json:"anomali,omitempty"`
}

func publikasikanEpisode(ep *Episode, lihatFinansial bool, anomali []string) *EpisodePub {
        pub := &EpisodePub{Episode: ep, Anomali: anomali}
        if !lihatFinansial {
                pub.TarifInacbg = 0
                pub.TotalBilling = 0
                pub.Selisih = 0
                pub.PctRugi = 0
                pub.KelasRugi = "RAHASIA"
                pub.Komponen = map[string]int64{}
                for _, a := range pub.AddOn {
                        a.Tarif = 0
                }
        }
        return pub
}

// ---------- Filter episode ----------

type FilterEpisode struct {
        Rs, Upload, Tipe, Dari, Sampai, KelasRugi, Kelas, Sev, Q string
}

func bacaFilter(r *http.Request) FilterEpisode {
        return FilterEpisode{
                Rs: r.URL.Query().Get("rs"), Upload: r.URL.Query().Get("upload"),
                Tipe: r.URL.Query().Get("tipe"), Dari: r.URL.Query().Get("dari"),
                Sampai: r.URL.Query().Get("sampai"), KelasRugi: r.URL.Query().Get("kelasRugi"),
                Kelas: r.URL.Query().Get("kelas"), Sev: r.URL.Query().Get("sev"),
                Q: strings.TrimSpace(r.URL.Query().Get("q")),
        }
}

func (f *FilterEpisode) terapkan(episodes []*Episode, uploads []*UploadBatch) []*Episode {
        uploadRS := map[string]string{}
        for _, u := range uploads {
                uploadRS[u.ID] = u.HospitalID
        }
        out := []*Episode{}
        for _, ep := range episodes {
                if ep.Superseded {
                        continue
                }
                if f.Rs != "" && uploadRS[ep.UploadBatchID] != f.Rs {
                        continue
                }
                if f.Upload != "" && ep.UploadBatchID != f.Upload {
                        continue
                }
                if f.Tipe != "" && f.Tipe != "SEMUA" && ep.TipeLayanan != f.Tipe {
                        continue
                }
                if f.Dari != "" && ep.TglMasuk < f.Dari {
                        continue
                }
                if f.Sampai != "" && ep.TglMasuk > f.Sampai {
                        continue
                }
                if f.KelasRugi != "" && f.KelasRugi != "SEMUA" && ep.KelasRugi != f.KelasRugi {
                        continue
                }
                if f.Kelas != "" && itoa(ep.KelasRawat) != f.Kelas {
                        continue
                }
                if f.Sev != "" && itoa(ep.SeverityLevel) != f.Sev {
                        continue
                }
                if f.Q != "" {
                        ql := strings.ToLower(f.Q)
                        cocok := strings.Contains(strings.ToLower(ep.NoSep), ql) ||
                                strings.Contains(strings.ToLower(ep.NoMr), ql) ||
                                strings.Contains(strings.ToLower(ep.NamaPasien), ql) ||
                                strings.Contains(strings.ToLower(ep.KodeIcbg), ql)
                        if !cocok {
                                for _, d := range ep.Diagnosa {
                                        if strings.Contains(strings.ToLower(d.Kode), ql) {
                                                cocok = true
                                                break
                                        }
                                }
                        }
                        if !cocok {
                                continue
                        }
                }
                out = append(out, ep)
        }
        return out
}

// scopeSesi — DPJP hanya episode miliknya & RS miliknya; staf non-superadmin terkunci RS-nya.
func scopeSesi(sesi *Sesi, episodes []*Episode) []*Episode {
        out := []*Episode{}
        for _, ep := range episodes {
                if sesi.Role == "DPJP" {
                        if ep.DpjpKode == nil || sesi.DpjpKode == nil || *ep.DpjpKode != *sesi.DpjpKode {
                                continue
                        }
                }
                out = append(out, ep)
        }
        return out
}

// ---------- Auth ----------

func handleLogin(w http.ResponseWriter, r *http.Request) {
        var body struct {
                Email    string `json:"email"`
                Password string `json:"password"`
        }
        if e := bacaBody(r, &body); e != nil {
                tulisGagal(w, e)
                return
        }
        email := strings.ToLower(strings.TrimSpace(body.Email))
        var user *User
        for _, u := range store.Users {
                if strings.ToLower(u.Email) == email && u.Aktif {
                        user = u
                        break
                }
        }
        if user == nil || !verifikasiSandi(body.Password, user.PasswordHash) {
                tulisGagal(w, &RespError{401, "Email atau kata sandi salah.", nil})
                return
        }
        now := sekarangISO()
        user.LastLoginAt = &now
        store.catatAudit(user.Email, "LOGIN", "USER", &user.ID, nil)
        store.simpan()
        setCookieSesi(w, user.ID)
        tulisJSON(w, 200, buatSesi(user))
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
        if s := getSesi(r); s != nil {
                store.catatAudit(s.Email, "LOGOUT", "USER", &s.ID, nil)
                store.simpan()
        }
        hapusCookieSesi(w)
        tulisJSON(w, 200, map[string]bool{"ok": true})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        tulisJSON(w, 200, sesi)
}

// ---------- Bootstrap ----------

func handleBootstrap(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        uploads := []*UploadBatch{}
        for _, u := range store.Uploads {
                if u.Status != "DIHAPUS" {
                        if sesi.Role != "SUPERADMIN" && sesi.HospitalID != nil && u.HospitalID != *sesi.HospitalID {
                                continue
                        }
                        uploads = append(uploads, u)
                }
        }
        sort.Slice(uploads, func(i, j int) bool { return uploads[i].CreatedAt > uploads[j].CreatedAt })

        dokter := []*User{}
        for _, u := range store.Users {
                if u.Role == "DPJP" && u.Aktif {
                        if sesi.Role != "SUPERADMIN" && sesi.HospitalID != nil && u.HospitalID != nil && *u.HospitalID != *sesi.HospitalID {
                                continue
                        }
                        dokter = append(dokter, u)
                }
        }

        resp := map[string]interface{}{
                "me":        sesi,
                "hospitals": store.Hospitals,
                "uploads":   uploads,
                "templates": store.Templates,
                "settings":  store.Settings,
                "dokter":    usersPublik(dokter),
        }
        if sesi.Role == "SUPERADMIN" {
                resp["users"] = usersPublik(store.Users)
        }
        tulisJSON(w, 200, resp)
}

func usersPublik(users []*User) []map[string]interface{} {
        out := []map[string]interface{}{}
        for _, u := range users {
                out = append(out, map[string]interface{}{
                        "id": u.ID, "email": u.Email, "nama": u.Nama, "role": u.Role,
                        "hospitalId": u.HospitalID, "spesialisasi": u.Spesialisasi,
                        "dpjpKode": u.DpjpKode, "aktif": u.Aktif, "createdAt": u.CreatedAt,
                        "lastLoginAt": u.LastLoginAt,
                })
        }
        return out
}

// ---------- Episodes ----------

func handleEpisodes(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        f := bacaFilter(r)
        if sesi.Role != "SUPERADMIN" && sesi.HospitalID != nil && f.Rs == "" {
                f.Rs = *sesi.HospitalID
        }
        eps := scopeSesi(sesi, f.terapkan(store.Episodes, store.Uploads))
        lihatF := bolehLihatFinansial(sesi.Role)
        items := make([]*EpisodePub, 0, len(eps))
        for _, ep := range eps {
                items = append(items, publikasikanEpisode(ep, lihatF, nil))
        }
        sort.Slice(items, func(i, j int) bool { return items[i].NoSep < items[j].NoSep })
        tulisJSON(w, 200, map[string]interface{}{"total": len(items), "items": items})
}

func handleEpisodeDetail(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        id := r.PathValue("id")
        ep := store.cariEpisode(id)
        if ep == nil {
                tulisGagal(w, &RespError{404, "Episode tidak ditemukan.", nil})
                return
        }
        if sesi.Role == "DPJP" && (ep.DpjpKode == nil || sesi.DpjpKode == nil || *ep.DpjpKode != *sesi.DpjpKode) {
                tulisGagal(w, &RespError{403, "Episode ini bukan pasien DPJP Anda.", nil})
                return
        }
        anom := []string{}
        for _, a := range deteksiAnomali([]*Episode{ep}) {
                if a.EpisodeID == ep.ID {
                        anom = append(anom, a.Detail)
                }
        }
        ready := readinessScore(ep)
        pub := publikasikanEpisode(ep, bolehLihatFinansial(sesi.Role), anom)
        pub.Readiness = &ready
        // riwayat versi
        versi := []*Episode{}
        for _, x := range store.Episodes {
                if x.NoSep == ep.NoSep {
                        versi = append(versi, publikasikanEpisode(x, bolehLihatFinansial(sesi.Role), nil).Episode)
                }
        }
        sort.Slice(versi, func(i, j int) bool { return versi[i].VersionNo < versi[j].VersionNo })
        tulisJSON(w, 200, map[string]interface{}{"episode": pub, "versi": versi})
}

// ---------- Dashboard & analitik ----------

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

func hitungDashboard(eps []*Episode) map[string]interface{} {
        var totalTarif, totalBilling int64
        nRugi, sumPct, sumLos := 0, 0.0, 0
        dist := map[string]int{"SURPLUS": 0, "S1_CODING": 0, "S2_MIXED": 0, "STRUCTURAL": 0}
        kelasDist := map[string]int{"1": 0, "2": 0, "3": 0}
        cbg := map[string]*struct {
                n, rugi int
                desk    string
        }{}
        readySum := 0
        readyKomp := map[string]int{}
        epsRugi := []*Episode{}
        for _, ep := range eps {
                totalTarif += ep.TarifInacbg
                totalBilling += ep.TotalBilling
                if ep.KelasRugi != "SURPLUS" {
                        nRugi++
                        sumPct += ep.PctRugi
                        epsRugi = append(epsRugi, ep)
                }
                dist[ep.KelasRugi]++
                kelasDist[itoa(ep.KelasRawat)]++
                sumLos += ep.Los
                c := cbg[ep.KodeIcbg]
                if c == nil {
                        c = &struct{ n, rugi int; desk string }{desk: ep.DeskripsiIcbg}
                        cbg[ep.KodeIcbg] = c
                }
                c.n++
                if ep.KelasRugi != "SURPLUS" && ep.Selisih < 0 {
                        c.rugi++
                }
                rd := readinessScore(ep)
                readySum += rd.Total
                for k, v := range rd.Komponen {
                        readyKomp[k] += v
                }
        }
        est := estimasiUpgradeTarif(eps)
        var potensi int64
        for _, ep := range epsRugi {
                potensi += est[ep.ID]
        }
        var estimasiRugi int64
        for _, ep := range eps {
                if ep.Selisih < 0 {
                        estimasiRugi += -ep.Selisih
                }
        }

        // tren per bulan
        tren := []map[string]interface{}{}
        perBulan := map[string]*struct{ n int; rugi int64 }{}
        var urut []string
        for _, ep := range eps {
                ym := ep.TglMasuk[:7]
                if perBulan[ym] == nil {
                        perBulan[ym] = &struct{ n int; rugi int64 }{}
                        urut = append(urut, ym)
                }
                perBulan[ym].n++
                if ep.Selisih < 0 {
                        perBulan[ym].rugi += -ep.Selisih
                }
        }
        sort.Strings(urut)
        for _, ym := range urut {
                tren = append(tren, map[string]interface{}{"periode": ym, "episode": perBulan[ym].n, "rugi": perBulan[ym].rugi})
        }

        // top CBG
        type CbgRow struct {
                Kode string `json:"kode"`
                Desk string `json:"desk"`
                N    int    `json:"jumlah"`
                Rugi int    `json:"rugi"`
        }
        cbgRows := []CbgRow{}
        for kode, c := range cbg {
                cbgRows = append(cbgRows, CbgRow{kode, c.desk, c.n, c.rugi})
        }
        sort.Slice(cbgRows, func(i, j int) bool {
                if cbgRows[i].N != cbgRows[j].N {
                        return cbgRows[i].N > cbgRows[j].N
                }
                return cbgRows[i].Kode < cbgRows[j].Kode
        })
        if len(cbgRows) > 8 {
                cbgRows = cbgRows[:8]
        }

        kelasRows := []map[string]interface{}{}
        for _, k := range []string{"1", "2", "3"} {
                kelasRows = append(kelasRows, map[string]interface{}{"kelas": k, "jumlah": kelasDist[k]})
        }

        n := len(eps)
        readinessRata := 0
        if n > 0 {
                readinessRata = readySum / n
        }
        kompRata := map[string]int{}
        for k, v := range readyKomp {
                if n > 0 {
                        kompRata[k] = v / n
                }
        }
        pctRata := 0.0
        if nRugi > 0 {
                pctRata = sumPct / float64(nRugi)
        }
        losRata := 0.0
        if n > 0 {
                losRata = float64(sumLos) / float64(n)
        }

        kpi := KPI{
                TotalEpisode: n, TotalTarif: totalTarif, TotalBilling: totalBilling,
                TotalSelisih: totalTarif - totalBilling, PctRugiRata: pctRata,
                PotensiRecovery: potensi, EstimasiRugi: estimasiRugi,
                ReadinessRata: readinessRata, LosRata: losRata,
                AnomaliJumlah: len(deteksiAnomali(eps)),
        }

        return map[string]interface{}{
                "kpi": kpi, "distribusiRugi": dist, "tren": tren,
                "topCbg": cbgRows, "kelasRawat": kelasRows,
                "readiness": map[string]interface{}{"total": readinessRata, "komponen": kompRata},
        }
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        f := bacaFilter(r)
        if sesi.Role != "SUPERADMIN" && sesi.HospitalID != nil && f.Rs == "" {
                f.Rs = *sesi.HospitalID
        }
        eps := scopeSesi(sesi, f.terapkan(store.Episodes, store.Uploads))

        resp := hitungDashboard(eps)
        kpi := resp["kpi"].(KPI)
        kpi.JumlahUpload = 0
        for _, u := range store.Uploads {
                if u.Status == "SELESAI" && (sesi.Role == "SUPERADMIN" || sesi.HospitalID == nil || u.HospitalID == *sesi.HospitalID) {
                        if f.Rs == "" || u.HospitalID == f.Rs {
                                kpi.JumlahUpload++
                        }
                }
        }
        now := sekarangISO()
        for _, q := range store.Queries {
                if sesi.Role == "DPJP" && (q.DpjpID != sesi.ID) {
                        continue
                }
                if q.Status == "TERKIRIM" {
                        if q.Deadline < now {
                                kpi.QueryTerlambat++
                        } else {
                                kpi.QueryTerbuka++
                        }
                }
        }
        kpi.LedgerAktif = 0
        for _, l := range store.Ledger {
                if l.Status != "HASIL" && l.Status != "TIDAK_BERUBAH" {
                        if sesi.Role == "SUPERADMIN" || sesi.HospitalID == nil || l.HospitalID == *sesi.HospitalID {
                                if f.Rs == "" || l.HospitalID == f.Rs {
                                        kpi.LedgerAktif++
                                }
                        }
                }
        }
        resp["kpi"] = kpi

        dqSum, dqN := 0, 0
        for _, u := range store.Uploads {
                if u.Status == "SELESAI" && (f.Rs == "" || u.HospitalID == f.Rs) && (sesi.Role == "SUPERADMIN" || sesi.HospitalID == nil || u.HospitalID == *sesi.HospitalID) {
                        dqSum += u.DQScore
                        dqN++
                }
        }
        if dqN > 0 {
                kpi.DQRata = dqSum / dqN
        }
        resp["kpi"] = kpi

        if !bolehLihatFinansial(sesi.Role) {
                resp = map[string]interface{}{"kpi": kpi, "readiness": resp["readiness"]}
        }
        tulisJSON(w, 200, resp)
}

func handleAnomali(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        f := bacaFilter(r)
        if sesi.Role != "SUPERADMIN" && sesi.HospitalID != nil && f.Rs == "" {
                f.Rs = *sesi.HospitalID
        }
        eps := scopeSesi(sesi, f.terapkan(store.Episodes, store.Uploads))
        hasil := deteksiAnomali(eps)
        items := []map[string]interface{}{}
        byID := map[string]*Episode{}
        for _, ep := range eps {
                byID[ep.ID] = ep
        }
        for _, a := range hasil {
                ep := byID[a.EpisodeID]
                if ep == nil {
                        continue
                }
                items = append(items, map[string]interface{}{
                        "episodeId": a.EpisodeID, "jenis": a.Jenis, "detail": a.Detail,
                        "sep": ep.NoSep, "mr": ep.NoMr, "nama": ep.NamaPasien,
                        "icbg": ep.KodeIcbg, "los": ep.Los, "tglMasuk": ep.TglMasuk,
                })
        }
        tulisJSON(w, 200, map[string]interface{}{"total": len(items), "items": items})
}

func handleFinancial(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehLihatFinansial(sesi.Role) {
                tulisGagal(w, &RespError{403, "Peran DPJP tidak memiliki akses ke modul finansial.", nil})
                return
        }
        f := bacaFilter(r)
        if sesi.Role != "SUPERADMIN" && sesi.HospitalID != nil && f.Rs == "" {
                f.Rs = *sesi.HospitalID
        }
        eps := f.terapkan(store.Episodes, store.Uploads)
        kompTotal := map[string]int64{}
        for _, ep := range eps {
                for _, k := range KOMPONEN_BILLING {
                        kompTotal[k] += ep.Komponen[k]
                }
        }
        var total int64
        for _, v := range kompTotal {
                total += v
        }
        komp := []map[string]interface{}{}
        for _, k := range KOMPONEN_BILLING {
                pct := 0.0
                if total > 0 {
                        pct = float64(kompTotal[k]) / float64(total)
                }
                komp = append(komp, map[string]interface{}{"komponen": k, "nilai": kompTotal[k], "pct": pct})
        }
        sort.Slice(komp, func(i, j int) bool {
                return komp[i]["nilai"].(int64) > komp[j]["nilai"].(int64)
        })
        tulisJSON(w, 200, map[string]interface{}{"total": total, "komponen": komp})
}

// ---------- Uploads ----------

type ReqUpload struct {
        HospitalID     string `json:"hospitalId"`
        Periode        string `json:"periode"`
        NamaFile       string `json:"namaFile"`
        Isi            string `json:"isi"`
        Override       bool   `json:"override"`
        AlasanOverride string `json:"alasanOverride"`
}

func handleUploadPreview(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehAkses("UPLOAD", sesi.Role) {
                tulisGagal(w, &RespError{403, "Hanya Super Admin dan Casemix yang dapat mengunggah data.", nil})
                return
        }
        var req ReqUpload
        if e := bacaBody(r, &req); e != nil {
                tulisGagal(w, e)
                return
        }
        h := store.cariHospital(req.HospitalID)
        if h == nil {
                tulisGagal(w, &RespError{400, "Rumah sakit tidak ditemukan.", nil})
                return
        }
        parse := parseTSV(req.Isi)
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
                PeriodeDipilih:         req.Periode,
                MedianUploadSebelumnya: medianUploadSebelumnya(store, req.HospitalID, ""),
                SepEksisting:           sepEksistingSet(store, req.HospitalID),
        }
        report := buatLaporanDQ(validasi.Valid, len(validasi.Gagal), ctx)
        gagal := validasi.Gagal
        if len(gagal) > 50 {
                gagal = gagal[:50]
        }
        duplikat := []string{}
        hitung := map[string]int{}
        for _, b := range validasi.Valid {
                hitung[b.NoSep]++
        }
        for sep, n := range hitung {
                if n > 1 {
                        duplikat = append(duplikat, sep+" ("+itoa(n)+"×)")
                }
        }
        sort.Strings(duplikat)
        tulisJSON(w, 200, map[string]interface{}{
                "header": parse.Header, "delimiter": parse.Delimiter, "format": format,
                "totalBaris": len(parse.Baris), "barisValid": len(validasi.Valid),
                "barisGagal": len(validasi.Gagal), "gagalDetail": gagal,
                "dq": report, "duplikatSep": duplikat,
                "kolomTersedia": validasi.KolomTersedia,
        })
}

func handleUploadCommit(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehAkses("UPLOAD", sesi.Role) {
                tulisGagal(w, &RespError{403, "Hanya Super Admin dan Casemix yang dapat mengunggah data.", nil})
                return
        }
        var req ReqUpload
        if e := bacaBody(r, &req); e != nil {
                tulisGagal(w, e)
                return
        }
        if req.Override && !bolehOverrideDQ(sesi.Role) {
                tulisGagal(w, &RespError{403, "Override gerbang DQ hanya oleh Casemix/Super Admin.", nil})
                return
        }
        if req.Override && len(strings.TrimSpace(req.AlasanOverride)) < 10 {
                tulisGagal(w, &RespError{422, "Alasan override wajib diisi minimal 10 karakter.", nil})
                return
        }
        hasil, errResp := store.ingestUpload(req.HospitalID, req.Periode, req.NamaFile, req.Isi, sesi.ID, req.AlasanOverride, req.Override)
        if errResp != nil {
                tulisGagal(w, errResp)
                return
        }
        meta := fmt.Sprintf("baru=%d berubah=%d unchanged=%d", hasil.BarisBaru, hasil.BarisBerubah, hasil.BarisUnchanged)
        store.catatAudit(sesi.Email, "UPLOAD_COMMIT", "UPLOAD", &hasil.Upload.ID, &meta)
        store.simpan()
        tulisJSON(w, 200, hasil)
}

func handleUploads(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        uploads := []*UploadBatch{}
        for _, u := range store.Uploads {
                if u.Status != "DIHAPUS" {
                        if sesi.Role != "SUPERADMIN" && sesi.HospitalID != nil && u.HospitalID != *sesi.HospitalID {
                                continue
                        }
                        uploads = append(uploads, u)
                }
        }
        sort.Slice(uploads, func(i, j int) bool { return uploads[i].CreatedAt > uploads[j].CreatedAt })
        tulisJSON(w, 200, map[string]interface{}{"items": uploads})
}

func handleUploadArsip(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehAkses("UPLOAD", sesi.Role) {
                tulisGagal(w, &RespError{403, "Tidak berwenang.", nil})
                return
        }
        up := store.cariUpload(r.PathValue("id"))
        if up == nil {
                tulisGagal(w, &RespError{404, "Upload tidak ditemukan.", nil})
                return
        }
        up.Status = "ARSIP"
        store.catatAudit(sesi.Email, "UPLOAD_ARSIP", "UPLOAD", &up.ID, nil)
        store.simpan()
        tulisJSON(w, 200, up)
}

func handleUploadDelete(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehHapusUpload(sesi.Role) {
                tulisGagal(w, &RespError{403, "Hanya Super Admin yang dapat menghapus upload.", nil})
                return
        }
        up := store.cariUpload(r.PathValue("id"))
        if up == nil {
                tulisGagal(w, &RespError{404, "Upload tidak ditemukan.", nil})
                return
        }
        up.Status = "DIHAPUS"
        store.catatAudit(sesi.Email, "UPLOAD_DELETE", "UPLOAD", &up.ID, nil)
        store.simpan()
        tulisJSON(w, 200, map[string]bool{"ok": true})
}

// ---------- Ledger ----------

func ledgerPublik(l *LedgerEntry, denganEvent bool) map[string]interface{} {
        m := map[string]interface{}{
                "id": l.ID, "episodeId": l.EpisodeID, "hospitalId": l.HospitalID,
                "sep": l.Sep, "nomorMr": l.NomorMr, "namaPasien": l.NamaPasien,
                "kodeIcbg": l.KodeIcbg, "kelasRugi": l.KelasRugi, "estimasi": l.Estimasi,
                "realisasi": l.Realisasi, "status": l.Status, "referensiRm": l.ReferensiRm,
                "catatan": l.Catatan, "strategy": l.Strategy,
                "createdAt": l.CreatedAt, "updatedAt": l.UpdatedAt,
        }
        if denganEvent {
                events := []*LedgerEvent{}
                for _, ev := range store.LedgerEvents {
                        if ev.LedgerID == l.ID {
                                events = append(events, ev)
                        }
                }
                sort.Slice(events, func(i, j int) bool { return events[i].Waktu < events[j].Waktu })
                m["events"] = events
        }
        return m
}

func handleLedgerList(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehAkses("LEDGER", sesi.Role) {
                tulisGagal(w, &RespError{403, "Peran Anda tidak memiliki akses ke Recovery Ledger.", nil})
                return
        }
        q := r.URL.Query()
        status := q.Get("status")
        cari := strings.ToLower(strings.TrimSpace(q.Get("q")))
        items := []map[string]interface{}{}
        for _, l := range store.Ledger {
                if sesi.Role != "SUPERADMIN" && sesi.HospitalID != nil && l.HospitalID != *sesi.HospitalID {
                        continue
                }
                if status != "" && status != "SEMUA" && l.Status != status {
                        continue
                }
                if cari != "" && !strings.Contains(strings.ToLower(l.Sep), cari) &&
                        !strings.Contains(strings.ToLower(l.NamaPasien), cari) &&
                        !strings.Contains(strings.ToLower(l.NomorMr), cari) {
                        continue
                }
                items = append(items, ledgerPublik(l, false))
        }
        sort.Slice(items, func(i, j int) bool { return items[i]["updatedAt"].(string) > items[j]["updatedAt"].(string) })
        tulisJSON(w, 200, map[string]interface{}{"items": items, "kalibrasi": kalibrasiLedger(store.Ledger)})
}

func handleLedgerDetail(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehAkses("LEDGER", sesi.Role) {
                tulisGagal(w, &RespError{403, "Peran Anda tidak memiliki akses ke Recovery Ledger.", nil})
                return
        }
        for _, l := range store.Ledger {
                if l.ID == r.PathValue("id") {
                        tulisJSON(w, 200, ledgerPublik(l, true))
                        return
                }
        }
        tulisGagal(w, &RespError{404, "Entri ledger tidak ditemukan.", nil})
}

func handleLedgerBuat(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehMutasi("LEDGER", sesi.Role) {
                tulisGagal(w, &RespError{403, "Peran Anda hanya dapat melihat ledger.", nil})
                return
        }
        var body struct {
                EpisodeID string `json:"episodeId"`
        }
        if e := bacaBody(r, &body); e != nil {
                tulisGagal(w, e)
                return
        }
        ep := store.cariEpisode(body.EpisodeID)
        if ep == nil {
                tulisGagal(w, &RespError{404, "Episode tidak ditemukan.", nil})
                return
        }
        if ep.KelasRugi == "SURPLUS" {
                tulisGagal(w, &RespError{422, "Episode surplus tidak menjadi kandidat recovery.", nil})
                return
        }
        for _, l := range store.Ledger {
                if l.Sep == ep.NoSep {
                        tulisGagal(w, &RespError{409, "SEP sudah terdaftar pada ledger.", nil})
                        return
                }
        }
        now := sekarangISO()
        estimasi := int64(math.Abs(float64(ep.Selisih)) / 2)
        strategi := strategiUntuk(ep)
        hospitalID := sesi.HospitalID
        if sesi.Role == "SUPERADMIN" {
                if u := store.cariUpload(ep.UploadBatchID); u != nil {
                        hospitalID = &u.HospitalID
                }
        }
        l := &LedgerEntry{
                ID: idBaru("lg"), EpisodeID: ep.ID, HospitalID: *hospitalID,
                Sep: ep.NoSep, NomorMr: ep.NoMr, NamaPasien: ep.NamaPasien,
                KodeIcbg: ep.KodeIcbg, KelasRugi: ep.KelasRugi, Estimasi: estimasi,
                Status: "KANDIDAT", Strategy: &strategi, CreatedAt: now, UpdatedAt: now,
        }
        store.Ledger = append(store.Ledger, l)
        store.catatAudit(sesi.Email, "LEDGER_CREATE", "LEDGER", &l.ID, nil)
        store.simpan()
        tulisJSON(w, 200, ledgerPublik(l, false))
}

func handleLedgerTransisi(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehMutasi("LEDGER", sesi.Role) {
                tulisGagal(w, &RespError{403, "Peran Anda hanya dapat melihat ledger (Direktur baca-saja).", nil})
                return
        }
        var body struct {
                Ke          string `json:"ke"`
                Alasan      string `json:"alasan"`
                ReferensiRm string `json:"referensiRm"`
                Realisasi   *int64 `json:"realisasi"`
        }
        if e := bacaBody(r, &body); e != nil {
                tulisGagal(w, e)
                return
        }
        var l *LedgerEntry
        for _, x := range store.Ledger {
                if x.ID == r.PathValue("id") {
                        l = x
                        break
                }
        }
        if l == nil {
                tulisGagal(w, &RespError{404, "Entri ledger tidak ditemukan.", nil})
                return
        }
        hasil := transisiLedger(PermintaanTransisi{
                Dari: l.Status, Ke: body.Ke, Peran: sesi.Role,
                ReferensiRm: body.ReferensiRm, Alasan: body.Alasan, Realisasi: body.Realisasi,
        })
        if !hasil.OK {
                tulisGagal(w, &RespError{422, hasil.Alasan, hasil})
                return
        }
        // HANYA aktor manusia yang boleh menuju DIAJUKAN/HASIL — di sini aktor adalah pengguna.
        var catatan *string
        if body.Alasan != "" {
                catatan = &body.Alasan
        }
        store.pindahLedger(l, body.Ke, body.Realisasi, sesi.Nama, catatan, sekarangISO())
        if body.ReferensiRm != "" {
                l.ReferensiRm = &body.ReferensiRm
        }
        store.catatAudit(sesi.Email, "LEDGER_TRANSISI", "LEDGER", &l.ID, strPtr(l.Status+" → "+body.Ke))
        store.simpan()
        tulisJSON(w, 200, ledgerPublik(l, true))
}

// ---------- Queries ----------

func queryPublik(q *QueryDok) map[string]interface{} {
        m := map[string]interface{}{
                "id": q.ID, "hospitalId": q.HospitalID, "sep": q.Sep, "nomorMr": q.NomorMr,
                "inisial": q.Inisial, "tglMasuk": q.TglMasuk, "tglKeluar": q.TglKeluar,
                "dpjpId": q.DpjpID, "dpjpNama": q.DpjpNama, "dpjpSpesialisasi": q.DpjpSpesialisasi,
                "templateNama": q.TemplateNama, "pertanyaan": q.Pertanyaan, "konteks": q.Konteks,
                "status": q.Status, "deadline": q.Deadline, "sentAt": q.SentAt,
                "answeredAt": q.AnsweredAt, "jawaban": q.Jawaban, "slaHari": q.SlaHari,
                "createdAt": q.CreatedAt,
        }
        if q.Status == "TERKIRIM" && q.Deadline < sekarangISO() {
                m["status"] = "TERLAMBAT"
        }
        return m
}

func handleQueries(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehAkses("QUERY", sesi.Role) && sesi.Role != "DPJP" {
                tulisGagal(w, &RespError{403, "Peran Anda tidak memiliki akses modul query.", nil})
                return
        }
        status := r.URL.Query().Get("status")
        items := []map[string]interface{}{}
        for _, q := range store.Queries {
                if sesi.Role == "DPJP" && q.DpjpID != sesi.ID {
                        continue
                }
                if sesi.Role != "SUPERADMIN" && sesi.HospitalID != nil && q.HospitalID != *sesi.HospitalID {
                        continue
                }
                pub := queryPublik(q)
                if status != "" && status != "SEMUA" && pub["status"] != status {
                        continue
                }
                items = append(items, pub)
        }
        sort.Slice(items, func(i, j int) bool { return items[i]["createdAt"].(string) > items[j]["createdAt"].(string) })
        tulisJSON(w, 200, map[string]interface{}{"items": items})
}

func handleQueryBuat(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehMutasi("QUERY", sesi.Role) {
                tulisGagal(w, &RespError{403, "Peran Anda tidak dapat membuat query.", nil})
                return
        }
        var body struct {
                Sep        string `json:"sep"`
                DpjpID     string `json:"dpjpId"`
                TemplateID string `json:"templateId"`
                Pertanyaan string `json:"pertanyaan"`
                Konteks    string `json:"konteks"`
        }
        if e := bacaBody(r, &body); e != nil {
                tulisGagal(w, e)
                return
        }
        var ep *Episode
        for _, x := range store.Episodes {
                if x.NoSep == strings.TrimSpace(body.Sep) && !x.Superseded {
                        ep = x
                        break
                }
        }
        if ep == nil {
                tulisGagal(w, &RespError{404, "SEP tidak ditemukan pada data aktif.", nil})
                return
        }
        var dpjp *User
        for _, u := range store.Users {
                if u.ID == body.DpjpID && u.Role == "DPJP" {
                        dpjp = u
                        break
                }
        }
        if dpjp == nil {
                tulisGagal(w, &RespError{404, "DPJP tidak ditemukan.", nil})
                return
        }
        if ep.DpjpKode == nil || dpjp.DpjpKode == nil || *ep.DpjpKode != *dpjp.DpjpKode {
                tulisGagal(w, &RespError{422, "DPJP yang dipilih bukan DPJP pada episode ini.", nil})
                return
        }
        pertanyaan := strings.TrimSpace(body.Pertanyaan)
        var tplNama *string
        if body.TemplateID != "" {
                for _, t := range store.Templates {
                        if t.ID == body.TemplateID {
                                tplNama = &t.Nama
                                if pertanyaan == "" {
                                        pertanyaan = t.Isi
                                }
                        }
                }
        }
        if pertanyaan == "" {
                tulisGagal(w, &RespError{422, "Pertanyaan wajib diisi.", nil})
                return
        }
        slaHari := 3
        for _, st := range store.Settings {
                if st.Key == "sla.query_hari" {
                        slaHari, _ = strconv.Atoi(st.Value)
                }
        }
        now := time.Now().UTC()
        deadline := now.AddDate(0, 0, slaHari)
        q := &QueryDok{
                ID: idBaru("qr"), HospitalID: sesiRoleHospital(sesi, ep), Sep: ep.NoSep,
                NomorMr: ep.NoMr, Inisial: inisial(ep.NamaPasien),
                TglMasuk: &ep.TglMasuk, TglKeluar: &ep.TglKeluar,
                DpjpID: dpjp.ID, DpjpNama: dpjp.Nama, DpjpSpesialisasi: dpjp.Spesialisasi,
                TemplateNama: tplNama, Pertanyaan: pertanyaan, Konteks: body.Konteks,
                Status: "TERKIRIM", Deadline: deadline.Format("2006-01-02T15:04:05.000Z"),
                SentAt: strPtr(now.Format("2006-01-02T15:04:05.000Z")),
                SlaHari: slaHari, CreatedAt: now.Format("2006-01-02T15:04:05.000Z"),
        }
        store.Queries = append(store.Queries, q)
        store.catatAudit(sesi.Email, "QUERY_CREATE", "QUERY", &q.ID, &q.Sep)
        store.simpan()
        tulisJSON(w, 200, queryPublik(q))
}

func sesiRoleHospital(sesi *Sesi, ep *Episode) string {
        if sesi.HospitalID != nil {
                return *sesi.HospitalID
        }
        if u := store.cariUpload(ep.UploadBatchID); u != nil {
                return u.HospitalID
        }
        return "rs1"
}

func handleQueryJawab(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        var q *QueryDok
        for _, x := range store.Queries {
                if x.ID == r.PathValue("id") {
                        q = x
                        break
                }
        }
        if q == nil {
                tulisGagal(w, &RespError{404, "Query tidak ditemukan.", nil})
                return
        }
        // Jawaban terkunci setelah DIJAWAB; DPJP hanya query miliknya
        if q.Status == "DIJAWAB" {
                tulisGagal(w, &RespError{409, "Query sudah dijawab — jawaban terkunci.", nil})
                return
        }
        if sesi.Role == "DPJP" && q.DpjpID != sesi.ID {
                tulisGagal(w, &RespError{403, "Query ini bukan untuk DPJP Anda.", nil})
                return
        }
        if sesi.Role != "DPJP" && !bolehMutasi("QUERY", sesi.Role) {
                tulisGagal(w, &RespError{403, "Peran Anda tidak dapat menjawab query.", nil})
                return
        }
        var body struct {
                Jawaban string `json:"jawaban"`
        }
        if e := bacaBody(r, &body); e != nil {
                tulisGagal(w, e)
                return
        }
        if len(strings.TrimSpace(body.Jawaban)) < 5 {
                tulisGagal(w, &RespError{422, "Jawaban minimal 5 karakter.", nil})
                return
        }
        now := sekarangISO()
        q.Status = "DIJAWAB"
        q.AnsweredAt = &now
        q.Jawaban = &body.Jawaban
        store.catatAudit(sesi.Email, "QUERY_ANSWER", "QUERY", &q.ID, nil)
        store.simpan()
        tulisJSON(w, 200, queryPublik(q))
}

func handleTemplates(w http.ResponseWriter, r *http.Request) {
        _, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        tulisJSON(w, 200, map[string]interface{}{"items": store.Templates})
}

// ---------- Admin ----------

func handleAdminUsers(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehAkses("ADMIN", sesi.Role) {
                tulisGagal(w, &RespError{403, "Hanya Super Admin yang dapat mengelola pengguna.", nil})
                return
        }
        tulisJSON(w, 200, map[string]interface{}{"items": usersPublik(store.Users)})
}

func handleAdminUserBuat(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehMutasi("ADMIN", sesi.Role) {
                tulisGagal(w, &RespError{403, "Hanya Super Admin yang dapat menambah pengguna.", nil})
                return
        }
        var body struct {
                Email        string  `json:"email"`
                Nama         string  `json:"nama"`
                Role         string  `json:"role"`
                HospitalID   *string `json:"hospitalId"`
                Spesialisasi *string `json:"spesialisasi"`
                DpjpKode     *string `json:"dpjpKode"`
        }
        if e := bacaBody(r, &body); e != nil {
                tulisGagal(w, e)
                return
        }
        email := strings.ToLower(strings.TrimSpace(body.Email))
        if email == "" || body.Nama == "" || !containsStr(SEMUA_PERAN, body.Role) {
                tulisGagal(w, &RespError{422, "Email, nama, dan peran wajib diisi dengan benar.", nil})
                return
        }
        for _, u := range store.Users {
                if strings.ToLower(u.Email) == email {
                        tulisGagal(w, &RespError{409, "Email sudah terdaftar.", nil})
                        return
                }
        }
        u := &User{
                ID: idBaru("u"), Email: email, PasswordHash: hashSandi("MedClaim#demo2026"),
                Nama: body.Nama, Role: body.Role, HospitalID: body.HospitalID,
                Spesialisasi: body.Spesialisasi, DpjpKode: body.DpjpKode,
                Aktif: true, CreatedAt: sekarangISO(),
        }
        store.Users = append(store.Users, u)
        store.catatAudit(sesi.Email, "USER_CREATE", "USER", &u.ID, &u.Email)
        store.simpan()
        tulisJSON(w, 200, usersPublik([]*User{u})[0])
}

func handleAdminUserToggle(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehMutasi("ADMIN", sesi.Role) {
                tulisGagal(w, &RespError{403, "Hanya Super Admin yang dapat mengubah status pengguna.", nil})
                return
        }
        u := store.cariUser(r.PathValue("id"))
        if u == nil {
                tulisGagal(w, &RespError{404, "Pengguna tidak ditemukan.", nil})
                return
        }
        if u.ID == sesi.ID {
                tulisGagal(w, &RespError{422, "Tidak dapat menonaktifkan akun sendiri.", nil})
                return
        }
        u.Aktif = !u.Aktif
        store.catatAudit(sesi.Email, "USER_TOGGLE", "USER", &u.ID, strPtr(itoaBool(u.Aktif)))
        store.simpan()
        tulisJSON(w, 200, usersPublik([]*User{u})[0])
}

func itoaBool(b bool) string {
        if b {
                return "aktif"
        }
        return "nonaktif"
}

func handleAdminAudit(w http.ResponseWriter, r *http.Request) {
        sesi, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        if !bolehAkses("ADMIN", sesi.Role) {
                tulisGagal(w, &RespError{403, "Hanya Super Admin yang dapat melihat jejak audit.", nil})
                return
        }
        audit := append([]*AuditLog(nil), store.Audit...)
        sort.Slice(audit, func(i, j int) bool { return audit[i].Waktu > audit[j].Waktu })
        if len(audit) > 200 {
                audit = audit[:200]
        }
        tulisJSON(w, 200, map[string]interface{}{"items": audit})
}

// ---------- Sampel (file demo) ----------

func handleSamples(w http.ResponseWriter, r *http.Request) {
        _, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        deskripsi := map[string]string{
                "vclaim-basic.txt":    "Format Basic (14 kolom) — periode 2025-10, 150 baris",
                "vclaim-detail.txt":   "Format Detail (41 kolom) — periode 2025-11, 150 baris",
                "vclaim-detail-v2.txt": "Revisi Detail (6 baris berubah: LOS & severity naik) — uji versioning",
                "vclaim-broken.txt":   "File rusak: duplikat SEP, tanggal 31-02, tanggal terbalik, tarif 0 — uji gerbang DQ",
        }
        items := []map[string]interface{}{}
        entries, _ := bacaDir(direktoriSample)
        for _, en := range entries {
                if en.IsDir() || !strings.HasSuffix(en.Name(), ".txt") {
                        continue
                }
                info, _ := en.Info()
                items = append(items, map[string]interface{}{
                        "nama": en.Name(), "ukuran": info.Size(), "deskripsi": deskripsi[en.Name()],
                })
        }
        sort.Slice(items, func(i, j int) bool { return items[i]["nama"].(string) < items[j]["nama"].(string) })
        tulisJSON(w, 200, map[string]interface{}{"items": items})
}

func handleSampleIsi(w http.ResponseWriter, r *http.Request) {
        _, e := butuhSesi(r)
        if e != nil {
                tulisGagal(w, e)
                return
        }
        nama := filepath.Base(r.PathValue("nama"))
        if !strings.HasSuffix(nama, ".txt") {
                tulisGagal(w, &RespError{400, "Hanya file .txt yang diizinkan.", nil})
                return
        }
        isi, err := bacaFile(filepath.Join(direktoriSample, nama))
        if err != nil {
                tulisGagal(w, &RespError{404, "Sampel tidak ditemukan.", nil})
                return
        }
        tulisJSON(w, 200, map[string]string{"nama": nama, "isi": string(isi)})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
        tulisJSON(w, 200, map[string]string{"status": "ok", "layanan": "medclaim-go", "waktu": sekarangISO()})
}
