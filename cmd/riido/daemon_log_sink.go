package main

import (
	"io"
	"os"

	"github.com/teamswyg/riido-daemon/internal/logging"
)

// openLogSink returns a Logger port for structured log lines. File logging
// also mirrors to stderr so tests and operators can both observe startup.
func openLogSink(logFile string) (logging.Logger, func(), error) {
	if logFile == "" {
		return logging.NewWriterLogger(os.Stderr), func() {}, nil
	}
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}
	w := daemonLogWriter(os.Stderr, f)
	return logging.NewWriterLogger(w), func() { _ = f.Close() }, nil
}

func daemonLogWriter(stderr, file *os.File) io.Writer {
	stderrInfo, stderrErr := stderr.Stat()
	fileInfo, fileErr := file.Stat()
	if stderrErr == nil && fileErr == nil && os.SameFile(stderrInfo, fileInfo) {
		return file
	}
	return io.MultiWriter(stderr, file)
}
