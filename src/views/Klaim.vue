<script setup lang="ts">
import { ref, watch, onMounted, computed } from "vue";
import { useRoute } from "vue-router";
import { api, modeDemo } from "../api";
import { store, queryFilter, bisaAkses } from "../store";
import { rupiah, persen, tanggal, usia, angka } from "../format";
import { LABEL_KELAS_RUGI } from "../types";
import type { Episode, Readiness } from "../types";
import PageHeader from "../components/PageHeader.vue";

const route = useRoute();

const items = ref<Episode[]>([]);
const total = ref(0);
const cari = ref("");
const kelasRugi = ref("SEMUA");
const kelas = ref("");
const memuat = ref(true);
const galat = ref("");

const halaman = ref(1);
const perHalaman = 25;

async function muat() {
  memuat.value = true;
  galat.value = "";
  try {
    const t = queryFilter({
      kelasRugi: kelasRugi.value,
      kelas: kelas.value,
      q: cari.value,
    });
    const d = await api.get<{ total: number; items: Episode[] }>("/api/episodes" + t);
    items.value = d.items;
    total.value = d.total;
    halaman.value = 1;
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal memuat episode.";
  } finally {
    memuat.value = false;
  }
}
onMounted(() => {
  const q = route.query.q;
  if (typeof q === "string" && q.trim()) cari.value = q.trim();
  muat();
});
let timer: ReturnType<typeof setTimeout> | undefined;
watch([() => store.filter.rs, () => store.filter.tipe, () => store.filter.dari, () => store.filter.sampai, () => store.filter.upload], muat);
watch(cari, () => {
  clearTimeout(timer);
  timer = setTimeout(muat, 350);
});
watch([kelasRugi, kelas], muat);

const halamanTotal = computed(() => Math.max(1, Math.ceil(total.value / perHalaman)));
const terpotong = computed(() => items.value.slice((halaman.value - 1) * perHalaman, halaman.value * perHalaman));

const badgeKelas: Record<string, string> = {
  SURPLUS: "hijau", S1_CODING: "kuning", S2_MIXED: "ungu", STRUCTURAL: "merah", RAHASIA: "abu",
};

// ---------- Detail ----------
const detail = ref<{ episode: Episode & { readiness?: Readiness }; versi: Episode[] } | null>(null);
const tabDetail = ref("ringkasan");
const muatDetail = ref(false);
async function bukaDetail(ep: Episode) {
  muatDetail.value = true;
  detail.value = null;
  tabDetail.value = "ringkasan";
  try {
    detail.value = await api.get(`/api/episodes/${encodeURIComponent(ep.id)}`);
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal memuat detail.";
  } finally {
    muatDetail.value = false;
  }
}
function warnaBar(ep: Episode): string {
  if (ep.kelasRugi === "SURPLUS" || ep.kelasRugi === "RAHASIA") return "var(--hijau)";
  if (ep.kelasRugi === "S1_CODING") return "var(--kuning)";
  if (ep.kelasRugi === "S2_MIXED") return "var(--ungu)";
  return "var(--merah)";
}
</script>

