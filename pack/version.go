package pack

import (
	"strings"
	"unicode"
)

// VersionLike is an interface for structs that can be treated as a version.
type VersionLike interface {
	// GetVersionStr returns the version represented by VersionLike as string.
	GetVersionStr() string
}

// VersionCmp compares two VersionLike structs.
// Returns 1 if a is newer than b, 0 if a and b are the same version, -1 if b is newer than a.
func VersionCmp(a, b VersionLike) int {
	return Rpmvercmp(a.GetVersionStr(), b.GetVersionStr())
}

// Rpmvercmp compares a and b version strings.
// Returns 1 if a is newer than b, 0 if a and b are the same version, -1 if b is newer than a.
// Tries to mimick behavior of rpmvercmp from libalpm: https://gitlab.archlinux.org/pacman/pacman/-/blob/master/lib/libalpm/version.c.
func Rpmvercmp(a, b string) int {
	var aSegStart, bSegStart, aSegEnd, bSegEnd int
	aSegStart, bSegStart, aSegEnd, bSegEnd = 0, 0, 0, 0
	var aSeg, bSeg string
	var isNum bool

	// Easy comparison to see if versions are identical.
	if a == b {
		return 0
	}

	// Loop through the version strings.
	for aSegStart < len(a) && bSegStart < len(b) {
		// Advance aSegStart and bSegStart to the next segment start -
		// the next character that is not a separator.
		for aSegStart < len(a) && !isAlNum(a[aSegStart]) {
			aSegStart++
		}
		for bSegStart < len(b) && !isAlNum(b[bSegStart]) {
			bSegStart++
		}

		// If we ran to the end of either strings, we are finished with the loop.
		if !(aSegStart < len(a) && bSegStart < len(b)) {
			break
		}

		// If the separator lengths were different, we are also finished.
		// aSegEnd and bSegEnd point at the PREVIOUS segment end at this point.
		if (aSegStart - aSegEnd) < (bSegStart - bSegEnd) {
			return -1
		} else if (aSegStart - aSegEnd) > (bSegStart - bSegEnd) {
			return 1
		}

		// Place aSegEnd and bSegEnd at the start of our current segments...
		aSegEnd = aSegStart
		bSegEnd = bSegStart

		// ... and advance them to the end of the completely alpha or completely
		// numeric substring. This is the end of the segment.
		if isDigit(a[aSegEnd]) {
			for aSegEnd < len(a) && isDigit(a[aSegEnd]) {
				aSegEnd++
			}
			for bSegEnd < len(b) && isDigit(b[bSegEnd]) {
				bSegEnd++
			}
			isNum = true
		} else {
			for aSegEnd < len(a) && isAlpha(a[aSegEnd]) {
				aSegEnd++
			}
			for bSegEnd < len(b) && isAlpha(b[bSegEnd]) {
				bSegEnd++
			}
			isNum = false
		}

		// If bSegEnd didn't advance at all it means the segments are of a different type.
		// The numeric one wins.
		if bSegEnd == bSegStart {
			if isNum {
				return 1
			} else {
				return -1
			}
		}

		// Let's cut out our segments.
		aSeg = a[aSegStart:aSegEnd]
		bSeg = b[bSegStart:bSegEnd]

		// Trim the leading zeros if they're are numeric.
		if isNum {
			aSeg = strings.TrimLeft(aSeg, "0")
			bSeg = strings.TrimLeft(bSeg, "0")
		}

		// Compare the segment strings.
		// String comparison will work correctly, no matter if they're alpha or numeric.
		if aSeg > bSeg {
			return 1
		} else if aSeg < bSeg {
			return -1
		}

		// Place aSegStart and bSegStart at the end of the current segment.
		aSegStart = aSegEnd
		bSegStart = bSegEnd
	}

	// This catches the case where all numeric and alpha segments have
	// compare identically, but the segment separating characters were different.
	if aSegStart >= len(a) && bSegStart >= len(b) {
		return 0
	}

	// The final showdown. We never want a remaining alpha string to beat an empty string.
	// The logic is a bit weird, but:
	// - if there is nothing remaining in a, and remaining in b is not alpha, b is newer;
	// - if remaining in a is alpha, b is newer;
	// - otherwise a is newer.
	if aSegStart >= len(a) && !isAlpha(b[bSegStart]) || aSegStart < len(a) && isAlpha(a[aSegStart]) {
		return -1
	} else {
		return 1
	}
}

func isAlNum(b byte) bool {
	return isDigit(b) || isAlpha(b)
}

func isDigit(b byte) bool {
	return unicode.IsNumber(rune(b))
}

func isAlpha(b byte) bool {
	return unicode.IsLetter(rune(b))
}
