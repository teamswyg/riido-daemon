package main

import (
	"os"
	"path/filepath"
)

func writeExecutable(dir, name string) (string, error) {
	path := filepath.Join(dir, name)
	body := []byte("#!/bin/sh\nprintf ok\n")
	if err := os.WriteFile(path, body, 0o755); err != nil {
		return "", err
	}
	return filepath.Clean(path), nil
}

func withPATH(value string, fn func() error) error {
	old, had := os.LookupEnv("PATH")
	if err := os.Setenv("PATH", value); err != nil {
		return err
	}
	defer restorePATH(old, had)
	return fn()
}

func restorePATH(old string, had bool) {
	if had {
		_ = os.Setenv("PATH", old)
		return
	}
	_ = os.Unsetenv("PATH")
}
