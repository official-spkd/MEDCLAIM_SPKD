// MedClaim — tipe domain: query dokumen (cermin respons backend Go)

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

// Body POST /api/queries
export interface BuatQueryBody {
  sep: string;
  dpjpId: string;
  templateId?: string;
  pertanyaan?: string;
  konteks?: string;
}
