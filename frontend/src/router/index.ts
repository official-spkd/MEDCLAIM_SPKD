// ============================================================
// MedClaim Vue — Router hash + guard sesi & peran. Halaman dimuat malas (lazy).
// ============================================================
import { watch } from "vue";
import { createRouter, createWebHashHistory } from "vue-router";
import { store, cobaSesi, bisaAkses } from "@/stores/session";
import { bolehBukaPath } from "@/config/navigation";
import Shell from "@/components/layout/Shell.vue";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/login", name: "login", component: () => import("@/views/auth/Login.vue"), meta: { publik: true } },
    {
      path: "/",
      component: Shell,
      children: [
        { path: "", name: "dashboard", component: () => import("@/views/dashboard/Dashboard.vue"), meta: { judul: "Dashboard" } },
        { path: "klaim", name: "klaim", component: () => import("@/views/claims/Klaim.vue"), meta: { judul: "Claim Explorer" } },
        { path: "pasien", name: "pasien", component: () => import("@/views/claims/Pasien.vue"), meta: { judul: "Patient Tracer" } },
        { path: "finansial", name: "finansial", component: () => import("@/views/finance/Finansial.vue"), meta: { judul: "Financial Intelligence" } },
        { path: "upload", name: "upload", component: () => import("@/views/upload/Upload.vue"), meta: { judul: "Upload Data Klaim" } },
        { path: "kualitas", name: "kualitas", component: () => import("@/views/quality/Kualitas.vue"), meta: { judul: "Claim Quality" } },
        { path: "anomali", name: "anomali", component: () => import("@/views/quality/Anomali.vue"), meta: { judul: "Anomaly Detection" } },
        { path: "ledger", name: "ledger", component: () => import("@/views/ledger/Ledger.vue"), meta: { judul: "Recovery Ledger" } },
        { path: "query", name: "query", component: () => import("@/views/query/Query.vue"), meta: { judul: "Query & Klarifikasi" } },
        { path: "admin", name: "admin", component: () => import("@/views/admin/Admin.vue"), meta: { judul: "Administrasi" } },
        { path: "dokumentasi", name: "dokumentasi", component: () => import("@/views/docs/Dokumentasi.vue"), meta: { judul: "Dokumentasi" } },
      ],
    },
    { path: "/:pathMatch(.*)*", redirect: "/" },
  ],
});

router.afterEach((to) => {
  const judul = (to.meta.judul as string) ?? (to.name === "login" ? "Masuk" : "MedClaim");
  document.title = `${judul} — MedClaim`;
});

router.beforeEach(async (to) => {
  if (!store.siap) await cobaSesi();
  if (!to.meta.publik && !store.sesi) return { name: "login" };
  if (to.name === "login" && store.sesi) return { name: "dashboard" };
  // Peran tanpa hak atas halaman ini dikembalikan ke dashboard (backend tetap menolak 403).
  if (store.sesi && !bolehBukaPath(to.path, bisaAkses.value, store.sesi.role)) return { name: "dashboard" };
  return true;
});

// Sesi hilang di tengah pemakaian (logout, 401 dari backend) -> kembali ke login.
watch(
  () => store.sesi,
  (sesi) => {
    if (!sesi && router.currentRoute.value.name !== "login") router.replace({ name: "login" });
  },
);

export default router;
