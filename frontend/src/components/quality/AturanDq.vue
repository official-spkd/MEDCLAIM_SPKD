<script setup lang="ts">
// Daftar aturan DQ. Aturan terpicu bisa diklik untuk melihat baris yang perlu
// diperbaiki; rekap seluruh baris bermasalah dapat diunduh sebagai CSV.
import { ref, computed } from "vue";
import { angka } from "@/utils/format";
import type { DQAturan } from "@/types";

const props = defineProps<{
  aturan: DQAturan[];
  namaFile?: string;
  tinggi?: number;
  tampilDimensi?: boolean;
}>();

const ikonSev: Record<string, string> = { KRITIS: "⛔", PERINGATAN: "⚠️", INFO: "ℹ️" };
const terpilih = ref<DQAturan | null>(null);

const adaRincian = computed(() => props.aturan.some((a) => a.terpicu && a.baris?.length));

function bisaDibuka(a: DQAturan): boolean {
  return a.terpicu && !!a.baris?.length;
}

function csvSel(v: string | number): string {
  const t = String(v ?? "");
  return /[",\n;]/.test(t) ? `"${t.replace(/"/g, '""')}"` : t;
}

function unduhRekap() {
  const kepala = ["Kode Aturan", "Nama Aturan", "Severitas", "No. Baris File", "No. SEP", "No. MR", "Nama Pasien", "Keterangan"];
  const baris: string[] = [kepala.map(csvSel).join(",")];
  for (const a of props.aturan) {
    if (!a.terpicu) continue;
    if (!a.baris?.length) {
      baris.push([a.kode, a.nama, a.severitas, "", "", "", "", a.pesan].map(csvSel).join(","));
      continue;
    }
    for (const b of a.baris) {
      baris.push([a.kode, a.nama, a.severitas, b.nomor, b.noSep, b.noMr, b.nama, b.keterangan].map(csvSel).join(","));
    }
  }
  // BOM agar Excel membaca UTF-8 dengan benar.
  const blob = new Blob(["﻿" + baris.join("\r\n")], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `rekap-dq-${(props.namaFile ?? "upload").replace(/\.[^.]+$/, "")}.csv`;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}
</script>

<template>
  <div v-if="adaRincian" style="display: flex; align-items: center; gap: 10px; margin-bottom: 10px">
    <span style="flex: 1; font-size: 12px; color: var(--text-3)">Klik aturan bertanda angka merah untuk melihat baris yang perlu diperbaiki.</span>
    <button class="btn garis kecil" type="button" @click="unduhRekap">Unduh rekap (CSV)</button>
  </div>

  <div :style="{ maxHeight: (tinggi ?? 340) + 'px', overflowY: 'auto' }">
    <div
      v-for="a in aturan"
      :key="a.kode"
      class="aturan-dq"
      :class="{ terpicu: a.terpicu, 'kartu-klik': bisaDibuka(a) }"
      :role="bisaDibuka(a) ? 'button' : undefined"
      :tabindex="bisaDibuka(a) ? 0 : undefined"
      @click="bisaDibuka(a) && (terpilih = a)"
      @keydown.enter="bisaDibuka(a) && (terpilih = a)"
    >
      <span class="ikon">{{ a.terpicu ? ikonSev[a.severitas] : "✅" }}</span>
      <div style="flex: 1">
        <b>{{ a.nama }}</b>
        <span class="badge" :class="a.severitas === 'KRITIS' ? 'merah' : a.severitas === 'PERINGATAN' ? 'kuning' : 'abu'" style="margin-left: 6px">{{ a.severitas }}</span>
        <div class="pesan">{{ a.pesan }} <span class="kode">· {{ a.kode }}<template v-if="tampilDimensi"> · {{ a.dimensi }}</template></span></div>
      </div>
      <b v-if="a.terpicu" style="color: var(--merah)">{{ a.jumlah }}</b>
    </div>
  </div>

  <!-- Rincian baris per aturan -->
  <div v-if="terpilih" class="modal-topeng" @click.self="terpilih = null">
    <div class="modal" style="max-width: 860px">
      <div class="kepala">
        <h3>{{ terpilih.nama }}</h3>
        <button class="btn garis kecil" @click="terpilih = null">✕</button>
      </div>
      <div class="badan">
        <p style="margin: 0 0 12px; font-size: 13px; color: var(--text-2)">
          {{ terpilih.pesan }} <span class="kode" style="font-family: ui-monospace, monospace; font-size: 11px">· {{ terpilih.kode }}</span>
        </p>
        <div class="tabel-wadah gulir-y" style="max-height: 420px">
          <table class="tabel">
            <thead><tr><th class="kanan">Baris</th><th>No. SEP</th><th>No. MR</th><th>Pasien</th><th>Yang perlu diperbaiki</th></tr></thead>
            <tbody>
              <tr v-for="(b, i) in terpilih.baris" :key="i">
                <td class="kanan tunggal">{{ b.nomor }}</td>
                <td class="tunggal">{{ b.noSep }}</td>
                <td class="tunggal">{{ b.noMr }}</td>
                <td>{{ b.nama }}</td>
                <td style="font-size: 12.5px">{{ b.keterangan }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-if="terpilih.terpotong" class="info-kecil" style="margin-top: 10px">
          Menampilkan {{ angka(terpilih.baris?.length ?? 0) }} dari {{ angka(terpilih.jumlah) }} baris.
        </div>
      </div>
      <div class="kaki"><button class="btn garis" @click="terpilih = null">Tutup</button></div>
    </div>
  </div>
</template>
