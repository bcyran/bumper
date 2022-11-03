// Package version implements logic for retrieving latest released software versions
// based on this software source URL.
package version

import (
	"regexp"
	"strings"
)

// Version represents the software version as returned by some VersionProvider.
type Version string

var versionRegexp = regexp.MustCompile(`^[\w.]+$`)

// parseVersion makes Version from a string.
// If the string cannot be interpreted as a Version, returns nil.
func parseVersion(rawVersion string) (Version, bool) {
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

// VersionProvider tries to find the latest software version based on its source URL.
type VersionProvider interface {
	LatestVersion() (Version, error)
}

// NewVersionProvider tries to create a VersionProvider instance for a given URL.
// Returns nil if there's no suitable provider.
func NewVersionProvider(url string) VersionProvider {
	return NewGitHubProvider(url)
}
