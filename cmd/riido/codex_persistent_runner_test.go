package main

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestCodexPersistentRunnerReusesAppServerAcrossSequentialTasks(t *testing.T) {
	running := process.NewFakeRunning()
	fake := process.NewFake()
	fake.NextRunning = running

	runner, err := newCodexPersistentRunner(runtimeactor.PersistentRunnerConfig{
		RuntimeID: "rt-codex",
		Adapter:   bridgeCodexAdapter{},
		Process:   fake,
		Now:       time.Now,
	})
	if err != nil {
		t.Fatalf("newCodexPersistentRunner: %v", err)
	}
	defer runner.Stop(context.Background())

	firstCh := make(chan agentbridge.RunHandle, 1)
	errCh := make(chan error, 1)
	go func() {
		h, err := runner.Submit(context.Background(), agentbridge.StartRequest{
			TaskID: "task-1",
			Prompt: "first",
			Cwd:    "/workspace/repo",
			Model:  "gpt-5.5",
		}, runtimeactor.PersistentRunOptions{})
		if err != nil {
			errCh <- err
			return
		}
		firstCh <- h
	}()

	assertNextMethod(t, running, "initialize")
	running.EmitStdout([]byte(`{"jsonrpc":"2.0","id":1,"result":{"server":"codex"}}` + "\n"))
	assertNextMethod(t, running, "initialized")

	first := recvHandle(t, firstCh, errCh)
	threadFrame := assertNextMethod(t, running, "thread/start")
	running.EmitStdout([]byte(`{"jsonrpc":"2.0","id":` + jsonFrameID(t, threadFrame) + `,"result":{"thread":{"id":"th-1"}}}` + "\n"))
	turnFrame := assertNextMethod(t, running, "turn/start")
	running.EmitStdout([]byte(`{"jsonrpc":"2.0","id":` + jsonFrameID(t, turnFrame) + `,"result":{}}` + "\n"))
	running.EmitStdout([]byte(`{"jsonrpc":"2.0","method":"turn_completed","params":{"output":"done"}}` + "\n"))
	res := recvResult(t, first)
	if res.Status != agentbridge.ResultCompleted || res.SessionID != "th-1" {
		t.Fatalf("first result = %+v", res)
	}
	assertNoKill(t, running)

	secondCh := make(chan agentbridge.RunHandle, 1)
	go func() {
		h, err := runner.Submit(context.Background(), agentbridge.StartRequest{
			TaskID: "task-2",
			Prompt: "second",
			Cwd:    "/workspace/repo",
			Model:  "gpt-5.5",
		}, runtimeactor.PersistentRunOptions{})
		if err != nil {
			errCh <- err
			return
		}
		secondCh <- h
	}()

	second := recvHandle(t, secondCh, errCh)
	threadFrame = assertNextMethod(t, running, "thread/start")
	if strings.Contains(threadFrame, `"method":"initialize"`) {
		t.Fatalf("second task reinitialized app-server: %s", threadFrame)
	}
	running.EmitStdout([]byte(`{"jsonrpc":"2.0","id":` + jsonFrameID(t, threadFrame) + `,"result":{"thread":{"id":"th-2"}}}` + "\n"))
	turnFrame = assertNextMethod(t, running, "turn/start")
	running.EmitStdout([]byte(`{"jsonrpc":"2.0","id":` + jsonFrameID(t, turnFrame) + `,"result":{}}` + "\n"))
	running.EmitStdout([]byte(`{"jsonrpc":"2.0","method":"turn_completed","params":{"output":"done"}}` + "\n"))
	res = recvResult(t, second)
	if res.Status != agentbridge.ResultCompleted || res.SessionID != "th-2" {
		t.Fatalf("second result = %+v", res)
	}
}

func jsonFrameID(t *testing.T, frame string) string {
	t.Helper()
	var payload struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(frame), &payload); err != nil {
		t.Fatalf("parse json frame: %v: %s", err, frame)
	}
	return strconvFormatInt(payload.ID)
}

func strconvFormatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}

func recvHandle(t *testing.T, ch <-chan agentbridge.RunHandle, errCh <-chan error) agentbridge.RunHandle {
	t.Helper()
	select {
	case h := <-ch:
		return h
	case err := <-errCh:
		t.Fatalf("submit failed: %v", err)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for submit")
	}
	return nil
}

func recvResult(t *testing.T, h agentbridge.RunHandle) agentbridge.Result {
	t.Helper()
	select {
	case res := <-h.Result():
		return res
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for result")
	}
	return agentbridge.Result{}
}

func assertNextMethod(t *testing.T, running *process.FakeRunning, method string) string {
	t.Helper()
	select {
	case b := <-running.StdinRecv():
		frame := string(b)
		if !strings.Contains(frame, `"method":"`+method+`"`) {
			t.Fatalf("expected method %s, got %s", method, frame)
		}
		return frame
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for method %s", method)
	}
	return ""
}

func assertNoKill(t *testing.T, running *process.FakeRunning) {
	t.Helper()
	select {
	case <-running.KillRecv():
		t.Fatal("persistent app-server was killed after completed run")
	default:
	}
}
