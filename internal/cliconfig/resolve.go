package cliconfig

import "strings"

type Resolved struct {
	Profile       string
	ActiveProfile string
	Stored        Profile
	APIOrigin     string
	Workspace     string
	Team          string
	Format        string
	Timeout       string
	Sources       map[string]string
}

// Resolve is pure. Its result is a single invocation snapshot, not authority.
func Resolve(c Config, profile string, explicit bool, getenv func(string) string) (Resolved, error) {
	if !c.Valid() {
		return Resolved{}, ErrInvalid
	}
	r := Resolved{ActiveProfile: c.ActiveProfile, Sources: map[string]string{}}
	r.Profile, r.Sources["profile"] = c.ActiveProfile, "config"
	if value := getenv("RIIDO_CLI_PROFILE"); value != "" {
		r.Profile, r.Sources["profile"] = value, "environment"
	}
	if explicit {
		r.Profile, r.Sources["profile"] = profile, "flag"
	}
	if !ValidProfile(r.Profile) {
		return Resolved{}, ErrInvalid
	}
	r.Stored = c.Profiles[r.Profile]
	r.APIOrigin, r.Sources["api_origin"] = r.Stored.APIOrigin, "profile"
	if r.APIOrigin == "" {
		r.Sources["api_origin"] = "default"
		switch r.Profile {
		case "production":
			r.APIOrigin = "https://monolith.riido.io"
		case "staging":
			r.APIOrigin = "https://monolith.api.riido.io"
		}
	}
	if value := strings.TrimSpace(getenv("RIIDO_CLI_API_URL")); value != "" {
		r.APIOrigin, r.Sources["api_origin"] = value, "environment"
	}
	if r.APIOrigin != "" {
		origin, err := Origin(r.APIOrigin, r.Profile == "local")
		if err != nil {
			return Resolved{}, err
		}
		r.APIOrigin = origin
	}
	if r.APIOrigin != "" && r.Stored.SelectionOrigin == r.APIOrigin {
		r.Workspace, r.Team = r.Stored.Workspace, r.Stored.Team
	}
	r.Format, r.Timeout = r.Stored.Format, r.Stored.Timeout
	for key, item := range map[string]struct {
		value *string
		env   string
	}{
		"workspace": {&r.Workspace, "RIIDO_CLI_WORKSPACE"}, "team": {&r.Team, "RIIDO_CLI_TEAM"}, "format": {&r.Format, "RIIDO_CLI_FORMAT"}, "timeout": {&r.Timeout, "RIIDO_CLI_TIMEOUT"},
	} {
		r.Sources[key] = "default"
		if *item.value != "" {
			r.Sources[key] = "profile"
		}
		if value := getenv(item.env); value != "" {
			*item.value = value
			r.Sources[key] = "environment"
		}
	}
	if r.Workspace != "" && !ValidID(r.Workspace) || r.Team != "" && !ValidID(r.Team) || r.Format != "" && r.Format != "json" && r.Format != "text" || r.Timeout != "" && !ValidTimeout(r.Timeout) {
		return Resolved{}, ErrInvalid
	}
	return r, nil
}
