package process

func (r *FakeRunning) runEmitActor() {
	for event := range r.emit {
		switch event.kind {
		case fakeEmitStdout:
			select {
			case <-r.done:
				return
			case r.stdout <- event.bytes:
			}
		case fakeEmitStderr:
			select {
			case <-r.done:
				return
			case r.stderr <- event.bytes:
			}
		case fakeEmitExit:
			r.exited <- event.status
			close(r.stdout)
			close(r.stderr)
			close(r.exited)
			close(r.done)
			return
		}
	}
}
