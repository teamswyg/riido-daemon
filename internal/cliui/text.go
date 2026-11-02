package cliui

import (
	"html"
	"strings"
)

// blockTags end a line in HTMLText; listTags start a bullet.
var blockTags = map[string]bool{
	"p": true, "div": true, "br": true, "hr": true, "ul": true, "ol": true, "li": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"blockquote": true, "pre": true, "tr": true, "table": true,
}

// HTMLText turns server HTML (comments, documents) into plain text for a
// terminal: block tags become line breaks, list items become bullets, other
// tags are dropped and entities are decoded. It approximates layout only; it
// is not a sanitizer, so callers still pass every line through Safe.
func HTMLText(s string) string {
	var b strings.Builder
	for s != "" {
		open := strings.IndexByte(s, '<')
		if open < 0 {
			b.WriteString(html.UnescapeString(s))
			break
		}
		b.WriteString(html.UnescapeString(s[:open]))
		s = s[open:]
		name, closing, ok := tagName(s)
		if !ok {
			b.WriteByte('<')
			s = s[1:]
			continue
		}
		end := strings.IndexByte(s, '>')
		if end < 0 {
			// No later '>' means no later tag: the rest is plain text.
			b.WriteString(html.UnescapeString(s))
			break
		}
		switch {
		case name == "li" && !closing:
			b.WriteString("\n• ")
		case name != "li" && blockTags[name]:
			b.WriteByte('\n')
		}
		s = s[end+1:]
	}
	return tidyLines(b.String())
}

// tagName reads "<name", "</name" or "<!..." and reports whether s starts a tag.
func tagName(s string) (string, bool, bool) {
	rest := s[1:]
	closing := strings.HasPrefix(rest, "/")
	rest = strings.TrimPrefix(rest, "/")
	if strings.HasPrefix(rest, "!") {
		return "", closing, true
	}
	end := 0
	for end < len(rest) && tagByte(rest[end], end > 0) {
		end++
	}
	return strings.ToLower(rest[:end]), closing, end > 0
}

func tagByte(c byte, digits bool) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || digits && '0' <= c && c <= '9'
}

// tidyLines trims trailing spaces, collapses blank runs and drops blank edges.
func tidyLines(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, " \t ")
		if line == "" && (len(out) == 0 || out[len(out)-1] == "") {
			continue
		}
		out = append(out, line)
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return strings.Join(out, "\n")
}

// Wrap breaks one line of plain (already escaped) text into lines of at most
// width cells, preferring spaces and splitting long words by cell width. It
// reads the text once, so large unbroken bodies stay linear.
func Wrap(s string, width int) []string {
	if width < 1 || DisplayWidth(s) <= width {
		return []string{s}
	}
	var lines []string
	var line strings.Builder
	used, breakAt, breakUsed := 0, -1, 0
	flush := func(keep string, keepUsed int) {
		lines = append(lines, strings.TrimRight(strings.TrimSuffix(line.String(), keep), " "))
		line.Reset()
		line.WriteString(keep)
		used, breakAt = keepUsed, -1
	}
	for _, r := range s {
		cells := runeCells(r)
		for used > 0 && used+cells > width {
			if text := line.String(); breakAt > 0 && breakAt < len(text) && r != ' ' {
				flush(text[breakAt:], used-breakUsed)
				continue
			}
			flush("", 0)
		}
		if r == ' ' && used == 0 && len(lines) > 0 {
			continue
		}
		line.WriteRune(r)
		used += cells
		if r == ' ' {
			breakAt, breakUsed = line.Len(), used
		}
	}
	if strings.TrimSpace(line.String()) != "" {
		flush("", 0)
	}
	return lines
}

// runeCells is DisplayWidth for one rune without allocating.
func runeCells(r rune) int {
	switch {
	case r < 0x20 || (r >= 0x7f && r < 0xa0) || (r >= 0x300 && r <= 0x36f) || r == 0x200b:
		return 0
	case wide(r):
		return 2
	}
	return 1
}
