#!/usr/bin/env bash
# run.sh — jalankan frontend MedClaim (Vite dev) per instance, sinkron dengan backend/run.sh.
#
# Pakai:  ./run.sh                  -> instance 1: frontend :7041 -> API http://localhost:7081
#         ./run.sh --instance 2     -> instance 2: frontend :7042 -> API http://localhost:7082
# (Blok 7xxx dipilih karena 303x dipakai matpad, 404x/808x/543x dipakai proyek lain.)
#         ./run.sh --demo           -> mode demo WASM (data dummy, tanpa backend)
#         ./run.sh --port 4000 --api-port 9090    -> port manual
set -euo pipefail
cd "$(dirname "$0")"

INSTANCE=1; PORT=0; API_PORT=0; DEMO=0
while [ $# -gt 0 ]; do
  case "$1" in
    --instance) INSTANCE="$2"; shift 2 ;;
    --port)     PORT="$2"; shift 2 ;;
    --api-port) API_PORT="$2"; shift 2 ;;
    --demo)     DEMO=1; shift ;;
    *) echo "argumen tidak dikenal: $1"; exit 1 ;;
  esac
done

[ "$INSTANCE" -ge 1 ] || { echo "--instance minimal 1"; exit 1; }
[ "$PORT"     -ne 0 ] || PORT=$((7040 + INSTANCE))     # i1=7041, i2=7042, ...
[ "$API_PORT" -ne 0 ] || API_PORT=$((7080 + INSTANCE)) # i1=7081, i2=7082, ...

[ -d node_modules ] || { echo "node_modules belum ada, menjalankan npm install..."; npm install; }

if [ "$DEMO" = 1 ]; then
  export VITE_MOCK=1
  export VITE_API_BASE=""
  echo "Instance $INSTANCE  ->  frontend :$PORT | MODE DEMO (WASM, data dummy)"
else
  export VITE_MOCK=""
  export VITE_API_BASE="http://localhost:$API_PORT"
  echo "Instance $INSTANCE  ->  frontend :$PORT | API http://localhost:$API_PORT"
  echo "Pastikan backend instance $INSTANCE sudah jalan: cd ../backend && ./run.sh --instance $INSTANCE"
fi

echo "Buka: http://localhost:$PORT/app/"
npx vite --port "$PORT" --host --strictPort
