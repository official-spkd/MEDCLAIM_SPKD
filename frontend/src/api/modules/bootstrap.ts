import { get } from "../client";
import type { Bootstrap } from "@/types";

export const bootstrapApi = {
  /** GET /api/bootstrap — sesi, RS, upload, template, pengaturan, dokter (+ users untuk SUPERADMIN). */
  get: () => get<Bootstrap>("/api/bootstrap"),
};
