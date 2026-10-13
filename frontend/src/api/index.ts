// ============================================================
// MedClaim — pintu masuk tunggal lapisan API.
// Halaman/komponen/store mengimpor dari "@/api", bukan dari berkas dalam.
// ============================================================
export { ApiError, pesanError } from "./errors";
export type { ApiErrorBody } from "./errors";
export { setPenangan401, buatQueryString } from "./client";
export type { Query } from "./client";
export { MODE_DEMO as modeDemo } from "./config";
export { siapkanDemo } from "./transport/demo";

export { authApi } from "./modules/auth";
export { bootstrapApi } from "./modules/bootstrap";
export { analyticsApi } from "./modules/analytics";
export { episodesApi } from "./modules/episodes";
export { uploadsApi, samplesApi } from "./modules/uploads";
export { ledgerApi } from "./modules/ledger";
export { queriesApi } from "./modules/queries";
export { adminApi } from "./modules/admin";
