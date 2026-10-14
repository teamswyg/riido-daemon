package supervisor

import (
	"sync"

	"github.com/teamswyg/riido-daemon/internal/workdir"
)

type countingWorkdir struct {
	inner  workdir.Adapter
	mu     sync.Mutex
	last   workdir.Workspace
	inject int
}

func newCountingWorkdir(root string) *countingWorkdir {
	return &countingWorkdir{inner: workdir.NewFSAdapter(root)}
}

func (w *countingWorkdir) Prepare(id workdir.TaskID) (workdir.Workspace, error) {
	ws, err := w.inner.Prepare(id)
	w.mu.Lock()
	w.last = ws
	w.mu.Unlock()
	return ws, err
}

func (w *countingWorkdir) InjectRuntimeConfig(
	ws workdir.Workspace,
	cfg workdir.RuntimeConfig,
) error {
	w.mu.Lock()
	w.inject++
	w.last = ws
	w.mu.Unlock()
	return w.inner.InjectRuntimeConfig(ws, cfg)
}

func (w *countingWorkdir) snapshot() (workdir.Workspace, int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.last, w.inject
}
