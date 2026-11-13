package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

const workReferenceQuery = `query RiidoCLIWorkReference($reference:String!,$workspaceId:String,$teamId:String){resolveCLIWorkReference(reference:$reference,workspaceId:$workspaceId,teamId:$teamId){schemaVersion workspaceId teamId id key keyNumber componentType}}`

const workReferenceV2Query = `query RiidoCLIWorkReference($reference:String!,$workspaceId:String,$teamId:String){resolveCLIWorkReference(reference:$reference,workspaceId:$workspaceId,teamId:$teamId,locatorVersion:2){schemaVersion workspaceId teamId id key keyNumber componentType}}`

var ErrAmbiguousWorkReference = errors.New("multiple readable works match this key; specify the work ID and team")

var resolvedWorkIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)

type WorkReferenceRequest struct {
	Reference      string `json:"reference"`
	WorkspaceID    string `json:"workspaceId,omitempty"`
	TeamID         string `json:"teamId,omitempty"`
	LocatorVersion int    `json:"-"`
}

func (q WorkReferenceRequest) Valid() bool {
	ref, err := q.parsed()
	if !validLocatorVersion(q.LocatorVersion) || (q.LocatorVersion == 2 && q.WorkspaceID == "") {
		return false
	}
	return err == nil && (ref.Kind == contextcli.WorkRefKey || q.LocatorVersion == 2 && resolvedWorkIDPattern.MatchString(q.Reference)) && ref.Value == q.Reference && len(q.Reference) <= 128 &&
		(q.WorkspaceID != "" || q.TeamID != "") &&
		(q.WorkspaceID == "" || workspaceIDPattern.MatchString(q.WorkspaceID)) &&
		(q.TeamID == "" || resolvedWorkIDPattern.MatchString(q.TeamID))
}

func (q WorkReferenceRequest) parsed() (contextcli.WorkRef, *contextcli.CLIError) {
	if q.LocatorVersion == 2 {
		return contextcli.ParseWorkspaceWorkRef(q.Reference)
	}
	return contextcli.ParseWorkRef(q.Reference)
}

type ResolvedWorkReference struct {
	SchemaVersion string  `json:"schemaVersion"`
	WorkspaceID   string  `json:"workspaceId"`
	TeamID        *string `json:"teamId"`
	ID            string  `json:"id"`
	Key           string  `json:"key"`
	KeyNumber     string  `json:"keyNumber"`
	ComponentType string  `json:"componentType"`
}

func (r Reader) ResolveWorkReference(ctx context.Context, q WorkReferenceRequest) (ResolvedWorkReference, error) {
	if !q.Valid() {
		return ResolvedWorkReference{}, ErrInvalidRequest
	}
	var data struct {
		Work json.RawMessage `json:"resolveCLIWorkReference"`
	}
	query := workReferenceQuery
	if q.LocatorVersion == 2 {
		query = workReferenceV2Query
	}
	if err := r.read(ctx, query, q, &data); err != nil {
		if errors.Is(err, ErrConflict) {
			return ResolvedWorkReference{}, ErrAmbiguousWorkReference
		}
		return ResolvedWorkReference{}, err
	}
	var w ResolvedWorkReference
	if json.Unmarshal(data.Work, &w) != nil || !nullableFieldPresent(data.Work, "teamId") || w.SchemaVersion != locatorSchema("riido-work-reference", q.LocatorVersion) || !workspaceIDPattern.MatchString(w.WorkspaceID) ||
		!validLocatorTeam(q.LocatorVersion, "component", w.ComponentType, w.TeamID) || !resolvedWorkIDPattern.MatchString(w.ID) ||
		(w.TeamID != nil && !resolvedWorkIDPattern.MatchString(*w.TeamID)) ||
		w.KeyNumber == "" || w.ComponentType == "" ||
		(q.WorkspaceID != "" && w.WorkspaceID != q.WorkspaceID) || (q.TeamID != "" && w.TeamID != nil && *w.TeamID != q.TeamID) {
		return ResolvedWorkReference{}, ErrInvalidResponse
	}
	ref, _ := q.parsed()
	if q.LocatorVersion == 2 {
		// Native aliases resolve to the current canonical key. Compare the
		// returned pair internally, never against a historical request key.
		if w.Key == "" || ref.Kind == contextcli.WorkRefID && w.ID != q.Reference {
			return ResolvedWorkReference{}, ErrInvalidResponse
		}
		if strings.ToUpper(w.KeyNumber) != strings.ToUpper(w.Key) {
			parts := strings.SplitN(strings.ToUpper(w.Key), "-", 2)
			if len(parts) != 2 || w.KeyNumber != parts[1] {
				return ResolvedWorkReference{}, ErrInvalidResponse
			}
		}
		return w, nil
	}
	if ref.Kind != contextcli.WorkRefKey {
		if w.ID != q.Reference {
			return ResolvedWorkReference{}, ErrInvalidResponse
		}
		return w, nil
	}
	if strings.ToUpper(w.Key) != q.Reference {
		return ResolvedWorkReference{}, ErrInvalidResponse
	}
	// keyNumber may contain either the full key or just its number.
	if strings.ToUpper(w.KeyNumber) != q.Reference && w.KeyNumber != strings.SplitN(q.Reference, "-", 2)[1] {
		return ResolvedWorkReference{}, ErrInvalidResponse
	}
	return w, nil
}
