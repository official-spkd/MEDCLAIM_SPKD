<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { api, ApiError, modeDemo } from "../api";
import { store, bisaAkses } from "../store";
import { rupiah, rupiahSingkat, tanggalWaktu, angka, persen } from "../format";
import { LABEL_KELAS_RUGI, LABEL_STATUS_LEDGER } from "../types";
import type { Ledger, Kalibrasi, Episode } from "../types";
import PageHeader from "../components/PageHeader.vue";

const items = ref<Ledger[]>([]);
const kalibrasi = ref<Kalibrasi | null>(null);
const status = ref("SEMUA");
const cari = ref("");
const memuat = ref(true);
const galat = ref("");

const badgeStatus: Record<string, string> = {
  KANDIDAT: "abu", DIPERIKSA: "teal", DIKOREKSI: "kuning", DIAJUKAN: "ungu",
  HASIL: "hijau", TIDAK_BERUBAH: "navy",
};

async function muat() {
  memuat.value = true;
  galat.value = "";
  try {
    const p = new URLSearchParams();
    if (status.value !== "SEMUA") p.set("status", status.value);
    if (cari.value) p.set("q", cari.value);
    const d = await api.get<{ items: Ledger[]; kalibrasi: Kalibrasi }>("/api/ledger" + (p.toString() ? `?${p}` : ""));
    items.value = d.items;
    kalibrasi.value = d.kalibrasi;
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal memuat ledger.";
  } finally {
    memuat.value = false;
  }
}
onMounted(muat);
let timer: ReturnType<typeof setTimeout> | undefined;
watch(status, muat);
watch(cari, () => { clearTimeout(timer); timer = setTimeout(muat, 350); });

// ---------- Detail + transisi ----------
const terpilih = ref<Ledger | null>(null);
const memuatDetail = ref(false);
const galatTransisi = ref("");
const form = ref({ ke: "", alasan: "", referensiRm: "", realisasi: "" });

async function buka(l: Ledger) {
  memuatDetail.value = true;
  terpilih.value = null;
  galatTransisi.value = "";
  form.value = { ke: "", alasan: "", referensiRm: "", realisasi: "" };
  try {
    terpilih.value = await api.get<Ledger>(`/api/ledger/${l.id}`);
  } finally {
    memuatDetail.value = false;
  }
}

const transisiSah: Record<string, string[]> = {
  KANDIDAT: ["DIPERIKSA"],
  DIPERIKSA: ["DIKOREKSI", "TIDAK_BERUBAH", "KANDIDAT"],
  DIKOREKSI: ["DIAJUKAN", "DIPERIKSA"],
  DIAJUKAN: ["HASIL", "DIKOREKSI"],
  HASIL: ["DIPERIKSA"],
  TIDAK_BERUBAH: ["DIPERIKSA"],
};

const perluReferensi = () => form.value.ke === "DIKOREKSI";
const perluAlasan = () => form.value.ke === "TIDAK_BERUBAH" || (form.value.ke === "DIPERIKSA" && (terpilih.value?.status === "HASIL" || terpilih.value?.status === "TIDAK_BERUBAH"));
const perluRealisasi = () => form.value.ke === "HASIL";

async function transisi() {
  if (!terpilih.value) return;
  galatTransisi.value = "";
  try {
    const body: Record<string, unknown> = { ke: form.value.ke, alasan: form.value.alasan, referensiRm: form.value.referensiRm };
    if (form.value.realisasi !== "") body.realisasi = Number(form.value.realisasi);
    terpilih.value = await api.post<Ledger>(`/api/ledger/${terpilih.value.id}/transisi`, body);
    form.value = { ke: "", alasan: "", referensiRm: "", realisasi: "" };
    await muat();
  } catch (e: any) {
    galatTransisi.value = e?.message ?? "Transisi ditolak.";
  }
}

