package api

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"parfume/backend/internal/phone"
)

var (
	slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

func normalizePhone(raw string) (string, bool) {
	return phone.Normalize(raw)
}

func lengthBetween(s string, min, max int) bool {
	n := utf8.RuneCountInString(s)
	return n >= min && n <= max
}

func clean(s string) string {
	return strings.TrimSpace(s)
}

func normalizeLocale(s string) string {
	switch s {
	case "uz", "ru", "en":
		return s
	}
	return "uz"
}
