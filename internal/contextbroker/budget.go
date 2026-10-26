package contextbroker

import (
	"bytes"
	"encoding/json"
	"unicode/utf8"
)

const budgetMinStringRunes = 16

// applyBudget shrinks the longest string values in data (halving each pass,
// marked with "…") until the JSON fits budget characters. It keeps the JSON
// valid; if the structure alone is larger than budget it returns what it
// reached and fits=false.
func applyBudget(data json.RawMessage, budget int) (out json.RawMessage, cut, fits bool) {
	if budget <= 0 || utf8.RuneCount(data) <= budget {
		return data, false, true
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var root any
	if dec.Decode(&root) != nil {
		return data, false, false
	}
	for range 256 {
		encoded, err := json.Marshal(root)
		if err != nil {
			return data, false, false
		}
		if utf8.RuneCount(encoded) <= budget {
			return encoded, cut, true
		}
		set, value := longestString(&root)
		if set == nil || utf8.RuneCountInString(value) < budgetMinStringRunes {
			return encoded, cut, false
		}
		runes := []rune(value)
		set(string(runes[:len(runes)/2]) + "…")
		cut = true
	}
	encoded, err := json.Marshal(root)
	if err != nil {
		return data, false, false
	}
	return encoded, cut, utf8.RuneCount(encoded) <= budget
}

// longestString returns a setter for the longest string reachable from v.
func longestString(v *any) (func(string), string) {
	var bestSet func(string)
	best := ""
	var walk func(get func() any, set func(any))
	walk = func(get func() any, set func(any)) {
		switch node := get().(type) {
		case string:
			if utf8.RuneCountInString(node) > utf8.RuneCountInString(best) {
				best, bestSet = node, func(s string) { set(s) }
			}
		case map[string]any:
			for key := range node {
				walk(func() any { return node[key] }, func(x any) { node[key] = x })
			}
		case []any:
			for i := range node {
				walk(func() any { return node[i] }, func(x any) { node[i] = x })
			}
		}
	}
	walk(func() any { return *v }, func(x any) { *v = x })
	return bestSet, best
}
