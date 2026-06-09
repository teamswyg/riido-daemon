package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
)

type codexPersistentRunner struct {
	cfg       runtimeactor.PersistentRunnerConfig
	proc      process.RunningProcess
	spawn     process.Command
	cwd       string
	nextRPCID int64
}

func newCodexPersistentRunner(cfg runtimeactor.PersistentRunnerConfig) (*codexPersistentRunner, error) {
	if cfg.Adapter == nil {
		return nil, errors.New("codex persistent runner: adapter is required")
	}
	if cfg.Process == nil {
		return nil, errors.New("codex persistent runner: process port is required")
	}
	if cfg.Now == nil {
		return nil, errors.New("codex persistent runner: clock is required")
	}
	return &codexPersistentRunner{cfg: cfg}, nil
}

func (r *codexPersistentRunner) Submit(ctx context.Context, req agentbridge.StartRequest, opts runtimeactor.PersistentRunOptions) (agentbridge.RunHandle, error) {
	if err := r.ensureProcess(ctx, req); err != nil {
		return nil, err
	}
	seed := r.nextRPCID
	if seed <= 0 {
		seed = 2
	}
	r.nextRPCID = seed + 1000
	driver, err := codex.NewTurnProtocolDriverWithIDSeed(req, seed)
	if err != nil {
		return nil, err
	}
	return session.Start(ctx, session.Config{
		TaskID:           req.TaskID,
		RuntimeID:        r.cfg.RuntimeID,
		Adapter:          r.cfg.Adapter,
		Spawn:            r.spawn,
		Running:          r.proc,
		KeepProcessAlive: true,
		Request:          req,
		HardTimeout:      opts.HardTimeout,
		SemanticIdle:     opts.SemanticIdle,
		AutoApprove:      opts.AutoApprove,
		ToolStartGate:    opts.ToolStartGate,
		ProtocolDriver:   driver,
		Now:              r.cfg.Now,
	})
}

func (r *codexPersistentRunner) Stop(ctx context.Context) error {
	if r.proc == nil {
		return nil
	}
	err := r.proc.Kill(ctx)
	r.proc = nil
	r.cwd = ""
	r.nextRPCID = 0
	return err
}

func (r *codexPersistentRunner) ensureProcess(ctx context.Context, req agentbridge.StartRequest) error {
	start, err := r.cfg.Adapter.BuildStart(req)
	if err != nil {
		return fmt.Errorf("codex persistent runner: BuildStart: %w", err)
	}
	spawn := codexPersistentProcessCommand(start)
	if spawn.Dir == "" {
		spawn.Dir = req.Cwd
	}

	if r.proc != nil && processExited(r.proc) {
		r.proc = nil
		r.cwd = ""
		r.nextRPCID = 0
	}
	if r.proc != nil && !sameCodexPersistentSpawn(r.spawn, spawn) {
		_ = r.proc.Kill(ctx)
		r.proc = nil
		r.cwd = ""
		r.nextRPCID = 0
	}
	if r.proc != nil {
		return nil
	}
	proc, err := r.cfg.Process.Start(ctx, spawn)
	if err != nil {
		return fmt.Errorf("codex persistent runner: process start: %w", err)
	}
	r.proc = proc
	r.spawn = spawn
	r.cwd = req.Cwd
	if r.cwd == "" {
		r.cwd = spawn.Dir
	}
	if err := initializeCodexAppServer(ctx, r.cfg.Adapter, proc); err != nil {
		_ = proc.Kill(context.Background())
		r.proc = nil
		r.cwd = ""
		r.nextRPCID = 0
		return err
	}
	r.nextRPCID = 2
	return nil
}

func sameCodexPersistentSpawn(a, b process.Command) bool {
	return a.Executable == b.Executable &&
		a.Dir == b.Dir &&
		slices.Equal(a.Args, b.Args) &&
		slices.Equal(a.Env, b.Env)
}

func processExited(proc process.RunningProcess) bool {
	select {
	case <-proc.Exited():
		return true
	default:
		return false
	}
}

func codexPersistentProcessCommand(c agentbridge.StartCommand) process.Command {
	return process.Command{
		Executable: c.Executable,
		Args:       append([]string(nil), c.Args...),
		Env:        append([]string(nil), c.Env...),
		Dir:        c.Dir,
	}
}

type codexPersistentProtocolIO struct {
	proc process.RunningProcess
}

func (io codexPersistentProtocolIO) WriteStdin(_ context.Context, b []byte) error {
	return io.proc.WriteStdin(b)
}

func (io codexPersistentProtocolIO) CloseStdin(_ context.Context) error {
	return io.proc.CloseStdin()
}

func initializeCodexAppServer(ctx context.Context, adapter agentbridge.Adapter, proc process.RunningProcess) error {
	io := codexPersistentProtocolIO{proc: proc}
	if err := writeJSONRPC(ctx, io, map[string]any{
		"jsonrpc": "2.0",
		"id":      int64(1),
		"method":  "initialize",
		"params": map[string]any{
			"clientInfo": map[string]any{"name": "riido", "version": "0.0.0"},
		},
	}); err != nil {
		return err
	}

	parser := adapter.NewParser()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case status, ok := <-proc.Exited():
			if !ok {
				return errors.New("codex app-server exited during initialize")
			}
			return fmt.Errorf("codex app-server exited during initialize code=%d err=%v", status.Code, status.Err)
		case chunk, ok := <-proc.Stdout():
			if !ok {
				return errors.New("codex app-server stdout closed during initialize")
			}
			raws, err := parser.FeedStdout(chunk)
			if err != nil {
				return fmt.Errorf("codex initialize parse stdout: %w", err)
			}
			for _, raw := range raws {
				done, err := handleCodexInitializeRaw(ctx, io, raw)
				if err != nil {
					return err
				}
				if done {
					return nil
				}
			}
		case chunk, ok := <-proc.Stderr():
			if !ok {
				continue
			}
			_, _ = parser.FeedStderr(chunk)
		}
	}
}

func handleCodexInitializeRaw(ctx context.Context, io codexPersistentProtocolIO, raw agentbridge.RawEvent) (bool, error) {
	if raw.Type == "error" {
		if id, ok := jsonRPCID(raw.Payload); ok && id == 1 {
			return false, fmt.Errorf("codex initialize rpc error: %s", jsonRPCErrorMessage(raw.Payload))
		}
	}
	if raw.Type != "response" {
		return false, nil
	}
	id, ok := jsonRPCID(raw.Payload)
	if !ok || id != 1 {
		return false, nil
	}
	if err := writeJSONRPC(ctx, io, map[string]any{
		"jsonrpc": "2.0",
		"method":  "initialized",
		"params":  map[string]any{},
	}); err != nil {
		return false, err
	}
	return true, nil
}

func writeJSONRPC(ctx context.Context, io codexPersistentProtocolIO, frame map[string]any) error {
	b, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return io.WriteStdin(ctx, b)
}

func jsonRPCID(payload map[string]any) (int64, bool) {
	if payload == nil {
		return 0, false
	}
	switch v := payload["id"].(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return n, err == nil
	default:
		return 0, false
	}
}

func jsonRPCErrorMessage(payload map[string]any) string {
	if payload == nil {
		return "unknown error"
	}
	errMap, _ := payload["error"].(map[string]any)
	if errMap == nil {
		return "unknown error"
	}
	if msg, _ := errMap["message"].(string); strings.TrimSpace(msg) != "" {
		return strings.TrimSpace(msg)
	}
	return "unknown error"
}
