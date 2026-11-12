package session

import (
	"fmt"
	"log"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (r *sessionRunner) finish() {
	// Clean up shared native queue work before observing the child exit. Keep
	// claim ownership fenced until the existing exit observer confirms it.
	_ = r.terminateProtocol(r.ctx)
	r.sess.childExitObserved = r.finishProcess()
	r.sess.claimAttempt.Exited(r.sess.childExitObserved)
	for _, ev := range cleanupTempFiles(r.cfg.TempFiles) {
		r.emit(ev)
	}
	if r.localDiagnosticsOmitted > 0 {
		text := fmt.Sprintf("provider diagnostic delivery limit reached, omitted=%d", r.localDiagnosticsOmitted)
		if !r.mirrorEvent(agentbridge.Event{Kind: agentbridge.EventLog, LocalOnly: true, At: r.cfg.Now(), Text: text}) {
			log.Print(text)
		}
	}
	r.sess.result <- r.finalResult()
}
