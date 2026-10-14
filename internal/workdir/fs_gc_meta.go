package workdir

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func writeGCMeta(ws Workspace, id TaskID, runID string) error {
	meta := map[string]any{
		metadatakeys.WorkspaceID.String(): id.Workspace,
		metadatakeys.TaskID.String():      id.Task,
		metadatakeys.RunID.String():       runID,
		"created_at":                      time.Now().UTC().Format(time.RFC3339Nano),
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("workdir: marshal gc meta: %w", err)
	}
	if err := os.WriteFile(filepath.Join(ws.Root, ".gc_meta.json"), metaBytes, 0o644); err != nil {
		return fmt.Errorf("workdir: write gc meta: %w", err)
	}
	return nil
}
