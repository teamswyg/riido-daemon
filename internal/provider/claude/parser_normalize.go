package claude

import "bytes"

func normalizeStreamLine(line []byte, prefixes []string) []byte {
	line = trimTrailingCR(line)
	for _, prefix := range prefixes {
		if bytes.HasPrefix(line, []byte(prefix)) {
			line = line[len(prefix):]
			break
		}
	}
	return bytes.TrimSpace(line)
}

func trimTrailingCR(line []byte) []byte {
	if n := len(line); n > 0 && line[n-1] == '\r' {
		return line[:n-1]
	}
	return line
}
