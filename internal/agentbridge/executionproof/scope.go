package executionproof

import "context"

// Scope carries the source-owned immutable claim binding separately from its
// tracker. Neither the source binding nor the tracker is serializable evidence.
type Scope struct {
	tracker *Tracker
	source  any
}

func NewScope(covered bool, source any) *Scope {
	return &Scope{tracker: NewTracker(covered), source: source}
}

func (s *Scope) Tracker() *Tracker {
	if s == nil {
		return nil
	}
	return s.tracker
}

func (s *Scope) Source() any {
	if s == nil {
		return nil
	}
	return s.source
}

func (*Scope) String() string   { return "[redacted execution scope]" }
func (*Scope) GoString() string { return "[redacted execution scope]" }

type (
	scopeContextKey struct{}
	proofContextKey struct{}
	contextProof    struct {
		proof Proof
		scope *Scope
	}
)

func ContextWithScope(ctx context.Context, scope *Scope) context.Context {
	return context.WithValue(ctx, scopeContextKey{}, scope)
}

func ScopeFromContext(ctx context.Context) *Scope {
	if ctx == nil {
		return nil
	}
	scope, _ := ctx.Value(scopeContextKey{}).(*Scope)
	return scope
}

func ContextWithProof(ctx context.Context, proof Proof) context.Context {
	return context.WithValue(ctx, proofContextKey{}, contextProof{proof: proof, scope: ScopeFromContext(ctx)})
}

func ProofFromContext(ctx context.Context) (Proof, bool) {
	if ctx == nil {
		return Proof{}, false
	}
	value, ok := ctx.Value(proofContextKey{}).(contextProof)
	if !ok || value.scope == nil || value.scope.Tracker() == nil || value.scope != ScopeFromContext(ctx) || value.proof.Kind != QuiescentKind {
		return Proof{}, false
	}
	return value.proof, true
}
