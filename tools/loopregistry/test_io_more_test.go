package main

import (
	"bytes"
	"os"
	"testing"
)

func captureStderr(t *testing.T) (func(), func() string) {
	t.Helper()
	original := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = writer
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = buf.ReadFrom(reader)
		close(done)
	}()
	restore := func() {
		_ = writer.Close()
		os.Stderr = original
		<-done
		_ = reader.Close()
	}
	return restore, buf.String
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}
