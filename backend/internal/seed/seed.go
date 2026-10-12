package seed

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Sandi akun demo. Bisa diganti lewat env SEED_PASSWORD.
const sandiSeedDefault = "MedClaim#demo2026"

type seedUser struct {
	id, email, nama, role string
	hospitalID            *string
	spesialisasi, dpjp    *string
}

func s(v string) *string { return &v }

// Run mengisi data master & akun demo untuk development.
// Idempoten: baris yang sudah ada dilewati (ON CONFLICT DO NOTHING).
// Data episode/ledger/query TIDAK di-seed — itu datang dari alur upload.
func Run(ctx context.Context, pool *pgxpool.Pool) error {
	if os.Getenv("APP_ENV") == "production" {
		return errors.New("seed ditolak: APP_ENV=production")
	}

	sandi := os.Getenv("SEED_PASSWORD")
	if sandi == "" {
		sandi = sandiSeedDefault
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(sandi), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	b := &pgx.Batch{}

	for _, h := range [][4]string{
		{"rs1", "RS001", "RSUD Kota Anggrek", "Bandung"},
		{"rs2", "RS002", "RS Harapan Medika", "Bekasi"},
		{"rs3", "RS003", "RS Pelayanan Sentosa", "Semarang"},
	} {
		kelas := map[string]string{"rs1": "B", "rs2": "B", "rs3": "C"}[h[0]]
		b.Queue(`INSERT INTO hospitals (id, kode, nama, kota, kelas) VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT DO NOTHING`, h[0], h[1], h[2], h[3], kelas)
	}

	rs1 := s("rs1")
	users := []seedUser{
		{"u-admin", "admin@demo.medclaim.id", "Admin SPKD", "SUPERADMIN", nil, nil, nil},
		{"u-casemix", "casemix@demo.medclaim.id", "Rina Casemix", "CASEMIX", rs1, nil, nil},
		{"u-coder", "coder@demo.medclaim.id", "Citra Coder", "CODER", rs1, nil, nil},
		{"u-direktur", "direktur@demo.medclaim.id", "Drs. Hendra Direktur", "DIREKTUR", rs1, nil, nil},
		{"u-dpjp", "dpjp@demo.medclaim.id", "dr. Budi Prakoso, Sp.PD", "DPJP", rs1, s("Penyakit Dalam"), s("DPJP-0101")},
		{"u-dpjp2", "dpjp02@demo.medclaim.id", "dr. Sinta Maharani, Sp.B", "DPJP", rs1, s("Bedah"), s("DPJP-0102")},
		{"u-dpjp3", "dpjp03@demo.medclaim.id", "dr. Rahmat Hidayat, Sp.An", "DPJP", rs1, s("Anestesi"), s("DPJP-0103")},
		{"u-dpjp4", "dpjp04@demo.medclaim.id", "dr. Yanti Lestari, Sp.OG", "DPJP", rs1, s("Obstetri Ginekologi"), s("DPJP-0104")},
		{"u-dpjp5", "dpjp05@demo.medclaim.id", "dr. Eko Prasetyo, Sp.P", "DPJP", rs1, s("Paru"), s("DPJP-0105")},
		{"u-dpjp6", "dpjp06@demo.medclaim.id", "dr. Fitri Utami, Sp.A", "DPJP", rs1, s("Anak"), s("DPJP-0106")},
	}
	for _, u := range users {
		b.Queue(`INSERT INTO users (id, email, password_hash, nama, role, hospital_id, spesialisasi, dpjp_kode)
			VALUES ($1,$2,$3,$4,$5::peran,$6,$7,$8) ON CONFLICT DO NOTHING`,
			u.id, u.email, string(hash), u.nama, u.role, u.hospitalID, u.spesialisasi, u.dpjp)
	}

	for _, t := range [][4]string{
		{"tpl-1", "Klarifikasi Diagnosis Utama", "Bagaimana kondisi klinis utama pasien pada episode ini dan temuan penunjang apa yang mendukung diagnosis utamanya?", "KODING"},
		{"tpl-2", "Kelengkapan Resume", "Dokumen apa yang perlu dilengkapi pada resume medis episode ini agar kondisi komorbid pasien terdokumentasi utuh?", "DOKUMEN"},
		{"tpl-3", "Catatan Operasi", "Apakah terdapat catatan tindakan/operasi yang belum tercatat pada episode ini dan kapan tindakan tersebut dilakukan?", "KODING"},
	} {
		b.Queue(`INSERT INTO query_templates (id, nama, isi, kategori) VALUES ($1,$2,$3,$4)
			ON CONFLICT DO NOTHING`, t[0], t[1], t[2], t[3])
	}

	for _, kv := range [][2]string{
		{"readiness.weights", `{"severity":30,"drug":25,"singleDiag":20,"icd":15,"los":10}`},
		{"dq.penalti", `{"kritis":100,"peringatan":25,"info":10}`},
		{"dq.max_invalid_row_pct", "5"},
		{"sla.query_hari", "3"},
	} {
		b.Queue(`INSERT INTO settings (key, value) VALUES ($1,$2) ON CONFLICT DO NOTHING`, kv[0], kv[1])
	}

	// Referensi INA-CBG yang muncul di file sample. Referensi lain ditambah otomatis saat upload.
	for _, r := range []struct {
		kode, deskripsi string
		tarif           int64
	}{
		{"I-4-08", "Gastritis & duodenitis", 2450000},
		{"Q-5-14", "Septicemia", 6200000},
		{"K-4-06", "Appendektomi", 5100000},
		{"I-4-09", "Gagal jantung", 4780598},
		{"E-4-02", "Diabetes mellitus", 3200501},
		{"J-4-07", "Pneumonia", 5450321},
		{"RJ-2-09", "Kunjungan lanjutan RJTP", 470249},
	} {
		b.Queue(`INSERT INTO ref_inacbg (kode, deskripsi, tarif) VALUES ($1,$2,$3)
			ON CONFLICT DO NOTHING`, r.kode, r.deskripsi, r.tarif)
	}

	res := tx.SendBatch(ctx, b)
	for i := 0; i < b.Len(); i++ {
		if _, err := res.Exec(); err != nil {
			res.Close()
			return fmt.Errorf("seed gagal pada statement #%d: %w", i+1, err)
		}
	}
	if err := res.Close(); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	fmt.Printf("seed selesai (%d statement). Akun demo: *@demo.medclaim.id\n", b.Len())
	if os.Getenv("SEED_PASSWORD") == "" {
		fmt.Println("sandi default dipakai — set SEED_PASSWORD untuk menggantinya")
	}
	return nil
}
