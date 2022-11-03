package version

import (
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestNewGithub_Valid(t *testing.T) {
	validUrl := "https://github.com/bcyran/timewall?foo=bar#whatever"
	expectedResult := gitHubProvider{
		owner: "bcyran",
		repo:  "timewall",
	}

	result := NewGitHubProvider(validUrl)

	assert.Equal(t, expectedResult, *result)
}

func TestNewGithub_Invalid(t *testing.T) {
	invalidUrl := "https://github.com/randompath"

	result := NewGitHubProvider(invalidUrl)

	assert.Nil(t, result)
}

func TestGithubLatestVersion_Release(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/releases").
		Reply(200).
		JSON([]map[string]interface{}{
			{
				"name":       "Draft",
				"tag_name":   "1.8.0",
				"prerelease": false,
				"draft":      true,
			},
			{
				"name":       "Prerelease",
				"tag_name":   "1.7.0",
				"prerelease": true,
				"draft":      false,
			},
			{
				"name":       "Release?",
				"tag_name":   "not-a-version!",
				"prerelease": false,
				"draft":      false,
			},
			{
				"name":       "Foo",
				"tag_name":   "1.6.9",
				"prerelease": false,
				"draft":      false,
			},
		})

	github := gitHubProvider{owner: "foo", repo: "bar"}

	result, err := github.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("1.6.9"), result)
}

func TestGithubLatestVersion_Tag(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/releases").
		Reply(200).
		JSON([]interface{}{})
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/tags").
		Reply(200).
		JSON([]map[string]interface{}{
			{"name": "what-is-this?"},
			{"name": "1.6.9"},
			{"name": "1.6.8"},
		})

	github := gitHubProvider{owner: "foo", repo: "bar"}

	result, err := github.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("1.6.9"), result)
}

func TestGithubLatestVersion_NoReleases(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/releases").
		Reply(200).
		JSON([]interface{}{})
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/tags").
		Reply(200).
		JSON([]interface{}{})

	github := gitHubProvider{owner: "foo", repo: "bar"}

	_, err := github.LatestVersion()

	assert.ErrorIs(t, err, ErrVersionNotFound)
}

func TestGithubLatestVersion_4xx(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/releases").
		Reply(404).
		JSON([]interface{}{})
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/tags").
		Reply(401).
		JSON([]interface{}{})

	github := gitHubProvider{owner: "foo", repo: "bar"}

	_, err := github.LatestVersion()

	assert.ErrorIs(t, err, ErrVersionNotFound)
}

func TestGithubLatestVersion_5xx(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/releases").
		Reply(500).
		JSON([]interface{}{})
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/tags").
		Reply(501).
		JSON([]interface{}{})

	github := gitHubProvider{owner: "foo", repo: "bar"}

	_, err := github.LatestVersion()

	assert.ErrorIs(t, err, ErrProviderError)
}
