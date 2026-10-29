package contextbroker

import (
	"bytes"
	"encoding/json"
	"unicode/utf8"
)

const budgetMinStringRunes = 16

// budgetShrinkableKeys: free text only; ids, keys and URLs stay intact for reuse.
var budgetShrinkableKeys = map[string]bool{
	"content": true, "alternate_content": true, "body": true, "html": true, "text": true, "description": true, "summary": true,
}

// applyBudget halves the longest free-text values (marked "…") until the JSON
// fits budget characters, keeping it valid; fits=false when it cannot.
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

// longestString returns a setter for the longest free-text string in v.
func longestString(v *any) (func(string), string) {
	var bestSet func(string)
	best := ""
	var walk func(get func() any, set func(any), key string)
	walk = func(get func() any, set func(any), key string) {
		switch node := get().(type) {
		case string:
			if budgetShrinkableKeys[key] && utf8.RuneCountInString(node) > utf8.RuneCountInString(best) {
				best, bestSet = node, func(s string) { set(s) }
			}
		case map[string]any:
			for k := range node {
				walk(func() any { return node[k] }, func(x any) { node[k] = x }, k)
			}
		case []any:
			for i := range node {
				walk(func() any { return node[i] }, func(x any) { node[i] = x }, key)
			}
		}
	}
	walk(func() any { return *v }, func(x any) { *v = x }, "")
	return bestSet, best
}
