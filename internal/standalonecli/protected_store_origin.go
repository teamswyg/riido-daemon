package standalonecli

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// NewProtectedStore separates credential/request payloads while retaining the
// legacy resource lock and the shared installation repair operation lock.
func NewProtectedStore(native securecredential.Store, resource, authOrigin string) (*ProtectedStore, error) {
	if authOrigin == "" {
		authOrigin = userauth.AuthOriginBrowser
	}
	u, err := url.Parse(resource)
	if err != nil || resource != u.Scheme+"://"+u.Host || (authOrigin != userauth.AuthOriginBrowser && authOrigin != userauth.AuthOriginDesktop) {
		return nil, securecredential.ErrUnavailable
	}
	_, err = userauth.NewClient(userauth.Config{ClientID: "riido-cli", Resource: resource, Endpoints: userauth.Endpoints{AuthorizeURL: resource + "/oauth/authorize", TokenURL: resource + "/oauth/token", RevokeURL: resource + "/oauth/revoke"}})
	if err != nil {
		return nil, securecredential.ErrUnavailable
	}
	custody, operation := native, native
	custody.Key = protectedItemKey("riido-cli:" + resource)
	operation.Key = protectedItemKey("riido-cli-repair:" + resource)
	credential, repair := custody, operation
	if authOrigin == userauth.AuthOriginDesktop {
		credential.Key = protectedItemKey("riido-cli-desktop:" + resource)
		repair.Key = protectedItemKey("riido-cli-repair-desktop:" + resource)
	}
	transfer := native
	transfer.Key = protectedItemKey("riido-cli-transfer:" + authOrigin + ":" + resource)
	return &ProtectedStore{TransferNative: &transfer, Native: &credential, RepairNative: &repair, CustodyLock: &custody, RepairLock: &operation, Resource: resource, AuthOrigin: authOrigin}, nil
}

func protectedItemKey(namespace string) string {
	digest := sha256.Sum256([]byte(namespace))
	return hex.EncodeToString(digest[:])
}
