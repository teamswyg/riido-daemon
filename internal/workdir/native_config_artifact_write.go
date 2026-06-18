package workdir

import (
	"fmt"
	"io/fs"
)

func writeNativeConfigArtifact(ws Workspace, rel string, content []byte) error {
	return writeNativeConfigArtifactWithMode(ws, rel, content, 0o644)
}

func writeNativeConfigArtifactWithMode(ws Workspace, rel string, content []byte, mode fs.FileMode) error {
	if err := writeFileUnder(ws.Workdir, rel, content, mode); err != nil {
		return err
	}
	if ws.NativeConfig != "" {
		if err := writeFileUnder(ws.NativeConfig, rel, content, mode); err != nil {
			return fmt.Errorf("workdir: write native-config copy %s: %w", rel, err)
		}
	}
	return nil
}
