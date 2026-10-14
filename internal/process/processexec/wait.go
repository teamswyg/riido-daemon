package processexec

import "github.com/teamswyg/riido-daemon/internal/process"

func (r *execRunning) waitExit() {
	err := r.cmd.Wait()
	close(r.done)
	r.cancel()
	close(r.stdout)
	close(r.stderr)
	code := r.cmd.ProcessState.ExitCode()
	code = normalizeExitCode(code, err)
	r.exited <- process.ExitStatus{Code: code, Err: err}
	close(r.exited)
}
