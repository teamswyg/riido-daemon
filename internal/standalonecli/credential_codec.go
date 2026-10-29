package standalonecli

import (
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func decodeCredential(raw []byte) (userauth.Credential, error) {
	var credential userauth.Credential
	if json.Unmarshal(raw, &credential) != nil || credential.AccessToken == "" || credential.TokenType != "Bearer" {
		return userauth.Credential{}, securecredential.ErrUnavailable
	}
	return credential, nil
}

func encodeCredential(credential userauth.Credential) ([]byte, error) {
	raw, err := json.Marshal(credential)
	if err != nil {
		return nil, securecredential.ErrUnavailable
	}
	return raw, nil
}
