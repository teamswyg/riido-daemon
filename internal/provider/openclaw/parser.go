package openclaw

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// parser handles OpenClaw's two output modes:
//
//  1. Full JSON result on stdout: one object emitted once at the end of the
//     run, either compact or pretty-printed. The parser buffers stdout until
//     Close, then tries to decode the whole buffer as a single JSON object →
//     RawEvent of Type "full_result".
//
//  2. NDJSON streaming: line-delimited JSON events. When we see a
//     line-terminated JSON object during Feed, we emit a RawEvent of
//     Type "ndjson:<event>" eagerly.
//
// The mode is detected per-line: any line that parses as JSON during
// FeedStdout is treated as NDJSON. If no lines are emitted during the
// run (i.e. all output came in one chunk with no trailing newline) the
// buffer is decoded once on Close.
//
// This matches the “안정 호환은 full JSON 우선, 실패 시 NDJSON fallback”
// pattern from spec §3.3.
type parser struct {
	fullStdoutBuf []byte
	ndjsonLineBuf []byte
	stderrBuf     []byte
	emittedNDJSON bool
}

func NewParser() agentbridge.Parser { return &parser{} }
