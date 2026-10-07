<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  potongan: { label: string; nilai: number; warna: string }[];
  ukuran?: number;
  teksTengah?: string;
  subTengah?: string;
}>();

const uk = computed(() => props.ukuran ?? 178);
const GAP = 10; // jarak antar segmen (px keliling)
const total = computed(() => props.potongan.reduce((a, p) => a + p.nilai, 0) || 1);

const lingkar = computed(() => {
  const R = uk.value / 2 - 15;
  const C = 2 * Math.PI * R;
  let mulai = 0;
  return props.potongan.map((p) => {
    const fraksi = p.nilai / total.value;
    const dash = Math.max(fraksi * C - GAP, 2);
    const offset = -(mulai * C + GAP / 2);
    mulai += fraksi;
    return { ...p, dash, offset, persen: fraksi, nilaiAbs: p.nilai };
  });
});
</script>

<template>
  <div style="display: flex; align-items: center; gap: 22px; flex-wrap: wrap; justify-content: center">
    <svg :width="uk" :height="uk" style="flex-shrink: 0">
      <g :transform="`rotate(-90 ${uk / 2} ${uk / 2})`">
        <circle
          v-for="(l, i) in lingkar"
          :key="i"
          class="segmen"
          :cx="uk / 2"
          :cy="uk / 2"
          :r="uk / 2 - 15"
          fill="none"
          :stroke="l.warna"
          stroke-width="23"
          :stroke-dasharray="`${l.dash} ${2 * Math.PI * (uk / 2 - 15)}`"
          :stroke-dashoffset="l.offset"
        >
          <title>{{ l.label }}: {{ l.nilai.toLocaleString("id-ID") }} episode</title>
        </circle>
      </g>
      <text
        :x="uk / 2" :y="uk / 2 - 3" text-anchor="middle"
        style="font-family: var(--font-judul); font-weight: 700; font-size: 24px"
        fill="var(--text)"
      >{{ teksTengah ?? total.toLocaleString("id-ID") }}</text>
      <text
        :x="uk / 2" :y="uk / 2 + 17" text-anchor="middle" style="font-size: 11px"
        fill="var(--text-3)"
      >{{ subTengah ?? "episode" }}</text>
    </svg>
    <div style="display: flex; flex-direction: column; gap: 9px; min-width: 168px">
      <div
        v-for="(l, i) in potongan"
        :key="i"
        style="display: flex; align-items: center; gap: 9px; font-size: 12.5px"
      >
        <span style="width: 11px; height: 11px; border-radius: 3.5px; flex-shrink: 0" :style="{ background: l.warna }" />
        <span style="flex: 1; color: var(--text-2); white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ l.label }}</span>
        <b style="font-variant-numeric: tabular-nums">{{ l.nilai.toLocaleString("id-ID") }}</b>
        <span style="color: var(--text-3); font-size: 11.5px; width: 44px; text-align: right; font-variant-numeric: tabular-nums">
          {{ (l.nilai / total * 100).toFixed(1) }}%
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.segmen {
  transition: stroke-width 0.16s ease;
  cursor: pointer;
}
.segmen:hover {
  stroke-width: 26;
}
</style>
