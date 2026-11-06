# ============================================================
# MedClaim SPKD — skrip runner (root)
#
# Pakai:
#   .\run.ps1                    -> dev lengkap instance 1: backend + frontend,
#                                   masing-masing di jendela PowerShell baru
#   .\run.ps1 -Instance 2        -> dev lengkap instance 2 (API 7082, FE 7042, DB 7434)
#   .\run.ps1 backend            -> backend saja (DB + migrasi + seed + server)
#   .\run.ps1 frontend           -> frontend saja (Vite, menunjuk API instance-nya)
#   .\run.ps1 demo               -> mode demo WASM (data dummy, tanpa DB/backend)
#   .\run.ps1 migrate|seed       -> utilitas database (hormati -Instance)
#   .\run.ps1 wasm               -> build ulang medclaim.wasm dari backend-mock
#
# Peta port per instance: API 7080+N | PostgreSQL 7432+N | frontend 7040+N
# (blok 7xxx: menghindari simrs/medpath/matpad; detail di backend\README.md)
# ============================================================
param(
    [string]$Mode = "dev",
    [int]$Instance = 1,
    [switch]$Fresh,
    [switch]$NoSeed
)

$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot
$Frontend = Join-Path $Root "frontend"
$Backend = Join-Path $Root "backend"
$BackendMock = Join-Path $Root "backend-mock"

function Set-InstanceEnv {
    # Dipakai mode migrate/seed agar menunjuk DB instance yang benar.
    $dbPort = 7432 + $Instance
    $env:DATABASE_URL = "postgres://medclaim:medclaim@localhost:$dbPort/medclaim?sslmode=disable"
}

switch ($Mode) {
    "dev" {
        # Backend + frontend instance N, masing-masing di jendela sendiri (log terpisah).
        $beArgs = @('-NoExit', '-ExecutionPolicy', 'Bypass', '-File', (Join-Path $Backend 'run.ps1'), '-Instance', "$Instance")
        if ($Fresh)  { $beArgs += '-Fresh' }
        if ($NoSeed) { $beArgs += '-NoSeed' }
        Start-Process powershell -ArgumentList $beArgs
        Start-Process powershell -ArgumentList @('-NoExit', '-ExecutionPolicy', 'Bypass', '-File', (Join-Path $Frontend 'run.ps1'), '-Instance', "$Instance")
        $api = 7080 + $Instance; $fe = 7040 + $Instance
        Write-Host "Instance $Instance dibuka di 2 jendela:" -ForegroundColor Green
        Write-Host "  Backend : http://localhost:$api" -ForegroundColor Cyan
        Write-Host "  Frontend: http://localhost:$fe/app/" -ForegroundColor Cyan
        Write-Host "  Login   : admin@demo.medclaim.id / MedClaim#demo2026" -ForegroundColor DarkGray
    }
    "backend" {
        $p = @('-Instance', "$Instance"); if ($Fresh) { $p += '-Fresh' }; if ($NoSeed) { $p += '-NoSeed' }
        & (Join-Path $Backend 'run.ps1') @p
    }
    "frontend" {
        & (Join-Path $Frontend 'run.ps1') -Instance $Instance
    }
    "demo" {
        # Mode demo: backend Go WASM + data dummy di browser, tanpa DB.
        & (Join-Path $Frontend 'run.ps1') -Instance $Instance -Demo
    }
    "migrate" {
        Set-InstanceEnv
        Set-Location $Backend
        if ($Fresh) { go run ./cmd/migrate -fresh } else { go run ./cmd/migrate }
    }
    "seed" {
        Set-InstanceEnv
        Set-Location $Backend
        go run ./cmd/seed
    }
    "wasm" {
        # Build ulang backend demo (backend-mock) menjadi WASM untuk frontend.
        Set-Location $BackendMock
        $env:GOOS = "js"; $env:GOARCH = "wasm"
        go build -ldflags="-s -w" -o (Join-Path $Frontend "public\medclaim.wasm") .
        $ok = $LASTEXITCODE
        Remove-Item Env:GOOS, Env:GOARCH -ErrorAction SilentlyContinue
        if ($ok -ne 0) { Write-Error "build WASM gagal"; exit 1 }
        Write-Host "medclaim.wasm diperbarui di frontend\public" -ForegroundColor Green
    }
    default {
        Write-Error "Mode tidak dikenal: $Mode. Pilihan: dev | backend | frontend | demo | migrate | seed | wasm"
        exit 1
    }
}
