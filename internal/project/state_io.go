package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
)

func DefaultStatePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "riido", "workspace-state.json"), nil
}

func SaveState(path string, state StateFile) error {
	if path == "" {
		return fmt.Errorf("state path is empty")
	}
	if err := fileutil.WriteJSONAtomic(path, state); err != nil {
		return fmt.Errorf("save state file: %w", err)
	}
	return nil
}

func LoadState(path string) (StateFile, error) {
	var state StateFile
	data, err := os.ReadFile(path)
	if err != nil {
		return state, err
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("decode state file: %w", err)
	}
	if state.SchemaVersion != StateSchemaVersion {
		return state, fmt.Errorf("state schema mismatch: got %q want %q", state.SchemaVersion, StateSchemaVersion)
	}
	return state, nil
}
