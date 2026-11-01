package cliconfig

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveProfileAndOriginIsolation(t *testing.T) {
	c := Default()
	c.ActiveProfile = "staging"
	c.Profiles["staging"] = Profile{APIOrigin: "https://staging.example", SelectionOrigin: "https://staging.example", Workspace: "ws-one", Team: "team-one", Format: "json", Timeout: "45s"}
	for _, tc := range []struct {
		name, profile                          string
		explicit                               bool
		env                                    map[string]string
		wantProfile, wantOrigin, wantWorkspace string
	}{
		{"stored", "", false, nil, "staging", "https://staging.example", "ws-one"},
		{"environment profile", "", false, map[string]string{"RIIDO_CLI_PROFILE": "production"}, "production", "https://monolith.riido.io", ""},
		{"flag wins", "staging", true, map[string]string{"RIIDO_CLI_PROFILE": "invalid"}, "staging", "https://staging.example", "ws-one"},
		{"new origin drops selection", "", false, map[string]string{"RIIDO_CLI_API_URL": "https://other.example"}, "staging", "https://other.example", ""},
		{"explicit environment workspace", "", false, map[string]string{"RIIDO_CLI_API_URL": "https://other.example", "RIIDO_CLI_WORKSPACE": "ws-two"}, "staging", "https://other.example", "ws-two"},
		{"unconfigured local", "local", true, nil, "local", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Resolve(c, tc.profile, tc.explicit, func(k string) string { return tc.env[k] })
			if err != nil || r.Profile != tc.wantProfile || r.APIOrigin != tc.wantOrigin || r.Workspace != tc.wantWorkspace {
				t.Fatal(r, err)
			}
			if tc.wantWorkspace == "" && r.Team != "" {
				t.Fatal("team crossed origin")
			}
		})
	}
}

func TestConfigRejectsUnsafeOrUnknownSettings(t *testing.T) {
	for _, raw := range []string{
		`{}`, `{"version":2,"activeProfile":"production","profiles":{}}`,
		`{"version":1,"activeProfile":"production","profiles":{},"token":"synthetic-secret"}`,
		`{"version":1,"activeProfile":"unknown","profiles":{}}`,
		`{"version":1,"activeProfile":"local","profiles":{"local":{"apiOrigin":"https://remote.example"}}}`,
		`{"version":1,"activeProfile":"staging","profiles":{"staging":{"workspace":"ws-one"}}}`,
		`{"version":1,"activeProfile":"staging","profiles":{"staging":{"apiOrigin":"https://user:secret@example.com"}}}`,
		`{"version":1,"activeProfile":"staging","profiles":{"staging":{"timeout":"11m"}}}`,
	} {
		t.Run(raw, func(t *testing.T) {
			s := Store{Dir: t.TempDir()}
			path := filepath.Join(s.Dir, "config.json")
			if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Load(); !errors.Is(err, ErrInvalid) {
				t.Fatal(err)
			}
			if err := s.Update(context.Background(), func(c *Config) error { *c = Default(); return nil }); !errors.Is(err, ErrInvalid) {
				t.Fatal(err)
			}
			after, _ := os.ReadFile(path)
			if string(after) != raw {
				t.Fatal("malformed file overwritten")
			}
		})
	}
}

func TestStoreSelectionCASAndNoSecrets(t *testing.T) {
	s := Store{Dir: filepath.Join(t.TempDir(), "new")}
	c, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	r, err := Resolve(c, "staging", true, func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set(context.Background(), r, "workspace", "ws-one"); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(context.Background(), r, "workspace", "ws-stale"); !errors.Is(err, ErrConflict) {
		t.Fatal("stale writer accepted", err)
	}
	c, _ = s.Load()
	r, _ = Resolve(c, "staging", true, func(string) string { return "" })
	if err := s.Set(context.Background(), r, "team", "team-one"); err != nil {
		t.Fatal(err)
	}
	c, _ = s.Load()
	r, _ = Resolve(c, "staging", true, func(string) string { return "" })
	if err := s.Set(context.Background(), r, "workspace", "ws-two"); err != nil {
		t.Fatal(err)
	}
	c, _ = s.Load()
	if c.Profiles["staging"].Team != "" {
		t.Fatal("old team retained")
	}
	r, _ = Resolve(c, "staging", true, func(string) string { return "" })
	if err := s.Set(context.Background(), r, "api_origin", "https://other.example"); err != nil {
		t.Fatal(err)
	}
	c, _ = s.Load()
	if p := c.Profiles["staging"]; p.Workspace != "" || p.Team != "" || p.SelectionOrigin != "" {
		t.Fatal("selection retained across origin")
	}
	r, _ = Resolve(c, "staging", true, func(string) string { return "" })
	for _, key := range []string{"access_token", "refresh_token", "cookie", "secret", "client_id"} {
		if err := s.Set(context.Background(), r, key, "synthetic-secret"); !errors.Is(err, ErrInvalid) {
			t.Fatal(key, err)
		}
	}
	info, _ := os.Stat(filepath.Join(s.Dir, "config.json"))
	if info.Mode().Perm() != 0o600 {
		t.Fatal("unsafe file permissions", info.Mode())
	}
}

func TestStorePreservesOtherProfileAndCanceledWrite(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	c := Default()
	r, _ := Resolve(c, "staging", true, func(string) string { return "" })
	other, _ := Resolve(c, "production", true, func(string) string { return "" })
	if err := s.Set(context.Background(), other, "format", "json"); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(context.Background(), r, "timeout", "45s"); err != nil {
		t.Fatal(err)
	}
	c, _ = s.Load()
	if c.Profiles["production"].Format != "json" || c.Profiles["staging"].Timeout != "45s" {
		t.Fatal(c)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Update(ctx, func(*Config) error { t.Fatal("called after cancellation"); return nil }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestConfigSymlinkAndSizeRejected(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	target := filepath.Join(t.TempDir(), "elsewhere")
	if err := os.WriteFile(target, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(s.Dir, "config.json")); err != nil {
		t.Skip("symlink unavailable")
	}
	if _, err := s.Load(); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	s = Store{Dir: t.TempDir()}
	if err := os.WriteFile(filepath.Join(s.Dir, "config.json"), make([]byte, maxConfigBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}
