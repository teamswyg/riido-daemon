package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/cliui"
)

// viewNow is the clock for relative times; tests pin it.
var viewNow = time.Now

// terminalView returns a renderer only for a person at a terminal. Pipes,
// files, agents, TERM=dumb and RIIDO_CLI_PLAIN=1 get nil, so succeed keeps
// printing the same indented JSON as before.
func terminalView(w io.Writer, view func(io.Writer, cliui.Terminal)) textRenderer {
	term := humanTerminal(w)
	if !term.Interactive {
		return nil
	}
	return func(out io.Writer, _ any) { view(out, term) }
}

// viewField is one aligned metadata row. value is raw server text; empty rows
// are skipped.
type viewField struct {
	label string
	value string
}

func writeViewFields(w io.Writer, term cliui.Terminal, fields []viewField) {
	width := 0
	for _, field := range fields {
		if field.value != "" {
			width = max(width, cliui.DisplayWidth(field.label))
		}
	}
	for _, field := range fields {
		if field.value == "" {
			continue
		}
		value := oneLine(field.value)
		if strings.Contains(value, " ") {
			// Prose is shortened; IDs, emails and URLs stay whole so they can be copied.
			value = cliui.Truncate(value, max(term.Width-width-4, 12))
		}
		fmt.Fprintf(w, "  %s  %s\n", term.Dim(cliui.PadRight(field.label, width)), value)
	}
}

// writeViewHeader prints the bold title with a dim key, then a meta line.
// meta entries are already escaped and styled.
func writeViewHeader(w io.Writer, term cliui.Terminal, title, key string, meta ...string) {
	key = cliui.Safe(key)
	room := max(term.Width-cliui.DisplayWidth(key)-1, 12)
	line := term.Dim("(제목 없음)")
	if title = oneLine(title); title != "" {
		line = term.Bold(cliui.Truncate(title, room))
	}
	if key != "" {
		line += " " + term.Dim(key)
	}
	fmt.Fprintln(w, line)
	parts := make([]string, 0, len(meta))
	for _, item := range meta {
		if item != "" {
			parts = append(parts, item)
		}
	}
	if len(parts) > 0 {
		fmt.Fprintln(w, strings.Join(parts, term.Dim(" · ")))
	}
}

// writeViewBody prints multi-line server text indented and wrapped to the
// terminal width. Every line is escaped before it is measured.
func writeViewBody(w io.Writer, term cliui.Terminal, text string, indent int) {
	pad := strings.Repeat(" ", indent)
	width := max(term.Width-indent, 20)
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\t", "    ")
	for line := range strings.SplitSeq(text, "\n") {
		for _, part := range cliui.Wrap(cliui.Safe(strings.TrimRight(line, " ")), width) {
			fmt.Fprintln(w, strings.TrimRight(pad+part, " "))
		}
	}
}

// writeViewSection prints a dim label and an indented body, or nothing.
func writeViewSection(w io.Writer, term cliui.Terminal, label, text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, term.Dim(label))
	writeViewBody(w, term, text, 2)
	return true
}

// oneLine escapes s and folds all whitespace runs into single spaces.
func oneLine(s string) string {
	return cliui.Safe(strings.Join(strings.Fields(s), " "))
}

// viewTime shows recent RFC 3339 times relatively and older ones as a date.
func viewTime(ts string) string {
	at, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return cliui.Safe(ts)
	}
	age := viewNow().Sub(at)
	switch {
	case age < 0 || age >= 30*24*time.Hour:
		return localTime(at).Format(time.DateOnly)
	case age < time.Minute:
		return "방금"
	case age < time.Hour:
		return fmt.Sprintf("%d분 전", int(age/time.Minute))
	case age < 24*time.Hour:
		return fmt.Sprintf("%d시간 전", int(age/time.Hour))
	}
	return fmt.Sprintf("%d일 전", int(age/(24*time.Hour)))
}

// viewDate shows an RFC 3339 value as a local date (and time when asked).
func viewDate(ts string, withTime bool) string {
	at, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return cliui.Safe(ts)
	}
	if withTime {
		return localTime(at).Format("2006-01-02 15:04")
	}
	return localTime(at).Format(time.DateOnly)
}

// localTime shows times in the reader's zone: these views are only for a
// person at their own terminal.
func localTime(at time.Time) time.Time {
	return at.Local() //nolint:gosmopolitan // a human terminal view shows the reader's wall clock
}

// viewState colors a work status: done green, cancelled dim, in progress in
// the brand color, anything else plain.
func viewState(term cliui.Terminal, name, kind string, done bool) string {
	label := oneLine(name)
	if label == "" {
		label = oneLine(kind)
	}
	norm := strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(kind))
	switch {
	case done || strings.Contains(norm, "done") || strings.Contains(norm, "complete"):
		return term.Success("● " + orText(label, "완료"))
	case strings.Contains(norm, "cancel"):
		return term.Dim("● " + orText(label, "취소"))
	case strings.Contains(norm, "progress") || strings.Contains(norm, "review"):
		return term.Accent("● " + orText(label, "진행 중"))
	}
	return "○ " + orText(label, "상태 없음")
}

// readStatusBadge describes how completely a live read returned its part.
func readStatusBadge(term cliui.Terminal, status string) string {
	switch status {
	case "available":
		return term.Success("● 전체")
	case "partial":
		return term.Warn("● 일부")
	case "empty":
		return term.Dim("○ 비어 있음")
	case "deferred":
		return term.Warn("● 나눠 읽기")
	case "denied":
		return term.Error("● 권한 없음")
	case "unavailable":
		return term.Error("● 읽을 수 없음")
	case "unsupported":
		return term.Error("● 지원하지 않음")
	}
	return oneLine(status)
}

func orText(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

// rawInt decodes a nullable JSON page number.
func rawInt(raw json.RawMessage) (int, bool) {
	var n *int
	if err := json.Unmarshal(raw, &n); err != nil || n == nil {
		return 0, false
	}
	return *n, true
}

// writeViewHint prints a dim next-step line. It is never shortened: hints carry
// versions and cursors that must be copied exactly.
func writeViewHint(w io.Writer, term cliui.Terminal, text string) {
	fmt.Fprintln(w, term.Dim(text))
}
