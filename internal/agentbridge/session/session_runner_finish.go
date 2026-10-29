package session

func (r *sessionRunner) finish() {
	// Also clean up queued work on provider/protocol failures that did not pass
	// through interruptDelivery. Completed drivers have nothing to remove.
	_ = r.terminateProtocol(r.ctx)
	_ = killProcess(r.ctx, r.proc, r.cfg.ProcessKillTimeout)
	drain(r.stdoutCh)
	drain(r.stderrCh)
	for _, ev := range cleanupTempFiles(r.cfg.TempFiles) {
		r.emit(ev)
	}
	r.sess.result <- r.finalResult()
}
