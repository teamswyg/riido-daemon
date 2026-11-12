package codex

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process/childenv"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

// Uses a real Codex binary, an isolated home, and a localhost fake model. No
// credentials, user's desktop threads, provider billing or external tools.
func TestIntegrationSharedOwnerStartsQueuedFollowup(t *testing.T) {
	exe := os.Getenv("RIIDO_CODEX_NATIVE_TEST_EXECUTABLE")
	if exe == "" {
		t.Skip("set RIIDO_CODEX_NATIVE_TEST_EXECUTABLE to a Codex binary")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 25*time.Second)
	defer cancel()
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, nativeSharedModelResponse)
	}))
	defer model.Close()
	home := t.TempDir()
	config := fmt.Sprintf("model_provider = 'fixture'\nmodel = 'gpt-5'\n[model_providers.fixture]\nname = 'fixture'\nbase_url = '%s/v1'\nwire_api = 'responses'\nrequires_openai_auth = false\n", model.URL)
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	req := agentbridge.StartRequest{Executable: exe, TaskID: "native-followup", Cwd: home, Prompt: "Reply OK only.", Env: map[string]string{"CODEX_HOME": home}}
	owner := exec.CommandContext(ctx, exe, "app-server", "--listen", "unix://")
	owner.Env = append(childenv.Environ(), "CODEX_HOME="+home)
	if err := owner.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = owner.Process.Kill(); _ = owner.Wait() }()
	var client *sharedOwnerClient
	var err error
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		client, err = openSharedOwner(ctx, req)
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("owner socket: %v", err)
	}
	defer client.close()
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
		Spawn: processCommandFromStart(spawn), Request: req, ProtocolDriver: driver,
		SemanticIdle: 4 * time.Second, HardTimeout: 15 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	go drainCodexIntegrationEvents(sess)
	result := <-sess.Result()
	if result.Status != agentbridge.ResultCompleted || result.SessionID != req.ResumeSessionID || strings.TrimSpace(result.Output) != "OK" {
		t.Fatalf("queued follow-up did not complete on same native thread: %+v", result)
	}
	// The owner's connection survives observer teardown, and its policy did not
	// inherit the detached Riido child's full-access launch configuration.
	resumed, err := client.request("thread/resume", map[string]any{"threadId": req.ResumeSessionID})
	if err != nil {
		t.Fatal(err)
	}
	if resumed["approvalPolicy"] != "on-request" || stringField(mapField(resumed, "sandbox"), "type") != "readOnly" {
		t.Fatalf("owner execution policy changed: approval=%v sandbox=%v", resumed["approvalPolicy"], resumed["sandbox"])
	}
	history, err := client.request("thread/turns/list", map[string]any{"threadId": req.ResumeSessionID, "itemsView": "full"})
	if err != nil || len(history["data"].([]any)) != 2 {
		t.Fatalf("follow-up duplicated or missing: %v %v", history, err)
	}
}

const nativeSharedModelResponse = `event: response.created
data: {"type":"response.created","response":{"id":"fixture","object":"response","status":"in_progress","output":[]}}

event: response.output_item.added
data: {"type":"response.output_item.added","output_index":0,"item":{"id":"fixture-message","type":"message","role":"assistant","status":"in_progress","content":[]}}

event: response.output_text.delta
data: {"type":"response.output_text.delta","item_id":"fixture-message","output_index":0,"content_index":0,"delta":"OK"}

event: response.output_item.done
data: {"type":"response.output_item.done","output_index":0,"item":{"id":"fixture-message","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"OK","annotations":[]}]}}

event: response.completed
data: {"type":"response.completed","response":{"id":"fixture","object":"response","status":"completed","output":[{"id":"fixture-message","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"OK","annotations":[]}]}],"usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2}}}

`
