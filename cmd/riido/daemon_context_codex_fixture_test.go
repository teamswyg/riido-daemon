package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
)

type codexContextClient struct {
	stdin  io.WriteCloser
	frames <-chan map[string]any
	nextID int
	stop   func()
}

func newCodexContextClient(t *testing.T, exe, home string) *codexContextClient {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	cmd := exec.CommandContext(ctx, exe, "app-server", "--listen", "stdio://")
	cmd.Env = []string{"HOME=" + home, "CODEX_HOME=" + home, "PATH=" + os.Getenv("PATH"), "UNRELATED_INHERITED=must-not-pass"}
	cmd.Dir = home
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	frames := make(chan map[string]any, 100)
	go func() {
		defer close(frames)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 4096), 4<<20)
		for scanner.Scan() {
			var frame map[string]any
			if json.Unmarshal(scanner.Bytes(), &frame) != nil {
				continue
			}
			select {
			case frames <- frame:
			case <-ctx.Done():
				return
			}
		}
	}()
	var once sync.Once
	c := &codexContextClient{stdin: stdin, frames: frames, stop: func() { once.Do(func() { cancel(); _ = stdin.Close(); _ = cmd.Wait() }) }}
	t.Cleanup(c.stop)
	c.call(t, "initialize", map[string]any{"clientInfo": map[string]any{"name": "riido-context-fixture", "version": "1"}})
	c.write(t, map[string]any{"method": "initialized", "params": map[string]any{}})
	return c
}

func (c *codexContextClient) write(t *testing.T, frame map[string]any) {
	t.Helper()
	if json.NewEncoder(c.stdin).Encode(frame) != nil {
		t.Fatal("write native Codex RPC")
	}
}

func (c *codexContextClient) next(t *testing.T) map[string]any {
	t.Helper()
	select {
	case f, ok := <-c.frames:
		if !ok {
			t.Fatal("native Codex exited")
		}
		return f
	case <-time.After(15 * time.Second):
		t.Fatal("native Codex RPC timed out")
		return nil
	}
}

func (c *codexContextClient) call(t *testing.T, method string, params map[string]any) map[string]any {
	t.Helper()
	c.nextID++
	c.write(t, map[string]any{"id": c.nextID, "method": method, "params": params})
	for {
		f := c.next(t)
		if f["id"] != float64(c.nextID) {
			continue
		}
		if f["error"] != nil {
			t.Fatalf("native Codex rejected %s", method)
		}
		return f["result"].(map[string]any)
	}
}

func (c *codexContextClient) shell(t *testing.T, thread, command string) string {
	t.Helper()
	c.call(t, "thread/shellCommand", map[string]any{"threadId": thread, "command": command, "timeoutMs": 5000})
	for {
		frame := c.next(t)
		if frame["method"] != "item/completed" {
			continue
		}
		item := frame["params"].(map[string]any)["item"].(map[string]any)
		out, ok := item["aggregatedOutput"].(string)
		if !ok {
			continue
		}
		if item["exitCode"] != float64(0) {
			var result struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			_ = json.Unmarshal([]byte(out), &result)
			t.Fatalf("native shell failed with exit %v: %s %s", item["exitCode"], result.Error.Code, result.Error.Message)
		}
		return out
	}
}

type codexContextCapture struct{ frames []map[string]any }

func (c *codexContextCapture) WriteStdin(_ context.Context, b []byte) error {
	var frame map[string]any
	if err := json.Unmarshal(b, &frame); err != nil {
		return err
	}
	c.frames = append(c.frames, frame)
	return nil
}
func (*codexContextCapture) CloseStdin(context.Context) error { return nil }

func codexContextThreadRequest(t *testing.T, req agentbridge.StartRequest) (string, map[string]any) {
	t.Helper()
	driver, err := codex.NewProtocolDriver(req)
	if err != nil {
		t.Fatal(err)
	}
	capture := &codexContextCapture{}
	if err := driver.OnStart(t.Context(), capture); err != nil {
		t.Fatal(err)
	}
	id := capture.frames[0]["id"]
	_, _, err = driver.OnRaw(t.Context(), agentbridge.RawEvent{Type: "response", Payload: map[string]any{"id": id, "result": map[string]any{}}}, capture)
	if err != nil {
		t.Fatal(err)
	}
	frame := capture.frames[len(capture.frames)-1]
	return frame["method"].(string), frame["params"].(map[string]any)
}

func quoteCodexContextCLI(path string) string {
	return "'" + strings.ReplaceAll(path, "'", "'\"'\"'") + "'"
}

func assertCodexContextSecretsNotPersisted(t *testing.T, root string, secrets []string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, secret := range secrets {
			if strings.Contains(string(data), secret) {
				t.Error("run capability persisted in Codex state")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
