import { get, post } from "../client";
import type { Sesi } from "@/types";

export const authApi = {
  /** POST /api/auth/login — membuat sesi (cookie). 401 di sini = kredensial salah, bukan sesi berakhir. */
  login: (email: string, password: string) =>
    post<Sesi>("/api/auth/login", { email, password }, { abaikan401: true }),

  /** POST /api/auth/logout — idempoten. */
  logout: () => post<{ ok: boolean }>("/api/auth/logout", undefined, { abaikan401: true }),

  /** GET /api/auth/me */
  me: () => get<Sesi>("/api/auth/me"),
};
