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
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, path)
	command.Stdin = bytes.NewReader(input)
	var out limitedOutput
	command.Stdout = &out
	command.Stderr = io.Discard
	if err := command.Run(); err != nil || out.overflow {
		return nil, ErrUnavailable
	}
	return out.Bytes(), nil
}
