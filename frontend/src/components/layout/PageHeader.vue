<script setup lang="ts">
// ============================================================
// MedClaim — PageHeader: H1 + badge demo + deskripsi + aksi
// (pola header halaman MedCredix)
// ============================================================
defineProps<{
  judul: string;
  sub?: string;
  demo?: boolean;
}>();
</script>

<template>
  <div class="ph">
    <div class="ph-kiri">
      <h1>
        {{ judul }}
        <span v-if="demo" class="badge-demo">● Data Demo</span>
      </h1>
      <p v-if="sub" class="ph-sub">{{ sub }}</p>
    </div>
    <div v-if="$slots.aksi" class="ph-aksi">
      <slot name="aksi" />
    </div>
  </div>
</template>
