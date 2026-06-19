//go:build windows

package lock

import "time"

func startLockMaintenance(path string) func() {
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(fileLockClaimStaleAfter / 3)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = refreshLockClaim(path)
			case <-stop:
				return
			}
		}
	}()
	return func() {
		close(stop)
		<-done
	}
}
