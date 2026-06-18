package session

func (r *sessionRunner) finish() {
	_ = killProcess(r.ctx, r.proc, r.cfg.ProcessKillTimeout)
	drain(r.stdoutCh)
	drain(r.stderrCh)
	for _, ev := range cleanupTempFiles(r.cfg.TempFiles) {
		r.emit(ev)
	}
	r.sess.result <- r.finalResult()
}
