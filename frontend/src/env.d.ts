/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Basis URL backend, tanpa slash di akhir. Contoh: http://localhost:8080 */
  readonly VITE_API_BASE?: string;
  /** "1" = mode demo (backend Go WASM di browser, data dummy). Diabaikan bila VITE_API_BASE diisi. */
  readonly VITE_MOCK?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
