package processexec

type streamWriter struct {
	out chan<- []byte
}

func (w streamWriter) Write(p []byte) (int, error) {
	chunk := make([]byte, len(p))
	copy(chunk, p)
	w.out <- chunk
	return len(p), nil
}
