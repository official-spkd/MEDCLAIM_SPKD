<script setup lang="ts">
import { ref, watch, onMounted } from "vue";
import { api, modeDemo } from "../api";
import { store, queryFilter } from "../store";
import { rupiah, rupiahSingkat, persen } from "../format";
import { LABEL_KOMPONEN } from "../types";
import GrafikDonut from "../components/GrafikDonut.vue";
import PageHeader from "../components/PageHeader.vue";

interface Fin {
  total: number;
  komponen: { komponen: string; nilai: number; pct: number }[];
}

const data = ref<Fin | null>(null);
const memuat = ref(true);
const galat = ref("");

const warna = ["#f0a640", "#0f766e", "#8a5a14", "#dc2626", "#7c3aed", "#059669", "#ea580c", "#0284c7", "#65a30d", "#c026d3", "#78716c", "#b45309"];

async function muat() {
  memuat.value = true;
  galat.value = "";
  try {
    data.value = await api.get<Fin>("/api/financial" + queryFilter());
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal memuat data finansial.";
  } finally {
    memuat.value = false;
  }
}
onMounted(muat);
watch(() => [store.filter.rs, store.filter.tipe, store.filter.dari, store.filter.sampai], muat);
</script>

<template>
  <PageHeader judul="Financial Intelligence" :demo="modeDemo" sub="Komposisi komponen billing format Detail — porsi tiap komponen terhadap total." />
  <div v-if="memuat" class="pemuat"><div class="spinner" /><span>Memuat finansial…</span></div>
  <div v-else-if="galat" class="bahaya-kecil">{{ galat }}</div>
  <template v-else-if="data">
    <div class="kartu kpi-kartu" style="margin-bottom: 14px; max-width: 480px">
      <span class="label">TOTAL BILLING SELURUH EPISODE TERFILTER</span>
      <span class="nilai" style="font-size: 30px">{{ rupiah(data.total) }}</span>
      <span class="catatan">agregasi 18 komponen billing format Detail — format Basic terbatas</span>
    </div>

    <div class="grid dua">
      <div class="kartu">
        <h3>Komposisi Komponen Billing</h3>
        <GrafikDonut
          :potongan="data.komponen.filter((k) => k.nilai > 0).slice(0, 10).map((k, i) => ({
            label: LABEL_KOMPONEN[k.komponen] ?? k.komponen,
            nilai: k.nilai,
            warna: warna[i % warna.length],
          }))"
          :ukuran="190"
          :teks-tengah="rupiahSingkat(data.total)"
          sub-tengah="total billing"
        />
      </div>
      <div class="kartu">
        <h3>Rincian per Komponen</h3>
        <div class="tabel-wadah" style="max-height: 420px; overflow-y: auto">
          <table class="tabel">
            <thead><tr><th>Komponen</th><th class="kanan">Nilai</th><th style="width: 200px">Porsi</th></tr></thead>
            <tbody>
              <tr v-for="k in data.komponen" :key="k.komponen">
                <td style="font-weight: 600">{{ LABEL_KOMPONEN[k.komponen] ?? k.komponen }}</td>
                <td class="kanan">{{ rupiah(k.nilai) }}</td>
                <td>
                  <div style="display: flex; align-items: center; gap: 8px">
                    <div class="bar-rugi" style="flex: 1"><div :style="{ width: (k.pct * 100).toFixed(1) + '%', background: 'var(--navy)' }" /></div>
                    <b style="font-size: 12px; width: 52px; text-align: right">{{ persen(k.pct) }}</b>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="peringatan-kecil" style="margin-top: 14px">
      Catatan analitik: komponen yang mendominasi (&gt;80% dari total billing pada satu episode) otomatis ditandai modul
      Anomaly Detection sebagai SELISIH_EKSTREM untuk verifikasi rincian.
    </div>
  </template>
</template>
