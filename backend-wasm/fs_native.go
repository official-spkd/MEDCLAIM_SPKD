//go:build !js

package main

import "os"

func bacaFile(p string) ([]byte, error)     { return os.ReadFile(p) }
func bacaDir(p string) ([]os.DirEntry, error) { return os.ReadDir(p) }
