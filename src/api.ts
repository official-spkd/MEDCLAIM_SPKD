// ============================================================
// MedClaim Vue — API client: semua request relatif dengan
// XTransformPort=3030 (pintu gateway sandbox menuju backend Go)
// ============================================================

// Basis URL backend. Kosong = same-origin (sandbox gateway atau rewrite Vercel).
// Di Vercel isi VITE_API_BASE, contoh: https://api.domain-anda.com
const BASE = ((import.meta as any).env?.VITE_API_BASE ?? "").replace(/\/+$/, "");
const PORT = "3030";

function url(path: string): string {
  // XTransformPort hanya dipakai saat tanpa VITE_API_BASE (gateway sandbox)
  if (BASE) return `${BASE}${path}`;
  const gabung = path.includes("?") ? "&" : "?";
  return `${path}${gabung}XTransformPort=${PORT}`;
}

export class ApiError extends Error {
  kode: number;
  data: any;
  constructor(kode: number, pesan: string, data?: any) {
    super(pesan);
    this.kode = kode;
    this.data = data;
  }
}

// ---------- Mode demo: backend Go berjalan di browser (WebAssembly), data dummy ----------
const MOCK = (import.meta as any).env?.VITE_MOCK === "1" && !BASE;

// Mode demo aktif? (backend Go WASM + data dummy di browser)
export const modeDemo: boolean = MOCK;
const KUNCI_COOKIE = "medclaim_demo_cookie";
let cookie = "";
try {
  cookie = sessionStorage.getItem(KUNCI_COOKIE) ?? "";
} catch {
  /* abaikan */
}
let janjiWasm: Promise<void> | null = null;

function muatScript(src: string): Promise<void> {
  return new Promise((ok, gagal) => {
    const el = document.createElement("script");
    el.src = src;
    el.onload = () => ok();
    el.onerror = () => gagal(new Error("Gagal memuat " + src));
    document.head.appendChild(el);
  });
}

export function siapkanDemo(): Promise<void> {
  if (!MOCK) return Promise.resolve();
  if (!janjiWasm) {
    janjiWasm = (async () => {
      const dasar = (import.meta as any).env.BASE_URL as string;
      await muatScript(`${dasar}wasm_exec.js`);
      const go = new (window as any).Go();
      const hasil = await WebAssembly.instantiateStreaming(fetch(`${dasar}medclaim.wasm`), go.importObject);
      go.run(hasil.instance);
      while (!(window as any).medclaimSiap) await new Promise((r) => setTimeout(r, 10));
    })();
  }
  return janjiWasm;
}

async function mintaDemo(path: string, opsi: RequestInit): Promise<{ ok: boolean; status: number; teks: string }> {
  await siapkanDemo();
  const hasil = JSON.parse(
    await (window as any).medclaimFetch(opsi.method ?? "GET", path, (opsi.body as string) ?? "", cookie),
  );
  if (hasil.setCookie) {
    const nilai = hasil.setCookie.split(";")[0];
    const kosong = nilai.endsWith("=") || /max-age=-1|max-age=0/i.test(hasil.setCookie);
    cookie = kosong ? "" : nilai;
    try {
      if (cookie) sessionStorage.setItem(KUNCI_COOKIE, cookie);
      else sessionStorage.removeItem(KUNCI_COOKIE);
    } catch {
      /* abaikan */
    }
  }
  return { ok: hasil.status >= 200 && hasil.status < 300, status: hasil.status, teks: hasil.body };
}

async function minta<T>(path: string, opsi: RequestInit = {}): Promise<T> {
  let res: { ok: boolean; status: number; teks: string };
  if (MOCK) {
    res = await mintaDemo(path, opsi);
  } else {
    const r = await fetch(url(path), {
      credentials: BASE ? "include" : "same-origin",
      headers: opsi.body ? { "Content-Type": "application/json" } : undefined,
      ...opsi,
    });
    res = { ok: r.ok, status: r.status, teks: await r.text() };
  }
  const teks = res.teks;
  let data: any = null;
  try {
    data = teks ? JSON.parse(teks) : null;
  } catch {
    data = null;
  }
  if (!res.ok) {
    throw new ApiError(res.status, data?.pesan ?? `HTTP ${res.status}`, data);
  }
  return data as T;
}

export const api = {
  get: <T>(path: string) => minta<T>(path),
  post: <T>(path: string, body?: unknown) =>
    minta<T>(path, { method: "POST", body: body ? JSON.stringify(body) : undefined }),
  hapus: <T>(path: string) => minta<T>(path, { method: "DELETE" }),
};
