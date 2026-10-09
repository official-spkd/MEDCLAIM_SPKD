<script setup lang="ts">
// ============================================================
// MedClaim — SearchModal (Ctrl/Cmd+K): navigasi + episode
// ============================================================
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { store, bisaAkses } from "../store";
import { api } from "../api";
import { usia } from "../format";
import Ikon from "./Ikon.vue";
import type { Episode } from "../types";

const emit = defineEmits<{ tutup: [] }>();
const router = useRouter();

const kunci = ref("");
const fokusInput = ref<HTMLInputElement | null>(null);
const indeksAktif = ref(0);
const hasilEpisode = ref<Episode[]>([]);
const mencari = ref(false);

onMounted(() => nextTick(() => fokusInput.value?.focus()));

const halaman = computed(() => {
  const a = bisaAkses.value;
  const semua: { ke: string; judul: string; ikon: string; tampil: boolean }[] = [
    { ke: "/", judul: "Dashboard", ikon: "dash", tampil: a.analitik || store.sesi?.role === "DPJP" },
    { ke: "/klaim", judul: "Claim Explorer", ikon: "klaim", tampil: a.analitik || store.sesi?.role === "DPJP" },
    { ke: "/pasien", judul: "Patient Tracer", ikon: "pasien", tampil: a.analitik || store.sesi?.role === "DPJP" },
    { ke: "/finansial", judul: "Financial Intelligence", ikon: "uang", tampil: a.finansial },
    { ke: "/upload", judul: "Upload Data Klaim", ikon: "unggah", tampil: a.upload },
    { ke: "/kualitas", judul: "Claim Quality", ikon: "mutu", tampil: a.upload || store.sesi?.role === "DIREKTUR" },
    { ke: "/anomali", judul: "Anomaly Detection", ikon: "anomali", tampil: a.finansial },
    { ke: "/ledger", judul: "Recovery Ledger", ikon: "buku", tampil: a.ledger },
    { ke: "/query", judul: "Query & Klarifikasi", ikon: "tanya", tampil: a.query || store.sesi?.role === "DPJP" },
    { ke: "/admin", judul: "Administrasi", ikon: "atur", tampil: a.admin },
    { ke: "/dokumentasi", judul: "Dokumentasi", ikon: "buku2", tampil: true },
  ];
  return semua.filter((h) => h.tampil);
});

const q = computed(() => kunci.value.trim().toLowerCase());
const halamanCocok = computed(() =>
  q.value ? halaman.value.filter((h) => h.judul.toLowerCase().includes(q.value)).slice(0, 5) : halaman.value.slice(0, 5),
);

const bolehCariEpisode = computed(() => bisaAkses.value.analitik || store.sesi?.role === "DPJP");

watch(kunci, () => {
  indeksAktif.value = 0;
  cariEpisode();
});
let timer: ReturnType<typeof setTimeout> | undefined;
async function cariEpisode() {
  clearTimeout(timer);
  if (!q.value || q.value.length < 2 || !bolehCariEpisode.value) {
    hasilEpisode.value = [];
    mencari.value = false;
    return;
  }
  mencari.value = true;
  timer = setTimeout(async () => {
    try {
      const d = await api.get<{ items: Episode[] }>("/api/episodes?q=" + encodeURIComponent(kunci.value.trim()));
      hasilEpisode.value = d.items.slice(0, 6);
    } catch {
      hasilEpisode.value = [];
    } finally {
      mencari.value = false;
    }
  }, 320);
}

const totalHasil = computed(() => halamanCocok.value.length + hasilEpisode.value.length);

function pilihHalaman(ke: string) {
  emit("tutup");
  router.push(ke);
}
function pilihEpisode(ep: Episode) {
  emit("tutup");
  router.push({ path: "/klaim", query: { q: ep.noSep } });
}

