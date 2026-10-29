package userapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func commentReader(t *testing.T, body string) Reader {
	t.Helper()
	return Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(_ context.Context, r *http.Request) (*http.Response, error) {
		var request struct{ Query string }
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(request.Query, "getComponent(") {
			return response(200, `{"data":{"getComponent":{"id":"work-one"}}}`), nil
		}
		return response(200, body), nil
	})}
}

func TestCommentsRejectNullMissingAndMalformedPageMetadata(t *testing.T) {
	for _, meta := range []string{
		`null`, `{}`, `{"total":null,"currentPage":1,"nextPage":null,"prevPage":null,"lastPage":0}`,
		`{"total":0,"currentPage":1,"prevPage":null,"lastPage":0}`,
		`{"total":0,"currentPage":null,"nextPage":null,"prevPage":null,"lastPage":0}`,
		`{"total":0,"currentPage":2,"nextPage":null,"prevPage":null,"lastPage":0}`,
		`{"total":-1,"currentPage":1,"nextPage":null,"prevPage":null,"lastPage":0}`,
		`{"total":0,"currentPage":1,"nextPage":null,"prevPage":null,"lastPage":1}`,
		`{"total":0,"currentPage":1,"nextPage":"cursor","prevPage":null,"lastPage":0}`,
		`{"total":0,"currentPage":1,"nextPage":2,"prevPage":null,"lastPage":0}`,
		`{"total":2147483648,"currentPage":1,"nextPage":null,"prevPage":null,"lastPage":0}`,
	} {
		body := fmt.Sprintf(`{"data":{"getComponentComments":{"comments":[],"meta":%s}}}`, meta)
		page, err := commentReader(t, body).Comments(context.Background(), "work-one", "team-one", CommentPagination{"1", "20"})
		if !errors.Is(err, ErrInvalidResponse) || page.Meta != nil || page.Comments != nil {
			t.Fatalf("malformed meta accepted: %s error=%v", meta, err)
		}
	}
}

func TestCommentCollectionsRejectMalformedOrOversizedResponses(t *testing.T) {
	for _, command := range []string{"comments", "replies"} {
		for _, body := range []string{`{"data":{}}`, `{"data":{"getComponentComments":null,"getCommentReplies":null}}`, `{"data":{"getComponentComments":{"comments":null,"meta":{"total":0,"currentPage":1,"nextPage":null,"prevPage":null,"lastPage":0}},"getCommentReplies":[null]}}`, strings.Repeat("x", maxBodyBytes+1)} {
			reader := commentReader(t, body)
			var err error
			if command == "comments" {
				_, err = reader.Comments(context.Background(), "work-one", "team-one", CommentPagination{"1", "20"})
			} else {
				_, err = reader.Replies(context.Background(), "work-one", "team-one", "comment-one")
			}
			if !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("%s accepted invalid body: %v", command, err)
			}
		}
	}
}

func TestCommentsTeamMismatchStopsBeforeContentRead(t *testing.T) {
	for _, body := range []string{`{"data":{"getComponent":{"id":"other-work"}}}`, `{"data":{"getComponent":null}}`, `{"errors":[{"message":"private","extensions":{"code":"permission_denied"}}]}`} {
		calls := 0
		reader := Reader{Endpoint: "https://api.example/graphql", Client: doerFunc(func(context.Context, *http.Request) (*http.Response, error) { calls++; return response(200, body), nil })}
		if _, err := reader.Comments(context.Background(), "work-one", "team-one", CommentPagination{"1", "20"}); err == nil || calls != 1 {
			t.Fatal("scope failure did not stop content read")
		}
	}
}

func TestCommentsAllowEmptyPagePastLastAndLargestSafeOffset(t *testing.T) {
	for _, page := range []string{"2", "2147483647"} {
		prev := "1"
		if page == "2147483647" {
			prev = "2147483646"
		}
		body := fmt.Sprintf(`{"data":{"getComponentComments":{"comments":[],"meta":{"total":0,"currentPage":%s,"nextPage":null,"prevPage":%s,"lastPage":0}}}}`, page, prev)
		if _, err := commentReader(t, body).Comments(context.Background(), "work-one", "team-one", CommentPagination{page, "1"}); err != nil {
			t.Fatal(err)
		}
	}
}
