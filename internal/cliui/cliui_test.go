package cliui

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDetectKeepsPipesAndFilesPlain(t *testing.T) {
	var buf bytes.Buffer
	if got := Detect(&buf, func(string) string { return "" }); got != Plain {
		t.Fatalf("buffer = %+v", got)
	}
	file, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if got := Detect(file, func(string) string { return "" }); got.Interactive || got.Color {
		t.Fatalf("regular file = %+v", got)
	}
}

func TestPlainTerminalWritesNoEscapes(t *testing.T) {
	for _, s := range []string{Plain.Bold("a"), Plain.Accent("b"), Plain.Error("c"), Plain.Logo(), Plain.Mark(), Plain.Symbol("ok")} {
		if strings.Contains(s, "\x1b") || strings.ContainsAny(s, "✓▀▄") {
			t.Fatalf("plain output styled: %q", s)
		}
	}
	if Plain.Symbol("warn") != "[warn]" || Plain.Logo() != "" {
		t.Fatal("plain fallbacks changed")
	}
}

func TestLogoNeedsColorAndWidth(t *testing.T) {
	wide := Terminal{Interactive: true, Color: true, TrueColor: true, Width: 120}
	logo := wide.Logo()
	if strings.Count(logo, "\n") != len(logoRows) || !strings.Contains(logo, "38;2;") {
		t.Fatalf("truecolor logo missing: %d lines", strings.Count(logo, "\n"))
	}
	for i, line := range strings.Split(strings.TrimSuffix(logo, "\n"), "\n") {
		if DisplayWidth(line) > LogoWidth() {
			t.Fatalf("line %d is %d cells", i, DisplayWidth(line))
		}
	}
	palette := Terminal{Interactive: true, Color: true, Width: 120}.Logo()
	if strings.Contains(palette, "38;2;") || !strings.Contains(palette, "38;5;") {
		t.Fatal("256-color fallback not used")
	}
	for _, term := range []Terminal{{Interactive: true, Color: true, Width: SymbolWidth() + 1}, {Interactive: true, Width: 200}} {
		if term.Logo() != "" {
			t.Fatalf("logo shown on %+v", term)
		}
	}
}

// A terminal narrower than the signature shows only the round symbol.
func TestNarrowTerminalShowsOnlySymbol(t *testing.T) {
	for _, width := range []int{SymbolWidth() + 2, LogoWidth() + 1} {
		logo := Terminal{Interactive: true, Color: true, TrueColor: true, Width: width}.Logo()
		if strings.Count(logo, "\n") != len(symbolRows) {
			t.Fatalf("width %d: %d lines, want the %d-row symbol", width, strings.Count(logo, "\n"), len(symbolRows))
		}
		for i, line := range strings.Split(strings.TrimSuffix(logo, "\n"), "\n") {
			if DisplayWidth(line) > SymbolWidth() {
				t.Fatalf("width %d line %d is %d cells", width, i, DisplayWidth(line))
			}
		}
	}
	if SymbolWidth() >= LogoWidth() {
		t.Fatalf("symbol %d cells is not narrower than the signature %d", SymbolWidth(), LogoWidth())
	}
}

func TestDisplayWidthCountsWideCharacters(t *testing.T) {
	for s, want := range map[string]int{"abc": 3, "작업": 4, "\x1b[1m굵게\x1b[0m": 4, "✓ ok": 4, "": 0} {
		if got := DisplayWidth(s); got != want {
			t.Fatalf("%q = %d, want %d", s, got, want)
		}
	}
	if got := Truncate("가나다라마", 7); got != "가나다…" {
		t.Fatalf("truncate = %q", got)
	}
}

func TestTableAlignsMixedWidthCells(t *testing.T) {
	var buf bytes.Buffer
	if err := Plain.Table(&buf, []string{"ID", "이름", "상태"}, [][]string{{"a1", "리도", "online"}, {"b22", "Codex", "offline"}}); err != nil {
		t.Fatal(err)
	}
	want := "ID   이름   상태\na1   리도   online\nb22  Codex  offline\n"
	if buf.String() != want {
		t.Fatalf("table:\n%s\nwant:\n%s", buf.String(), want)
	}
	buf.Reset()
	narrow := Terminal{Interactive: true, Width: 20}
	_ = narrow.Table(&buf, []string{"ID", "제목"}, [][]string{{"RIID-1", "아주 긴 작업 제목이 들어갑니다"}})
	for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		if DisplayWidth(line) > 20 {
			t.Fatalf("line exceeds width: %q", line)
		}
	}
}
