package standalonecli

import (
	"context"
	"sync"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

func (n *protectedNativeFixture) TryLock(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !n.operation.TryLock() {
		return nil, securecredential.ErrBusy
	}
	var once sync.Once
	return func() { once.Do(n.operation.Unlock) }, nil
}
