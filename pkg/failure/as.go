package failure

import "errors"

func AsClassified(err error) (Classified, bool) {
	var classified Classified
	if errors.As(err, &classified) {
		return classified, true
	}
	return nil, false
}

func AsOperational(err error) (Operational, bool) {
	var operational Operational
	if errors.As(err, &operational) {
		return operational, true
	}
	return nil, false
}
