package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteDaemonArchiveReportsCreateFailure(t *testing.T) {
	target := filepath.Join(t.TempDir(), "missing", "riido.tar.gz")
	if err := writeDaemonArchive(target); err == nil {
		t.Fatalf("expected archive create failure")
	}
}

func TestWriteArchiveAndSumsReportsAssetPathConflict(t *testing.T) {
	root := t.TempDir()
	assetDir := filepath.Join(root, "assets")
	if err := os.WriteFile(assetDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeArchiveAndSums(assetDir); err == nil {
		t.Fatalf("expected asset directory conflict")
	}
}

func TestAssertMarkerAbsentRejectsMissingOrPresentMarker(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "install-marker")
	if err := assertMarkerAbsent(marker); err == nil {
		t.Fatalf("expected missing marker failure")
	}
	if err := os.WriteFile(marker, []byte("present\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := assertMarkerAbsent(marker)
	if !errors.Is(err, errOldBinaryPresent) {
		t.Fatalf("expected old binary marker error, got %v", err)
	}
	if err := os.WriteFile(marker, []byte("absent\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := assertMarkerAbsent(marker); err != nil {
		t.Fatalf("expected absent marker to pass: %v", err)
	}
}
