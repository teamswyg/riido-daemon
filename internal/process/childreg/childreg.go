// Package childreg persists the process-group IDs of provider CLI children the
// daemon spawns, so a later daemon start can reap groups orphaned by a previous
// unclean exit (D6 of the runtime lifecycle review).
//
// Each provider process is started with Setpgid (see internal/process/
// processexec), so its PID equals its process-group ID. On a graceful session
// end the group is killed AND removed from this registry, so the file only ever
// holds groups that are currently live. After a daemon SIGKILL/crash the file
// survives with the live groups; the next daemon start calls ReapOrphans to kill
// those leftover groups before serving.
package childreg

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Registry tracks the live provider process-group IDs and mirrors them to a file.
// It satisfies processexec.ChildObserver (OnSpawn/OnExit) without importing it.
type Registry struct {
	path string
	mu   sync.Mutex
	live map[int]struct{}
}

// New returns a Registry that persists to path. A nil/blank path disables
// persistence (OnSpawn/OnExit become no-ops), which keeps tests and non-daemon
// callers working without a backing file.
func New(path string) *Registry {
	return &Registry{path: strings.TrimSpace(path), live: map[int]struct{}{}}
}

// OnSpawn records a newly spawned child's process-group id (== child pid).
func (r *Registry) OnSpawn(pid int) {
	if r == nil || pid <= 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.live[pid] = struct{}{}
	_ = r.persistLocked()
}

// OnExit drops a child whose process has been reaped.
func (r *Registry) OnExit(pid int) {
	if r == nil || pid <= 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.live, pid)
	_ = r.persistLocked()
}

func (r *Registry) persistLocked() error {
	if r.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}
	pids := make([]int, 0, len(r.live))
	for pid := range r.live {
		pids = append(pids, pid)
	}
	sort.Ints(pids)
	var b strings.Builder
	for _, pid := range pids {
		b.WriteString(strconv.Itoa(pid))
		b.WriteByte('\n')
	}
	tmp := r.path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, r.path)
}

// ReapOrphans reads the registry written by a previous daemon, kills each
// still-live process group, resets the file, and returns the number reaped.
// A missing file means a clean previous shutdown (nothing to reap).
func ReapOrphans(path string) (int, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return 0, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	reaped := 0
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil || pid <= 0 {
			continue
		}
		if reapProcessGroup(pid) {
			reaped++
		}
	}
	// Reset: these groups have been handled (killed or already gone).
	_ = os.WriteFile(path, nil, 0o644)
	return reaped, nil
}
