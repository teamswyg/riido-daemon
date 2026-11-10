package session

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const startupDiagnosticTimeout = 250 * time.Millisecond
const startupDiagnosticBytes = 64 * 1024

// The child can reject startup before accepting stdin. Read its explanation
// before failure teardown discards the streams; never resume a failed protocol.
func (r *sessionRunner) startFailureMessage(cause error) string {
	diagnostic := startupDiagnostic{}
	deadline := time.NewTimer(startupDiagnosticTimeout)
	defer deadline.Stop()
	remaining := startupDiagnosticBytes
drain:
	for remaining > 0 && (r.stdoutCh != nil || r.stderrCh != nil) {
		if r.ctx.Err() != nil {
			break
		}
		select {
		case <-r.ctx.Done():
			break drain
		case <-deadline.C:
			break drain
		case <-r.idleC:
			r.emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "semantic idle timeout"})
			return cause.Error()
		case chunk, ok := <-r.stdoutCh:
			if !ok {
				r.stdoutCh = nil
				continue
			}
			chunk = chunk[:min(len(chunk), remaining)]
			remaining -= len(chunk)
			raws, _ := r.parser.FeedStdout(chunk)
			diagnostic.collect(r.cfg.Adapter, raws)
		case chunk, ok := <-r.stderrCh:
			if !ok {
				r.stderrCh = nil
				continue
			}
			chunk = chunk[:min(len(chunk), remaining)]
			remaining -= len(chunk)
			raws, _ := r.parser.FeedStderr(chunk)
			diagnostic.collect(r.cfg.Adapter, raws)
		case status, ok := <-r.exitedCh:
			r.deferExit(status, ok)
		}
	}
	raws, _ := r.parser.Close()
	diagnostic.collect(r.cfg.Adapter, raws)
	if detail := diagnostic.detail(); detail != "" {
		return fmt.Sprintf("%s startup failed: %s (%s)", r.cfg.Adapter.Name(), detail, cause)
	}
	return cause.Error()
}
