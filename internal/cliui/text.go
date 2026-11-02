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
		end := strings.IndexByte(s, '>')
		name, closing, ok := tagName(s)
		if end < 0 || !ok {
			b.WriteByte('<')
			s = s[1:]
			continue
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
// width cells, preferring spaces and splitting long words by cell width.
func Wrap(s string, width int) []string {
	if width < 1 || DisplayWidth(s) <= width {
		return []string{s}
	}
	var lines []string
	line, used := "", 0
	flush := func() {
		lines = append(lines, strings.TrimRight(line, " "))
		line, used = "", 0
	}
	for _, word := range strings.SplitAfter(s, " ") {
		if used == 0 && len(lines) > 0 && strings.TrimSpace(word) == "" {
			continue
		}
		if used > 0 && used+DisplayWidth(strings.TrimRight(word, " ")) > width {
			flush()
		}
		for used == 0 && DisplayWidth(strings.TrimRight(word, " ")) > width {
			head, rest := splitCells(word, width)
			lines = append(lines, head)
			word = rest
		}
		line += word
		used += DisplayWidth(word)
	}
	if strings.TrimSpace(line) != "" {
		flush()
	}
	return lines
}

// splitCells cuts s after at most width cells, always keeping one rune.
func splitCells(s string, width int) (string, string) {
	used := 0
	for i, r := range s {
		w := DisplayWidth(string(r))
		if i > 0 && used+w > width {
			return s[:i], s[i:]
		}
		used += w
	}
	return s, ""
}
