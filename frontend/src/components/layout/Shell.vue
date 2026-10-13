<script setup lang="ts">
// ============================================================
// MedClaim — Shell v3.1 (bahasa visual MedCredix, sidebar kuning-oranye)
// Sidebar gradien amber · header: cari Ctrl+K, filter global, tema,
// notifikasi, avatar
// ============================================================
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { store, bisaAkses, keluar } from "../store";
import { api } from "../api";
import { inisialNama, tanggalWaktu, angka } from "../format";
import { LABEL_PERAN } from "../types";
import Ikon from "./Ikon.vue";
import SearchModal from "./SearchModal.vue";

const route = useRoute();
const router = useRouter();

// ---------- Sidebar & tema ----------
const sidebarBuka = ref(false);
const gelap = ref(false);

function toggleTema() {
  gelap.value = !gelap.value;
  localStorage.setItem("tema", gelap.value ? "gelap" : "terang");
  document.documentElement.classList.toggle("dark", gelap.value);
}

// ---------- Navigasi ----------
const itemNav = computed(() => {
  const a = bisaAkses.value;
  const items: { ke: string; judul: string; ikon: string; tampil: boolean }[] = [
    { ke: "/", judul: "Dashboard", ikon: "dash", tampil: a.analitik || store.sesi?.role === "DPJP" },
    { ke: "/klaim", judul: "Claim Explorer", ikon: "klaim", tampil: a.analitik || store.sesi?.role === "DPJP" },
    { ke: "/pasien", judul: "Patient Tracer", ikon: "pasien", tampil: a.analitik || store.sesi?.role === "DPJP" },
    { ke: "/finansial", judul: "Financial", ikon: "uang", tampil: a.finansial },
    { ke: "/upload", judul: "Upload Klaim", ikon: "unggah", tampil: a.upload },
    { ke: "/kualitas", judul: "Claim Quality", ikon: "mutu", tampil: a.upload || store.sesi?.role === "DIREKTUR" },
    { ke: "/anomali", judul: "Anomaly Detection", ikon: "anomali", tampil: a.finansial },
    { ke: "/ledger", judul: "Recovery Ledger", ikon: "buku", tampil: a.ledger },
    { ke: "/query", judul: "Query", ikon: "tanya", tampil: a.query || store.sesi?.role === "DPJP" },
    { ke: "/admin", judul: "Administrasi", ikon: "atur", tampil: a.admin },
    { ke: "/dokumentasi", judul: "Dokumentasi", ikon: "buku2", tampil: true },
  ];
  return items.filter((i) => i.tampil);
});

const grupNav = computed(() => {
  const ambil = (daftar: string[]) => itemNav.value.filter((i) => daftar.includes(i.ke));
  return [
    { nama: "ANALITIK", items: ambil(["/", "/klaim", "/pasien", "/finansial"]) },
    { nama: "OPERASIONAL", items: ambil(["/upload", "/kualitas", "/anomali", "/ledger", "/query"]) },
    { nama: "LAINNYA", items: ambil(["/admin", "/dokumentasi"]) },
  ].filter((g) => g.items.length);
});

function pindah(ke: string) {
  sidebarBuka.value = false;
  router.push(ke);
}
function aktif(ke: string): boolean {
  return ke === "/" ? route.path === "/" : route.path.startsWith(ke);
}
async function tombolKeluar() {
  tutupSemua();
  await keluar();
  router.push("/login");
}

// ---------- Filter global ----------
const filterAktif = computed(
  () => store.filter.rs || store.filter.tipe !== "SEMUA" || store.filter.dari || store.filter.sampai,
);
function resetFilter() {
  store.filter = { rs: "", upload: "", tipe: "SEMUA", dari: "", sampai: "" };
}

// ---------- Pencarian global (Ctrl+K) ----------
const cariBuka = ref(false);
function bukaCari() {
  tutupSemua();
  cariBuka.value = true;
}
function tekanGlobal(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
    e.preventDefault();
    cariBuka.value = !cariBuka.value;
    tutupSemua();
  } else if (e.key === "Escape") {
    sidebarBuka.value = false;
    tutupSemua();
  }
}

