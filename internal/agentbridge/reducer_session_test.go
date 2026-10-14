package agentbridge

import "testing"

func TestReduceSessionIdentifiedEmitsPersist(t *testing.T) {
	s := NewState()
	s, cmds := Reduce(s, Event{Kind: EventSessionIdentified, SessionID: "sess-1"}, nil)
	if s.SessionID != "sess-1" {
		t.Fatalf("session id not set: %q", s.SessionID)
	}
	if len(cmds) != 1 || cmds[0].Kind != CommandPersistSession {
		t.Fatalf("expected one CommandPersistSession, got %+v", cmds)
	}
}

func TestReduceSessionIdentifiedLate(t *testing.T) {
	s := NewState()
	s, _ = Reduce(s, Event{Kind: EventTextDelta, Text: "hi"}, nil)
	s, _ = Reduce(s, Event{Kind: EventSessionIdentified, SessionID: "late"}, nil)
	if s.SessionID != "late" {
		t.Fatalf("session id not updated: %q", s.SessionID)
	}
}
