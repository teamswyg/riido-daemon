package main

import (
	"io"
	"net/http"
	"strings"
	"time"
)

type routeProbe struct {
	StatusCode int
	FinalURL   string
	Body       string
	Err        error
}

func probeRoute(baseURL, path string) routeProbe {
	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(strings.TrimRight(baseURL, "/") + path)
	if err != nil {
		return routeProbe{Err: err}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	return routeProbe{
		StatusCode: resp.StatusCode,
		FinalURL:   resp.Request.URL.String(),
		Body:       string(body),
	}
}

func routeScenario(id, baseURL, path string) scenario {
	probe := probeRoute(baseURL, path)
	if probe.Err != nil {
		return failedRoute(id, "frontend_not_running", probe.Err.Error())
	}
	if probe.StatusCode >= 500 {
		return failedRoute(id, "frontend_server_error", probe.FinalURL)
	}
	if isMissingRoute(probe) {
		return failedRoute(id, "frontend_route_missing", probe.FinalURL)
	}
	if isAuthRedirect(probe, path) {
		return skippedRoute(id, "frontend_auth_required", probe.FinalURL)
	}
	return scenario{ID: id, Status: statusPassed}
}

func isMissingRoute(probe routeProbe) bool {
	return probe.StatusCode == http.StatusNotFound ||
		strings.Contains(probe.Body, "찾을 수 없는 페이지") ||
		strings.Contains(probe.Body, ">404<")
}

func isAuthRedirect(probe routeProbe, path string) bool {
	return path != "/login" && strings.Contains(probe.FinalURL, "login")
}
