package workdir

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
)

func claudeSettingsJSON() string {
	return `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PROJECT_DIR}/.riido/hooks/claude-audit-hook.sh",
            "timeout": 30,
            "statusMessage": "Riido audit"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PROJECT_DIR}/.riido/hooks/claude-audit-hook.sh",
            "timeout": 30,
            "statusMessage": "Riido audit"
          }
        ]
      }
    ]
  }
}
`
}

func claudeAuditHookScript() string {
	return `#!/bin/sh
set -eu

project_dir="${CLAUDE_PROJECT_DIR:-$(pwd)}"
event_dir="$project_dir/.riido/hooks"
mkdir -p "$event_dir"
cat >> "$event_dir/claude-hook-events.jsonl"
printf '\n' >> "$event_dir/claude-hook-events.jsonl"
exit 0
`
}

func codexConfigTOML() string {
	return `# Managed by riido-daemon.
# Reserved for future Codex native config materialization.
# Current Codex runs use adapter-owned full-access sandbox selection instead of task-scoped CODEX_HOME.
`
}

func sortedUniquePaths(paths []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		path = filepath.ToSlash(strings.TrimSpace(path))
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

// safePathSegment returns true if s is a non-empty string that does not
// contain path separators or upward traversal sequences. We do NOT
// trust caller-supplied workspace/task/provider names blindly; this
// guards against escapes from the per-task tree into the shared root.
func safePathSegment(s string) bool {
	if s == "" {
		return false
	}
	if strings.ContainsRune(s, os.PathSeparator) {
		return false
	}
	if strings.Contains(s, "..") {
		return false
	}
	return true
}

func localFileURI(path string) string {
	return (&url.URL{Scheme: "file", Path: path}).String()
}

func readArchiveRecord(path string) (ArchiveRecord, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return ArchiveRecord{}, fmt.Errorf("workdir: read archive manifest: %w", err)
	}
	var record ArchiveRecord
	if err := json.Unmarshal(body, &record); err != nil {
		return ArchiveRecord{}, fmt.Errorf("workdir: decode archive manifest: %w", err)
	}
	return record, nil
}

func cleanupEligible(record ArchiveRecord, cutoff time.Time) bool {
	if record.SchemaVersion != ArchiveRecordSchemaVersion {
		return false
	}
	if record.RetentionMode != RetentionModeKeepInPlace {
		return false
	}
	if record.ArchivedAt.IsZero() {
		return false
	}
	return record.ArchivedAt.UTC().Before(cutoff)
}

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

func writeJSONAtomic(path string, value any) error {
	return fileutil.WriteJSONAtomic(path, value)
}
