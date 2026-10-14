package toolargs

import "maps"

// FromPairs returns a redacted argument map from alternating key/value strings.
// Empty keys are ignored. An odd trailing value is ignored.
func FromPairs(pairs ...string) map[string]string {
	out := map[string]string{}
	for i := 0; i+1 < len(pairs) && len(out) < maxArgs; i += 2 {
		add(out, pairs[i], pairs[i+1])
	}
	return nilIfEmpty(out)
}

// FromValue flattens a provider raw argument object into a bounded, redacted
// string map. Nested fields use dot notation.
func FromValue(value any) map[string]string {
	out := map[string]string{}
	flatten(out, "", value, 0)
	return nilIfEmpty(out)
}

// Clone returns a defensive copy of args.
func Clone(args map[string]string) map[string]string {
	if len(args) == 0 {
		return nil
	}
	out := make(map[string]string, len(args))
	maps.Copy(out, args)
	return out
}
