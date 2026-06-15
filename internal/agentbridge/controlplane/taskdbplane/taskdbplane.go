// Package taskdbplane adapts riido-task-db.v1 into the agentbridge
// control-plane ports.
//
// It is intentionally outside the core controlplane package: the
// port definitions stay independent from project persistence, while
// this adapter is allowed to translate taskdb.TaskRecord rows into
// bridge.TaskRequest values and report guarded TaskState transitions.
package taskdbplane

import (
	"context"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

const (
	RuntimeRegistrySchemaVersion      = "riido-runtime-registry.v1"
	RuntimeLeaseRegistrySchemaVersion = "riido-runtime-lease-registry.v1"

	sourceName         = "riido.agentbridge.taskdb"
	metadataTaskDB     = "task_db_path"
	metadataDocument   = "source_document_path"
	commandIDPrefix    = "command:riido.agentbridge.taskdb:"
	defaultActor       = "daemon"
	defaultClaimReason = "runtime claimed queued task DB row"

	defaultRuntimeLeaseTTL = 30 * time.Second
)

// RuntimeRegistry is the task DB source sidecar written next to the
// riido-task-db.v1 file. It lets local GUI/Zed integrations inspect
// runtime registration and heartbeat without reaching into daemon memory.
type RuntimeRegistry struct {
	SchemaVersion string                           `json:"schema_version"`
	TaskDBPath    string                           `json:"task_db_path"`
	UpdatedAt     time.Time                        `json:"updated_at"`
	Runtimes      []controlplane.RegisteredRuntime `json:"runtimes"`
}

// RuntimeLeaseRegistry is the task DB source sidecar that records the
// latest local C9 fencing token per task.
type RuntimeLeaseRegistry struct {
	SchemaVersion string               `json:"schema_version"`
	TaskDBPath    string               `json:"task_db_path"`
	UpdatedAt     time.Time            `json:"updated_at"`
	Leases        []RuntimeLeaseRecord `json:"leases"`
}

type RuntimeLeaseRecord struct {
	LeaseID               string     `json:"lease_id"`
	TaskID                string     `json:"task_id"`
	RuntimeID             string     `json:"runtime_id"`
	CapabilityFingerprint string     `json:"capability_fingerprint,omitempty"`
	ClaimedAt             time.Time  `json:"claimed_at"`
	LeaseUntil            time.Time  `json:"lease_until"`
	FencingToken          int64      `json:"fencing_token"`
	ReleasedAt            *time.Time `json:"released_at,omitempty"`
}

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

func (p *Plane) RegisterRuntime(ctx context.Context, rt controlplane.RuntimeRegistration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if rt.RuntimeID == "" {
		return planeErrorf(ErrTaskDBPlaneRuntime, "register-runtime", "empty RuntimeID")
	}
	return p.withFileLock(ctx, func() error {
		if err := p.reloadRuntimeRegistry(); err != nil {
			return err
		}
		p.runtimes[rt.RuntimeID] = controlplane.RegisteredRuntime{
			RuntimeRegistration: rt,
			LastHeartbeat:       p.now().UTC(),
		}
		return p.saveRuntimeRegistry()
	})
}

func (p *Plane) DeregisterRuntime(ctx context.Context, runtimeID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if runtimeID == "" {
		return planeErrorf(ErrTaskDBPlaneRuntime, "deregister-runtime", "empty RuntimeID")
	}
	return p.withFileLock(ctx, func() error {
		if err := p.reloadRuntimeRegistry(); err != nil {
			return err
		}
		delete(p.runtimes, runtimeID)
		return p.saveRuntimeRegistry()
	})
}

func (p *Plane) Heartbeat(ctx context.Context, hb controlplane.RuntimeHeartbeat) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return p.withFileLock(ctx, func() error {
		if err := p.reloadRuntimeRegistry(); err != nil {
			return err
		}
		rec, ok := p.runtimes[hb.RuntimeID]
		if !ok {
			return planeErrorf(ErrTaskDBPlaneRuntime, "heartbeat", "heartbeat for unknown runtime %q", hb.RuntimeID)
		}
		rec.LastHeartbeat = p.now().UTC()
		applyHeartbeat(&rec.RuntimeRegistration, hb)
		p.runtimes[hb.RuntimeID] = rec
		if err := p.saveRuntimeRegistry(); err != nil {
			return err
		}
		leases, err := loadRuntimeLeaseRegistryOrEmpty(p.leasePath)
		if err != nil {
			return err
		}
		leases, changed := refreshRuntimeLeases(leases, rec, hb.RunningTaskIDs, rec.LastHeartbeat, p.leaseTTL)
		if !changed {
			return nil
		}
		return saveRuntimeLeaseRegistry(p.leasePath, p.path, leases, rec.LastHeartbeat)
	})
}
