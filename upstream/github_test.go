package upstream

import (
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestNewGithub_Valid(t *testing.T) {
	validUrl := "https://github.com/bcyran/timewall?foo=bar#whatever"
	expectedResult := GitHub{
		owner: "bcyran",
		repo:  "timewall",
	}

	result := NewGitHub(validUrl)

	assert.Equal(t, expectedResult, *result)
}

func TestNewGithub_Invalid(t *testing.T) {
	invalidUrl := "https://github.com/randompath"

	result := NewGitHub(invalidUrl)

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
				"name":       "Foo",
				"tag_name":   "1.6.9",
				"prerelease": false,
				"draft":      false,
			},
		})

	github := GitHub{owner: "foo", repo: "bar"}

	result, err := github.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, "1.6.9", result)
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
			{"name": "1.6.9"},
			{"name": "1.6.8"},
		})

	github := GitHub{owner: "foo", repo: "bar"}

	result, err := github.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, "1.6.9", result)
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

	github := GitHub{owner: "foo", repo: "bar"}

	result, err := github.LatestVersion()

	assert.ErrorIs(t, err, ErrVersionNotFound)
	assert.Equal(t, "", result)
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

	github := GitHub{owner: "foo", repo: "bar"}

	result, err := github.LatestVersion()

	assert.ErrorIs(t, err, ErrVersionNotFound)
	assert.Equal(t, "", result)
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

	github := GitHub{owner: "foo", repo: "bar"}

	result, err := github.LatestVersion()

	assert.ErrorIs(t, err, ErrUpstreamError)
	assert.Equal(t, "", result)
}
