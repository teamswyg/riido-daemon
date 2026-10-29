package agentbridge

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CheckReferenceDirectories proves host listing access, not provider/model use.
func CheckReferenceDirectories(paths []string) error {
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			return fmt.Errorf("reference folder %q: an absolute local path is required", path)
		}
		folder, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("reference folder %q: %w; check that the folder exists and this app has access", path, err)
		}
		_, readErr := folder.Readdirnames(1)
		closeErr := folder.Close()
		if readErr != nil && readErr != io.EOF {
			return fmt.Errorf("reference folder %q cannot be listed: %w; check folder permissions for this app", path, readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("reference folder %q: %w", path, closeErr)
		}
	}
	return nil
}
