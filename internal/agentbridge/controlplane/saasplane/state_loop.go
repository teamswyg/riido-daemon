package saasplane

func (p *Plane) loop() {
	state := newPlaneState()
	defer close(p.done)
	for op := range p.ops {
		if op.close {
			closePlaneState(op, state)
			return
		}
		op.fn(&state)
		close(op.ack)
	}
}

func closePlaneState(op stateOp, state planeState) {
	for _, ch := range state.cancelWatchers {
		close(ch)
	}
	close(op.ack)
}
