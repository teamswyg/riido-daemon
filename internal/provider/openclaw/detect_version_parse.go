package openclaw

import "strconv"

func parseVersion(s string) ([3]int, bool) {
	m := openClawVersionRE.FindStringSubmatch(s)
	if m == nil {
		return [3]int{}, false
	}

	year, ok := parseVersionPart(m[1])
	if !ok || year < 2020 || year > 2099 {
		return [3]int{}, false
	}
	month, ok := parseVersionPart(m[2])
	if !ok {
		return [3]int{}, false
	}
	day, ok := parseVersionPart(m[3])
	if !ok {
		return [3]int{}, false
	}
	return [3]int{year, month, day}, true
}

func parseVersionPart(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}
