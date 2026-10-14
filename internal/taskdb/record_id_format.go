package taskdb

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

func transitionID(taskID string, event ir.EventType, now time.Time, ordinal int) string {
	return fmt.Sprintf("transition:%s:%s:%s:%04d", taskID, event, recordTimestamp(now), ordinal)
}

func evidenceID(taskID string, now time.Time, ordinal int) string {
	return fmt.Sprintf("evidence:%s:%s:%04d", taskID, recordTimestamp(now), ordinal)
}

func commandReceiptID(taskID, kind string, now time.Time, ordinal int) string {
	return fmt.Sprintf("receipt:%s:%s:%s:%04d", kind, taskID, recordTimestamp(now), ordinal)
}

func generatedCommandID(taskID, kind string, now time.Time, ordinal int) string {
	return fmt.Sprintf("command:%s:%s:%s:%04d", kind, taskID, recordTimestamp(now), ordinal)
}

func recordTimestamp(now time.Time) string {
	return now.UTC().Format("20060102T150405.000000000Z")
}
