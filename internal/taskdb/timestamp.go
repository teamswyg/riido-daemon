package taskdb

import "time"

func timestamp(now time.Time) string {
	return now.UTC().Format(time.RFC3339Nano)
}
