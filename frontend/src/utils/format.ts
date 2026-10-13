// ============================================================
// MedClaim Vue — Format angka, rupiah, tanggal, persen (id-ID)
// ============================================================

export function rupiah(n: number | null | undefined): string {
  if (n === null || n === undefined) return "—";
  const neg = n < 0;
  const s = Math.abs(Math.round(n)).toLocaleString("id-ID");
  return `${neg ? "-" : ""}Rp ${s}`;
}

export function rupiahSingkat(n: number | null | undefined): string {
  if (n === null || n === undefined) return "—";
  const neg = n < 0;
  const a = Math.abs(n);
  let out: string;
  if (a >= 1_000_000_000) out = `${(a / 1_000_000_000).toLocaleString("id-ID", { maximumFractionDigits: 2 })} M`;
  else if (a >= 1_000_000) out = `${(a / 1_000_000).toLocaleString("id-ID", { maximumFractionDigits: 1 })} jt`;
  else if (a >= 1_000) out = `${(a / 1_000).toLocaleString("id-ID", { maximumFractionDigits: 0 })} rb`;
  else out = a.toLocaleString("id-ID");
  return `${neg ? "-" : ""}Rp ${out}`;
}

export function angka(n: number | null | undefined): string {
  if (n === null || n === undefined) return "—";
  return n.toLocaleString("id-ID");
}

export function persen(f: number | null | undefined, desimal = 1): string {
  if (f === null || f === undefined) return "—";
  return `${(f * 100).toLocaleString("id-ID", { maximumFractionDigits: desimal })}%`;
}

export function tanggal(iso: string | null | undefined): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  return d.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
}

export function tanggalWaktu(iso: string | null | undefined): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  return d.toLocaleString("id-ID", {
    day: "2-digit", month: "short", year: "numeric",
    hour: "2-digit", minute: "2-digit",
  });
}

export function usia(tglLahir: string, tglAcuan?: string): number {
  const l = new Date(tglLahir);
  const a = tglAcuan ? new Date(tglAcuan) : new Date();
  if (isNaN(l.getTime()) || isNaN(a.getTime())) return 0;
  let usia = a.getFullYear() - l.getFullYear();
  const m = a.getMonth() - l.getMonth();
  if (m < 0 || (m === 0 && a.getDate() < l.getDate())) usia--;
  return usia;
}

export function inisialNama(nama: string): string {
  const bagian = nama.trim().split(/\s+/);
  if (bagian.length === 1) return bagian[0].slice(0, 2).toUpperCase();
  return (bagian[0][0] + bagian[1][0]).toUpperCase();
}
