package scoring

import "testing"

func TestKlasifikasiRugi(t *testing.T) {
	cases := []struct {
		nama           string
		tarif, billing int64
		want           string
	}{
		{"surplus", 1000, 800, "SURPLUS"},
		{"impas", 1000, 1000, "SURPLUS"},
		{"rugi 10% = S1", 1000, 1100, "S1_CODING"},
		{"rugi 39% = S1", 1000, 1390, "S1_CODING"},
		{"rugi 40% = S2", 1000, 1400, "S2_MIXED"},
		{"rugi 69% = S2", 1000, 1690, "S2_MIXED"},
		{"rugi 70% = structural", 1000, 1700, "STRUCTURAL"},
		{"rugi >100% dibatasi", 1000, 9000, "STRUCTURAL"},
		{"tarif nol dianggap surplus", 0, 500, "SURPLUS"},
	}
	for _, c := range cases {
		if got := KlasifikasiRugi(c.tarif, c.billing); got != c.want {
			t.Errorf("%s: dapat %s, harusnya %s", c.nama, got, c.want)
		}
	}
	if p := HitungPctRugi(1000, 9000); p != 1 {
		t.Errorf("pct rugi harus dibatasi 1, dapat %v", p)
	}
}

func TestBasisGrupCbg(t *testing.T) {
	if got := BasisGrupCbg("Q-5-14-0"); got != "Q-5-14" {
		t.Errorf("dapat %s", got)
	}
	if got := BasisGrupCbg("I-4-08"); got != "I-4-08" {
		t.Errorf("dapat %s", got)
	}
}

func TestDrugAlignment(t *testing.T) {
	ok, flag := DrugAlignment([]string{"K29.7"}, map[string]int64{"OBAT_KEMO": 100})
	if ok || len(flag) != 1 {
		t.Errorf("kemo tanpa neoplasma harus ditandai: ok=%v flag=%v", ok, flag)
	}
	ok, _ = DrugAlignment([]string{"C34.90"}, map[string]int64{"OBAT_KEMO": 100})
	if !ok {
		t.Error("kemo dengan neoplasma harus lolos")
	}
	ok, _ = DrugAlignment([]string{"E11.9"}, map[string]int64{"OBAT_KRONIS": 5})
	if !ok {
		t.Error("obat kronis dengan DM harus lolos")
	}
}

func TestReadiness(t *testing.T) {
	ep := &Episode{
		TglMasuk: "2025-10-01", TglKeluar: "2025-10-04", Los: 3, SeverityLevel: 2,
		KodeIcbg: "Q-5-14-2", Diagnosa: []string{"A41.9", "J44.1"}, Komponen: map[string]int64{},
	}
	if r := Readiness(ep); r.Total != 100 || len(r.Flag) != 0 {
		t.Errorf("episode sehat harus 100, dapat %+v", r)
	}
	ep.Diagnosa = []string{"A41.9"}
	ep.Los = 9
	r := Readiness(ep)
	if r.Total != 100-20-10 || len(r.Flag) != 2 {
		t.Errorf("single diagnosis + LOS salah harus 70, dapat %+v", r)
	}
}

func TestAnomaliReadmisi(t *testing.T) {
	a := &Episode{ID: "a", NoSep: "1", NoMr: "RM1", TglMasuk: "2025-10-01", TglKeluar: "2025-10-05", Los: 4, KodeIcbg: "X-1-01", Diagnosa: []string{"J18.9"}, TotalBilling: 10, Komponen: map[string]int64{"LAB": 5, "RAD": 5}}
	b := &Episode{ID: "b", NoSep: "2", NoMr: "RM1", TglMasuk: "2025-10-12", TglKeluar: "2025-10-14", Los: 2, KodeIcbg: "X-1-01", Diagnosa: []string{"J18.0"}, TotalBilling: 10, Komponen: map[string]int64{"LAB": 5, "RAD": 5}}
	got := DeteksiAnomali([]*Episode{a, b})
	var readmisi int
	for _, x := range got {
		if x.Jenis == "READMISI" && x.EpisodeID == "b" {
			readmisi++
		}
	}
	if readmisi != 1 {
		t.Errorf("harus 1 anomali readmisi untuk episode b, dapat %+v", got)
	}
}

func TestFormatRibuan(t *testing.T) {
	for in, want := range map[int64]string{0: "0", 999: "999", 1000: "1.000", 1234567: "1.234.567", -1500: "-1.500"} {
		if got := FormatRibuan(in); got != want {
			t.Errorf("%d: dapat %s, harusnya %s", in, got, want)
		}
	}
}
