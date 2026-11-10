package supervisor

const assignmentGitDiagnosticLimit = 64 * 1024

// Keep only the tail so large progress/helper output cannot consume unbounded
// memory or hide Git's final diagnostic. This buffer is never persisted.
type assignmentGitDiagnosticBuffer struct {
	data []byte
}

func (b *assignmentGitDiagnosticBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if n >= assignmentGitDiagnosticLimit {
		b.data = append(b.data[:0], p[n-assignmentGitDiagnosticLimit:]...)
		return n, nil
	}
	if overflow := len(b.data) + n - assignmentGitDiagnosticLimit; overflow > 0 {
		copy(b.data, b.data[overflow:])
		b.data = b.data[:len(b.data)-overflow]
	}
	b.data = append(b.data, p...)
	return n, nil
}
