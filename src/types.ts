// ============================================================
// MedClaim Vue — Tipe domain (cermin dari backend Go)
// ============================================================

export type Peran = "SUPERADMIN" | "CASEMIX" | "CODER" | "DIREKTUR" | "DPJP";

export interface Sesi {
  id: string;
  email: string;
  nama: string;
  role: Peran;
  hospitalId: string | null;
  hospitalNama: string | null;
  hospitalKode: string | null;
  spesialisasi: string | null;
  dpjpKode: string | null;
}

export interface Hospital {
  id: string;
  kode: string;
  nama: string;
  kota: string;
  kelas: string;
}

export interface DQAturan {
  kode: string;
  nama: string;
  dimensi: string;
  severitas: "KRITIS" | "PERINGATAN" | "INFO";
  terpicu: boolean;
  jumlah: number;
  pesan: string;
}

export interface DQReport {
  skor: number;
  grade: string;
  aturan: DQAturan[];
  dimensi: { kode: string; label: string; skor: number }[];
  matriksModul: { modul: string; status: string }[];
  periodeUsulan: string;
  gerbang: "LOLOS" | "OVERRIDE" | "BLOKIR";
}

export interface Upload {
  id: string;
  hospitalId: string;
  namaFile: string;
  periode: string;
  format: "BASIC" | "DETAIL";
  dqScore: number;
  dqGrade: string;
  dqReport?: DQReport | null;
  status: "PROSES" | "SELESAI" | "ARSIP" | "DIHAPUS";
  totalBaris: number;
  barisValid: number;
  barisGagal: number;
  barisBaru: number;
  barisBerubah: number;
  barisUnchanged: number;
  overrideUsed: boolean;
  alasanOverride?: string | null;
  createdAt: string;
}

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

export interface QueryDok {
  id: string;
  hospitalId: string;
  sep: string;
  nomorMr: string;
  inisial: string;
  tglMasuk: string | null;
  tglKeluar: string | null;
  dpjpId: string;
  dpjpNama: string;
  dpjpSpesialisasi: string | null;
  templateNama: string | null;
  pertanyaan: string;
  konteks: string;
  status: "DRAFT" | "TERKIRIM" | "DIJAWAB" | "TERLAMBAT";
  deadline: string;
  sentAt: string | null;
  answeredAt: string | null;
  jawaban: string | null;
  slaHari: number;
  createdAt: string;
}

export interface TemplateDok {
  id: string;
  nama: string;
  isi: string;
  kategori: string;
  aktif: boolean;
}

export interface UserPub {
  id: string;
  email: string;
  nama: string;
  role: Peran;
  hospitalId: string | null;
  spesialisasi: string | null;
  dpjpKode: string | null;
  aktif: boolean;
  createdAt: string;
  lastLoginAt: string | null;
}

export interface Audit {
  id: string;
  userEmail: string;
  aksi: string;
  entitas: string;
  entitasId: string | null;
  meta: string | null;
  waktu: string;
}

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

export interface Bootstrap {
  me: Sesi;
  hospitals: Hospital[];
  uploads: Upload[];
  templates: TemplateDok[];
  settings: { key: string; value: string }[];
  dokter: UserPub[];
  users?: UserPub[];
}

export const LABEL_KELAS_RUGI: Record<string, string> = {
  SURPLUS: "Surplus",
  S1_CODING: "S1 — Coding Loss",
  S2_MIXED: "S2 — Mixed Loss",
  STRUCTURAL: "Structural Loss",
  RAHASIA: "—",
};

export const LABEL_STATUS_LEDGER: Record<string, string> = {
  KANDIDAT: "Kandidat",
  DIPERIKSA: "Diperiksa",
  DIKOREKSI: "Dikoreksi",
  DIAJUKAN: "Diajukan",
  HASIL: "Hasil",
  TIDAK_BERUBAH: "Tidak Berubah",
};

export const LABEL_STATUS_QUERY: Record<string, string> = {
  DRAFT: "Draf",
  TERKIRIM: "Terkirim",
  DIJAWAB: "Dijawab",
  TERLAMBAT: "Terlambat",
};

export const LABEL_PERAN: Record<string, string> = {
  SUPERADMIN: "Super Admin",
  CASEMIX: "Casemix Manager",
  CODER: "Medical Coder",
  DIREKTUR: "Direktur RS",
  DPJP: "DPJP",
};

export const LABEL_KOMPONEN: Record<string, string> = {
  VISITATION_FEE: "Visitation Fee",
  ACCOMODATION_FEE: "Akomodasi",
  ICU_FEE: "ICU",
  ICU_SUPPORT_FEE: "ICU Support",
  NURSE_FEE: "Keperawatan",
  TREATMENT_FEE: "Tindakan",
  DRUG_FEE: "Obat",
  DRUG_KRONIS_FEE: "Obat Kronis (fee)",
  DRUG_KEMO_FEE: "Obat Kemo (fee)",
  OBAT_KRONIS: "Obat Kronis",
  OBAT_KEMO: "Obat Kemo",
  ALKES_BHP: "Alkes & BHP",
  LAB: "Laboratorium",
  RAD: "Radiologi",
  REHAB_MEDIK: "Rehab Medik",
  KAMAR_AKOMODASI: "Kamar & Akomodasi",
  OPERASIONAL: "Operasional",
  DARAH: "Darah",
};
