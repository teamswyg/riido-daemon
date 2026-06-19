package main

const fixtureManifestSource = `{
  "schema_version": "riido-shutdown-authority.v1",
  "id": "provider-runtime-cancel-interrupt-input",
  "title": "Cancel, Interrupt, And Input",
  "generated_doc": "docs/20-domain/provider-runtime/adapter-draft-fields/cancel-interrupt-input.md",
  "evidence_artifact": "shutdown-authority-evidence",
  "sources": {"levels": "pkg/lifecycle/shutdown_level.go", "timeouts": "pkg/lifecycle/shutdown.go"},
  "levels": [
    {"name": "none", "const": "ShutdownNone", "order": 0},
    {"name": "graceful", "const": "ShutdownGraceful", "order": 1},
    {"name": "forced", "const": "ShutdownForced", "order": 2}
  ],
  "timeouts": [
    {"const": "DefaultGracefulShutdownTimeout", "duration": "5s"},
    {"const": "DefaultForcedShutdownTimeout", "duration": "1s"}
  ],
  "consumer_requirements": [
    {"file": "cmd/riido/daemon_ipc_request.go", "contains": "lifecycle.ParseShutdownLevel"},
    {"file": "cmd/riido/daemon_supervisor_start.go", "contains": "lifecycle.DetachedDefaultShutdown"},
    {"file": "internal/agentbridge/runtimeactor/stop.go", "contains": "lifecycle.NormalizeShutdownLevel"},
    {"file": "internal/agentbridge/supervisor/stop.go", "contains": "lifecycle.NormalizeShutdownLevel"},
    {"file": "internal/agentbridge/session/process_kill.go", "contains": "lifecycle.DetachedShutdown"}
  ]
}`

const fixtureLevelSource = `package lifecycle
type ShutdownLevel uint8
const ( ShutdownNone ShutdownLevel = iota; ShutdownGraceful; ShutdownForced )
func (l ShutdownLevel) String() string {
 switch l {
 case ShutdownNone: return "none"
 case ShutdownGraceful: return "graceful"
 case ShutdownForced: return "forced"
 default: return "unknown"
 }
}`

const fixtureTimeoutSource = `package lifecycle
import "time"
const (
 DefaultGracefulShutdownTimeout = 5 * time.Second
 DefaultForcedShutdownTimeout = time.Second
)`
