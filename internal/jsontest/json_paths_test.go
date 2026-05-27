package jsontest

import (
	"reflect"
	"testing"
)

type jsonPathFixture struct {
	Name     string            `json:"name"`
	Nested   jsonPathNested    `json:"nested"`
	Items    []jsonPathNested  `json:"items"`
	Optional *jsonPathNested   `json:"optional,omitempty"`
	Map      map[string]string `json:"map"`
	Untagged string            `json:",omitempty"`
	Ignored  string            `json:"-"`
	private  string
}

type jsonPathNested struct {
	Value string `json:"value"`
}

func TestStructJSONPaths(t *testing.T) {
	got := StructJSONPaths(reflect.TypeOf(jsonPathFixture{}))
	want := []string{
		"name",
		"nested",
		"nested.value",
		"items",
		"items.value",
		"optional",
		"optional.value",
		"map",
		"Untagged",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("StructJSONPaths() = %#v, want %#v", got, want)
	}
}
