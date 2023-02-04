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
	gitLabConfigProvider, _ = config.NewYAML(config.Source(strings.NewReader(
		"{empty: {}, gitlab: {apiKeys: {protected.gitlab.instance.com: test_api_key}}}",
	)))
	gitLabEmptyConfig  = gitLabConfigProvider.Get("empty")
	gitLabApiKeyConfig = gitLabConfigProvider.Get("gitlab")
)

func TestNewGitLab_Valid(t *testing.T) {
	cases := map[string]gitLabProvider{
		"https://gitlab.com/bcyran/timewall": {
			netloc: "gitlab.com",
			owner:  "bcyran",
			repo:   "timewall",
		},
		"https://gitlab.com/bcyran/timewall/-/foo": {
			netloc: "gitlab.com",
			owner:  "bcyran",
			repo:   "timewall",
		},
		"https://gitlab.com/bcyran/group/timewall": {
			netloc: "gitlab.com",
			owner:  "bcyran",
			repo:   "group/timewall",
		},
		"https://gitlab.com/bcyran/group/timewall/-/foo": {
			netloc: "gitlab.com",
			owner:  "bcyran",
			repo:   "group/timewall",
		},
		"https://mygit.bar.com/me/project": {
			netloc: "mygit.bar.com",
			owner:  "me",
			repo:   "project",
		},
		"https://protected.gitlab.instance.com/user/project": {
			netloc: "protected.gitlab.instance.com",
			owner:  "user",
			repo:   "project",
			apiKey: "test_api_key",
		},
	}

	for validUrl, expectedResult := range cases {
		result := newGitLabProvider(validUrl, gitLabApiKeyConfig)
		assert.Equal(t, &expectedResult, result)
	}
}

func TestNewGitLab_Invalid(t *testing.T) {
	invalidUrls := []string{
		"https://gitlab.com/whatever",
		"https://foo.com/user/project",
	}

	for _, invalidUrl := range invalidUrls {
		result := newGitLabProvider(invalidUrl, gitLabEmptyConfig)
		assert.Nil(t, result)
	}
}

func TestGitLabLatestVersion_Release(t *testing.T) {
	defer gock.Off()
	gock.New("https://gitlab.something.com").
		Get("/api/v4/projects/foo/bar/releases").
		AddMatcher(testutils.NoHeaderMatcher("Authorization")).
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
	gitLab := gitLabProvider{netloc: "gitlab.something.com", owner: "foo", repo: "bar"}

	result, err := gitLab.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("1.7.0"), result)
}

func TestGitLabLatestVersion_ReleaseWithApiKey(t *testing.T) {
	defer gock.Off()
	gock.New("https://gitlab.something.com").
		Get("/api/v4/projects/foo/bar/releases").
		MatchHeader("Authorization", "Bearer test_token").
		Reply(200).
		JSON([]map[string]interface{}{
			{
				"name":             "Release!",
				"tag_name":         "1.1.1",
				"upcoming_release": false,
			},
		})
	gitLab := gitLabProvider{netloc: "gitlab.something.com", owner: "foo", repo: "bar", apiKey: "test_token"}

	result, err := gitLab.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("1.1.1"), result)
}
func TestGitLabLatestVersion_Tag(t *testing.T) {
	defer gock.Off()
	gock.New("https://gitLab.com").
		Get("/api/v4/projects/foo/bar/releases").
		AddMatcher(testutils.NoHeaderMatcher("Authorization")).
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

	gitLab := gitLabProvider{netloc: "gitlab.com", owner: "foo", repo: "bar"}

	result, err := gitLab.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("4.2.0"), result)
}

func TestGitLabLatestVersion_TagWithApiKey(t *testing.T) {
	defer gock.Off()
	gock.New("https://gitLab.com").
		Get("/api/v4/projects/foo/bar/releases").
		MatchHeader("Authorization", "Bearer test_token").
		Reply(200).
		JSON([]interface{}{})
	gock.New("https://gitLab.com").
		Get("/api/v4/projects/foo/bar/repository/tags").
		MatchHeader("Authorization", "Bearer test_token").
		Reply(200).
		JSON([]map[string]interface{}{
			{"name": "1.1.1"},
		})

	gitLab := gitLabProvider{netloc: "gitlab.com", owner: "foo", repo: "bar", apiKey: "test_token"}

	result, err := gitLab.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("1.1.1"), result)
}

func TestGitLabLatestVersion_NoVersions(t *testing.T) {
	defer gock.Off()
	gock.New("https://gitlab.com").
		Get("/api/v4/projects/foo/bar/releases").
		AddMatcher(testutils.NoHeaderMatcher("Authorization")).
		Reply(200).
		JSON([]interface{}{})
	gock.New("https://gitlab.com").
		Get("/api/v4/projects/foo/bar/repository/tags").
		AddMatcher(testutils.NoHeaderMatcher("Authorization")).
		Reply(200).
		JSON([]interface{}{})

	gitLab := gitLabProvider{netloc: "gitlab.com", owner: "foo", repo: "bar"}

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

	gitLab := gitLabProvider{netloc: "gitlab.com", owner: "foo", repo: "bar"}

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

	gitLab := gitLabProvider{netloc: "gitlab.com", owner: "foo", repo: "bar"}

	_, err := gitLab.LatestVersion()

	assert.ErrorIs(t, err, ErrProviderError)
}
