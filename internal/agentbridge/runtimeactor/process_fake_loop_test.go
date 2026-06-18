package runtimeactor

import "github.com/teamswyg/riido-daemon/internal/process"

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
