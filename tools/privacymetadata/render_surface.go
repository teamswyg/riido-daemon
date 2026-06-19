package main

import (
	"fmt"
	"strings"
)

func renderSurface(b *strings.Builder, surface SurfaceSnapshot) {
	fmt.Fprintf(b, "## %s\n\n", surface.ID)
	fmt.Fprintf(b, "- owner context: `%s`\n\n", surface.OwnerContext)
	fmt.Fprintln(b, "| Boundary | JSON path |")
	fmt.Fprintln(b, "| --- | --- |")
	for _, path := range surface.AllowedJSONPaths {
		fmt.Fprintf(b, "| allowed | `%s` |\n", path)
	}
	for _, path := range surface.ForbiddenJSONPaths {
		fmt.Fprintf(b, "| forbidden | `%s` |\n", path)
	}
	fmt.Fprintln(b)
}

func renderRules(b *strings.Builder, policy PolicySnapshot) {
	server, _ := findSurface(policy, serverFacingSurfaceID)
	status, _ := findSurface(policy, providerStatusSurfaceID)
	fmt.Fprintln(b, "## Rules")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "- Server-facing metadata allowed paths: `%s`.\n", strings.Join(server.AllowedJSONPaths, "`, `"))
	fmt.Fprintf(b, "- Provider-status sync request allowed paths: `%s`.\n", strings.Join(status.AllowedJSONPaths, "`, `"))
	fmt.Fprintln(b, "- Provider executable paths, workspace absolute paths, tokens, API keys, and raw environments stay local.")
	fmt.Fprintln(b, "- `provider_available` and `provider_login_status` remain C11 projection fields for provider-status sync.")
	fmt.Fprintln(b, "- Capability fingerprint and binary version stay outside this metadata until a separate C3/C4 sync contract exists.")
}
