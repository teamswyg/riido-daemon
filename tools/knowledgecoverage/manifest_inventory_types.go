package main

type manifestInventory struct {
	Count   int                   `json:"count"`
	Groups  []manifestGroupCount  `json:"groups"`
	Samples []manifestGroupSample `json:"samples"`
}

type manifestGroupCount struct {
	Group string `json:"group"`
	Count int    `json:"count"`
}

type manifestGroupSample struct {
	Group string   `json:"group"`
	Paths []string `json:"paths"`
}
