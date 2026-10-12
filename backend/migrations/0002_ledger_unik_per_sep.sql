-- Satu SEP hanya boleh punya satu entri ledger per rumah sakit.
-- Sebelumnya hanya dijaga di kode aplikasi; indeks ini menutup race condition
-- saat dua request membuat kandidat untuk SEP yang sama bersamaan.
CREATE UNIQUE INDEX ledger_entries_hospital_sep_uq
    ON ledger_entries (hospital_id, sep);
