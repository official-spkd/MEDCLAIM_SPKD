// MedClaim — tipe domain: autentikasi (cermin respons backend Go)

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
