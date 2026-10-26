package contextbroker

import (
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// Broker answers `context_read` requests for runs on this daemon.
type Broker struct {
	Registry *Registry
	// Socket is injected into runs as RIIDO_DAEMON_SOCKET.
	Socket string
	// BinDir is prepended to PATH so runs resolve the bundled `riido`.
	BinDir string
	// Fetcher reads assignment context snapshots; nil means reads stay
	// CONTROL_PLANE_UNSUPPORTED (non-SaaS control planes).
	Fetcher Fetcher
}

// New returns a broker with an empty registry.
func New(socket, binDir string) *Broker {
	return &Broker{Registry: NewRegistry(), Socket: socket, BinDir: binDir}
}

// RunActive reports whether runKey still has a running session. certain is
// false when a runtime could not answer in time: the run may still be going.
type RunActive func(runKey string) (active, certain bool)

// WhoamiData is returned for `riido whoami` in agent mode. Requester
// identity needs the control-plane run grant, so it is reported as missing.
type WhoamiData struct {
	Mode      contextcli.Mode `json:"mode"`
	Run       Binding         `json:"run"`
	Requester *string         `json:"requester"`
}

// CapabilitiesData is returned for the broker capability probe.
type CapabilitiesData struct {
	Broker       string  `json:"broker"`
	ControlPlane string  `json:"control_plane"`
	Run          Binding `json:"run"`
}

const (
	controlPlaneUnavailable        = "unsupported"
	controlPlaneAssignmentSnapshot = "assignment_snapshot"
)

// Handle verifies the run capability and answers inside the run scope.
func (b *Broker) Handle(req contextcli.ContextReadRequest, active RunActive) contextcli.Envelope {
	requestID := req.RequestID
	if requestID == "" {
		requestID = contextcli.NewRequestID()
	}
	binding, cliErr := b.authorize(req, active)
	if cliErr != nil {
		return contextcli.FailureEnvelope(requestID, cliErr)
	}
	switch req.Command {
	case "capabilities":
		controlPlane := controlPlaneUnavailable
		if b.Fetcher != nil {
			controlPlane = controlPlaneAssignmentSnapshot
		}
		return contextcli.NewEnvelope(requestID, CapabilitiesData{Broker: "daemon", ControlPlane: controlPlane, Run: binding})
	case "whoami":
		env := contextcli.NewEnvelope(requestID, WhoamiData{Mode: contextcli.ModeAgent, Run: binding})
		env.Warnings = append(env.Warnings, contextcli.Warning{
			Code:    "REQUESTER_UNRESOLVED",
			Message: "요청자 확인에는 control-plane 실행 grant가 필요합니다",
		})
		return env
	default:
		return b.read(req, requestID, binding)
	}
}

func (b *Broker) authorize(req contextcli.ContextReadRequest, active RunActive) (Binding, *contextcli.CLIError) {
	if req.SchemaVersion != contextcli.ContextIPCSchemaVersion {
		return Binding{}, contextcli.Failf(contextcli.CodeUnsupported, "지원하지 않는 context_read 스키마입니다: %q", req.SchemaVersion)
	}
	if req.RunCapability == "" {
		return Binding{}, contextcli.Fail(contextcli.CodePermissionDenied,
			"실행 capability가 없습니다", "Riido가 시작한 Agent 실행 안에서만 에이전트 모드를 쓸 수 있습니다")
	}
	binding, ok := b.Registry.Lookup(req.RunCapability)
	if !ok {
		return Binding{}, contextcli.Fail(contextcli.CodePermissionDenied, "알 수 없거나 만료된 실행 capability입니다", "")
	}
	if active == nil {
		return Binding{}, contextcli.Fail(contextcli.CodePermissionDenied, "이미 끝난 실행의 capability입니다", "")
	}
	running, certain := active(binding.RunKey)
	if !certain {
		// Inconclusive: keep the capability so a retry can still succeed.
		return Binding{}, contextcli.Fail(contextcli.CodeUnavailable, "실행 상태를 확인하지 못했습니다", "잠시 후 다시 시도하세요")
	}
	if !running {
		b.Registry.Revoke(binding.RunKey)
		return Binding{}, contextcli.Fail(contextcli.CodePermissionDenied, "이미 끝난 실행의 capability입니다", "")
	}
	if outOfScope(req.WorkspaceID, binding.WorkspaceID) || outOfScope(req.TaskID, binding.TaskID) {
		return Binding{}, contextcli.Fail(contextcli.CodeOutOfRunScope, "요청 범위가 이 실행과 다릅니다", "")
	}
	return binding, nil
}

func outOfScope(requested, bound string) bool {
	return requested != "" && bound != "" && requested != bound
}
