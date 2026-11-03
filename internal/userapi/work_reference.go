package userapi

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

const workReferenceQuery = `query RiidoCLIWorkReference($reference:String!,$workspaceId:String,$teamId:String){resolveCLIWorkReference(reference:$reference,workspaceId:$workspaceId,teamId:$teamId){schemaVersion workspaceId teamId id key keyNumber componentType}}`

var ErrAmbiguousWorkReference = errors.New("multiple readable works match this key; specify the work ID and team")

var resolvedWorkIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)

type WorkReferenceRequest struct {
	Reference   string `json:"reference"`
	WorkspaceID string `json:"workspaceId,omitempty"`
	TeamID      string `json:"teamId,omitempty"`
}

func (q WorkReferenceRequest) Valid() bool {
	ref, err := contextcli.ParseWorkRef(q.Reference)
	return err == nil && ref.Kind == contextcli.WorkRefKey && ref.Value == q.Reference && len(q.Reference) <= 128 &&
		(q.WorkspaceID != "" || q.TeamID != "") &&
		(q.WorkspaceID == "" || workspaceIDPattern.MatchString(q.WorkspaceID)) &&
		(q.TeamID == "" || resolvedWorkIDPattern.MatchString(q.TeamID))
}

type ResolvedWorkReference struct {
	SchemaVersion string `json:"schemaVersion"`
	WorkspaceID   string `json:"workspaceId"`
	TeamID        string `json:"teamId"`
	ID            string `json:"id"`
	Key           string `json:"key"`
	KeyNumber     string `json:"keyNumber"`
	ComponentType string `json:"componentType"`
}

func (r Reader) ResolveWorkReference(ctx context.Context, q WorkReferenceRequest) (ResolvedWorkReference, error) {
	if !q.Valid() {
		return ResolvedWorkReference{}, ErrInvalidRequest
	}
	var data struct {
		Work *ResolvedWorkReference `json:"resolveCLIWorkReference"`
	}
	if err := r.read(ctx, workReferenceQuery, q, &data); err != nil {
		if errors.Is(err, ErrConflict) {
			return ResolvedWorkReference{}, ErrAmbiguousWorkReference
		}
		return ResolvedWorkReference{}, err
	}
	w := data.Work
	if w == nil || w.SchemaVersion != "riido-work-reference.v1" || !workspaceIDPattern.MatchString(w.WorkspaceID) ||
		!resolvedWorkIDPattern.MatchString(w.TeamID) || !resolvedWorkIDPattern.MatchString(w.ID) ||
		strings.ToUpper(w.Key) != q.Reference || w.KeyNumber == "" || w.ComponentType == "" ||
		(q.WorkspaceID != "" && w.WorkspaceID != q.WorkspaceID) || (q.TeamID != "" && w.TeamID != q.TeamID) {
		return ResolvedWorkReference{}, ErrInvalidResponse
	}
	// keyNumber may contain either the full key or just its number.
	if strings.ToUpper(w.KeyNumber) != q.Reference && w.KeyNumber != strings.SplitN(q.Reference, "-", 2)[1] {
		return ResolvedWorkReference{}, ErrInvalidResponse
	}
	return *w, nil
}
