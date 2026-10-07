<script setup lang="ts">
import { ref, onMounted, watch, computed } from "vue";
import { api, modeDemo } from "../api";
import { store, queryFilter, bisaAkses } from "../store";
import { rupiah, tanggal, usia, angka } from "../format";
import type { Episode } from "../types";
import PageHeader from "../components/PageHeader.vue";

const items = ref<Episode[]>([]);
const cari = ref("");
const memuat = ref(true);
const galat = ref("");

async function muat() {
  memuat.value = true;
  galat.value = "";
  try {
    items.value = (await api.get<{ items: Episode[] }>("/api/episodes" + queryFilter())).items;
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal memuat pasien.";
  } finally {
    memuat.value = false;
  }
}
onMounted(muat);
watch(() => [store.filter.rs, store.filter.tipe, store.filter.dari, store.filter.sampai], muat);

// kelompokkan per MR
interface Pasien {
  mr: string;
  nama: string;
  lahir: string;
  jk: string;
  episodes: Episode[];
}
const pasien = computed<Pasien[]>(() => {
  const map = new Map<string, Pasien>();
  for (const ep of items.value) {
    let p = map.get(ep.noMr);
    if (!p) {
      p = { mr: ep.noMr, nama: ep.namaPasien, lahir: ep.tglLahir, jk: ep.jenisKelamin, episodes: [] };
      map.set(ep.noMr, p);
    }
    p.episodes.push(ep);
  }
  const daftar = [...map.values()];
  daftar.sort((a, b) => b.episodes.length - a.episodes.length || a.mr.localeCompare(b.mr));
  return daftar;
});

const cariKecil = computed(() => cari.value.trim().toLowerCase());
const terfilter = computed(() =>
  pasien.value.filter(
    (p) =>
      !cariKecil.value ||
      p.mr.toLowerCase().includes(cariKecil.value) ||
      p.nama.toLowerCase().includes(cariKecil.value)
  )
);
const terpotong = computed(() => terfilter.value.slice(0, 60));

const terbuka = ref<string | null>(null);
function toggle(mr: string) {
  terbuka.value = terbuka.value === mr ? null : mr;
}

function tanggalDekat(a: string, b: string): boolean {
  const d1 = new Date(a).getTime();
  const d2 = new Date(b).getTime();
  return Math.abs(d2 - d1) <= 30 * 86400000;
}
</script>

<template>
  <PageHeader judul="Patient Tracer" :demo="modeDemo" sub="Lacak riwayat episode per pasien — kunjungan berulang dan readmisi ≤30 hari." />
  <div class="filter-bar">
    <div class="kolom-form cari">
      <label>Cari pasien (MR / nama)</label>
      <input v-model="cari" placeholder="mis. RM-10103, Andi" />
    </div>
    <div style="flex: 1" />
    <span class="badge navy">{{ angka(terfilter.length) }} pasien unik</span>
  </div>

  <div v-if="memuat" class="pemuat"><div class="spinner" /><span>Memuat tracer…</span></div>
  <div v-else-if="galat" class="bahaya-kecil">{{ galat }}</div>
  <div v-else style="display: flex; flex-direction: column; gap: 8px">
    <div v-for="p in terpotong" :key="p.mr" class="kartu" style="padding: 0; overflow: hidden">
      <div
        style="display: flex; align-items: center; gap: 14px; padding: 13px 18px; cursor: pointer; flex-wrap: wrap"
        @click="toggle(p.mr)"
      >
        <div class="avatar" style="background: var(--navy)">{{ p.nama.slice(0, 1) }}</div>
        <div style="min-width: 180px">
          <b>{{ p.nama }}</b>
          <div style="font-size: 11.5px; color: var(--text-3)">{{ p.mr }} · {{ usia(p.lahir) }} th · {{ p.jk }}</div>
        </div>
        <span class="badge" :class="p.episodes.length > 1 ? 'kuning' : 'abu'">{{ p.episodes.length }} episode</span>
        <span v-if="p.episodes.length > 1" class="badge" :class="p.episodes.some((e) => e.kelasRugi !== 'SURPLUS' && e.kelasRugi !== 'RAHASIA') ? 'merah' : 'hijau'">
          {{ bisaAkses.finansial ? p.episodes.filter((e) => e.kelasRugi !== "SURPLUS" && e.kelasRugi !== "RAHASIA").length + " rugi" : "riwayat berulang" }}
        </span>
        <div style="flex: 1" />
        <span style="color: var(--text-3); font-size: 12px">{{ terbuka === p.mr ? "▲" : "▼" }}</span>
      </div>

      <div v-if="terbuka === p.mr" style="border-top: 1px solid var(--border); padding: 14px 18px; background: var(--bg)">
        <div style="display: flex; flex-direction: column; gap: 8px">
          <div
            v-for="ep in [...p.episodes].sort((a, b) => a.tglMasuk.localeCompare(b.tglMasuk))"
            :key="ep.id"
            style="display: flex; gap: 12px; align-items: center; background: var(--surface); border: 1px solid var(--border); border-radius: 8px; padding: 10px 14px; flex-wrap: wrap"
          >
            <div style="min-width: 210px">
              <b style="font-size: 12.5px">{{ tanggal(ep.tglMasuk) }} → {{ tanggal(ep.tglKeluar) }}</b>
              <div style="font-size: 11.5px; color: var(--text-3)">LOS {{ ep.los }} hari · {{ ep.tipeLayanan }} · {{ ep.noSep }}</div>
            </div>
            <span class="badge navy">{{ ep.kodeIcbg }}</span>
            <span v-for="d in ep.diagnosa.slice(0, 3)" :key="d.kode" class="badge abu">{{ d.kode }}</span>
            <template v-if="bisaAkses.finansial">
              <div style="flex: 1" />
              <b :style="{ color: ep.selisih >= 0 ? 'var(--hijau)' : 'var(--merah)', fontSize: '12.5px' }">{{ rupiah(ep.selisih) }}</b>
            </template>
          </div>
          <div v-if="p.episodes.length > 1" class="info-kecil">
            Kunjungan berulang terdeteksi — jika ≤ 30 hari dengan diagnosis terkait, modul Anomaly menandai readmisi.
          </div>
        </div>
      </div>
    </div>
    <div v-if="terfilter.length > 60" class="info-kecil">Menampilkan 60 dari {{ angka(terfilter.length) }} pasien — persempit dengan pencarian.</div>
  </div>
</template>
