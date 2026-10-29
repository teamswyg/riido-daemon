package agentbridge

// StartRequest is the provider-neutral input to BuildStart.
type StartRequest struct {
	ReferenceDirectories []string
	TaskID               string
	Prompt               string
	Cwd                  string
	Executable           string
	Model                string
	SystemPrompt         string
	MaxTurns             int
	ResumeSessionID      string
	Env                  map[string]string
	CustomArgs           []string
	MCPConfig            []byte
	Metadata             map[string]string
	// ContextCLIEnabled is set by the daemon launch hook, never by run Env.
	ContextCLIEnabled bool
	// ContextCLIExecutable is the daemon-owned absolute binary, not assignment input.
	ContextCLIExecutable string
}

// StartCommand is the process-port input the session actor uses to spawn the
// provider CLI.
type StartCommand struct {
	Executable  string
	Args        []string
	Env         []string
	Dir         string
	StdinMode   StdinMode
	DroppedArgs []string
	TempFiles   []string
}

// StdinMode tells the session actor how to feed stdin to the provider process.
type StdinMode string

const (
	StdinNone   StdinMode = "none"
	StdinPipe   StdinMode = "pipe"
	StdinPrompt StdinMode = "prompt"
)
