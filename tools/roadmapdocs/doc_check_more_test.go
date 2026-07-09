package main

import (
	"strings"
	"testing"
)

func TestCheckDocAndMaybeWriteDoc(t *testing.T) {
	dir := t.TempDir()
	m := manifest{GeneratedDoc: "docs/roadmap.md"}
	body := "generated\n"
	if err := maybeWriteDoc(options{Repo: dir, WriteDoc: false}, m, body); err != nil {
		t.Fatal(err)
	}
	if problems := checkDoc(options{Repo: dir, CheckDoc: true}, m, body); len(problems) == 0 {
		t.Fatal("missing generated doc should report a problem")
	}
	if err := maybeWriteDoc(options{Repo: dir, WriteDoc: true}, m, body); err != nil {
		t.Fatal(err)
	}
	if problems := checkDoc(options{Repo: dir, CheckDoc: true}, m, body); len(problems) != 0 {
		t.Fatalf("doc should match: %#v", problems)
	}
	if problems := checkDoc(options{Repo: dir, CheckDoc: true}, m, "other\n"); len(problems) == 0 {
		t.Fatal("doc drift should be reported")
	}
}

func TestRenderIfValidSkipsInvalidManifest(t *testing.T) {
	m := manifest{Title: "Roadmap", Questions: []question{{ID: "Q-1"}}}
	if got := renderIfValid(m, []string{"problem"}); got != "" {
		t.Fatalf("rendered invalid doc: %q", got)
	}
	if got := renderIfValid(m, nil); !strings.Contains(got, generatedHeader) {
		t.Fatalf("valid doc missing generated header: %q", got)
	}
}
