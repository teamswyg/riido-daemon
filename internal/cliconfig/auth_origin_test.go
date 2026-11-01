package cliconfig

import (
	"context"
	"errors"
	"testing"
)

func TestAuthOriginDefaultsAndProfileIsolation(t *testing.T) {
	c := Default()
	c.Profiles["staging"] = Profile{AuthOrigin: "desktop"}
	for _, profile := range []string{"production", "staging"} {
		r, err := Resolve(c, profile, true, func(string) string { return "" })
		want := "browser"
		if profile == "staging" {
			want = "desktop"
		}
		if err != nil || r.AuthOrigin != want {
			t.Fatal(r, err)
		}
	}
	for _, value := range []string{"Desktop", "unknown", "browser desktop"} {
		c.Profiles["staging"] = Profile{AuthOrigin: value}
		if c.Valid() {
			t.Fatal("invalid selector accepted")
		}
	}
}

func TestAuthOriginChangeClearsScopeButPreservesOtherProfile(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	other := Profile{AuthOrigin: "desktop", APIOrigin: "https://other.example"}
	err := s.Update(context.Background(), func(c *Config) error {
		c.Profiles["production"] = Profile{SelectionOrigin: "https://monolith.riido.io", Workspace: "ws-one", Team: "team-one"}
		c.Profiles["staging"] = other
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	c, _ := s.Load()
	r, _ := Resolve(c, "production", true, func(string) string { return "" })
	if err := s.Set(context.Background(), r, "auth_origin", "desktop"); err != nil {
		t.Fatal(err)
	}
	c, _ = s.Load()
	p := c.Profiles["production"]
	if p.AuthOrigin != "desktop" || p.Workspace != "" || p.Team != "" || p.SelectionOrigin != "" || c.Profiles["staging"] != other {
		t.Fatal("selection or another profile changed incorrectly")
	}
	r, _ = Resolve(c, "production", true, func(string) string { return "" })
	if err := s.Set(context.Background(), r, "auth_origin", ""); err != nil {
		t.Fatal(err)
	}
	c, _ = s.Load()
	r, _ = Resolve(c, "production", true, func(string) string { return "" })
	if r.AuthOrigin != "browser" || r.Sources["auth_origin"] != "default" {
		t.Fatal("clear did not restore default")
	}
}

func TestAuthOriginActivationDoesNotOverwriteConcurrentDefault(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	c, _ := s.Load()
	r, _ := Resolve(c, "production", true, func(string) string { return "" })
	if err := s.Set(context.Background(), r, "auth_origin", "desktop"); err != nil {
		t.Fatal(err)
	}
	if err := s.ActivateAuthOrigin(context.Background(), r, "browser"); !errors.Is(err, ErrConflict) {
		t.Fatal("browser default overwrote later selection", err)
	}
	c, _ = s.Load()
	if c.Profiles["production"].AuthOrigin != "desktop" {
		t.Fatal("later preference lost")
	}
}

func TestAuthOriginScopeWriteRejectsOneShotOtherOrigin(t *testing.T) {
	for _, key := range []string{"workspace", "team"} {
		t.Run(key, func(t *testing.T) {
			s := Store{Dir: t.TempDir()}
			c, _ := s.Load()
			r, _ := Resolve(c, "production", true, func(string) string { return "" })
			r.AuthOrigin = "desktop"
			if err := s.Set(context.Background(), r, key, "other-origin-id"); !errors.Is(err, ErrConflict) {
				t.Fatal("scope saved for another origin", err)
			}
			c, _ = s.Load()
			if c.Profiles["production"].Workspace != "" || c.Profiles["production"].Team != "" {
				t.Fatal("scope changed")
			}
		})
	}
}

func TestAuthOriginEquivalentBrowserPreferencePreservesScope(t *testing.T) {
	for _, tc := range []struct{ old, next string }{{"", "browser"}, {"browser", ""}} {
		t.Run(tc.old+" to "+tc.next, func(t *testing.T) {
			s := Store{Dir: t.TempDir()}
			expected := Profile{AuthOrigin: tc.old, SelectionOrigin: "https://monolith.riido.io", Workspace: "ws-one", Team: "team-one"}
			if err := s.Update(context.Background(), func(c *Config) error { c.Profiles["production"] = expected; return nil }); err != nil {
				t.Fatal(err)
			}
			c, _ := s.Load()
			r, _ := Resolve(c, "production", true, func(string) string { return "" })
			if err := s.Set(context.Background(), r, "auth_origin", tc.next); err != nil {
				t.Fatal(err)
			}
			c, _ = s.Load()
			expected.AuthOrigin = tc.next
			if c.Profiles["production"] != expected {
				t.Fatal("equivalent browser selection erased workspace/team")
			}
		})
	}
}
