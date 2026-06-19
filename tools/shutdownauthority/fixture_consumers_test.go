package main

var fixtureConsumerFiles = []struct {
	path string
	body string
}{
	{"cmd/riido/daemon_ipc_request.go", "package main\nvar _ = lifecycle.ParseShutdownLevel\n"},
	{"cmd/riido/daemon_supervisor_start.go", "package main\nvar _ = lifecycle.DetachedDefaultShutdown\n"},
	{"internal/agentbridge/runtimeactor/stop.go", "package runtimeactor\nvar _ = lifecycle.NormalizeShutdownLevel\n"},
	{"internal/agentbridge/supervisor/stop.go", "package supervisor\nvar _ = lifecycle.NormalizeShutdownLevel\n"},
	{"internal/agentbridge/session/process_kill.go", "package session\nvar _ = lifecycle.DetachedShutdown\n"},
}
