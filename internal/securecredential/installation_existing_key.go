package securecredential

import (
	"context"
	"encoding/json"
)

// ExistingInstallationKey reads only an existing protected installation identity.
// Missing keys and older helpers never fall back to EnsureInstallationKey.
func (s *Store) ExistingInstallationKey(ctx context.Context, resource string) (InstallationPublicKey, error) {
	if s == nil || !installationResource(resource) {
		return InstallationPublicKey{}, ErrUnavailable
	}
	payload := struct {
		Resource string `json:"resource"`
		ClientID string `json:"client_id"`
	}{Resource: resource, ClientID: "riido-cli"}
	data, err := s.installationCall(ctx, resource, "installation-existing-key", payload)
	if err != nil {
		return InstallationPublicKey{}, err
	}
	var key InstallationPublicKey
	if _, err = installationObject(data, "algorithm", "public_key", "key_fingerprint"); err != nil || json.Unmarshal(data, &key) != nil {
		return InstallationPublicKey{}, ErrUnavailable
	}
	if _, err = key.publicKey(); err != nil {
		return InstallationPublicKey{}, err
	}
	return key, nil
}
