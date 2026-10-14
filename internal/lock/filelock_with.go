package lock

import "context"

// WithFile runs fn while holding an exclusive advisory file lock.
func WithFile(ctx context.Context, path string, fn func() error) (err error) {
	lock, err := AcquireFile(ctx, path)
	if err != nil {
		return err
	}
	defer func() {
		if releaseErr := lock.Release(); err == nil && releaseErr != nil {
			err = releaseErr
		}
	}()

	return fn()
}
