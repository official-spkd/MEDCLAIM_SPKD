<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api, modeDemo } from "../api";
import { store } from "../store";
import { tanggalWaktu, tanggal } from "../format";
import { LABEL_PERAN } from "../types";
import type { UserPub, Audit } from "../types";
import PageHeader from "../components/PageHeader.vue";

const tab = ref("pengguna");
const users = ref<UserPub[]>([]);
const audit = ref<Audit[]>([]);
const galat = ref("");

async function muat() {
  galat.value = "";
  try {
    if (tab.value === "pengguna") {
      users.value = (await api.get<{ items: UserPub[] }>("/api/admin/users")).items;
    } else {
      audit.value = (await api.get<{ items: Audit[] }>("/api/admin/audit")).items;
    }
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal memuat data admin.";
  }
}
onMounted(muat);
watchTab();
function watchTab() {
  // pemanggilan ulang ketika tab berubah
  const asli = tab;
  // Vue ref tidak punya watch inline di sini — pakai onMounted + handler tombol
  return asli;
}

async function toggle(u: UserPub) {
  galat.value = "";
  try {
    await api.post(`/api/admin/users/${u.id}/toggle`);
    await muat();
  } catch (e: any) {
    galat.value = e?.message ?? "Gagal mengubah status pengguna.";
  }
}

// Buat user
const bukaBuat = ref(false);
const form = ref({ email: "", nama: "", role: "CODER", hospitalId: "", spesialisasi: "", dpjpKode: "" });
const galatBuat = ref("");
async function buat() {
  galatBuat.value = "";
  try {
    await api.post("/api/admin/users", {
      email: form.value.email,
      nama: form.value.nama,
      role: form.value.role,
      hospitalId: form.value.hospitalId || null,
      spesialisasi: form.value.spesialisasi || null,
      dpjpKode: form.value.dpjpKode || null,
    });
    bukaBuat.value = false;
    form.value = { email: "", nama: "", role: "CODER", hospitalId: "", spesialisasi: "", dpjpKode: "" };
    await muat();
  } catch (e: any) {
    galatBuat.value = e?.message ?? "Gagal membuat pengguna.";
  }
}
</script>

