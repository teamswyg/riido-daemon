package session

func (r *sessionRunner) finish() {
	// Clean up shared native queue work before observing the child exit. Keep
	// claim ownership fenced until the existing exit observer confirms it.
	_ = r.terminateProtocol(r.ctx)
	r.sess.childExitObserved = r.finishProcess()
	r.sess.claimAttempt.Exited(r.sess.childExitObserved)
	for _, ev := range cleanupTempFiles(r.cfg.TempFiles) {
		r.emit(ev)
	}
	r.sess.result <- r.finalResult()
}
