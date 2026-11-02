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
	for _, s := range []string{Plain.Bold("a"), Plain.Accent("b"), Plain.Error("c"), Plain.Mark(), Plain.Symbol("ok")} {
		if strings.Contains(s, "\x1b") || strings.ContainsAny(s, "✓▀▄") {
			t.Fatalf("plain output styled: %q", s)
		}
	}
	if Plain.Symbol("warn") != "[warn]" || Plain.SymbolLines() != nil {
		t.Fatal("plain fallbacks changed")
	}
}

func TestSymbolNeedsColorAndWidth(t *testing.T) {
	lines := Terminal{Interactive: true, Color: true, TrueColor: true, Width: 120}.SymbolLines()
	if len(lines) != len(symbolRows) || !strings.Contains(strings.Join(lines, ""), "38;2;") {
		t.Fatalf("truecolor symbol missing: %d lines", len(lines))
	}
	for i, line := range lines {
		if DisplayWidth(line) != SymbolWidth() {
			t.Fatalf("line %d is %d cells, want %d", i, DisplayWidth(line), SymbolWidth())
		}
	}
	palette := strings.Join(Terminal{Interactive: true, Color: true, Width: 120}.SymbolLines(), "")
	if strings.Contains(palette, "38;2;") || !strings.Contains(palette, "38;5;") {
		t.Fatal("256-color fallback not used")
	}
	if got := (Terminal{Interactive: true, Color: true, Width: SymbolWidth() + 2}).SymbolLines(); len(got) != len(symbolRows) {
		t.Fatal("symbol missing at its minimum width")
	}
	for _, term := range []Terminal{{Interactive: true, Color: true, Width: SymbolWidth() + 1}, {Interactive: true, Width: 200}} {
		if term.SymbolLines() != nil {
			t.Fatalf("symbol shown on %+v", term)
		}
	}
}

func TestSideBySideCentersShorterBlock(t *testing.T) {
	left := []string{"aa", "bb", "cc", "dd", "ee"}
	got := SideBySide(left, 2, 1, []string{"x", "", "y"})
	want := []string{"aa", "bb x", "cc", "dd y", "ee"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("got %q", got)
	}
	got = SideBySide([]string{"aa"}, 2, 2, []string{"x", "y"})
	if strings.Join(got, "|") != "aa  x|    y" {
		t.Fatalf("taller right block: %q", got)
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

func TestAccentFollowsTerminalBackground(t *testing.T) {
	for value, want := range map[string]string{"15;0": "dark", "0;15": "light", "0;default;7": "light", "": "", "x": ""} {
		if got := backgroundFromColorFGBG(value); got != want {
			t.Fatalf("COLORFGBG %q = %q, want %q", value, got, want)
		}
	}
	for background, rgb := range map[string]string{"dark": "154;140;255", "light": "93;79;249", "": "133;119;255"} {
		term := Terminal{Interactive: true, Color: true, TrueColor: true, Background: background}
		if !strings.Contains(term.Accent("a"), rgb) || !strings.Contains(term.Brand("a"), "1;38;2;"+rgb) {
			t.Fatalf("%q background accent = %q", background, term.Accent("a"))
		}
	}
}

func TestLogoAddsPixelWordmarkOnWideTerminals(t *testing.T) {
	for _, row := range wordmarkPattern {
		if len(row) != len(wordmarkPattern[0]) || strings.Trim(row, "#.") != "" {
			t.Fatalf("wordmark row %q", row)
		}
	}
	wide := Terminal{Interactive: true, Color: true, TrueColor: true, Width: LogoWidth() + 2}
	lines := wide.LogoLines()
	if len(lines) != len(symbolRows) || !strings.Contains(strings.Join(lines, ""), "█") {
		t.Fatalf("logo lines = %d", len(lines))
	}
	for i, line := range lines {
		if DisplayWidth(line) > LogoWidth() {
			t.Fatalf("line %d is %d cells", i, DisplayWidth(line))
		}
	}
	for _, term := range []Terminal{{Interactive: true, Color: true, Width: LogoWidth() + 1}, {Interactive: true, Width: 200}, Plain} {
		if term.LogoLines() != nil {
			t.Fatalf("logo shown on %+v", term)
		}
	}
}
