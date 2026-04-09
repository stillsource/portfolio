package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	regexpNonAlpha = regexp.MustCompile(`[^a-z0-9]+`)
)

// Slugify converts a string into a clean, URL-friendly slug.
// It removes accents, converts to lowercase, and replaces non-alphanumeric chars with hyphens.
func Slugify(s string) string {
	// 1. Convert to lowercase
	s = strings.ToLower(s)

	// 2. Remove accents/diacritics using unicode normalization
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(t, s)

	// 3. Replace non-alphanumeric characters with hyphens
	s = regexpNonAlpha.ReplaceAllString(s, "-")

	// 4. Trim hyphens from ends
	return strings.Trim(s, "-")
}
