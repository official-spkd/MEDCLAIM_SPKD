<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  data: { label: string; nilai: number; warna?: string; sub?: string }[];
  tinggi?: number;
}>();

const tinggi = computed(() => props.tinggi ?? 220);
const maks = computed(() => Math.max(...props.data.map((d) => d.nilai), 1));
</script>

<template>
  <div style="overflow-x: auto">
    <div
      style="display: flex; align-items: flex-end; gap: 14px; border-bottom: 1px solid var(--border-kuat); padding-bottom: 2px"
      :style="{ height: tinggi + 'px', minWidth: data.length * 60 + 'px' }"
    >
      <div
        v-for="(d, i) in data"
        :key="i"
        style="flex: 1; display: flex; flex-direction: column; align-items: center; gap: 7px; height: 100%; justify-content: flex-end"
      >
        <div style="font-size: 11px; font-weight: 700; color: var(--text-2); font-variant-numeric: tabular-nums">
          {{ d.nilai.toLocaleString("id-ID") }}
        </div>
        <div
          class="batang"
          style="width: 68%; max-width: 46px; border-radius: 8px 8px 3px 3px; min-height: 5px"
          :style="{
            height: Math.max((d.nilai / maks) * (tinggi - 56), 5) + 'px',
            background: d.warna ?? 'linear-gradient(180deg, #f4b04e, #e8931f)',
            boxShadow: 'inset 0 10px 12px -8px rgba(255, 255, 255, 0.55)',
          }"
          :title="d.label + ': ' + d.nilai.toLocaleString('id-ID') + (d.sub ? ' — ' + d.sub : '')"
        />
        <div
          style="font-size: 10.5px; color: var(--text-3); text-align: center; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 92px"
        >
          {{ d.label }}<template v-if="d.sub"> · {{ d.sub }}</template>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.batang {
  transition: transform 0.18s ease, filter 0.18s ease, height 0.3s ease;
}
.batang:hover {
  filter: brightness(1.1) saturate(1.05);
  transform: translateY(-2px);
}
</style>
