package supervisor

import "testing"

func assertDefaultMailboxSize(t *testing.T, actor *Actor) {
	t.Helper()
	if got := cap(actor.mailbox); got != DefaultMailboxSize {
		t.Fatalf("mailbox size = %d, want %d", got, DefaultMailboxSize)
	}
}
