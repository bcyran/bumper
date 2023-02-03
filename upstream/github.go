package upstream

import (
	"errors"
	"fmt"
	"regexp"
)

var gitHubUrlRegex = regexp.MustCompile(`github\.com/([^/#?]+)/([^/#?]+)`)

// gitHubProvider tries to find the latest version both in releases and tags of a GitHub repo.
type gitHubProvider struct {
	owner string
	repo  string
}

type gitHubReleaseResp struct {
	Name       string `json:"name"`
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}

type gitHubTagResp struct {
	Name string `json:"name"`
}

func newGitHubProvider(url string) *gitHubProvider {
	match := gitHubUrlRegex.FindStringSubmatch(url)
	if len(match) == 0 {
		return nil
	}
	return &gitHubProvider{match[1], match[2]}
}

func (gitHub *gitHubProvider) Equal(other interface{}) bool {
	switch other := other.(type) {
	case *gitHubProvider:
		return gitHub.owner == other.owner && gitHub.repo == other.repo
	default:
		return false
	}
}

func (gitHub *gitHubProvider) LatestVersion() (Version, error) {
	latestReleaseVersion, releaseErr := gitHub.latestReleaseVersion()
	if releaseErr == nil {
		return latestReleaseVersion, nil
	}

	if !errors.Is(releaseErr, ErrVersionNotFound) {
		return "", releaseErr
	}

	latestTagVersion, tagErr := gitHub.latestTagVersion()
	if tagErr != nil {
		return "", errors.Join(releaseErr, tagErr)
	}

	return latestTagVersion, nil
}

func (gitHub *gitHubProvider) latestReleaseVersion() (Version, error) {
	var latestReleases []gitHubReleaseResp
	if err := httpGetJSON(gitHub.releasesURL(), &latestReleases); err != nil {
		return "", err
	}

	for _, release := range latestReleases {
		if release.Draft || release.Prerelease {
			continue
		}
		if version, isValid := parseVersion(release.TagName); isValid == true {
			return version, nil
		}
		if version, isValid := parseVersion(release.Name); isValid == true {
			return version, nil
		}
	}

	return "", ErrVersionNotFound
}

func (gitHub *gitHubProvider) releasesURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", gitHub.owner, gitHub.repo)
}

func (gitHub *gitHubProvider) latestTagVersion() (Version, error) {
	var latestTags []gitHubTagResp
	if err := httpGetJSON(gitHub.tagsURL(), &latestTags); err != nil {
		return "", err
	}

	for _, tag := range latestTags {
		if version, isValid := parseVersion(tag.Name); isValid == true {
			return version, nil
		}
	}

	return "", ErrVersionNotFound
}

func (gitHub *gitHubProvider) tagsURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", gitHub.owner, gitHub.repo)
}
