package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// textRenderer prints the data payload for people. nil falls back to JSON.
type textRenderer func(w io.Writer, data any)

// wantsJSON pre-scans raw args so even argument errors honor --json.
func wantsJSON(args []string) bool {
	for i, arg := range args {
		switch {
		case arg == "--json", arg == "--format=json":
			return true
		case arg == "--format" && i+1 < len(args) && args[i+1] == "json":
			return true
		}
	}
	return false
}

func (c contextIO) fail(jsonOut bool, requestID string, cliErr *contextcli.CLIError) error {
	if jsonOut {
		if err := writeEnvelope(c.stdout, contextcli.FailureEnvelope(requestID, cliErr)); err != nil {
			return err
		}
	} else {
		fmt.Fprintf(c.stderr, "riido: %s\n", cliErr.Message)
		if cliErr.NextAction != "" {
			fmt.Fprintf(c.stderr, "  다음 조치: %s\n", cliErr.NextAction)
		}
		fmt.Fprintf(c.stderr, "  (code=%s request_id=%s)\n", cliErr.Code, requestID)
	}
	return reportedExit{code: cliErr.ExitCode()}
}

func (c contextIO) succeed(jsonOut bool, env contextcli.Envelope, render textRenderer) error {
	if replyErr := contextcli.ReplyError(env); replyErr != nil {
		return c.fail(jsonOut, env.RequestID, replyErr)
	}
	if jsonOut {
		return writeEnvelope(c.stdout, env)
	}
	for _, warning := range env.Warnings {
		fmt.Fprintf(c.stderr, "riido: 경고 %s: %s\n", warning.Code, warning.Message)
	}
	if render != nil {
		render(c.stdout, env.Data)
		return nil
	}
	return writeIndentedJSON(c.stdout, env.Data)
}

func writeEnvelope(w io.Writer, env contextcli.Envelope) error {
	return writeIndentedJSON(w, env)
}

func writeIndentedJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode output: %w", err)
	}
	return nil
}
