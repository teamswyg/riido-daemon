package hostintegration

import (
	"path/filepath"
	"strings"
)

func joinHostPath(os HostOS, elems ...string) string {
	if os != HostOSWindows {
		return filepath.Join(elems...)
	}
	return joinWindowsPath(elems...)
}

func joinWindowsPath(elems ...string) string {
	var parts []string
	for i, elem := range elems {
		trimmed := strings.TrimSpace(elem)
		if trimmed == "" {
			continue
		}
		if i == 0 {
			parts = append(parts, strings.TrimRight(trimmed, `\/`))
			continue
		}
		parts = append(parts, strings.Trim(trimmed, `\/`))
	}
	return strings.Join(parts, `\`)
}
