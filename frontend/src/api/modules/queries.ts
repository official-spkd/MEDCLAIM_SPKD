import { get, post } from "../client";
import type { QueryDok, TemplateDok, BuatQueryBody } from "@/types";

export const queriesApi = {
  /** GET /api/queries — status: DRAFT | TERKIRIM | DIJAWAB | TERLAMBAT (kosong = semua). */
  list: (status?: string) => get<{ items: QueryDok[] }>("/api/queries", { status }),

  /** POST /api/queries — kirim query ke DPJP. */
  create: (body: BuatQueryBody) => post<QueryDok>("/api/queries", body),

  /** POST /api/queries/:id/answer */
  answer: (id: string, jawaban: string) =>
    post<QueryDok>(`/api/queries/${encodeURIComponent(id)}/answer`, { jawaban }),

  /** GET /api/templates */
  templates: () => get<{ items: TemplateDok[] }>("/api/templates"),
};
