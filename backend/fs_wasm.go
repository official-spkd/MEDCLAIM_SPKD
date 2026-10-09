//go:build js && wasm

package main

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
)

//go:embed embed/medclaim.json embed/samples/*.txt
var aset embed.FS

func bacaFile(p string) ([]byte, error) {
	if strings.HasSuffix(p, "medclaim.json") {
		return aset.ReadFile("embed/medclaim.json")
	}
	return aset.ReadFile("embed/samples/" + filepath.Base(p))
}

func bacaDir(p string) ([]os.DirEntry, error) { return aset.ReadDir("embed/samples") }
