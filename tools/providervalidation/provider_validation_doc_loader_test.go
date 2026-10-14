package providervalidation

import (
	"path/filepath"
	"strings"
	"testing"
)

func readIntegrationMatrixText(t *testing.T, root string) string {
	t.Helper()
	files := []string{
		"docs/30-architecture/integration-matrix.md",
		"docs/30-architecture/integration-matrix/gate-policy.md",
		"docs/30-architecture/integration-matrix/provider-matrix.md",
		"docs/30-architecture/integration-matrix/assertions.md",
		"docs/30-architecture/integration-matrix/instruction-effectiveness.md",
		"docs/30-architecture/integration-matrix/change-procedure.md",
	}
	var body strings.Builder
	for _, file := range files {
		body.WriteString(readText(t, filepath.Join(root, file)))
		body.WriteString("\n")
	}
	return body.String()
}
