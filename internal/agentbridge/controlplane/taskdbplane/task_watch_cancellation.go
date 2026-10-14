package taskdbplane

import "context"

func (p *Plane) WatchCancellation(_ context.Context, _ string) (<-chan error, error) {
	ch := make(chan error)
	close(ch)
	return ch, nil
}
