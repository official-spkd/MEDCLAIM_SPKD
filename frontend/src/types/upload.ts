// MedClaim — tipe domain: upload & kualitas data (cermin respons backend Go)

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

// Hasil POST /api/uploads/preview
export interface UploadPreview {
  header: string[];
  delimiter: string;
  format: string;
  totalBaris: number;
  barisValid: number;
  barisGagal: number;
  gagalDetail: { nomorBaris: number; alasan: string }[];
  dq: DQReport;
  duplikatSep: string[];
  kolomTersedia: string[];
}

export interface FieldDiff {
  label: string;
  lama: string;
  baru: string;
}

// Hasil POST /api/uploads/commit
export interface UploadHasil {
  upload?: Upload;
  barisBaru: number;
  barisBerubah: number;
  barisUnchanged: number;
  ledgerBaru: number;
  diff: Record<string, FieldDiff[]>;
}

export interface UploadRequest {
  hospitalId: string;
  periode: string;
  namaFile: string;
  isi: string;
  override?: boolean;
  alasanOverride?: string;
}

export interface SampelFile {
  nama: string;
  ukuran: number;
  deskripsi: string;
}
