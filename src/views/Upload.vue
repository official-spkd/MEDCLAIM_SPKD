<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { api, ApiError, modeDemo } from "../api";
import { store } from "../store";
import { tanggalWaktu, angka } from "../format";
import type { Upload, DQReport } from "../types";
import PageHeader from "../components/PageHeader.vue";

const langkah = ref(1);
const hospitalId = ref(store.hospitals[0]?.id ?? "rs1");
const periode = ref(new Date().toISOString().slice(0, 7));
const namaFile = ref("");
const isi = ref("");
const override = ref(false);
const alasanOverride = ref("");
const memproses = ref(false);
const galat = ref("");
const okPesan = ref("");

interface Pratinjau {
  header: string[];
  delimiter: string;
  format: string;
  totalBaris: number;
  barisValid: number;
  barisGagal: number;
  gagalDetail: { nomorBaris: number; alasan: string }[];
  dq: DQReport;
  duplikatSep: string[];
  kolomTersedia: string[];
}
const pratinjau = ref<Pratinjau | null>(null);
const hasil = ref<{ barisBaru: number; barisBerubah: number; barisUnchanged: number; ledgerBaru: number; diff: Record<string, { label: string; lama: string; baru: string }[]> } | null>(null);

const uploads = ref<Upload[]>([]);
const detailUpload = ref<Upload | null>(null);

async function muatUploads() {
  const d = await api.get<{ items: Upload[] }>("/api/uploads");
  uploads.value = d.items;
}
onMounted(muatUploads);

const sampel = ref<{ nama: string; ukuran: number; deskripsi: string }[]>([]);
onMounted(async () => {
  try {
    const d = await api.get<{ items: typeof sampel.value }>("/api/samples");
    sampel.value = d.items;
  } catch {
    /* abaikan */
  }
});

async function pakaiSampel(nama: string) {
  galat.value = "";
  try {
    const d = await api.get<{ nama: string; isi: string }>(`/api/samples/${encodeURIComponent(nama)}`);
    namaFile.value = d.nama;
    isi.value = d.isi;
    langkah.value = 1;
    pratinjau.value = null;
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal mengambil sampel.";
  }
}

function dariFile(e: Event) {
  const f = (e.target as HTMLInputElement).files?.[0];
  if (!f) return;
  namaFile.value = f.name;
  const reader = new FileReader();
  reader.onload = () => {
    isi.value = String(reader.result ?? "");
    pratinjau.value = null;
  };
  reader.readAsText(f);
}

const gerbangBadge = computed(() => {
  const g = pratinjau.value?.dq.gerbang;
  if (g === "LOLOS") return "hijau";
  if (g === "OVERRIDE") return "kuning";
  return "merah";
});

async function pratinjaukan() {
  galat.value = "";
  okPesan.value = "";
  if (!isi.value.trim()) {
    galat.value = "Isi file masih kosong — pilih file atau contoh dahulu.";
    return;
  }
  memproses.value = true;
  try {
    pratinjau.value = await api.post<Pratinjau>("/api/uploads/preview", {
      hospitalId: hospitalId.value,
      periode: periode.value,
      namaFile: namaFile.value || "tanpa-nama.txt",
      isi: isi.value,
    });
    langkah.value = 2;
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal memvalidasi file.";
  } finally {
    memproses.value = false;
  }
}

async function commit() {
  galat.value = "";
  memproses.value = true;
  try {
    hasil.value = await api.post("/api/uploads/commit", {
      hospitalId: hospitalId.value,
      periode: periode.value,
      namaFile: namaFile.value || "tanpa-nama.txt",
      isi: isi.value,
      override: override.value,
      alasanOverride: alasanOverride.value,
    });
    langkah.value = 3;
    await muatUploads();
  } catch (e: any) {
    if (e instanceof ApiError && e.data?.gerbang) galat.value = e.message;
    else galat.value = e?.message ?? "Gagal meng-commit upload.";
  } finally {
    memproses.value = false;
  }
}

function ulang() {
  langkah.value = 1;
  pratinjau.value = null;
  hasil.value = null;
  isi.value = "";
  namaFile.value = "";
  override.value = false;
  alasanOverride.value = "";
  galat.value = "";
  okPesan.value = "";
}

async function arsip(u: Upload) {
  await api.post(`/api/uploads/${u.id}/arsip`);
  await muatUploads();
}

