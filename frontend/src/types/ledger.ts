// MedClaim — tipe domain: recovery ledger (cermin respons backend Go)

export interface LedgerEvent {
  id: string;
  dariStatus: string | null;
  keStatus: string | null;
  aktorNama: string;
  catatan: string | null;
  waktu: string;
}

export interface Ledger {
  id: string;
  episodeId: string;
  hospitalId: string;
  sep: string;
  nomorMr: string;
  namaPasien: string;
  kodeIcbg: string;
  kelasRugi: string;
  estimasi: number;
  realisasi: number | null;
  status: string;
  referensiRm: string | null;
  catatan: string | null;
  strategy: string | null;
  createdAt: string;
  updatedAt: string;
  events?: LedgerEvent[];
}

export interface Kalibrasi {
  totalEstimasi: number;
  totalRealisasi: number;
  realizationRate: number;
  hitRate: number;
  nSampel: number;
  kalibrasiTersedia: boolean;
  realisasiNegatif: number;
}

export interface DaftarLedger {
  items: Ledger[];
  kalibrasi: Kalibrasi;
}

// Body POST /api/ledger/:id/transisi
export interface TransisiBody {
  ke: string;
  alasan?: string;
  referensiRm?: string;
  realisasi?: number;
}
