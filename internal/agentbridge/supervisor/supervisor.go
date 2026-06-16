// Package supervisor implements the Daemon tier of the
// Daemon -> Runtime -> Agent hierarchy.
//
// The supervisor owns the control-plane loop: register runtimes,
// heartbeat, claim tasks, submit them to the selected RuntimeActor, and report
// event/result streams back through TaskReporterPort. Its mutable state
// is owned by one goroutine; helper goroutines only translate external
// channels into mailbox messages.
package supervisor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/ir/ingest"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/workdir"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

var ErrStopped = errors.New("supervisor: stopped")

const (
	// DefaultMailboxSize is the supervisor actor mailbox size fixed by
	// docs/20-domain/provider-runtime.md §7.5.
	DefaultMailboxSize = 64
)

const (
	MetadataWorkspaceID              = string(metadatakeys.WorkspaceID)
	MetadataWorkspace                = string(metadatakeys.Workspace)
	MetadataRunID                    = string(metadatakeys.RunID)
	MetadataAgentName                = string(metadatakeys.AgentName)
	MetadataAgentIdentity            = string(metadatakeys.AgentIdentity)
	MetadataWorkflow                 = string(metadatakeys.Workflow)
	MetadataWorkdirRoot              = string(metadatakeys.WorkdirRoot)
	MetadataWorkdir                  = string(metadatakeys.Workdir)
	MetadataOutputDir                = string(metadatakeys.OutputDir)
	MetadataLogsDir                  = string(metadatakeys.LogsDir)
	MetadataArtifactsDir             = string(metadatakeys.ArtifactsDir)
	MetadataNativeConfig             = string(metadatakeys.NativeConfigDir)
	MetadataNativeConfigHome         = string(metadatakeys.NativeConfigHome)
	MetadataIRDir                    = string(metadatakeys.IRDir)
	MetadataNativeConfigVersion      = string(metadatakeys.NativeConfigVersion)
	MetadataRequiredSurfaces         = string(metadatakeys.RequiredSurfaces)
	MetadataAllowExperimentalRuntime = string(metadatakeys.AllowExperimentalRuntime)
)

type Config struct {
	DaemonID string
	// RiidoDaemonVersion is the A-axis daemon binary version stamped on
	// CanonicalEvent common envelopes.
	RiidoDaemonVersion string
	// Runtime is the legacy single-runtime path used by tests and older
	// embedders. New daemon wiring should pass Runtimes, one RuntimeActor per
	// provider capability boundary.
	Runtime *runtimeactor.Actor
	// Runtimes is the provider-runtime pool the supervisor dispatches over.
	Runtimes []*runtimeactor.Actor
	Source   controlplane.TaskSourcePort
	Reporter controlplane.TaskReporterPort
	Workdir  workdir.Adapter

	PollEvery           time.Duration
	IdlePollEvery       time.Duration
	HeartbeatEvery      time.Duration
	MailboxSize         int
	PolicyBundleVersion string
	PolicyBundle        policy.PolicyBundle
	RuntimeTrustTier    policy.TrustTier
}

type Actor struct {
	cfg Config

	mailbox   chan envelope
	stopReqCh chan lifecycle.ShutdownLevel
	stoppedCh chan struct{}
	stopErrCh chan error

	claimMu     sync.Mutex
	claimCancel context.CancelFunc
}

type envelope struct {
	taskEvent  *taskEventMsg
	taskResult *taskResultMsg
	cancel     *cancelMsg
}

type taskEventMsg struct {
	taskID string
	event  agentbridge.Event
}

type taskResultMsg struct {
	taskID string
	result agentbridge.Result
}

type cancelMsg struct {
	taskID string
	cause  error
}

type runningTask struct {
	taskID  string
	report  controlplane.TaskReportContext
	runtime *runtimeactor.Actor
	handle  *runtimeactor.SessionHandle

	workspace *workdir.Workspace
	events    *workspaceEventContext
}

type preparedWorkspace struct {
	workspace *workdir.Workspace
	events    *workspaceEventContext
}

type workspaceEventContext struct {
	taskID              string
	runID               string
	runtimeID           string
	capability          runtimeactor.Capability
	nativeConfigVersion string
	ingestor            *ingest.Ingestor
	agentIngestor       *ingest.Ingestor
}

func New(cfg Config) (*Actor, error) {
	if cfg.DaemonID == "" {
		return nil, errors.New("supervisor: DaemonID is required")
	}
	if len(configuredRuntimes(cfg)) == 0 {
		return nil, errors.New("supervisor: at least one Runtime is required")
	}
	if cfg.Source == nil {
		return nil, errors.New("supervisor: Source is required")
	}
	if cfg.Reporter == nil {
		return nil, errors.New("supervisor: Reporter is required")
	}
	if cfg.PollEvery <= 0 {
		cfg.PollEvery = time.Second
	}
	if cfg.IdlePollEvery <= 0 {
		cfg.IdlePollEvery = cfg.PollEvery
	}
	if cfg.IdlePollEvery < cfg.PollEvery {
		cfg.IdlePollEvery = cfg.PollEvery
	}
	if cfg.HeartbeatEvery <= 0 {
		cfg.HeartbeatEvery = 5 * time.Second
	}
	if cfg.MailboxSize <= 0 {
		cfg.MailboxSize = DefaultMailboxSize
	}
	if cfg.RiidoDaemonVersion == "" {
		cfg.RiidoDaemonVersion = "riido-agentd v0.0.0"
	}
	if cfg.PolicyBundleVersion == "" {
		cfg.PolicyBundleVersion = cfg.PolicyBundle.Version
		if cfg.PolicyBundleVersion == "" {
			cfg.PolicyBundleVersion = policy.DefaultLocalPolicyBundleVersion
		}
	}
	if cfg.PolicyBundle.SchemaVersion == "" {
		cfg.PolicyBundle = policy.DefaultLocalPolicyBundle()
		cfg.PolicyBundle.Version = cfg.PolicyBundleVersion
	} else {
		if err := cfg.PolicyBundle.Validate(); err != nil {
			return nil, fmt.Errorf("supervisor: policy bundle: %w", err)
		}
		if cfg.PolicyBundleVersion != cfg.PolicyBundle.Version {
			return nil, fmt.Errorf("supervisor: PolicyBundleVersion %q does not match policy bundle version %q", cfg.PolicyBundleVersion, cfg.PolicyBundle.Version)
		}
	}
	if cfg.RuntimeTrustTier == "" {
		cfg.RuntimeTrustTier = policy.TrustTierHost
	}
	return &Actor{
		cfg:       cfg,
		mailbox:   make(chan envelope, cfg.MailboxSize),
		stopReqCh: make(chan lifecycle.ShutdownLevel, cfg.MailboxSize),
		stoppedCh: make(chan struct{}),
		stopErrCh: make(chan error, 1),
	}, nil
}
