// Package upstream implements logic for retrieving latest released software versions
// based on this software upstream source URL.
package upstream

import (
	"regexp"
	"strings"
)

// Version represents the software version as returned by some VersionProvider.
type Version string

func (v Version) GetVersionStr() string {
	return string(v)
}

var versionRegexp = regexp.MustCompile(`^[\w.]+$`)

// ParseVersion makes Version from a string.
// If the string cannot be interpreted as a Version, returns nil.
func ParseVersion(rawVersion string) (Version, bool) {
	if !containsDigit(rawVersion) {
		return Version(""), false
	}
	if !versionRegexp.MatchString(rawVersion) {
		return Version(""), false
	}
	rawVersion = strings.TrimPrefix(rawVersion, "v")
	return Version(rawVersion), true
}

func containsDigit(str string) bool {
	for _, char := range str {
		if char >= '0' && char <= '9' {
			return true
		}
	}
	return false
}
