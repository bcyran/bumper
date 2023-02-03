package upstream

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
)

// Match any URL which *could* be a GitLab URL and contains 'git' in netloc.
// This is to try and handle instances other than gitlab.com, e.g.: foogit.bar.com.
var gitLabUrlRegex = regexp.MustCompile(`([^/#?]*git[^/#?]*)/([^/#?]+)/([^#?]+?)(/\-/.*)?$`)

// gitLabProvider tries to find the latest version both in releases and tags of a gitLab repo.
type gitLabProvider struct {
	netloc string
	owner  string
	repo   string
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
	return &gitLabProvider{match[1], match[2], match[3]}
}

func (gitLab *gitLabProvider) Equal(other interface{}) bool {
	switch other := other.(type) {
	case *gitLabProvider:
		return gitLab.owner == other.owner && gitLab.repo == other.repo
	default:
		return false
	}
}

func (gitLab *gitLabProvider) projectId() string {
	return url.PathEscape(fmt.Sprintf("%s/%s", gitLab.owner, gitLab.repo))
}

func (gitLab *gitLabProvider) apiURL() string {
	return fmt.Sprintf("https://%s/api/v4", gitLab.netloc)
}

func (gitLab *gitLabProvider) LatestVersion() (Version, error) {
	latestReleaseVersion, releaseErr := gitLab.latestReleaseVersion()
	if releaseErr == nil {
		return latestReleaseVersion, nil
	}

	if !errors.Is(releaseErr, ErrVersionNotFound) {
		return "", releaseErr
	}

	latestTagVersion, tagErr := gitLab.latestTagVersion()
	if tagErr != nil {
		return "", errors.Join(releaseErr, tagErr)
	}

	return latestTagVersion, nil
}

func (gitLab *gitLabProvider) releasesURL() string {
	return fmt.Sprintf("%s/projects/%s/releases", gitLab.apiURL(), gitLab.projectId())
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
	return fmt.Sprintf("%s/projects/%s/repository/tags", gitLab.apiURL(), gitLab.projectId())
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
