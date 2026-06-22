package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"
)

const streamProbeTimeout = 5 * time.Second

func apiStreamReplay(client apiClient, path string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), streamProbeTimeout)
	defer cancel()
	data, status, err := client.callStream(ctx, http.MethodGet, path)
	if err != nil && len(data) == 0 {
		return "", status, err
	}
	return string(data), status, nil
}

func (c apiClient) callStream(ctx context.Context, method, path string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-Riido-Ai-Agent-Token", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, readErr := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if readErr != nil && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return data, resp.StatusCode, readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return data, resp.StatusCode, errors.New(resp.Status)
	}
	return data, resp.StatusCode, nil
}
