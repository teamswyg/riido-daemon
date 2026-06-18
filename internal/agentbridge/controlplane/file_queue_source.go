package controlplane

import (
	"os"
	"time"
)

// FileQueueSource reads JSON-encoded TaskRequest files from a directory
// and writes runtime registry/heartbeat records under dir/runtimes/.
// Each successful ClaimTask atomically moves the top-level task file
// into dir/claims/ and replaces it with a claim receipt, so the same
// task is not replayed even if multiple daemon processes poll the same
// local queue. Useful for batch testing and for ad-hoc CLI-driven queues.
type FileQueueSource struct {
	dir string
	now func() time.Time
}

func NewFileQueueSource(dir string) (*FileQueueSource, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, controlPlaneWrapf(ErrControlPlaneQueue, "file-queue.new", err, "stat queue dir")
	}
	if !info.IsDir() {
		return nil, controlPlaneErrorf(ErrControlPlaneQueue, "file-queue.new", "%s is not a directory", dir)
	}
	return &FileQueueSource{dir: dir, now: time.Now}, nil
}
