package securecredential

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// A helper session keeps one verified helper process for the rest of one CLI
// command. The helper confirms the OS user once and reuses that confirmation
// for later credential and installation-key requests in the same command, so
// a command asks for Touch ID or the password once instead of once per
// keychain access. The helper bounds its own request count and lifetime and
// exits when this process closes its stdin (including on process exit).
type helperSession struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	broken bool
}

type sessionConn interface {
	roundTrip(ctx context.Context, body []byte) ([]byte, error)
	close()
}

var (
	sessionsMu sync.Mutex
	sessions   = map[string]sessionConn{}
	// Tests replace the starter; production starts the verified helper.
	startSession = startVerifiedSession
	// Sessions need the macOS helper; other platforms keep per-request calls.
	sessionSupported = runtime.GOOS == "darwin"
)

func sharedSession(ctx context.Context, path string) (sessionConn, error) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	if current, ok := sessions[path]; ok {
		return current, nil
	}
	started, err := startSession(ctx, path)
	if err != nil {
		return nil, ErrUnavailable
	}
	sessions[path] = started
	return started, nil
}

func dropSession(path string, conn sessionConn) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	if sessions[path] == conn {
		delete(sessions, path)
	}
	conn.close()
}

// CloseSessions ends every helper session of this process. A command calls it
// when it finishes; process exit has the same effect through stdin EOF.
func CloseSessions() {
	sessionsMu.Lock()
	current := sessions
	sessions = map[string]sessionConn{}
	sessionsMu.Unlock()
	for _, conn := range current {
		conn.close()
	}
}

// Keep the child unreaped while it serves requests so the verified PID cannot
// be reused. A prior path check is not a process identity check.
func startVerifiedSession(ctx context.Context, path string) (sessionConn, error) {
	command := exec.Command(path, "--session")
	command.Stderr = io.Discard
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, ErrUnavailable
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, ErrUnavailable
	}
	if err := command.Start(); err != nil {
		_ = stdin.Close()
		return nil, ErrUnavailable
	}
	session := &helperSession{cmd: command, stdin: stdin, stdout: bufio.NewReaderSize(io.LimitReader(stdout, 1<<30), 4096)}
	verifyCtx, cancel := context.WithTimeout(ctx, helperTimeout)
	defer cancel()
	if verifyHelperProcess(verifyCtx, command.Process.Pid) != nil || verifyCtx.Err() != nil {
		session.close()
		return nil, ErrUnavailable
	}
	return session, nil
}

func (h *helperSession) roundTrip(ctx context.Context, body []byte) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.broken || bytes.IndexByte(body, '\n') >= 0 {
		return nil, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, helperTimeout)
	defer cancel()
	type result struct {
		line []byte
		err  error
	}
	done := make(chan result, 1)
	go func() {
		if _, err := h.stdin.Write(append(append([]byte(nil), body...), '\n')); err != nil {
			done <- result{err: err}
			return
		}
		line, err := readBoundedLine(h.stdout, maxMessageBytes)
		done <- result{line: line, err: err}
	}()
	select {
	case got := <-done:
		if got.err != nil {
			h.breakLocked()
			return nil, ErrUnavailable
		}
		return got.line, nil
	case <-ctx.Done():
		h.breakLocked()
		<-done
		return nil, ctx.Err()
	}
}

func readBoundedLine(reader *bufio.Reader, limit int) ([]byte, error) {
	var line []byte
	for {
		part, err := reader.ReadSlice('\n')
		line = append(line, part...)
		if len(line) > limit+1 {
			return nil, ErrUnavailable
		}
		if err == nil {
			return line, nil
		}
		if !errors.Is(err, bufio.ErrBufferFull) {
			return nil, err
		}
	}
}

func (h *helperSession) close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.breakLocked()
}

// Breaking kills the helper; the next request starts a new session (and asks
// for OS confirmation again) rather than reusing an unknown state.
func (h *helperSession) breakLocked() {
	if h.broken {
		return
	}
	h.broken = true
	_ = h.stdin.Close()
	if h.cmd.Process != nil {
		_ = h.cmd.Process.Kill()
	}
	waited := make(chan struct{})
	go func() { _ = h.cmd.Wait(); close(waited) }()
	select {
	case <-waited:
	case <-time.After(time.Second):
	}
}
