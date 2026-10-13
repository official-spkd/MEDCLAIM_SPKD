// ============================================================
// MedClaim — klien HTTP inti. Semua modul API memakai `http` di sini;
// halaman/komponen TIDAK memanggil fetch langsung.
// ============================================================
import { API_BASE, GATEWAY_PORT, MODE_DEMO } from "./config";
import { ApiError, type ApiErrorBody } from "./errors";
import { kirimDemo, type RespMentah } from "./transport/demo";

/** Parameter query string. Nilai kosong ("", undefined, null) dibuang. */
export type Query = Record<string, string | number | boolean | null | undefined>;

export interface RequestOptions {
  method?: "GET" | "POST" | "DELETE";
  body?: unknown;
  query?: Query;
  /** true = 401 pada request ini TIDAK memicu penanganan "sesi berakhir" (mis. login). */
  abaikan401?: boolean;
}

let penangan401: (() => void) | null = null;

/**
 * Mendaftarkan callback yang dipanggil saat backend menjawab 401 (sesi berakhir)
 * di tengah pemakaian. Dipasang oleh store sesi; klien sendiri tidak tahu soal store.
 */
export function setPenangan401(fn: (() => void) | null): void {
  penangan401 = fn;
}

export function buatQueryString(query?: Query): string {
  if (!query) return "";
  const p = new URLSearchParams();
  for (const [k, v] of Object.entries(query)) {
    if (v === undefined || v === null || v === "") continue;
    p.set(k, String(v));
  }
  const s = p.toString();
  return s ? `?${s}` : "";
}

function bangunUrl(path: string, query?: Query): string {
  const penuh = path + buatQueryString(query);
  if (API_BASE) return `${API_BASE}${penuh}`;
  // XTransformPort hanya dipakai gateway sandbox (tanpa VITE_API_BASE).
  return `${penuh}${penuh.includes("?") ? "&" : "?"}XTransformPort=${GATEWAY_PORT}`;
}

async function kirim(url: string, opsi: RequestOptions, body: string | undefined): Promise<RespMentah> {
  const r = await fetch(url, {
    method: opsi.method ?? "GET",
    credentials: API_BASE ? "include" : "same-origin",
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body,
  });
  return { ok: r.ok, status: r.status, teks: await r.text() };
}

function parseJson(teks: string): unknown {
  try {
    return teks ? JSON.parse(teks) : null;
  } catch {
    return null;
  }
}

export async function http<T>(path: string, opsi: RequestOptions = {}): Promise<T> {
  const body = opsi.body === undefined || opsi.body === null ? undefined : JSON.stringify(opsi.body);
  const res = MODE_DEMO
    ? await kirimDemo(path + buatQueryString(opsi.query), opsi.method ?? "GET", body ?? "")
    : await kirim(bangunUrl(path, opsi.query), opsi, body);

  const data = parseJson(res.teks);
  if (!res.ok) {
    const err = (data && typeof data === "object" ? data : null) as ApiErrorBody | null;
    if (res.status === 401 && !opsi.abaikan401) penangan401?.();
    throw new ApiError(res.status, err?.pesan ?? `HTTP ${res.status}`, err);
  }
  return data as T;
}

export const get = <T>(path: string, query?: Query) => http<T>(path, { query });
export const post = <T>(path: string, body?: unknown, opsi?: Pick<RequestOptions, "query" | "abaikan401">) =>
  http<T>(path, { method: "POST", body, ...opsi });
export const hapus = <T>(path: string) => http<T>(path, { method: "DELETE" });
