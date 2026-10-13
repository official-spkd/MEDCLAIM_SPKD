<script setup lang="ts">
// ============================================================
// MedClaim — Dashboard v3.1 (hero gauge + KPI + tren dual-axis
// + tindakan prioritas ala Risk Radar MedCredix)
// ============================================================
import { ref, watch, onMounted, computed } from "vue";
import { useRouter } from "vue-router";
import { api, modeDemo } from "../api";
import { store, queryFilter, bisaAkses } from "../store";
import { rupiah, rupiahSingkat, angka, persen } from "../format";
import type { KPI } from "../types";
import PageHeader from "../components/PageHeader.vue";
import KpiCard from "../components/KpiCard.vue";
import Ikon from "../components/Ikon.vue";
import GrafikDonut from "../components/GrafikDonut.vue";
import GrafikTren from "../components/GrafikTren.vue";

const router = useRouter();

interface Dash {
  kpi: KPI;
  distribusiRugi: Record<string, number>;
  tren: { periode: string; episode: number; rugi: number }[];
  topCbg: { kode: string; desk: string; jumlah: number; rugi: number }[];
  kelasRawat: { kelas: string; jumlah: number }[];
  readiness: { total: number; komponen: Record<string, number> };
}

const data = ref<Dash | null>(null);
const galat = ref("");
const memuat = ref(true);

async function muat() {
  if (!bisaAkses.value.analitik && store.sesi?.role !== "DPJP") return;
  memuat.value = true;
  galat.value = "";
  try {
    data.value = await api.get<Dash>("/api/dashboard" + queryFilter());
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal memuat dashboard.";
  } finally {
    memuat.value = false;
  }
}
onMounted(muat);
watch(() => [store.filter.rs, store.filter.tipe, store.filter.dari, store.filter.sampai], muat);

const warnaRugi: Record<string, string> = {
  SURPLUS: "#059669",
  S1_CODING: "#d97706",
  S2_MIXED: "#ea580c",
  STRUCTURAL: "#dc2626",
};
const labelRugi: Record<string, string> = {
  SURPLUS: "Surplus",
  S1_CODING: "S1 — Coding Loss",
  S2_MIXED: "S2 — Mixed Loss",
  STRUCTURAL: "Structural Loss",
};
const labelReady: Record<string, string> = {
  severity: "Severity", drug: "Drug Align", singleDiag: "Multi-Dx", icd: "ICD-10", los: "LOS",
};

const statusReadiness = computed(() => {
  const n = data.value?.readiness.total ?? 0;
  if (n >= 85) return "Siap Klaim — komponen dokumentasi lengkap";
  if (n >= 70) return "Siap dengan Area Perbaikan";
  if (n >= 50) return "Perlu Perhatian — beberapa komponen lemah";
  return "Kritis — perbaikan dokumentasi menyeluruh";
});

// ---------- Ring gauge hero ----------
const C_RING = 2 * Math.PI * 63; // r=63
const ARK = C_RING * 0.75; // busur 270°
const dashRing = computed(() => `${(Math.min(100, data.value?.readiness.total ?? 0) / 100) * ARK} ${C_RING}`);

// ---------- Tren dual-axis ----------
const seriTren = computed(() => {
  const t = data.value?.tren ?? [];
  if (!t.length) return [];
  const seri: { nama: string; warna: string; nilai: number[]; sumbu: "kiri" | "kanan"; jenis: "area" | "garis" }[] = [
    { nama: "Episode", warna: "#f0a640", nilai: t.map((d) => d.episode), sumbu: "kiri", jenis: "area" },
  ];
  if (bisaAkses.value.finansial && t.some((d) => d.rugi > 0)) {
    seri.push({ nama: "Estimasi Rugi", warna: "#dc2626", nilai: t.map((d) => d.rugi), sumbu: "kanan", jenis: "garis" });
  }
  return seri;
});

