package main

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestLoadDaemonSettingsEnablesPprofForDevelopmentProfile(t *testing.T) {
	env := map[string]string{envDaemonProfile: "development"}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.PprofAddr != defaultDevelopmentPprofAddr {
		t.Fatalf("pprof addr = %q, want %q", settings.PprofAddr, defaultDevelopmentPprofAddr)
	}
}

func TestLoadDaemonSettingsKeepsPprofDisabledByDefault(t *testing.T) {
	settings, err := loadDaemonSettingsFromEnv(func(string) string { return "" }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.PprofAddr != "" {
		t.Fatalf("pprof addr = %q, want disabled", settings.PprofAddr)
	}
}

func TestLoadDaemonSettingsAllowsPprofDevelopmentOverrideOff(t *testing.T) {
	env := map[string]string{
		envDaemonProfile:   "development",
		envDaemonPprofAddr: "off",
	}
	settings, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.PprofAddr != "" {
		t.Fatalf("pprof addr = %q, want disabled", settings.PprofAddr)
	}
}

func TestLoadDaemonSettingsRejectsNonLoopbackPprofAddr(t *testing.T) {
	externalAddr := strings.Join([]string{"0", "0", "0", "0"}, ".") + ":6061"
	env := map[string]string{envDaemonPprofAddr: externalAddr}
	_, err := loadDaemonSettingsFromEnv(func(k string) string { return env[k] }, func() (string, error) {
		return "host", nil
	})
	if err == nil || !strings.Contains(err.Error(), envDaemonPprofAddr) {
		t.Fatalf("expected %s validation error, got %v", envDaemonPprofAddr, err)
	}
}

func TestStartDaemonPprofServerServesIndex(t *testing.T) {
	ctx, cancel := lifecycle.WithCancel(lifecycle.Background())
	defer cancel()
	stop, addr, err := startDaemonPprofServer(ctx, "127.0.0.1:0", logging.NewWriterLogger(io.Discard))
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	res, err := http.Get("http://" + addr + "/debug/pprof/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "profile") || !strings.Contains(string(body), "goroutine") {
		t.Fatalf("pprof index body does not look like pprof: %q", string(body))
	}
}
