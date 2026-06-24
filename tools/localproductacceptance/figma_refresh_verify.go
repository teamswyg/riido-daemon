package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func figmaRefreshProofAt(observed time.Time, validFor time.Duration, manifestPath, goldenPath string) figmaRefreshProof {
	proof := figmaRefreshProof{ManifestPath: manifestPath, GoldenPath: goldenPath, ValidSeconds: int64(validFor.Seconds())}
	entries, err := loadFigmaIntentEntries(manifestPath)
	if err != nil {
		proof.Err = err.Error()
		return proof
	}
	proof.EntryCount = len(entries)
	catalog, err := loadFigmaGoldenCatalogRaw(goldenPath)
	if err != nil {
		proof.Err = err.Error()
		return proof
	}
	proof.ScreenCount = len(catalog.Screens)
	proof.CapturedAt = catalog.CapturedAt
	capturedAt, err := time.Parse(time.RFC3339, catalog.CapturedAt)
	if err != nil {
		proof.Err = fmt.Sprintf("parse figma golden captured_at: %v", err)
		return proof
	}
	proof.AgeSeconds = int64(observed.Sub(capturedAt).Seconds())
	proof.Stale = observed.After(capturedAt.Add(validFor))
	return proof
}

func loadFigmaGoldenCatalogRaw(path string) (figmaGoldenCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return figmaGoldenCatalog{}, fmt.Errorf("read figma golden manifest: %w", err)
	}
	var catalog figmaGoldenCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return figmaGoldenCatalog{}, fmt.Errorf("decode figma golden manifest: %w", err)
	}
	if len(catalog.Screens) == 0 {
		return figmaGoldenCatalog{}, fmt.Errorf("figma golden manifest has no screens")
	}
	return catalog, nil
}
