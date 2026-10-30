package saasplane

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func (p *Plane) readAssignmentNotifications(ctx context.Context) (received, unsupported bool, err error) {
	// Keep the configured transport/auth, but a stream must outlive the JSON
	// request timeout. A separate header and idle watchdog bounds broken peers.
	client := *p.client
	client.Timeout = 0
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	req, err := http.NewRequestWithContext(streamCtx, http.MethodGet, p.cfg.BaseURL+"/v1/daemon/assignment-events", nil)
	if err != nil {
		return false, false, err
	}
	req.Header.Set("Accept", "text/event-stream")
	p.attachAuthHeaders(req)
	watchdog := time.AfterFunc(45*time.Second, cancel)
	defer watchdog.Stop()
	response, err := client.Do(req)
	if err != nil {
		return false, false, err
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusNotImplemented:
		return false, true, nil
	case http.StatusOK:
	default:
		return false, false, fmt.Errorf("assignment stream status %d", response.StatusCode)
	}
	if !strings.HasPrefix(response.Header.Get("Content-Type"), "text/event-stream") {
		return false, false, fmt.Errorf("invalid assignment stream content type")
	}
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	eventType, data := "", ""
	for scanner.Scan() {
		watchdog.Reset(45 * time.Second)
		line := scanner.Text()
		if line == "" {
			if eventType == "assignments" {
				var event assignmentNotification
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					return received, false, err
				}
				if event.SchemaVersion != assignmentEventsSchema {
					return received, false, fmt.Errorf("unsupported assignment stream schema")
				}
				p.acceptAssignmentNotification(ctx, event)
				if !received {
					logAssignmentStreamConnected()
					received = true
				}
			}
			eventType, data = "", ""
		} else if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data += strings.TrimSpace(strings.TrimPrefix(line, "data:")) + "\n"
			if len(data) > 1024*1024 {
				return received, false, fmt.Errorf("assignment stream event too large")
			}
		}
	}
	return received, false, scanner.Err()
}

func logAssignmentStreamConnected() { log.Printf("event=daemon_assignment_stream_connected") }
