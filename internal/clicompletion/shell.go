package clicompletion

import (
	"path"
	"strings"
	"unicode"
)

// ShellWords tokenizes a literal Bash/fish cursor prefix without evaluating
// variable, wildcard or command expansions. This preserves quotes, empty values, '=' and byte cursor offsets.
// Shell expressions requiring evaluation are deliberately left uncompleted.
func ShellWords(prefix string) []string {
	var words []string
	var word strings.Builder
	var quote rune
	escaped, started := false, false
	for _, r := range prefix {
		if escaped {
			word.WriteRune(r)
			escaped = false
			started = true
			continue
		}
		if r == '\\' && quote != '\'' {
			escaped = true
			started = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				word.WriteRune(r)
			}
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			started = true
			continue
		}
		if strings.ContainsRune("`$();|&<>", r) {
			return nil
		}
		if unicode.IsSpace(r) {
			if started {
				words = append(words, word.String())
				word.Reset()
				started = false
			}
			continue
		}
		word.WriteRune(r)
		started = true
	}
	if escaped {
		return nil
	}
	words = append(words, word.String())
	if len(words) < 2 || len(words) > 513 {
		return nil
	}
	if command := path.Base(words[0]); command != "riido" && command != "riido.exe" {
		return nil
	}
	return words[1:]
}
