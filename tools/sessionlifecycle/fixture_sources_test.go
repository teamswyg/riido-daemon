package main

var fixtureSourceFiles = []struct {
	path string
	body string
}{
	{
		"internal/agentbridge/supervisor/provider_event_draft.go",
		"package supervisor\nvar _ = EventSessionPinned\n",
	},
}