// ---------- Badge hitungan sidebar + notifikasi ----------
const hitung = ref({ anomali: 0, queryTerkirim: 0, ledgerAktif: 0 });
const notifikasi = ref<{ ikon: string; warna: string; judul: string; desk: string; waktu?: string; ke?: string }[]>([]);
const notifBaca = ref(false);

async function muatRingkasan() {
  if (!store.sesi) return;
  const a = bisaAkses.value;
  const tugas: Promise<void>[] = [];

  if (a.finansial) {
    tugas.push(
      api
        .get<{ total: number; items: unknown[] }>("/api/anomali")
        .then((d) => {
          hitung.value.anomali = d.total ?? d.items.length;
          if ((d.total ?? d.items.length) > 0) {
            notifikasi.value.push({
              ikon: "peringatan",
              warna: "kuning",
              judul: `${angka(d.total ?? d.items.length)} anomali terdeteksi`,
              desk: "LOS outlier · duplikat SEP · readmisi · selisih ekstrem",
              ke: "/anomali",
            });
          }
        })
        .catch(() => {}),
    );
  }
  if (a.query || store.sesi.role === "DPJP") {
    tugas.push(
      api
        .get<{ items: { status: string }[] }>("/api/queries")
        .then((d) => {
          const belum = d.items.filter((q) => q.status === "TERKIRIM").length;
          hitung.value.queryTerkirim = belum;
          if (belum > 0) {
            notifikasi.value.push({
              ikon: "tanya",
              warna: "info",
              judul: `${belum} query menunggu jawaban DPJP`,
              desk: "Klarifikasi coding yang belum dijawab — periksa inbox.",
              ke: "/query",
            });
          }
        })
        .catch(() => {}),
    );
  }
  if (a.ledger) {
    tugas.push(
      api
        .get<{ items: { status: string }[] }>("/api/ledger")
        .then((d) => {
          const aktif = d.items.filter((l) => l.status !== "HASIL" && l.status !== "TIDAK_BERUBAH").length;
          hitung.value.ledgerAktif = aktif;
          if (aktif > 0) {
            notifikasi.value.push({
              ikon: "buku",
              warna: "ungu",
              judul: `${aktif} entri ledger aktif`,
              desk: "Kandidat recovery yang masih diproses empat mata.",
              ke: "/ledger",
            });
          }
        })
        .catch(() => {}),
    );
  }
  const unggahTerbaru = store.uploads.filter((u) => u.status !== "DIHAPUS").slice(0, 2);
  for (const u of unggahTerbaru) {
    notifikasi.value.push({
      ikon: "unggah",
      warna: "teal",
      judul: `Upload ${u.namaFile}`,
      desk: `${u.periode} · DQ ${u.dqScore} (${u.dqGrade}) · ${angka(u.totalBaris)} baris`,
      waktu: u.createdAt,
      ke: "/kualitas",
    });
  }
  await Promise.all(tugas);
}

watch(
  () => store.sesi?.id,
  (id) => {
    notifikasi.value = [];
    hitung.value = { anomali: 0, queryTerkirim: 0, ledgerAktif: 0 };
    notifBaca.value = false;
    if (id) muatRingkasan();
  },
);

// ---------- Popover ----------
const notifBuka = ref(false);
const akunBuka = ref(false);
function tutupSemua() {
  notifBuka.value = false;
  akunBuka.value = false;
}
function klikLuar(e: MouseEvent) {
  const el = e.target as HTMLElement;
  if (!el.closest(".popover-wadah")) tutupSemua();
}
function keNotif(ke?: string) {
  notifBuka.value = false;
  if (ke) router.push(ke);
}

const judulHalaman = computed(() => (route.meta.judul as string) ?? "MedClaim");

onMounted(() => {
  gelap.value = localStorage.getItem("tema") === "gelap";
  document.documentElement.classList.toggle("dark", gelap.value);
  window.addEventListener("keydown", tekanGlobal);
  document.addEventListener("click", klikLuar);
  if (store.sesi) muatRingkasan();
});
onUnmounted(() => {
  window.removeEventListener("keydown", tekanGlobal);
  document.removeEventListener("click", klikLuar);
});
</script>

