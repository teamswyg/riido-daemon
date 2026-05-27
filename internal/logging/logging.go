// Package logging defines the daemon logging port.
//
// It owns the minimal structured-line logger contract used by cmd/riido.
// It does NOT own log routing policy such as which file path to open or
// whether stderr should be fan-out; those are launch-surface decisions in
// cmd/riido.
package logging

import (
	"fmt"
	"io"
	"time"
)

// Logger is the port the local daemon depends on for structured log lines.
type Logger interface {
	Printf(format string, args ...any)
}

// WriterLogger adapts an io.Writer to Logger. It is intentionally tiny:
// the caller owns writer lifetime and routing.
type WriterLogger struct {
	w   io.Writer
	now func() time.Time
}

func NewWriterLogger(w io.Writer) *WriterLogger {
	return &WriterLogger{w: w, now: time.Now}
}

func (l *WriterLogger) Printf(format string, args ...any) {
	fmt.Fprintf(l.w, "[%s] ", l.now().UTC().Format(time.RFC3339Nano))
	fmt.Fprintf(l.w, format, args...)
	fmt.Fprintln(l.w)
}
