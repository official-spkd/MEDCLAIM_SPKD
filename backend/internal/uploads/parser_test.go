package uploads

import "testing"

func TestAngkaBersih(t *testing.T) {
	cases := map[string]float64{
		"2450000":      2450000,
		"2.450.000":    2450000,
		"2,450,000":    2450000,
		"1.234.567,89": 1234567.89,
		"1,234,567.89": 1234567.89,
		"Rp 2.450.000": 2450000,
		"":             0,
		"-":            0,
		"NONE":         0,
		"abc":          0,
	}
	for in, want := range cases {
		if got := angkaBersih(in); got != want {
			t.Errorf("angkaBersih(%q) = %v, harusnya %v", in, got, want)
		}
	}
}

func TestParseDelimiterDanAlias(t *testing.T) {
	csv := "No.SEP;Nama;Tgl. Masuk;Tgl. Keluar;Kode INACBG;Diagnosa\n1;Budi;2025-10-01;2025-10-03;I-4-08;K29.7\n"
	p := parseTSV(csv)
	if p.Delimiter != ";" || len(p.Baris) != 1 {
		t.Fatalf("delimiter/baris salah: %q %d", p.Delimiter, len(p.Baris))
	}
	if p.Baris[0].Kolom["NOSEP"] != "1" || p.Baris[0].Kolom["NAMAPASIEN"] != "Budi" {
		t.Errorf("alias kolom tidak terpetakan: %+v", p.Baris[0].Kolom)
	}
	if p := parseTSV("\xef\xbb\xbfNOSEP\tDIAGNOSA\n1\tA09\n"); p.Header[0] != "NOSEP" {
		t.Errorf("BOM harus dibuang: %q", p.Header[0])
	}
}

func TestValidasiBarisMembuangBarisCacat(t *testing.T) {
	in := "NOSEP\tTGLMASUK\tTGLKELUAR\tKODEICBG\tDIAGNOSA\tTARIFINACBG\n" +
		"\t2025-10-01\t2025-10-02\tI-4-08\tA09\t100\n" + // tanpa SEP
		"S2\t2025-10-01\t2025-10-02\t\tA09\t100\n" + // tanpa CBG
		"S3\t2025-10-01\t2025-10-02\tI-4-08\t\t100\n" + // tanpa diagnosa
		"S4\t2025-10-01\t2025-10-02\tI-4-08\tA09;B34.9\t1.000\n" // sah
	v := validasiBaris(parseTSV(in))
	if len(v.Valid) != 1 || len(v.Gagal) != 3 {
		t.Fatalf("valid=%d gagal=%d", len(v.Valid), len(v.Gagal))
	}
	b := v.Valid[0]
	if b.NoSep != "S4" || len(b.Diagnosa) != 2 || b.TarifInacbg != 1000 || b.Los != 1 {
		t.Errorf("baris sah salah terparse: %+v", b)
	}
}

func contoh(t *testing.T, nama string) string {
	t.Helper()
	b, err := samplesFS.ReadFile("samples/" + nama)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func laporan(t *testing.T, nama, periode string) DQReport {
	t.Helper()
	v := validasiBaris(parseTSV(contoh(t, nama)))
	return buatLaporanDQ(v.Valid, KonteksDQ{KolomTersedia: v.KolomTersedia, PeriodeDipilih: periode, SepEksisting: map[string]bool{}})
}

func TestGerbangDQAtasFileContoh(t *testing.T) {
	if r := laporan(t, "vclaim-detail.txt", "2025-11"); r.Gerbang != "LOLOS" || r.Skor < 80 {
		t.Errorf("file detail harus lolos: skor %d gerbang %s", r.Skor, r.Gerbang)
	}
	if r := laporan(t, "vclaim-basic.txt", "2025-10"); r.Skor != 40 || r.Gerbang != "BLOKIR" {
		t.Errorf("file basic harus skor 40 & BLOKIR: skor %d gerbang %s", r.Skor, r.Gerbang)
	}
	r := laporan(t, "vclaim-broken.txt", "2025-12")
	if r.Gerbang != "BLOKIR" {
		t.Errorf("file rusak harus BLOKIR, dapat %s", r.Gerbang)
	}
	terpicu := map[string]bool{}
	for _, a := range r.Aturan {
		if a.Terpicu {
			terpicu[a.Kode] = true
		}
	}
	for _, kode := range []string{"DQ-UNQ-01", "DQ-VAL-02"} {
		if !terpicu[kode] {
			t.Errorf("aturan %s seharusnya terpicu pada file rusak", kode)
		}
	}
}

func TestBarisSiapDedupDanTanggal(t *testing.T) {
	baris := []BarisValid{
		{NoSep: "A", TglMasuk: "2025-10-01", TglKeluar: "2025-10-02", TglLahir: "1990-01-01", Los: 1},
		{NoSep: "A", TglMasuk: "2025-10-01", TglKeluar: "2025-10-05", TglLahir: "1990-01-01", Los: 4}, // duplikat: menang yang terakhir
		{NoSep: "B", TglMasuk: "2025-02-31", TglKeluar: "2025-03-01", TglLahir: "1990-01-01"},         // tanggal mustahil
		{NoSep: "C", TglMasuk: "2025-10-01", TglKeluar: "2025-10-02", TglLahir: "salah"},
	}
	siap, tak := barisSiap(baris)
	if len(siap) != 2 || len(tak) != 1 {
		t.Fatalf("siap=%d tak=%d", len(siap), len(tak))
	}
	if siap[0].NoSep != "A" || siap[0].Los != 4 {
		t.Errorf("duplikat harus memakai baris terakhir: %+v", siap[0])
	}
	if siap[1].TglLahir != "1900-01-01" {
		t.Errorf("tanggal lahir tak valid harus jadi default: %s", siap[1].TglLahir)
	}
}

func TestRowHashStabilDanSensitif(t *testing.T) {
	b := BarisValid{NoSep: "A", Diagnosa: []string{"B", "A"}, Komponen: map[string]int64{"LAB": 1}}
	urut := BarisValid{NoSep: "A", Diagnosa: []string{"A", "B"}, Komponen: map[string]int64{"LAB": 1}}
	if rowHash(b, "RS1") != rowHash(urut, "RS1") {
		t.Error("urutan diagnosis tidak boleh memengaruhi hash")
	}
	b2 := b
	b2.Los = 9
	if rowHash(b, "RS1") == rowHash(b2, "RS1") {
		t.Error("perubahan LOS harus mengubah hash")
	}
	if rowHash(b, "RS1") == rowHash(b, "RS2") {
		t.Error("salt tenant harus memengaruhi hash")
	}
}
