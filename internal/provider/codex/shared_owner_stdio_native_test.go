package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process/childenv"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

// Uses a real Codex binary, an isolated home, and a localhost fake model. No
// credentials, user's desktop threads, provider billing or external tools.
func TestIntegrationSharedStdioOwner(t *testing.T) {
	for _, release := range []bool{true, false} {
		t.Run(map[bool]string{true: "thread_released_without_exit", false: "retained_owner_hard_deadline"}[release], func(t *testing.T) {
			testNativeStdioOwner(t, release)
		})
	}
}

func testNativeStdioOwner(t *testing.T, release bool) {
	t.Helper()
	exe := os.Getenv("RIIDO_CODEX_NATIVE_TEST_EXECUTABLE")
	if exe == "" {
		t.Skip("set RIIDO_CODEX_NATIVE_TEST_EXECUTABLE to a Codex binary")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 25*time.Second)
	defer cancel()
	var modelCalls atomic.Int32
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		modelCalls.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, nativeSharedModelResponse)
	}))
	defer model.Close()
	home := t.TempDir()
	config := fmt.Sprintf("thread_unload_delay_secs = 0\nmodel_provider = 'fixture'\nmodel = 'gpt-5'\n[model_providers.fixture]\nname = 'fixture'\nbase_url = '%s/v1'\nwire_api = 'responses'\nrequires_openai_auth = false\n", model.URL)
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	req := agentbridge.StartRequest{Executable: exe, TaskID: "native-followup", Cwd: home, Prompt: "Reply OK only.", Env: map[string]string{"CODEX_HOME": home}}
	owner := exec.CommandContext(ctx, exe, "app-server", "--listen", "stdio://")
	owner.Env = append(childenv.Environ(), "CODEX_HOME="+home)
	ownerIn, err := owner.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	ownerOut, err := owner.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := owner.Start(); err != nil {
		t.Fatal(err)
	}
	var stopOnce sync.Once
	stopOwner := func() { stopOnce.Do(func() { _ = owner.Process.Kill(); _ = owner.Wait() }) }
	defer stopOwner()
	client := &nativeStdioFixture{in: ownerIn, out: bufio.NewScanner(ownerOut)}
	client.out.Buffer(make([]byte, 4096), 16*1024*1024)
	_, err = client.request("initialize", map[string]any{"clientInfo": map[string]any{"name": "isolated-test", "version": "1"}, "capabilities": map[string]any{"experimentalApi": true}})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(ownerIn, "{\"method\":\"initialized\"}\n")
	created, err := client.request("thread/start", map[string]any{"cwd": home, "approvalPolicy": "on-request", "sandbox": "read-only", "persistExtendedHistory": true})
	if err != nil {
		t.Fatal(err)
	}
	req.ResumeSessionID = stringField(mapField(created, "thread"), "id")
	if req.ResumeSessionID == "" {
		t.Fatal("no native thread ID")
	}
	_, err = client.request("turn/start", map[string]any{"threadId": req.ResumeSessionID, "input": []map[string]any{{"type": "text", "text": "Reply OK only."}}})
	if err != nil {
		t.Fatal(err)
	}
	for deadline := time.Now().Add(5 * time.Second); ; {
		history, readErr := client.request("thread/turns/list", map[string]any{"threadId": req.ResumeSessionID, "itemsView": "full"})
		turns, _ := history["data"].([]any)
		if readErr == nil && len(turns) == 1 && sharedTurnFinished(turns[0].(map[string]any)) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("initial native turn did not finish: %v", readErr)
		}
		time.Sleep(50 * time.Millisecond)
	}
	// Model navigating away from the idle thread, not closing the owner app.
	// The original Riido request must remain alive until Codex releases its lock.
	released := make(chan error, 1)
	if release {
		timer := time.AfterFunc(5*time.Second, func() {
			_, err := client.request("thread/unsubscribe", map[string]any{"threadId": req.ResumeSessionID})
			released <- err
		})
		defer timer.Stop()
	}
	result := runNativeSharedFollowup(t, ctx, req)
	if release {
		if err := <-released; err != nil {
			t.Fatal(err)
		}
		if result.Status != agentbridge.ResultCompleted || result.SessionID != req.ResumeSessionID || strings.TrimSpace(result.Output) != "OK" {
			t.Fatalf("original request did not automatically complete after thread release: %+v", result)
		}
	} else if result.Status != agentbridge.ResultTimeout || !strings.Contains(result.Error, "hard timeout") || result.SessionID != req.ResumeSessionID {
		t.Fatalf("retained owner must remain bounded by the hard deadline: %+v", result)
	}
	// The owner still answers RPCs after Riido has completed or timed out.
	queue, err := client.request("thread/queue/list", map[string]any{"threadId": req.ResumeSessionID})
	if err != nil || len(queue["data"].([]any)) != 0 {
		t.Fatalf("owner unavailable or pending request left behind: %v %v", queue, err)
	}
	history, err := client.request("thread/turns/list", map[string]any{"threadId": req.ResumeSessionID, "itemsView": "full"})
	want := 1
	if release {
		want = 2
	}
	if err != nil || len(history["data"].([]any)) != want || int(modelCalls.Load()) != want {
		t.Fatalf("request duplicated or missing: turns=%v modelCalls=%d err=%v", history, modelCalls.Load(), err)
	}
}

func runNativeSharedFollowup(t *testing.T, ctx context.Context, req agentbridge.StartRequest) agentbridge.Result {
	t.Helper()
	spawn, err := BuildStart(req, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	driver, err := NewProtocolDriver(req)
	if err != nil {
		t.Fatal(err)
	}
	sess, err := session.Start(ctx, session.Config{
		TaskID: req.TaskID, RuntimeID: "codex", Adapter: integrationAdapter{}, Process: processexec.New(),
		Spawn: processCommandFromStart(spawn), Request: req, ProtocolDriver: driver, SemanticIdle: 4 * time.Second, HardTimeout: 15 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	go drainCodexIntegrationEvents(sess)
	return <-sess.Result()
}

type nativeStdioFixture struct {
	in  io.Writer
	out *bufio.Scanner
	id  int64
}

func (c *nativeStdioFixture) request(method string, params map[string]any) (map[string]any, error) {
	c.id++
	b, _ := json.Marshal(map[string]any{"id": c.id, "method": method, "params": params})
	if _, err := c.in.Write(append(b, '\n')); err != nil {
		return nil, err
	}
	for c.out.Scan() {
		var frame map[string]any
		if err := json.Unmarshal(c.out.Bytes(), &frame); err != nil {
			return nil, err
		}
		if id, ok := rpcID(frame); !ok || id != c.id {
			continue
		}
		if frame["error"] != nil {
			return nil, fmt.Errorf("native fixture RPC: %v", frame["error"])
		}
		return frame["result"].(map[string]any), nil
	}
	if err := c.out.Err(); err != nil {
		return nil, fmt.Errorf("native fixture ended: %w", err)
	}
	return nil, io.EOF
}
