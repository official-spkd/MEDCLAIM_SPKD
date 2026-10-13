<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  nilai: number; // 0..100
  ukuran?: number;
  label?: string;
  warna?: string;
}>();

const uk = computed(() => props.ukuran ?? 150);
const R = computed(() => uk.value / 2 - 13);
const C = computed(() => 2 * Math.PI * R.value);
const dash = computed(() => `${(Math.min(100, Math.max(0, props.nilai)) / 100) * C.value * 0.75} ${C.value}`);
const warna = computed(() => props.warna ?? (props.nilai >= 90 ? "var(--hijau)" : props.nilai >= 60 ? "var(--kuning)" : "var(--merah)"));
</script>

<template>
  <div style="display: inline-flex; flex-direction: column; align-items: center; gap: 4px">
    <svg :width="uk" :height="uk">
      <g :transform="`rotate(135 ${uk / 2} ${uk / 2})`">
        <circle
          :cx="uk / 2" :cy="uk / 2" :r="R" fill="none" stroke="var(--border)" stroke-width="13"
          :stroke-dasharray="`${C * 0.75} ${C}`" stroke-linecap="round"
        />
        <circle
          :cx="uk / 2" :cy="uk / 2" :r="R" fill="none" :stroke="warna" stroke-width="13"
          :stroke-dasharray="dash" stroke-linecap="round" style="transition: stroke-dasharray .4s ease"
        />
      </g>
      <text :x="uk / 2" :y="uk / 2 + 2" text-anchor="middle" style="font-family: var(--font-judul); font-weight: 700; font-size: 23px" fill="var(--text)">
        {{ Math.round(nilai) }}
      </text>
      <text v-if="label" :x="uk / 2" :y="uk / 2 + 19" text-anchor="middle" style="font-size: 10.5px" fill="var(--text-3)">{{ label }}</text>
    </svg>
  </div>
</template>
