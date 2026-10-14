package runtimeactor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/process"
)

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

func (f *fakeProcess) Start(
	_ context.Context,
	cmd process.Command,
) (process.RunningProcess, error) {
	reply := make(chan *process.FakeRunning, 1)
	f.startCh <- startReq{cmd: cmd, reply: reply}
	return <-reply, nil
}