function tekan(e: KeyboardEvent) {
  if (e.key === "Escape") emit("tutup");
  else if (e.key === "ArrowDown") {
    e.preventDefault();
    indeksAktif.value = Math.min(indeksAktif.value + 1, totalHasil.value - 1);
  } else if (e.key === "ArrowUp") {
    e.preventDefault();
    indeksAktif.value = Math.max(indeksAktif.value - 1, 0);
  } else if (e.key === "Enter") {
    e.preventDefault();
    if (indeksAktif.value < halamanCocok.value.length) pilihHalaman(halamanCocok.value[indeksAktif.value].ke);
    else pilihEpisode(hasilEpisode.value[indeksAktif.value - halamanCocok.value.length]);
  }
}
function tekanGlobal(e: KeyboardEvent) {
  if (e.key === "Escape") emit("tutup");
}
onMounted(() => window.addEventListener("keydown", tekanGlobal));
onUnmounted(() => window.removeEventListener("keydown", tekanGlobal));
</script>

<template>
  <div class="cari-topeng" @click.self="emit('tutup')">
    <div class="kotak-cari">
      <div class="baris-cari">
        <Ikon nama="cari" :ukuran="19" />
        <input
          ref="fokusInput"
          v-model="kunci"
          placeholder="Cari menu, SEP, pasien, MR…"
          @keydown="tekan"
        />
        <span v-if="mencari" class="spinner" style="width: 16px; height: 16px; border-width: 2px; flex-shrink: 0" />
        <kbd style="font-family: var(--font-teks); font-size: 10px; color: var(--text-3); border: 1px solid var(--border); border-radius: 5px; padding: 1px 6px">ESC</kbd>
      </div>

      <div class="hasil-cari">
        <template v-if="halamanCocok.length">
          <div class="grup-cari">Navigasi</div>
          <div
            v-for="(h, i) in halamanCocok"
            :key="h.ke"
            class="hasil-item"
            :class="{ terpilih: indeksAktif === i }"
            @click="pilihHalaman(h.ke)"
            @mousemove="indeksAktif = i"
          >
            <span class="ikon-hasil"><Ikon :nama="h.ikon" :ukuran="15" /></span>
            <div>
              <b>{{ h.judul }}</b>
              <span class="sub-hasil">Buka halaman</span>
            </div>
            <span class="kbd-go">↵</span>
          </div>
        </template>

        <template v-if="q.length >= 2 && bolehCariEpisode">
          <div class="grup-cari">Episode &amp; Pasien</div>
          <div
            v-for="(ep, i) in hasilEpisode"
            :key="ep.id"
            class="hasil-item"
            :class="{ terpilih: indeksAktif === halamanCocok.length + i }"
            @click="pilihEpisode(ep)"
            @mousemove="indeksAktif = halamanCocok.length + i"
          >
            <span class="ikon-hasil"><Ikon nama="pasien" :ukuran="15" /></span>
            <div>
              <b>{{ ep.namaPasien }} <span class="tunggal" style="font-size: 11px">· {{ ep.noSep }}</span></b>
              <span class="sub-hasil">{{ ep.noMr }} · {{ usia(ep.tglLahir, ep.tglMasuk) }} th · {{ ep.kodeIcbg }} · LOS {{ ep.los }}</span>
            </div>
            <span class="kbd-go">↵</span>
          </div>
          <div v-if="!mencari && hasilEpisode.length === 0" class="kosong-cari" style="padding: 14px">
            Tidak ada episode yang cocok dengan "{{ kunci }}".
          </div>
        </template>

        <div v-if="totalHasil === 0 && !mencari" class="kosong-cari">Tidak ada hasil — coba kata kunci lain.</div>
      </div>

      <div
        style="padding: 9px 16px; border-top: 1px solid var(--border); font-size: 11px; color: var(--text-3); display: flex; gap: 14px"
      >
        <span>↑↓ navigasi</span>
        <span>↵ buka</span>
        <span>esc tutup</span>
      </div>
    </div>
  </div>
</template>
