package hostintegration

// PrivacyMetadataAllowlist is the executable C10/C11 policy artifact that
// keeps store metadata, public privacy policy, and server request fields aligned.
type PrivacyMetadataAllowlist struct {
	SchemaVersion string                         `json:"schema_version"`
	Surfaces      []PrivacyMetadataSurfacePolicy `json:"surfaces"`
}

// PrivacyMetadataSurfacePolicy describes one JSON boundary where local daemon
// facts may cross toward C10.
type PrivacyMetadataSurfacePolicy struct {
	ID                 string   `json:"id"`
	OwnerContext       string   `json:"owner_context"`
	AllowedJSONPaths   []string `json:"allowed_json_paths"`
	ForbiddenJSONPaths []string `json:"forbidden_json_paths"`
}
