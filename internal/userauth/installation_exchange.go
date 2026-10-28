package userauth

import (
	"context"
	"encoding/asn1"
	"encoding/base64"
	"math/big"
	"net/url"
	"time"
)

func (a *Authorization) exchange(ctx context.Context, code string) (Credential, error) {
	if a.installation == nil {
		return a.client.exchange(ctx, code, a.verifier, a.RedirectURI)
	}
	proof := a.installation.proof
	if code == "" || len(code) > 256 {
		return Credential{}, failure("AUTH_CALLBACK")
	}
	if err := ctx.Err(); err != nil {
		return Credential{}, err
	}
	if !time.Unix(proof.ExpiresAt, 0).After(time.Now()) {
		return Credential{}, context.DeadlineExceeded
	}
	proof.Code = code
	signature, err := a.installation.sign(ctx, proof)
	if ctx.Err() != nil {
		return Credential{}, ctx.Err()
	}
	if err != nil {
		return Credential{}, err
	}
	if !time.Unix(proof.ExpiresAt, 0).After(time.Now()) {
		return Credential{}, context.DeadlineExceeded
	}
	if !validInstallationSignature(signature) {
		return Credential{}, failure("STORAGE_ERROR")
	}
	return a.client.token(ctx, url.Values{"grant_type": {"authorization_code"}, "code": {code}, "code_verifier": {a.verifier}, "redirect_uri": {a.RedirectURI}, "installation_signature": {signature}})
}

func validInstallationSignature(value string) bool {
	if value == "" || len(value) > 96 {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(raw) > 72 || base64.RawURLEncoding.EncodeToString(raw) != value {
		return false
	}
	var signature struct{ R, S *big.Int }
	rest, err := asn1.Unmarshal(raw, &signature)
	return err == nil && len(rest) == 0 && signature.R != nil && signature.S != nil &&
		signature.R.Sign() > 0 && signature.S.Sign() > 0 && signature.R.BitLen() <= 256 && signature.S.BitLen() <= 256
}
