package session

func emitCancelableBurst(running burstStdoutEmitter, chunks int, done chan<- struct{}) {
	defer close(done)
	for range chunks {
		safeEmitStdout(running, []byte("x"))
	}
}

func closeAfterEvents(sess *Session, done chan<- struct{}) {
	for range sess.Events() {
	}
	close(done)
}
