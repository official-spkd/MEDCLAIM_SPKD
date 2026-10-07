<script setup lang="ts">
// ============================================================
// MedClaim — GrafikTren: grafik tren multi-seri (SVG) ala MedCredix
// - seri "area": kurva halus + isian gradien
// - seri "garis": kurva putus-putus (pembanding)
// - sumbu ganda (kiri/kanan), grid halus, label, tooltip titik
// ============================================================
import { computed } from "vue";

export interface SeriTren {
  nama: string;
  warna: string;
  nilai: number[];
  sumbu?: "kiri" | "kanan";
  jenis?: "area" | "garis";
}

const props = withDefaults(
  defineProps<{
    label: string[];
    seri: SeriTren[];
    tinggi?: number;
    formatKiri?: (n: number) => string;
    formatKanan?: (n: number) => string;
  }>(),
  { tinggi: 250 },
);

const uid = Math.random().toString(36).slice(2, 9);
const W = 780;
const pad = { atas: 18, kanan: 58, bawah: 32, kiri: 54 };

const tinggiPlot = computed(() => props.tinggi - pad.atas - pad.bawah);
const lebarPlot = W - pad.kiri - pad.kanan;

function angkaBagus(n: number): number {
  if (n <= 0) return 1;
  const p = Math.pow(10, Math.floor(Math.log10(n)));
  return Math.ceil(n / p) * p;
}

const maksKiri = computed(() => {
  const v = props.seri
    .filter((s) => (s.sumbu ?? "kiri") === "kiri")
    .flatMap((s) => s.nilai);
  return angkaBagus(Math.max(1, ...v));
});

const maksKanan = computed(() => {
  const v = props.seri
    .filter((s) => s.sumbu === "kanan")
    .flatMap((s) => s.nilai);
  return v.length ? angkaBagus(Math.max(1, ...v)) : 0;
});

const fmtKiri = computed(() => props.formatKiri ?? ((n: number) => n.toLocaleString("id-ID")));
const fmtKanan = computed(() => props.formatKanan ?? fmtKiri.value);

const fraksi = [0, 0.25, 0.5, 0.75, 1];

const x = (i: number) =>
  pad.kiri + (lebarPlot * i) / Math.max(1, props.label.length - 1);

function y(v: number, sumbu: "kiri" | "kanan" = "kiri"): number {
  const maks = sumbu === "kanan" && maksKanan.value ? maksKanan.value : maksKiri.value;
  return pad.atas + tinggiPlot.value * (1 - Math.min(v, maks) / maks);
}

function titik(seri: SeriTren): [number, number][] {
  const sumbu = seri.sumbu ?? "kiri";
  return seri.nilai.map((v, i) => [x(i), y(v, sumbu)]);
}

function jalurHalus(poin: [number, number][]): string {
  if (poin.length === 0) return "";
  if (poin.length === 1) return `M ${poin[0][0]},${poin[0][1]}`;
  let d = `M ${poin[0][0].toFixed(1)},${poin[0][1].toFixed(1)}`;
  for (let i = 0; i < poin.length - 1; i++) {
    const p0 = poin[Math.max(0, i - 1)];
    const p1 = poin[i];
    const p2 = poin[i + 1];
    const p3 = poin[Math.min(poin.length - 1, i + 2)];
    const c1x = p1[0] + (p2[0] - p0[0]) / 6;
    const c1y = p1[1] + (p2[1] - p0[1]) / 6;
    const c2x = p2[0] - (p3[0] - p1[0]) / 6;
    const c2y = p2[1] - (p3[1] - p1[1]) / 6;
    d += ` C ${c1x.toFixed(1)},${c1y.toFixed(1)} ${c2x.toFixed(1)},${c2y.toFixed(1)} ${p2[0].toFixed(1)},${p2[1].toFixed(1)}`;
  }
  return d;
}

