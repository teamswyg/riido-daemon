package runtimeactor

import (
	"context"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type resumeTestAdapter struct {
	stubAdapter
	probes int
}

func (a *resumeTestAdapter) ResumeExecutableIdentity(path, version string) agentbridge.DetectResult {
	r := a.detected
	r.Executable = path
	r.Version = version
	return r
}
func (a *resumeTestAdapter) ProbeResumeExecutable(_ context.Context, path string) (agentbridge.DetectResult, error) {
	a.probes++
	return a.ResumeExecutableIdentity(path, a.detected.Version), nil
}

func TestNativeResumeVerifiesExecutableBeforeProbeAndSubmit(t *testing.T) {
	file := filepath.Join(t.TempDir(), "codex")
	if err := os.WriteFile(file, []byte("original"), 0700); err != nil {
		t.Fatal(err)
	}
	adapter := &resumeTestAdapter{stubAdapter: stubAdapter{name: "codex", detected: agentbridge.DetectResult{Version: "original", Available: true, Executable: file, SupportsResume: true}}}
	a := &Actor{cfg: Config{RuntimeID: "runtime", Adapters: []agentbridge.Adapter{adapter}, Now: time.Now}}
	c, err := buildRuntimeCapability("runtime", "codex", adapter.detected, "", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	selected, err := a.ResolveNativeResume(context.Background(), "codex", c.CapabilityFingerprint, file, "original")
	if err != nil || adapter.probes != 1 {
		t.Fatalf("selection=%v probes=%d", err, adapter.probes)
	}
	msg := &submitMsg{ctx: context.Background(), nativeResume: selected, req: bridge.TaskRequest{Provider: "codex", ResumeSessionID: "session", ExpectedCapabilityFingerprint: c.CapabilityFingerprint}}
	if _, got, err := a.submitCapability(msg, indexAdapters(a.cfg.Adapters), nil, nil); err != nil || got.Executable != file {
		t.Fatalf("submit=%+v %v", got, err)
	}
	if err := os.WriteFile(file, []byte("changed"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, _, err := a.submitCapability(msg, indexAdapters(a.cfg.Adapters), nil, nil); err == nil {
		t.Fatal("changed CLI accepted")
	}
	if _, err := a.ResolveNativeResume(context.Background(), "codex", c.CapabilityFingerprint, file, "original"); err == nil {
		t.Fatal("changed candidate accepted")
	}
	if adapter.probes != 1 {
		t.Fatal("executed a candidate with the wrong fingerprint")
	}
}

func TestNativeResumeDoesNotFenceDefaultConcurrentTasks(t *testing.T) {
	tasks := []*runningTask{{handle: &SessionHandle{}, provider: "codex", capabilityFingerprint: "default"}, {handle: &SessionHandle{nativeResume: &NativeResumeRuntime{}}, provider: "codex", capabilityFingerprint: "retained"}}
	changed := Capability{Provider: "codex", CapabilityFingerprint: "new-default"}
	if !taskPinnedToDriftedCapability(tasks[0], changed) {
		t.Fatal("default drift lost fencing")
	}
	if taskPinnedToDriftedCapability(tasks[1], changed) {
		t.Fatal("default drift fenced another installed CLI")
	}
}

func TestNativeResumeContinuationRetainsSelection(t *testing.T) {
	a := &Actor{mailbox: make(chan envelope, 1), stoppedCh: make(chan struct{})}
	selected := &NativeResumeRuntime{owner: a}
	previous := &SessionHandle{TaskID: "task", nativeResume: selected}
	done := make(chan error, 1)
	go func() {
		_, err := a.SubmitContinuation(context.Background(), bridge.TaskRequest{ID: "task"}, previous)
		done <- err
	}()
	msg := <-a.mailbox
	if msg.submit.nativeResume != selected {
		t.Fatal("continuation lost executable selection")
	}
	msg.submit.reply <- submitReply{}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