<template>
  <div class="shell">
    <div
      v-if="sidebarBuka"
      style="position: fixed; inset: 0; background: rgba(12, 8, 4, 0.55); z-index: 55"
      @click="sidebarBuka = false"
    />

    <!-- ============ SIDEBAR ============ -->
    <aside class="sidebar" :class="{ buka: sidebarBuka }">
      <div class="brand">
        <div class="logo">
          <svg width="22" height="22" viewBox="0 0 32 32" fill="none">
            <path d="M8 20c3-8 6-12 8-12s5 4 8 12" stroke="#f6b04b" stroke-width="2.6" fill="none" stroke-linecap="round" />
            <circle cx="16" cy="20" r="3.2" fill="#f6b04b" />
          </svg>
        </div>
        <div>
          <b>MedClaim</b>
          <small>BY SPKD</small>
        </div>
      </div>

      <template v-for="grup in grupNav" :key="grup.nama">
        <div class="nav-grup">{{ grup.nama }}</div>
        <button
          v-for="it in grup.items"
          :key="it.ke"
          class="nav-item"
          :class="{ aktif: aktif(it.ke) }"
          @click="pindah(it.ke)"
          type="button"
        >
          <Ikon :nama="it.ikon" :ukuran="17" />
          <span class="judul">{{ it.judul }}</span>
          <span v-if="it.ke === '/anomali' && hitung.anomali" class="nav-badge">{{ hitung.anomali > 99 ? "99+" : hitung.anomali }}</span>
          <span v-else-if="it.ke === '/query' && hitung.queryTerkirim" class="nav-badge" style="background: #dc2626; color: #fff">{{ hitung.queryTerkirim > 99 ? "99+" : hitung.queryTerkirim }}</span>
          <span v-else-if="it.ke === '/ledger' && hitung.ledgerAktif" class="nav-badge" style="background: #0f766e; color: #fff">{{ hitung.ledgerAktif > 99 ? "99+" : hitung.ledgerAktif }}</span>
        </button>
      </template>

      <div class="spasi"></div>
      <div class="pengguna">
        <div class="avatar">{{ inisialNama(store.sesi?.nama ?? "?") }}</div>
        <div class="info">
          <b>{{ store.sesi?.nama }}</b>
          <span>{{ LABEL_PERAN[store.sesi?.role ?? ""] }}</span>
        </div>
      </div>
      <button class="nav-item keluar" style="margin-bottom: 14px" @click="tombolKeluar" type="button">
        <Ikon nama="keluar" :ukuran="17" />
        <span class="judul">Keluar</span>
      </button>
    </aside>

    <!-- ============ KONTEN ============ -->
    <div class="konten-area">
      <header class="header">
        <button class="btn-ikon tombol-menu-mobile" @click="sidebarBuka = !sidebarBuka" aria-label="Menu">
          <Ikon nama="menu" :ukuran="17" />
        </button>

        <div class="halaman">
          <span class="mini">MedClaim</span>
          <span class="pemisah">/</span>
          {{ judulHalaman }}
        </div>

        <div class="spasi"></div>

        <button class="cari-kotak" @click="bukaCari" type="button" aria-label="Cari (Ctrl+K)">
          <Ikon nama="cari" :ukuran="15" />
          <span class="teks">Cari SEP, pasien, menu…</span>
          <kbd>Ctrl K</kbd>
        </button>

        <div v-if="filterAktif" class="pita-ikut">
          <span class="badge brand">Filter aktif</span>
          <button class="btn tautan kecil" @click="resetFilter">reset</button>
        </div>
        <select v-if="bisaAkses.analitik" v-model="store.filter.rs" style="width: auto; min-width: 150px" aria-label="Filter rumah sakit">
          <option value="">Semua RS</option>
          <option v-for="h in store.hospitals" :key="h.id" :value="h.id">{{ h.kode }} — {{ h.nama }}</option>
        </select>
        <select v-if="bisaAkses.analitik" v-model="store.filter.tipe" style="width: auto; min-width: 120px" aria-label="Filter tipe layanan">
          <option value="SEMUA">Semua Layanan</option>
          <option value="RANAP">Rawat Inap</option>
          <option value="RJTP">Rawat Jalan</option>
        </select>

        <button class="btn-ikon" @click="toggleTema" :aria-label="gelap ? 'Mode terang' : 'Mode gelap'">
          <Ikon :nama="gelap ? 'matahari' : 'bulan'" :ukuran="16" />
        </button>

        <div class="popover-wadah" style="position: relative">
          <button class="btn-ikon" @click="notifBuka = !notifBuka; akunBuka = false; notifBaca = true" aria-label="Notifikasi">
            <Ikon nama="lonceng" :ukuran="16" />
            <span v-if="notifikasi.length && !notifBaca" class="titik-merah" />
          </button>
          <div v-if="notifBuka" class="popover">
            <div class="kepala-pop">
              <Ikon nama="lonceng" :ukuran="15" />
              Notifikasi
              <span style="margin-left: auto" class="badge abu">{{ notifikasi.length }}</span>
            </div>
            <div class="badan-pop">
              <div v-for="(n, i) in notifikasi" :key="i" class="item-notif" @click="keNotif(n.ke)">
                <span class="ikon-notif" :class="`warna-${n.warna}`">
                  <Ikon :nama="n.ikon" :ukuran="16" />
                </span>
                <div style="min-width: 0">
                  <b>{{ n.judul }}</b>
                  <div class="desk-notif">{{ n.desk }}</div>
                  <div v-if="n.waktu" class="waktu-notif">{{ tanggalWaktu(n.waktu) }}</div>
                </div>
              </div>
              <div v-if="notifikasi.length === 0" class="kosong" style="padding: 28px 16px">Tidak ada notifikasi.</div>
            </div>
          </div>
        </div>

        <div class="popover-wadah" style="position: relative">
          <button
            class="btn-ikon"
            style="width: auto; padding: 0 10px; gap: 9px; border-radius: 999px"
            @click="akunBuka = !akunBuka; notifBuka = false"
            aria-label="Menu akun"
          >
            <span class="avatar" style="width: 26px; height: 26px; font-size: 11px">{{ inisialNama(store.sesi?.nama ?? "?") }}</span>
            <Ikon nama="chevronBawah" :ukuran="13" />
          </button>
          <div v-if="akunBuka" class="popover" style="width: 280px">
            <div style="padding: 16px; display: flex; gap: 12px; align-items: center; border-bottom: 1px solid var(--border)">
              <span class="avatar" style="width: 42px; height: 42px; font-size: 15px">{{ inisialNama(store.sesi?.nama ?? "?") }}</span>
              <div style="min-width: 0">
                <b style="font-size: 13.5px; display: block; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ store.sesi?.nama }}</b>
                <span class="tunggal" style="font-size: 11.5px">{{ store.sesi?.email }}</span>
              </div>
            </div>
            <div style="padding: 10px 16px; display: flex; align-items: center; gap: 8px; border-bottom: 1px solid var(--border)">
              <span class="badge brand">{{ LABEL_PERAN[store.sesi?.role ?? ""] }}</span>
            </div>
            <button class="nav-item" style="color: var(--text-2); margin: 8px; width: calc(100% - 16px)" @click="tombolKeluar" type="button">
              <Ikon nama="keluar" :ukuran="16" />
              <span class="judul">Keluar</span>
            </button>
          </div>
        </div>
      </header>

      <main class="isi" style="flex: 1">
        <router-view />
      </main>

      <footer class="app-footer">
        <span><b>MedClaim</b> v3.0 — Go + Vue.js</span>
        <span>•</span>
        <span>Produk <b>SPKD</b> (Sistem Pelayanan Kesehatan dan Data)</span>
        <span>•</span>
        <span>Data demo sintetis — tanpa PII asli</span>
      </footer>
    </div>

    <!-- Modal pencarian global -->
    <SearchModal v-if="cariBuka" @tutup="cariBuka = false" />
  </div>
</template>
