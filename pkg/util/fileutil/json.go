package fileutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteJSONAtomic writes indented JSON and atomically replaces the target path.
// It creates the parent directory, writes a trailing newline, and removes the
// temporary file on failure.
func WriteJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return WriteFileAtomic(path, data, ".tmp-"+filepath.Base(path)+"-*")
}

// WriteFileAtomic writes content to a temp file in the target directory and
// renames it over the target path.
func WriteFileAtomic(path string, content []byte, tempPattern string) error {
	dir := filepath.Dir(path)
	if tempPattern == "" {
		tempPattern = ".tmp-" + filepath.Base(path) + "-*"
	}
	tmp, err := os.CreateTemp(dir, tempPattern)
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
