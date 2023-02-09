// Copyright 2023 Harness Inc. All rights reserved.

// Package slug provides utilities for working with slug values.
package slug

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var safeRanges = []*unicode.RangeTable{
	unicode.Letter,
	unicode.Number,
}

func safe(r rune) rune {
	switch {
	case unicode.IsOneOf(safeRanges, r):
		return unicode.ToLower(r)
	}
	return -1
}

// Create creates a slug from a string.
func Create(s string) string {
	s = norm.NFKD.String(s)
	s = strings.Map(safe, s)
	return s
}
