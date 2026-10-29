package saasplane

import (
	"bytes"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"math/big"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func cliLinkOrigin(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && u.Path == "" && u.RawPath == "" && u.RawQuery == "" && !u.ForceQuery && u.Fragment == "" && value == "https://"+u.Host
}

func cliLinkID(value string) bool {
	return value != "" && len(value) <= 512 && utf8.ValidString(value) && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "/\\\x00\r\n\t")
}

func cliLinkDigest(value string) bool {
	raw, err := hex.DecodeString(value)
	return err == nil && len(raw) == 32 && hex.EncodeToString(raw) == value
}

func (c *CLIInstallationLinkClient) validProof(p contextcli.CLIInstallationLinkProof) bool {
	expiry, err := time.Parse(time.RFC3339Nano, p.ExpiresAt)
	if err != nil || !expiry.After(time.Now()) || expiry.UTC().Format(time.RFC3339Nano) != p.ExpiresAt || p.Resource != c.cfg.Resource || p.ClientID != "riido-cli" || p.ExpectedDeviceID != c.cfg.DeviceID || !cliLinkDigest(p.ChallengeHash) || !cliLinkDigest(p.KeyFingerprint) {
		return false
	}
	for _, id := range []string{p.InstallationID, p.IntentID, p.ExpectedDeviceID, p.ApprovedPrincipalID} {
		if !cliLinkID(id) {
			return false
		}
	}
	return true
}

func cliLinkSignature(value string) bool {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(raw) == 0 || len(raw) > 72 || base64.RawURLEncoding.EncodeToString(raw) != value {
		return false
	}
	var signature struct{ R, S *big.Int }
	rest, err := asn1.Unmarshal(raw, &signature)
	if err != nil || len(rest) != 0 || signature.R == nil || signature.S == nil || signature.R.Sign() <= 0 || signature.S.Sign() <= 0 || signature.R.BitLen() > 256 || signature.S.BitLen() > 256 {
		return false
	}
	canonical, err := asn1.Marshal(signature)
	return err == nil && bytes.Equal(canonical, raw)
}
