package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func writeRequiredDocs(t *testing.T, root string) {
	t.Helper()
	writeFile(t, filepath.Join(root, "docs/20-domain/distribution-host-integration.md"), "# Distribution\n")
	writeFile(t, filepath.Join(root, "docs/30-architecture/store-distribution.md"), "# Store\n")
	writeFile(t, filepath.Join(root, "NOTICE.md"), requiredNoticeBody())
}

func requiredNoticeBody() string {
	return "# NOTICE\n" +
		"No source code from any third-party project is directly incorporated\n" +
		"Modified Apache License, Version 2.0\n" +
		"do not redistribute any vendor code or bundle provider CLI executables\n" +
		"No vendored third-party code\n"
}

func writeContract(t *testing.T, root string, value contract) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "packaging/store/riido_daemon_store_distribution.riido.json"), string(data))
}
