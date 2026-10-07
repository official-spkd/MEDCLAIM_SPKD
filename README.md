# MedClaim Frontend (mode demo, tanpa backend)

Vue 3 + Vite — redesign UI v3.1 "Amber Sunrise" (palette `#f0a640`, bahasa visual MedCredix / SPKD).

Backend Go dijalankan di browser lewat WebAssembly dengan data dummy
(`backend-wasm/embed/medclaim.json`). Perubahan data hanya hidup di memori tab, reset saat refresh.

Akun demo (sandi `MedClaim#demo2026`): admin / casemix / coder / direktur / dpjp @demo.medclaim.id

## Baru di v3.1 (redesign)

- **Sidebar kuning-oranye** — gradien amber (#fbc23e → #e78b12), teks espresso gelap,
  item aktif berupa pil putih dengan bayangan lembut; konsisten di mode terang & gelap.
- **Dashboard ala MedCredix** — hero Readiness Index dengan gauge ring 270°, grafik tren
  area multi-seri (Episode vs Estimasi Rugi, sumbu ganda), daftar **Tindakan Prioritas**
  (pola Risk Radar), distribusi kelas kerugian (donut dengan celah + hover),
  kelas rawat bar horizontal, dan tabel INA-CBG terbesar.
- **KPI card berikon** dengan lingkaran warna kanan-atas + efek angkat saat hover.
- **Pencarian global `Ctrl/Cmd+K`** — navigasi + episode/pasien (debounced, navigasi keyboard).
- **Panel notifikasi** — anomali, query menunggu jawaban, ledger aktif, upload terbaru.
- **Badge hitungan di sidebar** (Anomali / Query / Ledger) ala "Risk Radar" MedCredix.
- **Dark mode** netral hangat + toggle di header (tersimpan di localStorage).
- Favicon, judul dokumen per halaman, dan layar memuat ikut di-branding ulang.

## Menjalankan

    npm install
    npm run dev          # mode demo WASM aktif via .env.development (VITE_MOCK=1)

## Deploy Vercel

Import repo, Vercel otomatis memakai `vercel.json` (`npm run build:vercel` → mode `vercel` memuat `.env.vercel` dengan `VITE_MOCK=1`). Tidak perlu env apa pun.

Untuk backend Go sungguhan: kosongkan `VITE_MOCK`, isi `VITE_API_BASE` (lihat `.env.example`).

## Build ulang WASM (opsional, hanya bila Go diubah)

    cd backend-wasm
    GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o ../public/medclaim.wasm .
