package upstream

import (
	"strings"
	"testing"

	"github.com/bcyran/bumper/internal/testutils"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/config"
)

var (
	gitHubConfigProvider, _ = config.NewYAML(config.Source(strings.NewReader(
		"{empty: {}, github: {apiKey: test_api_key}}",
	)))
	gitHubEmptyConfig  = gitHubConfigProvider.Get("empty")
	gitHubAPIKeyConfig = gitHubConfigProvider.Get("github")
)

func TestNewGithub_Valid(t *testing.T) {
	validURL := "https://github.com/bcyran/timewall?foo=bar#whatever"
	expectedResult := gitHubProvider{
		owner: "bcyran",
		repo:  "timewall",
	}

	result := newGitHubProvider(validURL, gitHubEmptyConfig)

	assert.Equal(t, &expectedResult, result)
}

func TestNewGithub_ValidWithApiKey(t *testing.T) {
	validURL := "https://github.com/bcyran/timewall?foo=bar#whatever"
	expectedResult := gitHubProvider{
		owner:  "bcyran",
		repo:   "timewall",
		apiKey: "test_api_key",
	}

	result := newGitHubProvider(validURL, gitHubAPIKeyConfig)

	assert.Equal(t, &expectedResult, result)
}

func TestNewGithub_Invalid(t *testing.T) {
	invalidURL := "https://github.com/randompath"

	result := newGitHubProvider(invalidURL, gitHubEmptyConfig)

	assert.Nil(t, result)
}

func TestGithubLatestVersion_Release(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/releases").
		AddMatcher(testutils.NoHeaderMatcher("Authorization")).
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

	gitHub := gitHubProvider{owner: "foo", repo: "bar"}

	result, err := gitHub.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("1.6.9"), result)
}

func TestGithubLatestVersion_ReleaseWithApiKey(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/releases").
		MatchHeader("Authorization", "Bearer test_token").
		Reply(200).
		JSON([]map[string]interface{}{
			{
				"name":       "Foo",
				"tag_name":   "1.1.1",
				"prerelease": false,
				"draft":      false,
			},
		})

	gitHub := gitHubProvider{owner: "foo", repo: "bar", apiKey: "test_token"}

	result, err := gitHub.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("1.1.1"), result)
}

func TestGithubLatestVersion_Tag(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/releases").
		AddMatcher(testutils.NoHeaderMatcher("Authorization")).
		Reply(200).
		JSON([]interface{}{})
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/tags").
		AddMatcher(testutils.NoHeaderMatcher("Authorization")).
		Reply(200).
		JSON([]map[string]interface{}{
			{"name": "what-is-this?"},
			{"name": "1.6.9"},
			{"name": "1.6.8"},
		})

	gitHub := gitHubProvider{owner: "foo", repo: "bar"}

	result, err := gitHub.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("1.6.9"), result)
}

func TestGithubLatestVersion_TagWithApiKey(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/releases").
		MatchHeader("Authorization", "Bearer test_token").
		Reply(200).
		JSON([]interface{}{})
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/tags").
		MatchHeader("Authorization", "Bearer test_token").
		Reply(200).
		JSON([]map[string]interface{}{
			{"name": "1.6.9"},
		})

	gitHub := gitHubProvider{owner: "foo", repo: "bar", apiKey: "test_token"}

	result, err := gitHub.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("1.6.9"), result)
}

func TestGithubLatestVersion_NoVersions(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/releases").
		Reply(200).
		JSON([]interface{}{})
	gock.New("https://api.github.com").
		Get("/repos/foo/bar/tags").
		Reply(200).
		JSON([]interface{}{})

	gitHub := gitHubProvider{owner: "foo", repo: "bar"}

	_, err := gitHub.LatestVersion()

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

	gitHub := gitHubProvider{owner: "foo", repo: "bar"}

	_, err := gitHub.LatestVersion()

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

	gitHub := gitHubProvider{owner: "foo", repo: "bar"}

	_, err := gitHub.LatestVersion()

	assert.ErrorIs(t, err, ErrProviderError)
}
