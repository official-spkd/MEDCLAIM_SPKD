// ============================================================
// MedClaim — sumber tunggal menu navigasi + aturan siapa boleh membukanya.
// Dipakai Shell (sidebar) DAN router (guard), supaya keduanya tak pernah beda.
// ============================================================
import type { Peran } from "@/types";
import type { Akses } from "@/stores/session";

export type GrupNav = "ANALITIK" | "OPERASIONAL" | "LAINNYA";

export interface ItemNav {
  ke: string;
  judul: string;
  ikon: string;
  grup: GrupNav;
  /** Apakah peran/izin ini boleh melihat menu & membuka halamannya. */
  boleh: (a: Akses, peran?: Peran) => boolean;
}

const analitikAtauDpjp = (a: Akses, peran?: Peran) => a.analitik || peran === "DPJP";

// Urutan array = urutan tampil di sidebar.
export const ITEM_NAV: ItemNav[] = [
  { ke: "/", judul: "Dashboard", ikon: "dash", grup: "ANALITIK", boleh: analitikAtauDpjp },
  { ke: "/klaim", judul: "Claim Explorer", ikon: "klaim", grup: "ANALITIK", boleh: analitikAtauDpjp },
  { ke: "/pasien", judul: "Patient Tracer", ikon: "pasien", grup: "ANALITIK", boleh: analitikAtauDpjp },
  { ke: "/finansial", judul: "Financial", ikon: "uang", grup: "ANALITIK", boleh: (a) => a.finansial },
  { ke: "/upload", judul: "Upload Klaim", ikon: "unggah", grup: "OPERASIONAL", boleh: (a) => a.upload },
  { ke: "/kualitas", judul: "Claim Quality", ikon: "mutu", grup: "OPERASIONAL", boleh: (a, p) => a.upload || p === "DIREKTUR" },
  { ke: "/anomali", judul: "Anomaly Detection", ikon: "anomali", grup: "OPERASIONAL", boleh: (a) => a.finansial },
  { ke: "/ledger", judul: "Recovery Ledger", ikon: "buku", grup: "OPERASIONAL", boleh: (a) => a.ledger },
  { ke: "/query", judul: "Query", ikon: "tanya", grup: "OPERASIONAL", boleh: (a, p) => a.query || p === "DPJP" },
  { ke: "/admin", judul: "Administrasi", ikon: "atur", grup: "LAINNYA", boleh: (a) => a.admin },
  { ke: "/dokumentasi", judul: "Dokumentasi", ikon: "buku2", grup: "LAINNYA", boleh: () => true },
];

export const URUTAN_GRUP: GrupNav[] = ["ANALITIK", "OPERASIONAL", "LAINNYA"];

/** Boleh membuka path ini? Path yang tidak ada di menu dianggap boleh. */
export function bolehBukaPath(path: string, a: Akses, peran?: Peran): boolean {
  const item = ITEM_NAV.find((i) => i.ke === path);
  return item ? item.boleh(a, peran) : true;
}
