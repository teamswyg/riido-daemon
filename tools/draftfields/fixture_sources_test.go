package main

var fixtureSourceFiles = []struct {
	path string
	body string
}{
	{
		"internal/agentbridge/supervisor/provider_event_draft.go",
		"package supervisor\nfunc providerEventDraft() (ir.EventType, map[string]any, bool) { return \"\", nil, false }\n",
	},
}
