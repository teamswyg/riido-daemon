package contextbroker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

type cacheFetchFunc func(context.Context, Query) (Result, error)

func (f cacheFetchFunc) FetchAssignmentContext(ctx context.Context, q Query) (Result, error) {
	return f(ctx, q)
}

func cacheRequest(t *testing.T, b *Broker) contextcli.ContextReadRequest {
	t.Helper()
	req := request(issueForTest(t, b, saasBinding()), "context part")
	req.Args = map[string]string{"work": "work-one", "cursor": "page-one"}
	return req
}

func cacheResult() Result {
	return Result{Data: json.RawMessage(batchData), CapturedAt: time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC), ContentDigest: "sha256:" + strings.Repeat("a", 64), Sources: []ResultSource{{Kind: "task", ID: "work-one", Part: "body", Version: "v1"}}}
}

func TestContentCacheRevalidatesBeforeReusingBody(t *testing.T) {
	full := cacheResult()
	hit := full
	hit.Data = json.RawMessage("null")
	hit.NotModified = true
	f := &sequenceFetcher{results: []Result{full, hit}, errs: []error{nil, nil}}
	b := New("", "")
	b.Fetcher = f
	token := issueForTest(t, b, saasBinding())
	req := request(token, "context part")
	req.Args = map[string]string{"work": "work-one", "cursor": "page-one"}
	first := b.Handle(req, activeSet("run-1"))
	second := b.Handle(req, activeSet("run-1"))
	if first.Error != nil || second.Error != nil {
		t.Fatalf("errors: %v / %v", first.Error, second.Error)
	}
	if len(f.queries) != 2 || !f.queries[0].CacheEnabled || f.queries[0].KnownDigest != "" || f.queries[1].KnownDigest != full.ContentDigest {
		t.Fatal("each body read must revalidate the exact candidate")
	}
	a, _ := json.Marshal(first)
	c, _ := json.Marshal(second)
	if string(a) != string(c) {
		t.Fatal("conditional read changed body, continuation or source metadata")
	}
}

func TestContentCacheRefusesInvalidConditionalResponses(t *testing.T) {
	cases := map[string]func(*Result){
		"digest":         func(r *Result) { r.ContentDigest = "sha256:" + strings.Repeat("b", 64) },
		"malformed":      func(r *Result) { r.ContentDigest = "bad" },
		"missing_digest": func(r *Result) { r.ContentDigest = "" },
		"body":           func(r *Result) { r.Data = json.RawMessage(batchData) },
		"missing_null":   func(r *Result) { r.Data = nil },
		"live":           func(r *Result) { r.Live = true },
		"capture_time":   func(r *Result) { r.CapturedAt = r.CapturedAt.Add(time.Second) },
		"status":         func(r *Result) { r.Truncated = true },
		"source":         func(r *Result) { r.Sources[0].Version = "v2" },
		"source_updated": func(r *Result) { r.Sources[0].UpdatedAt = "2026-11-02T01:00:00Z" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			full := cacheResult()
			hit := copyContentResult(full)
			hit.Data = json.RawMessage("null")
			hit.NotModified = true
			mutate(&hit)
			f := &sequenceFetcher{results: []Result{full, hit, full}, errs: []error{nil, nil, nil}}
			b := New("", "")
			b.Fetcher = f
			req := cacheRequest(t, b)
			if env := b.Handle(req, activeSet("run-1")); env.Error != nil {
				t.Fatal(env.Error)
			}
			if env := b.Handle(req, activeSet("run-1")); env.Error == nil || env.Data != nil || len(env.Sources) != 0 {
				t.Fatal("invalid conditional response published cached content")
			}
			b.Handle(req, activeSet("run-1"))
			if f.queries[2].KnownDigest != "" {
				t.Fatal("invalid candidate retained")
			}
		})
	}
}

func TestContentCacheDoesNotMaskServerRefusal(t *testing.T) {
	for _, status := range []int{401, 403, 404, 409, 410, 429, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			f := &sequenceFetcher{results: []Result{cacheResult(), {}, cacheResult()}, errs: []error{nil, &FetchError{Status: status, Message: "refused"}, nil}}
			b := New("", "")
			b.Fetcher = f
			req := cacheRequest(t, b)
			b.Handle(req, activeSet("run-1"))
			if env := b.Handle(req, activeSet("run-1")); env.Error == nil || env.Data != nil || len(env.Sources) != 0 {
				t.Fatal("server refusal served stale data")
			}
			b.Handle(req, activeSet("run-1"))
			if len(f.queries) != 3 || f.queries[2].KnownDigest != "" {
				t.Fatal("refusal retried or stale candidate kept")
			}
		})
	}
}

