package process

import "testing"

func TestFakeProcessDefaultBuffersMatchProviderRuntimeBackpressureSSOT(t *testing.T) {
	running := NewFakeRunning()
	if got := cap(running.Stdout()); got != DefaultStdoutBuffer {
		t.Fatalf("stdout buffer = %d, want %d", got, DefaultStdoutBuffer)
	}
	if got := cap(running.Stderr()); got != DefaultStderrBuffer {
		t.Fatalf("stderr buffer = %d, want %d", got, DefaultStderrBuffer)
	}
}
