package session

func (r *sessionRunner) finish() {
	r.sess.childExitObserved = r.finishProcess()
	r.sess.claimAttempt.Exited(r.sess.childExitObserved)
	for _, ev := range cleanupTempFiles(r.cfg.TempFiles) {
		r.emit(ev)
	}
	r.sess.result <- r.finalResult()
}
