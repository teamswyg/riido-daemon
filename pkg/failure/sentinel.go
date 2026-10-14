package failure

type Sentinel struct {
	layer Layer
	kind  Kind
}

func NewSentinel(layer Layer, kind Kind) Sentinel {
	return Sentinel{layer: layer, kind: kind}
}

func (s Sentinel) Error() string {
	switch {
	case s.layer == "":
		return string(s.kind)
	case s.kind == "":
		return string(s.layer)
	default:
		return string(s.layer) + "/" + string(s.kind)
	}
}

func (s Sentinel) Layer() Layer {
	return s.layer
}

func (s Sentinel) Kind() Kind {
	return s.kind
}

func (s Sentinel) Is(target error) bool {
	return sentinelMatches(s, target)
}
