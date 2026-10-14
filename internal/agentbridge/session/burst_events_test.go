package session

func countDrainedEvents(sess *Session, drained chan<- int) {
	count := 0
	for range sess.Events() {
		count++
	}
	drained <- count
}
