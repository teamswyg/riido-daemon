package openclaw

import "bytes"

func nextParserLine(buf *[]byte) ([]byte, bool) {
	idx := bytes.IndexByte(*buf, '\n')
	if idx < 0 {
		return nil, false
	}

	line := (*buf)[:idx]
	*buf = (*buf)[idx+1:]
	return line, true
}
