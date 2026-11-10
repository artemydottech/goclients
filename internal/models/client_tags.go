package models

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

func NormalizeClientTags(tags []string) ([]string, error) {
	if len(tags) > 20 {
		return nil, Invalid("a client cannot have more than 20 tags")
	}
	unique := map[string]bool{}
	result := []string{}
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" || utf8.RuneCountInString(tag) > 32 || strings.IndexFunc(tag, unicode.IsControl) >= 0 {
			return nil, Invalid("tags must contain 1 to 32 characters without control characters")
		}
		if !unique[tag] {
			unique[tag] = true
			result = append(result, tag)
		}
	}
	sort.Strings(result)
	return result, nil
}
