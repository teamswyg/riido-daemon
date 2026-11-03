package cliconfig

import "context"

// Set changes only the profile captured for this command. Selection requires an
// already verified workspace; this package never grants API authority.
func (s Store) Set(ctx context.Context, r Resolved, key, value string) error {
	return s.Update(ctx, func(c *Config) error {
		if c.Profiles[r.Profile] != r.Stored {
			return ErrConflict
		}
		p := r.Stored
		if (key == "workspace" || key == "team") && value != "" && r.AuthOrigin != p.AuthenticationOrigin() {
			return ErrConflict
		}
		switch key {
		case "profile":
			if !ValidProfile(value) {
				return ErrInvalid
			}
			if c.ActiveProfile != r.ActiveProfile {
				return ErrConflict
			}
			c.ActiveProfile = value
			return nil
		case "auth_origin":
			if value != "" && !ValidAuthOrigin(value) {
				return ErrInvalid
			}
			if p.AuthenticationOrigin() != (Profile{AuthOrigin: value}).AuthenticationOrigin() {
				p.SelectionOrigin, p.Workspace, p.Team = "", "", ""
			}
			p.AuthOrigin = value
		case "api_origin":
			if value != "" {
				var err error
				value, err = Origin(value, r.Profile == "local")
				if err != nil {
					return err
				}
			}
			if p.APIOrigin != value {
				p.SelectionOrigin, p.Workspace, p.Team = "", "", ""
			}
			p.APIOrigin = value
		case "workspace":
			if value != "" && (!ValidID(value) || r.APIOrigin == "") {
				return ErrInvalid
			}
			if p.Workspace != value || p.SelectionOrigin != r.APIOrigin || value == "" {
				p.Team = ""
			}
			p.Workspace = value
			p.SelectionOrigin = r.APIOrigin
			if value == "" {
				p.SelectionOrigin = ""
			}
		case "team":
			if value != "" && (!ValidID(value) || r.APIOrigin == "") {
				return ErrInvalid
			}
			if p.SelectionOrigin != r.APIOrigin {
				p.Workspace = ""
			}
			p.Team = value
			p.SelectionOrigin = r.APIOrigin
		case "format":
			if value != "" && value != "text" && value != "json" {
				return ErrInvalid
			}
			p.Format = value
		case "timeout":
			if value != "" && !ValidTimeout(value) {
				return ErrInvalid
			}
			p.Timeout = value
		default:
			return ErrInvalid
		}
		c.Profiles[r.Profile] = p
		return nil
	})
}

// ActivateAuthOrigin changes only the successful login's captured profile. A
// concurrent profile or preference change is preserved, never silently replaced.
func (s Store) ActivateAuthOrigin(ctx context.Context, r Resolved, origin string) error {
	if !ValidAuthOrigin(origin) {
		return ErrInvalid
	}
	// Do not create a file on the legacy/default browser path. Still verify that
	// preferences did not change while authorization was in progress.
	if r.AuthOrigin == origin {
		c, err := s.Load()
		if err != nil {
			return err
		}
		if c.ActiveProfile != r.ActiveProfile || c.Profiles[r.Profile] != r.Stored {
			return ErrConflict
		}
		return nil
	}
	return s.Update(ctx, func(c *Config) error {
		if c.ActiveProfile != r.ActiveProfile || c.Profiles[r.Profile] != r.Stored {
			return ErrConflict
		}
		p := r.Stored
		p.AuthOrigin = origin
		p.SelectionOrigin, p.Workspace, p.Team = "", "", ""
		c.Profiles[r.Profile] = p
		return nil
	})
}
