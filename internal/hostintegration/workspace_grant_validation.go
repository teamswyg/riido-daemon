package hostintegration

import (
	"errors"
	"fmt"
	"strings"
)

// Validate checks C11 grant rules. It does not inspect security-scoped bookmark
// bytes or Windows shell item tokens; adapters own that OS-specific proof.
func (r WorkspaceGrantRecord) Validate() error {
	var errs []error
	workspaceID := strings.TrimSpace(r.WorkspaceID)
	rootPath := strings.TrimSpace(r.RootPath)
	if workspaceID == "" {
		errs = append(errs, errors.New("workspace id is required"))
	}
	if !r.Channel.Valid() {
		errs = append(errs, fmt.Errorf("unknown distribution channel %q", r.Channel))
	}
	if !r.HostOS.Valid() {
		errs = append(errs, fmt.Errorf("unknown host OS %q", r.HostOS))
	}
	if !r.Method.Valid() {
		errs = append(errs, fmt.Errorf("unknown workspace grant method %q", r.Method))
	}
	if rootPath == "" {
		errs = append(errs, errors.New("workspace root path is required"))
	}
	if r.GrantedAt.IsZero() {
		errs = append(errs, errors.New("granted time is required"))
	}
	if !r.RevokedAt.IsZero() && r.RevokedAt.Before(r.GrantedAt) {
		errs = append(errs, errors.New("revoked time must be after granted time"))
	}
	if r.Channel.Valid() && r.HostOS.Valid() && r.Method.Valid() {
		if err := validateWorkspaceGrantChannelMethod(r.Channel, r.HostOS, r.Method); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
