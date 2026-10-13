<script setup lang="ts">
// ============================================================
// MedClaim — KpiCard: kartu metrik ala MedCredix
// (label uppercase + nilai besar + ikon kanan-atas lingkaran)
// ============================================================
import Ikon from "./Ikon.vue";

withDefaults(
  defineProps<{
    label: string;
    nilai: string;
    catatan?: string;
    ikon?: string;
    warna?: string; // kelas varian ikon: brand | teal | hijau | merah | kuning | ungu | info
    garis?: string; // warna css garis aksen bawah
    nilaiWarna?: string; // warna css angka
  }>(),
  { warna: "brand", garis: "var(--brand)" },
);
</script>

<template>
  <div class="kartu kpi-kartu">
    <span class="label">{{ label }}</span>
    <span class="nilai" :style="nilaiWarna ? { color: nilaiWarna } : undefined">{{ nilai }}</span>
    <span v-if="catatan" class="catatan">{{ catatan }}</span>
    <div class="kpi-ikon" :class="`ikon-${warna}`">
      <Ikon :nama="ikon ?? 'aktivitas'" :ukuran="19" />
    </div>
    <div class="garis" :style="{ background: garis }" />
  </div>
</template>

<style scoped>
.kpi-ikon.ikon-brand { background: var(--brand-light); color: var(--brand-gelap); }
.kpi-ikon.ikon-teal { background: var(--teal-100); color: var(--teal); }
.kpi-ikon.ikon-hijau { background: var(--hijau-bg); color: var(--hijau); }
.kpi-ikon.ikon-merah { background: var(--merah-bg); color: var(--merah); }
.kpi-ikon.ikon-kuning { background: var(--kuning-bg); color: var(--kuning); }
.kpi-ikon.ikon-ungu { background: var(--ungu-bg); color: var(--ungu); }
.kpi-ikon.ikon-info { background: var(--info-bg); color: var(--info); }
.dark .kpi-ikon.ikon-brand { color: #f5b95e; }
.dark .kpi-ikon.ikon-teal { color: #5fd0c5; }
</style>
