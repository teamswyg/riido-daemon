package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion              string             `json:"schema_version"`
	ID                         string             `json:"id"`
	Title                      string             `json:"title"`
	GeneratedDoc               string             `json:"generated_doc"`
	Workflow                   string             `json:"workflow"`
	EvidenceArtifact           string             `json:"evidence_artifact"`
	RiidoTask                  string             `json:"riido_task"`
	Purpose                    string             `json:"purpose"`
	ProviderValidationManifest string             `json:"provider_validation_manifest"`
	RealCLIObservationManifest string             `json:"real_cli_observation_manifest"`
	SecurityDecisionRef        string             `json:"security_decision_ref"`
	SecurityDecisionLink       string             `json:"security_decision_link"`
	GatePolicy                 []string           `json:"gate_policy"`
	InstructionProbe           instructionProbe   `json:"instruction_probe"`
	ChangeProcedure            []string           `json:"change_procedure"`
	DetailDocs                 []detailDoc        `json:"detail_docs"`
	SourceChecks               []sourceCheck      `json:"source_checks"`
	Assertions                 []string           `json:"assertions"`
	ProviderValidation         providerValidation `json:"-"`
	RealCLIObservation         realCLIObservation `json:"-"`
}
