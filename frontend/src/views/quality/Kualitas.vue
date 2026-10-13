<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { api, modeDemo } from "../api";
import { store } from "../store";
import { tanggalWaktu, angka } from "../format";
import type { Upload, DQReport } from "../types";
import PageHeader from "../components/PageHeader.vue";

const uploads = ref<Upload[]>([]);
const terpilih = ref<Upload | null>(null);
const memuat = ref(true);

async function muat() {
  memuat.value = true;
  try {
    const d = await api.get<{ items: Upload[] }>("/api/uploads");
    uploads.value = d.items.filter((u) => u.status !== "DIHAPUS");
    if (!terpilih.value && uploads.value.length) terpilih.value = uploads.value[0];
  } finally {
    memuat.value = false;
  }
}
onMounted(muat);

const dq = computed<DQReport | null>(() => terpilih.value?.dqReport ?? null);
const warnaGrade: Record<string, string> = { A: "hijau", B: "teal", C: "kuning", D: "merah" };
const ikonSev: Record<string, string> = { KRITIS: "⛔", PERINGATAN: "⚠️", INFO: "ℹ️" };
</script>

<template>
  <PageHeader judul="Claim Quality" :demo="modeDemo" sub="Laporan Data Quality per upload batch — skor 7 dimensi dan 20 aturan DQ." />
  <div v-if="memuat" class="pemuat"><div class="spinner" /><span>Memuat…</span></div>
  <div v-else class="grid" style="grid-template-columns: 340px 1fr">
    <div class="kartu" style="padding: 12px">
      <h3 style="padding: 4px 6px 10px">Upload batches</h3>
      <div style="max-height: 640px; overflow-y: auto; display: flex; flex-direction: column; gap: 6px">
        <div
          v-for="u in uploads"
          :key="u.id"
          @click="terpilih = u"
          style="padding: 10px 12px; border-radius: 8px; cursor: pointer; border: 1.5px solid var(--border)"
          :style="terpilih?.id === u.id ? 'border-color: var(--teal); background: var(--teal-100)' : ''"
        >
          <b style="font-size: 13px">{{ u.namaFile }}</b>
          <div style="font-size: 11.5px; color: var(--text-3); margin-top: 2px">
            {{ u.hospitalId }} · {{ u.periode }} · {{ u.format }}
          </div>
          <div style="margin-top: 5px; display: flex; gap: 5px; align-items: center">
            <span class="badge" :class="warnaGrade[u.dqGrade]">{{ u.dqScore }} · {{ u.dqGrade }}</span>
            <span class="badge abu">{{ angka(u.totalBaris) }} baris</span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="terpilih && dq">
      <div class="grid tiga" style="margin-bottom: 14px">
        <div class="kartu" style="text-align: center">
          <div class="kpi-kartu" style="align-items: center">
            <span class="label">SKOR DQ</span>
            <span class="nilai" style="font-size: 40px">{{ dq.skor }}</span>
            <span class="badge" :class="warnaGrade[dq.grade]" style="font-size: 13px; padding: 4px 12px">Grade {{ dq.grade }}</span>
          </div>
        </div>
        <div class="kartu" style="text-align: center">
          <div class="kpi-kartu" style="align-items: center">
            <span class="label">GERBANG</span>
            <span class="nilai" style="font-size: 22px; margin: 8px 0">{{ dq.gerbang }}</span>
            <span class="badge" :class="dq.gerbang === 'BLOKIR' ? 'merah' : dq.gerbang === 'OVERRIDE' ? 'kuning' : 'hijau'">
              {{ terpilih.overrideUsed ? "override digunakan" : "tanpa override" }}
            </span>
          </div>
        </div>
        <div class="kartu">
          <div class="kpi-kartu">
            <span class="label">RINGKASAN</span>
            <span class="catatan" style="font-size: 12.5px; line-height: 1.9">
              File: <b>{{ terpilih.namaFile }}</b><br />
              Periode: <b>{{ terpilih.periode }}</b> (usulan: {{ dq.periodeUsulan }})<br />
              Baris: <b>{{ angka(terpilih.totalBaris) }}</b> · gagal <b>{{ angka(terpilih.barisGagal) }}</b><br />
              Ingest: <b style="color: var(--teal)">{{ angka(terpilih.barisBaru) }}</b> baru ·
              <b style="color: var(--kuning)">{{ angka(terpilih.barisBerubah) }}</b> ubah ·
              <b>{{ angka(terpilih.barisUnchanged) }}</b> sama<br />
              Waktu: {{ tanggalWaktu(terpilih.createdAt) }}
            </span>
          </div>
        </div>
      </div>

      <div class="kartu" style="margin-bottom: 14px">
        <h3>Skor per dimensi</h3>
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 6px 22px">
          <div v-for="d in dq.dimensi" :key="d.kode" style="display: flex; align-items: center; gap: 10px">
            <span style="width: 130px; font-size: 12.5px; color: var(--text-2); font-weight: 600">{{ d.label }}</span>
            <div class="bar-rugi" style="flex: 1; height: 8px"><div :style="{ width: d.skor + '%', background: d.skor >= 90 ? 'var(--hijau)' : d.skor >= 60 ? 'var(--kuning)' : 'var(--merah)' }" /></div>
            <b style="width: 34px; text-align: right; font-size: 12.5px">{{ d.skor }}</b>
          </div>
        </div>
      </div>

      <div class="kartu">
        <h3>Aturan DQ ({{ dq.aturan.filter((a) => a.terpicu).length }} dari 20 terpicu)</h3>
        <div style="max-height: 360px; overflow-y: auto">
          <div v-for="a in dq.aturan" :key="a.kode" class="aturan-dq" :class="{ terpicu: a.terpicu }">
            <span class="ikon">{{ a.terpicu ? ikonSev[a.severitas] : "✅" }}</span>
            <div style="flex: 1">
              <b>{{ a.nama }}</b>
              <span class="badge" :class="a.severitas === 'KRITIS' ? 'merah' : a.severitas === 'PERINGATAN' ? 'kuning' : 'abu'" style="margin-left: 6px">{{ a.severitas }}</span>
              <div class="pesan">{{ a.pesan }} <span class="kode">· {{ a.kode }}</span></div>
            </div>
            <b v-if="a.terpicu" style="color: var(--merah)">{{ a.jumlah }}</b>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="kartu"><div class="kosong">Pilih upload di kiri untuk melihat laporan DQ lengkap.</div></div>
  </div>
</template>
