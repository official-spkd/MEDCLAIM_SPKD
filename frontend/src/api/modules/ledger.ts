import { get, post, type Query } from "../client";
import type { DaftarLedger, Ledger, TransisiBody, HasilAjukanMassal } from "@/types";

export const ledgerApi = {
  /** GET /api/ledger — params: status, q. Menyertakan kalibrasi. */
  list: (params?: Query) => get<DaftarLedger>("/api/ledger", params),

  /** GET /api/ledger/:id — beserta riwayat event. */
  detail: (id: string) => get<Ledger>(`/api/ledger/${encodeURIComponent(id)}`),

  /** POST /api/ledger — kandidat manual dari satu episode. */
  create: (episodeId: string) => post<Ledger>("/api/ledger", { episodeId }),

  /** POST /api/ledger/:id/transisi — pindah status (mesin status di backend). */
  transisi: (id: string, body: TransisiBody) =>
    post<Ledger>(`/api/ledger/${encodeURIComponent(id)}/transisi`, body),

  /** POST /api/ledger/ajukan-massal — semua entri Dikoreksi pada satu periode → Diajukan. */
  ajukanMassal: (periode: string, rs?: string) =>
    post<HasilAjukanMassal>("/api/ledger/ajukan-massal", { periode }, { query: { rs } }),
};
