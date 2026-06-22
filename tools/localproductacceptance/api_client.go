package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type apiClient struct {
	base  string
	token string
	http  http.Client
}

func newAPIClient(base, token string) apiClient {
	return apiClient{
		base:  strings.TrimRight(base, "/"),
		token: strings.TrimSpace(token),
		http:  http.Client{Timeout: 20 * time.Second},
	}
}

func (c apiClient) call(method, path string, body any) (map[string]any, int, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("encode request: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, c.base+path, reader)
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Riido-Ai-Agent-Token", c.token)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	var decoded map[string]any
	_ = json.Unmarshal(data, &decoded)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decoded, resp.StatusCode, fmt.Errorf("http status %d", resp.StatusCode)
	}
	return decoded, resp.StatusCode, nil
}
