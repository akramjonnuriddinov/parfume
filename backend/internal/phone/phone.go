// Package phone normalizes customer phone numbers.
package phone

import "regexp"

var nonDigit = regexp.MustCompile(`\D`)

// Normalize keeps digits (with a leading +) and checks the length.
func Normalize(raw string) (string, bool) {
	digits := nonDigit.ReplaceAllString(raw, "")
	if len(digits) < 9 || len(digits) > 15 {
		return "", false
	}
	// Local Uzbek numbers (9 digits) get the country code.
	if len(digits) == 9 {
		digits = "998" + digits
	}
	return "+" + digits, true
}
