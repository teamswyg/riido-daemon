package runtimeactor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

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

func waitForRunning(
	t *testing.T,
	p *fakeProcess,
	i int,
	d time.Duration,
) *process.FakeRunning {
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
