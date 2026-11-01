// Package clicompletion matches static CLI grammar without side effects.
package clicompletion

import (
	"io"
	"slices"
	"strings"
	"unicode/utf8"
)

const MaxInput = 64 << 10

// Command owns completion grammar, not execution permissions. A nil value list
// means an opaque value: consume it but never suggest IDs, secrets or paths.
type Command struct {
	Path        string
	Flags       []string
	Values      map[string][]string
	Positionals func([]string) []string
}

// ReadWords accepts only bounded NUL-framed UTF-8. PowerShell's native pipe adds
// one final newline; no other trailing bytes or malformed frames are accepted.
func ReadWords(r io.Reader) []string {
	if r == nil {
		return nil
	}
	raw, err := io.ReadAll(io.LimitReader(r, MaxInput+1))
	if err != nil || len(raw) > MaxInput || !utf8.Valid(raw) {
		return nil
	}
	s := string(raw)
	if withoutNewline, ok := strings.CutSuffix(s, "\n"); ok {
		s = strings.TrimSuffix(withoutNewline, "\r")
	}
	if !strings.HasSuffix(s, "\x00") {
		return nil
	}
	parts := strings.Split(strings.TrimSuffix(s, "\x00"), "\x00")
	if len(parts) < 2 || len(parts) > 513 {
		return nil
	}
	switch parts[0] {
	case "words":
		return parts[1:]
	case "shell":
		if len(parts) == 2 {
			return ShellWords(parts[1])
		}
	}
	return nil
}

// Candidates never returns any input token, only catalog constants.
func Candidates(catalog []Command, words []string) []string {
	if len(words) == 0 {
		return nil
	}
	previous, current := words[:len(words)-1], words[len(words)-1]
	if len(previous) == 0 {
		roots := []string{"--help", "-h"}
		for _, cmd := range catalog {
			root, _, _ := strings.Cut(cmd.Path, " ")
			roots = append(roots, root)
		}
		return matching(roots, current)
	}
	var selected *Command
	pathLen := 0
	children := []string{}
	for i := range catalog {
		path := strings.Fields(catalog[i].Path)
		if len(previous) >= len(path) && slices.Equal(previous[:len(path)], path) && len(path) > pathLen {
			selected = &catalog[i]
			pathLen = len(path)
		}
		if len(previous) < len(path) && slices.Equal(previous, path[:len(previous)]) {
			children = append(children, path[len(previous)])
		}
	}
	if len(children) > 0 {
		return matching(append(children, "--help", "-h"), current)
	}
	if selected == nil {
		return nil
	}
	used := map[string]bool{}
	positionals := []string{}
	for i := pathLen; i < len(previous); i++ {
		word := previous[i]
		if !strings.HasPrefix(word, "-") {
			positionals = append(positionals, word)
			continue
		}
		flag, _, inline := strings.Cut(word, "=")
		if choices, ok := selected.Values[flag]; ok {
			used[flag] = true
			if inline {
				continue
			}
			if i+1 == len(previous) {
				return matching(choices, current)
			}
			i++
			continue
		}
		if inline || !slices.Contains(selected.Flags, flag) {
			return nil
		}
		used[flag] = true
	}
	if flag, value, inline := strings.Cut(current, "="); inline {
		choices, ok := selected.Values[flag]
		if !ok {
			return nil
		}
		matches := matching(choices, value)
		for i := range matches {
			matches[i] = flag + "=" + matches[i]
		}
		return matches
	}
	choices := []string{}
	if current == "" || strings.HasPrefix(current, "-") {
		for _, flag := range selected.Flags {
			if !used[flag] {
				choices = append(choices, flag)
			}
		}
		for flag := range selected.Values {
			if !used[flag] {
				choices = append(choices, flag)
			}
		}
	}
	if !strings.HasPrefix(current, "-") && selected.Positionals != nil {
		choices = append(choices, selected.Positionals(positionals)...)
	}
	return matching(choices, current)
}

func matching(values []string, prefix string) []string {
	var result []string
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			result = append(result, value)
		}
	}
	slices.Sort(result)
	return slices.Compact(result)
}
