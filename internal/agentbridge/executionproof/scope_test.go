package executionproof

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestScopePreservesClaimAcrossChildContextAndRedactsSource(t *testing.T) {
	source := &struct{ Secret string }{Secret: "private-claim-binding-fixture"}
	scope := NewScope(true, source)
	ctx := ContextWithScope(context.Background(), scope)
	child, cancel := context.WithCancel(ctx)
	cancel()
	got := ScopeFromContext(child)
	if got != scope || got == nil || got.Source() != source || got.Tracker() != scope.Tracker() {
		t.Fatal("child execution lost its original private scope")
	}
	proof, ok := got.Tracker().Seal()
	if !ok || proof.Kind != "local_invocations_quiescent" || proof.StartedInvocations != 0 {
		t.Fatalf("covered zero-start scope = %#v, %v", proof, ok)
	}
	encoded, err := json.Marshal(scope) //nolint:staticcheck // Verify opaque scope JSON cannot expose the private binding.
	if err != nil || string(encoded) != "{}" {
		t.Fatalf("scope JSON = %q, %v", encoded, err)
	}
	for _, value := range []any{scope, scope.Tracker(), ctx} {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			if strings.Contains(fmt.Sprintf(format, value), source.Secret) {
				t.Fatal("private scope leaked through formatting")
			}
		}
	}
}

func TestScopeAndProofContextDoNotInventOrCrossClaimEvidence(t *testing.T) {
	ctx := context.Background()
	if ScopeFromContext(ctx) != nil || ScopeFromContext(nil) != nil { //nolint:staticcheck // Explicit nil-boundary regression for legacy task contexts.
		t.Fatal("missing context invented scope")
	}
	var missing *Scope
	if missing.Tracker() != nil || missing.Source() != nil {
		t.Fatal("nil scope invented source")
	}
	if _, ok := ProofFromContext(ctx); ok {
		t.Fatal("missing context invented proof")
	}
	if _, ok := ProofFromContext(nil); ok { //nolint:staticcheck // Missing context must never become positive evidence.
		t.Fatal("nil context invented proof")
	}
	proof := Proof{Kind: "local_invocations_quiescent", StartedInvocations: 2}
	ctx = ContextWithScope(ctx, NewScope(true, "proof-claim"))
	ctx = ContextWithProof(ctx, proof)
	if got, ok := ProofFromContext(ctx); !ok || got != proof {
		t.Fatalf("proof context = %#v, %v", got, ok)
	}
	ctx = ContextWithProof(ctx, Proof{Kind: "provider_completed", StartedInvocations: 2})
	if _, ok := ProofFromContext(ctx); ok {
		t.Fatal("provider result became termination proof")
	}
	first := NewScope(true, "first")
	second := NewScope(false, "second")
	ctx = ContextWithProof(ContextWithScope(ctx, first), proof)
	ctx = ContextWithScope(ctx, second)
	if ScopeFromContext(ctx) != second {
		t.Fatal("new claim inherited old scope")
	}
	if _, ok := ProofFromContext(ctx); ok {
		t.Fatal("new claim inherited previous claim proof")
	}
	if _, ok := ScopeFromContext(ctx).Tracker().Seal(); ok {
		t.Fatal("uncovered new claim inherited old coverage")
	}
}

func TestProofContextRequiresScopeAndTracker(t *testing.T) {
	proof := Proof{Kind: "local_invocations_quiescent", StartedInvocations: 1}
	for _, scope := range []*Scope{nil, {}} {
		ctx := ContextWithProof(ContextWithScope(context.Background(), scope), proof)
		if got, ok := ProofFromContext(ctx); ok || got != (Proof{}) {
			t.Fatalf("missing claim source became proof: %#v, %v", got, ok)
		}
	}
}
