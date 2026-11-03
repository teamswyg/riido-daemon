package userauth

import (
	"context"
	"time"
)

func (*Authorization) String() string   { return "[redacted OAuth authorization]" }
func (*Authorization) GoString() string { return "[redacted OAuth authorization]" }
func (a *Authorization) Close()         { a.closeOnce.Do(func() { a.cancel(); _ = a.server.Close() }) }

func (a *Authorization) Wait(ctx context.Context) (Credential, error) {
	a.mu.Lock()
	if a.used {
		a.mu.Unlock()
		return Credential{}, failure("AUTH_USED")
	}
	a.used = true
	a.mu.Unlock()
	defer a.Close()
	authCtx, cancelAuth := a.waitContext()
	defer cancelAuth()
	if err := ctx.Err(); err != nil {
		return Credential{}, err
	}
	if err := authCtx.Err(); err != nil {
		return Credential{}, err
	}
	select {
	case <-ctx.Done():
		return Credential{}, ctx.Err()
	case <-authCtx.Done():
		return Credential{}, authCtx.Err()
	case result := <-a.result:
		if result.err != nil {
			return Credential{}, result.err
		}
		exchangeCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		if result.installation != nil {
			// Only Wait owns this mutable proof; callback handlers send values.
			a.installation = result.installation
			var deadlineCancel context.CancelFunc
			exchangeCtx, deadlineCancel = context.WithDeadline(exchangeCtx, time.Unix(a.installation.proof.ExpiresAt, 0))
			defer deadlineCancel()
		}
		stop := context.AfterFunc(authCtx, cancel)
		defer stop()
		credential, err := a.exchange(exchangeCtx, result.code)
		if authCtx.Err() != nil {
			return Credential{}, authCtx.Err()
		}
		return credential, err
	}
}
