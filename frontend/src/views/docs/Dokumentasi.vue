<script setup lang="ts">
import { modeDemo } from "../api";
import { ref } from "vue";
import PageHeader from "../components/PageHeader.vue";

const buka = ref<string>("pengantar");
function toggle(id: string) {
  buka.value = buka.value === id ? "" : id;
}

const faq = [
  {
    id: "pengantar",
    tanya: "Apa itu MedClaim?",
    jawab:
      "MedClaim adalah modul analitik pra-pengajuan klaim INA-CBG pada produk SPKD (Sistem Pelayanan Kesehatan dan Data). Sistem menerima ekspor V-Claim (TSV), menilai kualitas data lewat 20 aturan DQ 7 dimensi, menghitung kelas kerugian (S1 Coding Loss, S2 Mixed Loss, Structural), mendeteksi anomali, lalu mengelola upaya recovery melalui Recovery Ledger dengan prinsip empat mata.",
  },
  {
    id: "format",
    tanya: "Format file apa yang didukung upload?",
    jawab:
      "Ekspor TSV/CSV V-Claim berbasis NAMA KOLOM (bukan posisi) dengan alias umum: NOSEP, NOKARTU, NOMR, NAMAPASIEN, TGLLAHIR, JENISKELAMIN, TGLMASUK, TGLKELUAR, LOS, KELASRAWAT, TIPELAYANAN, DIAGNOSA (ICD-10, pemisah koma/titik-koma), PROSEDUR (ICD-9-CM), KODEICBG, SEVERITYLEVEL, TARIFINACBG, hingga 18 kolom komponen billing (VISITATION_FEE … DARAH). Format Basic (tanpa komponen) tetap diterima — modul finansial akan ditandai TERBATAS.",
  },
  {
    id: "dq",
    tanya: "Bagaimana skor DQ dan gerbang dihitung?",
    jawab:
      "Setiap aturan terpicu memberi penalti tetap: KRITIS 100, PERINGATAN 25, INFO 10. Skor = 100 − Σ penalti (batas 0–100). Grade: A ≥ 90, B ≥ 80, C ≥ 60, D < 60. Gerbang BLOKIR bila skor < 60 ATAU ada aturan KRITIS terpicu — upload tidak bisa di-ingest kecuali Casemix/Super Admin memberi override dengan alasan minimal 10 karakter (semua tercatat di audit).",
  },
  {
    id: "versi",
    tanya: "Apa yang terjadi jika SEP yang sama diunggah ulang?",
    jawab:
      "Sistem menghitung row-hash (FNV-1a ganda atas bentuk kanonik baris + kode RS). Jika hash sama → baris dihitung UNCHANGED (idempoten, tidak menimpa). Jika berbeda → episode lama menjadi superseded dan dibuat versi baru dengan nomor versi naik; perubahan field ditampilkan sebagai diff (diagnosa, severity, tarif, LOS, dll.).",
  },
  {
    id: "rugi",
    tanya: "Bagaimana klasifikasi rugi S1/S2/Structural?",
    jawab:
      "selisih = tarif INA-CBG − total billing (penjumlahan 18 komponen). pct_rugi = |selisih|/tarif bila negatif. Surplus bila selisih ≥ 0; S1 Coding Loss bila 0 < pct < 40%; S2 Mixed Loss bila 40% ≤ pct < 70%; Structural Loss bila pct ≥ 70%. Potensi recovery diestimasi dari delta tarif rata-rata antar severity level pada grup CBG basis yang sama.",
  },
  {
    id: "ledger",
    tanya: "Apa aturan state machine Recovery Ledger?",
    jawab:
      "KANDIDAT → DIPERIKSA → DIKOREKSI → DIAJUKAN → HASIL (atau TIDAK_BERUBAH). Lompatan status ditolak. DIPERIKSA oleh Coder/Casemix; DIKOREKSI wajib referensi rekam medis; DIAJUKAN & HASIL hanya Casemix/Super Admin (empat mata); HASIL wajib mengisi realisasi; reopen status terminal wajib alasan ≥ 10 karakter. Sistem tidak pernah men-set DIAJUKAN/HASIL secara otomatis.",
  },
  {
    id: "rbac",
    tanya: "Siapa boleh melihat apa?",
    jawab:
      "Super Admin: semua RS + administrasi. Casemix: upload, ledger (tulis), query. Coder: analitik + query. Direktur: analitik + ledger (baca-saja). DPJP: hanya episode pasiennya sendiri, TANPA kolom finansial, dengan inbox query. Penyembunyian finansial diterapkan di server (bukan hanya UI).",
  },
  {
    id: "arsitektur",
    tanya: "Apa arsitektur versi Go + Vue ini?",
    jawab:
      "Backend Go murni (net/http stdlib — tanpa dependensi eksternal) pada port 3030 di belakang gateway: API JSON + penyajian SPA. Data disimpan sebagai JSON file dengan mutex (data/medclaim.json). Autentikasi PBKDF2-HMAC-SHA256 (210k iterasi) + cookie HMAC. Frontend Vue 3 + vue-router (hash mode) dibangun Vite menjadi satu file HTML yang disajikan Go. Semua logika domain (parser, DQ, analytics, state machine, RBAC) diporting setia dari versi TypeScript.",
  },
];
</script>

<template>
  <PageHeader judul="Dokumentasi" :demo="modeDemo" sub="Rumus kunci dan tanya-jawab penggunaan MedClaim." />
  <div style="max-width: 880px">
    <div class="kartu" style="margin-bottom: 14px">
      <h3>Rumus kunci</h3>
      <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-top: 8px">
        <div style="background: var(--bg); border-radius: 8px; padding: 12px 14px">
          <b style="font-size: 12.5px">Selisih</b>
          <div class="tunggal" style="font-size: 12.5px; margin-top: 3px">selisih = tarif_inacbg − total_billing</div>
        </div>
        <div style="background: var(--bg); border-radius: 8px; padding: 12px 14px">
          <b style="font-size: 12.5px">% Rugi</b>
          <div class="tunggal" style="font-size: 12.5px; margin-top: 3px">pct_rugi = max(0, −selisih) / tarif</div>
        </div>
        <div style="background: var(--bg); border-radius: 8px; padding: 12px 14px">
          <b style="font-size: 12.5px">Kelas rugi</b>
          <div class="tunggal" style="font-size: 12.5px; margin-top: 3px">0 → Surplus · &lt;40% → S1 · &lt;70% → S2 · ≥70% → Structural</div>
        </div>
        <div style="background: var(--bg); border-radius: 8px; padding: 12px 14px">
          <b style="font-size: 12.5px">Readiness (bobot 30/25/20/15/10)</b>
          <div class="tunggal" style="font-size: 12.5px; margin-top: 3px">severity + drug + multi-dx + icd + los</div>
        </div>
      </div>
    </div>

    <div v-for="f in faq" :key="f.id" class="kartu" style="padding: 0; margin-bottom: 8px; overflow: hidden">
      <div style="display: flex; align-items: center; gap: 10px; padding: 14px 18px; cursor: pointer" @click="toggle(f.id)">
        <b style="flex: 1">{{ f.tanya }}</b>
        <span style="color: var(--text-3)">{{ buka === f.id ? "▲" : "▼" }}</span>
      </div>
      <div v-if="buka === f.id" style="padding: 0 18px 16px; color: var(--text-2); font-size: 13.5px; border-top: 1px solid var(--border); padding-top: 12px">
        {{ f.jawab }}
      </div>
    </div>
  </div>
</template>
