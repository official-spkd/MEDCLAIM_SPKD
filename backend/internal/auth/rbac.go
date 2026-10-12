package auth

import (
	"slices"

	"medclaim-backend/internal/response"
	"medclaim-backend/internal/types"
)

// Grup akses fitur (cermin MATRIX_AKSES di mock dan store.ts di frontend).
const (
	GrupAnalitik = "ANALITIK"
	GrupUpload   = "UPLOAD"
	GrupLedger   = "LEDGER"
	GrupQuery    = "QUERY"
	GrupAdmin    = "ADMIN"
)

var matriks = map[string][]string{
	GrupAnalitik: {"SUPERADMIN", "CASEMIX", "CODER", "DIREKTUR"},
	GrupUpload:   {"SUPERADMIN", "CASEMIX"},
	GrupLedger:   {"SUPERADMIN", "CASEMIX", "DIREKTUR"},
	GrupQuery:    {"SUPERADMIN", "CASEMIX", "CODER"},
	GrupAdmin:    {"SUPERADMIN"},
}

func BolehAkses(grup, peran string) bool { return slices.Contains(matriks[grup], peran) }

// BolehMutasi: akses tulis. LEDGER hanya Casemix/Superadmin (Direktur baca-saja).
func BolehMutasi(grup, peran string) bool {
	if !BolehAkses(grup, peran) {
		return false
	}
	switch grup {
	case GrupLedger:
		return peran == "SUPERADMIN" || peran == "CASEMIX"
	case GrupAdmin:
		return peran == "SUPERADMIN"
	}
	return true
}

func IsCasemix(peran string) bool           { return peran == "SUPERADMIN" || peran == "CASEMIX" }
func BolehOverrideDQ(peran string) bool     { return IsCasemix(peran) }
func BolehLihatFinansial(peran string) bool { return peran != "DPJP" }

func (p *Principal) IsSuperadmin() bool { return p.Role == "SUPERADMIN" }

// ScopeHospital menentukan RS yang boleh diakses untuk request ini.
// Non-superadmin SELALU terkunci ke RS miliknya (parameter rs diabaikan);
// superadmin boleh memilih RS (string kosong = semua RS).
func (p *Principal) ScopeHospital(requested string) (string, error) {
	if !p.IsSuperadmin() {
		if p.HospitalID == nil {
			return "", response.Forbidden("Akun belum terikat ke rumah sakit.")
		}
		return *p.HospitalID, nil
	}
	if requested != "" && !types.IsUUID(requested) {
		return "", response.BadRequest("Parameter rs tidak valid.")
	}
	return requested, nil
}
