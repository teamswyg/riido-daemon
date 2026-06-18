package claude

import (
	"strings"
	"testing"
)

func TestParserAcceptsLargeLine(t *testing.T) {
	p := NewParser()
	big := strings.Repeat("a", 9*1024*1024)
	chunk := `{"type":"tool_result","content":"` + big + `"}` + "\n"
	raws, err := p.FeedStdout([]byte(chunk))
	if err != nil {
		t.Fatalf("FeedStdout big line: %v", err)
	}
	if len(raws) != 1 || raws[0].Type != "tool_result" {
		t.Fatalf("big line: %+v", raws)
	}
}
