import { get, post } from "../client";
import type { UserPub, Audit, BuatUserBody, UserDibuat, UbahUserBody } from "@/types";

export const adminApi = {
  users: () => get<{ items: UserPub[] }>("/api/admin/users"),
  createUser: (body: BuatUserBody) => post<UserDibuat>("/api/admin/users", body),
  toggleUser: (id: string) => post<UserPub>(`/api/admin/users/${encodeURIComponent(id)}/toggle`),
  /** POST /api/admin/users/:id/update — ubah nama/email/password (field kosong tidak diubah). */
  updateUser: (id: string, body: UbahUserBody) =>
    post<UserPub>(`/api/admin/users/${encodeURIComponent(id)}/update`, body),
  audit:() => get<{ items: Audit[] }>("/api/admin/audit"),
};
