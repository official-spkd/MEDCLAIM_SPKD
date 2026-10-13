/** Bentuk body error dari backend: {"pesan": "...", "data": ...}. */
export interface ApiErrorBody {
  pesan?: string;
  data?: unknown;
  // Beberapa error membawa field tambahan (mis. gerbang DQ).
  [kunci: string]: unknown;
}

/** Error HTTP dari backend; `kode` = status HTTP, `message` = `pesan` dari backend. */
export class ApiError extends Error {
  readonly kode: number;
  readonly data: ApiErrorBody | null;

  constructor(kode: number, pesan: string, data: ApiErrorBody | null = null) {
    super(pesan);
    this.name = "ApiError";
    this.kode = kode;
    this.data = data;
  }

  get tidakTerautentikasi(): boolean {
    return this.kode === 401;
  }
  get dilarang(): boolean {
    return this.kode === 403;
  }
}

/** Pesan aman untuk ditampilkan dari error apa pun (pengganti `catch (e: any)`). */
export function pesanError(e: unknown, bawaan: string): string {
  if (e instanceof Error && e.message) return e.message;
  return bawaan;
}
