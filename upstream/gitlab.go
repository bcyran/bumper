package upstream

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"go.uber.org/config"
)

// Match any URL which *could* be a GitLab URL and contains 'git' in netloc.
// This is to try and handle instances other than gitlab.com, e.g.: foogit.bar.com.
var gitLabURLRegex = regexp.MustCompile(`([^/#?]*git[^/#?]*)/([^/#?]+)/([^#?]+?)(/\-/.*)?$`)

// gitLabProvider tries to find the latest version both in releases and tags of a gitLab repo.
type gitLabProvider struct {
	netloc string
	owner  string
	repo   string
	apiKey string
}

type gitLabReleaseResp struct {
	Name     string `json:"name"`
	TagName  string `json:"tag_name"`
	Upcoming bool   `json:"upcoming_release"`
}

type gitLabTagResp struct {
	Name string `json:"name"`
}

func newGitLabProvider(url string, gitLabConfig config.Value) *gitLabProvider {
	match := gitLabURLRegex.FindStringSubmatch(url)
	if len(match) == 0 {
		return nil
	}

	provider := gitLabProvider{netloc: match[1], owner: match[2], repo: match[3]}
	// config.Value.Get(path string) doesn't work when path contains dots, like URLs
	apiKeysMap := map[string]string{}
	gitLabConfig.Get("apiKeys").Populate(&apiKeysMap) //nolint:errcheck
	if apiKey, apiKeyPresent := apiKeysMap[provider.netloc]; apiKeyPresent {
		provider.apiKey = apiKey
	}

	return &provider
}

func (gitLab *gitLabProvider) Equal(other interface{}) bool {
	switch other := other.(type) {
	case *gitLabProvider:
		return gitLab.owner == other.owner && gitLab.repo == other.repo
	default:
		return false
	}
}

func (gitLab *gitLabProvider) projectID() string {
	return url.PathEscape(fmt.Sprintf("%s/%s", gitLab.owner, gitLab.repo))
}

func (gitLab *gitLabProvider) apiURL() string {
	return fmt.Sprintf("https://%s/api/v4", gitLab.netloc)
}

func (gitLab *gitLabProvider) apiHeaders() map[string]string {
	headers := map[string]string{}
	if gitLab.apiKey != "" {
		headers["Authorization"] = "Bearer " + gitLab.apiKey
	}
	return headers
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
	return fmt.Sprintf("%s/projects/%s/releases", gitLab.apiURL(), gitLab.projectID())
}

func (gitLab *gitLabProvider) latestReleaseVersion() (Version, error) {
	var latestReleases []gitLabReleaseResp
	if err := httpGetJSON(gitLab.releasesURL(), &latestReleases, gitLab.apiHeaders()); err != nil {
		return "", err
	}

	for _, release := range latestReleases {
		if release.Upcoming {
			continue
		}
		if version, isValid := parseVersion(release.TagName); isValid {
			return version, nil
		}
		if version, isValid := parseVersion(release.Name); isValid {
			return version, nil
		}
	}

	return "", ErrVersionNotFound
}

func (gitLab *gitLabProvider) tagsURL() string {
	return fmt.Sprintf("%s/projects/%s/repository/tags", gitLab.apiURL(), gitLab.projectID())
}

func (gitLab *gitLabProvider) latestTagVersion() (Version, error) {
	var latestTags []gitLabTagResp
	if err := httpGetJSON(gitLab.tagsURL(), &latestTags, gitLab.apiHeaders()); err != nil {
		return "", err
	}

	for _, tag := range latestTags {
		if version, isValid := parseVersion(tag.Name); isValid {
			return version, nil
		}
	}

	return "", ErrVersionNotFound
}