// ---------- Tindakan prioritas (pola Risk Radar MedCredix) ----------
interface Prioritas {
  ikon: string;
  tingkat: "tinggi" | "sedang" | "rendah";
  judul: string;
  desk: string;
  label: string;
  badge: string;
  ke: string;
}
const prioritas = computed<Prioritas[]>(() => {
  const k = data.value?.kpi;
  if (!k) return [];
  const a = bisaAkses.value;
  const out: Prioritas[] = [];
  if (a.finansial && k.anomaliJumlah > 0) {
    out.push({
      ikon: "peringatan", tingkat: "tinggi",
      judul: `Tindak lanjuti ${angka(k.anomaliJumlah)} anomali`,
      desk: "LOS outlier · duplikat SEP · readmisi · selisih ekstrem",
      label: "Tinggi", badge: "merah", ke: "/anomali",
    });
  }
  if ((a.query || store.sesi?.role === "DPJP") && k.queryTerlambat > 0) {
    out.push({
      ikon: "jam", tingkat: "tinggi",
      judul: `${angka(k.queryTerlambat)} query melewati SLA`,
      desk: "Klarifikasi coding terlambat dijawab DPJP",
      label: "Tinggi", badge: "merah", ke: "/query",
    });
  }
  if ((a.query || store.sesi?.role === "DPJP") && k.queryTerbuka > 0) {
    out.push({
      ikon: "tanya", tingkat: "sedang",
      judul: `${angka(k.queryTerbuka)} query menunggu jawaban`,
      desk: "Klarifikasi coding yang belum dijawab DPJP",
      label: "Sedang", badge: "kuning", ke: "/query",
    });
  }
  if (a.ledger && k.ledgerAktif > 0) {
    out.push({
      ikon: "buku", tingkat: "sedang",
      judul: `${angka(k.ledgerAktif)} entri ledger aktif`,
      desk: "Kandidat recovery yang masih diproses empat mata",
      label: "Sedang", badge: "ungu", ke: "/ledger",
    });
  }
  if (a.upload && k.dqRata > 0 && k.dqRata < 85) {
    out.push({
      ikon: "mutu", tingkat: k.dqRata < 70 ? "tinggi" : "rendah",
      judul: `Skor DQ rata-rata ${k.dqRata}/100`,
      desk: "Perbaiki dimensi kualitas data sebelum verifikasi klaim",
      label: k.dqRata < 70 ? "Tinggi" : "Rendah", badge: k.dqRata < 70 ? "merah" : "info", ke: "/kualitas",
    });
  }
  return out;
});

// ---------- Kelas rawat ----------
const totalKelas = computed(() =>
  (data.value?.kelasRawat ?? []).reduce((a, k) => a + k.jumlah, 0),
);
</script>

