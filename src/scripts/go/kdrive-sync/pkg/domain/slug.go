package domain

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	slugNonAlpha   = regexp.MustCompile(`[^a-z0-9]+`)
	slugDiacritics = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
)

// Slugify converts an arbitrary string to a URL-safe lowercase slug.
//
// Diacritics are stripped via Unicode NFD decomposition; the result contains
// only [a-z0-9-] with no leading or trailing dashes.
func Slugify(s string) string {
	s = strings.ToLower(s)
	s, _, _ = transform.String(slugDiacritics, s)
	s = slugNonAlpha.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
