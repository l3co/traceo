package slug

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func Generate(name, idPrefix string) string {
	s := removeAccents(name)
	s = strings.ToLower(s)
	s = keepAlphanumericAndSpaces(s)
	s = strings.TrimSpace(s)
	s = compactSpaces(s)
	s = strings.ReplaceAll(s, " ", "-")

	if s == "" {
		s = "missing"
	}

	if len(idPrefix) > 8 {
		idPrefix = idPrefix[:8]
	}

	if idPrefix != "" {
		s = s + "-" + idPrefix
	}

	return s
}

func removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, s)
	if err != nil {
		return s
	}
	return result
}

func keepAlphanumericAndSpaces(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func compactSpaces(s string) string {
	var b strings.Builder
	prev := false
	for _, r := range s {
		if r == ' ' {
			if !prev {
				b.WriteRune(r)
			}
			prev = true
		} else {
			b.WriteRune(r)
			prev = false
		}
	}
	return b.String()
}
