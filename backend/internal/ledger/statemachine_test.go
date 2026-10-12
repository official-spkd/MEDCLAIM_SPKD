package ledger

import "testing"

func i64(n int64) *int64 { return &n }

func TestCekTransisi(t *testing.T) {
	panjang := "alasan yang cukup panjang"
	cases := []struct {
		nama string
		req  PermintaanTransisi
		ok   bool
	}{
		{"kandidat -> diperiksa oleh coder", PermintaanTransisi{Dari: "KANDIDAT", Ke: "DIPERIKSA", Peran: "CODER"}, true},
		{"kandidat -> hasil loncat status", PermintaanTransisi{Dari: "KANDIDAT", Ke: "HASIL", Peran: "CASEMIX", Realisasi: i64(1)}, false},
		{"dikoreksi tanpa referensi RM", PermintaanTransisi{Dari: "DIPERIKSA", Ke: "DIKOREKSI", Peran: "CODER"}, false},
		{"dikoreksi dengan referensi RM", PermintaanTransisi{Dari: "DIPERIKSA", Ke: "DIKOREKSI", Peran: "CODER", ReferensiRm: "RM-1"}, true},
		{"diajukan oleh coder ditolak", PermintaanTransisi{Dari: "DIKOREKSI", Ke: "DIAJUKAN", Peran: "CODER"}, false},
		{"diajukan oleh casemix", PermintaanTransisi{Dari: "DIKOREKSI", Ke: "DIAJUKAN", Peran: "CASEMIX"}, true},
		{"hasil tanpa realisasi", PermintaanTransisi{Dari: "DIAJUKAN", Ke: "HASIL", Peran: "CASEMIX"}, false},
		{"hasil realisasi 0 sah", PermintaanTransisi{Dari: "DIAJUKAN", Ke: "HASIL", Peran: "CASEMIX", Realisasi: i64(0)}, true},
		{"hasil oleh coder ditolak", PermintaanTransisi{Dari: "DIAJUKAN", Ke: "HASIL", Peran: "CODER", Realisasi: i64(5)}, false},
		{"tidak berubah tanpa alasan", PermintaanTransisi{Dari: "DIPERIKSA", Ke: "TIDAK_BERUBAH", Peran: "CASEMIX"}, false},
		{"tidak berubah dengan alasan", PermintaanTransisi{Dari: "DIPERIKSA", Ke: "TIDAK_BERUBAH", Peran: "CASEMIX", Alasan: panjang}, true},
		{"reopen terminal oleh coder ditolak", PermintaanTransisi{Dari: "HASIL", Ke: "DIPERIKSA", Peran: "CODER", Alasan: panjang}, false},
		{"reopen terminal tanpa alasan", PermintaanTransisi{Dari: "HASIL", Ke: "DIPERIKSA", Peran: "CASEMIX"}, false},
		{"reopen terminal sah", PermintaanTransisi{Dari: "HASIL", Ke: "DIPERIKSA", Peran: "SUPERADMIN", Alasan: panjang}, true},
		{"direktur tidak boleh memeriksa", PermintaanTransisi{Dari: "KANDIDAT", Ke: "DIPERIKSA", Peran: "DIREKTUR"}, false},
		{"status asal tidak dikenal", PermintaanTransisi{Dari: "ENTAH", Ke: "DIPERIKSA", Peran: "CASEMIX"}, false},
	}
	for _, c := range cases {
		if got := CekTransisi(c.req); got.OK != c.ok {
			t.Errorf("%s: OK=%v (alasan: %q), harusnya %v", c.nama, got.OK, got.Alasan, c.ok)
		}
	}
}

func TestKalibrasi(t *testing.T) {
	var es []Entry
	for i := 0; i < 30; i++ {
		es = append(es, Entry{Estimasi: 100, Realisasi: i64(80), Status: "HASIL"})
	}
	es = append(es, Entry{Estimasi: 50, Realisasi: i64(0), Status: "TIDAK_BERUBAH"}, Entry{Estimasi: 50, Status: "KANDIDAT"})
	k := HitungKalibrasi(es)
	if k.NSampel != 31 || !k.KalibrasiTersedia {
		t.Errorf("sampel harus 31 dan kalibrasi tersedia: %+v", k)
	}
	if k.RealizationRate != 0.8 {
		t.Errorf("realization rate harus 0.8, dapat %v", k.RealizationRate)
	}
	if k.HitRate < 0.96 || k.HitRate > 0.97 {
		t.Errorf("hit rate harus 30/31, dapat %v", k.HitRate)
	}
}
