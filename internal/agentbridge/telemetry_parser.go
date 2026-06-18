package agentbridge

import "strings"

func (p *TelemetryParser) Feed(text string) []Event {
	if p == nil || text == "" {
		return nil
	}
	p.buf += text
	if len(p.buf) > 64*1024 {
		p.buf = p.buf[len(p.buf)-64*1024:]
	}
	out := []Event{}
	for {
		start := strings.Index(p.buf, telemetryLogStart)
		if start < 0 {
			p.buf = suffixThatCanStartTag(p.buf)
			return out
		}
		if start > 0 {
			p.buf = p.buf[start:]
		}
		afterStart := p.buf[len(telemetryLogStart):]
		before, after, ok := strings.Cut(afterStart, telemetryLogEnd)
		if !ok {
			return out
		}
		message := strings.TrimSpace(before)
		if event, ok := progressEventFromTelemetryMessage(message); ok {
			out = append(out, event)
		}
		p.buf = after
	}
}

func suffixThatCanStartTag(s string) string {
	limit := min(len(s), len(telemetryLogStart)-1)
	for n := limit; n > 0; n-- {
		if strings.HasSuffix(s, telemetryLogStart[:n]) {
			return s[len(s)-n:]
		}
	}
	return ""
}
