// ============================================================
// MedClaim Vue — State global (sesi, bootstrap, filter global)
// ============================================================
import { reactive, computed } from "vue";
import { api, ApiError } from "./api";
import type { Bootstrap, Sesi, Hospital, Upload } from "./types";

interface State {
  siap: boolean;
  sesi: Sesi | null;
  hospitals: Hospital[];
  uploads: Upload[];
  templates: Bootstrap["templates"];
  settings: Bootstrap["settings"];
  dokter: Bootstrap["dokter"];
  users: Bootstrap["users"];
  filter: {
    rs: string;
    upload: string;
    tipe: "SEMUA" | "RANAP" | "RJTP";
    dari: string;
    sampai: string;
  };
  memuat: boolean;
}

export const store = reactive<State>({
  siap: false,
  sesi: null,
  hospitals: [],
  uploads: [],
  templates: [],
  settings: [],
  dokter: [],
  users: [],
  filter: { rs: "", upload: "", tipe: "SEMUA", dari: "", sampai: "" },
  memuat: false,
});

export function queryFilter(tambahan: Record<string, string> = {}): string {
  const p = new URLSearchParams();
  if (store.filter.rs) p.set("rs", store.filter.rs);
  if (store.filter.upload) p.set("upload", store.filter.upload);
  if (store.filter.tipe && store.filter.tipe !== "SEMUA") p.set("tipe", store.filter.tipe);
  if (store.filter.dari) p.set("dari", store.filter.dari);
  if (store.filter.sampai) p.set("sampai", store.filter.sampai);
  for (const [k, v] of Object.entries(tambahan)) if (v) p.set(k, v);
  const s = p.toString();
  return s ? `?${s}` : "";
}

export async function muatBootstrap(): Promise<void> {
  const data = await api.get<Bootstrap>("/api/bootstrap");
  store.sesi = data.me;
  store.hospitals = data.hospitals;
  store.uploads = data.uploads;
  store.templates = data.templates;
  store.settings = data.settings;
  store.dokter = data.dokter;
  store.users = data.users;
}

export async function masuk(email: string, sandi: string): Promise<void> {
  store.sesi = await api.post<Sesi>("/api/auth/login", { email, password: sandi });
  await muatBootstrap();
}

export async function keluar(): Promise<void> {
  try {
    await api.post("/api/auth/logout");
  } catch {
    /* abaikan */
  }
  store.sesi = null;
  store.uploads = [];
}

export async function cobaSesi(): Promise<void> {
  store.memuat = true;
  try {
    await muatBootstrap();
  } catch (e) {
    if (e instanceof ApiError && e.kode === 401) {
      store.sesi = null;
    }
  } finally {
    store.memuat = false;
    store.siap = true;
  }
}

// ---------- Izin RBAC (cermin rbac.go) ----------

export const bisaAkses = computed(() => {
  const peran = store.sesi?.role;
  return {
    analitik: ["SUPERADMIN", "CASEMIX", "CODER", "DIREKTUR"].includes(peran ?? ""),
    upload: peran === "SUPERADMIN" || peran === "CASEMIX",
    ledger: ["SUPERADMIN", "CASEMIX", "DIREKTUR"].includes(peran ?? ""),
    ledgerTulis: peran === "SUPERADMIN" || peran === "CASEMIX",
    query: ["SUPERADMIN", "CASEMIX", "CODER"].includes(peran ?? ""),
    admin: peran === "SUPERADMIN",
    finansial: peran !== "DPJP",
  };
});
