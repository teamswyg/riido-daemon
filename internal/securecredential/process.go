package securecredential

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"time"
)

// limitedOutput always reports a full write to the child while bounding memory.
// Helper diagnostics and native error text are deliberately never returned.
type limitedOutput struct {
	bytes.Buffer
	overflow bool
}

func (b *limitedOutput) Write(p []byte) (int, error) {
	size := len(p)
	remaining := maxMessageBytes - b.Len()
	if len(p) > remaining {
		b.overflow = true
		p = p[:remaining]
	}
	_, _ = b.Buffer.Write(p)
	return size, nil
}

func runHelper(ctx context.Context, path string, input []byte) ([]byte, error) {
	return runVerifiedHelper(ctx, path, input, verifyHelperProcess)
}

// Keep the child unreaped until authentication and request delivery finish so
// the PID cannot be reused. A prior path check is not a process identity check.
func runVerifiedHelper(ctx context.Context, path string, input []byte, verify func(context.Context, int) error) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, path)
	command.WaitDelay = time.Second
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, ErrUnavailable
	}
	var out limitedOutput
	command.Stdout = &out
	command.Stderr = io.Discard
	if err := command.Start(); err != nil {
		_ = stdin.Close()
		return nil, ErrUnavailable
	}
	reaped := false
	defer func() {
		_ = stdin.Close()
		if !reaped {
			_ = command.Process.Kill()
			_ = command.Wait()
		}
	}()
	if verify == nil || verify(ctx, command.Process.Pid) != nil || ctx.Err() != nil {
		return nil, ErrUnavailable
	}
	if n, err := stdin.Write(input); err != nil || n != len(input) {
		return nil, ErrUnavailable
	}
	if stdin.Close() != nil {
		return nil, ErrUnavailable
	}
	err = command.Wait()
	reaped = true
	if err != nil || out.overflow {
		return nil, ErrUnavailable
	}
	return out.Bytes(), nil
}
