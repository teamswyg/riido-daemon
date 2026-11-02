package cliui

import (
	"fmt"
	"strings"
)

// Brand colors from the Riido signature gradient. The core brand purple is
// too dark to read on dark terminals (contrast 2.6:1), so text uses a lighter
// tint there and the core color only on light backgrounds.
const (
	brandRGB      = 0x5d4ff9 // 5.3:1 on white
	brandLightRGB = 0x9a8cff // 5.0:1 on a dark terminal
	brandMidRGB   = 0x8577ff // about 4:1 on both when the background is unknown
	brandDark     = 0x3e00bf
)

func (t Terminal) accentRGB() uint32 {
	switch t.Background {
	case "light":
		return brandRGB
	case "dark":
		return brandLightRGB
	}
	return brandMidRGB
}

func (t Terminal) wrap(codes, s string) string {
	if !t.Color || s == "" {
		return s
	}
	return "\x1b[" + codes + "m" + s + "\x1b[0m"
}

func (t Terminal) rgb(rgb uint32) string {
	r, g, b := rgb>>16&0xff, rgb>>8&0xff, rgb&0xff
	if t.TrueColor {
		return fmt.Sprintf("2;%d;%d;%d", r, g, b)
	}
	return fmt.Sprintf("5;%d", xterm256(r, g, b))
}

// xterm256 maps a color onto the 6x6x6 cube or the gray ramp.
func xterm256(r, g, b uint32) int {
	if r == g && g == b {
		if r < 8 {
			return 16
		}
		if r > 248 {
			return 231
		}
		return 232 + int((r-8)*24/247)
	}
	level := func(v uint32) int { return int((v*5 + 127) / 255) }
	return 16 + 36*level(r) + 6*level(g) + level(b)
}

func (t Terminal) Bold(s string) string   { return t.wrap("1", s) }
func (t Terminal) Dim(s string) string    { return t.wrap("2", s) }
func (t Terminal) Accent(s string) string { return t.wrap("38;"+t.rgb(t.accentRGB()), s) }

// Brand is bold accent text for the product name.
func (t Terminal) Brand(s string) string   { return t.wrap("1;38;"+t.rgb(t.accentRGB()), s) }
func (t Terminal) Success(s string) string { return t.wrap("32", s) }
func (t Terminal) Warn(s string) string    { return t.wrap("33", s) }
func (t Terminal) Error(s string) string   { return t.wrap("31", s) }

// Symbol returns a status marker; plain terminals keep ASCII brackets.
func (t Terminal) Symbol(status string) string {
	if !t.Interactive {
		return "[" + status + "]"
	}
	switch status {
	case "ok":
		return t.Success("✓")
	case "warn":
		return t.Warn("!")
	case "fail":
		return t.Error("✗")
	default:
		return t.Dim("·")
	}
}

// StripANSI removes SGR sequences so widths can be measured.
func StripANSI(s string) string {
	if !strings.Contains(s, "\x1b[") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] < 0x40 || s[j] > 0x7e) {
				j++
			}
			i = j
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
