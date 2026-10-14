package main

import (
	"os"
	"path/filepath"
	"strings"
)

const manifestSampleLimit = 3

func scanManifestInventory(root string) (manifestInventory, error) {
	inventory := manifestInventory{}
	samples := map[string][]string{}
	counts := map[string]int{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".riido.json") {
			return nil
		}
		group := manifestGroup(root, path)
		inventory.Count++
		counts[group]++
		if len(samples[group]) < manifestSampleLimit {
			samples[group] = append(samples[group], slashPath(root, path))
		}
		return nil
	})
	inventory.Groups = manifestGroups(counts)
	inventory.Samples = manifestSamples(inventory.Groups, samples)
	return inventory, err
}

func manifestGroup(root, path string) string {
	parts := strings.Split(slashPath(root, path), "/")
	if len(parts) == 1 {
		return "."
	}
	return parts[0]
}
