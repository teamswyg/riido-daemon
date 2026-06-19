package main

var fixtureSourceFiles = []struct {
	path string
	body string
}{
	{"internal/agentbridge/adapter.go", "package agentbridge\ntype Adapter interface { BuildStart(); NewParser() Parser; Translate() }\n"},
	{"internal/agentbridge/adapter_parser.go", "package agentbridge\ntype Parser interface { FeedStdout(); FeedStderr(); Close() }\n"},
	{"internal/process/port.go", "package process\ntype Process interface { Start() }\ntype RunningProcess interface { Stdout(); Stderr(); Exited(); WriteStdin(); CloseStdin(); Kill() }\n"},
	{"internal/agentbridge/session/start.go", "package session\nfunc x(){ cfg.Process.Start(ctx, cfg.Spawn) }\n"},
	{"internal/agentbridge/session/session_runner_new.go", "package session\nfunc x(){ _ = struct{stdoutCh any}{stdoutCh:  proc.Stdout()} }\n"},
}
