// ============================================================
// MedClaim — konfigurasi API (satu-satunya tempat membaca env)
// ============================================================

// import.meta.env bisa tidak ada di luar Vite (mis. skrip uji di Node).
const env: Partial<ImportMetaEnv> = (import.meta as { env?: ImportMetaEnv }).env ?? {};

/**
 * Basis URL backend. Kosong = same-origin (gateway sandbox atau rewrite Vercel).
 * Lokal: VITE_API_BASE=http://localhost:8080 (lihat .env.example).
 */
export const API_BASE: string = (env.VITE_API_BASE ?? "").replace(/\/+$/, "");

/** Port gateway sandbox; hanya ditambahkan ke URL saat API_BASE kosong. */
export const GATEWAY_PORT = "3030";

/** Mode demo: backend Go WASM + data dummy di browser. Tidak aktif bila API_BASE diisi. */
export const MODE_DEMO: boolean = env.VITE_MOCK === "1" && !API_BASE;

/** Base path aplikasi (Vite `base`), dipakai memuat aset demo. */
export const BASE_URL: string = env.BASE_URL ?? "/";
