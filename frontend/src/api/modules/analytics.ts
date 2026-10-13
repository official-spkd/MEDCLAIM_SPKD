import { get, type Query } from "../client";
import type { Dash, Fin, DaftarAnomali } from "@/types";

// `params` = filter global (rs, upload, tipe, dari, sampai) dari stores/session.filterParams().
export const analyticsApi = {
  /** GET /api/dashboard — DPJP hanya menerima kpi & readiness. */
  dashboard: (params?: Query) => get<Dash>("/api/dashboard", params),

  /** GET /api/anomali */
  anomali: (params?: Query) => get<DaftarAnomali>("/api/anomali", params),

  /** GET /api/financial — 403 untuk DPJP. */
  financial: (params?: Query) => get<Fin>("/api/financial", params),
};
