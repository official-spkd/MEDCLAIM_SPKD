// MedClaim — tipe domain: klaim / episode (cermin respons backend Go)

export interface DiagnosisRef {
  kode: string;
  nama: string;
}

export interface AddOn {
  kode: string;
  jenis: string;
  tarif: number;
  status: string;
}

export type KelasRugi = "SURPLUS" | "S1_CODING" | "S2_MIXED" | "STRUCTURAL" | "RAHASIA";

export interface Episode {
  id: string;
  uploadBatchId: string;
  noSep: string;
  noKartu: string;
  noMr: string;
  namaPasien: string;
  tglLahir: string;
  jenisKelamin: string;
  tglMasuk: string;
  tglKeluar: string;
  los: number;
  kelasRawat: number;
  tipeLayanan: "RANAP" | "RJTP";
  caraPulang: string | null;
  dpjpKode: string | null;
  dpjpNama: string | null;
  diagnosa: DiagnosisRef[];
  prosedur: DiagnosisRef[];
  kodeIcbg: string;
  deskripsiIcbg: string;
  severityLevel: number;
  idrgKode: string | null;
  komponen: Record<string, number>;
  totalBilling: number;
  tarifInacbg: number;
  selisih: number;
  pctRugi: number;
  kelasRugi: KelasRugi;
  addOn: AddOn[];
  dqScore: number;
  versionNo: number;
  superseded: boolean;
  createdAt: string;
}

export interface Readiness {
  total: number;
  komponen: Record<string, number>;
  flag: string[];
}

// Hasil GET /api/episodes/:id
export interface EpisodeDetail {
  episode: Episode & { readiness?: Readiness; anomali?: string[] };
  versi: Episode[];
}

export interface DaftarEpisode {
  total: number;
  items: Episode[];
}
