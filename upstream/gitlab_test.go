package upstream

import (
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestNewGitLab_Valid(t *testing.T) {
	cases := map[string]gitLabProvider{
		"https://gitlab.com/bcyran/timewall":             {owner: "bcyran", repo: "timewall"},
		"https://gitlab.com/bcyran/timewall/-/foo":       {owner: "bcyran", repo: "timewall"},
		"https://gitlab.com/bcyran/group/timewall":       {owner: "bcyran", repo: "group/timewall"},
		"https://gitlab.com/bcyran/group/timewall/-/foo": {owner: "bcyran", repo: "group/timewall"},
	}

	for validUrl, expectedResult := range cases {
		result := newGitLabProvider(validUrl)
		assert.Equal(t, &expectedResult, result)
	}
}

func TestNewGitLab_Invalid(t *testing.T) {
	invalidUrl := "https://gitlab.com/whatever"

	result := newGitLabProvider(invalidUrl)

	assert.Nil(t, result)
}

func TestGitLabLatestVersion_Release(t *testing.T) {
	defer gock.Off()
	gock.New("https://gitlab.com").
		Get("/api/v4/projects/foo/bar/releases").
		Reply(200).
		JSON([]map[string]interface{}{
			{
				"name":             "Upcoming",
				"tag_name":         "1.8.0",
				"upcoming_release": true,
			},
			{
				"name":             "Release?",
				"tag_name":         "not-a-version!",
				"upcoming_release": false,
			},
			{
				"name":             "Release!",
				"tag_name":         "1.7.0",
				"upcoming_release": false,
			},
		})
	gitLab := gitLabProvider{owner: "foo", repo: "bar"}

	result, err := gitLab.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("1.7.0"), result)
}

func TestGitLabLatestVersion_Tag(t *testing.T) {
	defer gock.Off()
	gock.New("https://gitLab.com").
		Get("/api/v4/projects/foo/bar/releases").
		Reply(200).
		JSON([]interface{}{})
	gock.New("https://gitLab.com").
		Get("/api/v4/projects/foo/bar/repository/tags").
		Reply(200).
		JSON([]map[string]interface{}{
			{"name": "what-is-this?"},
			{"name": "4.2.0"},
			{"name": "4.1.0"},
		})

	gitLab := gitLabProvider{owner: "foo", repo: "bar"}

	result, err := gitLab.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("4.2.0"), result)
}

func TestGitLabLatestVersion_NoVersions(t *testing.T) {
	defer gock.Off()
	gock.New("https://gitlab.com").
		Get("/api/v4/projects/foo/bar/releases").
		Reply(200).
		JSON([]interface{}{})
	gock.New("https://gitlab.com").
		Get("/api/v4/projects/foo/bar/repository/tags").
		Reply(200).
		JSON([]interface{}{})

	gitLab := gitLabProvider{owner: "foo", repo: "bar"}

	_, err := gitLab.LatestVersion()

	assert.ErrorIs(t, err, ErrVersionNotFound)
}

func TestGitLabLatestVersion_4xx(t *testing.T) {
	defer gock.Off()
	gock.New("https://gitlab.com").
		Get("/api/v4/projects/foo/bar/releases").
		Reply(404).
		JSON([]interface{}{})
	gock.New("https://gitlab.com").
		Get("/api/v4/projects/foo/bar/repository/tags").
		Reply(401).
		JSON([]interface{}{})

	gitLab := gitLabProvider{owner: "foo", repo: "bar"}

	_, err := gitLab.LatestVersion()

	assert.ErrorIs(t, err, ErrVersionNotFound)
}

func TestGitLabLatestVersion_5xx(t *testing.T) {
	defer gock.Off()
	gock.New("https://gitlab.com").
		Get("/api/v4/projects/foo/bar/releases").
		Reply(500).
		JSON([]interface{}{})
	gock.New("https://gitlab.com").
		Get("/api/v4/projects/foo/bar/repository/tags").
		Reply(501).
		JSON([]interface{}{})

	gitLab := gitLabProvider{owner: "foo", repo: "bar"}

	_, err := gitLab.LatestVersion()

	assert.ErrorIs(t, err, ErrProviderError)
}
