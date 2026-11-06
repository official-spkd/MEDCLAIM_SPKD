# run.ps1 — jalankan frontend MedClaim (Vite dev) per instance, sinkron dengan backend/run.ps1.
#
# Pakai:  .\run.ps1                 -> instance 1: frontend :7041 -> API http://localhost:7081
#         .\run.ps1 -Instance 2     -> instance 2: frontend :7042 -> API http://localhost:7082
# (Blok 7xxx dipilih karena 303x dipakai matpad, 404x/808x/543x dipakai proyek lain.)
#         .\run.ps1 -Demo           -> mode demo WASM (data dummy, tanpa backend)
#         .\run.ps1 -Port 4000 -ApiPort 9090      -> port manual
#
# Env OS menang atas file .env* Vite, jadi VITE_API_BASE diatur di sini per instance.
param(
    [int]$Instance = 1,
    [int]$Port = 0,
    [int]$ApiPort = 0,
    [switch]$Demo
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

if ($Instance -lt 1) { throw "-Instance minimal 1." }
if ($Port    -eq 0) { $Port    = 7040 + $Instance }   # i1=7041, i2=7042, ...
if ($ApiPort -eq 0) { $ApiPort = 7080 + $Instance }   # i1=7081, i2=7082, ...

if (-not (Test-Path "node_modules")) {
    Write-Host "node_modules belum ada, menjalankan npm install..." -ForegroundColor Yellow
    npm install
    if ($LASTEXITCODE -ne 0) { throw "npm install gagal." }
}

if ($Demo) {
    $env:VITE_MOCK = "1"
    $env:VITE_API_BASE = ""
    Write-Host "Instance $Instance  ->  frontend :$Port | MODE DEMO (WASM, data dummy)" -ForegroundColor Magenta
} else {
    $env:VITE_MOCK = ""
    $env:VITE_API_BASE = "http://localhost:$ApiPort"
    Write-Host "Instance $Instance  ->  frontend :$Port | API http://localhost:$ApiPort" -ForegroundColor Magenta
    Write-Host "Pastikan backend instance $Instance sudah jalan: cd ../backend; .\run.ps1 -Instance $Instance" -ForegroundColor DarkGray
}

Write-Host "Buka: http://localhost:$Port/app/" -ForegroundColor Green
npx vite --port $Port --host --strictPort
