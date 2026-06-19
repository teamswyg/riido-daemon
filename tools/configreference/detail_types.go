package main

type DetailDoc struct {
	Title  string        `json:"title"`
	Path   string        `json:"path"`
	Blocks []DetailBlock `json:"blocks"`
}

type DetailBlock struct {
	Kind    string       `json:"kind"`
	Text    string       `json:"text,omitempty"`
	Items   []string     `json:"items,omitempty"`
	EnvVars []string     `json:"env_vars,omitempty"`
	Table   *DetailTable `json:"table,omitempty"`
}

type DetailTable struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}