func TestContentCacheMissAndLegacyReplaceCandidate(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprint(legacy), func(t *testing.T) {
			changed := cacheResult()
			changed.Data = json.RawMessage(`{"source_id":"work-one","content":"changed"}`)
			changed.ContentDigest = "sha256:" + strings.Repeat("b", 64)
			if legacy {
				changed.ContentDigest = ""
			}
			f := &sequenceFetcher{results: []Result{cacheResult(), changed, changed}, errs: []error{nil, nil, nil}}
			b := New("", "")
			b.Fetcher = f
			req := cacheRequest(t, b)
			b.Handle(req, activeSet("run-1"))
			out := b.Handle(req, activeSet("run-1"))
			data, _ := json.Marshal(out.Data)
			if out.Error != nil || !strings.Contains(string(data), "changed") {
				t.Fatal("changed full response was not used")
			}
			b.Handle(req, activeSet("run-1"))
			if f.queries[2].KnownDigest != changed.ContentDigest {
				t.Fatal("candidate not replaced")
			}
		})
	}
}

func TestContentCacheSeparatesPageBudgetAndCapability(t *testing.T) {
	f := &fakeFetcher{result: cacheResult()}
	b := New("", "")
	b.Fetcher = f
	req := cacheRequest(t, b)
	b.Handle(req, activeSet("run-1"))
	for _, args := range []map[string]string{{"work": "work-two", "cursor": "page-one"}, {"work": "work-one", "cursor": "page-two"}, {"work": "work-one", "cursor": "page-one", "budget_chars": "9000"}} {
		req.Args = args
		b.Handle(req, activeSet("run-1"))
		if f.queries[len(f.queries)-1].KnownDigest != "" {
			t.Fatal("different representation reused candidate")
		}
	}
	old := req.RunCapability
	req = cacheRequest(t, b)
	b.Handle(req, activeSet("run-1"))
	if f.queries[len(f.queries)-1].KnownDigest != "" {
		t.Fatal("new capability inherited candidate")
	}
	if _, ok := b.Registry.Lookup(old); ok {
		t.Fatal("old capability retained")
	}
}

func TestContentCacheNeverReturnsAcrossMidFlightRevoke(t *testing.T) {
	for _, replace := range []bool{false, true} {
		t.Run(fmt.Sprint(replace), func(t *testing.T) {
			b := New("", "")
			req := cacheRequest(t, b)
			b.Fetcher = cacheFetchFunc(func(_ context.Context, _ Query) (Result, error) {
				if replace {
					issueForTest(t, b, saasBinding())
				} else {
					b.Registry.Revoke("run-1")
				}
				return cacheResult(), nil
			})
			out := b.Handle(req, activeSet("run-1"))
			if out.Error == nil || out.Data != nil {
				t.Fatal("revoked read published")
			}
			if b.Registry.cacheBytes != 0 {
				t.Fatal("revoked response restored cache")
			}
		})
	}
}

func TestContentCacheBoundsCopiesAndCleanup(t *testing.T) {
	b := New("", "")
	token := issueForTest(t, b, saasBinding())
	r := b.Registry
	result := cacheResult()
	key := contentCacheKey(Query{View: viewContent})
	r.storeContent(token, key, result)
	result.Data[0] = 'x'
	result.Sources[0].ID = "mutated"
	got, ok := r.cachedContent(token, key)
	if !ok || !json.Valid(got.Data) || got.Sources[0].ID == "mutated" {
		t.Fatal("stored alias changed")
	}
	got.Data[0] = 'x'
	again, _ := r.cachedContent(token, key)
	if !json.Valid(again.Data) {
		t.Fatal("read alias changed")
	}
	for i := range contentCacheMaxEntries + 2 {
		r.storeContent(token, contentCacheKey(Query{Cursor: fmt.Sprint(i)}), cacheResult())
	}
	if _, ok := r.cachedContent(token, key); ok {
		t.Fatal("old entry not evicted")
	}
	large := cacheResult()
	large.Data = make([]byte, contentCacheMaxEntryBytes)
	r.storeContent(token, key, large)
	if _, ok := r.cachedContent(token, key); ok {
		t.Fatal("oversized entry stored")
	}
	r.Revoke("run-1")
	if r.cacheBytes != 0 || len(r.cache) != 0 {
		t.Fatal("revoke did not release content")
	}
	for i := range 30 {
		binding := saasBinding()
		binding.RunKey = fmt.Sprint(i)
		tok := issueForTest(t, b, binding)
		large.Data = make([]byte, contentCacheMaxEntryBytes-1024)
		r.storeContent(tok, key, large)
	}
	if r.cacheBytes > contentCacheMaxBytes {
		t.Fatal("global cache budget exceeded")
	}
}

