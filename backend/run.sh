#!/usr/bin/env bash
# run.sh — jalankan backend MedClaim dari nol, mendukung banyak instance di satu mesin.
#
# Pakai:  ./run.sh                  -> instance 1: API 7081, DB 7433 (frontend 7041)
#         ./run.sh --instance 2     -> instance 2: API 7082, DB 7434 (frontend 7042)
#         ./run.sh --fresh          -> reset database dulu (hapus semua data)
#         ./run.sh --no-seed        -> lewati seed
#         ./run.sh --api-port 9090 --db-port 5500   -> port manual
#
# Tiap instance punya container, volume, dan network Docker sendiri (project medclaim-iN).
set -euo pipefail
cd "$(dirname "$0")"

INSTANCE=1; API_PORT=0; DB_PORT=0; FRESH=0; SEED=1
while [ $# -gt 0 ]; do
  case "$1" in
    --instance) INSTANCE="$2"; shift 2 ;;
    --api-port) API_PORT="$2"; shift 2 ;;
    --db-port)  DB_PORT="$2";  shift 2 ;;
    --fresh)    FRESH=1; shift ;;
    --no-seed)  SEED=0; shift ;;
    *) echo "argumen tidak dikenal: $1"; exit 1 ;;
  esac
done

[ "$INSTANCE" -ge 1 ] || { echo "--instance minimal 1"; exit 1; }
# Blok 7xxx: menghindari sistem lain di mesin ini (simrs 8080/5432, medpath 8082/5434, matpad 303x).
[ "$API_PORT" -ne 0 ] || API_PORT=$((7080 + INSTANCE))   # i1=7081, i2=7082, ...
[ "$DB_PORT"  -ne 0 ] || DB_PORT=$((7432 + INSTANCE))    # i1=7433, i2=7434, ...
FE_PORT=$((7040 + INSTANCE))                             # i1=7041, i2=7042, ...
PROJECT="medclaim-i$INSTANCE"

# Env OS menang atas .env (godotenv tidak menimpa env yang sudah ada).
export MEDCLAIM_DB_PORT="$DB_PORT"
export PORT="$API_PORT"
export DATABASE_URL="postgres://medclaim:medclaim@localhost:$DB_PORT/medclaim?sslmode=disable"
export CORS_ORIGINS="http://localhost:$FE_PORT,http://localhost:5173"

echo "Instance $INSTANCE  ->  API :$API_PORT | PostgreSQL :$DB_PORT | frontend :$FE_PORT | project $PROJECT"

echo "[1/4] Menyalakan PostgreSQL (docker compose -p $PROJECT)..."
docker compose -p "$PROJECT" up -d db

for i in $(seq 1 30); do
  if docker compose -p "$PROJECT" exec -T db pg_isready -U medclaim -d medclaim >/dev/null 2>&1; then
    break
  fi
  [ "$i" = 30 ] && { echo "PostgreSQL tidak siap setelah 30 detik."; exit 1; }
  sleep 1
done
echo "      PostgreSQL siap di localhost:$DB_PORT"

echo "[2/4] Menjalankan migrasi..."
if [ "$FRESH" = 1 ]; then go run ./cmd/migrate -fresh; else go run ./cmd/migrate; fi

if [ "$SEED" = 1 ]; then
  echo "[3/4] Menjalankan seed (akun demo)..."
  go run ./cmd/seed
else
  echo "[3/4] Seed dilewati."
fi

echo "[4/4] Menjalankan server di http://localhost:$API_PORT (Ctrl+C untuk berhenti)"
echo "      Login demo: admin@demo.medclaim.id / MedClaim#demo2026"
echo "      Frontend instance ini: cd ../frontend && ./run.sh --instance $INSTANCE  (port $FE_PORT)"
go run ./cmd/server
