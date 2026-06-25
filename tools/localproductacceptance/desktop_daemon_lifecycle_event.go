package main

import (
	"bufio"
	"encoding/json"
	"os"
)

type desktopDaemonStopEvent struct {
	ObservedAt string `json:"observed_at"`
	Reason     string `json:"reason"`
	Method     string `json:"method"`
	Profile    string `json:"profile"`
	SocketPath string `json:"socketPath"`
}

func readDesktopDaemonStopEvents(path string) ([]desktopDaemonStopEvent, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	events := []desktopDaemonStopEvent{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event desktopDaemonStopEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
