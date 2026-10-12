// ============================================================
// MedClaim Go — State machine Recovery Ledger + RBAC 5 peran
// (port dari state-machine.ts & rbac.ts)
// ============================================================
package main

import "strings"

var StatusLedgerUrut = []string{
	"KANDIDAT", "DIPERIKSA", "DIKOREKSI", "DIAJUKAN", "HASIL", "TIDAK_BERUBAH",
}

var sah = map[string][]string{
	"KANDIDAT":      {"DIPERIKSA"},
	"DIPERIKSA":     {"DIKOREKSI", "TIDAK_BERUBAH", "KANDIDAT"},
	"DIKOREKSI":     {"DIAJUKAN", "DIPERIKSA"},
	"DIAJUKAN":      {"HASIL", "DIKOREKSI"},
	"HASIL":         {"DIPERIKSA"},
	"TIDAK_BERUBAH": {"DIPERIKSA"},
}

type HasilTransisi struct {
	OK               bool   `json:"ok"`
	Alasan           string `json:"alasan,omitempty"`
	PerluReferensiRm bool   `json:"perluReferensiRm,omitempty"`
	PerluAlasan      bool   `json:"perluAlasan,omitempty"`
	PerluRealisasi   bool   `json:"perluRealisasi,omitempty"`
}

type PermintaanTransisi struct {
	Dari        string
	Ke          string
	Peran       string
	ReferensiRm string
	Alasan      string
	Realisasi   *int64
}

func isCasemix(peran string) bool {
	return peran == "SUPERADMIN" || peran == "CASEMIX"
}

func transisiLedger(req PermintaanTransisi) HasilTransisi {
	dari, ke, peran := req.Dari, req.Ke, req.Peran
	izin, ada := sah[dari]
	if !ada {
		return HasilTransisi{OK: false, Alasan: "Status asal tidak dikenal."}
	}
	if !containsStr(izin, ke) {
		return HasilTransisi{OK: false, Alasan: "Transisi " + dari + " → " + ke + " tidak sah. Transisi yang diizinkan: " + strings.Join(izin, ", ") + "."}
	}

	switch ke {
	case "DIPERIKSA":
		if !containsStr([]string{"SUPERADMIN", "CASEMIX", "CODER"}, peran) {
			return HasilTransisi{OK: false, Alasan: "Hanya Casemix, Coder, atau Super Admin yang dapat memeriksa kandidat."}
		}
		if (dari == "HASIL" || dari == "TIDAK_BERUBAH") && !isCasemix(peran) {
			return HasilTransisi{OK: false, Alasan: "Reopen status terminal hanya oleh Casemix/Super Admin."}
		}
		if (dari == "HASIL" || dari == "TIDAK_BERUBAH") && len(strings.TrimSpace(req.Alasan)) < 10 {
			return HasilTransisi{OK: false, Alasan: "Reopen status terminal wajib menyertakan alasan (minimal 10 karakter).", PerluAlasan: true}
		}
		return HasilTransisi{OK: true}

	case "TIDAK_BERUBAH":
		if !isCasemix(peran) {
			return HasilTransisi{OK: false, Alasan: "Penetapan Tidak Berubah hanya oleh Casemix/Super Admin."}
		}
		if len(strings.TrimSpace(req.Alasan)) < 10 {
			return HasilTransisi{OK: false, Alasan: "Wajib menyertakan alasan penetapan (minimal 10 karakter).", PerluAlasan: true}
		}
		return HasilTransisi{OK: true}

	case "DIKOREKSI":
		if !containsStr([]string{"SUPERADMIN", "CASEMIX", "CODER"}, peran) {
			return HasilTransisi{OK: false, Alasan: "Hanya Casemix, Coder, atau Super Admin yang dapat menandai dikoreksi."}
		}
		if len(strings.TrimSpace(req.ReferensiRm)) < 3 {
			return HasilTransisi{OK: false, Alasan: "Referensi rekam medis wajib diisi (nomor RM / resume / tautan).", PerluReferensiRm: true}
		}
		return HasilTransisi{OK: true}

	case "DIAJUKAN":
		if !isCasemix(peran) {
			return HasilTransisi{OK: false, Alasan: "Pengajuan hanya oleh Casemix/Super Admin."}
		}
		return HasilTransisi{OK: true}

	case "HASIL":
		if !isCasemix(peran) {
			return HasilTransisi{OK: false, Alasan: "Konfirmasi HASIL hanya oleh Casemix/Super Admin (prinsip empat mata)."}
		}
		if req.Realisasi == nil {
			return HasilTransisi{OK: false, Alasan: "Nilai realisasi wajib diisi saat konfirmasi HASIL.", PerluRealisasi: true}
		}
		return HasilTransisi{OK: true}

	case "KANDIDAT":
		return HasilTransisi{OK: true}
	}
	return HasilTransisi{OK: false, Alasan: "Status tujuan tidak dikenal."}
}

// transisiSistemDiizinkan — sistem tidak pernah boleh set DIAJUKAN/HASIL otomatis.
func transisiSistemDiizinkan(ke string) bool {
	return ke != "DIAJUKAN" && ke != "HASIL"
}

func containsStr(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}

// ---------- RBAC ----------

var SEMUA_PERAN = []string{"SUPERADMIN", "CASEMIX", "CODER", "DIREKTUR", "DPJP"}

var LabelPeran = map[string]string{
	"SUPERADMIN": "Super Admin",
	"CASEMIX":    "Casemix Manager",
	"CODER":      "Medical Coder",
	"DIREKTUR":   "Direktur RS",
	"DPJP":       "DPJP",
}

var MATRIX_AKSES = map[string][]string{
	"ANALITIK":   {"SUPERADMIN", "CASEMIX", "CODER", "DIREKTUR"},
	"UPLOAD":     {"SUPERADMIN", "CASEMIX"},
	"LEDGER":     {"SUPERADMIN", "CASEMIX", "DIREKTUR"},
	"QUERY":      {"SUPERADMIN", "CASEMIX", "CODER"},
	"ADMIN":      {"SUPERADMIN"},
	"DOKUMENTASI": SEMUA_PERAN,
}

func bolehAkses(grup, peran string) bool {
	return containsStr(MATRIX_AKSES[grup], peran)
}

func bolehMutasi(grup, peran string) bool {
	if !bolehAkses(grup, peran) {
		return false
	}
	switch grup {
	case "LEDGER":
		return peran == "SUPERADMIN" || peran == "CASEMIX"
	case "ADMIN":
		return peran == "SUPERADMIN"
	default:
		return true
	}
}

func bolehKonfirmasiHasil(peran string) bool {
	return peran == "SUPERADMIN" || peran == "CASEMIX"
}

func bolehOverrideDQ(peran string) bool {
	return peran == "SUPERADMIN" || peran == "CASEMIX"
}

func bolehHapusUpload(peran string) bool {
	return peran == "SUPERADMIN"
}

func rsTerkses(peran string) bool {
	return peran == "SUPERADMIN"
}

func bolehLihatFinansial(peran string) bool {
	return peran != "DPJP"
}
