package mwsdbridge

import "time"

func (c Client) timeout() time.Duration {
	if c.Timeout == 0 {
		return defaultClientTimeout
	}
	return c.Timeout
}
