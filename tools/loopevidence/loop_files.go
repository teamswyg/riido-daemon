package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func expandLoopFiles(root string, m manifest) (manifest, error) {
	for _, path := range m.LoopFiles {
		item, err := loadLoopFile(resolvePath(root, path))
		if err != nil {
			return manifest{}, fmt.Errorf("load loop file %q: %w", path, err)
		}
		m.Loops = append(m.Loops, item)
	}
	return m, nil
}

func loadLoopFile(path string) (loop, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return loop{}, err
	}
	var out loop
	if err := json.Unmarshal(data, &out); err != nil {
		return loop{}, err
	}
	return out, nil
}
