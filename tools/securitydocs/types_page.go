package main

type page struct {
	SchemaVersion string  `json:"schema_version"`
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	LoopSource    string  `json:"loop_source,omitempty"`
	GeneratedDoc  string  `json:"generated_doc"`
	BackTitle     string  `json:"back_title"`
	BackPath      string  `json:"back_path"`
	Blocks        []block `json:"blocks"`
}

type block struct {
	Kind     string     `json:"kind"`
	Text     string     `json:"text,omitempty"`
	Items    []string   `json:"items,omitempty"`
	Links    []link     `json:"links,omitempty"`
	Columns  []string   `json:"columns,omitempty"`
	Rows     [][]string `json:"rows,omitempty"`
	Language string     `json:"language,omitempty"`
	Code     string     `json:"code,omitempty"`
}
