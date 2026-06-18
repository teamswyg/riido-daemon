package detectutil

import (
	"os"
	"strings"
)

// LaunchPATH returns the PATH value provider child processes should inherit
// when the caller has not supplied an explicit PATH.
func LaunchPATH() string {
	return strings.Join(augmentedSearchDirs(), string(os.PathListSeparator))
}
