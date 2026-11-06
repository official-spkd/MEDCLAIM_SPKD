# MedClaim SPKD — Backend

API Go (Gin + GORM + PostgreSQL) untuk MedClaim SPKD. Tiga binary di `cmd/`:

| Binary | Fungsi |
| --- | --- |
| `cmd/server` | HTTP API di port `7081` (endpoint `/health`, `/ready`, `/api/...`) |
| `cmd/migrate` | Menjalankan migrasi SQL di `migrations/` (flag `-fresh` = reset schema, hanya development) |
| `cmd/seed` | Mengisi RS demo, akun demo, template query, pengaturan, referensi INA-CBG |

## Prasyarat

- Go 1.26+
- Docker (untuk PostgreSQL) — atau PostgreSQL lokal sendiri, tinggal sesuaikan `DATABASE_URL`

## Cara menjalankan — full stack Docker (paling simpel)

```bash
cd backend
docker compose up -d --build
# atau, kalau punya make (bisa juga dari root repo):
make up
```

Satu perintah itu menyalakan **db + backend + frontend** sekaligus:

- Frontend: <http://localhost:7041/app/>
- API: <http://localhost:7081> (migrasi + seed jalan otomatis saat container start)

Hentikan dengan `docker compose down` (tambah `-v` untuk hapus data).
Port bisa digeser lewat env `MEDCLAIM_API_PORT`, `MEDCLAIM_FE_PORT`, `MEDCLAIM_DB_PORT`
(plus `-p medclaim-iN` untuk instance lain).

Dengan `make` semuanya jadi satu kata (`make help` untuk daftar lengkap):

```bash
make up          # full stack instance 1
make up I=2      # instance 2 — project medclaim-i2, port 7082/7434/7042 otomatis
make logs        # ikuti log       (make logs I=2 untuk instance 2, dst)
make down        # stop            (make clean = stop + hapus data)
make db          # hanya database — pasangan mode dev go run
make migrate && make seed && make server   # mode dev manual, env per instance otomatis
``` Catatan: frontend di mode ini hasil
build statis — untuk coding harian dengan hot-reload, pakai skrip di bawah.

## Cara menjalankan — mode dev hot-reload (sekali perintah)

```powershell
# Windows (PowerShell)
.\run.ps1                # instance 1 (default)
.\run.ps1 -Instance 2    # instance 2 — bisa jalan bersamaan dengan instance 1
.\run.ps1 -Fresh         # reset database dulu (hapus semua data)
.\run.ps1 -NoSeed        # tanpa seed
.\run.ps1 -ApiPort 9090 -DbPort 5500   # port manual
```

```bash
# Git Bash / Linux / macOS
./run.sh                 # instance 1 (default)
./run.sh --instance 2
./run.sh --fresh
./run.sh --no-seed
```

Skrip ini otomatis: menyalakan PostgreSQL via `docker compose` (hanya service `db`),
menunggu DB siap, migrasi, seed, lalu menjalankan server dengan `go run`.
Jangan gabungkan dengan full stack Docker pada instance yang sama (port API/frontend
akan bentrok) — pilih salah satu.

Dari root repo juga bisa: `.\run.ps1` (buka backend + frontend dev dalam 2 jendela),
`.\run.ps1 -Instance 2`, atau mode lain (`backend|frontend|demo|migrate|seed|wasm`).

## Banyak instance di satu mesin

Karena satu mesin bisa dipakai beberapa orang/proyek sekaligus, setiap instance
mendapat blok port sendiri dan Docker project sendiri (`medclaim-iN`) sehingga
container, volume (data), dan network-nya terpisah total:

| Instance | API | PostgreSQL | Frontend | Docker project |
| --- | --- | --- | --- | --- |
| 1 | 7081 | 7433 | 7041 | `medclaim-i1` |
| 2 | 7082 | 7434 | 7042 | `medclaim-i2` |
| 3 | 7083 | 7435 | 7043 | `medclaim-i3` |
| N | 7080+N | 7432+N | 7040+N | `medclaim-iN` |

MedClaim sengaja memakai blok **7xxx** karena port yang **sudah dipakai proyek
lain** di mesin ini cukup banyak (jangan dipakai): `8080`, `3000`, `8000` (simrs),
`5432` (postgres simrs), `8082`, `5434`, `5173` (medpath), `6379` (redis),
`9000-9001` (minio), `3031` dan blok 303x (matpad), `3306` (mysql papi).

Frontend untuk instance N (skrip `frontend/run.ps1` / `run.sh` sudah mengatur
port dan `VITE_API_BASE` secara otomatis):

```powershell
cd ../frontend
.\run.ps1 -Instance 2    # frontend :7042 -> API http://localhost:7082
```

CORS backend otomatis mengizinkan `http://localhost:704N` sesuai instance-nya.

Mematikan sebuah instance: `Ctrl+C` di terminal server, lalu bila ingin
mematikan database-nya juga: `docker compose -p medclaim-iN down`
(tambah `-v` untuk sekalian menghapus datanya).

## Cara manual (per langkah, instance 1)

```bash
docker compose -p medclaim-i1 up -d db   # PostgreSQL di localhost:7433
go run ./cmd/migrate                     # migrasi (tambah -fresh untuk reset)
go run ./cmd/seed                        # akun demo
go run ./cmd/server                      # API di http://localhost:7081
```

(Nilai default `.env` sudah menunjuk instance 1. Untuk instance lain, pakai
`run.ps1`/`run.sh` yang mengatur semua env per-instance secara otomatis.)

## Konfigurasi (`.env`)

Lihat `.env.example`. Variabel penting:

- `DATABASE_URL` — wajib. Default skrip: `postgres://medclaim:medclaim@localhost:7433/medclaim?sslmode=disable`
- `PORT` — default `7081`
- `CORS_ORIGINS` — asal frontend (default sudah memuat `http://localhost:7041` untuk Vite)
- `SEED_PASSWORD` — sandi semua akun demo (default `MedClaim#demo2026`)

## Akun demo (hasil seed)

Sandi semua akun: `MedClaim#demo2026` (atau nilai `SEED_PASSWORD`).

- `admin@demo.medclaim.id` — SUPERADMIN
- `casemix@demo.medclaim.id` — CASEMIX (RS001)
- `coder@demo.medclaim.id` — CODER (RS001)
- `direktur@demo.medclaim.id` — DIREKTUR (RS001)
- `dpjp@demo.medclaim.id` … `dpjp06@demo.medclaim.id` — DPJP (RS001)

## Tes cepat

```bash
curl http://localhost:7081/health
curl -X POST http://localhost:7081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@demo.medclaim.id","password":"MedClaim#demo2026"}'
```

Koleksi Postman lengkap ada di `docs/` (`MedClaim SPKD API.postman_collection.json`).

## Menyambungkan frontend

```bash
cd ../frontend
# kosongkan VITE_MOCK, isi VITE_API_BASE=http://localhost:7081 (lihat .env.example frontend)
npm install && npm run dev    # jalan di http://localhost:7041
```

## Unit test

```bash
go test ./...
```
