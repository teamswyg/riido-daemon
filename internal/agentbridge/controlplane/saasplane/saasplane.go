package saasplane

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const (
	MetadataAssignmentID    = "riido_saas_assignment_id"
	MetadataAgentID         = "riido_saas_agent_id"
	MetadataComponentID     = "riido_saas_component_id"
	MetadataLeaseToken      = "riido_saas_lease_token"
	MetadataModelID         = "riido_saas_model_id"
	MetadataRuntimeProvider = "riido_saas_runtime_provider"
)

const runtimeSnapshotHeartbeatMinInterval = 4 * time.Second

const agentBindingCacheTTL = time.Second

const (
	jsonRequestMaxAttempts = 3
	jsonRequestRetryBase   = 50 * time.Millisecond
)

const (
	defaultLongPollWait       = 30 * time.Second
	longPollRequestTimeoutPad = 5 * time.Second
)

// Live assistant-body streaming. Raw text deltas are never forwarded one line
// per token (they are tiny, incoherent fragments). Instead the daemon
// accumulates them per task and periodically forwards the FULL text-so-far as
// one evolving progress line, tagged so the client can render it as the live
// body (replaced by the final result on completion).
const (
	// assistantPartialProgressCode is a sentinel progress code OUTSIDE the
	// control-plane canonical template range (1001-1203), so the control-plane
	// keeps the verbatim accumulated body instead of rendering a template.
	assistantPartialProgressCode agentbridge.ProgressCode = 9001
	// assistantPartialProgressKey is relayed verbatim to the client as the
	// line's message_key so it can distinguish the evolving body from status
	// narration lines.
	assistantPartialProgressKey = "assistant.partial"
	// partialBodyFlushInterval / partialBodyFlushChars debounce forwarding so
	// many deltas coalesce into a coherent, low-frequency update.
	partialBodyFlushInterval = 350 * time.Millisecond
	partialBodyFlushChars    = 24
)

// AgentBinding maps a SaaS agent identity to one local provider runtime.
type AgentBinding struct {
	AgentID         string
	RuntimeProvider string
}

type Config struct {
	BaseURL        string
	DaemonID       string
	DeviceID       string
	DeviceSecret   string
	Profile        string
	AppVersion     string
	PID            int
	StartedAt      time.Time
	Agents         []AgentBinding
	BearerToken    string
	HTTPClient     *http.Client
	RequestTimeout time.Duration
	LongPollWait   time.Duration
}

type RuntimeModelRecord struct {
	ModelID   string `json:"model_id"`
	Label     string `json:"label"`
	IsDefault bool   `json:"is_default"`
}

type RuntimeSnapshotRecord struct {
	RuntimeID                 string               `json:"runtime_id"`
	Kind                      string               `json:"kind"`
	Availability              string               `json:"availability,omitempty"`
	DetectionState            string               `json:"detection_state,omitempty"`
	ProviderVersion           string               `json:"provider_version,omitempty"`
	RequiresExperimentalOptIn bool                 `json:"requires_experimental_opt_in,omitempty"`
	Models                    []RuntimeModelRecord `json:"models,omitempty"`
}

type DeviceRuntimeSnapshotSyncRequest struct {
	DaemonID          string                  `json:"daemon_id"`
	DeviceID          string                  `json:"device_id,omitempty"`
	DeviceDisplayName string                  `json:"device_display_name,omitempty"`
	Profile           string                  `json:"profile,omitempty"`
	AppVersion        string                  `json:"app_version,omitempty"`
	PID               int                     `json:"pid,omitempty"`
	UptimeSeconds     int64                   `json:"uptime_seconds,omitempty"`
	StartedAt         time.Time               `json:"started_at,omitzero"`
	Runtimes          []RuntimeSnapshotRecord `json:"runtimes"`
}

type AgentRuntimeBindingListResponse struct {
	SchemaVersion string                                   `json:"schema_version"`
	Bindings      []assignmentcontract.AgentRuntimeBinding `json:"bindings"`
}

// Plane implements both TaskSourcePort and TaskReporterPort against the
// control-plane assignment polling API. Internal state is owned by a mailbox
// goroutine so the supervisor can use the adapter without shared mutable maps.
type Plane struct {
	cfg    Config
	client *http.Client
	ops    chan stateOp
	done   chan struct{}
}

func New(cfg Config) (*Plane, error) {
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if cfg.BaseURL == "" {
		return nil, errors.New("saasplane: BaseURL is required")
	}
	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("saasplane: invalid BaseURL: %w", err)
	}
	cfg.DaemonID = strings.TrimSpace(cfg.DaemonID)
	if cfg.DaemonID == "" {
		return nil, errors.New("saasplane: DaemonID is required")
	}
	cfg.DeviceID = strings.TrimSpace(cfg.DeviceID)
	if cfg.DeviceID == "" {
		cfg.DeviceID = cfg.DaemonID
	}
	cfg.DeviceSecret = strings.TrimSpace(cfg.DeviceSecret)
	cfg.Profile = strings.TrimSpace(cfg.Profile)
	cfg.AppVersion = strings.TrimSpace(cfg.AppVersion)
	cfg.BearerToken = strings.TrimSpace(cfg.BearerToken)
	if cfg.DeviceSecret != "" && cfg.DeviceID == "" {
		return nil, errors.New("saasplane: DeviceID is required when DeviceSecret is set")
	}
	cfg.Agents = normalizeAgents(cfg.Agents)
	if len(cfg.Agents) == 0 && cfg.DeviceSecret == "" {
		return nil, errors.New("saasplane: at least one static agent binding or a device credential is required")
	}
	if cfg.LongPollWait <= 0 {
		cfg.LongPollWait = defaultLongPollWait
	}
	minRequestTimeout := cfg.LongPollWait + longPollRequestTimeoutPad
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = minRequestTimeout
	}
	if cfg.RequestTimeout < minRequestTimeout {
		cfg.RequestTimeout = minRequestTimeout
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.RequestTimeout}
	}
	p := &Plane{
		cfg:    cfg,
		client: client,
		ops:    make(chan stateOp, 64),
		done:   make(chan struct{}),
	}
	go p.loop()
	return p, nil
}

func (p *Plane) Close() {
	ack := make(chan struct{})
	select {
	case p.ops <- stateOp{close: true, ack: ack}:
		<-ack
	case <-p.done:
	}
}
