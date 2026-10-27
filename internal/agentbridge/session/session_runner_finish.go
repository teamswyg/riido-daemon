package session

func (r *sessionRunner) finish() {
	r.sess.childExitObserved = r.finishProcess()
	for _, ev := range cleanupTempFiles(r.cfg.TempFiles) {
		r.emit(ev)
	}
	r.sess.result <- r.finalResult()
}
