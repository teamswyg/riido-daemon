package cliui

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Safe makes a remote value printable on one terminal line. Control
// characters (ESC, newlines, C1 codes), bidi overrides and invalid UTF-8 are
// written as escapes, so a name cannot clear the screen, forge table rows or
// reorder text. Apply it before adding trusted styling.
func Safe(s string) string {
	if !needsEscape(s) {
		return s
	}
	var b strings.Builder
	for i, r := range s {
		if r == utf8.RuneError {
			if _, size := utf8.DecodeRuneInString(s[i:]); size <= 1 {
				b.WriteString(`\x` + strconv.FormatUint(uint64(s[i]), 16))
				continue
			}
		}
		if unsafeRune(r) {
			quoted := strconv.QuoteRuneToASCII(r)
			b.WriteString(quoted[1 : len(quoted)-1])
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func needsEscape(s string) bool {
	if !utf8.ValidString(s) {
		return true
	}
	for _, r := range s {
		if unsafeRune(r) {
			return true
		}
	}
	return false
}

func unsafeRune(r rune) bool {
	return unicode.IsControl(r) || r >= 0x202A && r <= 0x202E || r >= 0x2066 && r <= 0x2069
}
