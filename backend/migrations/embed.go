// Package migrations meng-embed file SQL migrasi supaya ikut ter-compile ke binary.
// Tambahkan migrasi baru cukup dengan menaruh file NNNN_nama.sql di folder ini.
package migrations

import "embed"

//go:embed *.sql
var Files embed.FS
