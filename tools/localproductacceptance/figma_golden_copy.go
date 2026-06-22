package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func copyFigmaGolden(id string, golden figmaGoldenScreen, screenshotDir string) (string, error) {
	if screenshotDir == "" {
		return golden.ResolvedPath, nil
	}
	if err := os.MkdirAll(screenshotDir, 0o755); err != nil {
		return "", fmt.Errorf("create figma screenshot dir: %w", err)
	}
	target := filepath.Join(screenshotDir, strings.ReplaceAll(id, ".", "-")+".png")
	data, err := os.ReadFile(golden.ResolvedPath)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		return "", fmt.Errorf("write figma screenshot copy: %w", err)
	}
	return target, nil
}

func entryNodeMatches(entries []figmaIntentEntry, nodeID string) bool {
	for _, entry := range entries {
		if entry.NodeID == nodeID {
			return true
		}
	}
	return false
}

func figmaGoldenObserved(golden figmaGoldenScreen) map[string]any {
	return map[string]any{
		"node_id":         golden.NodeID,
		"name":            golden.Name,
		"width":           golden.Width,
		"height":          golden.Height,
		"original_width":  golden.OriginalWidth,
		"original_height": golden.OriginalHeight,
		"sha256":          golden.SHA256,
	}
}
