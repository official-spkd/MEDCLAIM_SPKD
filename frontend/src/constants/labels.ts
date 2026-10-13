// MedClaim — label tampilan untuk kode enum dari backend

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