<template>
  <PageHeader
    judul="Dashboard Eksekutif"
    :demo="modeDemo"
    sub="Kondisi klaim dalam 30 detik — setiap angka dapat ditelusuri ke episode, komponen, dan tindakannya."
  >
    <template #aksi>
      <button class="btn garis" @click="router.push('/klaim')">
        <Ikon nama="klaim" :ukuran="15" />
        Buka Claim Explorer
      </button>
    </template>
  </PageHeader>

  <div v-if="store.sesi?.role === 'DPJP'" style="margin-bottom: 16px" class="info-kecil">
    Anda masuk sebagai DPJP — dashboard menampilkan hanya episode pasien Anda; kolom finansial disembunyikan.
  </div>

  <div v-if="memuat" class="pemuat"><div class="spinner" /><span>Memuat dashboard…</span></div>
  <div v-else-if="galat" class="bahaya-kecil">{{ galat }}</div>

  <template v-else-if="data">
    <!-- ====== Kartu hero: indeks readiness (gauge ring) ====== -->
    <div class="kartu-unggulan" style="margin-bottom: 16px">
      <div class="hero-grid">
        <div class="hero-ring">
          <svg width="164" height="164" viewBox="0 0 164 164">
            <defs>
              <linearGradient id="ringGradHero" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0%" stop-color="#ffd98a" />
                <stop offset="100%" stop-color="#f0a640" />
              </linearGradient>
            </defs>
            <g transform="rotate(135 82 82)">
              <circle
                cx="82" cy="82" r="63" fill="none" stroke="rgba(255, 255, 255, 0.13)"
                stroke-width="13" :stroke-dasharray="`${ARK} ${C_RING}`" stroke-linecap="round"
              />
              <circle
                cx="82" cy="82" r="63" fill="none" stroke="url(#ringGradHero)"
                stroke-width="13" :stroke-dasharray="dashRing" stroke-linecap="round"
                style="transition: stroke-dasharray 0.5s ease; filter: drop-shadow(0 0 10px rgba(240, 166, 64, 0.45))"
              />
            </g>
            <text
              x="82" y="80" text-anchor="middle"
              style="font-family: var(--font-judul); font-weight: 800; font-size: 37px"
              fill="#fff"
            >{{ data.readiness.total }}</text>
            <text x="82" y="100" text-anchor="middle" style="font-size: 11.5px" fill="#cf9d4e">/ 100</text>
          </svg>
        </div>

        <div class="hero-info">
          <div class="label-hero">MedClaim Readiness Index</div>
          <div class="status-hero">{{ statusReadiness }}</div>
          <div style="display: flex; flex-direction: column; gap: 8px; margin-top: 14px; max-width: 300px">
            <span class="chip-hero">Data Quality <b>{{ data.kpi.dqRata }}</b>/100</span>
            <span class="chip-hero">Episode Aktif <b>{{ angka(data.kpi.totalEpisode) }}</b></span>
            <span class="chip-hero" v-if="bisaAkses.finansial">% Rugi Rata-rata <b>{{ persen(data.kpi.pctRugiRata) }}</b></span>
            <span class="chip-hero">LOS Rata-rata <b>{{ data.kpi.losRata.toFixed(1) }}</b> hari</span>
          </div>
        </div>

        <div class="hero-komponen">
          <div class="label-hero" style="margin-bottom: 12px">Komponen Readiness</div>
          <div
            v-for="(bobot, k) in data.readiness?.komponen ?? {}"
            :key="k"
            style="margin-bottom: 11px"
          >
            <div style="display: flex; justify-content: space-between; font-size: 12px; margin-bottom: 5px; color: #e8d9ba">
              <span style="font-weight: 600">{{ labelReady[k] ?? k }}</span>
              <b style="color: #fff">{{ bobot }}</b>
            </div>
            <div class="bar-hero"><div :style="{ width: (bobot * 100) / 30 + '%' }" /></div>
          </div>
        </div>
      </div>
    </div>

    <!-- ====== KPI ====== -->
    <div class="grid kpi" style="margin-bottom: 16px">
      <KpiCard
        label="Total Episode Aktif"
        :nilai="angka(data.kpi.totalEpisode)"
        :catatan="`${angka(data.kpi.jumlahUpload)} upload · LOS rata-rata ${data.kpi.losRata.toFixed(1)} hari`"
        ikon="lapis"
        warna="brand"
      />
      <KpiCard
        v-if="bisaAkses.finansial"
        label="Estimasi Rugi (Σ billing > tarif)"
        :nilai="rupiahSingkat(data.kpi.estimasiRugi)"
        :catatan="`dari total tarif ${rupiahSingkat(data.kpi.totalTarif)}`"
        ikon="turun"
        warna="merah"
        garis="var(--merah)"
        nilai-warna="var(--merah)"
      />
      <KpiCard
        v-if="bisaAkses.finansial"
        label="Potensi Recovery"
        :nilai="rupiahSingkat(data.kpi.potensiRecovery)"
        catatan="estimasi upgrade severity antar-grup"
        ikon="naik"
        warna="hijau"
        garis="var(--hijau)"
        nilai-warna="var(--hijau)"
      />
      <KpiCard
        v-if="bisaAkses.finansial"
        label="% Rugi Rata-rata (Episode Rugi)"
        :nilai="persen(data.kpi.pctRugiRata)"
        :catatan="`selisih total ${rupiahSingkat(data.kpi.totalSelisih)}`"
        ikon="persen"
        warna="kuning"
        garis="var(--kuning)"
      />
      <KpiCard
        v-if="bisaAkses.finansial"
        label="Anomali Terdeteksi"
        :nilai="angka(data.kpi.anomaliJumlah)"
        catatan="LOS outlier · duplikat · readmisi · ekstrem"
        ikon="peringatan"
        warna="kuning"
        garis="var(--kuning)"
        nilai-warna="var(--kuning)"
      />
      <KpiCard
        v-if="bisaAkses.ledger"
        label="Ledger Aktif"
        :nilai="angka(data.kpi.ledgerAktif)"
        catatan="kandidat s.d. diajukan"
        ikon="buku"
        warna="ungu"
        garis="var(--ungu)"
      />
    </div>

    <!-- ====== Tren + Tindakan Prioritas ====== -->
    <div class="grid dua-lebar" style="margin-bottom: 16px">
      <div class="kartu">
        <div class="kepala-kartu">
          <h3>Tren Episode &amp; Rugi per Periode</h3>
        </div>
        <GrafikTren
          v-if="seriTren.length"
          :label="data.tren.map((t) => t.periode)"
          :seri="seriTren"
          :tinggi="252"
          :format-kanan="(n: number) => rupiahSingkat(n)"
        />
        <div v-else class="kosong">Belum ada data tren.</div>
      </div>

      <div class="kartu">
        <div class="kepala-kartu">
          <h3>Tindakan Prioritas</h3>
          <span class="badge" :class="prioritas.length ? 'merah' : 'hijau'">{{ prioritas.length }}</span>
        </div>
        <div v-if="prioritas.length" style="max-height: 268px; overflow-y: auto; padding-right: 2px">
          <button
            v-for="(p, i) in prioritas"
            :key="i"
            class="prioritas-item"
            :class="p.tingkat"
            type="button"
            @click="router.push(p.ke)"
          >
            <span class="pi-ikon" :class="p.tingkat"><Ikon :nama="p.ikon" :ukuran="16" /></span>
            <span class="pi-badan">
              <b>{{ p.judul }}</b>
              <span>{{ p.desk }}</span>
            </span>
            <span class="badge" :class="p.badge">{{ p.label }}</span>
            <Ikon nama="chevronKanan" :ukuran="15" class="chev" />
          </button>
        </div>
        <div v-else class="kosong" style="padding: 34px 16px">
          <div style="display: grid; place-items: center; gap: 8px">
            <span style="width: 42px; height: 42px; border-radius: 50%; background: var(--hijau-bg); color: var(--hijau); display: grid; place-items: center">
              <Ikon nama="ceklis" :ukuran="20" />
            </span>
            Semua tindakan selesai — tidak ada prioritas tertunda.
          </div>
        </div>
      </div>
    </div>

    <!-- ====== Distribusi + Kelas Rawat ====== -->
    <div class="grid dua" style="margin-bottom: 16px">
      <div class="kartu" v-if="bisaAkses.finansial">
        <div class="kepala-kartu">
          <h3>Distribusi Kelas Kerugian</h3>
        </div>
        <GrafikDonut
          :potongan="['SURPLUS', 'S1_CODING', 'S2_MIXED', 'STRUCTURAL'].map((k) => ({ label: labelRugi[k], nilai: data!.distribusiRugi[k] ?? 0, warna: warnaRugi[k] }))"
          :teks-tengah="angka(data.kpi.totalEpisode)"
        />
      </div>
      <div class="kartu">
        <div class="kepala-kartu">
          <h3>Kelas Rawat</h3>
          <span class="badge abu">{{ angka(totalKelas) }} episode</span>
        </div>
        <div v-if="(data.kelasRawat ?? []).length" style="padding-top: 4px">
          <div v-for="k in data.kelasRawat" :key="k.kelas" class="kelas-baris">
            <span class="kb-label">Kelas {{ k.kelas }}</span>
            <div class="bar-rugi">
              <div :style="{ width: (k.jumlah / Math.max(...data.kelasRawat.map((x) => x.jumlah), 1)) * 100 + '%', background: 'linear-gradient(90deg, #c07f1f, #f0a640)' }" />
            </div>
            <b class="kb-nilai">{{ angka(k.jumlah) }}</b>
          </div>
        </div>
        <div v-else class="kosong">Belum ada data kelas rawat.</div>
      </div>
    </div>

    <!-- ====== Top CBG ====== -->
    <div class="kartu">
      <div class="kepala-kartu">
        <h3>Grup INA-CBG Terbesar</h3>
        <button class="tautan-kartu" type="button" @click="router.push('/klaim')">
          Claim Explorer
          <Ikon nama="chevronKanan" :ukuran="13" />
        </button>
      </div>
      <div class="tabel-wadah gulir-y" style="max-height: 360px">
        <table class="tabel">
          <thead>
            <tr><th>Kode</th><th>Deskripsi</th><th class="kanan">Episode</th><th class="kanan">Rugi</th></tr>
          </thead>
          <tbody>
            <tr v-for="c in data.topCbg ?? []" :key="c.kode">
              <td class="tunggal">{{ c.kode }}</td>
              <td>{{ c.desk }}</td>
              <td class="kanan"><b>{{ c.jumlah }}</b></td>
              <td class="kanan"><span class="badge" :class="c.rugi > 0 ? 'merah' : 'hijau'">{{ c.rugi }}</span></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </template>
</template>

<style scoped>
.hero-grid {
  display: flex; gap: 36px; flex-wrap: wrap; align-items: center;
  position: relative; z-index: 1;
}
.hero-ring { flex-shrink: 0; }
.hero-info { min-width: 225px; }
.hero-komponen { flex: 1; min-width: 260px; }
@media (max-width: 700px) {
  .hero-grid { gap: 20px; }
}
</style>
