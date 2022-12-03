package upstream

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
)

var gitLabUrlRegex = regexp.MustCompile(`gitlab\.com/([^/#?]+)/([^#?]+?)(/\-/.*)?$`)

// gitLabProvider tries to find the latest version both in releases and tags of a gitLab repo.
type gitLabProvider struct {
	owner string
	repo  string
}

type gitLabReleaseResp struct {
	Name     string `json:"name"`
	TagName  string `json:"tag_name"`
	Upcoming bool   `json:"upcoming_release"`
}

type gitLabTagResp struct {
	Name string `json:"name"`
}

func newGitLabProvider(url string) *gitLabProvider {
	match := gitLabUrlRegex.FindStringSubmatch(url)
	if len(match) == 0 {
		return nil
	}
	return &gitLabProvider{match[1], match[2]}
}

func (gitLab *gitLabProvider) projectId() string {
	return url.PathEscape(fmt.Sprintf("%s/%s", gitLab.owner, gitLab.repo))
}

func (gitLab *gitLabProvider) LatestVersion() (Version, error) {
	latestReleaseVersion, err := gitLab.latestReleaseVersion()
	if err == nil {
		return latestReleaseVersion, nil
	}

	if !errors.Is(err, ErrVersionNotFound) {
		return "", err
	}

	return gitLab.latestTagVersion()
}

func (gitLab *gitLabProvider) releasesURL() string {
	return fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/releases", gitLab.projectId())
}

func (gitLab *gitLabProvider) latestReleaseVersion() (Version, error) {
	var latestReleases []gitLabReleaseResp
	if err := httpGetJSON(gitLab.releasesURL(), &latestReleases); err != nil {
		return "", err
	}

	for _, release := range latestReleases {
		if release.Upcoming {
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

func (gitLab *gitLabProvider) tagsURL() string {
	return fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/repository/tags", gitLab.projectId())
}

func (gitLab *gitLabProvider) latestTagVersion() (Version, error) {
	var latestTags []gitLabTagResp
	if err := httpGetJSON(gitLab.tagsURL(), &latestTags); err != nil {
		return "", err
	}

	for _, tag := range latestTags {
		if version, isValid := parseVersion(tag.Name); isValid == true {
			return version, nil
		}
	}

	return "", ErrVersionNotFound
}
