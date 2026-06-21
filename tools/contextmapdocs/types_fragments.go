package main

type aclFragment struct {
	SchemaVersion string   `json:"schema_version"`
	LoopSource    string   `json:"loop_source"`
	Rows          []aclRow `json:"rows"`
}

type aclRow struct {
	ACL    string `json:"acl"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

type dependencyFragment struct {
	SchemaVersion          string   `json:"schema_version"`
	LoopSource             string   `json:"loop_source"`
	Diagram                []string `json:"diagram"`
	Notes                  []string `json:"notes"`
	ForbiddenImports       []string `json:"forbidden_imports"`
	RetiredPrivateRepoRule string   `json:"retired_private_repo_rule"`
}

type figmaFragment struct {
	SchemaVersion               string         `json:"schema_version"`
	LoopSource                  string         `json:"loop_source"`
	Sections                    []figmaSection `json:"sections"`
	DirectHostHelperRule        string         `json:"direct_host_helper_rule,omitempty"`
	AssignmentAuthorizationRule string         `json:"assignment_authorization_rule,omitempty"`
}

type onboardingFragment struct {
	SchemaVersion   string         `json:"schema_version"`
	LoopSource      string         `json:"loop_source"`
	Sections        []figmaSection `json:"sections"`
	MustNotHardcode []string       `json:"must_not_hardcode"`
}
