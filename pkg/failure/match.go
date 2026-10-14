package failure

import "errors"

func Is(err error, sentinel Sentinel) bool {
	return errors.Is(err, sentinel)
}

func sentinelMatches(sentinel Sentinel, target error) bool {
	if target == nil {
		return false
	}
	var t Sentinel
	if errors.As(target, &t) {
		return sentinel == t
	}
	var t1 *Sentinel
	if errors.As(target, &t1) {
		return t1 != nil && sentinel == *t1
	}
	var t2 Classified
	if errors.As(target, &t2) {
		return sentinel.Layer() == t2.Layer() && sentinel.Kind() == t2.Kind()
	}
	return false
}
