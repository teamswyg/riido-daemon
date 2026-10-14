package main

import (
	"crypto/sha256"
	"fmt"
	"image/png"
	"os"
)

func verifyFigmaGolden(
	id string,
	matches []figmaIntentEntry,
	golden figmaGoldenScreen,
	screenshotDir string,
) (string, map[string]any, error) {
	if golden.ScenarioID == "" {
		return "", nil, fmt.Errorf("figma golden missing for %s", id)
	}
	if !entryNodeMatches(matches, golden.NodeID) {
		return "", nil, fmt.Errorf("figma golden node %s is not present in intent entries", golden.NodeID)
	}
	if err := verifyGoldenFile(golden); err != nil {
		return "", nil, err
	}
	screenshot, err := copyFigmaGolden(id, golden, screenshotDir)
	if err != nil {
		return "", nil, err
	}
	observed := figmaObserved(matches)
	observed["golden"] = figmaGoldenObserved(golden)
	return screenshot, observed, nil
}

func verifyGoldenFile(golden figmaGoldenScreen) error {
	width, height, err := pngSize(golden.ResolvedPath)
	if err != nil {
		return err
	}
	if width != golden.Width || height != golden.Height {
		return fmt.Errorf("figma golden size = %dx%d, want %dx%d", width, height, golden.Width, golden.Height)
	}
	hash, err := fileSHA256(golden.ResolvedPath)
	if err != nil {
		return err
	}
	if hash != golden.SHA256 {
		return fmt.Errorf("figma golden sha256 = %s, want %s", hash, golden.SHA256)
	}
	return nil
}

func pngSize(path string) (int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	cfg, err := png.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}
