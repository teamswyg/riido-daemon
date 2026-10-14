package claude

import (
	"context"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestClaudePartialFlag(t *testing.T) {
	c, e := BuildStart(agentbridge.StartRequest{}, StartOptions{PermissionMode: PermissionModeApproval})
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(strings.Join(c.Args, " "), "--include-partial-messages") {
		t.Fatal("token streaming is not enabled by generated CLI args")
	}
}

func TestClaudeStreamDelta(t *testing.T) {
	raws := feedStdoutAll(t, NewParser(), `{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hello"}}}`+"\n")
	for _, r := range raws {
		es, _, e := Translate(r)
		if e != nil {
			t.Fatal(e)
		}
		for _, v := range es {
			if v.Kind == agentbridge.EventTextDelta && v.Text == "hello" {
				return
			}
		}
		t.Logf("events=%+v", es)
	}
	t.Fatal("documented stream_event text delta does not reach neutral text stream")
}

func TestStreamingMessageDoesNotDuplicateFinalBlocks(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	var text, thinking string
	frames := []string{
		`{"type":"stream_event","event":{"type":"message_start"}}`,
		`{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hello"}}}`,
		`{"type":"stream_event","event":{"type":"content_block_delta","index":1,"delta":{"type":"thinking_delta","thinking":"hmm"}}}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"hello"},{"type":"thinking","thinking":"hmm"}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"fallback"}]}}`,
	}
	for _, frame := range frames {
		for _, raw := range feedStdoutAll(t, NewParser(), frame+"\n") {
			events, _, err := d.OnRaw(context.Background(), raw, nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, event := range events {
				if event.Kind == agentbridge.EventTextDelta {
					text += event.Text
				}
				if event.Kind == agentbridge.EventThinkingDelta {
					thinking += event.Text
				}
			}
		}
	}
	if text != "hellofallback" || thinking != "hmm" {
		t.Fatalf("duplicated output: text=%q thinking=%q", text, thinking)
	}
}
