package userapi

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestWorkInspectorRejectsMissingOrNullCollections(t *testing.T) {
	for name, fields := range map[string]string{
		"missing members": `"componentLabels":[]`,
		"missing labels":  `"componentMembers":[]`,
		"null members":    `"componentMembers":null,"componentLabels":[]`,
		"null labels":     `"componentMembers":[],"componentLabels":null`,
	} {
		t.Run(name, func(t *testing.T) {
			reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(context.Context, *http.Request) (*http.Response, error) {
				return response(http.StatusOK, `{"data":{"getComponent":{"id":"work-1","title":"Incomplete",`+fields+`}}}`), nil
			})}
			work, err := reader.Work(t.Context(), "work-1", "team-1")
			if !errors.Is(err, ErrInvalidResponse) || work.ID != "" {
				t.Fatal("incomplete inspector was presented as a successful read", err)
			}
		})
	}
}

func TestWorkInspectorPartialPermissionErrorDiscardsDetails(t *testing.T) {
	reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(context.Context, *http.Request) (*http.Response, error) {
		return response(http.StatusOK, `{"data":{"getComponent":{"id":"work-1","title":"private","componentMembers":[{"id":"private-member"}]}},"errors":[{"extensions":{"code":"permission_denied"}}]}`), nil
	})}
	work, err := reader.Work(t.Context(), "work-1", "team-1")
	if !errors.Is(err, ErrPermission) || work.ID != "" {
		t.Fatal("partial inspector data escaped permission failure", err)
	}
}
