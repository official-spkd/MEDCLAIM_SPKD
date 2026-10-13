// ============================================================
// Transport mode demo: backend Go berjalan di browser (WebAssembly),
// data dummy. Dipisah dari klien HTTP agar mudah dihapus bila tak dipakai lagi.
// ============================================================
import { BASE_URL, MODE_DEMO } from "../config";

export interface RespMentah {
  ok: boolean;
  status: number;
  teks: string;
}

interface GoRuntime {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): void;
}

interface DemoWindow extends Window {
  Go?: new () => GoRuntime;
  medclaimSiap?: boolean;
  medclaimFetch?: (method: string, path: string, body: string, cookie: string) => string | Promise<string>;
}

const demoWindow = (): DemoWindow => window as DemoWindow;

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

/** Menyiapkan runtime WASM (sekali). Tidak melakukan apa-apa di luar mode demo. */
export function siapkanDemo(): Promise<void> {
  if (!MODE_DEMO) return Promise.resolve();
  if (!janjiWasm) {
    janjiWasm = (async () => {
      await muatScript(`${BASE_URL}wasm_exec.js`);
      const w = demoWindow();
      if (!w.Go) throw new Error("Runtime Go WASM tidak tersedia");
      const go = new w.Go();
      const hasil = await WebAssembly.instantiateStreaming(fetch(`${BASE_URL}medclaim.wasm`), go.importObject);
      go.run(hasil.instance);
      while (!demoWindow().medclaimSiap) await new Promise((r) => setTimeout(r, 10));
    })();
  }
  return janjiWasm;
}

/** Kirim request ke backend WASM; cookie sesi dikelola manual di sessionStorage. */
export async function kirimDemo(path: string, method: string, body: string): Promise<RespMentah> {
  await siapkanDemo();
  const fetchDemo = demoWindow().medclaimFetch;
  if (!fetchDemo) throw new Error("Backend demo belum siap");
  const hasil = JSON.parse(await fetchDemo(method, path, body, cookie));
  if (hasil.setCookie) {
    const nilai = String(hasil.setCookie).split(";")[0];
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
