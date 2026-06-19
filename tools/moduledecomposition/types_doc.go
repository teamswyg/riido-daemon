package main

type detailDoc struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

type checkResult struct {
	Name   string `json:"name"`
	File   string `json:"file,omitempty"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail,omitempty"`
}

type packageInfo struct {
	ImportPath string
	Imports    []string
	Name       string
}
