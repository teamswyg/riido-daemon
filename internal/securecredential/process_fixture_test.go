package securecredential

import (
	"bytes"
	"io"
	"os"
	"testing"
)

const (
	processFixtureMode    = "RIIDO_TEST_PROTECTED_HELPER_MODE"
	processFixtureCapture = "RIIDO_TEST_PROTECTED_HELPER_CAPTURE"
)

// The test executable provides a real pipe peer on every OS, without a shell
// or a native credential provider. Only synthetic test requests reach it.
func TestMain(m *testing.M) {
	if mode := os.Getenv(processFixtureMode); mode != "" {
		os.Exit(runProcessFixture(mode, os.Getenv(processFixtureCapture)))
	}
	os.Exit(m.Run())
}

func runProcessFixture(mode, capture string) int {
	if mode != "echo" && mode != "overflow" && mode != "failure" {
		return 2
	}
	file, err := os.OpenFile(capture, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return 2
	}
	defer func() { _ = file.Close() }()
	// Record each read immediately; buffering until EOF hides preapproval bytes
	// when the parent rejects and kills a child without closing its input.
	n, err := io.Copy(file, io.LimitReader(os.Stdin, maxMessageBytes+1))
	if err != nil || n > maxMessageBytes {
		return 2
	}
	switch mode {
	case "failure":
		_, _ = os.Stderr.WriteString("synthetic-sensitive-diagnostic")
		return 7
	case "overflow":
		_, err = os.Stdout.Write(bytes.Repeat([]byte("x"), maxMessageBytes+1))
	default:
		_, err = os.Stdout.WriteString("fixture-result")
	}
	if err != nil {
		return 2
	}
	return 0
}
