package agentbridge

// StartRequest is the provider-neutral input to BuildStart.
type StartRequest struct {
	TaskID          string
	Prompt          string
	Cwd             string
	Executable      string
	Model           string
	SystemPrompt    string
	MaxTurns        int
	ResumeSessionID string
	Env             map[string]string
	CustomArgs      []string
	MCPConfig       []byte
	Metadata        map[string]string
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
