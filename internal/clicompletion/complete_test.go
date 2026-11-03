package clicompletion

import (
	"reflect"
	"strings"
	"testing"
)

func TestCompletionFraming(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want []string
	}{
		{"words\x00work\x00\x00", []string{"work", ""}},
		{"words\x00work\x00한 글\x00--pro\x00\n", []string{"work", "한 글", "--pro"}},
		{"words\x00--profile=st\x00\r\n", []string{"--profile=st"}},
		{"words\x00work\x00\r", nil},
		{"words\x00work\x00\n\n", nil},
		{"words\x00unterminated", nil},
		{"words\x00\xff\x00", nil},
		{"unknown\x00work\x00", nil},
		{strings.Repeat("x", MaxInput+1), nil},
		{"words\x00" + strings.Repeat("x\x00", 513), nil},
		{"shell\x00riido config set profile st\x00", []string{"config", "set", "profile", "st"}},
	} {
		if got := ReadWords(strings.NewReader(tc.raw)); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("framing mismatch: got %q want %q", got, tc.want)
		}
	}
}

func TestShellCursorPrefixAndQuotes(t *testing.T) {
	for _, tc := range []struct {
		line string
		want []string
	}{
		{"riido ", []string{""}},
		{"riido config list --profile=st", []string{"config", "list", "--profile=st"}},
		{"riido work search --query '한글 여러 단어' --f", []string{"work", "search", "--query", "한글 여러 단어", "--f"}},
		{"riido config set workspace '' --f", []string{"config", "set", "workspace", "", "--f"}},
		{"riido config list --profile \"st", []string{"config", "list", "--profile", "st"}},
		{"riido work search --query two\\ words --f", []string{"work", "search", "--query", "two words", "--f"}},
		{"riido work search --query $(touch sentinel) --f", nil},
		{"riido work search --query `touch sentinel` --f", nil},
		{"riido work get > file --f", nil},
	} {
		if got := ShellWords(tc.line); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%q: got %q want %q", tc.line, got, tc.want)
		}
	}
}

func TestStaticCompletionGrammar(t *testing.T) {
	catalog := []Command{{Path: "work get", Flags: []string{"--json"}, Values: map[string][]string{"--profile": {"staging", "production"}, "--work": nil}}, {Path: "work search"}}
	for _, tc := range []struct{ words, want []string }{
		{[]string{"wo"}, []string{"work"}},
		{[]string{"work", "g"}, []string{"get"}},
		{[]string{"work", "get", "--profile", "st"}, []string{"staging"}},
		{[]string{"work", "get", "--profile=st"}, []string{"--profile=staging"}},
		{[]string{"work", "get", "--work", ""}, nil},
		{[]string{"work", "get", "--token", "private"}, nil},
		{[]string{"work", "get", "--json", "--j"}, nil},
		{[]string{"work", "get", "--work", "한글 여러 단어", "--p"}, []string{"--profile"}},
		{[]string{"work", "not-a-command", ""}, nil},
	} {
		if got := Candidates(catalog, tc.words); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%q: got %q want %q", tc.words, got, tc.want)
		}
	}
}
