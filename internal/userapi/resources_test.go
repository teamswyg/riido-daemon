package userapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func resourceScopeFixture() map[string]any {
	return map[string]any{"schemaVersion": ReadScopeSchemaVersion, "rootId": "work-one", "teamId": "team-one", "status": "available", "works": []any{
		map[string]any{
			"id": "work-one", "kind": "component", "teamId": "team-one", "reasons": []any{map[string]any{"kind": "root"}},
			"links":       []any{map[string]any{"id": "link-one", "title": nil, "url": "https://private.example/path"}},
			"attachments": []any{map[string]any{"id": "file-one", "fileName": "", "fileLocation": "s3://stored/file"}},
			"figma":       map[string]any{"id": "figma-one", "url": "https://www.figma.com/design/x", "fileKey": nil, "nodeId": nil},
			"partsStatus": []any{map[string]any{"name": "links", "status": "available", "complete": true}, map[string]any{"name": "attachments", "status": "available", "complete": true}, map[string]any{"name": "figma", "status": "available", "complete": true}},
		},
		map[string]any{"id": "related", "kind": "component", "teamId": "other-team", "reasons": []any{map[string]any{"kind": "related"}}},
	}}
}

func resourceFixtureJSON(f map[string]any) string {
	raw, _ := json.Marshal(map[string]any{"data": map[string]any{"getComponentReadScope": f}})
	return string(raw)
}

func TestResourcesProjectOnlyRootStoredMetadata(t *testing.T) {
	got, err := relationReader(resourceFixtureJSON(resourceScopeFixture())).Resources(context.Background(), "work-one", "team-one")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Complete || got.WorkID != "work-one" || len(got.Links) != 1 || got.Links[0].Title != nil || *got.Links[0].URL != "https://private.example/path" || len(got.Attachments) != 1 || *got.Attachments[0].FileName != "" || got.Figma == nil {
		t.Fatalf("metadata lost: %+v", got)
	}
	if got.MetadataVersion != nil || len(got.PartsStatus) != 3 {
		t.Fatal("invented version or status")
	}
}

func TestResourcesRejectMalformedRootAndPreserveUnknownAvailability(t *testing.T) {
	for name, change := range map[string]func(map[string]any){
		"schema":         func(f map[string]any) { f["schemaVersion"] = "future" },
		"root":           func(f map[string]any) { f["rootId"] = "other" },
		"team":           func(f map[string]any) { f["teamId"] = "other" },
		"missing-root":   func(f map[string]any) { f["works"] = f["works"].([]any)[1:] },
		"duplicate-root": func(f map[string]any) { f["works"].([]any)[1] = f["works"].([]any)[0] },
		"missing-links":  func(f map[string]any) { delete(f["works"].([]any)[0].(map[string]any), "links") },
		"missing-figma":  func(f map[string]any) { delete(f["works"].([]any)[0].(map[string]any), "figma") },
		"missing-status": func(f map[string]any) { f["works"].([]any)[0].(map[string]any)["partsStatus"] = []any{} },
		"duplicate-status": func(f map[string]any) {
			r := f["works"].([]any)[0].(map[string]any)
			r["partsStatus"] = append(r["partsStatus"].([]any), r["partsStatus"].([]any)[0])
		},
		"deferred-without-page": func(f map[string]any) {
			p := f["works"].([]any)[0].(map[string]any)["partsStatus"].([]any)[0].(map[string]any)
			p["status"] = "deferred"
			p["complete"] = false
		},
		"contradictory-status": func(f map[string]any) {
			f["works"].([]any)[0].(map[string]any)["partsStatus"].([]any)[0].(map[string]any)["complete"] = false
		},
		"missing-filename": func(f map[string]any) {
			delete(f["works"].([]any)[0].(map[string]any)["attachments"].([]any)[0].(map[string]any), "fileName")
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := resourceScopeFixture()
			change(f)
			got, err := relationReader(resourceFixtureJSON(f)).Resources(t.Context(), "work-one", "team-one")
			if got != nil || !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("accepted %+v %v", got, err)
			}
		})
	}
	f := resourceScopeFixture()
	root := f["works"].([]any)[0].(map[string]any)
	root["links"] = []any{}
	root["figma"] = nil
	status := root["partsStatus"].([]any)[0].(map[string]any)
	status["status"] = "unavailable"
	status["complete"] = false
	got, err := relationReader(resourceFixtureJSON(f)).Resources(t.Context(), "work-one", "team-one")
	if err != nil || got.Complete || got.Figma != nil || got.Links == nil {
		t.Fatalf("availability replaced with empty success: %+v %v", got, err)
	}
}

