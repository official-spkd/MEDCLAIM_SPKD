// Package seed mengisi data master dan akun demo untuk development.
// Idempoten: dapat dijalankan berulang tanpa membuat data ganda.
// Data episode/ledger/query TIDAK di-seed — itu datang dari alur upload.
package seed

import (
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"medclaim-backend/internal/claims"
	"medclaim-backend/internal/hospitals"
	"medclaim-backend/internal/queries"
	"medclaim-backend/internal/settings"
	"medclaim-backend/internal/users"
)

func s(v string) *string { return &v }

type userSeed struct {
	email, nama, role, rs string
	spesialisasi, dpjp    *string
}

// Run menjalankan seed dalam satu transaksi. password = sandi semua akun demo.
func Run(db *gorm.DB, password string) error {
	if len(password) < 8 {
		return errors.New("password seed minimal 8 karakter")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// --- Rumah sakit ---
		for _, h := range []hospitals.Hospital{
			{Kode: "RS001", Nama: "RSUD Kota Anggrek", Kota: "Bandung", Kelas: "B"},
			{Kode: "RS002", Nama: "RS Harapan Medika", Kota: "Bekasi", Kelas: "B"},
			{Kode: "RS003", Nama: "RS Pelayanan Sentosa", Kota: "Semarang", Kelas: "C"},
		} {
			if err := tx.Where("kode = ?", h.Kode).FirstOrCreate(&h).Error; err != nil {
				return fmt.Errorf("seed hospital %s: %w", h.Kode, err)
			}
		}
		rsID := map[string]string{}
		var hs []hospitals.Hospital
		if err := tx.Find(&hs).Error; err != nil {
			return err
		}
		for _, h := range hs {
			rsID[h.Kode] = h.ID
		}

		// --- Pengguna demo ---
		for _, u := range []userSeed{
			{"admin@demo.medclaim.id", "Admin SPKD", "SUPERADMIN", "", nil, nil},
			{"casemix@demo.medclaim.id", "Rina Casemix", "CASEMIX", "RS001", nil, nil},
			{"coder@demo.medclaim.id", "Citra Coder", "CODER", "RS001", nil, nil},
			{"direktur@demo.medclaim.id", "Drs. Hendra Direktur", "DIREKTUR", "RS001", nil, nil},
			{"dpjp@demo.medclaim.id", "dr. Budi Prakoso, Sp.PD", "DPJP", "RS001", s("Penyakit Dalam"), s("DPJP-0101")},
			{"dpjp02@demo.medclaim.id", "dr. Sinta Maharani, Sp.B", "DPJP", "RS001", s("Bedah"), s("DPJP-0102")},
			{"dpjp03@demo.medclaim.id", "dr. Rahmat Hidayat, Sp.An", "DPJP", "RS001", s("Anestesi"), s("DPJP-0103")},
			{"dpjp04@demo.medclaim.id", "dr. Yanti Lestari, Sp.OG", "DPJP", "RS001", s("Obstetri Ginekologi"), s("DPJP-0104")},
			{"dpjp05@demo.medclaim.id", "dr. Eko Prasetyo, Sp.P", "DPJP", "RS001", s("Paru"), s("DPJP-0105")},
			{"dpjp06@demo.medclaim.id", "dr. Fitri Utami, Sp.A", "DPJP", "RS001", s("Anak"), s("DPJP-0106")},
		} {
			var n int64
			if err := tx.Model(&users.User{}).Where("email = ?", u.email).Count(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				continue
			}
			row := users.User{
				Email: u.email, PasswordHash: string(hash), Nama: u.nama, Role: u.role,
				Spesialisasi: u.spesialisasi, DpjpKode: u.dpjp, Aktif: true,
			}
			if u.rs != "" {
				id, ok := rsID[u.rs]
				if !ok {
					return fmt.Errorf("seed user %s: RS %s tidak ada", u.email, u.rs)
				}
				row.HospitalID = &id
			}
			if err := tx.Create(&row).Error; err != nil {
				return fmt.Errorf("seed user %s: %w", u.email, err)
			}
		}

		// --- Template query ---
		for _, t := range []queries.Template{
			{Nama: "Klarifikasi Diagnosis Utama", Kategori: "KODING", Aktif: true,
				Isi: "Bagaimana kondisi klinis utama pasien pada episode ini dan temuan penunjang apa yang mendukung diagnosis utamanya?"},
			{Nama: "Kelengkapan Resume", Kategori: "DOKUMEN", Aktif: true,
				Isi: "Dokumen apa yang perlu dilengkapi pada resume medis episode ini agar kondisi komorbid pasien terdokumentasi utuh?"},
			{Nama: "Catatan Operasi", Kategori: "KODING", Aktif: true,
				Isi: "Apakah terdapat catatan tindakan/operasi yang belum tercatat pada episode ini dan kapan tindakan tersebut dilakukan?"},
		} {
			if err := tx.Where("nama = ?", t.Nama).FirstOrCreate(&t).Error; err != nil {
				return err
			}
		}

		// --- Pengaturan sistem ---
		for _, kv := range []settings.Setting{
			{Key: "readiness.weights", Value: `{"severity":30,"drug":25,"singleDiag":20,"icd":15,"los":10}`},
			{Key: "dq.penalti", Value: `{"kritis":100,"peringatan":25,"info":10}`},
			{Key: "dq.max_invalid_row_pct", Value: "5"},
			{Key: "sla.query_hari", Value: "3"},
		} {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&kv).Error; err != nil {
				return err
			}
		}

		// --- Referensi INA-CBG yang muncul pada file contoh ---
		// Referensi lain ditambahkan otomatis saat upload.
		for _, r := range []claims.RefInacbg{
			{Kode: "I-4-08", Deskripsi: "Gastritis & duodenitis", Tarif: 2450000},
			{Kode: "Q-5-14", Deskripsi: "Septicemia", Tarif: 6200000},
			{Kode: "K-4-06", Deskripsi: "Appendektomi", Tarif: 5100000},
			{Kode: "I-4-09", Deskripsi: "Gagal jantung", Tarif: 4780598},
			{Kode: "E-4-02", Deskripsi: "Diabetes mellitus", Tarif: 3200501},
			{Kode: "J-4-07", Deskripsi: "Pneumonia", Tarif: 5450321},
			{Kode: "RJ-2-09", Deskripsi: "Kunjungan lanjutan RJTP", Tarif: 470249},
		} {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&r).Error; err != nil {
				return err
			}
		}
		log.Println("seed: RS, akun demo (*@demo.medclaim.id), template, pengaturan, ref INA-CBG")
		return nil
	})
}
