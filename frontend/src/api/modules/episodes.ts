import { get, type Query } from "../client";
import type { DaftarEpisode, EpisodeDetail } from "@/types";

export const episodesApi = {
  /**
   * GET /api/episodes. Parameter: q, kelasRugi, kelas, sev + filter global
   * (rs, upload, tipe, dari, sampai).
   */
  list: (params?: Query) => get<DaftarEpisode>("/api/episodes", params),

  /** GET /api/episodes/:id — episode + readiness + anomali + riwayat versi. */
  detail: (id: string) => get<EpisodeDetail>(`/api/episodes/${encodeURIComponent(id)}`),
};
