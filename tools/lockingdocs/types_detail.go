package main

type detailDoc struct {
	SchemaVersion string  `json:"schema_version"`
	LoopSource    string  `json:"loop_source"`
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	GeneratedDoc  string  `json:"generated_doc"`
	Blocks        []block `json:"blocks"`
}

type block struct {
	Kind     string     `json:"kind"`
	Text     string     `json:"text,omitempty"`
	Items    []string   `json:"items,omitempty"`
	Columns  []string   `json:"columns,omitempty"`
	Rows     [][]string `json:"rows,omitempty"`
	Language string     `json:"language,omitempty"`
	Code     string     `json:"code,omitempty"`
}

type sourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}
