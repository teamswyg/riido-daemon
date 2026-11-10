package session

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

type startupDiagnostic struct {
	providerError string
	stderr        string
}

func (d *startupDiagnostic) collect(adapter agentbridge.Adapter, raws []agentbridge.RawEvent) {
	for _, raw := range raws {
		// Translation is pure. Do not call the failed driver's OnRaw or execute
		// returned commands: startup diagnostics cannot authorize provider work.
		events, _, _ := adapter.Translate(raw)
		for _, event := range events {
			switch event.Kind {
			case agentbridge.EventResult:
				if event.Result.Status != agentbridge.ResultCompleted && event.Result.Error != "" {
					d.providerError = safeStartupDetail(event.Result.Error)
				}
			case agentbridge.EventError:
				if event.Err != "" {
					d.providerError = safeStartupDetail(event.Err)
				}
			case agentbridge.EventLog:
				if raw.Source == agentbridge.RawSourceStderr {
					d.stderr = safeStartupDetail(d.stderr + "\n" + event.Text)
				}
			}
		}
	}
}

func (d startupDiagnostic) detail() string {
	if d.providerError != "" {
		return d.providerError
	}
	return d.stderr
}

func safeStartupDetail(text string) string {
	text, _ = policy.RedactSecretPatterns(text, nil)
	chars := []rune(strings.TrimSpace(text))
	if len(chars) > 2048 {
		chars = append([]rune("…"), chars[len(chars)-2048:]...)
	}
	return string(chars)
}
