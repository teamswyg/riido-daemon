package contextbroker

import (
	"encoding/json"
	"testing"
	"time"
)

func TestExplicitReadsPreserveBoundRootAndLivePagination(t *testing.T) {
	for command, view := range map[string]string{"work activity": "activity", "work body": "body", "thread list": "threads", "thread show": "thread", "work comments": "comments"} {
		t.Run(command, func(t *testing.T) {
			f := &fakeFetcher{result: Result{Data: json.RawMessage(`{"resource":{"nextPage":2,"complete":false,"version":"source-version"}}`), Live: true, Truncated: true, CapturedAt: time.Now()}}
			b := New("", "")
			b.Fetcher = f
			token := issueForTest(t, b, saasBinding())
			r := request(token, command)
			r.Args = map[string]string{"work": "RAX-43", "page": "1", "version": "source-version", "thread": "thread-one", "principal_id": "forged"}
			env := b.Handle(r, activeSet("run-1"))
			if env.Error != nil {
				t.Fatal(env.Error)
			}
			q := f.queries[0]
			if q.TaskID != "task-1" || q.Target != "RAX-43" || q.View != view || q.Options["page"] != "1" || q.Options["principal_id"] != "" {
				t.Fatalf("wrong query %+v", q)
			}
			if env.Page != nil {
				t.Fatal("live page received snapshot cursor contract")
			}
			if q.CacheEnabled {
				t.Fatal("live result cached")
			}
		})
	}
}