// ---------- Buat kandidat manual ----------
const bukaBuat = ref(false);
const sepBuat = ref("");
const episodeDitemukan = ref<Episode | null>(null);
const cariEpisode = ref<Episode[]>([]);
async function cariSep() {
  const q = sepBuat.value.trim();
  if (!q) return;
  const d = await api.get<{ items: Episode[] }>("/api/episodes?q=" + encodeURIComponent(q));
  cariEpisode.value = d.items.filter((e) => e.kelasRugi !== "SURPLUS").slice(0, 8);
  episodeDitemukan.value = d.items.find((e) => e.noSep === q) ?? null;
}
async function buatKandidat(ep: Episode) {
  try {
    await api.post("/api/ledger", { episodeId: ep.id });
    okPesan("Kandidat dibuat untuk SEP " + ep.noSep);
    bukaBuat.value = false;
    await muat();
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal membuat kandidat.";
  }
}
function okPesan(p: string) { galat.value = ""; setTimeout(() => (pesanOk.value = ""), 4000); pesanOk.value = p; }
const pesanOk = ref("");
</script>

<template>
  <PageHeader judul="Recovery Ledger" :demo="modeDemo" sub="Kelola kandidat recovery empat mata: kandidat → diperiksa → dikoreksi → diajukan → hasil." />
  <div v-if="galat" class="bahaya-kecil" style="margin-bottom: 12px">{{ galat }}</div>
  <div v-if="pesanOk" class="info-kecil" style="margin-bottom: 12px">{{ pesanOk }}</div>

  <!-- Kalibrasi -->
  <div class="grid" style="grid-template-columns: repeat(4, 1fr); margin-bottom: 14px" v-if="kalibrasi">
    <div class="kartu kpi-kartu"><span class="label">TOTAL ESTIMASI (HASIL)</span><span class="nilai">{{ rupiahSingkat(kalibrasi.totalEstimasi) }}</span></div>
    <div class="kartu kpi-kartu"><span class="label">TOTAL REALISASI</span><span class="nilai" style="color: var(--hijau)">{{ rupiahSingkat(kalibrasi.totalRealisasi) }}</span></div>
    <div class="kartu kpi-kartu"><span class="label">REALIZATION RATE</span><span class="nilai">{{ persen(kalibrasi.realizationRate) }}</span><span class="catatan">hit rate {{ persen(kalibrasi.hitRate) }} · n={{ kalibrasi.nSampel }}</span></div>
    <div class="kartu kpi-kartu">
      <span class="label">KALIBRASI</span>
      <span class="nilai" style="font-size: 16px; margin-top: 6px">{{ kalibrasi.kalibrasiTersedia ? "Tersedia" : "Menunggu sampel" }}</span>
      <span class="catatan">ambang 30 sampel · negatif: {{ kalibrasi.realisasiNegatif }}</span>
    </div>
  </div>

  <div class="filter-bar">
    <div class="kolom-form cari">
      <label>Cari SEP / MR / nama</label>
      <input v-model="cari" placeholder="mis. 030390…" />
    </div>
    <div class="kolom-form" style="max-width: 200px">
      <label>Status</label>
      <select v-model="status">
        <option value="SEMUA">Semua status</option>
        <option v-for="(l, k) in LABEL_STATUS_LEDGER" :key="k" :value="k">{{ l }}</option>
      </select>
    </div>
    <div style="flex: 1" />
    <button v-if="bisaAkses.ledgerTulis" class="btn navy" @click="bukaBuat = true">+ Kandidat Manual</button>
  </div>

  <div v-if="memuat" class="pemuat"><div class="spinner" /><span>Memuat ledger…</span></div>
  <div v-else class="kartu" style="padding: 0; overflow: hidden">
    <div class="tabel-wadah gulir-y" style="border: none; max-height: 560px">
      <table class="tabel">
        <thead>
          <tr><th>SEP</th><th>Pasien</th><th>INA-CBG</th><th>Kelas Rugi</th><th class="kanan">Estimasi</th><th class="kanan">Realisasi</th><th>Status</th><th>Strategi</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="l in items" :key="l.id">
            <td class="tunggal">{{ l.sep }}</td>
            <td><b>{{ l.namaPasien }}</b><br /><span style="font-size: 11.5px; color: var(--text-3)">{{ l.nomorMr }}</span></td>
            <td class="tunggal">{{ l.kodeIcbg }}</td>
            <td>
              <span class="badge" :class="l.kelasRugi === 'S1_CODING' ? 'kuning' : l.kelasRugi === 'S2_MIXED' ? 'ungu' : 'merah'">
                {{ LABEL_KELAS_RUGI[l.kelasRugi] ?? l.kelasRugi }}
              </span>
            </td>
            <td class="kanan"><b>{{ rupiah(l.estimasi) }}</b></td>
            <td class="kanan" :style="l.realisasi != null ? (l.realisasi >= 0 ? 'color: var(--hijau)' : 'color: var(--merah)') : ''">
              {{ l.realisasi != null ? rupiah(l.realisasi) : "—" }}
            </td>
            <td><span class="badge" :class="badgeStatus[l.status]">{{ LABEL_STATUS_LEDGER[l.status] ?? l.status }}</span></td>
            <td style="font-size: 11.5px; color: var(--text-2); max-width: 220px">{{ l.strategy ?? "—" }}</td>
            <td><button class="btn garis kecil" @click="buka(l)">Detail</button></td>
          </tr>
          <tr v-if="items.length === 0"><td colspan="9" class="kosong">Belum ada entri ledger yang cocok.</td></tr>
        </tbody>
      </table>
    </div>
  </div>

  <!-- Modal detail + transisi -->
  <div v-if="terpilih || memuatDetail" class="modal-topeng" @click.self="terpilih = null; memuatDetail = false">
    <div class="modal">
      <div class="kepala">
        <h3 v-if="terpilih">Ledger {{ terpilih.sep }} — {{ terpilih.namaPasien }}</h3>
        <button class="btn garis kecil" @click="terpilih = null; memuatDetail = false">✕</button>
      </div>
      <div class="badan" v-if="terpilih">
        <div class="grid tiga" style="margin-bottom: 16px">
          <div class="kartu kpi-kartu" style="box-shadow: none; background: var(--bg)">
            <span class="label">ESTIMASI</span><span class="nilai" style="font-size: 19px">{{ rupiah(terpilih.estimasi) }}</span>
          </div>
          <div class="kartu kpi-kartu" style="box-shadow: none; background: var(--bg)">
            <span class="label">REALISASI</span>
            <span class="nilai" style="font-size: 19px" :style="terpilih.realisasi != null && terpilih.realisasi < 0 ? 'color: var(--merah)' : 'color: var(--hijau)'">
              {{ terpilih.realisasi != null ? rupiah(terpilih.realisasi) : "—" }}
            </span>
          </div>
          <div class="kartu kpi-kartu" style="box-shadow: none; background: var(--bg)">
            <span class="label">STATUS</span>
            <span style="margin-top: 6px"><span class="badge" :class="badgeStatus[terpilih.status]">{{ LABEL_STATUS_LEDGER[terpilih.status] }}</span></span>
          </div>
        </div>

        <div class="grid dua">
          <div>
            <b style="font-size: 13px">Timeline peristiwa</b>
            <div class="timeline" style="margin-top: 12px">
              <div v-for="ev in terpilih.events ?? []" :key="ev.id" class="peristiwa">
                <b>{{ ev.dariStatus ? LABEL_STATUS_LEDGER[ev.dariStatus] + " → " + LABEL_STATUS_LEDGER[ev.keStatus ?? ""] : "Dibuat" }}</b>
                <div style="font-size: 12px; color: var(--text-2)">oleh {{ ev.aktorNama }} <span v-if="ev.catatan">— {{ ev.catatan }}</span></div>
                <div class="waktu">{{ tanggalWaktu(ev.waktu) }}</div>
              </div>
              <div v-if="(terpilih.events ?? []).length === 0" class="kosong">Belum ada peristiwa.</div>
            </div>
          </div>
          <div v-if="bisaAkses.ledgerTulis">
            <b style="font-size: 13px">Transisi status</b>
            <div class="kartu" style="box-shadow: none; background: var(--bg); margin-top: 10px">
              <div class="kolom-form" style="margin-bottom: 10px">
                <label>Transisi ke</label>
                <select v-model="form.ke">
                  <option value="" disabled>Pilih status tujuan…</option>
                  <option v-for="ke in transisiSah[terpilih.status] ?? []" :key="ke" :value="ke">{{ LABEL_STATUS_LEDGER[ke] }}</option>
                </select>
              </div>
              <div v-if="perluReferensi()" class="kolom-form" style="margin-bottom: 10px">
                <label>Referensi rekam medis *</label>
                <input v-model="form.referensiRm" placeholder="no. RM / resume / tautan" />
              </div>
              <div v-if="perluRealisasi()" class="kolom-form" style="margin-bottom: 10px">
                <label>Nilai realisasi (Rp) *</label>
                <input v-model="form.realisasi" type="number" placeholder="mis. 1200000" />
              </div>
              <div v-if="perluAlasan() || form.alasan" class="kolom-form" style="margin-bottom: 10px">
                <label>Alasan {{ perluAlasan() ? "(min. 10 karakter) *" : "(opsional)" }}</label>
                <textarea v-model="form.alasan" />
              </div>
              <div v-if="galatTransisi" class="bahaya-kecil" style="margin-bottom: 10px">{{ galatTransisi }}</div>
              <button class="btn utama" style="width: 100%" :disabled="!form.ke" @click="transisi">Jalankan Transisi</button>
              <p style="font-size: 11.5px; color: var(--text-3); margin: 9px 0 0">
                Prinsip empat mata: konfirmasi HASIL & penetapan TIDAK BERUBAH hanya oleh Casemix/Super Admin. Sistem tidak pernah men-set DIAJUKAN/HASIL otomatis.
              </p>
            </div>
            <div v-if="terpilih.strategy" class="info-kecil" style="margin-top: 10px">Strategi: {{ terpilih.strategy }}</div>
          </div>
        </div>
      </div>
      <div class="kaki"><button class="btn garis" @click="terpilih = null; memuatDetail = false">Tutup</button></div>
    </div>
  </div>

  <!-- Modal buat kandidat -->
  <div v-if="bukaBuat" class="modal-topeng" @click.self="bukaBuat = false">
    <div class="modal" style="max-width: 620px">
      <div class="kepala"><h3>Tambah Kandidat Recovery Manual</h3><button class="btn garis kecil" @click="bukaBuat = false">✕</button></div>
      <div class="badan">
        <div class="kolom-form">
          <label>Nomor SEP episode (aktif, rugi)</label>
          <input v-model="sepBuat" @keydown.enter="cariSep" placeholder="mis. 030390000000013" />
        </div>
        <button class="btn garis" style="margin-top: 10px" @click="cariSep">Cari</button>
        <div v-if="cariEpisode.length" style="margin-top: 12px">
          <div v-for="ep in cariEpisode" :key="ep.id" class="aturan-dq terpicu" style="align-items: center">
            <div style="flex: 1">
              <b class="tunggal">{{ ep.noSep }}</b> — <b>{{ ep.namaPasien }}</b>
              <div class="pesan">{{ ep.kodeIcbg }} · {{ LABEL_KELAS_RUGI[ep.kelasRugi] }} · selisih {{ rupiah(ep.selisih) }}</div>
            </div>
            <button class="btn utama kecil" @click="buatKandidat(ep)">Tambah</button>
          </div>
        </div>
      </div>
      <div class="kaki"><button class="btn garis" @click="bukaBuat = false">Tutup</button></div>
    </div>
  </div>
</template>
