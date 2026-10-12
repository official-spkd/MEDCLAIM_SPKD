package ledger

import (
	"slices"
	"strings"

	"medclaim-backend/internal/auth"
)

// Transisi status yang sah (mesin status Recovery Ledger).
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
	Dari, Ke, Peran string
	ReferensiRm     string
	Alasan          string
	Realisasi       *int64
}

func gagal(alasan string) HasilTransisi { return HasilTransisi{Alasan: alasan} }

// CekTransisi menerapkan aturan transisi + peran + kelengkapan data.
func CekTransisi(req PermintaanTransisi) HasilTransisi {
	izin, ada := sah[req.Dari]
	if !ada {
		return gagal("Status asal tidak dikenal.")
	}
	if !slices.Contains(izin, req.Ke) {
		return gagal("Transisi " + req.Dari + " → " + req.Ke + " tidak sah. Transisi yang diizinkan: " + strings.Join(izin, ", ") + ".")
	}
	peran := req.Peran
	terminal := req.Dari == "HASIL" || req.Dari == "TIDAK_BERUBAH"
	alasanCukup := len(strings.TrimSpace(req.Alasan)) >= 10

	switch req.Ke {
	case "DIPERIKSA":
		if !slices.Contains([]string{"SUPERADMIN", "CASEMIX", "CODER"}, peran) {
			return gagal("Hanya Casemix, Coder, atau Super Admin yang dapat memeriksa kandidat.")
		}
		if terminal && !auth.IsCasemix(peran) {
			return gagal("Reopen status terminal hanya oleh Casemix/Super Admin.")
		}
		if terminal && !alasanCukup {
			return HasilTransisi{Alasan: "Reopen status terminal wajib menyertakan alasan (minimal 10 karakter).", PerluAlasan: true}
		}
	case "TIDAK_BERUBAH":
		if !auth.IsCasemix(peran) {
			return gagal("Penetapan Tidak Berubah hanya oleh Casemix/Super Admin.")
		}
		if !alasanCukup {
			return HasilTransisi{Alasan: "Wajib menyertakan alasan penetapan (minimal 10 karakter).", PerluAlasan: true}
		}
	case "DIKOREKSI":
		if !slices.Contains([]string{"SUPERADMIN", "CASEMIX", "CODER"}, peran) {
			return gagal("Hanya Casemix, Coder, atau Super Admin yang dapat menandai dikoreksi.")
		}
		if len(strings.TrimSpace(req.ReferensiRm)) < 3 {
			return HasilTransisi{Alasan: "Referensi rekam medis wajib diisi (nomor RM / resume / tautan).", PerluReferensiRm: true}
		}
	case "DIAJUKAN":
		if !auth.IsCasemix(peran) {
			return gagal("Pengajuan hanya oleh Casemix/Super Admin.")
		}
	case "HASIL":
		if !auth.IsCasemix(peran) {
			return gagal("Konfirmasi HASIL hanya oleh Casemix/Super Admin (prinsip empat mata).")
		}
		if req.Realisasi == nil {
			return HasilTransisi{Alasan: "Nilai realisasi wajib diisi saat konfirmasi HASIL.", PerluRealisasi: true}
		}
	case "KANDIDAT":
		// dari DIPERIKSA saja (sudah dicek di tabel transisi)
	default:
		return gagal("Status tujuan tidak dikenal.")
	}
	return HasilTransisi{OK: true}
}
