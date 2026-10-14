package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func noEventRecordingAdapter() *recordingAdapter {
	return &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			return nil, nil, nil
		},
	}
}
