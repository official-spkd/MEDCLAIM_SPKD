// MedClaim — tipe domain: bootstrap (cermin respons backend Go)

import type { Sesi } from "./auth";
import type { Hospital } from "./hospital";
import type { Upload } from "./upload";
import type { TemplateDok } from "./query";
import type { UserPub } from "./admin";

export interface Bootstrap {
  me: Sesi;
  hospitals: Hospital[];
  uploads: Upload[];
  templates: TemplateDok[];
  settings: { key: string; value: string }[];
  dokter: UserPub[];
  users?: UserPub[];
}
