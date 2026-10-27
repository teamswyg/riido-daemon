package userauth

import "context"

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
	if err := ctx.Err(); err != nil {
		return Credential{}, err
	}
	if err := a.ctx.Err(); err != nil {
		return Credential{}, err
	}
	select {
	case <-ctx.Done():
		return Credential{}, ctx.Err()
	case <-a.ctx.Done():
		return Credential{}, a.ctx.Err()
	case result := <-a.result:
		if result.err != nil {
			return Credential{}, result.err
		}
		exchangeCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		stop := context.AfterFunc(a.ctx, cancel)
		defer stop()
		credential, err := a.client.exchange(exchangeCtx, result.code, a.verifier, a.RedirectURI)
		if a.ctx.Err() != nil {
			return Credential{}, a.ctx.Err()
		}
		return credential, err
	}
}
