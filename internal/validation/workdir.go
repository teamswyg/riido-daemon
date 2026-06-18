package validation

import (
	"fmt"
	"os"
	"strings"
)

func resolveValidationWorkdir(workdir string) (string, error) {
	workdir = strings.TrimSpace(workdir)
	if workdir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve validation workdir: %w", err)
		}
		workdir = cwd
	}
	if info, err := os.Stat(workdir); err != nil {
		return "", fmt.Errorf("stat validation workdir: %w", err)
	} else if !info.IsDir() {
		return "", fmt.Errorf("validation workdir is not a directory: %s", workdir)
	}
	return workdir, nil
}
