package workdir

import "github.com/teamswyg/riido-daemon/pkg/util/fileutil"

func writeJSONAtomic(path string, value any) error {
	return fileutil.WriteJSONAtomic(path, value)
}
