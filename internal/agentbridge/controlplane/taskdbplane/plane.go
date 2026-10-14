package taskdbplane

import (
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

// Plane implements both TaskSourcePort and TaskReporterPort over one
// riido-task-db.v1 JSON file. The supervisor actor owns this value and
// calls it serially; the adapter therefore uses no mutex.
type Plane struct {
	path         string
	registryPath string
	leasePath    string
	lockPath     string
	leaseTTL     time.Duration
	now          func() time.Time
	runtimes     map[string]controlplane.RegisteredRuntime
}

func New(path string) (*Plane, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, planeErrorf(ErrTaskDBPlaneInput, "new", "empty task DB path")
	}
	if _, err := taskdb.LoadTaskDBOrEmpty(path); err != nil {
		return nil, planeWrapf(ErrTaskDBPlanePersistence, "new.load-task-db", err, "load task DB")
	}
	registryPath := runtimeRegistryPath(path)
	leasePath := runtimeLeaseRegistryPath(path)
	runtimes, err := loadRuntimeRegistryOrEmpty(registryPath)
	if err != nil {
		return nil, err
	}
	return &Plane{
		path:         path,
		registryPath: registryPath,
		leasePath:    leasePath,
		lockPath:     path + ".lock",
		leaseTTL:     defaultRuntimeLeaseTTL,
		now:          time.Now,
		runtimes:     runtimes,
	}, nil
}
