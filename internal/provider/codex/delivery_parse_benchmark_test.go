package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func BenchmarkAnswerDeltaParseAndTranslate(b *testing.B) {
	chunk := []byte("{\"jsonrpc\":\"2.0\",\"method\":\"item/agentMessage/delta\",\"params\":{\"delta\":\"한글 답변 조각\"}}\n")
	parser := NewParser()
	b.ReportAllocs()
	for b.Loop() {
		raws, err := parser.FeedStdout(chunk)
		if err != nil || len(raws) != 1 {
			b.Fatal("parse", err)
		}
		events, _, err := Translate(raws[0])
		if err != nil || len(events) != 1 || events[0].Kind != agentbridge.EventTextDelta {
			b.Fatal("translate", err)
		}
	}
}
