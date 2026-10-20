// ============================================================
// MedClaim Vue — State global (sesi, bootstrap, filter global, izin)
// ============================================================
import { reactive, computed } from "vue";
import { ApiError, authApi, bootstrapApi, setPenangan401, type Query } from "@/api";
import type { Bootstrap, Sesi, Hospital, Upload } from "@/types";

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

/**
 * Filter global (RS, upload, tipe, rentang tanggal) sebagai parameter query,
 * siap diteruskan ke fungsi di "@/api". `tambahan` ikut dikirim bila tidak kosong.
 */
export function filterParams(tambahan: Query = {}): Query {
  const p: Query = {};
  if (store.filter.rs) p.rs = store.filter.rs;
  if (store.filter.upload) p.upload = store.filter.upload;
  if (store.filter.tipe && store.filter.tipe !== "SEMUA") p.tipe = store.filter.tipe;
  if (store.filter.dari) p.dari = store.filter.dari;
  if (store.filter.sampai) p.sampai = store.filter.sampai;
  for (const [k, v] of Object.entries(tambahan)) if (v) p[k] = v;
  return p;
}

export async function muatBootstrap(): Promise<void> {
  const data = await bootstrapApi.get();
  store.sesi = data.me;
  store.hospitals = data.hospitals;
  store.uploads = data.uploads;
  store.templates = data.templates;
  store.settings = data.settings;
  store.dokter = data.dokter;
  store.users = data.users;
}

export async function masuk(email: string, sandi: string): Promise<void> {
  store.sesi = await authApi.login(email, sandi);
  await muatBootstrap();
}

export async function keluar(): Promise<void> {
  try {
    await authApi.logout();
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

// Sesi berakhir di tengah pemakaian (cookie kedaluwarsa / akun dinonaktifkan):
// kosongkan sesi; router (router/index.ts) memantau store.sesi dan mengarahkan ke login.
setPenangan401(() => {
  if (store.sesi) {
    store.sesi = null;
    store.uploads = [];
  }
});

// ---------- Izin RBAC (cermin internal/auth/rbac.go di backend) ----------

export interface Akses {
  analitik: boolean;
  upload: boolean;
  ledger: boolean;
  ledgerTulis: boolean;
  ledgerTransisi: boolean;
  query: boolean;
  admin: boolean;
  finansial: boolean;
}

export const bisaAkses = computed<Akses>(() => {
  const peran = store.sesi?.role;
  return {
    analitik: ["SUPERADMIN", "CASEMIX", "CODER", "DIREKTUR"].includes(peran ?? ""),
    upload: peran === "SUPERADMIN" || peran === "CASEMIX",
    ledger: ["SUPERADMIN", "CASEMIX", "CODER", "DIREKTUR"].includes(peran ?? ""),
    ledgerTulis: peran === "SUPERADMIN" || peran === "CASEMIX",
    ledgerTransisi: ["SUPERADMIN", "CASEMIX", "CODER"].includes(peran ?? ""),
    query: ["SUPERADMIN", "CASEMIX", "CODER"].includes(peran ?? ""),
    admin: peran === "SUPERADMIN",
    finansial: peran !== "DPJP",
  };
});
