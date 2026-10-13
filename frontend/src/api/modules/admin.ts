import { get, post } from "../client";
import type { UserPub, Audit, BuatUserBody, UserDibuat } from "@/types";

export const adminApi = {
  users: () => get<{ items: UserPub[] }>("/api/admin/users"),
  createUser: (body: BuatUserBody) => post<UserDibuat>("/api/admin/users", body),
  toggleUser: (id: string) => post<UserPub>(`/api/admin/users/${encodeURIComponent(id)}/toggle`),
  audit: () => get<{ items: Audit[] }>("/api/admin/audit"),
};
