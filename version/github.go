package version

import (
	"errors"
	"fmt"
	"regexp"
)

const githubUrlRegex = "github.com/([^/#?]+)/([^/#?]+)"

// gitHubProvider tries to find the latest version both in releases and tags of a GitHub repo.
type gitHubProvider struct {
	owner string
	repo  string
}

type githubReleaseResp struct {
	Name       string `json:"name"`
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}

type githubTagResp struct {
	Name string `json:"name"`
}

func newGitHubProvider(url string) *gitHubProvider {
	match := regexp.MustCompile(githubUrlRegex).FindStringSubmatch(url)
	if len(match) == 0 {
		return nil
	}
	return &gitHubProvider{match[1], match[2]}
}

func (github gitHubProvider) LatestVersion() (Version, error) {
	latestReleaseVersion, err := github.latestReleaseVersion()
	if err == nil {
		return latestReleaseVersion, nil
	}

	if !errors.Is(err, ErrVersionNotFound) {
		return "", err
	}

	return github.latestTagVersion()
}

func (github gitHubProvider) latestReleaseVersion() (Version, error) {
	var latestReleases []githubReleaseResp
	if err := httpGetJSON(github.releasesURL(), &latestReleases); err != nil {
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

func (github gitHubProvider) releasesURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", github.owner, github.repo)
}

func (github gitHubProvider) latestTagVersion() (Version, error) {
	var latestTags []githubTagResp
	if err := httpGetJSON(github.tagsURL(), &latestTags); err != nil {
		return "", err
	}

	for _, tag := range latestTags {
		if version, isValid := parseVersion(tag.Name); isValid == true {
			return version, nil
		}
	}

	return "", ErrVersionNotFound
}

func (github gitHubProvider) tagsURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", github.owner, github.repo)
}
