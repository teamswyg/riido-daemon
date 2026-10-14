package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

//go:embed qa_i18n.generated.json
var qaI18NFS embed.FS
var qaI18NPlaceholderPattern = regexp.MustCompile(`\{[A-Za-z0-9_.-]+\}`)

func qaI18NJSON() ([]byte, error) {
	body, err := qaI18NFS.ReadFile("qa_i18n.generated.json")
	if err != nil {
		return nil, fmt.Errorf("read QA i18n catalog: %w", err)
	}
	return body, nil
}

func qaI18NContract() map[string]any {
	body, err := qaI18NJSON()
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return map[string]any{"error": err.Error()}
	}
	return out
}

func qaI18NScenario() scenario {
	spec := qaI18NContract()
	if errText, ok := spec["error"].(string); ok {
		return scenario{
			ID:             "contract.ui.i18n_dsl",
			Status:         statusFailed,
			FailureSummary: errText,
			Observed:       spec,
		}
	}
	coverage := qaI18NTranslationCoverage(spec)
	spec["capture_owner"] = "qa-codex"
	spec["source_dsl"] = "docs/30-architecture/qa-i18n.dsl.json"
	spec["translation_coverage"] = coverage
	status := statusPassed
	if coverage["passed"] != true {
		status = statusFailed
	}
	return scenario{
		ID:       "contract.ui.i18n_dsl",
		Status:   status,
		Observed: spec,
	}
}

func qaI18NTranslationCoverage(spec map[string]any) map[string]any {
	defaultLocale, _ := spec["default_locale"].(string)
	fallbackLocale, _ := spec["fallback_locale"].(string)
	locales := []string{defaultLocale, fallbackLocale}
	localeColumns := map[string]int{defaultLocale: 1, fallbackLocale: 2}
	localeErrors := []string{}
	if defaultLocale == "" {
		localeErrors = append(localeErrors, "default_locale is empty")
	}
	if fallbackLocale == "" {
		localeErrors = append(localeErrors, "fallback_locale is empty")
	}
	if defaultLocale != "" && defaultLocale == fallbackLocale {
		localeErrors = append(localeErrors, "default_locale and fallback_locale must differ")
	}
	namespaceRows := []map[string]any{}
	missing := []map[string]string{}
	placeholderMismatches := []map[string]any{}
	duplicateKeys := []string{}
	translatedByLocale := map[string]int{defaultLocale: 0, fallbackLocale: 0}
	messageCount := 0
	namespaces, _ := spec["namespaces"].([]any)
	for _, rawNS := range namespaces {
		ns, _ := rawNS.(map[string]any)
		nsID, _ := ns["id"].(string)
		messages, _ := ns["messages"].([]any)
		seen := map[string]bool{}
		nsMissing := 0
		for _, rawMessage := range messages {
			message, _ := rawMessage.([]any)
			if len(message) < 3 {
				missing = append(missing, map[string]string{"namespace": nsID, "key": fmt.Sprint(message), "locale": "schema", "reason": "message tuple must be [key, ko, en]"})
				nsMissing++
				continue
			}
			key := fmt.Sprint(message[0])
			messageCount++
			if seen[key] {
				duplicateKeys = append(duplicateKeys, nsID+"."+key)
			}
			seen[key] = true
			values := map[string]string{}
			for _, locale := range locales {
				value := strings.TrimSpace(fmt.Sprint(message[localeColumns[locale]]))
				values[locale] = value
				if value == "" {
					missing = append(missing, map[string]string{"namespace": nsID, "key": key, "locale": locale, "reason": "empty translation"})
					nsMissing++
					continue
				}
				translatedByLocale[locale]++
			}
			if !sameStringSet(qaI18NPlaceholders(values[defaultLocale]), qaI18NPlaceholders(values[fallbackLocale])) {
				placeholderMismatches = append(placeholderMismatches, map[string]any{
					"namespace": nsID,
					"key":       key,
					"default":   qaI18NPlaceholders(values[defaultLocale]),
					"fallback":  qaI18NPlaceholders(values[fallbackLocale]),
				})
			}
		}
		namespaceRows = append(namespaceRows, map[string]any{"id": nsID, "message_count": len(messages), "missing_cell_count": nsMissing})
	}
	requiredCells := messageCount * len(locales)
	translatedCells := 0
	for _, locale := range locales {
		translatedCells += translatedByLocale[locale]
	}
	ratio := 0.0
	if requiredCells > 0 {
		ratio = float64(translatedCells) / float64(requiredCells)
	}
	return map[string]any{
		"passed":                     requiredCells > 0 && len(localeErrors) == 0 && len(missing) == 0 && len(placeholderMismatches) == 0 && len(duplicateKeys) == 0,
		"required_locales":           locales,
		"locale_error_count":         len(localeErrors),
		"locale_errors":              localeErrors,
		"namespace_count":            len(namespaces),
		"message_count":              messageCount,
		"required_cell_count":        requiredCells,
		"translated_cell_count":      translatedCells,
		"translated_by_locale":       translatedByLocale,
		"missing_cell_count":         len(missing),
		"missing_cells":              missing,
		"coverage_ratio":             ratio,
		"placeholder_mismatch_count": len(placeholderMismatches),
		"placeholder_mismatches":     placeholderMismatches,
		"duplicate_key_count":        len(duplicateKeys),
		"duplicate_keys":             duplicateKeys,
		"namespaces":                 namespaceRows,
	}
}

func qaI18NPlaceholders(text string) []string {
	matches := qaI18NPlaceholderPattern.FindAllString(text, -1)
	sort.Strings(matches)
	out := []string{}
	for _, match := range matches {
		if len(out) == 0 || out[len(out)-1] != match {
			out = append(out, match)
		}
	}
	return out
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
