package main

var fixtureSourceFiles = []struct {
	path string
	body string
}{
	{"internal/agentbridge/session/session_runner_timers.go", "package session\nvar _ = time.NewTimer(r.cfg.HardTimeout)\nvar _ = time.NewTimer(r.cfg.SemanticIdle)\n"},
	{"internal/agentbridge/session/session_runner_emit.go", "package session\nfunc x(){ _ = ev.Kind.IsSemanticActivity() }\n"},
	{"internal/agentbridge/session/session_tool_approval_resolver.go", "package session\nfunc x(){ select { case <-r.hardC: case <-r.idleC: } }\n"},
}