const warnaGrade: Record<string, string> = { A: "hijau", B: "teal", C: "kuning", D: "merah" };
const ikonSev: Record<string, string> = { KRITIS: "⛔", PERINGATAN: "⚠️", INFO: "ℹ️" };
const labelDimensi: Record<string, string> = {
  COMPLETENESS: "Kelengkapan", VALIDITY: "Keabsahan", CONSISTENCY: "Konsistensi",
  TIMELINESS: "Ketepatan Waktu", UNIQUENESS: "Keunikan", ACCURACY: "Akurasi", CONFORMITY: "Kesesuaian",
};
</script>

<template>
  <PageHeader judul="Upload Data Klaim" :demo="modeDemo" sub="Unggah ekspor V-Claim (TSV) — validasi DQ, gerbang ingest, lalu versioning otomatis." />
  <div class="bar-langkah">
    <div class="langkah" :class="{ aktif: langkah === 1, selesai: langkah > 1 }"><span>1</span> Pilih file</div>
    <div class="langkah" :class="{ aktif: langkah === 2, selesai: langkah > 2 }"><span>2</span> Laporan DQ & gerbang</div>
    <div class="langkah" :class="{ aktif: langkah === 3, selesai: langkah > 3 }"><span>3</span> Ingest & versioning</div>
  </div>

  <!-- LANGKAH 1 -->
  <div v-if="langkah === 1" class="kartu" style="max-width: 860px">
    <h3>Unggah ekspor V-Claim (TSV)</h3>
    <p class="sub">Parser berbasis nama kolom + alias — mendukung format Basic maupun Detail (41 kolom + komponen billing).</p>

    <div class="grid dua">
      <div>
        <div class="kolom-form" style="margin-bottom: 12px">
          <label>Rumah sakit</label>
          <select v-model="hospitalId">
            <option v-for="h in store.hospitals" :key="h.id" :value="h.id">{{ h.kode }} — {{ h.nama }}</option>
          </select>
        </div>
        <div class="kolom-form" style="margin-bottom: 12px">
          <label>Periode evaluasi</label>
          <input v-model="periode" type="month" />
        </div>
        <div class="kolom-form">
          <label>File ekspor (.txt / .tsv / .csv)</label>
          <input type="file" accept=".txt,.tsv,.csv" @change="dariFile" />
        </div>
      </div>
      <div>
        <b style="font-size: 12.5px; color: var(--text-2)">ATAU pakai contoh siap-uji:</b>
        <div style="display: flex; flex-direction: column; gap: 7px; margin-top: 8px">
          <button v-for="s in sampel" :key="s.nama" class="btn garis" style="justify-content: flex-start; text-align: left" @click="pakaiSampel(s.nama)">
            <span>
              <b>{{ s.nama }}</b><br />
              <span style="font-size: 11.5px; color: var(--text-3)">{{ s.deskripsi }}</span>
            </span>
          </button>
        </div>
      </div>
    </div>

    <div v-if="galat" class="bahaya-kecil" style="margin-top: 12px">{{ galat }}</div>
    <div style="display: flex; gap: 10px; margin-top: 16px">
      <button class="btn utama" :disabled="memproses" @click="pratinjaukan">
        {{ memproses ? "Memvalidasi…" : "Validasi & Lihat Laporan DQ" }}
      </button>
    </div>
  </div>

  <!-- LANGKAH 2 -->
  <template v-else-if="langkah === 2 && pratinjau">
    <div class="grid" style="grid-template-columns: 300px 1fr; margin-bottom: 14px" v-if="pratinjau">
      <div class="kartu" style="text-align: center">
        <div style="font-size: 12px; color: var(--text-2); font-weight: 700; letter-spacing: 1px">SKOR DATA QUALITY</div>
        <div style="font-family: var(--font-judul); font-size: 52px; font-weight: 800; color: var(--text)">{{ pratinjau.dq.skor }}</div>
        <span class="badge" :class="warnaGrade[pratinjau.dq.grade]" style="font-size: 14px; padding: 4px 14px">Grade {{ pratinjau.dq.grade }}</span>
        <div style="margin-top: 12px">
          <span class="badge" :class="gerbangBadge" style="font-size: 12.5px; padding: 5px 14px">
            GERBANG: {{ pratinjau.dq.gerbang }}
          </span>
        </div>
        <div style="margin-top: 14px; font-size: 12.5px; color: var(--text-2); text-align: left">
          <div>Format terdeteksi: <b>{{ pratinjau.format }}</b> (delimiter "{{ pratinjau.delimiter === "\t" ? "TAB" : pratinjau.delimiter }}")</div>
          <div>Total baris: <b>{{ angka(pratinjau.totalBaris) }}</b></div>
          <div>Valid: <b style="color: var(--hijau)">{{ angka(pratinjau.barisValid) }}</b> · Gagal: <b style="color: var(--merah)">{{ angka(pratinjau.barisGagal) }}</b></div>
          <div v-if="pratinjau.duplikatSep.length">SEP ganda: <b style="color: var(--kuning)">{{ pratinjau.duplikatSep.length }}</b></div>
        </div>
      </div>

      <div class="kartu">
        <h3>Skor per dimensi</h3>
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 6px 22px">
          <div v-for="d in pratinjau.dq.dimensi" :key="d.kode" style="display: flex; align-items: center; gap: 10px">
            <span style="width: 130px; font-size: 12.5px; color: var(--text-2); font-weight: 600">{{ d.label }}</span>
            <div class="bar-rugi" style="flex: 1; height: 8px"><div :style="{ width: d.skor + '%', background: d.skor >= 90 ? 'var(--hijau)' : d.skor >= 60 ? 'var(--kuning)' : 'var(--merah)' }" /></div>
            <b style="width: 34px; text-align: right; font-size: 12.5px">{{ d.skor }}</b>
          </div>
        </div>
        <h3 style="margin-top: 16px">Matriks modul</h3>
        <div style="display: flex; flex-wrap: wrap; gap: 6px">
          <span v-for="m in pratinjau.dq.matriksModul" :key="m.modul" class="badge" :class="m.status === 'PENUH' ? 'hijau' : m.status === 'TERBATAS' ? 'kuning' : 'abu'">
            {{ m.modul }}: {{ m.status }}
          </span>
        </div>
      </div>
    </div>

    <div class="kartu" style="margin-bottom: 14px">
      <h3>20 aturan DQ ({{ pratinjau.dq.aturan.filter((a) => a.terpicu).length }} terpicu)</h3>
      <div style="max-height: 340px; overflow-y: auto">
        <div v-for="a in pratinjau.dq.aturan" :key="a.kode" class="aturan-dq" :class="{ terpicu: a.terpicu }">
          <span class="ikon">{{ a.terpicu ? ikonSev[a.severitas] : "✅" }}</span>
          <div style="flex: 1">
            <b>{{ a.nama }}</b>
            <span class="badge" :class="a.severitas === 'KRITIS' ? 'merah' : a.severitas === 'PERINGATAN' ? 'kuning' : 'abu'" style="margin-left: 6px">{{ a.severitas }}</span>
            <div class="pesan">{{ a.pesan }} <span class="kode">· {{ a.kode }} · {{ a.dimensi }}</span></div>
          </div>
          <b v-if="a.terpicu" style="color: var(--merah)">{{ a.jumlah }}</b>
        </div>
      </div>
    </div>

    <div v-if="pratinjau.dq.gerbang === 'BLOKIR'" class="kartu" style="border-color: var(--merah); margin-bottom: 14px">
      <h3 style="color: var(--merah)">Gerbang DQ: BLOKIR</h3>
      <p class="sub">Skor di bawah 60 atau terdapat aturan KRITIS terpicu. Data tidak dapat di-ingest tanpa override.</p>
      <label style="display: flex; align-items: center; gap: 10px; cursor: pointer; font-weight: 700">
        <input type="checkbox" v-model="override" style="width: 16px; min-height: 16px" />
        Override gerbang (hanya Casemix / Super Admin)
      </label>
      <div v-if="override" class="kolom-form" style="margin-top: 10px">
        <label>Alasan override (minimal 10 karakter)</label>
        <textarea v-model="alasanOverride" placeholder="mis. Format Basic memang tanpa komponen billing — disetujui untuk ingest sebagai versi awal…" />
      </div>
    </div>

    <div v-if="galat" class="bahaya-kecil" style="margin-bottom: 14px">{{ galat }}</div>

    <div style="display: flex; gap: 10px">
      <button class="btn garis" @click="langkah = 1">← Kembali</button>
      <button class="btn utama" :disabled="memproses || (pratinjau.dq.gerbang === 'BLOKIR' && (!override || alasanOverride.trim().length < 10))" @click="commit">
        {{ memproses ? "Meng-ingest…" : "Commit & Ingest Data" }}
      </button>
    </div>
  </template>

  <!-- LANGKAH 3 -->
  <template v-else-if="langkah === 3 && hasil">
    <div class="kartu" style="max-width: 900px; border-color: var(--hijau)">
      <h3 style="color: var(--hijau)">✅ Ingest selesai</h3>
      <div class="grid" style="grid-template-columns: repeat(4, 1fr); margin-top: 12px">
        <div class="kartu kpi-kartu" style="box-shadow: none; background: var(--bg)"><span class="label">BARIS BARU</span><span class="nilai" style="color: var(--teal)">{{ angka(hasil.barisBaru) }}</span></div>
        <div class="kartu kpi-kartu" style="box-shadow: none; background: var(--bg)"><span class="label">BERUBAH (VERSI BARU)</span><span class="nilai" style="color: var(--kuning)">{{ angka(hasil.barisBerubah) }}</span></div>
        <div class="kartu kpi-kartu" style="box-shadow: none; background: var(--bg)"><span class="label">TIDAK BERUBAH</span><span class="nilai" style="color: var(--text-2)">{{ angka(hasil.barisUnchanged) }}</span></div>
        <div class="kartu kpi-kartu" style="box-shadow: none; background: var(--bg)"><span class="label">KANDIDAT LEDGER BARU</span><span class="nilai" style="color: var(--ungu)">{{ angka(hasil.ledgerBaru) }}</span></div>
      </div>

      <template v-if="Object.keys(hasil.diff).length">
        <h3 style="margin-top: 18px">Perubahan terdeteksi ({{ Object.keys(hasil.diff).length }} SEP)</h3>
        <div style="max-height: 300px; overflow-y: auto">
          <div v-for="(diffs, sep) in hasil.diff" :key="sep" style="border: 1px solid var(--border); border-radius: 8px; padding: 10px 14px; margin-bottom: 8px">
            <b class="tunggal" style="font-size: 12px">{{ sep }}</b>
            <table class="tabel tabel-dimensi" style="margin-top: 6px">
              <thead><tr><th>Field</th><th>Lama</th><th>Baru</th></tr></thead>
              <tbody>
                <tr v-for="d in diffs" :key="d.label">
                  <td style="font-weight: 600">{{ d.label }}</td>
                  <td style="color: var(--merah)">{{ d.lama }}</td>
                  <td style="color: var(--hijau)">{{ d.baru }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>
    <div style="margin-top: 14px; display: flex; gap: 10px">
      <button class="btn utama" @click="ulang">Unggah file lain</button>
      <button class="btn garis" @click="langkah = 4; detailUpload = null">Lihat daftar upload</button>
    </div>
  </template>

  <!-- DAFTAR UPLOAD -->
  <div v-if="langkah === 4 || uploads.length" class="kartu" style="margin-top: 16px" :style="langkah === 4 ? '' : 'margin-top: 16px'">
    <h3>Riwayat upload ({{ uploads.length }})</h3>
    <div class="tabel-wadah gulir-y">
      <table class="tabel">
        <thead>
          <tr>
            <th>File</th><th>RS</th><th>Periode</th><th>Format</th>
            <th class="kanan">DQ</th><th class="kanan">Baris</th><th class="kanan">Baru</th>
            <th class="kanan">Ubah</th><th class="kanan">Sama</th><th>Status</th><th>Waktu</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in uploads" :key="u.id">
            <td style="font-weight: 600">{{ u.namaFile }}<div v-if="u.alasanOverride" style="font-size: 11px; color: var(--kuning); max-width: 240px">Override: {{ u.alasanOverride }}</div></td>
            <td class="tunggal">{{ u.hospitalId }}</td>
            <td>{{ u.periode }}</td>
            <td><span class="badge" :class="u.format === 'DETAIL' ? 'teal' : 'abu'">{{ u.format }}</span></td>
            <td class="kanan"><span class="badge" :class="warnaGrade[u.dqGrade]">{{ u.dqScore }} · {{ u.dqGrade }}</span></td>
            <td class="kanan">{{ angka(u.totalBaris) }}</td>
            <td class="kanan">{{ angka(u.barisBaru) }}</td>
            <td class="kanan">{{ angka(u.barisBerubah) }}</td>
            <td class="kanan">{{ angka(u.barisUnchanged) }}</td>
            <td><span class="badge" :class="u.status === 'SELESAI' ? 'hijau' : 'abu'">{{ u.status }}</span></td>
            <td style="font-size: 11.5px; color: var(--text-3)">{{ tanggalWaktu(u.createdAt) }}</td>
            <td><button v-if="u.status === 'SELESAI'" class="btn garis kecil" @click="arsip(u)">Arsipkan</button></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
