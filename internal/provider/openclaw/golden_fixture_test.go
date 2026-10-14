package openclaw

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type goldenFixtureTransform func([]byte) []byte

func goldenRawEvents(t *testing.T, name string, transform goldenFixtureTransform) []agentbridge.RawEvent {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	p := NewParser()
	feed, err := p.FeedStdout(transform(body))
	if err != nil {
		t.Fatalf("FeedStdout: %v", err)
	}
	closed, _ := p.Close()
	raws := append([]agentbridge.RawEvent{}, feed...)
	return append(raws, closed...)
}

func trimGoldenTrailingNewline(body []byte) []byte {
	return bytes.TrimRight(body, "\n")
}

func keepGoldenFixtureBytes(body []byte) []byte {
	return body
}