<template>
  <PageHeader judul="Administrasi" :demo="modeDemo" sub="Pengguna, peran, jejak audit, dan pengaturan sistem MedClaim." />
  <div class="tabs">
    <button :class="{ aktif: tab === 'pengguna' }" @click="tab = 'pengguna'; muat()">Pengguna</button>
    <button :class="{ aktif: tab === 'audit' }" @click="tab = 'audit'; muat()">Jejak Audit</button>
    <button :class="{ aktif: tab === 'pengaturan' }" @click="tab = 'pengaturan'">Pengaturan Sistem</button>
  </div>

  <div v-if="galat" class="bahaya-kecil" style="margin-bottom: 12px">{{ galat }}</div>

  <template v-if="tab === 'pengguna'">
    <div style="display: flex; justify-content: flex-end; margin-bottom: 12px">
      <button class="btn navy" @click="bukaBuat = true">+ Pengguna Baru</button>
    </div>
    <div class="tabel-wadah">
      <table class="tabel">
        <thead>
          <tr><th>Nama</th><th>Email</th><th>Peran</th><th>RS</th><th>DPJP</th><th>Status</th><th>Login Terakhir</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td><b>{{ u.nama }}</b></td>
            <td class="tunggal">{{ u.email }}</td>
            <td><span class="badge navy">{{ LABEL_PERAN[u.role] ?? u.role }}</span></td>
            <td class="tunggal">{{ u.hospitalId ?? "—" }}</td>
            <td class="tunggal">{{ u.dpjpKode ?? "—" }}</td>
            <td><span class="badge" :class="u.aktif ? 'hijau' : 'abu'">{{ u.aktif ? "Aktif" : "Nonaktif" }}</span></td>
            <td style="font-size: 12px; color: var(--text-3)">{{ u.lastLoginAt ? tanggalWaktu(u.lastLoginAt) : "—" }}</td>
            <td>
              <button
                class="btn kecil"
                :class="u.aktif ? 'bahaya' : 'utama'"
                :disabled="u.email === store.sesi?.email"
                @click="toggle(u)"
              >
                {{ u.aktif ? "Nonaktifkan" : "Aktifkan" }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </template>

  <template v-else-if="tab === 'audit'">
    <div class="tabel-wadah gulir-y">
      <table class="tabel">
        <thead><tr><th>Waktu</th><th>Pengguna</th><th>Aksi</th><th>Entitas</th><th>Meta</th></tr></thead>
        <tbody>
          <tr v-for="a in audit" :key="a.id">
            <td style="font-size: 12px; white-space: nowrap">{{ tanggalWaktu(a.waktu) }}</td>
            <td class="tunggal">{{ a.userEmail }}</td>
            <td><span class="badge teal">{{ a.aksi }}</span></td>
            <td>{{ a.entitas }}<span v-if="a.entitasId" class="tunggal"> · {{ a.entitasId }}</span></td>
            <td style="font-size: 12px; color: var(--text-2)">{{ a.meta ?? "—" }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </template>

  <template v-else>
    <div class="grid dua" style="max-width: 860px">
      <div class="kartu">
        <h3>Pengaturan (settings)</h3>
        <div class="tabel-wadah">
          <table class="tabel">
            <thead><tr><th>Kunci</th><th>Nilai</th></tr></thead>
            <tbody>
              <tr v-for="s in store.settings" :key="s.key">
                <td class="tunggal">{{ s.key }}</td>
                <td style="font-size: 12.5px"><code>{{ s.value }}</code></td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div class="kartu">
        <h3>Informasi layanan</h3>
        <p style="font-size: 13px; color: var(--text-2)">
          Backend <b>Go</b> (net/http stdlib, port 3030 via gateway) · Frontend <b>Vue 3</b> (Vite, single-file build).<br />
          Penyimpanan: JSON file <code>data/medclaim.json</code> dengan mutex — tanpa dependensi eksternal.<br />
          Autentikasi: PBKDF2-HMAC-SHA256 (210.000 iterasi) + cookie token HMAC-SHA256 (7 hari).<br />
          Audit: seluruh aksi penting (login, upload, transisi ledger, query, admin) tercatat append-only.
        </p>
      </div>
    </div>
  </template>

  <!-- Modal buat user -->
  <div v-if="bukaBuat" class="modal-topeng" @click.self="bukaBuat = false">
    <div class="modal" style="max-width: 560px">
      <div class="kepala"><h3>Pengguna Baru</h3><button class="btn garis kecil" @click="bukaBuat = false">✕</button></div>
      <div class="badan">
        <div class="grid dua" style="gap: 10px">
          <div class="kolom-form"><label>Nama</label><input v-model="form.nama" /></div>
          <div class="kolom-form"><label>Email</label><input v-model="form.email" type="email" /></div>
          <div class="kolom-form">
            <label>Peran</label>
            <select v-model="form.role">
              <option v-for="(l, k) in LABEL_PERAN" :key="k" :value="k">{{ l }}</option>
            </select>
          </div>
          <div class="kolom-form">
            <label>Rumah sakit</label>
            <select v-model="form.hospitalId">
              <option value="">— Tanpa RS (Superadmin) —</option>
              <option v-for="h in store.hospitals" :key="h.id" :value="h.id">{{ h.nama }}</option>
            </select>
          </div>
          <div class="kolom-form" v-if="form.role === 'DPJP'"><label>Spesialisasi</label><input v-model="form.spesialisasi" /></div>
          <div class="kolom-form" v-if="form.role === 'DPJP'"><label>Kode DPJP</label><input v-model="form.dpjpKode" placeholder="DPJP-01xx" /></div>
        </div>
        <p style="font-size: 12px; color: var(--text-3)">Sandi awal: <b>MedClaim#demo2026</b></p>
        <div v-if="galatBuat" class="bahaya-kecil">{{ galatBuat }}</div>
      </div>
      <div class="kaki">
        <button class="btn garis" @click="bukaBuat = false">Batal</button>
        <button class="btn utama" :disabled="!form.email || !form.nama" @click="buat">Buat Pengguna</button>
      </div>
    </div>
  </div>
</template>
