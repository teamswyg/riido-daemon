package workdir

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func injectedFileHashes(root string) ([]nativeConfigFileHash, error) {
	files := []nativeConfigFileHash{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(content)
		files = append(files, nativeConfigFileHash{
			Path:   filepath.ToSlash(rel),
			SHA256: fmt.Sprintf("%x", sum[:]),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("workdir: walk native-config: %w", err)
	}
	sortNativeConfigFiles(files)
	return files, nil
}

func sortNativeConfigFiles(files []nativeConfigFileHash) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}
