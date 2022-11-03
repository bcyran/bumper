package version

import (
	"regexp"
	"strings"
)

type Version string

var versionRegexp = regexp.MustCompile(`^[\w.]+$`)

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

type VersionProvider interface {
	LatestVersion() (Version, error)
}

func NewVersionProvider(url string) VersionProvider {
	return NewGitHubProvider(url)
}
