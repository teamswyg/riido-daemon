package main

type detailDoc struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

type sourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}

type instructionProbe struct {
	PublicCI  string              `json:"public_ci"`
	Builder   string              `json:"builder"`
	Validator string              `json:"validator"`
	Providers []instructionTarget `json:"providers"`
}

type instructionTarget struct {
	Provider string `json:"provider"`
	Marker   string `json:"marker"`
	Surface  string `json:"surface"`
}
