package main

import (
	"fmt"
	"os"
	"strings"
)

func maybeDoc(root, rel, body string, write, check bool) error {
	path := repoPath(root, rel)
	if write {
		return os.WriteFile(path, []byte(body), 0o644)
	}
	if !check {
		return nil
	}
	current, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(current)) != strings.TrimSpace(body) {
		return fmt.Errorf("generated doc drift: run go run ./tools/localqacandidatedecision -write-doc")
	}
	return nil
}
