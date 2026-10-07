<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { api, ApiError, modeDemo } from "../api";
import { store, bisaAkses } from "../store";
import { tanggal, tanggalWaktu, angka } from "../format";
import { LABEL_STATUS_QUERY } from "../types";
import type { QueryDok } from "../types";
import PageHeader from "../components/PageHeader.vue";

const items = ref<QueryDok[]>([]);
const status = ref("SEMUA");
const memuat = ref(true);
const galat = ref("");

const badgeStatus: Record<string, string> = {
  DRAFT: "abu", TERKIRIM: "teal", DIJAWAB: "hijau", TERLAMBAT: "merah",
};

async function muat() {
  memuat.value = true;
  galat.value = "";
  try {
    const s = status.value !== "SEMUA" ? `?status=${status.value}` : "";
    items.value = (await api.get<{ items: QueryDok[] }>("/api/queries" + s)).items;
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal memuat query.";
  } finally {
    memuat.value = false;
  }
}
onMounted(muat);
watch(status, muat);

// ---------- Buat query ----------
const bukaBuat = ref(false);
const form = ref({ sep: "", dpjpId: "", templateId: "", pertanyaan: "", konteks: "" });
const galatBuat = ref("");
const memproses = ref(false);

function pakaiTemplate() {
  const t = store.templates.find((t) => t.id === form.value.templateId);
  if (t) form.value.pertanyaan = t.isi;
}

async function kirim() {
  galatBuat.value = "";
  memproses.value = true;
  try {
    await api.post("/api/queries", form.value);
    bukaBuat.value = false;
    form.value = { sep: "", dpjpId: "", templateId: "", pertanyaan: "", konteks: "" };
    await muat();
  } catch (e: any) {
    galatBuat.value = e?.message ?? "Gagal mengirim query.";
  } finally {
    memproses.value = false;
  }
}

// ---------- Jawab ----------
const terpilih = ref<QueryDok | null>(null);
const jawaban = ref("");
const galatJawab = ref("");
async function buka(q: QueryDok) {
  terpilih.value = q;
  jawaban.value = "";
  galatJawab.value = "";
}
async function kirimJawaban() {
  if (!terpilih.value) return;
  galatJawab.value = "";
  try {
    terpilih.value = await api.post<QueryDok>(`/api/queries/${terpilih.value.id}/answer`, { jawaban: jawaban.value });
    await muat();
  } catch (e: any) {
    galatJawab.value = e?.message ?? "Gagal menyimpan jawaban.";
  }
}
</script>

