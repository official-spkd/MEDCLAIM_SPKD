package database

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"medclaim-backend/migrations"
)

// Migrate menerapkan file migrations/*.sql yang belum tercatat di
// schema_migrations, berurutan menurut nama file. Setiap file dijalankan dalam
// satu transaksi: gagal berarti tidak ada perubahan untuk file itu.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	// Cegah dua proses migrasi berjalan bersamaan.
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock(727001)"); err != nil {
		return err
	}
	defer conn.Exec(context.Background(), "SELECT pg_advisory_unlock(727001)")

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		versi      varchar PRIMARY KEY,
		diterapkan timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return err
	}

	sudah := map[string]bool{}
	rows, err := conn.Query(ctx, "SELECT versi FROM schema_migrations")
	if err != nil {
		return err
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		sudah[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	nama, err := fs.Glob(migrations.Files, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(nama)

	diterapkan := 0
	for _, f := range nama {
		versi := strings.TrimSuffix(f, ".sql")
		if sudah[versi] {
			continue
		}
		sql, err := migrations.Files.ReadFile(f)
		if err != nil {
			return err
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		if err := terapkan(ctx, tx, versi, string(sql)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("migrasi %s gagal (dibatalkan): %w", versi, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		fmt.Println("diterapkan:", versi)
		diterapkan++
	}
	if diterapkan == 0 {
		fmt.Println("skema sudah terbaru")
	}
	return nil
}

// Reset menghapus seluruh objek di schema public (tabel, enum, riwayat
// migrasi) lalu membuatnya kosong kembali. HANYA untuk development.
func Reset(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	return err
}

func terapkan(ctx context.Context, tx pgx.Tx, versi, sql string) error {
	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, "INSERT INTO schema_migrations (versi) VALUES ($1)", versi)
	return err
}
