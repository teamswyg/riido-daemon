package processexec

import "errors"

func (r *execRunning) WriteStdin(b []byte) error {
	r.stdinMu.Lock()
	defer r.stdinMu.Unlock()
	if r.stdin == nil {
		return errors.New("processexec: stdin closed")
	}
	_, err := r.stdin.Write(b)
	return err
}

func (r *execRunning) CloseStdin() error {
	var err error
	r.stdinOnce.Do(func() {
		r.stdinMu.Lock()
		defer r.stdinMu.Unlock()
		if r.stdin != nil {
			err = r.stdin.Close()
			r.stdin = nil
		}
	})
	return err
}