func TestContentCacheConcurrentReadAndRevoke(t *testing.T) {
	b := New("", "")
	token := issueForTest(t, b, saasBinding())
	key := contentCacheKey(Query{View: viewContent})
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 50 {
				b.Registry.storeContent(token, key, cacheResult())
				b.Registry.cachedContent(token, key)
				b.Registry.forgetContent(token, key)
			}
		})
	}
	b.Registry.Revoke("run-1")
	wg.Wait()
	if b.Registry.cacheBytes != 0 {
		t.Fatal("concurrent response restored revoked cache")
	}
}

func TestContentCacheFallbackClearsNegotiation(t *testing.T) {
	f := &sequenceFetcher{results: []Result{{}, cacheResult()}, errs: []error{&FetchError{Status: 400, Message: `unknown field "max_bytes"`}, nil}}
	b := New("", "")
	b.Fetcher = f
	req := cacheRequest(t, b)
	if out := b.Handle(req, activeSet("run-1")); out.Error != nil {
		t.Fatal(out.Error)
	}
	if len(f.queries) != 2 || f.queries[1].CacheEnabled || f.queries[1].KnownDigest != "" || b.Registry.cacheBytes != 0 {
		t.Fatal("legacy fallback cached different representation")
	}
}

func TestContentCacheTransportFailureDoesNotReturnData(t *testing.T) {
	f := &sequenceFetcher{results: []Result{cacheResult(), {}}, errs: []error{nil, errors.New("offline")}}
	b := New("", "")
	b.Fetcher = f
	req := cacheRequest(t, b)
	b.Handle(req, activeSet("run-1"))
	if out := b.Handle(req, activeSet("run-1")); out.Error == nil || out.Data != nil {
		t.Fatal("offline cache served")
	}
}

func TestContentCacheRechecksCapabilityAfterActivityWait(t *testing.T) {
	for _, running := range []bool{false, true} {
		t.Run(fmt.Sprint(running), func(t *testing.T) {
			b := New("", "")
			b.Fetcher = &fakeFetcher{result: cacheResult()}
			req := cacheRequest(t, b)
			calls := 0
			replacement := ""
			out := b.Handle(req, func(string) (bool, bool) {
				calls++
				if calls == 2 {
					replacement = issueForTest(t, b, saasBinding())
					return running, true
				}
				return true, true
			})
			if out.Error == nil || out.Data != nil {
				t.Fatal("retired capability passed the activity wait")
			}
			if _, ok := b.Registry.Lookup(replacement); !ok {
				t.Fatal("stale activity result revoked replacement capability")
			}
		})
	}
}

func TestContentCacheEvictionRetiresExactCapability(t *testing.T) {
	b := New("", "")
	token := issueForTest(t, b, saasBinding())
	key := contentCacheKey(Query{View: viewContent})
	b.Registry.storeContent(token, key, cacheResult())
	for i := range maxBindings {
		binding := saasBinding()
		binding.RunKey = fmt.Sprintf("eviction-%d", i)
		issueForTest(t, b, binding)
	}
	if _, ok := b.Registry.Lookup(token); ok || b.Registry.cacheBytes != 0 {
		t.Fatal("binding eviction retained body")
	}
	newToken := issueForTest(t, b, saasBinding())
	b.Registry.storeContent(newToken, key, cacheResult())
	b.Registry.RevokeCapability(token)
	if _, ok := b.Registry.cachedContent(newToken, key); !ok {
		t.Fatal("old terminal event removed replacement cache")
	}
}
