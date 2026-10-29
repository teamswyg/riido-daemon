package userapi

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestTypedPartUsesFixedBodyRouteAndNormalizedSelector(t *testing.T) {
	for _, tc := range []struct {
		name, kind, part, input, normalized string
	}{
		{"work metadata", "component", "metadata", `{}`, `{}`},
		{"backlog comments", "backlog", "comments_page", `{}`, `{"page":1,"limit":20}`},
		{"intake replies", "intake", "comments_page", `{"parentCommentId":"comment-one","page":2,"limit":3}`, `{"parentCommentId":"comment-one","page":2,"limit":3}`},
		{"meeting board", "meeting", "board_page", `{"boardCursor":"board-next","limit":5}`, `{"boardCursor":"board-next","limit":5}`},
		{"meeting metadata", "meeting", "metadata", `{}`, `{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			version := "sha256:" + strings.Repeat("a", 64)
			var q ReadBodyRequest
			if err := json.Unmarshal([]byte(`{"componentId":"resource-one","teamId":"team-one","part":"`+tc.part+`","version":"`+version+`","offset":0,"limit":64,"selector":`+tc.input+`}`), &q); err != nil {
				t.Fatal(err)
			}
			q.Kind = tc.kind
			calls := 0
			_, field, idKey := readBodyOperation(tc.kind)
			reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, req *http.Request) (*http.Response, error) {
				calls++
				var input struct {
					Query     string
					Variables map[string]json.RawMessage
				}
				if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
					t.Fatal(err)
				}
				var got, want any
				_ = json.Unmarshal(input.Variables["selector"], &got)
				_ = json.Unmarshal([]byte(tc.normalized), &want)
				if !reflect.DeepEqual(got, want) || string(input.Variables[idKey]) != `"resource-one"` || !strings.Contains(input.Query, "selector:$selector") {
					t.Fatal("typed selector or actual resource was not forwarded to a fixed body query")
				}
				return response(200, `{"data":{"`+field+`":{"resourceId":"resource-one","part":"`+tc.part+`","source":"riido_`+tc.part+`","format":"json","version":"`+version+`","offset":0,"content":"{}","nextOffset":null,"complete":true,"status":"available","selector":`+tc.normalized+`}}}`), nil
			})}
			page, err := reader.Body(t.Context(), q)
			if err != nil || page.Content != "{}" || calls != 1 {
				t.Fatalf("fixed typed part not readable: calls=%d err=%v", calls, err)
			}
		})
	}
}