const jalurSeri = computed(() =>
  props.seri.map((s) => {
    const poin = titik(s);
    const garis = jalurHalus(poin);
    const area =
      poin.length > 1
        ? `${garis} L ${poin[poin.length - 1][0].toFixed(1)},${(pad.atas + tinggiPlot.value).toFixed(1)} L ${poin[0][0].toFixed(1)},${(pad.atas + tinggiPlot.value).toFixed(1)} Z`
        : "";
    return { seri: s, poin, garis, area };
  }),
);

// Label-x: tampilkan maksimal ~9 agar tidak bertumpuk
const langkahLabel = computed(() => Math.max(1, Math.ceil(props.label.length / 9)));

function fmtSeri(s: SeriTren, v: number): string {
  return (s.sumbu === "kanan" ? fmtKanan.value : fmtKiri.value)(v);
}
</script>

<template>
  <div>
    <div class="grafik-legend">
      <span v-for="(s, i) in seri" :key="i" class="legenda">
        <span class="bulet" :style="{ background: s.warna }" />
        {{ s.nama }}
      </span>
    </div>

    <svg :viewBox="`0 0 ${W} ${tinggi}`" style="width: 100%; height: auto; display: block">
      <!-- Grid + label sumbu -->
      <g v-for="(f, i) in fraksi" :key="i">
        <line
          :x1="pad.kiri" :y1="pad.atas + tinggiPlot * f"
          :x2="W - pad.kanan" :y2="pad.atas + tinggiPlot * f"
          :stroke="i === fraksi.length - 1 ? 'var(--border-kuat)' : 'var(--border)'"
          :stroke-width="i === fraksi.length - 1 ? 1.2 : 1"
          :stroke-dasharray="i === 0 ? 'none' : '3 5'"
        />
        <text
          :x="pad.kiri - 8" :y="pad.atas + tinggiPlot * f + 3.5"
          text-anchor="end" font-size="11" fill="var(--text-2)"
          style="font-family: var(--font-teks)"
        >{{ fmtKiri(maksKiri * (1 - f)) }}</text>
        <text
          v-if="maksKanan"
          :x="W - pad.kanan + 8" :y="pad.atas + tinggiPlot * f + 3.5"
          text-anchor="start" font-size="11" fill="var(--text-2)"
          style="font-family: var(--font-teks)"
        >{{ fmtKanan(maksKanan * (1 - f)) }}</text>
      </g>

      <!-- Label-x -->
      <g>
        <text
          v-for="(l, i) in label" v-show="i % langkahLabel === 0 || i === label.length - 1"
          :key="i"
          :x="x(i)" :y="tinggi - 10"
          text-anchor="middle" font-size="11" fill="var(--text-2)"
          style="font-family: var(--font-teks)"
        >{{ l }}</text>
      </g>

      <!-- Seri -->
      <g v-for="(j, i) in jalurSeri" :key="i">
        <defs v-if="j.seri.jenis !== 'garis'">
          <linearGradient :id="`grad-s-${uid}-${i}`" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" :stop-color="j.seri.warna" stop-opacity="0.3" />
            <stop offset="100%" :stop-color="j.seri.warna" stop-opacity="0.02" />
          </linearGradient>
        </defs>
        <path
          v-if="j.seri.jenis !== 'garis'"
          :d="j.area" :fill="`url(#grad-s-${uid}-${i})`"
        />
        <path
          :d="j.garis" fill="none" :stroke="j.seri.warna"
          :stroke-width="j.seri.jenis === 'garis' ? 2.2 : 2.6"
          :stroke-dasharray="j.seri.jenis === 'garis' ? '6 5' : 'none'"
          stroke-linecap="round" stroke-linejoin="round"
        />
        <g v-if="j.seri.jenis !== 'garis' || label.length <= 12">
          <circle
            v-for="(p, k) in j.poin" :key="k"
            :cx="p[0]" :cy="p[1]" r="3.4"
            :fill="j.seri.warna" stroke="var(--surface)" stroke-width="2"
          >
            <title>{{ j.seri.nama }} · {{ label[k] }}: {{ fmtSeri(j.seri, j.seri.nilai[k]) }}</title>
          </circle>
        </g>
      </g>
    </svg>
  </div>
</template>
