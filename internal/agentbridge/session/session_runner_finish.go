package session

func (r *sessionRunner) finish() {
	// Also clean up queued work on provider/protocol failures that did not pass
	// through interruptDelivery. Completed drivers have nothing to remove.
	_ = r.terminateProtocol(r.ctx)
	r.sess.childExitObserved = r.finishProcess()
	r.sess.claimAttempt.Exited(r.sess.childExitObserved)
	for _, ev := range cleanupTempFiles(r.cfg.TempFiles) {
		r.emit(ev)
	}
	r.sess.result <- r.finalResult()
}
