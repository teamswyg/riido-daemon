package session

import (
	"bytes"
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type burstAdapter struct {
	done []byte
}

func (a *burstAdapter) Name() string { return "burst" }

func (a *burstAdapter) Detect(_ context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}

func (a *burstAdapter) BuildStart(_ agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return agentbridge.StartCommand{}, nil
}

func (a *burstAdapter) NewParser() agentbridge.Parser {
	done := a.done
	if len(done) == 0 {
		done = []byte("DONE")
	}
	return &burstParser{done: done}
}

func (a *burstAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	switch raw.Type {
	case "sentinel":
		result := agentbridge.Result{Status: agentbridge.ResultCompleted}
		return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: result}}, nil, nil
	case "chunk":
		event := agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)}
		return []agentbridge.Event{event}, nil, nil
	default:
		return nil, nil, nil
	}
}

func (a *burstAdapter) BlockedArgs() []string { return nil }

type burstParser struct {
	done []byte
}

func (p *burstParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	if bytes.Equal(chunk, p.done) {
		return []agentbridge.RawEvent{burstRaw(agentbridge.RawSourceStdout, "sentinel", chunk)}, nil
	}
	return []agentbridge.RawEvent{burstRaw(agentbridge.RawSourceStdout, "chunk", chunk)}, nil
}

func (p *burstParser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) {
	return []agentbridge.RawEvent{burstRaw(agentbridge.RawSourceStderr, "chunk", chunk)}, nil
}

func (p *burstParser) Close() ([]agentbridge.RawEvent, error) { return nil, nil }

func burstRaw(source agentbridge.RawSource, eventType string, chunk []byte) agentbridge.RawEvent {
	return agentbridge.RawEvent{Source: source, Type: eventType, Bytes: chunk}
}
