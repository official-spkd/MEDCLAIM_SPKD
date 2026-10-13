import { get, post, hapus } from "../client";
import type { Upload, UploadPreview, UploadHasil, UploadRequest, SampelFile } from "@/types";

export const uploadsApi = {
  /** GET /api/uploads */
  list: () => get<{ items: Upload[] }>("/api/uploads"),

  /** POST /api/uploads/preview — validasi + skor DQ tanpa menyimpan. */
  preview: (body: UploadRequest) => post<UploadPreview>("/api/uploads/preview", body),

  /** POST /api/uploads/commit — simpan; gerbang DQ BLOKIR menjawab 409 kecuali override. */
  commit: (body: UploadRequest) => post<UploadHasil>("/api/uploads/commit", body),

  /** POST /api/uploads/:id/arsip */
  arsip: (id: string) => post<Upload>(`/api/uploads/${encodeURIComponent(id)}/arsip`),

  /** DELETE /api/uploads/:id — hanya SUPERADMIN. */
  hapus: (id: string) => hapus<{ ok: boolean }>(`/api/uploads/${encodeURIComponent(id)}`),
};

export const samplesApi = {
  /** GET /api/samples */
  list: () => get<{ items: SampelFile[] }>("/api/samples"),

  /** GET /api/samples/:nama */
  get: (nama: string) => get<{ nama: string; isi: string }>(`/api/samples/${encodeURIComponent(nama)}`),
};
