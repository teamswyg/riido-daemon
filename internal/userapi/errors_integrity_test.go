package userapi

import (
	"encoding/json"
	"errors"
	"testing"
)

// RIID-6008: only the read-scope integrity reason turns a conflict into a
// data-integrity failure; plain conflicts stay version conflicts.
func TestClassifyGraphQLErrorsDataIntegrity(t *testing.T) {
	for name, tc := range map[string]struct {
		raw  string
		want error
	}{
		"integrity":    {`{"message":"x","extensions":{"code":"conflict","reason":"read_scope_integrity"}}`, ErrDataIntegrity},
		"conflict":     {`{"message":"x","extensions":{"code":"conflict"}}`, ErrConflict},
		"other reason": {`{"message":"x","extensions":{"code":"conflict","reason":"version_changed"}}`, ErrConflict},
		"permission":   {`{"message":"x","extensions":{"code":"permission_denied","reason":"read_scope_integrity"}}`, ErrPermission},
	} {
		t.Run(name, func(t *testing.T) {
			if got := classifyGraphQLErrors([]json.RawMessage{json.RawMessage(tc.raw)}); !errors.Is(got, tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
	mixed := []json.RawMessage{
		json.RawMessage(`{"extensions":{"code":"conflict","reason":"read_scope_integrity"}}`),
		json.RawMessage(`{"extensions":{"code":"conflict"}}`),
	}
	if got := classifyGraphQLErrors(mixed); !errors.Is(got, ErrGraphQL) {
		t.Fatal("mixed integrity and version conflicts must not collapse", got)
	}
}
