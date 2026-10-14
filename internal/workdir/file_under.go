package workdir

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func writeFileUnder(root, rel string, content []byte, mode fs.FileMode) error {
	path, err := safeJoin(root, rel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("workdir: mkdir %s: %w", filepath.Dir(path), err)
	}
	if mode == 0 {
		mode = 0o644
	}
	if err := os.WriteFile(path, content, mode); err != nil {
		return fmt.Errorf("workdir: write %s: %w", path, err)
	}
	if mode&0o111 == 0 {
		return nil
	}
	if err := os.Chmod(path, mode); err != nil {
		return fmt.Errorf("workdir: chmod %s: %w", path, err)
	}
	return nil
}
