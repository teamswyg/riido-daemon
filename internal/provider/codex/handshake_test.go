package codex

import "testing"

// TestCodexFullHandshakeComposition is the M-3 regression: drive every
// step of the Codex app-server JSON-RPC handshake through Parser +
// RPCActor + Translator together, with a fake process providing stdio.
//
// Steps exercised (spec §3.2):
//
//  1. initialize request -> response
//  2. initialized notification (no response)
//  3. thread/start request -> response -> thread_started notification
//  4. turn/start request -> response -> turn_started notification
//  5. streaming notification: agent_message
//  6. server_request: approve_command -> daemon replies with approval
//  7. turn_completed notification -> translated to EventResult
//  8. pending-map cleanup: a stray response with an unknown id is a no-op
//
// Plus: semantic-idle equivalent: RPC actor releases pending callers
// when Close is invoked.
func TestCodexFullHandshakeComposition(t *testing.T) {
	handshake := newCodexHandshakeFixture(t)
	defer handshake.close()

	handshake.initialize()
	handshake.notifyInitialized()
	handshake.startThread()
	handshake.startTurn()
	handshake.streamAgentMessage()
	handshake.approveCommand()
	handshake.completeTurn()
	handshake.resolveOrphanResponse()
	handshake.finish()
}
