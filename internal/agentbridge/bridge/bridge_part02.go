package bridge

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// Run spawns a session for the named provider and returns a Session
// handle. The caller MUST drain Events() until it is closed; otherwise
// the session goroutine will block on send.
func (c *Client) Run(ctx context.Context, req TaskRequest) (*Session, error) {
	adapter, ok := c.adapters[req.Provider]
	if !ok {
		return nil, fmt.Errorf("bridge: unknown provider %q", req.Provider)
	}
	if c.process == nil {
		return nil, errors.New("bridge: Process port not configured")
	}
	det, err := adapter.Detect(ctx, agentbridge.DetectEnv{})
	if err != nil {
		return nil, fmt.Errorf("bridge: detect %s: %w", req.Provider, err)
	}
	if !det.Available {
		return nil, fmt.Errorf("bridge: provider %s unavailable: %s", req.Provider, det.Reason)
	}

	launchEnv := detectutil.EnvMapWithLaunchPATH(req.Env)
	startReq := agentbridge.StartRequest{
		TaskID:          req.ID,
		Prompt:          req.Prompt,
		Cwd:             req.Cwd,
		Executable:      det.Executable,
		Model:           req.Model,
		SystemPrompt:    req.SystemPrompt,
		MaxTurns:        req.MaxTurns,
		ResumeSessionID: req.ResumeSessionID,
		Env:             launchEnv,
		CustomArgs:      req.CustomArgs,
		MCPConfig:       req.MCPConfig,
		Metadata:        req.Metadata,
	}
	spawnCmd, err := adapter.BuildStart(startReq)
	if err != nil {
		return nil, fmt.Errorf("bridge: BuildStart %s: %w", req.Provider, err)
	}

	var driver agentbridge.ProtocolDriver
	if provider, ok := adapter.(agentbridge.ProtocolDriverProvider); ok {
		drv, err := provider.NewProtocolDriver(startReq)
		if err != nil {
			return nil, fmt.Errorf("bridge: NewProtocolDriver %s: %w", req.Provider, err)
		}
		driver = drv
	}

	spawnProcess := toProcessCommand(spawnCmd)
	if spawnProcess.Dir == "" {
		spawnProcess.Dir = startReq.Cwd
	}
	spawnProcess.Env = detectutil.EnvListWithLaunchPATHFromMap(spawnProcess.Env, launchEnv)

	cfg := session.Config{
		TaskID:         req.ID,
		RuntimeID:      string(req.Provider),
		Adapter:        adapter,
		Process:        c.process,
		Spawn:          spawnProcess,
		Request:        startReq,
		HardTimeout:    firstNonZero(req.Timeout, c.defaults.timeout),
		SemanticIdle:   firstNonZero(req.SemanticIdle, c.defaults.semanticIdle),
		AutoApprove:    c.autoApprove,
		ToolStartGate:  c.toolStartGate,
		ProtocolDriver: driver,
		TempFiles:      spawnCmd.TempFiles,
	}
	inner, err := session.Start(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Session{inner: inner, droppedArgs: spawnCmd.DroppedArgs}, nil
}

// Session is the caller-facing handle for one run.
type Session struct {
	inner       *session.Session
	droppedArgs []string
}

// Events returns the run-scope event stream, closed when the session ends.
func (s *Session) Events() <-chan agentbridge.Event { return s.inner.Events() }

// Result returns the single-value terminal result channel.
func (s *Session) Result() <-chan agentbridge.Result { return s.inner.Result() }

// Cancel signals the session to terminate as ResultCancelled.
func (s *Session) Cancel(cause error) { s.inner.Cancel(cause) }

// DroppedArgs returns the custom args that BuildStart removed because
// they collided with the adapter's BlockedArgs.
func (s *Session) DroppedArgs() []string { return s.droppedArgs }

func toProcessCommand(c agentbridge.StartCommand) process.Command {
	return process.Command{
		Executable: c.Executable,
		Args:       c.Args,
		Env:        c.Env,
		Dir:        c.Dir,
	}
}

func firstNonZero(a, b time.Duration) time.Duration {
	if a > 0 {
		return a
	}
	return b
}

func sortByProvider(caps []RuntimeCapability) {
	// Simple insertion sort — order matters only for test stability.
	for i := 1; i < len(caps); i++ {
		for j := i; j > 0 && caps[j-1].Provider > caps[j].Provider; j-- {
			caps[j-1], caps[j] = caps[j], caps[j-1]
		}
	}
}
