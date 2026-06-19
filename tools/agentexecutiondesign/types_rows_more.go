package main

type ownerRule struct {
	Field string `json:"field"`
	Owner string `json:"owner"`
	Rule  string `json:"rule"`
}

type streamEvent struct {
	Kind          string `json:"kind"`
	Store         string `json:"store"`
	ClientMeaning string `json:"client_meaning"`
}

type retryPolicy struct {
	Class string `json:"class"`
	Retry string `json:"retry"`
	Rule  string `json:"rule"`
}

type sliceSpec struct {
	Title string   `json:"title"`
	Items []string `json:"items"`
}

type repoOwner struct {
	Repo           string `json:"repo"`
	Responsibility string `json:"responsibility"`
}

type decision struct {
	ID       string `json:"id"`
	Decision string `json:"decision"`
	Default  string `json:"default"`
}
