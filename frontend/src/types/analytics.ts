// MedClaim — tipe domain: analitik (cermin respons backend Go)

export interface KPI {
  totalEpisode: number;
  totalTarif: number;
  totalBilling: number;
  totalSelisih: number;
  pctRugiRata: number;
  potensiRecovery: number;
  estimasiRugi: number;
  readinessRata: number;
  dqRata: number;
  losRata: number;
  jumlahUpload: number;
  ledgerAktif: number;
  queryTerbuka: number;
  queryTerlambat: number;
  anomaliJumlah: number;
}

// Hasil GET /api/dashboard (DPJP hanya menerima kpi & readiness)
export interface Dash {
  kpi: KPI;
  distribusiRugi: Record<string, number>;
  tren: { periode: string; episode: number; rugi: number }[];
  topCbg: { kode: string; desk: string; jumlah: number; rugi: number }[];
  kelasRawat: { kelas: string; jumlah: number }[];
  readiness: { total: number; komponen: Record<string, number> };
}

// Hasil GET /api/financial
export interface Fin {
  total: number;
  komponen: { komponen: string; nilai: number; pct: number }[];
}

export interface AnomaliItem {
  episodeId: string;
  jenis: string;
  detail: string;
  sep: string;
  mr: string;
  nama: string;
  icbg: string;
  los: number;
  tglMasuk: string;
}

export interface DaftarAnomali {
  total: number;
  items: AnomaliItem[];
}
