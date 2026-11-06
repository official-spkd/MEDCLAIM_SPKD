# run.ps1 — jalankan backend MedClaim dari nol, mendukung banyak instance di satu mesin.
#
# Pakai:  .\run.ps1                 -> instance 1: API 7081, DB 7433 (frontend 7041)
#         .\run.ps1 -Instance 2     -> instance 2: API 7082, DB 7434 (frontend 7042)
#         .\run.ps1 -Instance 3     -> instance 3: API 7083, DB 7435 (frontend 7043)
#         .\run.ps1 -Fresh          -> reset database dulu (hapus semua data)
#         .\run.ps1 -NoSeed         -> lewati seed
#         .\run.ps1 -ApiPort 9090 -DbPort 5500   -> port manual (abaikan skema instance)
#
# Tiap instance punya container, volume, dan network Docker sendiri (project medclaim-iN),
# jadi datanya terpisah dan bisa jalan bersamaan.
param(
    [int]$Instance = 1,
    [int]$ApiPort = 0,
    [int]$DbPort = 0,
    [switch]$Fresh,
    [switch]$NoSeed
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

if ($Instance -lt 1) { throw "-Instance minimal 1." }
# Blok 7xxx: menghindari sistem lain di mesin ini (simrs 8080/5432, medpath 8082/5434, matpad 303x).
if ($ApiPort -eq 0) { $ApiPort = 7080 + $Instance }   # i1=7081, i2=7082, ...
if ($DbPort  -eq 0) { $DbPort  = 7432 + $Instance }   # i1=7433, i2=7434, ...
$FePort  = 7040 + $Instance                            # i1=7041, i2=7042, ...
$Project = "medclaim-i$Instance"

# Env OS menang atas .env (godotenv tidak menimpa env yang sudah ada),
# jadi semua nilai per-instance diatur di sini.
$env:MEDCLAIM_DB_PORT = "$DbPort"
$env:PORT             = "$ApiPort"
$env:DATABASE_URL     = "postgres://medclaim:medclaim@localhost:$DbPort/medclaim?sslmode=disable"
$env:CORS_ORIGINS     = "http://localhost:$FePort,http://localhost:5173"

Write-Host "Instance $Instance  ->  API :$ApiPort | PostgreSQL :$DbPort | frontend :$FePort | project $Project" -ForegroundColor Magenta

# 1. Database (Docker, terisolasi per instance lewat nama project)
Write-Host "[1/4] Menyalakan PostgreSQL (docker compose -p $Project)..." -ForegroundColor Cyan
docker compose -p $Project up -d db
if ($LASTEXITCODE -ne 0) { throw "docker compose gagal. Pastikan Docker Desktop berjalan dan port $DbPort kosong." }

# Tunggu DB siap
$ready = $false
for ($i = 0; $i -lt 30; $i++) {
    docker compose -p $Project exec -T db pg_isready -U medclaim -d medclaim *> $null
    if ($LASTEXITCODE -eq 0) { $ready = $true; break }
    Start-Sleep -Seconds 1
}
if (-not $ready) { throw "PostgreSQL tidak siap setelah 30 detik." }
Write-Host "      PostgreSQL siap di localhost:$DbPort" -ForegroundColor Green

# 2. Migrasi
Write-Host "[2/4] Menjalankan migrasi..." -ForegroundColor Cyan
if ($Fresh) { go run ./cmd/migrate -fresh } else { go run ./cmd/migrate }
if ($LASTEXITCODE -ne 0) { throw "Migrasi gagal." }

# 3. Seed
if (-not $NoSeed) {
    Write-Host "[3/4] Menjalankan seed (akun demo)..." -ForegroundColor Cyan
    go run ./cmd/seed
    if ($LASTEXITCODE -ne 0) { throw "Seed gagal." }
} else {
    Write-Host "[3/4] Seed dilewati." -ForegroundColor Yellow
}

# 4. Server
Write-Host "[4/4] Menjalankan server di http://localhost:$ApiPort (Ctrl+C untuk berhenti)" -ForegroundColor Cyan
Write-Host "      Login demo: admin@demo.medclaim.id / MedClaim#demo2026" -ForegroundColor Green
Write-Host "      Frontend instance ini: cd ../frontend; .\run.ps1 -Instance $Instance  (port $FePort)" -ForegroundColor DarkGray
go run ./cmd/server
