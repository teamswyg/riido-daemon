package bridge

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
)

// Run spawns a session for the named provider and returns a Session
// handle. The caller MUST drain Events() until it is closed; otherwise
// the session goroutine will block on send.
func (c *Client) Run(ctx context.Context, req TaskRequest) (*Session, error) {
	runtime, err := c.resolveRuntime(ctx, req)
	if err != nil {
		return nil, err
	}

	startReq, launchEnv := newStartRequest(req, runtime.detect.Executable)
	spawnCmd, err := runtime.adapter.BuildStart(startReq)
	if err != nil {
		return nil, fmt.Errorf("bridge: BuildStart %s: %w", req.Provider, err)
	}

	driver, err := newProtocolDriver(runtime.adapter, startReq, req.Provider)
	if err != nil {
		return nil, err
	}

	spawnProcess := newSpawnProcess(spawnCmd, startReq.Cwd, launchEnv)
	cfg := c.newSessionConfig(req, runtime.adapter, startReq, spawnProcess, driver, spawnCmd.TempFiles)
	inner, err := session.Start(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return newSession(inner, spawnCmd.DroppedArgs), nil
}
