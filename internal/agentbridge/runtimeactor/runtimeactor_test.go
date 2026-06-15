package runtimeactor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// --- Test scaffolding ---

type stubAdapter struct {
	name       string
	detected   agentbridge.DetectResult
	startReqCh chan agentbridge.StartRequest
}

func (a *stubAdapter) Name() string { return a.name }
func (a *stubAdapter) Detect(_ context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return a.detected, nil
}

func (a *stubAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	if a.startReqCh != nil {
		select {
		case a.startReqCh <- req:
		default:
		}
	}
	exe := req.Executable
	if exe == "" {
		exe = a.name
	}
	return agentbridge.StartCommand{Executable: exe}, nil
}
func (a *stubAdapter) NewParser() agentbridge.Parser { return &stubParser{} }
func (a *stubAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "chunk" {
		return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultCompleted, Output: string(raw.Bytes)}}}, nil, nil
	}
	return nil, nil, nil
}
func (a *stubAdapter) BlockedArgs() []string { return nil }

type stubParser struct{}

func (p *stubParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "chunk", Bytes: chunk}}, nil
}
func (p *stubParser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) { return nil, nil }
func (p *stubParser) Close() ([]agentbridge.RawEvent, error)                  { return nil, nil }

// fakeProcess is a channel-owned multi-spawn fake. Single goroutine
// (the producer loop) owns the slice; Start, at, count send messages.
// No mutex — same discipline as the runtime actor mailbox boundary.
type fakeProcess struct {
	startCh chan startReq
	atCh    chan atReq
	cmdCh   chan cmdReq
	cntCh   chan chan int
}

type startReq struct {
	cmd   process.Command
	reply chan *process.FakeRunning
}

type atReq struct {
	idx   int
	reply chan *process.FakeRunning
}

type cmdReq struct {
	idx   int
	reply chan process.Command
}

func newFakeProcess() *fakeProcess {
	f := &fakeProcess{
		startCh: make(chan startReq, 8),
		atCh:    make(chan atReq, 8),
		cmdCh:   make(chan cmdReq, 8),
		cntCh:   make(chan chan int, 4),
	}
	go f.run()
	return f
}

func (f *fakeProcess) run() {
	var produced []*process.FakeRunning
	var commands []process.Command
	for {
		select {
		case msg := <-f.startCh:
			r := process.NewFakeRunning()
			produced = append(produced, r)
			commands = append(commands, msg.cmd)
			msg.reply <- r
		case msg := <-f.atCh:
			if msg.idx >= len(produced) {
				msg.reply <- nil
			} else {
				msg.reply <- produced[msg.idx]
			}
		case msg := <-f.cmdCh:
			if msg.idx >= len(commands) {
				msg.reply <- process.Command{}
			} else {
				msg.reply <- commands[msg.idx]
			}
		case reply := <-f.cntCh:
			reply <- len(produced)
		}
	}
}

func (f *fakeProcess) Start(_ context.Context, cmd process.Command) (process.RunningProcess, error) {
	reply := make(chan *process.FakeRunning, 1)
	f.startCh <- startReq{cmd: cmd, reply: reply}
	return <-reply, nil
}

func (f *fakeProcess) at(i int) *process.FakeRunning {
	reply := make(chan *process.FakeRunning, 1)
	f.atCh <- atReq{idx: i, reply: reply}
	return <-reply
}

func (f *fakeProcess) count() int {
	reply := make(chan int, 1)
	f.cntCh <- reply
	return <-reply
}

func (f *fakeProcess) commandAt(i int) process.Command {
	reply := make(chan process.Command, 1)
	f.cmdCh <- cmdReq{idx: i, reply: reply}
	return <-reply
}

func waitForRunning(t *testing.T, p *fakeProcess, i int, d time.Duration) *process.FakeRunning {
	t.Helper()
	end := time.Now().Add(d)
	for time.Now().Before(end) {
		if r := p.at(i); r != nil {
			return r
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("fake process #%d not created within %v", i, d)
	return nil
}

func envListValue(env []string, wantKey string) (string, bool) {
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if ok && strings.EqualFold(key, wantKey) {
			return value, true
		}
	}
	return "", false
}
