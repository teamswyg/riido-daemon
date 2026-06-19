package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentManifestAndGeneratedDoc(t *testing.T) {
	if err := run("../..", "docs/30-architecture/runtime-secret-private-evidence.riido.json", "", false, true); err != nil {
		t.Fatal(err)
	}
}

func TestRejectsRawSecretAllowedField(t *testing.T) {
	dir, manifestPath, _ := fixture(t)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	raw := strings.Replace(string(data), `"secret_ref"`, `"token"`, 1)
	mustWrite(t, manifestPath, raw)
	err = run(dir, manifestPath, "", false, false)
	if err == nil || !strings.Contains(err.Error(), `forbidden field "token"`) {
		t.Fatalf("expected forbidden field error, got %v", err)
	}
}

func TestEvidenceOutputIsSanitized(t *testing.T) {
	dir, manifestPath, docPath := fixture(t)
	mustWrite(t, docPath, renderMarkdown(mustLoad(t, manifestPath)))
	evidencePath := filepath.Join(dir, "evidence.json")
	if err := run(dir, manifestPath, evidencePath, false, true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatal(err)
	}
	var evidence evidenceFile
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.Status != "verified" {
		t.Fatalf("status=%q, want verified", evidence.Status)
	}
	if contains(evidence.AllowedFields, "token") || contains(evidence.AllowedOperations, "ssm:GetParameter") {
		t.Fatalf("evidence allowed raw secret path: %+v", evidence)
	}
}

func fixture(t *testing.T) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	docPath := filepath.Join(dir, "doc.md")
	mustWrite(t, filepath.Join(dir, "workflow.yml"), "name: test\n")
	data := `{"schema_version":"riido-runtime-secret-private-evidence.v1","id":"test","title":"Test","generated_doc":"doc.md","workflow":"workflow.yml","evidence_artifact":"artifact","private_owner":"riido-infra","public_scope":["metadata only"],"evidence_kinds":[{"id":"runtime-secret-readiness","actual_id":"actual-runtime-secret-readiness","proves":["shape"],"forbids":["raw token"]}],"allowed_packet_fields":["schema_version","secret_ref"],"allowed_aws_operations":["ssm:DescribeParameters"],"forbidden_aws_operations":["ssm:GetParameter","ssm:GetParameters","ssm:GetParametersByPath","kms:Decrypt","secretsmanager:GetSecretValue"],"forbidden_field_names":["value","token","secret","secret_string","payload","parameter_value","decrypted_value","authorization","bearer"]}`
	mustWrite(t, manifestPath, data)
	return dir, manifestPath, docPath
}