<template>
  <PageHeader
    judul="Claim Explorer"
    :demo="modeDemo"
    sub="Jelajahi seluruh episode klaim INA-CBG — filter SEP, pasien, kelas kerugian, hingga batch upload."
  />
  <div class="filter-bar">
    <div class="kolom-form cari">
      <label>Cari (SEP / MR / nama / ICD / CBG)</label>
      <input v-model="cari" placeholder="mis. 030390…, K29.7, Andi" />
    </div>
    <div class="kolom-form" style="max-width: 190px">
      <label>Kelas kerugian</label>
      <select v-model="kelasRugi">
        <option value="SEMUA">Semua</option>
        <option value="SURPLUS">Surplus</option>
        <option value="S1_CODING">S1 — Coding Loss</option>
        <option value="S2_MIXED">S2 — Mixed Loss</option>
        <option value="STRUCTURAL">Structural Loss</option>
      </select>
    </div>
    <div class="kolom-form" style="max-width: 140px">
      <label>Kelas rawat</label>
      <select v-model="kelas">
        <option value="">Semua</option>
        <option value="1">Kelas 1</option>
        <option value="2">Kelas 2</option>
        <option value="3">Kelas 3</option>
      </select>
    </div>
    <div class="kolom-form" style="max-width: 160px">
      <label>Upload batch</label>
      <select v-model="store.filter.upload">
        <option value="">Semua</option>
        <option v-for="u in store.uploads" :key="u.id" :value="u.id">{{ u.namaFile.slice(0, 26) }}</option>
      </select>
    </div>
    <div class="kolom-form" style="max-width: 150px">
      <label>Dari (tgl masuk)</label>
      <input v-model="store.filter.dari" type="date" />
    </div>
    <div class="kolom-form" style="max-width: 150px">
      <label>Sampai</label>
      <input v-model="store.filter.sampai" type="date" />
    </div>
  </div>

  <div v-if="memuat" class="pemuat"><div class="spinner" /><span>Memuat episode…</span></div>
  <div v-else-if="galat" class="bahaya-kecil" style="margin-bottom: 12px">{{ galat }}</div>

  <div class="kartu" style="padding: 0; overflow: hidden" v-else>
    <div style="display: flex; justify-content: space-between; align-items: center; padding: 13px 16px; flex-wrap: wrap; gap: 8px">
      <h3>{{ angka(total) }} episode</h3>
      <div style="display: flex; gap: 6px; align-items: center">
        <button class="btn garis kecil" :disabled="halaman <= 1" @click="halaman--">‹</button>
        <span style="font-size: 12.5px; color: var(--text-2)">Hal. {{ halaman }}/{{ halamanTotal }}</span>
        <button class="btn garis kecil" :disabled="halaman >= halamanTotal" @click="halaman++">›</button>
      </div>
    </div>
    <div class="tabel-wadah gulir-y" style="border: none; border-top: 1px solid var(--border); border-radius: 0">
      <table class="tabel">
        <thead>
          <tr>
            <th>SEP</th><th>Pasien</th><th>Masuk–Keluar</th><th class="kanan">LOS</th>
            <th>INA-CBG</th><th>Diagnosa</th>
            <th v-if="bisaAkses.finansial" class="kanan">Tarif</th>
            <th v-if="bisaAkses.finansial" class="kanan">Billing</th>
            <th v-if="bisaAkses.finansial" class="kanan">Rugi</th>
            <th v-if="bisaAkses.finansial">Kelas</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="ep in terpotong" :key="ep.id">
            <td class="tunggal">{{ ep.noSep }}<br /><span style="color: var(--text-3)">v{{ ep.versionNo }}</span></td>
            <td>
              <b>{{ ep.namaPasien }}</b><br />
              <span style="font-size: 11.5px; color: var(--text-3)">{{ ep.noMr }} · {{ usia(ep.tglLahir, ep.tglMasuk) }} th · {{ ep.jenisKelamin }}</span>
            </td>
            <td style="font-size: 12.5px">{{ tanggal(ep.tglMasuk) }}<br /><span style="color: var(--text-3)">{{ tanggal(ep.tglKeluar) }}</span></td>
            <td class="kanan"><b>{{ ep.los }}</b></td>
            <td>
              <span class="tunggal">{{ ep.kodeIcbg }}</span><br />
              <span style="font-size: 11.5px; color: var(--text-3)">{{ ep.deskripsiIcbg }}</span>
            </td>
            <td>
              <span v-for="(d, i) in ep.diagnosa.slice(0, 2)" :key="i" class="badge navy" style="margin-right: 4px; margin-bottom: 2px">{{ d.kode }}</span>
              <span v-if="ep.diagnosa.length > 2" class="badge abu">+{{ ep.diagnosa.length - 2 }}</span>
            </td>
            <template v-if="bisaAkses.finansial">
              <td class="kanan">{{ rupiah(ep.tarifInacbg) }}</td>
              <td class="kanan">{{ rupiah(ep.totalBilling) }}</td>
              <td class="kanan" style="min-width: 110px">
                <div style="display: flex; align-items: center; gap: 7px; justify-content: flex-end">
                  <div class="bar-rugi"><div :style="{ width: Math.round(ep.pctRugi * 100) + '%', background: warnaBar(ep) }" /></div>
                  <b style="font-size: 12px">{{ persen(ep.pctRugi, 0) }}</b>
                </div>
              </td>
              <td><span class="badge" :class="badgeKelas[ep.kelasRugi]">{{ LABEL_KELAS_RUGI[ep.kelasRugi] }}</span></td>
            </template>
            <td><button class="btn garis kecil" @click="bukaDetail(ep)">Detail</button></td>
          </tr>
          <tr v-if="terpotong.length === 0">
            <td colspan="12" class="kosong">Tidak ada episode yang cocok dengan filter.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <!-- Modal detail -->
  <div v-if="detail || muatDetail" class="modal-topeng" @click.self="detail = null; muatDetail = false">
    <div class="modal lebar">
      <div class="kepala">
        <h3 v-if="detail">Episode {{ detail.episode.noSep }} — {{ detail.episode.namaPasien }}</h3>
        <h3 v-else>Memuat…</h3>
        <button class="btn garis kecil" @click="detail = null; muatDetail = false">✕</button>
      </div>
      <div class="badan" v-if="detail">
        <div class="tabs">
          <button :class="{ aktif: tabDetail === 'ringkasan' }" @click="tabDetail = 'ringkasan'">Ringkasan</button>
          <button v-if="bisaAkses.finansial" :class="{ aktif: tabDetail === 'finansial' }" @click="tabDetail = 'finansial'">Finansial</button>
          <button :class="{ aktif: tabDetail === 'readiness' }" @click="tabDetail = 'readiness'">Readiness & Anomali</button>
          <button :class="{ aktif: tabDetail === 'versi' }" @click="tabDetail = 'versi'">Riwayat Versi</button>
        </div>

        <template v-if="tabDetail === 'ringkasan'">
          <div class="grid dua">
            <div>
              <table class="tabel tabel-dimensi" style="border: 1px solid var(--border); border-radius: 8px">
                <tbody>
                  <tr><td style="color: var(--text-2)">No. Kartu</td><td class="tunggal">{{ detail.episode.noKartu }}</td></tr>
                  <tr><td style="color: var(--text-2)">No. MR</td><td class="tunggal">{{ detail.episode.noMr }}</td></tr>
                  <tr><td style="color: var(--text-2)">Tgl Lahir</td><td>{{ tanggal(detail.episode.tglLahir) }} ({{ usia(detail.episode.tglLahir, detail.episode.tglMasuk) }} tahun)</td></tr>
                  <tr><td style="color: var(--text-2)">Masuk → Keluar</td><td>{{ tanggal(detail.episode.tglMasuk) }} → {{ tanggal(detail.episode.tglKeluar) }} ({{ detail.episode.los }} hari)</td></tr>
                  <tr><td style="color: var(--text-2)">Kelas Rawat</td><td>Kelas {{ detail.episode.kelasRawat }}</td></tr>
                  <tr><td style="color: var(--text-2)">Tipe Layanan</td><td>{{ detail.episode.tipeLayanan }}</td></tr>
                  <tr><td style="color: var(--text-2)">Cara Pulang</td><td>{{ detail.episode.caraPulang ?? "—" }}</td></tr>
                  <tr><td style="color: var(--text-2)">DPJP</td><td>{{ detail.episode.dpjpNama ?? detail.episode.dpjpKode ?? "—" }}</td></tr>
                </tbody>
              </table>
            </div>
            <div>
              <div class="kartu" style="box-shadow: none; background: var(--bg)">
                <h3>{{ detail.episode.kodeIcbg }} — {{ detail.episode.deskripsiIcbg }}</h3>
                <p style="font-size: 12.5px; color: var(--text-2); margin: 0 0 8px">
                  Severity level {{ detail.episode.severityLevel }}
                  <template v-if="detail.episode.idrgKode"> · In-DRG {{ detail.episode.idrgKode }}</template>
                </p>
                <div style="margin-top: 10px">
                  <b style="font-size: 12.5px">Diagnosa (ICD-10)</b>
                  <div style="margin-top: 5px; display: flex; flex-wrap: wrap; gap: 5px">
                    <span v-for="d in detail.episode.diagnosa" :key="d.kode" class="badge navy" :title="d.nama">{{ d.kode }} — {{ d.nama }}</span>
                  </div>
                </div>
                <div v-if="detail.episode.prosedur.length" style="margin-top: 10px">
                  <b style="font-size: 12.5px">Prosedur (ICD-9-CM)</b>
                  <div style="margin-top: 5px; display: flex; flex-wrap: wrap; gap: 5px">
                    <span v-for="p in detail.episode.prosedur" :key="p.kode" class="badge teal" :title="p.nama">{{ p.kode }} — {{ p.nama }}</span>
                  </div>
                </div>
                <div v-if="detail.episode.addOn?.length" style="margin-top: 10px">
                  <b style="font-size: 12.5px">Add-on tarif</b>
                  <div style="margin-top: 5px">
                    <span v-for="(a, i) in detail.episode.addOn" :key="i" class="badge" :class="a.status === 'DIKLAIM' ? 'hijau' : 'merah'" style="margin-right: 4px">
                      {{ a.kode }} · {{ a.status }} · {{ rupiah(a.tarif) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>

        <template v-else-if="tabDetail === 'finansial'">
          <div class="grid tiga" style="margin-bottom: 14px">
            <div class="kartu kpi-kartu" style="box-shadow: none"><span class="label">TARIF INA-CBG</span><span class="nilai">{{ rupiah(detail.episode.tarifInacbg) }}</span></div>
            <div class="kartu kpi-kartu" style="box-shadow: none"><span class="label">TOTAL BILLING</span><span class="nilai">{{ rupiah(detail.episode.totalBilling) }}</span></div>
            <div class="kartu kpi-kartu" style="box-shadow: none">
              <span class="label">SELISIH</span>
              <span class="nilai" :style="{ color: detail.episode.selisih >= 0 ? 'var(--hijau)' : 'var(--merah)' }">{{ rupiah(detail.episode.selisih) }}</span>
              <span class="catatan">{{ LABEL_KELAS_RUGI[detail.episode.kelasRugi] }} · rugi {{ persen(detail.episode.pctRugi) }}</span>
            </div>
          </div>
          <div class="tabel-wadah">
            <table class="tabel">
              <thead><tr><th>Komponen</th><th class="kanan">Nilai</th><th class="kanan">Porsi billing</th></tr></thead>
              <tbody>
                <tr v-for="(v, k) in detail.episode.komponen" :key="k" v-show="v > 0">
                  <td>{{ k }}</td>
                  <td class="kanan">{{ rupiah(v) }}</td>
                  <td class="kanan">{{ detail.episode.totalBilling ? ((v / detail.episode.totalBilling) * 100).toFixed(1) : 0 }}%</td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>

        <template v-else-if="tabDetail === 'readiness'">
          <div class="grid dua">
            <div class="kartu" style="box-shadow: none">
              <h3>Readiness Score</h3>
              <div v-if="detail.episode.readiness" style="font-family: var(--font-judul); font-size: 34px; font-weight: 700; color: var(--teal)">
                {{ detail.episode.readiness.total }}<span style="font-size: 14px; color: var(--text-3)">/100</span>
              </div>
              <ul v-if="detail.episode.readiness" style="margin: 8px 0 0; padding-left: 18px; font-size: 12.5px; color: var(--text-2)">
                <li v-for="(f, i) in detail.episode.readiness.flag" :key="i">{{ f }}</li>
                <li v-if="detail.episode.readiness.flag.length === 0" style="color: var(--hijau)">Seluruh komponen siap — tidak ada flag.</li>
              </ul>
            </div>
            <div class="kartu" style="box-shadow: none">
              <h3>Anomali Episode</h3>
              <div v-if="detail.episode.anomali?.length">
                <div v-for="(a, i) in detail.episode.anomali" :key="i" class="peringatan-kecil" style="margin-bottom: 7px">{{ a }}</div>
              </div>
              <div v-else class="info-kecil">Tidak ada anomali pada episode ini.</div>
            </div>
          </div>
        </template>

        <template v-else>
          <div class="tabel-wadah">
            <table class="tabel">
              <thead>
                <tr><th>Versi</th><th>Tarif</th><th>Sev.</th><th>LOS</th><th>Diagnosa</th><th>Status</th></tr>
              </thead>
              <tbody>
                <tr v-for="v in detail.versi" :key="v.id" :style="v.id === detail.episode.id ? 'background: var(--teal-100)' : ''">
                  <td><b>v{{ v.versionNo }}</b></td>
                  <td>{{ rupiah(v.tarifInacbg) }}</td>
                  <td>{{ v.severityLevel }}</td>
                  <td>{{ v.los }}</td>
                  <td class="tunggal">{{ v.diagnosa.map((d) => d.kode).join(", ") }}</td>
                  <td>
                    <span class="badge" :class="v.superseded ? 'abu' : 'hijau'">{{ v.superseded ? "Digantikan" : "Aktif" }}</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </div>
      <div class="kaki">
        <button class="btn garis" @click="detail = null; muatDetail = false">Tutup</button>
      </div>
    </div>
  </div>
</template>
