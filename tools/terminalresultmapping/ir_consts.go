package main

import "github.com/teamswyg/riido-contracts/ir"

func eventTypeValues() map[string]string {
	return map[string]string{
		"EventRunReportedDone": string(ir.EventRunReportedDone),
		"EventTaskCancelled":   string(ir.EventTaskCancelled),
		"EventTaskFailed":      string(ir.EventTaskFailed),
		"EventTaskTimedOut":    string(ir.EventTaskTimedOut),
	}
}
