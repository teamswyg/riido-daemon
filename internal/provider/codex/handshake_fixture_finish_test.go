package codex

func (f *codexHandshakeFixture) resolveOrphanResponse() {
	f.rpc.Resolve(99999, map[string]any{"orphan": true}, nil)
}

func (f *codexHandshakeFixture) finish() {
	f.running.EmitExit(0, nil)
}