<template>
  <PageHeader judul="Query & Klarifikasi" :demo="modeDemo" sub="Klarifikasi coding ke DPJP dengan template dan SLA — jawaban terkunci setelah dikirim." />
  <div v-if="galat" class="bahaya-kecil" style="margin-bottom: 12px">{{ galat }}</div>

  <div class="filter-bar">
    <div class="kolom-form" style="max-width: 200px">
      <label>Status</label>
      <select v-model="status">
        <option value="SEMUA">Semua status</option>
        <option v-for="(l, k) in LABEL_STATUS_QUERY" :key="k" :value="k">{{ l }}</option>
      </select>
    </div>
    <div style="flex: 1" />
    <button v-if="bisaAkses.query" class="btn navy" @click="bukaBuat = true">+ Query Baru</button>
  </div>

  <div v-if="store.sesi?.role === 'DPJP'" class="info-kecil" style="margin-bottom: 12px">
    Inbox DPJP — query klarifikasi yang ditujukan kepada Anda. Jawaban akan terkunci setelah dikirim.
  </div>

  <div v-if="memuat" class="pemuat"><div class="spinner" /><span>Memuat query…</span></div>
  <div v-else class="tabel-wadah">
    <table class="tabel">
      <thead>
        <tr><th>SEP</th><th>Pasien</th><th>DPJP</th><th>Template / Pertanyaan</th><th>Status</th><th>Deadline</th><th></th></tr>
      </thead>
      <tbody>
        <tr v-for="q in items" :key="q.id">
          <td class="tunggal">{{ q.sep }}<br /><span style="color: var(--text-3)">{{ q.inisial }}</span></td>
          <td><b>MR {{ q.nomorMr }}</b><br /><span style="font-size: 11.5px; color: var(--text-3)">{{ tanggal(q.tglMasuk) }} → {{ tanggal(q.tglKeluar) }}</span></td>
          <td style="max-width: 180px">{{ q.dpjpNama }}<br /><span style="font-size: 11.5px; color: var(--text-3)">{{ q.dpjpSpesialisasi }}</span></td>
          <td style="max-width: 300px">
            <span v-if="q.templateNama" class="badge navy" style="margin-bottom: 3px">{{ q.templateNama }}</span>
            <div style="font-size: 12px; color: var(--text-2); display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden">{{ q.pertanyaan }}</div>
          </td>
          <td><span class="badge" :class="badgeStatus[q.status]">{{ LABEL_STATUS_QUERY[q.status] ?? q.status }}</span></td>
          <td style="font-size: 12px">{{ tanggal(q.deadline) }}</td>
          <td><button class="btn garis kecil" @click="buka(q)">Buka</button></td>
        </tr>
        <tr v-if="items.length === 0"><td colspan="7" class="kosong">Tidak ada query.</td></tr>
      </tbody>
    </table>
  </div>

  <!-- Modal buat -->
  <div v-if="bukaBuat" class="modal-topeng" @click.self="bukaBuat = false">
    <div class="modal" style="max-width: 680px">
      <div class="kepala"><h3>Query Klarifikasi Baru</h3><button class="btn garis kecil" @click="bukaBuat = false">✕</button></div>
      <div class="badan">
        <div class="kolom-form" style="margin-bottom: 12px">
          <label>Nomor SEP (episode aktif)</label>
          <input v-model="form.sep" placeholder="mis. 030390000000013" />
        </div>
        <div class="kolom-form" style="margin-bottom: 12px">
          <label>DPJP</label>
          <select v-model="form.dpjpId">
            <option value="" disabled>Pilih DPJP…</option>
            <option v-for="d in store.dokter" :key="d.id" :value="d.id">{{ d.nama }} ({{ d.dpjpKode }})</option>
          </select>
        </div>
        <div class="kolom-form" style="margin-bottom: 12px">
          <label>Template pertanyaan</label>
          <select v-model="form.templateId" @change="pakaiTemplate">
            <option value="">— Tanpa template —</option>
            <option v-for="t in store.templates" :key="t.id" :value="t.id">{{ t.nama }}</option>
          </select>
        </div>
        <div class="kolom-form" style="margin-bottom: 12px">
          <label>Pertanyaan</label>
          <textarea v-model="form.pertanyaan" placeholder="Tulis pertanyaan klarifikasi…" />
        </div>
        <div class="kolom-form">
          <label>Konteks (opsional)</label>
          <input v-model="form.konteks" placeholder="mis. Episode rugi S1, CC terlewat…" />
        </div>
        <div v-if="galatBuat" class="bahaya-kecil" style="margin-top: 12px">{{ galatBuat }}</div>
      </div>
      <div class="kaki">
        <button class="btn garis" @click="bukaBuat = false">Batal</button>
        <button class="btn utama" :disabled="memproses || !form.sep || !form.dpjpId || !form.pertanyaan" @click="kirim">
          {{ memproses ? "Mengirim…" : "Kirim ke DPJP" }}
        </button>
      </div>
    </div>
  </div>

  <!-- Modal detail / jawab -->
  <div v-if="terpilih" class="modal-topeng" @click.self="terpilih = null">
    <div class="modal" style="max-width: 640px">
      <div class="kepala">
        <h3>Query {{ terpilih.sep }} — {{ terpilih.inisial }}</h3>
        <button class="btn garis kecil" @click="terpilih = null">✕</button>
      </div>
      <div class="badan">
        <div style="display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 12px">
          <span class="badge" :class="badgeStatus[terpilih.status]">{{ LABEL_STATUS_QUERY[terpilih.status] }}</span>
          <span class="badge abu">SLA {{ terpilih.slaHari }} hari</span>
          <span class="badge navy">DPJP: {{ terpilih.dpjpNama }}</span>
          <span v-if="terpilih.templateNama" class="badge teal">{{ terpilih.templateNama }}</span>
        </div>
        <div class="kartu" style="box-shadow: none; background: var(--bg); margin-bottom: 12px">
          <b style="font-size: 12.5px">Pertanyaan</b>
          <p style="margin: 6px 0 0; font-size: 13.5px">{{ terpilih.pertanyaan }}</p>
          <p v-if="terpilih.konteks" style="margin: 8px 0 0; font-size: 12px; color: var(--text-2)">Konteks: {{ terpilih.konteks }}</p>
        </div>

        <template v-if="terpilih.jawaban">
          <div class="kartu" style="box-shadow: none; border-color: var(--hijau)">
            <b style="font-size: 12.5px; color: var(--hijau)">Jawaban DPJP (terkunci {{ tanggalWaktu(terpilih.answeredAt) }})</b>
            <p style="margin: 6px 0 0; font-size: 13.5px">{{ terpilih.jawaban }}</p>
          </div>
        </template>
        <template v-else-if="store.sesi && (store.sesi.id === terpilih.dpjpId || bisaAkses.query)">
          <div class="kolom-form">
            <label>Jawaban Anda</label>
            <textarea v-model="jawaban" placeholder="Jelaskan temuan klinis / dokumen pendukung…" />
          </div>
          <div v-if="galatJawab" class="bahaya-kecil" style="margin-top: 10px">{{ galatJawab }}</div>
          <button class="btn utama" style="margin-top: 12px; width: 100%" :disabled="jawaban.trim().length < 5" @click="kirimJawaban">
            Kirim Jawaban & Kunci
          </button>
        </template>
      </div>
      <div class="kaki"><button class="btn garis" @click="terpilih = null">Tutup</button></div>
    </div>
  </div>
</template>
