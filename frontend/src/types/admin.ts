// MedClaim — tipe domain: administrasi (cermin respons backend Go)

import type { Peran } from "./auth";

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

// Body POST /api/admin/users
export interface BuatUserBody {
  email: string;
  nama: string;
  role: string;
  hospitalId: string | null;
  spesialisasi: string | null;
  dpjpKode: string | null;
  password?: string;
}

// Respons pembuatan akun; passwordSementara hanya ada bila password tidak diisi.
export interface UserDibuat extends UserPub {
  passwordSementara?: string;
}