func deferredResourceReader(t *testing.T, metadata string, mutate func(map[string]any), followStatus int) (Reader, *[]int) {
	t.Helper()
	canonical, _ := json.Marshal([]string{"riido-read-part.v1", "component", "work-one", "team-one", "metadata", "{}", "riido_metadata", "json", metadata})
	version := fmt.Sprintf("sha256:%x", sha256.Sum256(canonical))
	offsets := []int{}
	page := func(offset, limit int) map[string]any {
		end := min(offset+limit, len(metadata))
		complete := end == len(metadata)
		status := "partial"
		var next any = end
		if complete {
			status = "available"
			next = nil
		}
		return map[string]any{"selector": map[string]any{}, "resourceId": "work-one", "part": "metadata", "status": status, "source": "riido_metadata", "format": "json", "version": version, "offset": offset, "content": metadata[offset:end], "nextOffset": next, "complete": complete}
	}
	r := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, req *http.Request) (*http.Response, error) {
		var body struct {
			Query     string
			Variables map[string]json.RawMessage
		}
		if json.NewDecoder(req.Body).Decode(&body) != nil {
			t.Fatal("bad request")
		}
		if strings.Contains(body.Query, "getComponentReadScopeBody") {
			if followStatus != http.StatusOK {
				return response(followStatus, `{"errors":[{"message":"private-error"}]}`), nil
			}
			var offset, limit int
			_ = json.Unmarshal(body.Variables["offset"], &offset)
			_ = json.Unmarshal(body.Variables["limit"], &limit)
			if string(body.Variables["version"]) != fmt.Sprintf("%q", version) || string(body.Variables["componentId"]) != `"work-one"` || string(body.Variables["teamId"]) != `"team-one"` || string(body.Variables["part"]) != `"metadata"` {
				t.Fatal("scope/version drift")
			}
			offsets = append(offsets, offset)
			p := page(offset, limit)
			if mutate != nil {
				mutate(p)
			}
			raw, _ := json.Marshal(map[string]any{"data": map[string]any{"getComponentReadScopeBody": p}})
			return response(200, string(raw)), nil
		}
		if !strings.Contains(body.Query, "query RiidoCLIResources") || strings.Contains(body.Query, "mutation") || strings.Contains(body.Query, "meetings") || strings.Contains(body.Query, "github") {
			t.Fatal("not a fixed metadata selection")
		}
		f := resourceScopeFixture()
		root := f["works"].([]any)[0].(map[string]any)
		root["links"] = []any{}
		root["attachments"] = []any{}
		root["figma"] = nil
		root["contentPage"] = page(0, min(1024, len(metadata)))
		for _, v := range root["partsStatus"].([]any) {
			s := v.(map[string]any)
			s["status"] = "deferred"
			s["complete"] = false
		}
		return response(200, resourceFixtureJSON(f)), nil
	})}
	return r, &offsets
}

func resourceMetadataFixture() string {
	root := resourceScopeFixture()["works"].([]any)[0].(map[string]any)
	root["details"] = map[string]any{"id": "work-one"}
	root["unselected"] = strings.Repeat("x", 150000)
	raw, _ := json.Marshal(root)
	return string(raw)
}

func TestResourcesCompleteDeferredMetadataWithOriginalVersion(t *testing.T) {
	r, offsets := deferredResourceReader(t, resourceMetadataFixture(), nil, 200)
	got, err := r.Resources(t.Context(), "work-one", "team-one")
	if err != nil || !got.Complete || len(got.Links) != 1 || got.MetadataVersion == nil {
		t.Fatalf("deferred: %+v %v", got, err)
	}
	if !reflect.DeepEqual(*offsets, []int{1024, 66560, 132096}) {
		t.Fatalf("offsets=%v", *offsets)
	}
	for _, p := range got.PartsStatus {
		if p.Status != "available" || !*p.Complete {
			t.Fatal("deferred state not resolved")
		}
	}
}

func TestResourcesNoPartialOutputWhenContinuationFails(t *testing.T) {
	for _, status := range []int{403, 409, 429} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			r, _ := deferredResourceReader(t, resourceMetadataFixture(), nil, status)
			got, err := r.Resources(t.Context(), "work-one", "team-one")
			if got != nil || err == nil {
				t.Fatal("partial data escaped")
			}
		})
	}
	for name, mutate := range map[string]func(map[string]any){
		"version": func(p map[string]any) { p["version"] = "sha256:" + strings.Repeat("b", 64) },
		"offset":  func(p map[string]any) { p["offset"] = 0 },
		"source":  func(p map[string]any) { p["source"] = "wrong" },
		"format":  func(p map[string]any) { p["format"] = "html" },
		"hash":    func(p map[string]any) { p["content"] = strings.Replace(p["content"].(string), "x", "y", 1) },
	} {
		t.Run(name, func(t *testing.T) {
			r, _ := deferredResourceReader(t, resourceMetadataFixture(), mutate, 200)
			got, err := r.Resources(t.Context(), "work-one", "team-one")
			if got != nil || !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("accepted %+v %v", got, err)
			}
		})
	}
}

func TestResourcesBoundAssemblyWithoutTruncatingList(t *testing.T) {
	base := `{"details":{"id":"work-one"},"links":[],"attachments":[],"figma":null,"padding":""}`
	for _, extra := range []int{0, 1} {
		t.Run(fmt.Sprint(extra), func(t *testing.T) {
			metadata := strings.Replace(base, `"padding":""`, `"padding":"`+strings.Repeat("x", maxResourceMetadataBytes-len(base)+extra)+`"`, 1)
			r, _ := deferredResourceReader(t, metadata, nil, http.StatusOK)
			got, err := r.Resources(t.Context(), "work-one", "team-one")
			if extra == 0 {
				if err != nil || got == nil || !got.Complete || len(got.Links) != 0 {
					t.Fatalf("exact bound rejected: %v", err)
				}
			} else if got != nil || !errors.Is(err, ErrResourceMetadataTooLarge) {
				t.Fatalf("overflow not explicit: %+v %v", got, err)
			}
		})
	}
}

func TestResourcesCancelDuringContinuationDiscardsList(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	r, offsets := deferredResourceReader(t, resourceMetadataFixture(), func(map[string]any) { cancel() }, http.StatusOK)
	got, err := r.Resources(ctx, "work-one", "team-one")
	if got != nil || !errors.Is(err, context.Canceled) || len(*offsets) != 1 {
		t.Fatalf("cancelled read continued: %+v %v offsets=%v", got, err, *offsets)
	}
}
