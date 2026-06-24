package main

import (
	"path/filepath"
	"testing"
)

func mustRegistry(t *testing.T, root string) registry {
	t.Helper()
	reg, err := loadRegistry(filepath.Join(root, filepath.FromSlash(defaultManifest)))
	if err != nil {
		t.Fatal(err)
	}
	return reg
}
