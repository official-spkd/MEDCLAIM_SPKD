// ============================================================
// MedClaim Vue — Router hash + guard sesi
// ============================================================
import { createRouter, createWebHashHistory } from "vue-router";
import { store, cobaSesi } from "./store";
import Shell from "./components/Shell.vue";
import Login from "./views/Login.vue";
import Dashboard from "./views/Dashboard.vue";
import Klaim from "./views/Klaim.vue";
import Upload from "./views/Upload.vue";
import Kualitas from "./views/Kualitas.vue";
import Ledger from "./views/Ledger.vue";
import Query from "./views/Query.vue";
import Pasien from "./views/Pasien.vue";
import Finansial from "./views/Finansial.vue";
import Anomali from "./views/Anomali.vue";
import Admin from "./views/Admin.vue";
import Dokumentasi from "./views/Dokumentasi.vue";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/login", name: "login", component: Login, meta: { publik: true } },
    {
      path: "/",
      component: Shell,
      children: [
        { path: "", name: "dashboard", component: Dashboard, meta: { judul: "Dashboard" } },
        { path: "klaim", name: "klaim", component: Klaim, meta: { judul: "Claim Explorer" } },
        { path: "pasien", name: "pasien", component: Pasien, meta: { judul: "Patient Tracer" } },
        { path: "finansial", name: "finansial", component: Finansial, meta: { judul: "Financial Intelligence" } },
        { path: "upload", name: "upload", component: Upload, meta: { judul: "Upload Data Klaim" } },
        { path: "kualitas", name: "kualitas", component: Kualitas, meta: { judul: "Claim Quality" } },
        { path: "anomali", name: "anomali", component: Anomali, meta: { judul: "Anomaly Detection" } },
        { path: "ledger", name: "ledger", component: Ledger, meta: { judul: "Recovery Ledger" } },
        { path: "query", name: "query", component: Query, meta: { judul: "Query & Klarifikasi" } },
        { path: "admin", name: "admin", component: Admin, meta: { judul: "Administrasi" } },
        { path: "dokumentasi", name: "dokumentasi", component: Dokumentasi, meta: { judul: "Dokumentasi" } },
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
  return true;
});

export default router;
