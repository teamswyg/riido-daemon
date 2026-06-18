package agentbridge

// Parser owns the per-process line buffer state. Each Feed call returns zero
// or more RawEvent envelopes that Translate maps to run-scope Events.
type Parser interface {
	FeedStdout(chunk []byte) ([]RawEvent, error)
	FeedStderr(chunk []byte) ([]RawEvent, error)
	Close() ([]RawEvent, error)
}

// RawEvent is the post-parse, pre-translate envelope. The Type field is
// provider-specific and meaningful only to the adapter's Translate.
type RawEvent struct {
	Source  RawSource
	Type    string
	Payload map[string]any
	Bytes   []byte
}

// RawSource tags which stream produced the raw event.
type RawSource string

const (
	RawSourceStdout RawSource = "stdout"
	RawSourceStderr RawSource = "stderr"
	RawSourceClose  RawSource = "close"
)
