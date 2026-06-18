// Package session implements the run-scope session actor: one goroutine
// that owns one provider session's State. It wires Process -> Parser ->
// Adapter.Translate -> Reducer -> emit Events/Result.
//
// The actor is the single owner of the agentbridge.State for its run.
// The reducer is called inline; no other goroutine ever touches State.
// No sync.Mutex / sync.RWMutex is used: backpressure and shutdown are
// expressed via bounded channels per docs/20-domain/provider-runtime.md.
package session
