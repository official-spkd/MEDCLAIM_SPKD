<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { api, modeDemo } from "../api";
import { store, queryFilter } from "../store";
import { tanggal } from "../format";
import { angka } from "../format";
import PageHeader from "../components/PageHeader.vue";

interface Anomali {
  episodeId: string;
  jenis: string;
  detail: string;
  sep: string;
  mr: string;
  nama: string;
  icbg: string;
  los: number;
  tglMasuk: string;
}

const items = ref<Anomali[]>([]);
const jenis = ref("");
const memuat = ref(true);
const galat = ref("");

const labelJenis: Record<string, string> = {
  LOS_OUTLIER: "LOS Outlier",
  READMISI: "Readmisi ≤30 hari",
  DUPLIKAT: "Duplikat SEP",
  SELISIH_EKSTREM: "Selisih Ekstrem",
};
const badgeJenis: Record<string, string> = {
  LOS_OUTLIER: "kuning", READMISI: "ungu", DUPLIKAT: "merah", SELISIH_EKSTREM: "teal",
};

async function muat() {
  memuat.value = true;
  galat.value = "";
  try {
    const t = queryFilter(jenis.value ? { jenis: jenis.value } : {});
    const d = await api.get<{ total: number; items: Anomali[] }>("/api/anomali" + t);
    items.value = d.items.filter((a) => !jenis.value || a.jenis === jenis.value);
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal memuat anomali.";
  } finally {
    memuat.value = false;
  }
}
onMounted(muat);
watch(() => [store.filter.rs, store.filter.tipe, store.filter.dari, store.filter.sampai], muat);
watch(jenis, muat);
</script>

<template>
  <PageHeader judul="Anomaly Detection" :demo="modeDemo" sub="LOS outlier, duplikat SEP, readmisi ≤30 hari, dan selisih ekstrem yang perlu verifikasi." />
  <div class="filter-bar">
    <div class="kolom-form" style="max-width: 220px">
      <label>Jenis anomali</label>
      <select v-model="jenis">
        <option value="">Semua jenis</option>
        <option v-for="(l, k) in labelJenis" :key="k" :value="k">{{ l }}</option>
      </select>
    </div>
    <div style="flex: 1" />
    <span class="badge merah" style="font-size: 12.5px; padding: 5px 14px">{{ angka(items.length) }} anomali</span>
  </div>

  <div v-if="memuat" class="pemuat"><div class="spinner" /><span>Menjalankan deteksi anomali…</span></div>
  <div v-else-if="galat" class="bahaya-kecil">{{ galat }}</div>
  <div v-else style="display: flex; flex-direction: column; gap: 8px">
    <div v-for="(a, i) in items" :key="i" class="kartu" style="display: flex; gap: 14px; align-items: center; flex-wrap: wrap; padding: 13px 16px">
      <span class="badge" :class="badgeJenis[a.jenis]" style="min-width: 150px; justify-content: center">{{ labelJenis[a.jenis] ?? a.jenis }}</span>
      <div style="min-width: 190px">
        <b>{{ a.nama }}</b>
        <div style="font-size: 11.5px; color: var(--text-3)">{{ a.mr }} · {{ tanggal(a.tglMasuk) }} · LOS {{ a.los }} hari</div>
      </div>
      <span class="badge navy">{{ a.icbg }}</span>
      <span class="tunggal">{{ a.sep }}</span>
      <div style="flex: 1" />
      <span style="font-size: 12.5px; color: var(--text-2); max-width: 420px">{{ a.detail }}</span>
    </div>
    <div v-if="items.length === 0" class="kartu"><div class="kosong">Tidak ada anomali yang cocok — data dalam kondisi sehat. ✨</div></div>
  </div>
</template>
