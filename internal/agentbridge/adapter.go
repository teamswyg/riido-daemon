package agentbridge

import (
	"context"
	"strings"
)

// Adapter is the run-scope provider port: a single provider CLI plugin.
//
// Adapters are translators, not state owners. They build a process
// command, parse raw provider output, and translate each raw provider
// event into one or more run-scope Events. Adapters MUST NOT own a
// state machine of their own (docs/20-domain/provider-runtime.md §1),
// MUST NOT import any other provider's package, and MUST NOT touch the
// filesystem or network outside of their Detect/BuildStart returned
// values.
type Adapter interface {
	Name() string
	Detect(ctx context.Context, env DetectEnv) (DetectResult, error)
	BuildStart(req StartRequest) (StartCommand, error)
	NewParser() Parser
	Translate(raw RawEvent) ([]Event, []Command, error)
	BlockedArgs() []string
}

// RunHandle is the provider-neutral handle for one task/run. Most providers
// implement this with one spawned process per run. Persistent providers can
// return a run handle backed by a long-lived process and a run-scoped protocol
// driver.
type RunHandle interface {
	Events() <-chan Event
	Result() <-chan Result
	Done() <-chan struct{}
	Cancel(error)
}

// DetectEnv carries the environment an adapter consults during Detect.
// Adapters MUST read only from this struct, never os.Environ directly —
// the daemon supplies everything they need.
type DetectEnv struct {
	Executable  string
	PathExtra   []string
	EnvOverride map[string]string
}

// DetectResult is the snapshot of capability the adapter returns from Detect.
// This is distinct from the full C3 ProviderCapability contract in
// github.com/teamswyg/riido-contracts/provider/capability: DetectResult is the
// raw daemon observation that a later runtimeactor migration promotes into C3
// through reconciliation.
type DetectResult struct {
	Available         bool
	Executable        string
	Version           string
	SupportsStreaming bool
	SupportsResume    bool
	SupportsSystem    bool
	SupportsMaxTurns  bool
	SupportsMCP       bool
	SupportsToolHooks bool
	SupportsUsage     bool
	Reason            string
	Metadata          map[string]string
}

// StartRequest is the provider-neutral input to BuildStart.
type StartRequest struct {
	TaskID          string // stable task id; adapters may use it only as a provider-neutral correlation seed
	Prompt          string
	Cwd             string
	Executable      string // runtime-selected executable path from Detect; adapter options may still override it
	Model           string
	SystemPrompt    string
	MaxTurns        int
	ResumeSessionID string
	Env             map[string]string
	CustomArgs      []string
	MCPConfig       []byte
	Metadata        map[string]string
}

// StartCommand is the process-port input the session actor uses to
// spawn the provider CLI.
type StartCommand struct {
	Executable  string
	Args        []string
	Env         []string
	Dir         string
	StdinMode   StdinMode
	DroppedArgs []string // custom args removed because they collide with BlockedArgs
	TempFiles   []string // adapter-owned temp paths to delete on process exit
}

// StdinMode tells the session actor how to feed stdin to the provider
// process. StdinPrompt means "write the StartRequest.Prompt as the
// first stdin frame"; StdinPipe means "open a writable stdin and let
// the adapter's translator drive WriteProviderInput commands".
type StdinMode string

const (
	StdinNone   StdinMode = "none"
	StdinPipe   StdinMode = "pipe"
	StdinPrompt StdinMode = "prompt"
)

// Parser owns the per-process line buffer state. Each Feed call returns
// zero or more RawEvent envelopes that Translate then maps to run-scope
// Events.
type Parser interface {
	FeedStdout(chunk []byte) ([]RawEvent, error)
	FeedStderr(chunk []byte) ([]RawEvent, error)
	Close() ([]RawEvent, error)
}

// RawEvent is the post-parse, pre-translate envelope. The Type field
// is provider-specific and meaningful only to the adapter's Translate.
type RawEvent struct {
	Source  RawSource
	Type    string
	Payload map[string]any
	Bytes   []byte
}

// RawSource tags which stream produced the raw event.
type RawSource string

const (
	RawSourceStdout RawSource = "stdout"
	RawSourceStderr RawSource = "stderr"
	RawSourceClose  RawSource = "close"
)

// FilterBlockedArgs removes adapter-blocked args from caller-supplied
// custom args. Blocked args are typically protocol-critical flags the adapter
// sets itself, such as a provider output-format flag. The dropped list is
// reported so the session actor can emit a Warning event per the C4 provider
// runtime contract.
//
// Both the bare form (--flag value) and the equals form (--flag=value)
// are recognized.
func FilterBlockedArgs(custom []string, blocked []string) (kept []string, dropped []string) {
	blockedSet := make(map[string]struct{}, len(blocked))
	for _, b := range blocked {
		blockedSet[b] = struct{}{}
	}
	for i := 0; i < len(custom); i++ {
		arg := custom[i]
		if _, isBlocked := blockedSet[arg]; isBlocked {
			dropped = append(dropped, arg)
			if i+1 < len(custom) && !strings.HasPrefix(custom[i+1], "-") {
				dropped = append(dropped, custom[i+1])
				i++
			}
			continue
		}
		if eq := strings.IndexByte(arg, '='); eq > 0 {
			if _, isBlocked := blockedSet[arg[:eq]]; isBlocked {
				dropped = append(dropped, arg)
				continue
			}
		}
		kept = append(kept, arg)
	}
	return kept, dropped
}
