package types

import "regexp"

var reUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID memeriksa format sebelum nilai dikirim ke kolom uuid; tanpa ini,
// input seperti "abc" membuat PostgreSQL menolak query (error 500, bukan 404).
func IsUUID(s string) bool { return reUUID.MatchString(s) }
