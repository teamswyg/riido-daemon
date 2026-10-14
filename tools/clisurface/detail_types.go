package main

type DetailDoc struct {
	Title  string        `json:"title"`
	Path   string        `json:"path"`
	Blocks []DetailBlock `json:"blocks"`
}

type DetailBlock struct {
	Kind  string   `json:"kind"`
	Text  string   `json:"text,omitempty"`
	Items []string `json:"items,omitempty"`
}
