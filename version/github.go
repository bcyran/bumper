package version

import (
	"errors"
	"fmt"
	"regexp"
)

const githubUrlRegex = "github.com/([^/#?]+)/([^/#?]+)"

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

func NewGitHubProvider(url string) *gitHubProvider {
	match := regexp.MustCompile(githubUrlRegex).FindStringSubmatch(url)
	if len(match) == 0 {
		return nil
	}
	return &gitHubProvider{match[1], match[2]}
}

func (github gitHubProvider) LatestVersion() (string, error) {
	latestReleaseVersion, err := github.latestReleaseVersion()
	if err == nil {
		return latestReleaseVersion, nil
	}

	if !errors.Is(err, ErrVersionNotFound) {
		return "", err
	}

	return github.latestTagVersion()
}

func (github gitHubProvider) latestReleaseVersion() (string, error) {
	var latestReleases []githubReleaseResp
	if err := httpGetJSON(github.releasesURL(), &latestReleases); err != nil {
		return "", err
	}

	latestRelease := getLatestPublishedRelease(latestReleases)
	if latestRelease == nil {
		return "", ErrVersionNotFound
	}

	if latestRelease.TagName != "" {
		return latestRelease.TagName, nil
	}
	if latestRelease.Name != "" {
		return latestRelease.Name, nil
	}

	return "", ErrVersionNotFound
}

func getLatestPublishedRelease(releases []githubReleaseResp) *githubReleaseResp {
	for _, release := range releases {
		if !release.Draft && !release.Prerelease {
			return &release
		}
	}
	return nil
}

func (github gitHubProvider) releasesURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", github.owner, github.repo)
}

func (github gitHubProvider) latestTagVersion() (string, error) {
	var latestTags []githubTagResp
	if err := httpGetJSON(github.tagsURL(), &latestTags); err != nil {
		return "", err
	}

	if len(latestTags) == 0 {
		return "", ErrVersionNotFound
	}

	return latestTags[0].Name, nil
}

func (github gitHubProvider) tagsURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", github.owner, github.repo)
}
